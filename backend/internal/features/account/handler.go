package account

import (
	_ "embed"
	"net/http"

	"github.com/go-chi/chi/v5"

	apimw "github.com/mahmudovbahrom555-lab/study_in/backend/internal/middleware"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/response"
)

//go:embed delete_account.html
var deletePage []byte

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes — защищённые маршруты API (внутри /api/v1 с проверкой токена).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Delete("/me", h.deleteMe)
}

// RegisterPage — публичная страница удаления аккаунта без приложения
// (ссылка для Google Play). Страница сама входит по SMS-коду через API.
func (h *Handler) RegisterPage(r chi.Router) {
	r.Get("/account/delete", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(deletePage)
	})
}

func (h *Handler) deleteMe(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteAccount(r.Context(), apimw.UserIDFromCtx(r.Context())); err != nil {
		response.Error(w, err)
		return
	}
	response.NoContent(w)
}
