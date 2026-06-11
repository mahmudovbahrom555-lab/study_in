package feed_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/feed"
)

// --- in-memory feed repo ---

type memFeedRepo struct {
	posts map[uuid.UUID]*domain.Post
	atts  map[uuid.UUID][]*domain.PostAttachment // postID → attachments
}

func newMemFeedRepo() *memFeedRepo {
	return &memFeedRepo{
		posts: make(map[uuid.UUID]*domain.Post),
		atts:  make(map[uuid.UUID][]*domain.PostAttachment),
	}
}

func (r *memFeedRepo) CreatePost(_ context.Context, p *domain.Post) error {
	cp := *p
	r.posts[p.ID] = &cp
	return nil
}

func (r *memFeedRepo) GetPostByID(_ context.Context, id uuid.UUID) (*domain.Post, error) {
	p, ok := r.posts[id]
	if !ok || p.DeletedAt != nil {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (r *memFeedRepo) ListGroupPosts(_ context.Context, groupID uuid.UUID, limit, offset int) ([]*domain.PostWithMeta, error) {
	var all []*domain.PostWithMeta
	for _, p := range r.posts {
		if p.GroupID == groupID && p.DeletedAt == nil {
			cp := *p
			all = append(all, &domain.PostWithMeta{Post: cp, AuthorName: "Teacher"})
		}
	}
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (r *memFeedRepo) UpdatePost(_ context.Context, p *domain.Post) error {
	if _, ok := r.posts[p.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *p
	r.posts[p.ID] = &cp
	return nil
}

func (r *memFeedRepo) SoftDeletePost(_ context.Context, id uuid.UUID) error {
	if p, ok := r.posts[id]; ok {
		now := time.Now()
		p.DeletedAt = &now
	}
	return nil
}

func (r *memFeedRepo) PinPost(_ context.Context, id uuid.UUID, pinned bool) error {
	if p, ok := r.posts[id]; ok {
		p.Pinned = pinned
	}
	return nil
}

func (r *memFeedRepo) AddAttachment(_ context.Context, a *domain.PostAttachment) error {
	cp := *a
	r.atts[a.PostID] = append(r.atts[a.PostID], &cp)
	return nil
}

func (r *memFeedRepo) ListAttachments(_ context.Context, postID uuid.UUID) ([]*domain.PostAttachment, error) {
	return r.atts[postID], nil
}

// --- in-memory group checker ---

type memGroupChecker struct {
	groups  map[uuid.UUID]*domain.Group
	members map[string]*domain.GroupMember
}

func newGroupChecker(teacherID uuid.UUID) (*memGroupChecker, uuid.UUID) {
	gc := &memGroupChecker{
		groups:  make(map[uuid.UUID]*domain.Group),
		members: make(map[string]*domain.GroupMember),
	}
	gid := uuid.New()
	gc.groups[gid] = &domain.Group{
		ID:        gid,
		TeacherID: teacherID,
		Name:      "Test Group",
	}
	return gc, gid
}

func (gc *memGroupChecker) addMember(groupID, studentID uuid.UUID) {
	gc.members[groupID.String()+":"+studentID.String()] = &domain.GroupMember{
		GroupID: groupID, StudentID: studentID,
	}
}

func (gc *memGroupChecker) GetGroupByID(_ context.Context, id uuid.UUID) (*domain.Group, error) {
	g, ok := gc.groups[id]
	if !ok {
		return nil, nil
	}
	cp := *g
	return &cp, nil
}

func (gc *memGroupChecker) GetMember(_ context.Context, groupID, userID uuid.UUID) (*domain.GroupMember, error) {
	m, ok := gc.members[groupID.String()+":"+userID.String()]
	if !ok {
		return nil, nil
	}
	cp := *m
	return &cp, nil
}

// --- helpers ---

var (
	teacherID = uuid.New()
	studentID = uuid.New()
)

func buildSvc() (*feed.Service, *memGroupChecker, uuid.UUID) {
	gc, gid := newGroupChecker(teacherID)
	gc.addMember(gid, studentID)
	repo := newMemFeedRepo()
	svc := feed.NewService(repo, gc)
	return svc, gc, gid
}

// --- tests ---

func TestCreatePost_Success(t *testing.T) {
	svc, _, gid := buildSvc()
	p, err := svc.CreatePost(context.Background(), teacherID, gid, feed.CreatePostRequest{Body: "Hello!"})
	require.NoError(t, err)
	assert.Equal(t, "Hello!", p.Body)
	assert.Equal(t, gid, p.GroupID)
}

func TestCreatePost_NonTeacherForbidden(t *testing.T) {
	svc, _, gid := buildSvc()
	_, err := svc.CreatePost(context.Background(), studentID, gid, feed.CreatePostRequest{Body: "Hi"})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreatePost_UnknownGroup(t *testing.T) {
	svc, _, _ := buildSvc()
	_, err := svc.CreatePost(context.Background(), teacherID, uuid.New(), feed.CreatePostRequest{Body: "x"})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestListPosts_TeacherSeesOwn(t *testing.T) {
	svc, _, gid := buildSvc()
	_, _ = svc.CreatePost(context.Background(), teacherID, gid, feed.CreatePostRequest{Body: "A"})
	_, _ = svc.CreatePost(context.Background(), teacherID, gid, feed.CreatePostRequest{Body: "B"})

	posts, err := svc.ListPosts(context.Background(), gid, teacherID, domain.RoleTeacher, 10, 0)
	require.NoError(t, err)
	assert.Len(t, posts, 2)
}

func TestListPosts_StudentSeesGroup(t *testing.T) {
	svc, _, gid := buildSvc()
	_, _ = svc.CreatePost(context.Background(), teacherID, gid, feed.CreatePostRequest{Body: "A"})

	posts, err := svc.ListPosts(context.Background(), gid, studentID, domain.RoleStudent, 10, 0)
	require.NoError(t, err)
	assert.Len(t, posts, 1)
}

func TestListPosts_NonMemberForbidden(t *testing.T) {
	svc, _, gid := buildSvc()
	stranger := uuid.New()
	_, err := svc.ListPosts(context.Background(), gid, stranger, domain.RoleStudent, 10, 0)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpdatePost_Success(t *testing.T) {
	svc, _, gid := buildSvc()
	p, _ := svc.CreatePost(context.Background(), teacherID, gid, feed.CreatePostRequest{Body: "Original"})

	newBody := "Updated"
	updated, err := svc.UpdatePost(context.Background(), p.ID, teacherID, feed.UpdatePostRequest{Body: &newBody})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Body)
}

func TestUpdatePost_NonAuthorForbidden(t *testing.T) {
	svc, _, gid := buildSvc()
	p, _ := svc.CreatePost(context.Background(), teacherID, gid, feed.CreatePostRequest{Body: "X"})

	other := uuid.New()
	body := "hack"
	_, err := svc.UpdatePost(context.Background(), p.ID, other, feed.UpdatePostRequest{Body: &body})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestDeletePost_Success(t *testing.T) {
	svc, _, gid := buildSvc()
	p, _ := svc.CreatePost(context.Background(), teacherID, gid, feed.CreatePostRequest{Body: "bye"})

	err := svc.DeletePost(context.Background(), p.ID, teacherID)
	require.NoError(t, err)

	// post should be gone
	posts, _ := svc.ListPosts(context.Background(), gid, teacherID, domain.RoleTeacher, 10, 0)
	assert.Len(t, posts, 0)
}

func TestPinPost_TeacherOnly(t *testing.T) {
	svc, _, gid := buildSvc()
	p, _ := svc.CreatePost(context.Background(), teacherID, gid, feed.CreatePostRequest{Body: "pin me"})

	updated, err := svc.PinPost(context.Background(), p.ID, teacherID, true)
	require.NoError(t, err)
	assert.True(t, updated.Pinned)
}

func TestPinPost_NonTeacherForbidden(t *testing.T) {
	svc, _, gid := buildSvc()
	p, _ := svc.CreatePost(context.Background(), teacherID, gid, feed.CreatePostRequest{Body: "x"})

	_, err := svc.PinPost(context.Background(), p.ID, studentID, true)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}
