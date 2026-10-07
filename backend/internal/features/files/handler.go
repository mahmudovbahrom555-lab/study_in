package files

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	apimw "github.com/mahmudovbahrom555-lab/study_in/backend/internal/middleware"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/response"
)

// multipartOverhead — запас на заголовки multipart сверх самого файла.
const multipartOverhead = 1 << 20

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes:
//
// POST /files — загрузить файл (multipart: file, purpose). Возвращает id для
// последующего прикрепления, например при сдаче ДЗ.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/files", h.upload)
}

type FileResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	MimeType  string    `json:"mime_type"`
	SizeBytes int64     `json:"size_bytes"`
	Width     *int      `json:"width,omitempty"`
	Height    *int      `json:"height,omitempty"`
	URL       string    `json:"url"`
}

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	purpose := domain.FilePurpose(r.FormValue("purpose"))
	// Пока единственное назначение — сдача ДЗ, и загружает её только ученик.
	if purpose != domain.FilePurposeSubmission {
		response.Error(w, domain.NewError("PURPOSE_INVALID", "Неизвестное назначение файла", domain.ErrValidation))
		return
	}
	if apimw.RoleFromCtx(r.Context()) != domain.RoleStudent {
		response.Error(w, domain.ErrForbidden)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize+multipartOverhead)
	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, domain.NewError("FILE_MISSING", "Файл не передан или больше 10 МБ", domain.ErrValidation))
		return
	}
	defer file.Close()

	f, err := h.svc.Upload(r.Context(), apimw.UserIDFromCtx(r.Context()), purpose, header.Filename, file)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Created(w, h.toResponse(r, f))
}

// toResponse — вид загруженного файла в ответе.
func (h *Handler) toResponse(r *http.Request, f *domain.File) FileResponse {
	url, _ := h.svc.SignedURL(r.Context(), f)
	return FileResponse{
		ID:        f.ID,
		Name:      f.OriginalName,
		MimeType:  f.MimeType,
		SizeBytes: f.SizeBytes,
		Width:     f.Width,
		Height:    f.Height,
		URL:       url,
	}
}
