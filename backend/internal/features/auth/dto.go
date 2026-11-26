package auth

import (
	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// --- Request DTOs ---

type SendCodeRequest struct {
	Phone string `json:"phone" validate:"required,min=7,max=20"`
}

type VerifyRequest struct {
	Phone      string `json:"phone"      validate:"required"`
	Code       string `json:"code"       validate:"required,len=6"`
	DeviceInfo string `json:"device_info"`
	// PrivacyVersion — версия политики конфиденциальности, с которой человек
	// согласился на экране входа. Сохраняется при создании аккаунта.
	PrivacyVersion string `json:"privacy_version" validate:"max=20"`

	// Заполняет обработчик — для журнала согласий.
	IP        string `json:"-"`
	UserAgent string `json:"-"`
}

// GoogleSignInRequest — вход через Google: ID-токен, выданный Google клиенту.
type GoogleSignInRequest struct {
	IDToken        string `json:"id_token"        validate:"required"`
	DeviceInfo     string `json:"device_info"`
	PrivacyVersion string `json:"privacy_version" validate:"max=20"`

	IP        string `json:"-"`
	UserAgent string `json:"-"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type SetRoleRequest struct {
	Role domain.Role `json:"role" validate:"required,oneof=teacher student parent"`
}

type UpdateMeRequest struct {
	Name     *string          `json:"name"     validate:"omitempty,min=1,max=100"`
	Language *domain.Language `json:"language"` // проверяет сервис: domain.SupportedLanguages
	// @username репетитора (по нему ученики находят его открытые группы); "" — убрать.
	Username *string `json:"username"`
}

type AvatarPresetRequest struct {
	Preset string `json:"preset" validate:"required"`
}

type AddDeviceTokenRequest struct {
	Token    string `json:"token"    validate:"required"`
	Platform string `json:"platform" validate:"required,oneof=ios android"`
}

// --- Response DTOs ---

type UserResponse struct {
	ID           uuid.UUID    `json:"id"`
	Phone        string       `json:"phone"`
	Email        string       `json:"email,omitempty"`
	Name         string       `json:"name"`
	Username     *string      `json:"username"`
	Role         *domain.Role `json:"role"`
	AvatarURL    *string      `json:"avatar_url"` // подписанная ссылка на фото
	AvatarPreset *string      `json:"avatar_preset"`
	Language     string       `json:"language"`
	IsActive     bool         `json:"is_active"`
}

func userToResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:           u.ID,
		Phone:        deref(u.Phone),
		Email:        deref(u.Email),
		Name:         u.Name,
		Username:     u.Username,
		Role:         u.Role,
		AvatarPreset: u.AvatarPreset,
		Language:     string(u.Language),
		IsActive:     u.IsActive,
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type Response struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
	IsNewUser    bool         `json:"is_new_user"`
}
