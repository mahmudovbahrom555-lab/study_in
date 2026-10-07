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
	files       map[uuid.UUID]*domain.File     // модуль files (memRepo служит и FileGetter)
	subFiles    map[uuid.UUID][]uuid.UUID      // submissionID → fileIDs по порядку
}

func newMemRepo() *memRepo {
	return &memRepo{
		assignments: make(map[uuid.UUID]*domain.Assignment),
		aAtts:       make(map[uuid.UUID][]*domain.AssignmentAttachment),
		submissions: make(map[string]*domain.Submission),
		files:       make(map[uuid.UUID]*domain.File),
		subFiles:    make(map[uuid.UUID][]uuid.UUID),
	}
}

// addFile кладёт в «модуль files» загруженный, но не прикреплённый файл.
func (r *memRepo) addFile(owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	r.files[id] = &domain.File{ID: id, OwnerID: owner, Purpose: domain.FilePurposeSubmission, Status: domain.FileStatusPending}
	return id
}

func (r *memRepo) GetByIDs(_ context.Context, ids []uuid.UUID) ([]*domain.File, error) {
	var out []*domain.File
	for _, id := range ids {
		if f, ok := r.files[id]; ok { cp := *f; out = append(out, &cp) }
	}
	return out, nil
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
// SaveSubmission повторяет семантику Postgres-реализации: прежние файлы → pending, новые → attached.
func (r *memRepo) SaveSubmission(_ context.Context, s *domain.Submission, fileIDs []uuid.UUID) error {
	cp := *s
	if old, ok := r.submissions[subKey(s.AssignmentID, s.StudentID)]; ok {
		cp.Grade, cp.TeacherNote, cp.GradedAt = old.Grade, old.TeacherNote, old.GradedAt
	}
	r.submissions[subKey(s.AssignmentID, s.StudentID)] = &cp
	for _, id := range r.subFiles[s.ID] { r.files[id].Status = domain.FileStatusPending }
	r.subFiles[s.ID] = append([]uuid.UUID(nil), fileIDs...)
	for _, id := range fileIDs { r.files[id].Status = domain.FileStatusAttached }
	return nil
}
func (r *memRepo) ListSubmissionFiles(_ context.Context, ids []uuid.UUID) (map[uuid.UUID][]*domain.File, error) {
	out := map[uuid.UUID][]*domain.File{}
	for _, sid := range ids {
		for _, fid := range r.subFiles[sid] { cp := *r.files[fid]; out[sid] = append(out[sid], &cp) }
	}
	return out, nil
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
	svc, mg, gid, _ := buildSvcWithRepo()
	return svc, mg, gid
}

func buildSvcWithRepo() (*assignments.Service, *memGroups, uuid.UUID, *memRepo) {
	mg, gid := newGroups(teacherID)
	mg.addMember(gid, studentID)
	repo := newMemRepo()
	return assignments.NewService(repo, mg, noopSigner{}, noopStore{}, repo), mg, gid, repo
}

var comment = "Сделал"

func withComment() assignments.SubmitRequest { return assignments.SubmitRequest{Comment: &comment} }

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
	sub, err := svc.Submit(context.Background(), a.ID, studentID, withComment())
	require.NoError(t, err)
	assert.Equal(t, a.ID, sub.AssignmentID)
}

func TestSubmit_NonMemberForbidden(t *testing.T) {
	svc, _, gid := buildSvc()
	a, _ := svc.CreateAssignment(context.Background(), teacherID, gid,
		assignments.CreateAssignmentRequest{Title: "C"})
	_, err := svc.Submit(context.Background(), a.ID, uuid.New(), withComment())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGrade_Success(t *testing.T) {
	svc, _, gid := buildSvc()
	a, _ := svc.CreateAssignment(context.Background(), teacherID, gid,
		assignments.CreateAssignmentRequest{Title: "D"})
	sub, _ := svc.Submit(context.Background(), a.ID, studentID, withComment())

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
	sub, _ := svc.Submit(context.Background(), a.ID, studentID, withComment())

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

// --- G10: сдача с файлами ---

func newAssignment(t *testing.T, svc *assignments.Service, gid uuid.UUID) *domain.Assignment {
	t.Helper()
	a, err := svc.CreateAssignment(context.Background(), teacherID, gid, assignments.CreateAssignmentRequest{Title: "ДЗ"})
	require.NoError(t, err)
	return a
}

func TestSubmit_EmptyRejected(t *testing.T) {
	svc, _, gid := buildSvc()
	a := newAssignment(t, svc, gid)
	_, err := svc.Submit(context.Background(), a.ID, studentID, assignments.SubmitRequest{})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestSubmit_WithFiles_AttachesThem(t *testing.T) {
	svc, _, gid, repo := buildSvcWithRepo()
	a := newAssignment(t, svc, gid)
	f1, f2 := repo.addFile(studentID), repo.addFile(studentID)

	sub, err := svc.Submit(context.Background(), a.ID, studentID, assignments.SubmitRequest{FileIDs: []uuid.UUID{f1, f2}})
	require.NoError(t, err)

	files, _ := svc.SubmissionFiles(context.Background(), []uuid.UUID{sub.ID})
	require.Len(t, files[sub.ID], 2)
	assert.Equal(t, f1, files[sub.ID][0].ID, "порядок файлов сохраняется")
	assert.Equal(t, domain.FileStatusAttached, repo.files[f1].Status)
}

func TestSubmit_TooManyFiles(t *testing.T) {
	svc, _, gid, repo := buildSvcWithRepo()
	a := newAssignment(t, svc, gid)
	ids := make([]uuid.UUID, assignments.MaxSubmissionFiles+1)
	for i := range ids {
		ids[i] = repo.addFile(studentID)
	}
	_, err := svc.Submit(context.Background(), a.ID, studentID, assignments.SubmitRequest{FileIDs: ids})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestSubmit_ForeignFileForbidden(t *testing.T) {
	svc, _, gid, repo := buildSvcWithRepo()
	a := newAssignment(t, svc, gid)
	someoneElses := repo.addFile(uuid.New())
	_, err := svc.Submit(context.Background(), a.ID, studentID, assignments.SubmitRequest{FileIDs: []uuid.UUID{someoneElses}})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestSubmit_UnknownFile(t *testing.T) {
	svc, _, gid := buildSvc()
	a := newAssignment(t, svc, gid)
	_, err := svc.Submit(context.Background(), a.ID, studentID, assignments.SubmitRequest{FileIDs: []uuid.UUID{uuid.New()}})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestSubmit_FileOfAnotherSubmissionInUse(t *testing.T) {
	svc, _, gid, repo := buildSvcWithRepo()
	a1, a2 := newAssignment(t, svc, gid), newAssignment(t, svc, gid)
	f := repo.addFile(studentID)
	_, err := svc.Submit(context.Background(), a1.ID, studentID, assignments.SubmitRequest{FileIDs: []uuid.UUID{f}})
	require.NoError(t, err)

	_, err = svc.Submit(context.Background(), a2.ID, studentID, assignments.SubmitRequest{FileIDs: []uuid.UUID{f}})
	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestResubmit_ReplacesFilesAndKeepsID(t *testing.T) {
	svc, _, gid, repo := buildSvcWithRepo()
	a := newAssignment(t, svc, gid)
	keep, drop, added := repo.addFile(studentID), repo.addFile(studentID), repo.addFile(studentID)

	first, err := svc.Submit(context.Background(), a.ID, studentID, assignments.SubmitRequest{FileIDs: []uuid.UUID{keep, drop}})
	require.NoError(t, err)
	second, err := svc.Submit(context.Background(), a.ID, studentID, assignments.SubmitRequest{FileIDs: []uuid.UUID{keep, added}})
	require.NoError(t, err)

	assert.Equal(t, first.ID, second.ID, "пересдача обновляет ту же запись")
	assert.Equal(t, domain.FileStatusAttached, repo.files[keep].Status)
	assert.Equal(t, domain.FileStatusPending, repo.files[drop].Status, "убранный файл уйдёт в очистку")
	assert.Equal(t, domain.FileStatusAttached, repo.files[added].Status)
}

func TestResubmit_AfterGradeRejected(t *testing.T) {
	svc, _, gid := buildSvc()
	a := newAssignment(t, svc, gid)
	sub, err := svc.Submit(context.Background(), a.ID, studentID, withComment())
	require.NoError(t, err)
	_, err = svc.Grade(context.Background(), sub.ID, teacherID, assignments.GradeRequest{Grade: 80})
	require.NoError(t, err)

	_, err = svc.Submit(context.Background(), a.ID, studentID, withComment())
	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestMySubmission_ReturnsFiles(t *testing.T) {
	svc, _, gid, repo := buildSvcWithRepo()
	a := newAssignment(t, svc, gid)
	f := repo.addFile(studentID)
	_, err := svc.Submit(context.Background(), a.ID, studentID, assignments.SubmitRequest{FileIDs: []uuid.UUID{f}})
	require.NoError(t, err)

	sub, files, err := svc.MySubmission(context.Background(), a.ID, studentID)
	require.NoError(t, err)
	require.NotNil(t, sub)
	require.Len(t, files, 1)
	assert.Equal(t, f, files[0].ID)
}
