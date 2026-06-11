package attendance_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/attendance"
)

// ─── in-memory repo ───────────────────────────────────────────────────────────

type memRepo struct {
	records map[uuid.UUID]*domain.Attendance
}

func newMemRepo() *memRepo { return &memRepo{records: make(map[uuid.UUID]*domain.Attendance)} }

func (r *memRepo) Upsert(_ context.Context, a *domain.Attendance) error {
	// Simulate ON CONFLICT (group_id, student_id, lesson_date) DO UPDATE
	for _, existing := range r.records {
		if existing.GroupID == a.GroupID &&
			existing.StudentID == a.StudentID &&
			existing.LessonDate.Equal(a.LessonDate) {
			existing.Status = a.Status
			existing.Note = a.Note
			existing.UpdatedAt = time.Now()
			return nil
		}
	}
	cp := *a
	r.records[a.ID] = &cp
	return nil
}
func (r *memRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Attendance, error) {
	a, ok := r.records[id]; if !ok { return nil, nil }; cp := *a; return &cp, nil
}
func (r *memRepo) ListByGroupDate(_ context.Context, gid uuid.UUID, date time.Time) ([]*domain.Attendance, error) {
	var out []*domain.Attendance
	for _, a := range r.records {
		if a.GroupID == gid && a.LessonDate.Equal(date) { cp := *a; out = append(out, &cp) }
	}
	return out, nil
}
func (r *memRepo) ListByGroupStudent(_ context.Context, gid, sid uuid.UUID) ([]*domain.Attendance, error) {
	var out []*domain.Attendance
	for _, a := range r.records {
		if a.GroupID == gid && a.StudentID == sid { cp := *a; out = append(out, &cp) }
	}
	return out, nil
}
func (r *memRepo) Delete(_ context.Context, id uuid.UUID) error { delete(r.records, id); return nil }

// ─── group checker ────────────────────────────────────────────────────────────

type memGroups struct {
	groups  map[uuid.UUID]*domain.Group
	members map[string]struct{}
}

func newGroups(tid uuid.UUID) (*memGroups, uuid.UUID) {
	gid := uuid.New()
	return &memGroups{
		groups:  map[uuid.UUID]*domain.Group{gid: {ID: gid, TeacherID: tid}},
		members: make(map[string]struct{}),
	}, gid
}
func (mg *memGroups) add(gid, sid uuid.UUID) { mg.members[gid.String()+":"+sid.String()] = struct{}{} }
func (mg *memGroups) GetGroupByID(_ context.Context, id uuid.UUID) (*domain.Group, error) {
	g, ok := mg.groups[id]; if !ok { return nil, nil }; cp := *g; return &cp, nil
}
func (mg *memGroups) GetMember(_ context.Context, gid, uid uuid.UUID) (*domain.GroupMember, error) {
	if _, ok := mg.members[gid.String()+":"+uid.String()]; !ok { return nil, nil }
	return &domain.GroupMember{GroupID: gid, StudentID: uid}, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

var (
	teacherID = uuid.New()
	studentID = uuid.New()
)

func buildSvc() (*attendance.Service, uuid.UUID) {
	mg, gid := newGroups(teacherID)
	mg.add(gid, studentID)
	return attendance.NewService(newMemRepo(), mg), gid
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestMark_Success(t *testing.T) {
	svc, gid := buildSvc()
	a, err := svc.Mark(context.Background(), teacherID, gid, attendance.MarkRequest{
		StudentID:  studentID,
		LessonDate: "2024-09-01",
		Status:     domain.AttendancePresent,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.AttendancePresent, a.Status)
}

func TestMark_ForbiddenNonTeacher(t *testing.T) {
	svc, gid := buildSvc()
	_, err := svc.Mark(context.Background(), uuid.New(), gid, attendance.MarkRequest{
		StudentID:  studentID,
		LessonDate: "2024-09-01",
		Status:     domain.AttendancePresent,
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestMark_StudentNotMember(t *testing.T) {
	svc, gid := buildSvc()
	_, err := svc.Mark(context.Background(), teacherID, gid, attendance.MarkRequest{
		StudentID:  uuid.New(),
		LessonDate: "2024-09-01",
		Status:     domain.AttendancePresent,
	})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestMark_Upsert(t *testing.T) {
	svc, gid := buildSvc()
	_, _ = svc.Mark(context.Background(), teacherID, gid, attendance.MarkRequest{
		StudentID:  studentID,
		LessonDate: "2024-09-01",
		Status:     domain.AttendancePresent,
	})
	a2, err := svc.Mark(context.Background(), teacherID, gid, attendance.MarkRequest{
		StudentID:  studentID,
		LessonDate: "2024-09-01",
		Status:     domain.AttendanceAbsent,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.AttendanceAbsent, a2.Status)
}

func TestListByDate_TeacherOnly(t *testing.T) {
	svc, gid := buildSvc()
	_, _ = svc.Mark(context.Background(), teacherID, gid, attendance.MarkRequest{
		StudentID:  studentID,
		LessonDate: "2024-09-02",
		Status:     domain.AttendanceLate,
	})
	list, err := svc.ListByDate(context.Background(), gid, "2024-09-02", teacherID)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestListByDate_ForbiddenNonTeacher(t *testing.T) {
	svc, gid := buildSvc()
	_, err := svc.ListByDate(context.Background(), gid, "2024-09-02", uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestListStudent_StudentSelf(t *testing.T) {
	svc, gid := buildSvc()
	_, _ = svc.Mark(context.Background(), teacherID, gid, attendance.MarkRequest{
		StudentID:  studentID,
		LessonDate: "2024-09-03",
		Status:     domain.AttendanceExcused,
	})
	list, err := svc.ListStudent(context.Background(), gid, studentID, studentID, domain.RoleStudent)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestListStudent_StudentForbiddenForOther(t *testing.T) {
	svc, gid := buildSvc()
	_, err := svc.ListStudent(context.Background(), gid, studentID, uuid.New(), domain.RoleStudent)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestDelete_TeacherOnly(t *testing.T) {
	svc, gid := buildSvc()
	a, _ := svc.Mark(context.Background(), teacherID, gid, attendance.MarkRequest{
		StudentID:  studentID,
		LessonDate: "2024-09-04",
		Status:     domain.AttendanceAbsent,
	})
	err := svc.Delete(context.Background(), a.ID, teacherID)
	require.NoError(t, err)

	list, _ := svc.ListByDate(context.Background(), gid, "2024-09-04", teacherID)
	assert.Len(t, list, 0)
}
