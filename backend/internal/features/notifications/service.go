package notifications

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

const timeFormat = "2006-01-02T15:04:05Z07:00"

// Service handles in-app notifications and push delivery.
type Service struct {
	repo    Repository
	enqueue TaskEnqueuer
}

func NewService(repo Repository, enqueue TaskEnqueuer) *Service {
	return &Service{repo: repo, enqueue: enqueue}
}

// RegisterDevice saves or updates a device FCM token for the user.
func (s *Service) RegisterDevice(ctx context.Context, userID uuid.UUID, token, platform string) error {
	if err := s.repo.UpsertDeviceToken(ctx, userID, token, platform); err != nil {
		return fmt.Errorf("register device: %w", err)
	}
	return nil
}

// UnregisterDevice removes a device token (e.g. on logout).
func (s *Service) UnregisterDevice(ctx context.Context, token string) error {
	if err := s.repo.DeleteDeviceToken(ctx, token); err != nil {
		return fmt.Errorf("unregister device: %w", err)
	}
	return nil
}

// Send creates an in-app notification and enqueues a push to all user devices.
func (s *Service) Send(ctx context.Context, userID uuid.UUID, title, body string, data any) error {
	n := &domain.Notification{
		UserID: userID,
		Title:  title,
		Body:   body,
		Data:   "{}",
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return fmt.Errorf("create notification: %w", err)
	}

	tokens, err := s.repo.ListDeviceTokens(ctx, userID)
	if err != nil {
		return fmt.Errorf("list device tokens: %w", err)
	}
	for _, tok := range tokens {
		_ = s.enqueue.EnqueuePush(ctx, tok, title, body, data)
	}
	return nil
}

// List returns unread notifications for the caller.
func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]*domain.Notification, error) {
	ns, err := s.repo.ListUnread(ctx, userID, 50)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	return ns, nil
}

// MarkRead marks a single notification as read.
func (s *Service) MarkRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	if err := s.repo.MarkRead(ctx, notificationID, userID); err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	return nil
}

// MarkAllRead marks all of the user's notifications as read.
func (s *Service) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.MarkAllRead(ctx, userID); err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}

// UnreadCount returns the badge count for the user.
func (s *Service) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("unread count: %w", err)
	}
	return count, nil
}

// NotificationDTO is the API representation.
type NotificationDTO struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

func toDTO(n *domain.Notification) NotificationDTO {
	return NotificationDTO{
		ID:        n.ID.String(),
		Title:     n.Title,
		Body:      n.Body,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt.Format(timeFormat),
	}
}

func toDTOs(ns []*domain.Notification) []NotificationDTO {
	out := make([]NotificationDTO, len(ns))
	for i, n := range ns {
		out[i] = toDTO(n)
	}
	return out
}
