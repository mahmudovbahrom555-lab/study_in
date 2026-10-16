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
		response.Error(w, err)
		return
	}
	response.OK(w, toResponses(gs))
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
		response.Error(w, err)
		return
	}
	response.OK(w, toResponses(gs))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	if apimw.RoleFromCtx(r.Context()) != domain.RoleTeacher {
		response.Error(w, domain.ErrForbidden)
		return
	}

	var req CreateGradeRequest
	if !h.decode(w, r, &req) {
		return
	}
	g, err := h.svc.CreateGrade(r.Context(), callerID, groupID, req)
	if err != nil {
		response.Error(w, err)
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
	if !h.decode(w, r, &req) {
		return
	}
	g, err := h.svc.UpdateGrade(r.Context(), gradeID, callerID, req)
	if err != nil {
		response.Error(w, err)
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
		response.Error(w, err)
		return
	}
	response.NoContent(w)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func toResponses(gs []*domain.Grade) []GradeResponse {
	out := make([]GradeResponse, len(gs))
	for i, g := range gs {
		out[i] = gradeToResponse(g)
	}
	return out
}

func parseUUID(w http.ResponseWriter, s string) (uuid.UUID, bool) {
	id, err := uuid.Parse(s)
	if err != nil {
		response.Error(w, domain.NewError("INVALID_ID", "invalid id", domain.ErrValidation))
		return uuid.Nil, false
	}
	return id, true
}

// decode читает JSON и проверяет теги validate; при ошибке сам отвечает 400.
func (h *Handler) decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		response.Error(w, domain.NewError("INVALID_JSON", "invalid JSON", domain.ErrValidation))
		return false
	}
	if err := h.validate.Struct(dst); err != nil {
		details := map[string]any{}
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, fe := range ve {
				details[fe.Field()] = fe.Tag()
			}
		}
		response.ErrorWithDetails(w, domain.ErrValidation, details)
		return false
	}
	return true
}
