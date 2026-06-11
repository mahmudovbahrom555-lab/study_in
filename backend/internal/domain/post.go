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

type PostAttachment struct {
	ID        uuid.UUID `db:"id"`
	PostID    uuid.UUID `db:"post_id"`
	ObjectKey string    `db:"object_key"`
	Filename  string    `db:"filename"`
	MimeType  string    `db:"mime_type"`
	SizeBytes int64     `db:"size_bytes"`
	CreatedAt time.Time `db:"created_at"`
}

// PostWithMeta — пост плюс имя автора и вложения (для ответа API).
type PostWithMeta struct {
	Post
	AuthorName string            `db:"author_name"`
	Attachments []PostAttachment `db:"-"`
}
