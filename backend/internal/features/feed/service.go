package feed

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct {
	repo   Repository
	groups GroupChecker
}

func NewService(repo Repository, groups GroupChecker) *Service {
	return &Service{repo: repo, groups: groups}
}

// CreatePost publishes a post to a group (teacher only).
func (s *Service) CreatePost(ctx context.Context, authorID, groupID uuid.UUID, req CreatePostRequest) (*domain.PostWithMeta, error) {
	g, err := s.groups.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("feed.CreatePost fetch group: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}
	if g.TeacherID != authorID {
		return nil, domain.ErrForbidden
	}

	now := time.Now()
	p := &domain.Post{
		ID:        uuid.New(),
		GroupID:   groupID,
		AuthorID:  authorID,
		Body:      strings.TrimSpace(req.Body),
		Pinned:    req.Pinned,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreatePost(ctx, p); err != nil {
		return nil, fmt.Errorf("feed.CreatePost: %w", err)
	}

	return &domain.PostWithMeta{Post: *p}, nil
}

// GetPost returns a single post if the caller can see it.
func (s *Service) GetPost(ctx context.Context, postID, callerID uuid.UUID, role domain.Role) (*domain.PostWithMeta, error) {
	p, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("feed.GetPost: %w", err)
	}
	if p == nil {
		return nil, domain.ErrNotFound
	}

	if err := s.assertVisible(ctx, p.GroupID, callerID, role); err != nil {
		return nil, err
	}
	out := []*domain.PostWithMeta{{Post: *p}}
	if err := s.attachReactions(ctx, out); err != nil {
		return nil, err
	}
	return out[0], nil
}

// ListPosts returns paginated posts for a group.
func (s *Service) ListPosts(ctx context.Context, groupID, callerID uuid.UUID, role domain.Role, limit, offset int) ([]*domain.PostWithMeta, error) {
	if err := s.assertVisible(ctx, groupID, callerID, role); err != nil {
		return nil, err
	}

	limit = clampLimit(limit)
	posts, err := s.repo.ListGroupPosts(ctx, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("feed.ListPosts: %w", err)
	}
	if err := s.attachReactions(ctx, posts); err != nil {
		return nil, err
	}
	return posts, nil
}

// React ставит или меняет реакцию ученика на объявление.
func (s *Service) React(ctx context.Context, postID, callerID uuid.UUID, role domain.Role, kind domain.ReactionKind) error {
	if err := s.assertCanReact(ctx, postID, callerID, role); err != nil {
		return err
	}
	if err := s.repo.SetReaction(ctx, postID, callerID, kind); err != nil {
		return fmt.Errorf("feed.React: %w", err)
	}
	return nil
}

// Unreact снимает реакцию ученика.
func (s *Service) Unreact(ctx context.Context, postID, callerID uuid.UUID, role domain.Role) error {
	if err := s.assertCanReact(ctx, postID, callerID, role); err != nil {
		return err
	}
	if err := s.repo.DeleteReaction(ctx, postID, callerID); err != nil {
		return fmt.Errorf("feed.Unreact: %w", err)
	}
	return nil
}

// Реагируют только ученики класса: репетитор — автор объявлений, родитель —
// наблюдатель.
func (s *Service) assertCanReact(ctx context.Context, postID, callerID uuid.UUID, role domain.Role) error {
	if role != domain.RoleStudent {
		return domain.ErrForbidden
	}
	p, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("feed.assertCanReact: %w", err)
	}
	if p == nil {
		return domain.ErrNotFound
	}
	return s.assertVisible(ctx, p.GroupID, callerID, role)
}

func (s *Service) attachReactions(ctx context.Context, posts []*domain.PostWithMeta) error {
	if len(posts) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(posts))
	byID := make(map[uuid.UUID]*domain.PostWithMeta, len(posts))
	for i, p := range posts {
		ids[i] = p.ID
		byID[p.ID] = p
	}
	reactions, err := s.repo.ListReactions(ctx, ids)
	if err != nil {
		return fmt.Errorf("feed.attachReactions: %w", err)
	}
	for _, r := range reactions {
		if p := byID[r.PostID]; p != nil {
			p.Reactions = append(p.Reactions, r)
		}
	}
	return nil
}

// UpdatePost edits body or pinned flag (author only).
func (s *Service) UpdatePost(ctx context.Context, postID, authorID uuid.UUID, req UpdatePostRequest) (*domain.PostWithMeta, error) {
	p, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("feed.UpdatePost fetch: %w", err)
	}
	if p == nil {
		return nil, domain.ErrNotFound
	}
	if p.AuthorID != authorID {
		return nil, domain.ErrForbidden
	}

	if req.Body != nil {
		trimmed := strings.TrimSpace(*req.Body)
		p.Body = trimmed
	}
	if req.Pinned != nil {
		p.Pinned = *req.Pinned
	}
	p.UpdatedAt = time.Now()

	if err := s.repo.UpdatePost(ctx, p); err != nil {
		return nil, fmt.Errorf("feed.UpdatePost save: %w", err)
	}

	return &domain.PostWithMeta{Post: *p}, nil
}

// DeletePost soft-deletes a post (author only).
func (s *Service) DeletePost(ctx context.Context, postID, authorID uuid.UUID) error {
	p, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("feed.DeletePost fetch: %w", err)
	}
	if p == nil {
		return domain.ErrNotFound
	}
	if p.AuthorID != authorID {
		return domain.ErrForbidden
	}

	if err := s.repo.SoftDeletePost(ctx, postID); err != nil {
		return fmt.Errorf("feed.DeletePost: %w", err)
	}
	return nil
}

// PinPost toggles pinned flag (group teacher only).
func (s *Service) PinPost(ctx context.Context, postID, teacherID uuid.UUID, pinned bool) (*domain.PostWithMeta, error) {
	p, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("feed.PinPost fetch: %w", err)
	}
	if p == nil {
		return nil, domain.ErrNotFound
	}

	g, err := s.groups.GetGroupByID(ctx, p.GroupID)
	if err != nil {
		return nil, fmt.Errorf("feed.PinPost fetch group: %w", err)
	}
	if g == nil || g.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	if err := s.repo.PinPost(ctx, postID, pinned); err != nil {
		return nil, fmt.Errorf("feed.PinPost save: %w", err)
	}

	p.Pinned = pinned
	return &domain.PostWithMeta{Post: *p}, nil
}

// assertVisible checks that callerID may read posts in groupID.
// Teacher must own the group; student/parent must be a member.
func (s *Service) assertVisible(ctx context.Context, groupID, callerID uuid.UUID, role domain.Role) error {
	g, err := s.groups.GetGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("assertVisible: %w", err)
	}
	if g == nil {
		return domain.ErrNotFound
	}

	if role == domain.RoleTeacher {
		if g.TeacherID != callerID {
			return domain.ErrForbidden
		}
		return nil
	}

	m, err := s.groups.GetMember(ctx, groupID, callerID)
	if err != nil {
		return fmt.Errorf("assertVisible: %w", err)
	}
	if m == nil {
		return domain.ErrForbidden
	}
	return nil
}

func clampLimit(l int) int {
	if l <= 0 {
		return defaultLimit
	}
	if l > maxLimit {
		return maxLimit
	}
	return l
}
