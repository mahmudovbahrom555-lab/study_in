package chat

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	apimw "github.com/mahmudovbahrom555-lab/study_in/backend/internal/middleware"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/response"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes — маршруты чата (внутри /api/v1 с проверкой токена).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/groups/{groupID}/chat", func(r chi.Router) {
		r.Get("/messages", h.list)
		r.Post("/messages", h.send)
		r.Delete("/messages/{messageID}", h.delete)
		r.Post("/read", h.read)
		r.Patch("/settings", h.settings)
	})
	r.Post("/groups/{groupID}/members/{userID}/mute", h.mute(true))
	r.Post("/groups/{groupID}/members/{userID}/unmute", h.mute(false))
	r.Post("/groups/{groupID}/members/{userID}/ban", h.ban)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	groupID, ok := idParam(w, r, "groupID")
	if !ok {
		return
	}
	var before *uuid.UUID
	if b := r.URL.Query().Get("before"); b != "" {
		id, err := uuid.Parse(b)
		if err != nil {
			response.Error(w, domain.NewError("INVALID_ID", "invalid before", domain.ErrValidation))
			return
		}
		before = &id
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	ms, err := h.svc.List(r.Context(), groupID, apimw.UserIDFromCtx(r.Context()), before, limit)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, ms)
}

func (h *Handler) send(w http.ResponseWriter, r *http.Request) {
	groupID, ok := idParam(w, r, "groupID")
	if !ok {
		return
	}
	var req SendRequest
	if !decode(w, r, &req) {
		return
	}
	if req.ClientID == uuid.Nil {
		response.Error(w, domain.NewError("CLIENT_ID_REQUIRED", "client_id is required", domain.ErrValidation))
		return
	}
	m, err := h.svc.Send(r.Context(), groupID, apimw.UserIDFromCtx(r.Context()), req.ClientID, req.Body)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Created(w, m)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	groupID, ok := idParam(w, r, "groupID")
	if !ok {
		return
	}
	messageID, ok := idParam(w, r, "messageID")
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), groupID, messageID, apimw.UserIDFromCtx(r.Context())); err != nil {
		response.Error(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) read(w http.ResponseWriter, r *http.Request) {
	groupID, ok := idParam(w, r, "groupID")
	if !ok {
		return
	}
	var req ReadRequest
	if !decode(w, r, &req) {
		return
	}
	if err := h.svc.MarkRead(r.Context(), groupID, apimw.UserIDFromCtx(r.Context()), req.LastReadAt); err != nil {
		response.Error(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) settings(w http.ResponseWriter, r *http.Request) {
	groupID, ok := idParam(w, r, "groupID")
	if !ok {
		return
	}
	var req SettingsResponse
	if !decode(w, r, &req) {
		return
	}
	err := h.svc.UpdateSettings(r.Context(), groupID, apimw.UserIDFromCtx(r.Context()), req.TTLSeconds, domain.ChatMode(req.Mode))
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, req)
}

func (h *Handler) mute(muted bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID, ok := idParam(w, r, "groupID")
		if !ok {
			return
		}
		memberID, ok := idParam(w, r, "userID")
		if !ok {
			return
		}
		if err := h.svc.SetMuted(r.Context(), groupID, apimw.UserIDFromCtx(r.Context()), memberID, muted); err != nil {
			response.Error(w, err)
			return
		}
		response.NoContent(w)
	}
}

func (h *Handler) ban(w http.ResponseWriter, r *http.Request) {
	groupID, ok := idParam(w, r, "groupID")
	if !ok {
		return
	}
	memberID, ok := idParam(w, r, "userID")
	if !ok {
		return
	}
	if err := h.svc.Ban(r.Context(), groupID, apimw.UserIDFromCtx(r.Context()), memberID); err != nil {
		response.Error(w, err)
		return
	}
	response.NoContent(w)
}

func idParam(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		response.Error(w, domain.NewError("INVALID_ID", "invalid "+name, domain.ErrValidation))
		return uuid.Nil, false
	}
	return id, true
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		response.Error(w, domain.NewError("INVALID_JSON", "invalid JSON", domain.ErrValidation))
		return false
	}
	return true
}
