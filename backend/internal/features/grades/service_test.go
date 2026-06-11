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
	mg, gid := newGroups(teacherID)
	mg.addMember(gid, studentID)
	return grades.NewService(newMemRepo(), mg), gid
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestCreateGrade_Success(t *testing.T) {
	svc, gid := buildSvc()
	g, err := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID,
		Subject:   "Математика",
		Value:     85,
		MaxValue:  100,
	})
	require.NoError(t, err)
	assert.Equal(t, float64(85), g.Value)
	assert.Equal(t, "Математика", g.Subject)
}

func TestCreateGrade_ForbiddenNonTeacher(t *testing.T) {
	svc, gid := buildSvc()
	_, err := svc.CreateGrade(context.Background(), uuid.New(), gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "X", Value: 10,
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateGrade_StudentNotMember(t *testing.T) {
	svc, gid := buildSvc()
	_, err := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: uuid.New(), Subject: "X", Value: 10,
	})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestCreateGrade_DefaultMaxValue(t *testing.T) {
	svc, gid := buildSvc()
	g, err := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "X", Value: 50,
	})
	require.NoError(t, err)
	assert.Equal(t, float64(100), g.MaxValue)
}

func TestListGroupGrades_TeacherSeesAll(t *testing.T) {
	svc, gid := buildSvc()
	student2 := uuid.New()
	mg, _ := newGroups(teacherID)
	mg.addMember(gid, student2)

	_, _ = svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "X", Value: 10,
	})
	_, _ = svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "Y", Value: 20,
	})

	list, err := svc.ListGroupGrades(context.Background(), gid, teacherID, domain.RoleTeacher)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestListGroupGrades_StudentSeesOnlyOwn(t *testing.T) {
	svc, gid := buildSvc()
	_, _ = svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "Math", Value: 90,
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
		StudentID: studentID, Subject: "Math", Value: 70,
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
		StudentID: studentID, Subject: "Math", Value: 70,
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
		StudentID: studentID, Subject: "Math", Value: 80,
	})
	err := svc.DeleteGrade(context.Background(), g.ID, teacherID)
	require.NoError(t, err)

	list, _ := svc.ListGroupGrades(context.Background(), gid, teacherID, domain.RoleTeacher)
	assert.Len(t, list, 0)
}

func TestDeleteGrade_ForbiddenNonTeacher(t *testing.T) {
	svc, gid := buildSvc()
	g, _ := svc.CreateGrade(context.Background(), teacherID, gid, grades.CreateGradeRequest{
		StudentID: studentID, Subject: "Math", Value: 80,
	})
	err := svc.DeleteGrade(context.Background(), g.ID, uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}
