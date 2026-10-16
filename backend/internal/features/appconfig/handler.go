// Package appconfig отдаёт приложению настройки, которые меняются без релиза
// в стор: минимальную поддерживаемую версию и включённые фичи.
//
// Зачем: версии приложения из стора живут месяцами. Ломающее изменение API
// выпускаем вместе с поднятием MinVersion — старые версии просят обновиться,
// а не падают. Новую фичу выкатываем выключенной и включаем флагом.
package appconfig

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/response"
)

// Config — что сервер сообщает приложению.
type Config struct {
	// MinVersion — версия приложения (semver), ниже которой работать нельзя.
	MinVersion string
	// Features — ключи включённых фич (например "quizzes"). Неизвестные
	// приложению ключи оно игнорирует, отсутствующие считает выключенными.
	Features []string
}

type Handler struct{ cfg Config }

func NewHandler(cfg Config) *Handler { return &Handler{cfg: cfg} }

// RegisterRoutes — публичный маршрут: нужен до входа (экран обновления).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/app-config", h.get)
}

type configResponse struct {
	MinVersion string          `json:"min_version"`
	Features   map[string]bool `json:"features"`
}

func (h *Handler) get(w http.ResponseWriter, _ *http.Request) {
	features := make(map[string]bool, len(h.cfg.Features))
	for _, f := range h.cfg.Features {
		features[f] = true
	}
	response.OK(w, configResponse{MinVersion: h.cfg.MinVersion, Features: features})
}
