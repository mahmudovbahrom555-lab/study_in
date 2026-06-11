// Package auth реализует фичу авторизации: SMS OTP, JWT, профиль пользователя.
package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// Repository — интерфейс хранилища данных для auth-фичи.
// Реализуется в infrastructure/postgres.
type Repository interface {
	// Users
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error

	// Verification codes
	CreateVerificationCode(ctx context.Context, phone, code string, expiresAt time.Time) error
	GetLatestVerificationCode(ctx context.Context, phone string) (*domain.VerificationCode, error)
	MarkCodeUsed(ctx context.Context, id uuid.UUID) error
	IncrementCodeAttempts(ctx context.Context, id uuid.UUID) error

	// Refresh tokens
	CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error

	// Device tokens
	UpsertDeviceToken(ctx context.Context, userID uuid.UUID, token, platform string) error
}

// Limiter — абстракция rate-limiter для unit-тестируемости сервиса.
type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) error
}
