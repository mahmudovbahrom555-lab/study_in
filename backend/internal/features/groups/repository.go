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
	// IsBanned — репетитор заблокировал пользователя в классе (вступить снова нельзя).
	IsBanned(ctx context.Context, groupID, userID uuid.UUID) (bool, error)

	SearchPublicGroups(ctx context.Context, pattern, username string, limit int) ([]*domain.GroupPreview, error)
	GetApplicant(ctx context.Context, studentID uuid.UUID) (*domain.Applicant, error)
	ListParents(ctx context.Context, studentID uuid.UUID) ([]domain.ParentRef, error)
	ListParentConsents(ctx context.Context, parentID uuid.UUID) ([]*domain.ParentConsent, error)
	DecideParentConsent(ctx context.Context, requestID, parentID uuid.UUID, approve bool) (bool, error)
	GetGroupPreviewByCode(ctx context.Context, code string) (*domain.GroupPreview, error)
	// UpsertJoinRequest создаёт заявку или снова делает прежнюю ожидающей
	// (status — pending или parent_pending у ребёнка младше 13).
	UpsertJoinRequest(ctx context.Context, groupID, studentID uuid.UUID, status domain.JoinRequestStatus) (*domain.GroupJoinRequest, error)
	GetJoinRequest(ctx context.Context, id uuid.UUID) (*domain.GroupJoinRequest, error)
	ListPendingJoinRequests(ctx context.Context, groupID uuid.UUID) ([]*domain.JoinRequestDetail, error)
	ListMyPendingJoinRequests(ctx context.Context, studentID uuid.UUID) ([]*domain.MyJoinRequest, error)
	// AcceptJoinRequest в одной транзакции добавляет ученика и закрывает заявку.
	AcceptJoinRequest(ctx context.Context, requestID uuid.UUID, m *domain.GroupMember) error
	DeclineJoinRequest(ctx context.Context, requestID uuid.UUID) error
}
