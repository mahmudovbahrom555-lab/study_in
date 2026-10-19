// Package chat — чат в классе: сообщения со сроком жизни, модерация
// репетитором (docs/design/CHAT.md).
package chat

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// Repository — хранилище чата.
type Repository interface {
	// Access — права пользователя в чате класса одним запросом; nil — класса нет.
	Access(ctx context.Context, groupID, userID uuid.UUID) (*domain.ChatAccess, error)

	// Insert сохраняет сообщение; повтор с тем же (author_id, client_id)
	// возвращает уже сохранённое, не создавая дубль.
	Insert(ctx context.Context, m *domain.ChatMessage) (*domain.ChatMessage, error)
	// List — живые сообщения класса, новые первыми; before — id, раньше которого.
	List(ctx context.Context, groupID uuid.UUID, before *uuid.UUID, limit int, now time.Time) ([]*domain.ChatMessage, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.ChatMessage, error)
	Delete(ctx context.Context, id uuid.UUID) error
	// DeleteExpired удаляет истёкшие сообщения пачкой, возвращает число удалённых.
	DeleteExpired(ctx context.Context, now time.Time, batch int) (int, error)

	MarkRead(ctx context.Context, groupID, userID uuid.UUID, at time.Time) error
	// UnreadCounts — непрочитанные чужие живые сообщения по классам пользователя.
	UnreadCounts(ctx context.Context, userID uuid.UUID, groupIDs []uuid.UUID, now time.Time) (map[uuid.UUID]int, error)

	UpdateSettings(ctx context.Context, groupID uuid.UUID, ttlSeconds int, mode domain.ChatMode) error
	SetMuted(ctx context.Context, groupID, userID uuid.UUID, muted bool) error
	// Ban удаляет участника из класса и запрещает вступать снова (одна транзакция).
	Ban(ctx context.Context, groupID, userID uuid.UUID) error
}

// Limiter ограничивает частоту отправки (Redis, как у SMS).
type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) error
}

// Publisher рассылает события чата подключённым клиентам (WebSocket).
// nil — рассылки нет, сообщения доступны через REST.
type Publisher interface {
	Publish(ctx context.Context, groupID uuid.UUID, e Event)
}

// Event — событие чата для подключённых клиентов.
type Event struct {
	Type    string `json:"type"` // message.created | message.deleted | chat.settings_changed | member.muted | member.unmuted | member.removed
	GroupID string `json:"group_id"`
	Data    any    `json:"data,omitempty"`
}
