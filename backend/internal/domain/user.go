package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleTeacher Role = "teacher"
	RoleStudent Role = "student"
	RoleParent  Role = "parent"
)

func (r Role) Valid() bool {
	return r == RoleTeacher || r == RoleStudent || r == RoleParent
}

type Language string

const (
	LanguageUz     Language = "uz"
	LanguageUzCyrl Language = "uz-cyrl"
	LanguageRu     Language = "ru"
	LanguageEn     Language = "en"
)

func (l Language) Valid() bool {
	return l == LanguageUz || l == LanguageUzCyrl || l == LanguageRu || l == LanguageEn
}

type User struct {
	ID             uuid.UUID  `db:"id"`
	Phone          string     `db:"phone"`
	Name           string     `db:"name"`
	Role           *Role      `db:"role"`
	AvatarURL      *string    `db:"avatar_url"`
	Language       Language   `db:"language"`
	OrganizationID *uuid.UUID `db:"organization_id"`
	IsActive       bool       `db:"is_active"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
	DeletedAt      *time.Time `db:"deleted_at"`
}

func (u *User) HasRole() bool {
	return u.Role != nil
}

type VerificationCode struct {
	ID        uuid.UUID `db:"id"`
	Phone     string    `db:"phone"`
	Code      string    `db:"code"`
	Attempts  int       `db:"attempts"`
	ExpiresAt time.Time `db:"expires_at"`
	Used      bool      `db:"used"`
	CreatedAt time.Time `db:"created_at"`
}

type RefreshToken struct {
	ID         uuid.UUID  `db:"id"`
	UserID     uuid.UUID  `db:"user_id"`
	TokenHash  string     `db:"token_hash"`
	DeviceInfo *string    `db:"device_info"`
	ExpiresAt  time.Time  `db:"expires_at"`
	RevokedAt  *time.Time `db:"revoked_at"`
	CreatedAt  time.Time  `db:"created_at"`
}

type DeviceToken struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Token     string    `db:"token"`
	Platform  string    `db:"platform"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
