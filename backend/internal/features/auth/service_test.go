package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/auth"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/sms"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/jwt"
)

// ─── Mocks ───────────────────────────────────────────────────────────────────

type mockRepo struct {
	users         map[string]*domain.User
	codes         map[string]*domain.VerificationCode
	tokens        map[string]*domain.RefreshToken
	attempts      map[uuid.UUID]int
	markedUsed    map[uuid.UUID]bool
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		users:      make(map[string]*domain.User),
		codes:      make(map[string]*domain.VerificationCode),
		tokens:     make(map[string]*domain.RefreshToken),
		attempts:   make(map[uuid.UUID]int),
		markedUsed: make(map[uuid.UUID]bool),
	}
}

func (m *mockRepo) CreateUser(_ context.Context, u *domain.User) error {
	m.users[u.Phone] = u
	return nil
}
func (m *mockRepo) GetUserByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}
func (m *mockRepo) GetUserByPhone(_ context.Context, phone string) (*domain.User, error) {
	return m.users[phone], nil
}
func (m *mockRepo) UpdateUser(_ context.Context, u *domain.User) error {
	m.users[u.Phone] = u
	return nil
}
func (m *mockRepo) CreateVerificationCode(_ context.Context, phone, code string, exp time.Time) error {
	m.codes[phone] = &domain.VerificationCode{ID: uuid.New(), Phone: phone, Code: code, ExpiresAt: exp}
	return nil
}
func (m *mockRepo) GetLatestVerificationCode(_ context.Context, phone string) (*domain.VerificationCode, error) {
	vc := m.codes[phone]
	if vc == nil {
		return nil, nil
	}
	copy := *vc
	copy.Used = m.markedUsed[vc.ID]
	copy.Attempts = m.attempts[vc.ID]
	return &copy, nil
}
func (m *mockRepo) MarkCodeUsed(_ context.Context, id uuid.UUID) error {
	m.markedUsed[id] = true
	return nil
}
func (m *mockRepo) IncrementCodeAttempts(_ context.Context, id uuid.UUID) error {
	m.attempts[id]++
	return nil
}
func (m *mockRepo) CreateRefreshToken(_ context.Context, rt *domain.RefreshToken) error {
	m.tokens[rt.TokenHash] = rt
	return nil
}
func (m *mockRepo) GetRefreshToken(_ context.Context, hash string) (*domain.RefreshToken, error) {
	return m.tokens[hash], nil
}
func (m *mockRepo) RevokeRefreshToken(_ context.Context, hash string) error {
	if rt, ok := m.tokens[hash]; ok {
		now := time.Now()
		rt.RevokedAt = &now
	}
	return nil
}
func (m *mockRepo) RevokeAllUserTokens(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockRepo) UpsertDeviceToken(_ context.Context, _ uuid.UUID, _, _ string) error {
	return nil
}

type mockSMS struct{ sent []string }

func (m *mockSMS) Send(_ context.Context, phone, _ string) error {
	m.sent = append(m.sent, phone)
	return nil
}

// Compile-time check: *mockSMS implements sms.Sender.
var _ sms.Sender = (*mockSMS)(nil)

type allowLimiter struct{}

func (allowLimiter) Allow(_ context.Context, _ string, _ int, _ time.Duration) error { return nil }

type denyLimiter struct{}

func (denyLimiter) Allow(_ context.Context, _ string, _ int, _ time.Duration) error {
	return domain.ErrRateLimit
}

// ─── Builder ─────────────────────────────────────────────────────────────────

func buildService(repo *mockRepo, smsMock *mockSMS, limiter auth.Limiter) *auth.Service {
	jwtMgr := jwt.NewManager(
		"test_access_secret_at_least_32ch",
		"test_refresh_secret_at_least_32ch",
		15*time.Minute,
		30*24*time.Hour,
	)
	return auth.NewService(repo, smsMock, jwtMgr, limiter)
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestSendCode_Success(t *testing.T) {
	repo, smsMock := newMockRepo(), &mockSMS{}
	svc := buildService(repo, smsMock, allowLimiter{})

	err := svc.SendCode(context.Background(), &auth.SendCodeRequest{Phone: "+998901234567"})

	require.NoError(t, err)
	assert.Len(t, smsMock.sent, 1)
}

func TestSendCode_RateLimited(t *testing.T) {
	repo, smsMock := newMockRepo(), &mockSMS{}
	svc := buildService(repo, smsMock, denyLimiter{})

	err := svc.SendCode(context.Background(), &auth.SendCodeRequest{Phone: "+998901234567"})

	assert.ErrorIs(t, err, domain.ErrRateLimit)
	assert.Empty(t, smsMock.sent)
}

func TestVerify_NewUser(t *testing.T) {
	repo, smsMock := newMockRepo(), &mockSMS{}
	svc := buildService(repo, smsMock, allowLimiter{})
	phone := "+998901234567"

	_ = svc.SendCode(context.Background(), &auth.SendCodeRequest{Phone: phone})
	code := repo.codes[phone].Code

	result, err := svc.Verify(context.Background(), &auth.VerifyRequest{Phone: phone, Code: code})

	require.NoError(t, err)
	assert.True(t, result.IsNewUser)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Nil(t, result.User.Role)
}

func TestVerify_ExistingUser(t *testing.T) {
	repo, smsMock := newMockRepo(), &mockSMS{}
	svc := buildService(repo, smsMock, allowLimiter{})
	phone := "+998901234567"
	role := domain.RoleTeacher
	repo.users[phone] = &domain.User{
		ID: uuid.New(), Phone: phone, Role: &role,
		Language: domain.LanguageUz, IsActive: true,
	}

	_ = svc.SendCode(context.Background(), &auth.SendCodeRequest{Phone: phone})
	code := repo.codes[phone].Code
	result, err := svc.Verify(context.Background(), &auth.VerifyRequest{Phone: phone, Code: code})

	require.NoError(t, err)
	assert.False(t, result.IsNewUser)
	assert.Equal(t, &role, result.User.Role)
}

func TestVerify_WrongCode(t *testing.T) {
	repo, smsMock := newMockRepo(), &mockSMS{}
	svc := buildService(repo, smsMock, allowLimiter{})
	phone := "+998901234567"

	_ = svc.SendCode(context.Background(), &auth.SendCodeRequest{Phone: phone})
	_, err := svc.Verify(context.Background(), &auth.VerifyRequest{Phone: phone, Code: "000000"})

	var domErr *domain.Error
	require.ErrorAs(t, err, &domErr)
	assert.Equal(t, "INVALID_CODE", domErr.Code)
}

func TestVerify_ExpiredCode(t *testing.T) {
	repo, smsMock := newMockRepo(), &mockSMS{}
	svc := buildService(repo, smsMock, allowLimiter{})
	phone := "+998901234567"

	_ = svc.SendCode(context.Background(), &auth.SendCodeRequest{Phone: phone})
	repo.codes[phone].ExpiresAt = time.Now().Add(-1 * time.Minute) // expire
	code := repo.codes[phone].Code

	_, err := svc.Verify(context.Background(), &auth.VerifyRequest{Phone: phone, Code: code})

	var domErr *domain.Error
	require.ErrorAs(t, err, &domErr)
	assert.Equal(t, "CODE_EXPIRED", domErr.Code)
}

func TestSetRole_Success(t *testing.T) {
	repo, smsMock := newMockRepo(), &mockSMS{}
	svc := buildService(repo, smsMock, allowLimiter{})
	phone := "+998901234567"

	_ = svc.SendCode(context.Background(), &auth.SendCodeRequest{Phone: phone})
	code := repo.codes[phone].Code
	verifyResult, _ := svc.Verify(context.Background(), &auth.VerifyRequest{Phone: phone, Code: code})

	userResp, err := svc.SetRole(context.Background(), verifyResult.User.ID, &auth.SetRoleRequest{Role: domain.RoleTeacher})

	require.NoError(t, err)
	require.NotNil(t, userResp.Role)
	assert.Equal(t, domain.RoleTeacher, *userResp.Role)
}

func TestSetRole_AlreadySet(t *testing.T) {
	repo, smsMock := newMockRepo(), &mockSMS{}
	svc := buildService(repo, smsMock, allowLimiter{})
	phone := "+998901234567"
	role := domain.RoleStudent
	repo.users[phone] = &domain.User{
		ID: uuid.New(), Phone: phone, Role: &role,
		Language: domain.LanguageUz, IsActive: true,
	}

	_, err := svc.SetRole(context.Background(), repo.users[phone].ID, &auth.SetRoleRequest{Role: domain.RoleTeacher})

	var domErr *domain.Error
	require.ErrorAs(t, err, &domErr)
	assert.Equal(t, "ROLE_ALREADY_SET", domErr.Code)
}

func TestRefresh_TokenRotation(t *testing.T) {
	repo, smsMock := newMockRepo(), &mockSMS{}
	svc := buildService(repo, smsMock, allowLimiter{})
	phone := "+998901234567"

	_ = svc.SendCode(context.Background(), &auth.SendCodeRequest{Phone: phone})
	code := repo.codes[phone].Code
	first, _ := svc.Verify(context.Background(), &auth.VerifyRequest{Phone: phone, Code: code})

	second, err := svc.Refresh(context.Background(), &auth.RefreshRequest{RefreshToken: first.RefreshToken})

	require.NoError(t, err)
	assert.NotEmpty(t, second.AccessToken)
	assert.NotEqual(t, first.RefreshToken, second.RefreshToken, "refresh token must rotate")
}

func TestRefresh_RevokedToken(t *testing.T) {
	repo, smsMock := newMockRepo(), &mockSMS{}
	svc := buildService(repo, smsMock, allowLimiter{})
	phone := "+998901234567"

	_ = svc.SendCode(context.Background(), &auth.SendCodeRequest{Phone: phone})
	code := repo.codes[phone].Code
	first, _ := svc.Verify(context.Background(), &auth.VerifyRequest{Phone: phone, Code: code})

	_, _ = svc.Refresh(context.Background(), &auth.RefreshRequest{RefreshToken: first.RefreshToken})
	// Второй раз с тем же токеном — должен отклонить
	_, err := svc.Refresh(context.Background(), &auth.RefreshRequest{RefreshToken: first.RefreshToken})

	require.Error(t, err)
}
