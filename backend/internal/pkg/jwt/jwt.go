// Package jwt обёртывает golang-jwt для выпуска и проверки access-токенов.
package jwt

import (
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// Claims — payload access-токена.
type Claims struct {
	UserID uuid.UUID   `json:"uid"`
	Role   domain.Role `json:"role"`
	gojwt.RegisteredClaims
}

// Manager выпускает и проверяет JWT.
type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

// IssueAccess создаёт подписанный access-токен.
func (m *Manager) IssueAccess(userID uuid.UUID, role domain.Role) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(m.accessTTL)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.accessSecret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// ParseAccess проверяет и разбирает access-токен.
func (m *Manager) ParseAccess(tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.accessSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse access token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// RefreshTTL возвращает TTL refresh-токена (нужен сервису при создании записи в БД).
func (m *Manager) RefreshTTL() time.Duration {
	return m.refreshTTL
}
