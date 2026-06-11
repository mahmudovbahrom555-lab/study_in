package attendance

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type Repository interface {
	Upsert(ctx context.Context, a *domain.Attendance) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Attendance, error)
	ListByGroupDate(ctx context.Context, groupID uuid.UUID, date time.Time) ([]*domain.Attendance, error)
	ListByGroupStudent(ctx context.Context, groupID, studentID uuid.UUID) ([]*domain.Attendance, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// GroupChecker avoids a circular import on the groups package.
type GroupChecker interface {
	GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.Group, error)
	GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.GroupMember, error)
}
