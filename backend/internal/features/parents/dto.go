package parents

import (
	"time"

	"github.com/google/uuid"
)

type LinkRequest struct {
	StudentID uuid.UUID `json:"student_id" validate:"required"`
}

type ParentLinkResponse struct {
	ID        uuid.UUID `json:"id"`
	ParentID  uuid.UUID `json:"parent_id"`
	StudentID uuid.UUID `json:"student_id"`
	CreatedAt time.Time `json:"created_at"`
}
