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

// PostResponse — запись ленты. Файлы к записям появятся через модуль files
// (как у ДЗ), когда понадобятся материалы в ленте.
type PostResponse struct {
	ID         uuid.UUID `json:"id"`
	GroupID    uuid.UUID `json:"group_id"`
	AuthorID   uuid.UUID `json:"author_id"`
	AuthorName string    `json:"author_name"`
	Body       string    `json:"body"`
	Pinned     bool      `json:"pinned"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func postToResponse(p *domain.PostWithMeta) PostResponse {
	return PostResponse{
		ID:         p.ID,
		GroupID:    p.GroupID,
		AuthorID:   p.AuthorID,
		AuthorName: p.AuthorName,
		Body:       p.Body,
		Pinned:     p.Pinned,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}
