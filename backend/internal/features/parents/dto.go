package parents

import (
	"time"

	"github.com/google/uuid"
)

type LinkByCodeRequest struct {
	Code string `json:"code" validate:"required,min=4,max=12"`
}

type LinkCodeResponse struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ParentLinkResponse struct {
	ID        uuid.UUID `json:"id"`
	ParentID  uuid.UUID `json:"parent_id"`
	StudentID uuid.UUID `json:"student_id"`
	CreatedAt time.Time `json:"created_at"`
}
