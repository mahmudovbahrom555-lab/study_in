package feed

import (
	"context"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// Repository — хранилище постов и вложений.
type Repository interface {
	CreatePost(ctx context.Context, p *domain.Post) error
	GetPostByID(ctx context.Context, id uuid.UUID) (*domain.Post, error)
	ListGroupPosts(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*domain.PostWithMeta, error)
	UpdatePost(ctx context.Context, p *domain.Post) error
	SoftDeletePost(ctx context.Context, id uuid.UUID) error
	PinPost(ctx context.Context, id uuid.UUID, pinned bool) error

	// Attachments — созданы вместе с постом, удаляются каскадно.
	AddAttachment(ctx context.Context, a *domain.PostAttachment) error
	ListAttachments(ctx context.Context, postID uuid.UUID) ([]*domain.PostAttachment, error)
}

// GroupChecker — минимальный интерфейс для проверки членства,
// чтобы не тащить весь groups.Repository в зависимость.
type GroupChecker interface {
	GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.Group, error)
	GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.GroupMember, error)
}
