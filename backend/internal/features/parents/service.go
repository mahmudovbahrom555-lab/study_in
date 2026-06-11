package parents

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type Service struct {
	repo        Repository
	grades      GradeReader
	attendance  AttendanceReader
	groupMember GroupMemberLister
}

func NewService(
	repo Repository,
	grades GradeReader,
	attendance AttendanceReader,
	groupMember GroupMemberLister,
) *Service {
	return &Service{
		repo:        repo,
		grades:      grades,
		attendance:  attendance,
		groupMember: groupMember,
	}
}

// LinkChild links a parent to a student. Only users with the parent role may call this.
func (s *Service) LinkChild(ctx context.Context, parentID uuid.UUID, req LinkRequest) (*domain.ParentLink, error) {
	existing, err := s.repo.GetLink(ctx, parentID, req.StudentID)
	if err != nil {
		return nil, fmt.Errorf("parents.LinkChild get: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrConflict
	}

	link := &domain.ParentLink{
		ID:        uuid.New(),
		ParentID:  parentID,
		StudentID: req.StudentID,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Link(ctx, link); err != nil {
		return nil, fmt.Errorf("parents.LinkChild: %w", err)
	}
	return link, nil
}

// UnlinkChild removes the parent→student relationship.
func (s *Service) UnlinkChild(ctx context.Context, parentID, studentID uuid.UUID) error {
	existing, err := s.repo.GetLink(ctx, parentID, studentID)
	if err != nil {
		return fmt.Errorf("parents.UnlinkChild get: %w", err)
	}
	if existing == nil {
		return domain.ErrNotFound
	}
	if err := s.repo.Unlink(ctx, parentID, studentID); err != nil {
		return fmt.Errorf("parents.UnlinkChild: %w", err)
	}
	return nil
}

// ListChildren returns all students linked to the parent.
func (s *Service) ListChildren(ctx context.Context, parentID uuid.UUID) ([]*domain.ParentLink, error) {
	links, err := s.repo.ListChildren(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("parents.ListChildren: %w", err)
	}
	return links, nil
}

// assertLinked verifies that parentID is linked to studentID.
func (s *Service) assertLinked(ctx context.Context, parentID, studentID uuid.UUID) error {
	link, err := s.repo.GetLink(ctx, parentID, studentID)
	if err != nil {
		return fmt.Errorf("parents.assertLinked: %w", err)
	}
	if link == nil {
		return domain.ErrForbidden
	}
	return nil
}

// ChildGroups returns groups the child belongs to (read-only for parent).
func (s *Service) ChildGroups(ctx context.Context, parentID, studentID uuid.UUID) ([]*domain.Group, error) {
	if err := s.assertLinked(ctx, parentID, studentID); err != nil {
		return nil, err
	}
	gs, err := s.groupMember.ListGroupsByStudent(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("parents.ChildGroups: %w", err)
	}
	return gs, nil
}

// ChildGrades returns grades for a child in a group (parent read-only).
func (s *Service) ChildGrades(ctx context.Context, parentID, studentID, groupID uuid.UUID) ([]*domain.Grade, error) {
	if err := s.assertLinked(ctx, parentID, studentID); err != nil {
		return nil, err
	}
	gs, err := s.grades.ListByGroupStudent(ctx, groupID, studentID)
	if err != nil {
		return nil, fmt.Errorf("parents.ChildGrades: %w", err)
	}
	return gs, nil
}

// ChildAttendance returns attendance records for a child in a group.
func (s *Service) ChildAttendance(ctx context.Context, parentID, studentID, groupID uuid.UUID) ([]*domain.Attendance, error) {
	if err := s.assertLinked(ctx, parentID, studentID); err != nil {
		return nil, err
	}
	list, err := s.attendance.ListByGroupStudent(ctx, groupID, studentID)
	if err != nil {
		return nil, fmt.Errorf("parents.ChildAttendance: %w", err)
	}
	return list, nil
}
