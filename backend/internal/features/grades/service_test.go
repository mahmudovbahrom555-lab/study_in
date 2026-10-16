package grades_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/grades"
)

// ─── in-memory repo ───────────────────────────────────────────────────────────

type memRepo struct {
	grades map[uuid.UUID]*domain.Grade
}

func newMemRepo() *memRepo { return &memRepo{grades: make(map[uuid.UUID]*domain.Grade)} }

func (r *memRepo) Create(_ context.Context, g *domain.Grade) error {
	cp := *g; r.grades[g.ID] = &cp; return nil
}
func (r *memRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Grade, error) {
	g, ok := r.grades[id]; if !ok { return nil, nil }; cp := *g; return &cp, nil
}
func (r *memRepo) ListByGroupStudent(_ context.Context, gid, sid uuid.UUID) ([]*domain.Grade, error) {
	var out []*domain.Grade
	for _, g := range r.grades {
		if g.GroupID == gid && g.StudentID == sid { cp := *g; out = append(out, &cp) }
	}
	return out, nil
}
func (r *memRepo) ListByGroup(_ context.Context, gid uuid.UUID) ([]*domain.Grade, error) {
	var out []*domain.Grade
	for _, g := range r.grades { if g.GroupID == gid { cp := *g; out = append(out, &cp) } }
	return out, nil
}
func (r *memRepo) Update(_ context.Context, g *domain.Grade) error {
	cp := *g; r.grades[g.ID] = &cp; return nil
}
func (r *memRepo) Delete(_ context.Context, id uuid.UUID) error { delete(r.grades, id); return nil }

// ─── group checker ────────────────────────────────────────────────────────────

type memGroups struct {
	groups  map[uuid.UUID]*domain.Group
	members map[string]struct{}
}

func newGroups(teacherID uuid.UUID) (*memGroups, uuid.UUID) {
	gid := uuid.New()
	return &memGroups{
		groups:  map[uuid.UUID]*domain.Group{gid: {ID: gid, TeacherID: teacherID}},
		members: make(map[string]struct{}),
	}, gid
}
func (mg *memGroups) addMember(gid, sid uuid.UUID) {
	mg.members[gid.String()+":"+sid.String()] = struct{}{}
}
func (mg *memGroups) GetGroupByID(_ context.Context, id uuid.UUID) (*domain.Group, error) {
	g, ok := mg.groups[id]; if !ok { return nil, nil }; cp := *g; return &cp, nil
}
func (mg *memGroups) GetMember(_ context.Context, gid, uid uuid.UUID) (*domain.GroupMember, error) {
	_, ok := mg.members[gid.String()+":"+uid.String()]
	if !ok { return nil, nil }
	return &domain.GroupMember{GroupID: gid, StudentID: uid}, nil
}

// ─── fixtures ────────────────────────────────────────────────────────────────

var (
	teacherID = uuid.New()
	studentID = uuid.New()
)

func buildSvc() (*grades.Service, uuid.UUID) {
	svc, gid, _ := buildSvcRepo()
	return svc, gid
}

func buildSvcRepo() (*grades.Service, uuid.UUID, *memRepo) {
	mg, gid := newGroups(teacherID)
	mg.addMember(gid, studentID)
	repo := newMemRepo()
	return grades.NewService(repo, mg), gid, repo
}

func v(f float64) *float64 { return &f }

func assertCode(t *testing.T, err error, code string) {
	t.Helper()
	var de *domain.Error
	require.ErrorAs(t, err, &de)
	assert.Equal(t, code, de.Code)
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestCreateGrade_Success(t *testing.T) {
	svc, gid := buildSvc()
	g, err := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID,
		Subject:   "Математика",
		Value: v(85),
		MaxValue: 100,
	})
	require.NoError(t, err)
	assert.Equal(t, float64(85), g.Value)
	assert.Equal(t, "Математика", g.Subject)
}

func TestCreateGrade_ForbiddenNonTeacher(t *testing.T) {
	svc, gid := buildSvc()
	_, err := svc.CreateGrade(context.Background(), uuid.New(), gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "X", Value: v(10),
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateGrade_StudentNotMember(t *testing.T) {
	svc, gid := buildSvc()
	_, err := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: uuid.New(), Subject: "X", Value: v(10),
	})
	assertCode(t, err, "STUDENT_NOT_IN_GROUP")
}

func TestCreateGrade_DefaultMaxValue(t *testing.T) {
	svc, gid := buildSvc()
	g, err := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "X", Value: v(50),
	})
	require.NoError(t, err)
	assert.Equal(t, float64(100), g.MaxValue)
}

func TestCreateGrade_ZeroAllowed(t *testing.T) {
	svc, gid := buildSvc()
	g, err := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "Диктант", Value: v(0), MaxValue: 5,
	})
	require.NoError(t, err)
	assert.Equal(t, float64(0), g.Value)
	assert.Equal(t, domain.GradeKindClass, g.Kind)
}

func TestCreateGrade_AboveMaxRejected(t *testing.T) {
	svc, gid := buildSvc()
	_, err := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "Диктант", Value: v(6), MaxValue: 5,
	})
	assertCode(t, err, "GRADE_ABOVE_MAX")
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestUpdateGrade_AboveMaxRejected(t *testing.T) {
	svc, gid := buildSvc()
	g, _ := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "X", Value: v(4), MaxValue: 5,
	})
	_, err := svc.UpdateGrade(context.Background(), g.ID, teacherID, grades.UpdateGradeRequest{Value: v(7)})
	assertCode(t, err, "GRADE_ABOVE_MAX")
}

func TestHomeworkGrade_ReadonlyInJournal(t *testing.T) {
	svc, gid, repo := buildSvcRepo()
	aid := uuid.New()
	hw := &domain.Grade{
		ID: uuid.New(), GroupID: gid, StudentID: studentID, TeacherID: teacherID,
		Subject: "Workbook", Value: 90, MaxValue: 100,
		Kind: domain.GradeKindHomework, AssignmentID: &aid,
	}
	require.NoError(t, repo.Create(context.Background(), hw))

	assertCode(t, svc.DeleteGrade(context.Background(), hw.ID, teacherID), "HOMEWORK_GRADE_READONLY")
	_, err := svc.UpdateGrade(context.Background(), hw.ID, teacherID, grades.UpdateGradeRequest{Value: v(50)})
	assertCode(t, err, "HOMEWORK_GRADE_READONLY")

	list, _ := svc.ListGroupGrades(context.Background(), gid, teacherID, domain.RoleTeacher)
	require.Len(t, list, 1)
	assert.Equal(t, float64(90), list[0].Value, "оценка за ДЗ не изменилась")
}

func TestListGroupGrades_TeacherSeesAll(t *testing.T) {
	svc, gid := buildSvc()
	student2 := uuid.New()
	mg, _ := newGroups(teacherID)
	mg.addMember(gid, student2)

	_, _ = svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "X", Value: v(10),
	})
	_, _ = svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "Y", Value: v(20),
	})

	list, err := svc.ListGroupGrades(context.Background(), gid, teacherID, domain.RoleTeacher)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestListGroupGrades_StudentSeesOnlyOwn(t *testing.T) {
	svc, gid := buildSvc()
	_, _ = svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "Math", Value: v(90),
	})

	list, err := svc.ListGroupGrades(context.Background(), gid, studentID, domain.RoleStudent)
	require.NoError(t, err)
	for _, g := range list {
		assert.Equal(t, studentID, g.StudentID)
	}
}

func TestUpdateGrade_TeacherOnly(t *testing.T) {
	svc, gid := buildSvc()
	g, _ := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "Math", Value: v(70),
	})

	newVal := float64(95)
	updated, err := svc.UpdateGrade(context.Background(), g.ID, teacherID, grades.UpdateGradeRequest{
		Value: &newVal,
	})
	require.NoError(t, err)
	assert.Equal(t, float64(95), updated.Value)
}

func TestUpdateGrade_ForbiddenNonTeacher(t *testing.T) {
	svc, gid := buildSvc()
	g, _ := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "Math", Value: v(70),
	})

	val := float64(50)
	_, err := svc.UpdateGrade(context.Background(), g.ID, uuid.New(), grades.UpdateGradeRequest{
		Value: &val,
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestDeleteGrade_TeacherOnly(t *testing.T) {
	svc, gid := buildSvc()
	g, _ := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "Math", Value: v(80),
	})
	err := svc.DeleteGrade(context.Background(), g.ID, teacherID)
	require.NoError(t, err)

	list, _ := svc.ListGroupGrades(context.Background(), gid, teacherID, domain.RoleTeacher)
	assert.Len(t, list, 0)
}

func TestDeleteGrade_ForbiddenNonTeacher(t *testing.T) {
	svc, gid := buildSvc()
	g, _ := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "Math", Value: v(80),
	})
	err := svc.DeleteGrade(context.Background(), g.ID, uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}
