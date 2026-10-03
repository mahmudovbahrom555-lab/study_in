package domain

import (
	"time"

	"github.com/google/uuid"
)

type ParentLink struct {
	ID        uuid.UUID `db:"id"`
	ParentID  uuid.UUID `db:"parent_id"`
	StudentID uuid.UUID `db:"student_id"`
	CreatedAt time.Time `db:"created_at"`
}

// ParentLinkCode — короткий код, который ученик передаёт родителю для привязки.
type ParentLinkCode struct {
	Code      string    `db:"code"`
	StudentID uuid.UUID `db:"student_id"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}
