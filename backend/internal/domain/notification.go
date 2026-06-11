package domain

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Title     string    `db:"title"`
	Body      string    `db:"body"`
	Data      string    `db:"data"` // JSON string
	IsRead    bool      `db:"is_read"`
	CreatedAt time.Time `db:"created_at"`
}

