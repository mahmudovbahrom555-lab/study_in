package domain

import (
	"regexp"
	"strings"
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

// Language — язык пользователя: на нём сервер пишет push и SMS.
type Language string

const (
	LanguageUz     Language = "uz"
	LanguageUzCyrl Language = "uz-cyrl"
	LanguageRu     Language = "ru"
	LanguageEn     Language = "en"
)

// SupportedLanguages — единственный список языков на сервере. Новый язык —
// одна строка здесь (в БД ограничения нет, миграция не нужна).
var SupportedLanguages = []Language{LanguageUz, LanguageUzCyrl, LanguageRu, LanguageEn}

func (l Language) Valid() bool {
	for _, s := range SupportedLanguages {
		if l == s {
			return true
		}
	}
	return false
}

type User struct {
	ID             uuid.UUID  `db:"id"`
	Phone          *string    `db:"phone"` // nil у вошедших через Google
	Email          *string    `db:"email"`
	GoogleSub      *string    `db:"google_sub"`
	Name           string     `db:"name"`
	Username       *string    `db:"username"` // @username репетитора, у других ролей nil
	Role           *Role      `db:"role"`
	AvatarURL      *string    `db:"avatar_url"` // ключ фото в хранилище, не ссылка
	AvatarPreset   *string    `db:"avatar_preset"`
	Language       Language   `db:"language"`
	OrganizationID *uuid.UUID `db:"organization_id"`
	IsActive       bool       `db:"is_active"`
	ChatBanned     bool       `db:"chat_banned"` // запрет писать во всех чатах (по жалобе)
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
	DeletedAt      *time.Time `db:"deleted_at"`
}

// usernameRe — @username репетитора: латиница в нижнем регистре, цифры и _,
// 4–32 символа, первая — буква.
var usernameRe = regexp.MustCompile(`^[a-z][a-z0-9_]{3,31}$`)

func ValidUsername(s string) bool { return usernameRe.MatchString(s) }

// NormalizeUsername убирает пробелы и ведущий @ и приводит к нижнему регистру.
func NormalizeUsername(s string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(s), "@"))
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

// AvatarPresets — аватары-животные. Ученики выбирают только из них: фото
// детей в системе не храним (политика Google Play для семей, модерация).
var AvatarPresets = []string{
	"fox", "panda", "koala", "lion", "tiger", "frog",
	"monkey", "penguin", "bear", "rabbit", "cat", "dog",
}

func ValidAvatarPreset(p string) bool {
	for _, v := range AvatarPresets {
		if v == p {
			return true
		}
	}
	return false
}
