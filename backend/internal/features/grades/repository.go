package grades

import (
	"context"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, g *domain.Grade) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Grade, error)
	ListByGroupStudent(ctx context.Context, groupID, studentID uuid.UUID) ([]*domain.Grade, error)
	ListByGroup(ctx context.Context, groupID uuid.UUID) ([]*domain.Grade, error)
	Update(ctx context.Context, g *domain.Grade) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// GroupChecker allows grades to verify group membership without importing the groups package.
type GroupChecker interface {
	GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.Group, error)
	GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.GroupMember, error)
}
