// Package groups реализует фичу групп: создание, участники, приглашения.
package groups

import (
	"context"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// Repository — хранилище данных для groups-фичи.
type Repository interface {
	// Groups
	CreateGroup(ctx context.Context, g *domain.Group) error
	GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.Group, error)
	GetGroupByInviteCode(ctx context.Context, code string) (*domain.Group, error)
	ListTeacherGroups(ctx context.Context, teacherID uuid.UUID) ([]*domain.Group, error)
	ListStudentGroups(ctx context.Context, studentID uuid.UUID) ([]*domain.Group, error)
	UpdateGroup(ctx context.Context, g *domain.Group) error
	SoftDeleteGroup(ctx context.Context, id uuid.UUID) error

	// Members
	AddMember(ctx context.Context, m *domain.GroupMember) error
	RemoveMember(ctx context.Context, groupID, studentID uuid.UUID) error
	GetMember(ctx context.Context, groupID, studentID uuid.UUID) (*domain.GroupMember, error)
	ListMembers(ctx context.Context, groupID uuid.UUID) ([]*domain.GroupMemberDetail, error)
	UpdateMemberPayment(ctx context.Context, groupID, studentID uuid.UUID, status domain.PaymentStatus) error
}
