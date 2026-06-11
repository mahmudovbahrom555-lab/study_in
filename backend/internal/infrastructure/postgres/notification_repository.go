package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type NotificationRepository struct{ db *sqlx.DB }

func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	q := `INSERT INTO notifications (user_id, title, body, data)
	      VALUES ($1, $2, $3, $4)
	      RETURNING id, created_at`
	return r.db.QueryRowxContext(ctx, q, n.UserID, n.Title, n.Body, n.Data).
		Scan(&n.ID, &n.CreatedAt)
}

func (r *NotificationRepository) ListUnread(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.Notification, error) {
	var ns []*domain.Notification
	q := `SELECT id, user_id, title, body, data, is_read, created_at
	      FROM notifications
	      WHERE user_id = $1 AND is_read = FALSE
	      ORDER BY created_at DESC
	      LIMIT $2`
	if err := r.db.SelectContext(ctx, &ns, q, userID, limit); err != nil {
		return nil, fmt.Errorf("postgres ListUnread: %w", err)
	}
	return ns, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	q := `UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2`
	if _, err := r.db.ExecContext(ctx, q, notificationID, userID); err != nil {
		return fmt.Errorf("postgres MarkRead: %w", err)
	}
	return nil
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	q := `UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`
	if _, err := r.db.ExecContext(ctx, q, userID); err != nil {
		return fmt.Errorf("postgres MarkAllRead: %w", err)
	}
	return nil
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	q := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`
	if err := r.db.GetContext(ctx, &count, q, userID); err != nil {
		return 0, fmt.Errorf("postgres CountUnread: %w", err)
	}
	return count, nil
}

func (r *NotificationRepository) UpsertDeviceToken(ctx context.Context, userID uuid.UUID, token, platform string) error {
	q := `INSERT INTO device_tokens (user_id, token, platform)
	      VALUES ($1, $2, $3)
	      ON CONFLICT (token) DO UPDATE SET user_id = $1, platform = $3, updated_at = NOW()`
	if _, err := r.db.ExecContext(ctx, q, userID, token, platform); err != nil {
		return fmt.Errorf("postgres UpsertDeviceToken: %w", err)
	}
	return nil
}

func (r *NotificationRepository) DeleteDeviceToken(ctx context.Context, token string) error {
	q := `DELETE FROM device_tokens WHERE token = $1`
	if _, err := r.db.ExecContext(ctx, q, token); err != nil {
		return fmt.Errorf("postgres DeleteDeviceToken: %w", err)
	}
	return nil
}

func (r *NotificationRepository) ListDeviceTokens(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var tokens []string
	q := `SELECT token FROM device_tokens WHERE user_id = $1`
	if err := r.db.SelectContext(ctx, &tokens, q, userID); err != nil {
		return nil, fmt.Errorf("postgres ListDeviceTokens: %w", err)
	}
	return tokens, nil
}
