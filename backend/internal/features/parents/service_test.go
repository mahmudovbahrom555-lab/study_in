package parents_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/parents"
)

// ─── in-memory repo ───────────────────────────────────────────────────────────

type memRepo struct {
	links map[string]*domain.ParentLink
	codes map[string]*domain.ParentLinkCode
}

func newMemRepo() *memRepo {
	return &memRepo{
		links: make(map[string]*domain.ParentLink),
		codes: make(map[string]*domain.ParentLinkCode),
	}
}

func (r *memRepo) CreateLinkCode(_ context.Context, c *domain.ParentLinkCode) error {
	cp := *c
	r.codes[c.Code] = &cp
	return nil
}
func (r *memRepo) GetActiveLinkCode(_ context.Context, code string) (*domain.ParentLinkCode, error) {
	c, ok := r.codes[code]
	if !ok || time.Now().After(c.ExpiresAt) {
		return nil, nil
	}
	cp := *c
	return &cp, nil
}

func key(p, s uuid.UUID) string { return p.String() + ":" + s.String() }

func (r *memRepo) Link(_ context.Context, l *domain.ParentLink) error {
	cp := *l; r.links[key(l.ParentID, l.StudentID)] = &cp; return nil
}
func (r *memRepo) Unlink(_ context.Context, p, s uuid.UUID) error {
	delete(r.links, key(p, s)); return nil
}
func (r *memRepo) GetLink(_ context.Context, p, s uuid.UUID) (*domain.ParentLink, error) {
	l, ok := r.links[key(p, s)]; if !ok { return nil, nil }; cp := *l; return &cp, nil
}
func (r *memRepo) ListChildren(_ context.Context, p uuid.UUID) ([]*domain.ParentLink, error) {
	var out []*domain.ParentLink
	for _, l := range r.links { if l.ParentID == p { cp := *l; out = append(out, &cp) } }
	return out, nil
}
func (r *memRepo) ListParents(_ context.Context, s uuid.UUID) ([]*domain.ParentLink, error) {
	var out []*domain.ParentLink
	for _, l := range r.links { if l.StudentID == s { cp := *l; out = append(out, &cp) } }
	return out, nil
}

// ─── stubs ────────────────────────────────────────────────────────────────────

type stubGrades struct{}

func (stubGrades) ListByGroupStudent(_ context.Context, _, _ uuid.UUID) ([]*domain.Grade, error) {
	return []*domain.Grade{{Value: 90, MaxValue: 100}}, nil
}

type stubAttendance struct{}

func (stubAttendance) ListByGroupStudent(_ context.Context, _, _ uuid.UUID) ([]*domain.Attendance, error) {
	return []*domain.Attendance{{Status: domain.AttendancePresent, LessonDate: time.Now()}}, nil
}

type stubGroups struct{}

func (stubGroups) ListGroupsByStudent(_ context.Context, _ uuid.UUID) ([]*domain.Group, error) {
	return []*domain.Group{{Name: "Math"}}, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

var (
	parentID  = uuid.New()
	studentID = uuid.New()
	groupID   = uuid.New()
)

func buildSvc() *parents.Service {
	return parents.NewService(newMemRepo(), stubGrades{}, stubAttendance{}, stubGroups{})
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestLinkChild_Success(t *testing.T) {
	svc := buildSvc()
	link, err := svc.LinkChild(context.Background(), parentID, studentID)
	require.NoError(t, err)
	assert.Equal(t, studentID, link.StudentID)
}

func TestLinkChild_DuplicateForbidden(t *testing.T) {
	svc := buildSvc()
	_, _ = svc.LinkChild(context.Background(), parentID, studentID)
	_, err := svc.LinkChild(context.Background(), parentID, studentID)
	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestUnlinkChild_Success(t *testing.T) {
	svc := buildSvc()
	_, _ = svc.LinkChild(context.Background(), parentID, studentID)
	err := svc.UnlinkChild(context.Background(), parentID, studentID)
	require.NoError(t, err)

	children, _ := svc.ListChildren(context.Background(), parentID)
	assert.Len(t, children, 0)
}

func TestUnlinkChild_NotFound(t *testing.T) {
	svc := buildSvc()
	err := svc.UnlinkChild(context.Background(), parentID, studentID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestChildGrades_RequiresLink(t *testing.T) {
	svc := buildSvc()
	_, err := svc.ChildGrades(context.Background(), parentID, studentID, groupID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestChildGrades_LinkedParent(t *testing.T) {
	svc := buildSvc()
	_, _ = svc.LinkChild(context.Background(), parentID, studentID)
	grades, err := svc.ChildGrades(context.Background(), parentID, studentID, groupID)
	require.NoError(t, err)
	assert.Len(t, grades, 1)
}

func TestChildAttendance_RequiresLink(t *testing.T) {
	svc := buildSvc()
	_, err := svc.ChildAttendance(context.Background(), parentID, studentID, groupID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestChildGroups_LinkedParent(t *testing.T) {
	svc := buildSvc()
	_, _ = svc.LinkChild(context.Background(), parentID, studentID)
	groups, err := svc.ChildGroups(context.Background(), parentID, studentID)
	require.NoError(t, err)
	assert.Len(t, groups, 1)
}

// ─── link codes ───────────────────────────────────────────────────────────────

func TestCreateLinkCode_Format(t *testing.T) {
	svc := buildSvc()
	c, err := svc.CreateLinkCode(context.Background(), studentID)
	require.NoError(t, err)
	assert.Len(t, c.Code, 6)
	assert.Regexp(t, `^[2-9A-HJ-NP-Z]{6}$`, c.Code, "no ambiguous 0/O/1/I")
	assert.Equal(t, studentID, c.StudentID)
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), c.ExpiresAt, time.Minute)
}

func TestLinkByCode_Success_NormalizesInput(t *testing.T) {
	svc := buildSvc()
	c, err := svc.CreateLinkCode(context.Background(), studentID)
	require.NoError(t, err)

	// Родитель может ввести код строчными буквами и с пробелами.
	typed := " " + c.Code[:3] + " " + strings.ToLower(c.Code[3:]) + " "
	link, err := svc.LinkByCode(context.Background(), parentID, typed)
	require.NoError(t, err)
	assert.Equal(t, studentID, link.StudentID)
	assert.Equal(t, parentID, link.ParentID)
}

func TestLinkByCode_InvalidCode(t *testing.T) {
	svc := buildSvc()
	_, err := svc.LinkByCode(context.Background(), parentID, "ZZZZZZ")
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestLinkByCode_ExpiredCode(t *testing.T) {
	repo := newMemRepo()
	svc := parents.NewService(repo, stubGrades{}, stubAttendance{}, stubGroups{})
	require.NoError(t, repo.CreateLinkCode(context.Background(), &domain.ParentLinkCode{
		Code: "ABC234", StudentID: studentID, ExpiresAt: time.Now().Add(-time.Minute),
	}))
	_, err := svc.LinkByCode(context.Background(), parentID, "ABC234")
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestLinkByCode_SecondParentCanUseSameCode(t *testing.T) {
	svc := buildSvc()
	c, err := svc.CreateLinkCode(context.Background(), studentID)
	require.NoError(t, err)
	_, err = svc.LinkByCode(context.Background(), parentID, c.Code)
	require.NoError(t, err)
	_, err = svc.LinkByCode(context.Background(), uuid.New(), c.Code)
	assert.NoError(t, err)
}
