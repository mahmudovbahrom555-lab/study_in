package grades

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

// RegisterRoutes mounts grade endpoints under /groups/{groupID}/grades.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/groups/{groupID}/grades", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)

		r.Route("/{gradeID}", func(r chi.Router) {
			r.Patch("/", h.update)
			r.Delete("/", h.delete)
		})
	})

	// Per-student view: /groups/{groupID}/students/{studentID}/grades
	r.Get("/groups/{groupID}/students/{studentID}/grades", h.listStudent)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	gs, err := h.svc.ListGroupGrades(r.Context(), groupID, callerID, role)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]GradeResponse, len(gs))
	for i, g := range gs {
		out[i] = gradeToResponse(g)
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

	gs, err := h.svc.ListStudentGrades(r.Context(), groupID, studentID, callerID, role)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]GradeResponse, len(gs))
	for i, g := range gs {
		out[i] = gradeToResponse(g)
	}
	response.OK(w, out)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	if role != domain.RoleTeacher {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teachers can add grades")
		return
	}

	var req CreateGradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	g, err := h.svc.CreateGrade(r.Context(), callerID, groupID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.Created(w, gradeToResponse(g))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	gradeID, ok := parseUUID(w, chi.URLParam(r, "gradeID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	var req UpdateGradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	g, err := h.svc.UpdateGrade(r.Context(), gradeID, callerID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.OK(w, gradeToResponse(g))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	gradeID, ok := parseUUID(w, chi.URLParam(r, "gradeID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	if err := h.svc.DeleteGrade(r.Context(), gradeID, callerID); err != nil {
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
