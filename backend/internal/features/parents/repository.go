package parents

import (
	"context"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type Repository interface {
	Link(ctx context.Context, link *domain.ParentLink) error
	Unlink(ctx context.Context, parentID, studentID uuid.UUID) error
	ListChildren(ctx context.Context, parentID uuid.UUID) ([]*domain.ParentLink, error)
	ListParents(ctx context.Context, studentID uuid.UUID) ([]*domain.ParentLink, error)
	GetLink(ctx context.Context, parentID, studentID uuid.UUID) (*domain.ParentLink, error)
}

// GradeReader and AttendanceReader allow parents to read child data without
// importing the grades/attendance packages directly.

type GradeReader interface {
	ListByGroupStudent(ctx context.Context, groupID, studentID uuid.UUID) ([]*domain.Grade, error)
}

type AttendanceReader interface {
	ListByGroupStudent(ctx context.Context, groupID, studentID uuid.UUID) ([]*domain.Attendance, error)
}

// GroupMembership lets parents find which groups their child belongs to.
type GroupMemberLister interface {
	ListGroupsByStudent(ctx context.Context, studentID uuid.UUID) ([]*domain.Group, error)
}
