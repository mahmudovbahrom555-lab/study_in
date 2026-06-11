package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	apimw "github.com/mahmudovbahrom555-lab/study_in/backend/internal/middleware"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/jwt"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/response"
)

// Handler — HTTP handlers для auth-фичи.
type Handler struct {
	service    *Service
	jwtManager *jwt.Manager
	validate   *validator.Validate
}

func NewHandler(service *Service, jwtManager *jwt.Manager) *Handler {
	return &Handler{
		service:    service,
		jwtManager: jwtManager,
		validate:   validator.New(),
	}
}

// RegisterRoutes добавляет все auth и me маршруты в роутер.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Public
	r.Post("/auth/send-code", h.sendCode)
	r.Post("/auth/verify", h.verify)
	r.Post("/auth/refresh", h.refresh)

	// Protected
	r.Group(func(r chi.Router) {
		r.Use(apimw.Auth(h.jwtManager))
		r.Post("/auth/logout", h.logout)
		r.Post("/auth/role", h.setRole)

		r.Get("/me", h.getMe)
		r.Patch("/me", h.updateMe)
		r.Post("/me/avatar", h.uploadAvatar)
		r.Post("/me/device-token", h.addDeviceToken)
	})
}

func (h *Handler) sendCode(w http.ResponseWriter, r *http.Request) {
	var req SendCodeRequest
	if err := decode(r, &req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, map[string]any{"parse": err.Error()})
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, validationDetails(err))
		return
	}
	if err := h.service.SendCode(r.Context(), &req); err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, map[string]string{"message": "Код отправлен"})
}

func (h *Handler) verify(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := decode(r, &req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, map[string]any{"parse": err.Error()})
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, validationDetails(err))
		return
	}
	result, err := h.service.Verify(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, result)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := decode(r, &req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, map[string]any{"parse": err.Error()})
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, validationDetails(err))
		return
	}
	result, err := h.service.Refresh(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, result)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	userID := apimw.UserIDFromCtx(r.Context())
	var req LogoutRequest
	if err := decode(r, &req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, map[string]any{"parse": err.Error()})
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, validationDetails(err))
		return
	}
	if err := h.service.Logout(r.Context(), userID, &req); err != nil {
		response.Error(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) setRole(w http.ResponseWriter, r *http.Request) {
	userID := apimw.UserIDFromCtx(r.Context())
	var req SetRoleRequest
	if err := decode(r, &req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, map[string]any{"parse": err.Error()})
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, validationDetails(err))
		return
	}
	user, err := h.service.SetRole(r.Context(), userID, &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, user)
}

func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	userID := apimw.UserIDFromCtx(r.Context())
	user, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, user)
}

func (h *Handler) updateMe(w http.ResponseWriter, r *http.Request) {
	userID := apimw.UserIDFromCtx(r.Context())
	var req UpdateMeRequest
	if err := decode(r, &req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, map[string]any{"parse": err.Error()})
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, validationDetails(err))
		return
	}
	user, err := h.service.UpdateMe(r.Context(), userID, &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, user)
}

func (h *Handler) uploadAvatar(w http.ResponseWriter, _ *http.Request) {
	// MinIO интеграция — Этап 4 (см. DECISIONS.md)
	http.Error(w, `{"error":{"code":"NOT_IMPLEMENTED","message":"Avatar upload available in next release"}}`, http.StatusNotImplemented)
}

func (h *Handler) addDeviceToken(w http.ResponseWriter, r *http.Request) {
	userID := apimw.UserIDFromCtx(r.Context())
	var req AddDeviceTokenRequest
	if err := decode(r, &req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, map[string]any{"parse": err.Error()})
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, validationDetails(err))
		return
	}
	if err := h.service.AddDeviceToken(r.Context(), userID, &req); err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, map[string]string{"message": "Device token saved"})
}

// --- helpers ---

func decode(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func validationDetails(err error) map[string]any {
	details := make(map[string]any)
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			details[fe.Field()] = fe.Tag()
		}
	}
	return details
}
