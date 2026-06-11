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
	Language *domain.Language `json:"language" validate:"omitempty,oneof=uz uz-cyrl ru en"`
}

type AddDeviceTokenRequest struct {
	Token    string `json:"token"    validate:"required"`
	Platform string `json:"platform" validate:"required,oneof=ios android"`
}

// --- Response DTOs ---

type UserResponse struct {
	ID        uuid.UUID    `json:"id"`
	Phone     string       `json:"phone"`
	Name      string       `json:"name"`
	Role      *domain.Role `json:"role"`
	AvatarURL *string      `json:"avatar_url"`
	Language  string       `json:"language"`
	IsActive  bool         `json:"is_active"`
}

func userToResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Phone:     u.Phone,
		Name:      u.Name,
		Role:      u.Role,
		AvatarURL: u.AvatarURL,
		Language:  string(u.Language),
		IsActive:  u.IsActive,
	}
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
	IsNewUser    bool         `json:"is_new_user"`
}
