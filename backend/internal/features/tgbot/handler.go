package tgbot

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc    *Service
	secret string
}

func NewHandler(svc *Service, secret string) *Handler {
	return &Handler{svc: svc, secret: secret}
}

// RegisterRoutes: вебхук публичный — Telegram подтверждает себя секретом
// в заголовке, без него запрос отклоняется.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/telegram/webhook", h.webhook)
}

func (h *Handler) webhook(w http.ResponseWriter, r *http.Request) {
	got := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	if h.secret == "" || subtle.ConstantTimeCompare([]byte(got), []byte(h.secret)) != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var u Update
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Ошибку только пишем в журнал: иначе Telegram будет слать то же обновление снова.
	if err := h.svc.HandleUpdate(r.Context(), u); err != nil {
		slog.Warn("telegram update", slog.String("error", err.Error()))
	}
	w.WriteHeader(http.StatusOK)
}
