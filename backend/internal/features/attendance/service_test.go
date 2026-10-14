package attendance_test

import (
	"context"
	"sort"
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
	groups  *memGroups // для Roster: состав группы и имена
}

func (r *memRepo) find(a *domain.Attendance) *domain.Attendance {
	for _, ex := range r.records {
		if ex.GroupID == a.GroupID && ex.StudentID == a.StudentID && ex.LessonDate.Equal(a.LessonDate) {
			return ex
		}
	}
	return nil
}

// Upsert повторяет ON CONFLICT … RETURNING id: при повторе — id существующей строки.
func (r *memRepo) Upsert(_ context.Context, a *domain.Attendance) error {
	if ex := r.find(a); ex != nil {
		ex.Status, ex.Note, ex.UpdatedAt = a.Status, a.Note, time.Now()
		a.ID, a.CreatedAt = ex.ID, ex.CreatedAt
		return nil
	}
	cp := *a
	r.records[a.ID] = &cp
	return nil
}
func (r *memRepo) UpsertMany(ctx context.Context, list []*domain.Attendance) error {
	for _, a := range list {
		_ = r.Upsert(ctx, a)
	}
	return nil
}
func (r *memRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Attendance, error) {
	a, ok := r.records[id]
	if !ok {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}
func (r *memRepo) Roster(_ context.Context, gid uuid.UUID, date time.Time) ([]*domain.RosterEntry, error) {
	ids := map[uuid.UUID]bool{}
	for key := range r.groups.members {
		if key.gid == gid {
			ids[key.sid] = true
		}
	}
	for _, a := range r.records {
		if a.GroupID == gid && a.LessonDate.Equal(date) {
			ids[a.StudentID] = true
		}
	}
	var out []*domain.RosterEntry
	for sid := range ids {
		e := &domain.RosterEntry{StudentID: sid, StudentName: r.groups.names[sid]}
		if ex := r.find(&domain.Attendance{GroupID: gid, StudentID: sid, LessonDate: date}); ex != nil {
			id, st := ex.ID, ex.Status
			e.AttendanceID, e.Status = &id, &st
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StudentName < out[j].StudentName })
	return out, nil
}
func (r *memRepo) ListByGroupStudent(_ context.Context, gid, sid uuid.UUID) ([]*domain.Attendance, error) {
	var out []*domain.Attendance
	for _, a := range r.records {
		if a.GroupID == gid && a.StudentID == sid {
			cp := *a
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (r *memRepo) Delete(_ context.Context, id uuid.UUID) error { delete(r.records, id); return nil }

// ─── group checker ────────────────────────────────────────────────────────────

type memberKey struct{ gid, sid uuid.UUID }

type memGroups struct {
	groups  map[uuid.UUID]*domain.Group
	members map[memberKey]struct{}
	names   map[uuid.UUID]string
}

func (mg *memGroups) add(gid, sid uuid.UUID, name string) {
	mg.members[memberKey{gid, sid}] = struct{}{}
	mg.names[sid] = name
}
func (mg *memGroups) remove(gid, sid uuid.UUID) { delete(mg.members, memberKey{gid, sid}) }
func (mg *memGroups) GetGroupByID(_ context.Context, id uuid.UUID) (*domain.Group, error) {
	g, ok := mg.groups[id]
	if !ok {
		return nil, nil
	}
	cp := *g
	return &cp, nil
}
func (mg *memGroups) GetMember(_ context.Context, gid, uid uuid.UUID) (*domain.GroupMember, error) {
	if _, ok := mg.members[memberKey{gid, uid}]; !ok {
		return nil, nil
	}
	return &domain.GroupMember{GroupID: gid, StudentID: uid}, nil
}
func (mg *memGroups) ListMembers(_ context.Context, gid uuid.UUID) ([]*domain.GroupMemberDetail, error) {
	var out []*domain.GroupMemberDetail
	for key := range mg.members {
		if key.gid == gid {
			out = append(out, &domain.GroupMemberDetail{
				GroupMember: domain.GroupMember{GroupID: gid, StudentID: key.sid},
				Name:        mg.names[key.sid],
			})
		}
	}
	return out, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

var (
	teacherID = uuid.New()
	studentID = uuid.New() // «Bekzod»
	student2  = uuid.New() // «Aziza»
)

const day = "2024-09-01"

func buildSvc() (*attendance.Service, uuid.UUID, *memGroups, *memRepo) {
	gid := uuid.New()
	mg := &memGroups{
		groups:  map[uuid.UUID]*domain.Group{gid: {ID: gid, TeacherID: teacherID}},
		members: map[memberKey]struct{}{},
		names:   map[uuid.UUID]string{},
	}
	mg.add(gid, studentID, "Bekzod")
	mg.add(gid, student2, "Aziza")
	repo := &memRepo{records: map[uuid.UUID]*domain.Attendance{}, groups: mg}
	return attendance.NewService(repo, mg), gid, mg, repo
}

func mark(t *testing.T, svc *attendance.Service, gid, sid uuid.UUID, st domain.AttendanceStatus) *domain.Attendance {
	t.Helper()
	a, err := svc.Mark(context.Background(), teacherID, gid,
		attendance.MarkRequest{StudentID: sid, LessonDate: day, Status: st})
	require.NoError(t, err)
	return a
}

// ─── Mark ─────────────────────────────────────────────────────────────────────

func TestMark_Success(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	a := mark(t, svc, gid, studentID, domain.AttendancePresent)
	assert.Equal(t, domain.AttendancePresent, a.Status)
}

func TestMark_ForbiddenNonTeacher(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	_, err := svc.Mark(context.Background(), uuid.New(), gid,
		attendance.MarkRequest{StudentID: studentID, LessonDate: day, Status: domain.AttendancePresent})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestMark_StudentNotMember(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	_, err := svc.Mark(context.Background(), teacherID, gid,
		attendance.MarkRequest{StudentID: uuid.New(), LessonDate: day, Status: domain.AttendancePresent})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestMark_BadDateIsValidationError(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	_, err := svc.Mark(context.Background(), teacherID, gid,
		attendance.MarkRequest{StudentID: studentID, LessonDate: "01.09.2024", Status: domain.AttendancePresent})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

// Повторная отметка за тот же день меняет статус и возвращает id существующей
// записи — иначе приложение потом удаляет несуществующую запись.
func TestMark_RemarkKeepsRecordID(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	first := mark(t, svc, gid, studentID, domain.AttendancePresent)
	second := mark(t, svc, gid, studentID, domain.AttendanceAbsent)

	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, domain.AttendanceAbsent, second.Status)
	require.NoError(t, svc.Delete(context.Background(), second.ID, teacherID))
	roster, _ := svc.Roster(context.Background(), gid, day, teacherID)
	for _, e := range roster {
		assert.Nil(t, e.Status, "после удаления отметки ученик снова не отмечен")
	}
}

// ─── Roster ───────────────────────────────────────────────────────────────────

func TestRoster_ListsWholeGroupWithMarks(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	mark(t, svc, gid, studentID, domain.AttendanceLate)

	roster, err := svc.Roster(context.Background(), gid, day, teacherID)
	require.NoError(t, err)
	require.Len(t, roster, 2, "неотмеченные ученики тоже в списке")
	assert.Equal(t, "Aziza", roster[0].StudentName)
	assert.Nil(t, roster[0].Status)
	assert.Equal(t, "Bekzod", roster[1].StudentName)
	require.NotNil(t, roster[1].Status)
	assert.Equal(t, domain.AttendanceLate, *roster[1].Status)
	assert.NotNil(t, roster[1].AttendanceID)
}

func TestRoster_KeepsMarkOfStudentWhoLeft(t *testing.T) {
	svc, gid, mg, _ := buildSvc()
	mark(t, svc, gid, studentID, domain.AttendancePresent)
	mg.remove(gid, studentID)

	roster, err := svc.Roster(context.Background(), gid, day, teacherID)
	require.NoError(t, err)
	assert.Len(t, roster, 2, "вышедший ученик с отметкой за этот день остаётся в списке")
}

func TestRoster_ForbiddenNonTeacher(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	_, err := svc.Roster(context.Background(), gid, day, uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestRoster_BadDate(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	_, err := svc.Roster(context.Background(), gid, "tomorrow", teacherID)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

// ─── MarkMany ─────────────────────────────────────────────────────────────────

func TestMarkMany_MarksAllInOneGo(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	existing := mark(t, svc, gid, studentID, domain.AttendanceAbsent)

	list, err := svc.MarkMany(context.Background(), teacherID, gid, attendance.BatchMarkRequest{
		LessonDate: day,
		Marks: []attendance.BatchMark{
			{StudentID: studentID, Status: domain.AttendancePresent},
			{StudentID: student2, Status: domain.AttendancePresent},
		},
	})
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, existing.ID, list[0].ID, "существующая отметка обновлена, а не задублирована")

	roster, _ := svc.Roster(context.Background(), gid, day, teacherID)
	for _, e := range roster {
		require.NotNil(t, e.Status)
		assert.Equal(t, domain.AttendancePresent, *e.Status)
	}
}

func TestMarkMany_RejectsForeignStudent(t *testing.T) {
	svc, gid, _, repo := buildSvc()
	_, err := svc.MarkMany(context.Background(), teacherID, gid, attendance.BatchMarkRequest{
		LessonDate: day,
		Marks: []attendance.BatchMark{
			{StudentID: studentID, Status: domain.AttendancePresent},
			{StudentID: uuid.New(), Status: domain.AttendancePresent},
		},
	})
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Empty(t, repo.records, "при ошибке не сохраняется ничего")
}

func TestMarkMany_RejectsDuplicateStudent(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	_, err := svc.MarkMany(context.Background(), teacherID, gid, attendance.BatchMarkRequest{
		LessonDate: day,
		Marks: []attendance.BatchMark{
			{StudentID: studentID, Status: domain.AttendancePresent},
			{StudentID: studentID, Status: domain.AttendanceAbsent},
		},
	})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestMarkMany_ForbiddenNonTeacher(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	_, err := svc.MarkMany(context.Background(), uuid.New(), gid, attendance.BatchMarkRequest{
		LessonDate: day,
		Marks:      []attendance.BatchMark{{StudentID: studentID, Status: domain.AttendancePresent}},
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

// ─── ListStudent / Delete ────────────────────────────────────────────────────

func TestListStudent_StudentSelf(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	mark(t, svc, gid, studentID, domain.AttendanceExcused)
	list, err := svc.ListStudent(context.Background(), gid, studentID, studentID, domain.RoleStudent)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestListStudent_StudentForbiddenForOther(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	_, err := svc.ListStudent(context.Background(), gid, studentID, uuid.New(), domain.RoleStudent)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestDelete_ForbiddenForOtherTeacher(t *testing.T) {
	svc, gid, _, _ := buildSvc()
	a := mark(t, svc, gid, studentID, domain.AttendanceAbsent)
	assert.ErrorIs(t, svc.Delete(context.Background(), a.ID, uuid.New()), domain.ErrForbidden)
}
