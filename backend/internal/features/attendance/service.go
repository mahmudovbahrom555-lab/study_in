package attendance

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type Service struct {
	repo   Repository
	groups GroupChecker
}

func NewService(repo Repository, groups GroupChecker) *Service {
	return &Service{repo: repo, groups: groups}
}

// Mark creates or updates an attendance record (upsert by group+student+date).
func (s *Service) Mark(ctx context.Context, teacherID, groupID uuid.UUID, req MarkRequest) (*domain.Attendance, error) {
	if err := s.assertOwner(ctx, groupID, teacherID); err != nil {
		return nil, err
	}
	m, err := s.groups.GetMember(ctx, groupID, req.StudentID)
	if err != nil {
		return nil, fmt.Errorf("attendance.Mark fetch member: %w", err)
	}
	if m == nil {
		return nil, domain.ErrNotFound
	}
	date, err := parseDate(req.LessonDate)
	if err != nil {
		return nil, err
	}

	a := newRecord(groupID, req.StudentID, teacherID, date, req.Status)
	a.Note = req.Note
	if err := s.repo.Upsert(ctx, a); err != nil {
		return nil, fmt.Errorf("attendance.Mark: %w", err)
	}
	return a, nil
}

// MarkMany отмечает нескольких учеников за один день одной транзакцией:
// либо сохраняются все отметки, либо ни одной.
func (s *Service) MarkMany(ctx context.Context, teacherID, groupID uuid.UUID, req BatchMarkRequest) ([]*domain.Attendance, error) {
	if err := s.assertOwner(ctx, groupID, teacherID); err != nil {
		return nil, err
	}
	date, err := parseDate(req.LessonDate)
	if err != nil {
		return nil, err
	}

	members, err := s.groups.ListMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("attendance.MarkMany members: %w", err)
	}
	inGroup := make(map[uuid.UUID]bool, len(members))
	for _, m := range members {
		inGroup[m.StudentID] = true
	}

	seen := make(map[uuid.UUID]bool, len(req.Marks))
	list := make([]*domain.Attendance, 0, len(req.Marks))
	for _, mk := range req.Marks {
		if seen[mk.StudentID] {
			return nil, domain.ErrValidation // один ученик дважды в одном запросе
		}
		seen[mk.StudentID] = true
		if !inGroup[mk.StudentID] {
			return nil, domain.ErrNotFound
		}
		list = append(list, newRecord(groupID, mk.StudentID, teacherID, date, mk.Status))
	}

	if err := s.repo.UpsertMany(ctx, list); err != nil {
		return nil, fmt.Errorf("attendance.MarkMany: %w", err)
	}
	return list, nil
}

// Roster — список занятия на дату: все ученики группы и их отметки (teacher only).
func (s *Service) Roster(ctx context.Context, groupID uuid.UUID, dateStr string, teacherID uuid.UUID) ([]*domain.RosterEntry, error) {
	if err := s.assertOwner(ctx, groupID, teacherID); err != nil {
		return nil, err
	}
	date, err := parseDate(dateStr)
	if err != nil {
		return nil, err
	}
	list, err := s.repo.Roster(ctx, groupID, date)
	if err != nil {
		return nil, fmt.Errorf("attendance.Roster: %w", err)
	}
	return list, nil
}

// ListStudent returns attendance history for a student in a group.
// Teacher can view any student; students can only view themselves.
func (s *Service) ListStudent(ctx context.Context, groupID, studentID, callerID uuid.UUID, role domain.Role) ([]*domain.Attendance, error) {
	g, err := s.groups.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("attendance.ListStudent: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}

	if role == domain.RoleTeacher {
		if g.TeacherID != callerID {
			return nil, domain.ErrForbidden
		}
	} else {
		if callerID != studentID {
			return nil, domain.ErrForbidden
		}
		m, err := s.groups.GetMember(ctx, groupID, callerID)
		if err != nil {
			return nil, fmt.Errorf("attendance.ListStudent member: %w", err)
		}
		if m == nil {
			return nil, domain.ErrForbidden
		}
	}

	return s.repo.ListByGroupStudent(ctx, groupID, studentID)
}

// Delete removes an attendance record (teacher only).
func (s *Service) Delete(ctx context.Context, id, teacherID uuid.UUID) error {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("attendance.Delete fetch: %w", err)
	}
	if a == nil {
		return domain.ErrNotFound
	}
	if a.TeacherID != teacherID {
		return domain.ErrForbidden
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("attendance.Delete: %w", err)
	}
	return nil
}

func (s *Service) assertOwner(ctx context.Context, groupID, teacherID uuid.UUID) error {
	g, err := s.groups.GetGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("attendance fetch group: %w", err)
	}
	if g == nil {
		return domain.ErrNotFound
	}
	if g.TeacherID != teacherID {
		return domain.ErrForbidden
	}
	return nil
}

func parseDate(s string) (time.Time, error) {
	d, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, domain.ErrValidation
	}
	return d, nil
}

func newRecord(groupID, studentID, teacherID uuid.UUID, date time.Time, status domain.AttendanceStatus) *domain.Attendance {
	now := time.Now()
	return &domain.Attendance{
		ID:         uuid.New(),
		GroupID:    groupID,
		StudentID:  studentID,
		TeacherID:  teacherID,
		LessonDate: date,
		Status:     status,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
