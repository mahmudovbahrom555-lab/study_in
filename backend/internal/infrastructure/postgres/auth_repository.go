package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// AuthRepository реализует auth.Repository поверх PostgreSQL.
type AuthRepository struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// --- Users ---

func (r *AuthRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, phone, name, role, language, organization_id, is_active, created_at, updated_at)
		VALUES (:id, :phone, :name, :role, :language, :organization_id, :is_active, :created_at, :updated_at)`
	if _, err := r.db.NamedExecContext(ctx, query, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user,
		`SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

func (r *AuthRepository) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user,
		`SELECT * FROM users WHERE phone = $1 AND deleted_at IS NULL`, phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by phone: %w", err)
	}
	return &user, nil
}

func (r *AuthRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET name = :name, role = :role, language = :language,
		                 avatar_url = :avatar_url, updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`
	if _, err := r.db.NamedExecContext(ctx, query, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// --- Verification codes ---

func (r *AuthRepository) CreateVerificationCode(ctx context.Context, phone, code string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO verification_codes (id, phone, code, expires_at) VALUES ($1, $2, $3, $4)`,
		uuid.New(), phone, code, expiresAt)
	if err != nil {
		return fmt.Errorf("create verification code: %w", err)
	}
	return nil
}

func (r *AuthRepository) GetLatestVerificationCode(ctx context.Context, phone string) (*domain.VerificationCode, error) {
	var vc domain.VerificationCode
	err := r.db.GetContext(ctx, &vc,
		`SELECT * FROM verification_codes WHERE phone = $1 ORDER BY created_at DESC LIMIT 1`, phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get verification code: %w", err)
	}
	return &vc, nil
}

func (r *AuthRepository) MarkCodeUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE verification_codes SET used = TRUE WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("mark code used: %w", err)
	}
	return nil
}

func (r *AuthRepository) IncrementCodeAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE verification_codes SET attempts = attempts + 1 WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("increment attempts: %w", err)
	}
	return nil
}

// --- Refresh tokens ---

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, device_info, expires_at, created_at)
		VALUES (:id, :user_id, :token_hash, :device_info, :expires_at, :created_at)`
	if _, err := r.db.NamedExecContext(ctx, query, token); err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

func (r *AuthRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := r.db.GetContext(ctx, &rt,
		`SELECT * FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return &rt, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = $1 WHERE token_hash = $2 AND revoked_at IS NULL`,
		now, tokenHash)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (r *AuthRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`,
		now, userID)
	if err != nil {
		return fmt.Errorf("revoke all tokens: %w", err)
	}
	return nil
}

// --- Device tokens ---

func (r *AuthRepository) UpsertDeviceToken(ctx context.Context, userID uuid.UUID, token, platform string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO device_tokens (id, user_id, token, platform, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (token) DO UPDATE SET user_id = $2, platform = $4, updated_at = NOW()`,
		uuid.New(), userID, token, platform)
	if err != nil {
		return fmt.Errorf("upsert device token: %w", err)
	}
	return nil
}
