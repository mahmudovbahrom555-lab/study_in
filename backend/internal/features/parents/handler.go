package parents

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

// RegisterRoutes mounts parent-specific endpoints.
//
// POST   /parent/children                          — link a child
// DELETE /parent/children/{studentID}              — unlink
// GET    /parent/children                          — list children
// GET    /parent/children/{studentID}/groups       — child's groups
// GET    /parent/children/{studentID}/groups/{groupID}/grades     — child's grades
// GET    /parent/children/{studentID}/groups/{groupID}/attendance — child's attendance
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/parent/children", func(r chi.Router) {
		r.Get("/", h.listChildren)
		r.Post("/", h.linkChild)

		r.Route("/{studentID}", func(r chi.Router) {
			r.Delete("/", h.unlinkChild)
			r.Get("/groups", h.childGroups)
			r.Get("/groups/{groupID}/grades", h.childGrades)
			r.Get("/groups/{groupID}/attendance", h.childAttendance)
		})
	})
}

func (h *Handler) linkChild(w http.ResponseWriter, r *http.Request) {
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())
	if role != domain.RoleParent {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only parents can link children")
		return
	}

	var req LinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	link, err := h.svc.LinkChild(r.Context(), callerID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.Created(w, ParentLinkResponse{
		ID:        link.ID,
		ParentID:  link.ParentID,
		StudentID: link.StudentID,
		CreatedAt: link.CreatedAt,
	})
}

func (h *Handler) unlinkChild(w http.ResponseWriter, r *http.Request) {
	studentID, ok := parseUUID(w, chi.URLParam(r, "studentID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	if err := h.svc.UnlinkChild(r.Context(), callerID, studentID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listChildren(w http.ResponseWriter, r *http.Request) {
	callerID := apimw.UserIDFromCtx(r.Context())

	links, err := h.svc.ListChildren(r.Context(), callerID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]ParentLinkResponse, len(links))
	for i, l := range links {
		out[i] = ParentLinkResponse{
			ID:        l.ID,
			ParentID:  l.ParentID,
			StudentID: l.StudentID,
			CreatedAt: l.CreatedAt,
		}
	}
	response.OK(w, out)
}

func (h *Handler) childGroups(w http.ResponseWriter, r *http.Request) {
	studentID, ok := parseUUID(w, chi.URLParam(r, "studentID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	groups, err := h.svc.ChildGroups(r.Context(), callerID, studentID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.OK(w, groups)
}

func (h *Handler) childGrades(w http.ResponseWriter, r *http.Request) {
	studentID, ok := parseUUID(w, chi.URLParam(r, "studentID"))
	if !ok {
		return
	}
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	grades, err := h.svc.ChildGrades(r.Context(), callerID, studentID, groupID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.OK(w, grades)
}

func (h *Handler) childAttendance(w http.ResponseWriter, r *http.Request) {
	studentID, ok := parseUUID(w, chi.URLParam(r, "studentID"))
	if !ok {
		return
	}
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	list, err := h.svc.ChildAttendance(r.Context(), callerID, studentID, groupID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.OK(w, list)
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
