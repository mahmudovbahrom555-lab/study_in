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

type ReactRequest struct {
	Kind domain.ReactionKind `json:"kind" validate:"required,oneof=understood question"`
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

	Reactions ReactionsResponse `json:"reactions"`
}

// ReactionsResponse — счётчики видят все; имена — только репетитор класса.
type ReactionsResponse struct {
	Understood   int                 `json:"understood"`
	Question     int                 `json:"question"`
	Mine         domain.ReactionKind `json:"mine,omitempty"`
	UnderstoodBy []string            `json:"understood_by,omitempty"`
	QuestionBy   []string            `json:"question_by,omitempty"`
}

func reactionsToResponse(rs []domain.PostReaction, viewerID uuid.UUID, isTeacher bool) ReactionsResponse {
	var out ReactionsResponse
	for _, r := range rs {
		if r.UserID == viewerID {
			out.Mine = r.Kind
		}
		switch r.Kind {
		case domain.ReactionUnderstood:
			out.Understood++
			if isTeacher {
				out.UnderstoodBy = append(out.UnderstoodBy, r.UserName)
			}
		case domain.ReactionQuestion:
			out.Question++
			if isTeacher {
				out.QuestionBy = append(out.QuestionBy, r.UserName)
			}
		}
	}
	return out
}

func postToResponse(p *domain.PostWithMeta, viewerID uuid.UUID, isTeacher bool) PostResponse {
	return PostResponse{
		ID:         p.ID,
		GroupID:    p.GroupID,
		AuthorID:   p.AuthorID,
		AuthorName: p.AuthorName,
		Body:       p.Body,
		Pinned:     p.Pinned,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
		Reactions:  reactionsToResponse(p.Reactions, viewerID, isTeacher),
	}
}
