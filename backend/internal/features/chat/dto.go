package chat

import (
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type SendRequest struct {
	Body     string    `json:"body"`
	ClientID uuid.UUID `json:"client_id"`
}

type ReadRequest struct {
	LastReadAt time.Time `json:"last_read_at"`
}

type SettingsResponse struct {
	TTLSeconds int    `json:"ttl_seconds"`
	Mode       string `json:"mode"`
}

type MessageResponse struct {
	ID         uuid.UUID `json:"id"`
	GroupID    uuid.UUID `json:"group_id"`
	AuthorID   uuid.UUID `json:"author_id"`
	AuthorName string    `json:"author_name"`
	// FromOwner — сообщение владельца класса: только в них ссылки кликабельны.
	FromOwner bool      `json:"from_owner"`
	Body      string    `json:"body"`
	ClientID  uuid.UUID `json:"client_id"`
	CreatedAt time.Time `json:"created_at"`
	// ExpiresAt — приложение само скрывает сообщение после этого момента.
	ExpiresAt time.Time `json:"expires_at"`
}

func toResponse(m *domain.ChatMessage, ownerID uuid.UUID) MessageResponse {
	return MessageResponse{
		ID:         m.ID,
		GroupID:    m.GroupID,
		AuthorID:   m.AuthorID,
		AuthorName: m.AuthorName,
		FromOwner:  m.AuthorID == ownerID,
		Body:       m.Body,
		ClientID:   m.ClientID,
		CreatedAt:  m.CreatedAt,
		ExpiresAt:  m.ExpiresAt,
	}
}
