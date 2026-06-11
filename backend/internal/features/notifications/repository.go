package notifications

import (
	"context"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// Repository persists notifications and device tokens.
type Repository interface {
	// Notifications
	Create(ctx context.Context, n *domain.Notification) error
	ListUnread(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.Notification, error)
	MarkRead(ctx context.Context, notificationID, userID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)

	// Device tokens
	UpsertDeviceToken(ctx context.Context, userID uuid.UUID, token, platform string) error
	DeleteDeviceToken(ctx context.Context, token string) error
	ListDeviceTokens(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// TaskEnqueuer schedules async push notification delivery.
type TaskEnqueuer interface {
	EnqueuePush(ctx context.Context, token, title, body string, data any) error
}
