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
	repo    Repository
	groups  GroupChecker
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
func (s *Service) GetPost(ctx context.Context, postID, callerID uuid.UUID, role domain.Role) (*domain.PostWithMeta, []*domain.PostAttachment, error) {
	p, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, nil, fmt.Errorf("feed.GetPost: %w", err)
	}
	if p == nil {
		return nil, nil, domain.ErrNotFound
	}

	if err := s.assertVisible(ctx, p.GroupID, callerID, role); err != nil {
		return nil, nil, err
	}

	atts, err := s.repo.ListAttachments(ctx, postID)
	if err != nil {
		return nil, nil, fmt.Errorf("feed.GetPost attachments: %w", err)
	}

	meta := &domain.PostWithMeta{Post: *p}
	return meta, atts, nil
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
	return posts, nil
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

// ListAttachments returns attachments for a single post (no auth check — caller must own the post).
func (s *Service) ListAttachments(ctx context.Context, postID uuid.UUID) ([]*domain.PostAttachment, error) {
	atts, err := s.repo.ListAttachments(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("feed.ListAttachments: %w", err)
	}
	return atts, nil
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
