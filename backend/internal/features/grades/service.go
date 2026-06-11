package grades

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

// CreateGrade adds a grade for a student. Only the group's teacher may do this.
func (s *Service) CreateGrade(ctx context.Context, teacherID, groupID uuid.UUID, req CreateGradeRequest) (*domain.Grade, error) {
	g, err := s.groups.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("grades.CreateGrade fetch group: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}
	if g.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	m, err := s.groups.GetMember(ctx, groupID, req.StudentID)
	if err != nil {
		return nil, fmt.Errorf("grades.CreateGrade fetch member: %w", err)
	}
	if m == nil {
		return nil, domain.ErrNotFound
	}

	maxValue := req.MaxValue
	if maxValue == 0 {
		maxValue = 100
	}

	gradedAt := time.Now()
	if req.GradedAt != nil {
		if t, err2 := time.Parse(time.RFC3339, *req.GradedAt); err2 == nil {
			gradedAt = t
		}
	}

	now := time.Now()
	grade := &domain.Grade{
		ID:        uuid.New(),
		GroupID:   groupID,
		StudentID: req.StudentID,
		TeacherID: teacherID,
		Subject:   req.Subject,
		Value:     req.Value,
		MaxValue:  maxValue,
		Comment:   req.Comment,
		GradedAt:  gradedAt,
		CreatedAt: now,
	}
	if err := s.repo.Create(ctx, grade); err != nil {
		return nil, fmt.Errorf("grades.CreateGrade: %w", err)
	}
	return grade, nil
}

// ListGroupGrades returns all grades for a group.
// Teachers see all; students see only their own.
func (s *Service) ListGroupGrades(ctx context.Context, groupID, callerID uuid.UUID, role domain.Role) ([]*domain.Grade, error) {
	g, err := s.groups.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("grades.ListGroupGrades: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}

	if role == domain.RoleTeacher {
		if g.TeacherID != callerID {
			return nil, domain.ErrForbidden
		}
		return s.repo.ListByGroup(ctx, groupID)
	}

	m, err := s.groups.GetMember(ctx, groupID, callerID)
	if err != nil {
		return nil, fmt.Errorf("grades.ListGroupGrades member: %w", err)
	}
	if m == nil {
		return nil, domain.ErrForbidden
	}
	return s.repo.ListByGroupStudent(ctx, groupID, callerID)
}

// ListStudentGrades returns grades for a specific student in a group.
// Teacher can query any student; student can only query themselves.
func (s *Service) ListStudentGrades(ctx context.Context, groupID, studentID, callerID uuid.UUID, role domain.Role) ([]*domain.Grade, error) {
	g, err := s.groups.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("grades.ListStudentGrades: %w", err)
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
			return nil, fmt.Errorf("grades.ListStudentGrades member: %w", err)
		}
		if m == nil {
			return nil, domain.ErrForbidden
		}
	}

	return s.repo.ListByGroupStudent(ctx, groupID, studentID)
}

// UpdateGrade lets the teacher edit an existing grade.
func (s *Service) UpdateGrade(ctx context.Context, gradeID, teacherID uuid.UUID, req UpdateGradeRequest) (*domain.Grade, error) {
	grade, err := s.repo.GetByID(ctx, gradeID)
	if err != nil {
		return nil, fmt.Errorf("grades.UpdateGrade fetch: %w", err)
	}
	if grade == nil {
		return nil, domain.ErrNotFound
	}
	if grade.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	if req.Subject != nil {
		grade.Subject = *req.Subject
	}
	if req.Value != nil {
		grade.Value = *req.Value
	}
	if req.MaxValue != nil {
		grade.MaxValue = *req.MaxValue
	}
	if req.Comment != nil {
		grade.Comment = req.Comment
	}
	if req.GradedAt != nil {
		if t, err2 := time.Parse(time.RFC3339, *req.GradedAt); err2 == nil {
			grade.GradedAt = t
		}
	}

	if err := s.repo.Update(ctx, grade); err != nil {
		return nil, fmt.Errorf("grades.UpdateGrade: %w", err)
	}
	return grade, nil
}

// DeleteGrade removes a grade (teacher only).
func (s *Service) DeleteGrade(ctx context.Context, gradeID, teacherID uuid.UUID) error {
	grade, err := s.repo.GetByID(ctx, gradeID)
	if err != nil {
		return fmt.Errorf("grades.DeleteGrade fetch: %w", err)
	}
	if grade == nil {
		return domain.ErrNotFound
	}
	if grade.TeacherID != teacherID {
		return domain.ErrForbidden
	}
	if err := s.repo.Delete(ctx, gradeID); err != nil {
		return fmt.Errorf("grades.DeleteGrade: %w", err)
	}
	return nil
}
