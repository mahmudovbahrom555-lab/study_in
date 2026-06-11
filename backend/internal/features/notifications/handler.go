package notifications

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	apimw "github.com/mahmudovbahrom555-lab/study_in/backend/internal/middleware"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes mounts notification endpoints.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/notifications", func(r chi.Router) {
		r.Get("/", h.list)
		r.Get("/unread-count", h.unreadCount)
		r.Post("/read-all", h.markAllRead)
		r.Post("/{id}/read", h.markRead)
	})
	r.Route("/devices", func(r chi.Router) {
		r.Post("/", h.registerDevice)
		r.Delete("/{token}", h.unregisterDevice)
	})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID := apimw.UserIDFromCtx(r.Context())
	ns, err := h.svc.List(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, toDTOs(ns))
}

func (h *Handler) unreadCount(w http.ResponseWriter, r *http.Request) {
	userID := apimw.UserIDFromCtx(r.Context())
	count, err := h.svc.UnreadCount(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, map[string]int{"count": count})
}

func (h *Handler) markAllRead(w http.ResponseWriter, r *http.Request) {
	userID := apimw.UserIDFromCtx(r.Context())
	if err := h.svc.MarkAllRead(r.Context(), userID); err != nil {
		response.Error(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) markRead(w http.ResponseWriter, r *http.Request) {
	userID := apimw.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, domain.NewError("INVALID_ID", "invalid notification id", domain.ErrValidation))
		return
	}
	if err := h.svc.MarkRead(r.Context(), id, userID); err != nil {
		response.Error(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) registerDevice(w http.ResponseWriter, r *http.Request) {
	userID := apimw.UserIDFromCtx(r.Context())
	var req struct {
		Token    string `json:"token"`
		Platform string `json:"platform"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		response.Error(w, domain.NewError("INVALID_REQUEST", "token is required", domain.ErrValidation))
		return
	}
	if err := h.svc.RegisterDevice(r.Context(), userID, req.Token, req.Platform); err != nil {
		response.Error(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) unregisterDevice(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if err := h.svc.UnregisterDevice(r.Context(), token); err != nil {
		response.Error(w, err)
		return
	}
	response.NoContent(w)
}
