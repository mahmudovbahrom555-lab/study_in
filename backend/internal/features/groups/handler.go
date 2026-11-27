package groups

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

// RegisterRoutes mounts all group routes under the given router.
// All routes require a valid JWT (caller must pass apimw.Auth middleware upstream).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/groups", func(r chi.Router) {
		r.Post("/", h.createGroup)
		r.Get("/", h.listGroups)
		r.Post("/join", h.joinGroup)
		r.Post("/enroll", h.enrollChild)
		r.Get("/preview/{code}", h.previewGroup)
		r.Get("/search", h.searchGroups)
		r.Get("/consents", h.parentConsents)
		r.Post("/consents/{requestID}/approve", h.approveConsent)
		r.Post("/consents/{requestID}/decline", h.declineConsent)
		r.Get("/requests/mine", h.myJoinRequests)

		r.Route("/{groupID}", func(r chi.Router) {
			r.Get("/", h.getGroup)
			r.Patch("/", h.updateGroup)
			r.Delete("/", h.deleteGroup)
			r.Post("/archive", h.archiveGroup)

			r.Get("/members", h.listMembers)
			r.Delete("/members/{studentID}", h.removeMember)
			r.Post("/members/leave", h.leaveGroup)
			r.Patch("/members/{studentID}/payment", h.updatePayment)

			r.Get("/requests", h.listJoinRequests)
			r.Post("/requests/{requestID}/accept", h.decideJoinRequest(true))
			r.Post("/requests/{requestID}/decline", h.decideJoinRequest(false))
		})
	})
}

// POST /groups
func (h *Handler) createGroup(w http.ResponseWriter, r *http.Request) {
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	if role != domain.RoleTeacher {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teachers can create groups")
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	resp, err := h.svc.CreateGroup(r.Context(), callerID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// GET /groups
func (h *Handler) listGroups(w http.ResponseWriter, r *http.Request) {
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	groups, err := h.svc.ListMyGroups(r.Context(), callerID, role)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, groups)
}

// POST /groups/join
func (h *Handler) joinGroup(w http.ResponseWriter, r *http.Request) {
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	if role != domain.RoleStudent {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only students can join groups")
		return
	}

	var req JoinGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	resp, err := h.svc.JoinGroup(r.Context(), callerID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// POST /groups/enroll — родитель записывает своего ребёнка по коду группы.
func (h *Handler) enrollChild(w http.ResponseWriter, r *http.Request) {
	if apimw.RoleFromCtx(r.Context()) != domain.RoleParent {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only parents can enroll a child")
		return
	}
	var req EnrollChildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}
	resp, err := h.svc.EnrollChild(r.Context(), apimw.UserIDFromCtx(r.Context()), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /groups/preview/{code} — карточка группы по ссылке, до заявки.
func (h *Handler) previewGroup(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.PreviewGroup(r.Context(), chi.URLParam(r, "code"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /groups/search?q= — открытые группы по названию, предмету или @username.
func (h *Handler) searchGroups(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.SearchGroups(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /groups/consents — заявки детей, ждущие согласия родителя.
func (h *Handler) parentConsents(w http.ResponseWriter, r *http.Request) {
	if apimw.RoleFromCtx(r.Context()) != domain.RoleParent {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only parents")
		return
	}
	resp, err := h.svc.ParentConsents(r.Context(), apimw.UserIDFromCtx(r.Context()))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) approveConsent(w http.ResponseWriter, r *http.Request) {
	h.decideConsent(w, r, true)
}

func (h *Handler) declineConsent(w http.ResponseWriter, r *http.Request) {
	h.decideConsent(w, r, false)
}

// POST /groups/consents/{requestID}/approve|decline — решение родителя.
func (h *Handler) decideConsent(w http.ResponseWriter, r *http.Request, approve bool) {
	if apimw.RoleFromCtx(r.Context()) != domain.RoleParent {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only parents")
		return
	}
	requestID, ok := parseUUID(w, chi.URLParam(r, "requestID"))
	if !ok {
		return
	}
	if err := h.svc.DecideParentConsent(r.Context(), requestID, apimw.UserIDFromCtx(r.Context()), approve); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /groups/requests/mine — свои ожидающие заявки ученика.
func (h *Handler) myJoinRequests(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.MyJoinRequests(r.Context(), apimw.UserIDFromCtx(r.Context()))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /groups/{groupID}/requests — заявки в группу (репетитор).
func (h *Handler) listJoinRequests(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	resp, err := h.svc.ListJoinRequests(r.Context(), groupID, apimw.UserIDFromCtx(r.Context()))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// POST /groups/{groupID}/requests/{requestID}/accept|decline (репетитор).
func (h *Handler) decideJoinRequest(accept bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
		if !ok {
			return
		}
		requestID, ok := parseUUID(w, chi.URLParam(r, "requestID"))
		if !ok {
			return
		}
		callerID := apimw.UserIDFromCtx(r.Context())
		if err := h.svc.DecideJoinRequest(r.Context(), groupID, requestID, callerID, accept); err != nil {
			writeServiceError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// GET /groups/{groupID}
func (h *Handler) getGroup(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	resp, err := h.svc.GetGroup(r.Context(), groupID, callerID, role)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// PATCH /groups/{groupID}
func (h *Handler) updateGroup(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	if role != domain.RoleTeacher {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teachers can update groups")
		return
	}

	var req UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	resp, err := h.svc.UpdateGroup(r.Context(), groupID, callerID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// DELETE /groups/{groupID}
func (h *Handler) deleteGroup(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	if role != domain.RoleTeacher {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teachers can delete groups")
		return
	}

	if err := h.svc.DeleteGroup(r.Context(), groupID, callerID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /groups/{groupID}/archive
func (h *Handler) archiveGroup(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	if role != domain.RoleTeacher {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teachers can archive groups")
		return
	}

	resp, err := h.svc.ArchiveGroup(r.Context(), groupID, callerID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /groups/{groupID}/members
func (h *Handler) listMembers(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	members, err := h.svc.ListMembers(r.Context(), groupID, callerID, role)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, members)
}

// DELETE /groups/{groupID}/members/{studentID}
func (h *Handler) removeMember(w http.ResponseWriter, r *http.Request) {
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

	if role != domain.RoleTeacher {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teachers can remove members")
		return
	}

	if err := h.svc.RemoveMember(r.Context(), groupID, studentID, callerID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /groups/{groupID}/members/leave
func (h *Handler) leaveGroup(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	if err := h.svc.LeaveGroup(r.Context(), groupID, callerID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /groups/{groupID}/members/{studentID}/payment
func (h *Handler) updatePayment(w http.ResponseWriter, r *http.Request) {
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

	if role != domain.RoleTeacher {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teachers can update payment status")
		return
	}

	var req UpdatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	if err := h.svc.UpdatePayment(r.Context(), groupID, studentID, callerID, req); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

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
	// Ошибка с конкретным кодом (BANNED_FROM_GROUP…) — отдаём код как есть,
	// чтобы приложение показало точную причину.
	var de *domain.Error
	if errors.As(err, &de) && de.Code != "" {
		response.Error(w, err)
		return
	}
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "not found")
	case errors.Is(err, domain.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
	case errors.Is(err, domain.ErrAlreadyMember):
		writeError(w, http.StatusConflict, "ALREADY_MEMBER", "already a member of this group")
	case errors.Is(err, domain.ErrConflict):
		writeError(w, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, domain.ErrNotMember):
		writeError(w, http.StatusNotFound, "NOT_MEMBER", "not a member of this group")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
