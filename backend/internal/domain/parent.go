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
