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
	g, err := s.groups.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("attendance.Mark fetch group: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}
	if g.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	m, err := s.groups.GetMember(ctx, groupID, req.StudentID)
	if err != nil {
		return nil, fmt.Errorf("attendance.Mark fetch member: %w", err)
	}
	if m == nil {
		return nil, domain.ErrNotFound
	}

	lessonDate, err := time.Parse("2006-01-02", req.LessonDate)
	if err != nil {
		return nil, domain.ErrConflict // malformed date handled as conflict
	}

	now := time.Now()
	a := &domain.Attendance{
		ID:         uuid.New(),
		GroupID:    groupID,
		StudentID:  req.StudentID,
		TeacherID:  teacherID,
		LessonDate: lessonDate,
		Status:     req.Status,
		Note:       req.Note,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.Upsert(ctx, a); err != nil {
		return nil, fmt.Errorf("attendance.Mark: %w", err)
	}
	return a, nil
}

// ListByDate returns all attendance records for a group on a given date (teacher only).
func (s *Service) ListByDate(ctx context.Context, groupID uuid.UUID, dateStr string, teacherID uuid.UUID) ([]*domain.Attendance, error) {
	g, err := s.groups.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("attendance.ListByDate: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}
	if g.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, domain.ErrConflict
	}
	return s.repo.ListByGroupDate(ctx, groupID, date)
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
