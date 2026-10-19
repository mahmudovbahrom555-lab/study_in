package domain

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID        uuid.UUID  `db:"id"`
	GroupID   uuid.UUID  `db:"group_id"`
	AuthorID  uuid.UUID  `db:"author_id"`
	Body      string     `db:"body"`
	Pinned    bool       `db:"pinned"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

// PostWithMeta — пост плюс имя автора (для ответа API).
type PostWithMeta struct {
	Post
	AuthorName string `db:"author_name"`
}
