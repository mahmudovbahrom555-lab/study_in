package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/sms"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/jwt"
)

const (
	codeLength    = 6
	codeTTL       = 10 * time.Minute
	maxAttempts   = 3
	smsRateLimit  = 1
	smsRateWindow = time.Minute
)

// Service — бизнес-логика авторизации.
type Service struct {
	repo    Repository
	sms     sms.Sender
	jwt     *jwt.Manager
	limiter Limiter
}

func NewService(repo Repository, smsSender sms.Sender, jwtManager *jwt.Manager, limiter Limiter) *Service {
	return &Service{
		repo:    repo,
		sms:     smsSender,
		jwt:     jwtManager,
		limiter: limiter,
	}
}

// SendCode отправляет SMS с кодом подтверждения.
func (s *Service) SendCode(ctx context.Context, req *SendCodeRequest) error {
	key := fmt.Sprintf("sms:%s", req.Phone)
	if err := s.limiter.Allow(ctx, key, smsRateLimit, smsRateWindow); err != nil {
		return err
	}

	code, err := generateCode(codeLength)
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	expiresAt := time.Now().Add(codeTTL)
	if err := s.repo.CreateVerificationCode(ctx, req.Phone, code, expiresAt); err != nil {
		return fmt.Errorf("save code: %w", err)
	}

	message := fmt.Sprintf("RepetApp: %s — ваш код подтверждения. Никому не сообщайте!", code)
	if err := s.sms.Send(ctx, req.Phone, message); err != nil {
		return fmt.Errorf("send sms: %w", err)
	}
	return nil
}

// Verify проверяет код и выдаёт токены. Создаёт пользователя при первом входе.
func (s *Service) Verify(ctx context.Context, req *VerifyRequest) (*AuthResponse, error) {
	vc, err := s.repo.GetLatestVerificationCode(ctx, req.Phone)
	if err != nil {
		return nil, fmt.Errorf("get code: %w", err)
	}
	if vc == nil || vc.Used || time.Now().After(vc.ExpiresAt) {
		return nil, domain.NewError("CODE_EXPIRED", "Код истёк или не найден", domain.ErrValidation)
	}
	if vc.Attempts >= maxAttempts {
		return nil, domain.NewError("TOO_MANY_ATTEMPTS", "Превышено количество попыток", domain.ErrRateLimit)
	}
	if vc.Code != req.Code {
		_ = s.repo.IncrementCodeAttempts(ctx, vc.ID)
		return nil, domain.NewError("INVALID_CODE", "Неверный код", domain.ErrValidation)
	}
	if err := s.repo.MarkCodeUsed(ctx, vc.ID); err != nil {
		return nil, fmt.Errorf("mark code used: %w", err)
	}

	user, err := s.repo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	isNewUser := user == nil
	if isNewUser {
		user = &domain.User{
			ID:        uuid.New(),
			Phone:     req.Phone,
			Name:      "",
			Language:  domain.LanguageUz,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
	}

	accessToken, refreshToken, err := s.issueTokenPair(ctx, user, req.DeviceInfo)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userToResponse(user),
		IsNewUser:    isNewUser,
	}, nil
}

// Refresh обменивает refresh-токен на новую пару токенов.
func (s *Service) Refresh(ctx context.Context, req *RefreshRequest) (*AuthResponse, error) {
	tokenHash := hashToken(req.RefreshToken)

	rt, err := s.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	if rt == nil || rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
		return nil, domain.NewError("TOKEN_INVALID", "Недействительный refresh-токен", domain.ErrUnauthorized)
	}

	user, err := s.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil || !user.IsActive {
		return nil, domain.ErrUnauthorized
	}

	// Ротация токена: отзываем старый, выпускаем новый
	if err := s.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("revoke token: %w", err)
	}

	deviceInfo := ""
	if rt.DeviceInfo != nil {
		deviceInfo = *rt.DeviceInfo
	}
	accessToken, refreshToken, err := s.issueTokenPair(ctx, user, deviceInfo)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userToResponse(user),
	}, nil
}

// Logout отзывает refresh-токен.
func (s *Service) Logout(ctx context.Context, userID uuid.UUID, req *LogoutRequest) error {
	tokenHash := hashToken(req.RefreshToken)
	if err := s.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	return nil
}

// SetRole устанавливает роль при первом входе.
func (s *Service) SetRole(ctx context.Context, userID uuid.UUID, req *SetRoleRequest) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrNotFound
	}
	if user.HasRole() {
		return nil, domain.NewError("ROLE_ALREADY_SET", "Роль уже установлена", domain.ErrConflict)
	}

	user.Role = &req.Role
	user.UpdatedAt = time.Now()
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	resp := userToResponse(user)
	return &resp, nil
}

// GetMe возвращает профиль текущего пользователя.
func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrNotFound
	}
	resp := userToResponse(user)
	return &resp, nil
}

// UpdateMe обновляет имя и/или язык профиля.
func (s *Service) UpdateMe(ctx context.Context, userID uuid.UUID, req *UpdateMeRequest) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrNotFound
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Language != nil {
		user.Language = *req.Language
	}
	user.UpdatedAt = time.Now()

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	resp := userToResponse(user)
	return &resp, nil
}

// AddDeviceToken сохраняет FCM-токен устройства.
func (s *Service) AddDeviceToken(ctx context.Context, userID uuid.UUID, req *AddDeviceTokenRequest) error {
	return s.repo.UpsertDeviceToken(ctx, userID, req.Token, req.Platform)
}

// --- helpers ---

func (s *Service) issueTokenPair(ctx context.Context, user *domain.User, deviceInfo string) (string, string, error) {
	role := domain.Role("")
	if user.Role != nil {
		role = *user.Role
	}

	accessToken, err := s.jwt.IssueAccess(user.ID, role)
	if err != nil {
		return "", "", fmt.Errorf("issue access token: %w", err)
	}

	rawToken, err := generateRawToken()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	var di *string
	if deviceInfo != "" {
		di = &deviceInfo
	}

	rt := &domain.RefreshToken{
		ID:         uuid.New(),
		UserID:     user.ID,
		TokenHash:  hashToken(rawToken),
		DeviceInfo: di,
		ExpiresAt:  time.Now().Add(s.jwt.RefreshTTL()),
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreateRefreshToken(ctx, rt); err != nil {
		return "", "", fmt.Errorf("save refresh token: %w", err)
	}

	return accessToken, rawToken, nil
}

func generateCode(length int) (string, error) {
	digits := "0123456789"
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[n.Int64()]
	}
	return string(code), nil
}

func generateRawToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken вычисляет SHA-256 хэш токена для хранения в БД.
// Решение: SHA-256 достаточен для случайных токенов (см. DECISIONS.md).
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
