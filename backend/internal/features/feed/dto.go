package feed

import (
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// --- Requests ---

type CreatePostRequest struct {
	Body   string `json:"body"   validate:"required,min=1,max=5000"`
	Pinned bool   `json:"pinned"`
}

type UpdatePostRequest struct {
	Body   *string `json:"body"   validate:"omitempty,min=1,max=5000"`
	Pinned *bool   `json:"pinned"`
}

// --- Responses ---

type AttachmentResponse struct {
	ID        uuid.UUID `json:"id"`
	Filename  string    `json:"filename"`
	MimeType  string    `json:"mime_type"`
	SizeBytes int64     `json:"size_bytes"`
	URL       string    `json:"url"` // signed URL, filled by handler
}

type PostResponse struct {
	ID          uuid.UUID            `json:"id"`
	GroupID     uuid.UUID            `json:"group_id"`
	AuthorID    uuid.UUID            `json:"author_id"`
	AuthorName  string               `json:"author_name"`
	Body        string               `json:"body"`
	Pinned      bool                 `json:"pinned"`
	Attachments []AttachmentResponse `json:"attachments"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type ListResponse struct {
	Posts  []PostResponse `json:"posts"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func postToResponse(p *domain.PostWithMeta, attachments []AttachmentResponse) PostResponse {
	return PostResponse{
		ID:          p.ID,
		GroupID:     p.GroupID,
		AuthorID:    p.AuthorID,
		AuthorName:  p.AuthorName,
		Body:        p.Body,
		Pinned:      p.Pinned,
		Attachments: attachments,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
