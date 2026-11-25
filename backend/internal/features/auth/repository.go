// Package auth реализует фичу авторизации: SMS OTP, JWT, профиль пользователя.
package auth

import (
	"context"
	"io"
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
	GetUserByGoogleSub(ctx context.Context, sub string) (*domain.User, error)
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

// GoogleIdentity — проверенные данные из Google ID-токена.
type GoogleIdentity struct {
	Sub   string
	Email string
	Name  string
}

// GoogleVerifier проверяет подпись, срок и audience Google ID-токена.
type GoogleVerifier interface {
	Verify(ctx context.Context, idToken string) (*GoogleIdentity, error)
}

// Limiter — абстракция rate-limiter для unit-тестируемости сервиса.
type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) error
}

// AvatarStorage — хранилище фото профиля (то же, что у файлов заданий).
type AvatarStorage interface {
	PutObject(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	RemoveObject(ctx context.Context, key string) error
	PresignedGetURL(ctx context.Context, objectKey, filename string) (string, error)
}
