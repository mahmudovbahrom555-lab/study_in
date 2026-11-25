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
	AuthorName string         `db:"author_name"`
	Reactions  []PostReaction `db:"-"`
}

// ReactionKind — реакция ученика на объявление.
type ReactionKind string

const (
	ReactionUnderstood ReactionKind = "understood"
	ReactionQuestion   ReactionKind = "question"
)

// PostReaction — реакция с именем ученика; имена видит только репетитор.
type PostReaction struct {
	PostID   uuid.UUID    `db:"post_id"`
	UserID   uuid.UUID    `db:"user_id"`
	Kind     ReactionKind `db:"kind"`
	UserName string       `db:"user_name"`
}
