package attendance

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	apimw "github.com/mahmudovbahrom555-lab/study_in/backend/internal/middleware"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/response"
)

type Handler struct {
	svc      *Service
	validate *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// RegisterRoutes mounts attendance endpoints.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/groups/{groupID}/attendance", func(r chi.Router) {
		// GET /groups/{groupID}/attendance?date=2024-09-01
		r.Get("/", h.listByDate)
		// POST /groups/{groupID}/attendance
		r.Post("/", h.mark)
		// DELETE /groups/{groupID}/attendance/{id}
		r.Delete("/{id}", h.delete)
	})
	// GET /groups/{groupID}/students/{studentID}/attendance
	r.Get("/groups/{groupID}/students/{studentID}/attendance", h.listStudent)
}

func (h *Handler) mark(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	if role != domain.RoleTeacher {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teachers can mark attendance")
		return
	}

	var req MarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	a, err := h.svc.Mark(r.Context(), callerID, groupID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.Created(w, toResponse(a))
}

func (h *Handler) listByDate(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "date query param required (YYYY-MM-DD)")
		return
	}

	list, err := h.svc.ListByDate(r.Context(), groupID, dateStr, callerID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]AttendanceResponse, len(list))
	for i, a := range list {
		out[i] = toResponse(a)
	}
	response.OK(w, out)
}

func (h *Handler) listStudent(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	studentID, ok := parseUUID(w, chi.URLParam(r, "studentID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	list, err := h.svc.ListStudent(r.Context(), groupID, studentID, callerID, role)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]AttendanceResponse, len(list))
	for i, a := range list {
		out[i] = toResponse(a)
	}
	response.OK(w, out)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, chi.URLParam(r, "id"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	if err := h.svc.Delete(r.Context(), id, callerID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func parseUUID(w http.ResponseWriter, s string) (uuid.UUID, bool) {
	id, err := uuid.Parse(s)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid UUID")
		return uuid.Nil, false
	}
	return id, true
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
			"code": "VALIDATION_ERROR", "message": "validation failed", "details": details,
		},
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "not found")
	case errors.Is(err, domain.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
	case errors.Is(err, domain.ErrConflict):
		writeError(w, http.StatusConflict, "CONFLICT", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
