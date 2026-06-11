package feed

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	apimw "github.com/mahmudovbahrom555-lab/study_in/backend/internal/middleware"
)

// URLSigner generates a signed URL for a MinIO object key.
// In Etap 4, this will be replaced by the real MinIO implementation.
type URLSigner interface {
	SignURL(ctx interface{}, objectKey string) (string, error)
}

// noopSigner returns the object key as-is until MinIO is wired in Etap 4.
type noopSigner struct{}

func (noopSigner) SignURL(_ interface{}, key string) (string, error) { return key, nil }

type Handler struct {
	svc      *Service
	validate *validator.Validate
	signer   URLSigner
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, validate: validator.New(), signer: noopSigner{}}
}

// RegisterRoutes mounts feed routes.
// All routes require auth middleware upstream.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/groups/{groupID}/feed", func(r chi.Router) {
		r.Get("/", h.listPosts)
		r.Post("/", h.createPost)

		r.Route("/{postID}", func(r chi.Router) {
			r.Get("/", h.getPost)
			r.Patch("/", h.updatePost)
			r.Delete("/", h.deletePost)
			r.Post("/pin", h.pinPost)
			r.Post("/unpin", h.unpinPost)
		})
	})
}

// GET /groups/{groupID}/feed?limit=20&offset=0
func (h *Handler) listPosts(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	posts, err := h.svc.ListPosts(r.Context(), groupID, callerID, role, limit, offset)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := make([]PostResponse, len(posts))
	for i, p := range posts {
		atts, _ := h.svc.ListAttachments(r.Context(), p.ID)
		out[i] = postToResponse(p, h.buildAttachments(r, atts))
	}
	writeJSON(w, http.StatusOK, out)
}

// POST /groups/{groupID}/feed
func (h *Handler) createPost(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	if role != domain.RoleTeacher {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teachers can post")
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	p, err := h.svc.CreatePost(r.Context(), callerID, groupID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, postToResponse(p, nil))
}

// GET /groups/{groupID}/feed/{postID}
func (h *Handler) getPost(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	postID, ok := parseUUID(w, chi.URLParam(r, "postID"))
	if !ok {
		return
	}
	_ = groupID // validated via assertVisible inside service
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	p, atts, err := h.svc.GetPost(r.Context(), postID, callerID, role)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, postToResponse(p, h.buildAttachments(r, atts)))
}

// PATCH /groups/{groupID}/feed/{postID}
func (h *Handler) updatePost(w http.ResponseWriter, r *http.Request) {
	postID, ok := parseUUID(w, chi.URLParam(r, "postID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	var req UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	p, err := h.svc.UpdatePost(r.Context(), postID, callerID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, postToResponse(p, nil))
}

// DELETE /groups/{groupID}/feed/{postID}
func (h *Handler) deletePost(w http.ResponseWriter, r *http.Request) {
	postID, ok := parseUUID(w, chi.URLParam(r, "postID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	if err := h.svc.DeletePost(r.Context(), postID, callerID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /groups/{groupID}/feed/{postID}/pin
func (h *Handler) pinPost(w http.ResponseWriter, r *http.Request) {
	h.setPinned(w, r, true)
}

// POST /groups/{groupID}/feed/{postID}/unpin
func (h *Handler) unpinPost(w http.ResponseWriter, r *http.Request) {
	h.setPinned(w, r, false)
}

func (h *Handler) setPinned(w http.ResponseWriter, r *http.Request, pinned bool) {
	postID, ok := parseUUID(w, chi.URLParam(r, "postID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	p, err := h.svc.PinPost(r.Context(), postID, callerID, pinned)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, postToResponse(p, nil))
}

// buildAttachments converts domain attachments into response DTOs with signed URLs.
func (h *Handler) buildAttachments(r *http.Request, atts []*domain.PostAttachment) []AttachmentResponse {
	if len(atts) == 0 {
		return nil
	}
	out := make([]AttachmentResponse, len(atts))
	for i, a := range atts {
		url, _ := h.signer.SignURL(r.Context(), a.ObjectKey)
		out[i] = AttachmentResponse{
			ID:        a.ID,
			Filename:  a.Filename,
			MimeType:  a.MimeType,
			SizeBytes: a.SizeBytes,
			URL:       url,
		}
	}
	return out
}

// --- helpers (same pattern as groups/handler.go) ---

func parseUUID(w http.ResponseWriter, s string) (uuid.UUID, bool) {
	id, err := uuid.Parse(s)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid UUID")
		return uuid.Nil, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeValidationError(w http.ResponseWriter, err error) {
	var ve validator.ValidationErrors
	details := map[string]any{}
	if errors.As(err, &ve) {
		for _, fe := range ve {
			details[fe.Field()] = fe.Tag()
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":    "VALIDATION_ERROR",
			"message": "validation failed",
			"details": details,
		},
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "not found")
	case errors.Is(err, domain.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
