package assignments_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/assignments"
)

// --- in-memory repo ---

type memRepo struct {
	assignments map[uuid.UUID]*domain.Assignment
	aAtts       map[uuid.UUID][]*domain.AssignmentAttachment
	submissions map[string]*domain.Submission // "assignmentID:studentID"
	sAtts       map[uuid.UUID][]*domain.SubmissionAttachment
}

func newMemRepo() *memRepo {
	return &memRepo{
		assignments: make(map[uuid.UUID]*domain.Assignment),
		aAtts:       make(map[uuid.UUID][]*domain.AssignmentAttachment),
		submissions: make(map[string]*domain.Submission),
		sAtts:       make(map[uuid.UUID][]*domain.SubmissionAttachment),
	}
}

func subKey(aid, sid uuid.UUID) string { return aid.String() + ":" + sid.String() }

func (r *memRepo) CreateAssignment(_ context.Context, a *domain.Assignment) error {
	cp := *a; r.assignments[a.ID] = &cp; return nil
}
func (r *memRepo) GetAssignmentByID(_ context.Context, id uuid.UUID) (*domain.Assignment, error) {
	a, ok := r.assignments[id]
	if !ok || a.DeletedAt != nil { return nil, nil }
	cp := *a; return &cp, nil
}
func (r *memRepo) ListGroupAssignments(_ context.Context, groupID uuid.UUID) ([]*domain.Assignment, error) {
	var out []*domain.Assignment
	for _, a := range r.assignments {
		if a.GroupID == groupID && a.DeletedAt == nil { cp := *a; out = append(out, &cp) }
	}
	return out, nil
}
func (r *memRepo) UpdateAssignment(_ context.Context, a *domain.Assignment) error {
	cp := *a; r.assignments[a.ID] = &cp; return nil
}
func (r *memRepo) SoftDeleteAssignment(_ context.Context, id uuid.UUID) error {
	if a, ok := r.assignments[id]; ok { now := time.Now(); a.DeletedAt = &now }; return nil
}
func (r *memRepo) AddAssignmentAttachment(_ context.Context, a *domain.AssignmentAttachment) error {
	cp := *a; r.aAtts[a.AssignmentID] = append(r.aAtts[a.AssignmentID], &cp); return nil
}
func (r *memRepo) ListAssignmentAttachments(_ context.Context, id uuid.UUID) ([]*domain.AssignmentAttachment, error) {
	return r.aAtts[id], nil
}
func (r *memRepo) CreateSubmission(_ context.Context, s *domain.Submission) error {
	cp := *s; r.submissions[subKey(s.AssignmentID, s.StudentID)] = &cp; return nil
}
func (r *memRepo) GetSubmission(_ context.Context, aid, sid uuid.UUID) (*domain.Submission, error) {
	s, ok := r.submissions[subKey(aid, sid)]
	if !ok { return nil, nil }; cp := *s; return &cp, nil
}
func (r *memRepo) GetSubmissionByID(_ context.Context, id uuid.UUID) (*domain.Submission, error) {
	for _, s := range r.submissions { if s.ID == id { cp := *s; return &cp, nil } }
	return nil, nil
}
func (r *memRepo) ListSubmissions(_ context.Context, aid uuid.UUID) ([]*domain.Submission, error) {
	var out []*domain.Submission
	for _, s := range r.submissions { if s.AssignmentID == aid { cp := *s; out = append(out, &cp) } }
	return out, nil
}
func (r *memRepo) GradeSubmission(_ context.Context, id uuid.UUID, grade int16, note string) error {
	for _, s := range r.submissions {
		if s.ID == id { now := time.Now(); s.Grade = &grade; s.TeacherNote = &note; s.GradedAt = &now }
	}
	return nil
}
func (r *memRepo) AddSubmissionAttachment(_ context.Context, a *domain.SubmissionAttachment) error {
	cp := *a; r.sAtts[a.SubmissionID] = append(r.sAtts[a.SubmissionID], &cp); return nil
}
func (r *memRepo) ListSubmissionAttachments(_ context.Context, id uuid.UUID) ([]*domain.SubmissionAttachment, error) {
	return r.sAtts[id], nil
}

// --- group checker ---

type memGroups struct {
	groups  map[uuid.UUID]*domain.Group
	members map[string]*domain.GroupMember
}

func newGroups(teacherID uuid.UUID) (*memGroups, uuid.UUID) {
	gid := uuid.New()
	mg := &memGroups{
		groups:  map[uuid.UUID]*domain.Group{gid: {ID: gid, TeacherID: teacherID}},
		members: make(map[string]*domain.GroupMember),
	}
	return mg, gid
}
func (mg *memGroups) addMember(gid, sid uuid.UUID) {
	mg.members[gid.String()+":"+sid.String()] = &domain.GroupMember{GroupID: gid, StudentID: sid}
}
func (mg *memGroups) GetGroupByID(_ context.Context, id uuid.UUID) (*domain.Group, error) {
	g, ok := mg.groups[id]; if !ok { return nil, nil }; cp := *g; return &cp, nil
}
func (mg *memGroups) GetMember(_ context.Context, gid, uid uuid.UUID) (*domain.GroupMember, error) {
	m, ok := mg.members[gid.String()+":"+uid.String()]; if !ok { return nil, nil }; cp := *m; return &cp, nil
}

// --- noop store/signer ---

type noopStore struct{}
func (noopStore) PutObject(_ context.Context, _ string, _ io.Reader, _ int64, _ string) error { return nil }

type noopSigner struct{}
func (noopSigner) PresignedGetURL(_ context.Context, key string) (string, error) { return key, nil }

// --- helpers ---

var teacherID = uuid.New()
var studentID = uuid.New()

func buildSvc() (*assignments.Service, *memGroups, uuid.UUID) {
	mg, gid := newGroups(teacherID)
	mg.addMember(gid, studentID)
	return assignments.NewService(newMemRepo(), mg, noopSigner{}, noopStore{}), mg, gid
}

// --- tests ---

func TestCreateAssignment_Success(t *testing.T) {
	svc, _, gid := buildSvc()
	a, err := svc.CreateAssignment(context.Background(), teacherID, gid,
		assignments.CreateAssignmentRequest{Title: "Задача 1"})
	require.NoError(t, err)
	assert.Equal(t, "Задача 1", a.Title)
}

func TestCreateAssignment_ForbiddenForNonOwner(t *testing.T) {
	svc, _, gid := buildSvc()
	other := uuid.New()
	_, err := svc.CreateAssignment(context.Background(), other, gid,
		assignments.CreateAssignmentRequest{Title: "hack"})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestListAssignments_StudentSeesGroup(t *testing.T) {
	svc, _, gid := buildSvc()
	_, _ = svc.CreateAssignment(context.Background(), teacherID, gid,
		assignments.CreateAssignmentRequest{Title: "A"})
	as, err := svc.ListAssignments(context.Background(), gid, studentID, domain.RoleStudent)
	require.NoError(t, err)
	assert.Len(t, as, 1)
}

func TestListAssignments_NonMemberForbidden(t *testing.T) {
	svc, _, gid := buildSvc()
	_, err := svc.ListAssignments(context.Background(), gid, uuid.New(), domain.RoleStudent)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestSubmit_Success(t *testing.T) {
	svc, _, gid := buildSvc()
	a, _ := svc.CreateAssignment(context.Background(), teacherID, gid,
		assignments.CreateAssignmentRequest{Title: "B"})
	sub, err := svc.Submit(context.Background(), a.ID, studentID, assignments.SubmitRequest{})
	require.NoError(t, err)
	assert.Equal(t, a.ID, sub.AssignmentID)
}

func TestSubmit_NonMemberForbidden(t *testing.T) {
	svc, _, gid := buildSvc()
	a, _ := svc.CreateAssignment(context.Background(), teacherID, gid,
		assignments.CreateAssignmentRequest{Title: "C"})
	_, err := svc.Submit(context.Background(), a.ID, uuid.New(), assignments.SubmitRequest{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGrade_Success(t *testing.T) {
	svc, _, gid := buildSvc()
	a, _ := svc.CreateAssignment(context.Background(), teacherID, gid,
		assignments.CreateAssignmentRequest{Title: "D"})
	sub, _ := svc.Submit(context.Background(), a.ID, studentID, assignments.SubmitRequest{})

	note := "Хорошо"
	graded, err := svc.Grade(context.Background(), sub.ID, teacherID,
		assignments.GradeRequest{Grade: 90, Note: &note})
	require.NoError(t, err)
	require.NotNil(t, graded.Grade)
	assert.Equal(t, int16(90), *graded.Grade)
}

func TestGrade_WrongTeacherForbidden(t *testing.T) {
	svc, _, gid := buildSvc()
	a, _ := svc.CreateAssignment(context.Background(), teacherID, gid,
		assignments.CreateAssignmentRequest{Title: "E"})
	sub, _ := svc.Submit(context.Background(), a.ID, studentID, assignments.SubmitRequest{})

	_, err := svc.Grade(context.Background(), sub.ID, uuid.New(),
		assignments.GradeRequest{Grade: 50})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUploadAssignmentFile_SizeLimit(t *testing.T) {
	svc, _, gid := buildSvc()
	a, _ := svc.CreateAssignment(context.Background(), teacherID, gid,
		assignments.CreateAssignmentRequest{Title: "F"})

	bigSize := int64(51 << 20) // 51 MB > maxFileSize
	_, err := svc.UploadAssignmentFile(context.Background(), a.ID, teacherID,
		"big.pdf", bytes.NewReader(nil), bigSize)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestDeleteAssignment_Success(t *testing.T) {
	svc, _, gid := buildSvc()
	a, _ := svc.CreateAssignment(context.Background(), teacherID, gid,
		assignments.CreateAssignmentRequest{Title: "G"})
	err := svc.DeleteAssignment(context.Background(), a.ID, teacherID)
	require.NoError(t, err)

	as, _ := svc.ListAssignments(context.Background(), gid, teacherID, domain.RoleTeacher)
	assert.Len(t, as, 0)
}
