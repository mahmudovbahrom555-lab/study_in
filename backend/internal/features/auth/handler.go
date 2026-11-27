package auth

import (
	"encoding/json"
	"io"
	"net"
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
	r.Post("/auth/google", h.googleSignIn)
	r.Post("/auth/refresh", h.refresh)

	// Protected
	r.Group(func(r chi.Router) {
		r.Use(apimw.Auth(h.jwtManager))
		r.Post("/auth/logout", h.logout)
		r.Post("/auth/role", h.setRole)

		r.Get("/me", h.getMe)
		r.Patch("/me", h.updateMe)
		r.Post("/me/avatar", h.uploadAvatar)
		r.Put("/me/avatar/preset", h.setAvatarPreset)
		r.Delete("/me/avatar", h.deleteAvatar)
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
	resp, err := h.service.SendCode(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, resp)
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
	req.IP = clientIP(r)
	req.UserAgent = r.UserAgent()
	result, err := h.service.Verify(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, result)
}

func (h *Handler) googleSignIn(w http.ResponseWriter, r *http.Request) {
	var req GoogleSignInRequest
	if err := decode(r, &req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, map[string]any{"parse": err.Error()})
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, validationDetails(err))
		return
	}
	req.IP = clientIP(r)
	req.UserAgent = r.UserAgent()
	result, err := h.service.GoogleSignIn(r.Context(), &req)
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

// POST /me/avatar — фото профиля (multipart, поле file). Только взрослые.
func (h *Handler) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxAvatarSize+(1<<20))
	file, _, err := r.FormFile("file")
	if err != nil {
		response.Error(w, domain.NewError("AVATAR_TOO_LARGE", "Фото больше 2 МБ или не передано", domain.ErrValidation))
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, MaxAvatarSize+1))
	if err != nil {
		response.Error(w, err)
		return
	}
	user, err := h.service.UploadAvatar(r.Context(), apimw.UserIDFromCtx(r.Context()), data)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, user)
}

// PUT /me/avatar/preset — аватар-животное.
func (h *Handler) setAvatarPreset(w http.ResponseWriter, r *http.Request) {
	var req AvatarPresetRequest
	if err := decode(r, &req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, map[string]any{"parse": err.Error()})
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.ErrorWithDetails(w, domain.ErrValidation, validationDetails(err))
		return
	}
	user, err := h.service.SetAvatarPreset(r.Context(), apimw.UserIDFromCtx(r.Context()), req.Preset)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, user)
}

// DELETE /me/avatar — убрать фото и животное.
func (h *Handler) deleteAvatar(w http.ResponseWriter, r *http.Request) {
	user, err := h.service.DeleteAvatar(r.Context(), apimw.UserIDFromCtx(r.Context()))
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, user)
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

// clientIP — адрес клиента без порта (middleware RealIP уже учёл X-Real-IP / X-Forwarded-For).
func clientIP(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

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
