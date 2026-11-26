package auth_test

import (
	"context"
	"io"
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
	users      map[string]*domain.User
	codes      map[string]*domain.VerificationCode
	tokens     map[string]*domain.RefreshToken
	attempts   map[uuid.UUID]int
	markedUsed map[uuid.UUID]bool
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

// userKey — ключ в mockRepo.users: телефон или google:<sub>.
func userKey(u *domain.User) string {
	if u.Phone != nil {
		return *u.Phone
	}
	return "google:" + *u.GoogleSub
}

func (m *mockRepo) CreateUser(_ context.Context, u *domain.User) error {
	m.users[userKey(u)] = u
	return nil
}
func (m *mockRepo) GetUserByGoogleSub(_ context.Context, sub string) (*domain.User, error) {
	return m.users["google:"+sub], nil
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
	m.users[userKey(u)] = u
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

type memConsents struct{ events []*domain.ConsentEvent }

func (m *memConsents) RecordConsent(_ context.Context, e *domain.ConsentEvent) error {
	m.events = append(m.events, e)
	return nil
}

func TestVerify_RecordsPrivacyConsentForNewUserOnly(t *testing.T) {
	repo, smsMock := newMockRepo(), &mockSMS{}
	consents := &memConsents{}
	svc := buildService(repo, smsMock, allowLimiter{}).WithConsentRecorder(consents)
	phone := "+998901234567"
	login := func() {
		_ = svc.SendCode(context.Background(), &auth.SendCodeRequest{Phone: phone})
		_, err := svc.Verify(context.Background(), &auth.VerifyRequest{
			Phone: phone, Code: repo.codes[phone].Code,
			PrivacyVersion: "2026-10-17", IP: "203.0.113.7", UserAgent: "MaMuLearn/0.1",
		})
		require.NoError(t, err)
	}

	login()
	require.Len(t, consents.events, 1)
	e := consents.events[0]
	assert.Equal(t, auth.ConsentPrivacyPolicy, e.ConsentType)
	assert.Equal(t, "2026-10-17", e.ConsentVersion)
	assert.True(t, e.Granted)
	require.NotNil(t, e.IPAddress)
	assert.Equal(t, "203.0.113.7", *e.IPAddress)

	login() // повторный вход — аккаунт уже есть, согласие уже записано
	assert.Len(t, consents.events, 1)
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
		ID: uuid.New(), Phone: &phone, Role: &role,
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
		ID: uuid.New(), Phone: &phone, Role: &role,
		Language: domain.LanguageUz, IsActive: true,
	}

	_, err := svc.SetRole(context.Background(), repo.users[phone].ID, &auth.SetRoleRequest{Role: domain.RoleTeacher})

	var domErr *domain.Error
	require.ErrorAs(t, err, &domErr)
	assert.Equal(t, "ROLE_ALREADY_SET", domErr.Code)
}

func TestUpdateMe_Language(t *testing.T) {
	repo, smsMock := newMockRepo(), &mockSMS{}
	svc := buildService(repo, smsMock, allowLimiter{})
	phone := "+998901234567"
	repo.users[phone] = &domain.User{
		ID: uuid.New(), Phone: &phone, Language: domain.LanguageUz, IsActive: true,
	}
	id := repo.users[phone].ID

	ru := domain.LanguageRu
	resp, err := svc.UpdateMe(context.Background(), id, &auth.UpdateMeRequest{Language: &ru})
	require.NoError(t, err)
	assert.Equal(t, "ru", resp.Language)

	xx := domain.Language("xx")
	_, err = svc.UpdateMe(context.Background(), id, &auth.UpdateMeRequest{Language: &xx})
	var domErr *domain.Error
	require.ErrorAs(t, err, &domErr)
	assert.Equal(t, "UNSUPPORTED_LANGUAGE", domErr.Code)
}

func TestUpdateMe_Username(t *testing.T) {
	repo, smsMock := newMockRepo(), &mockSMS{}
	svc := buildService(repo, smsMock, allowLimiter{})
	teacher, student := domain.RoleTeacher, domain.RoleStudent
	tPhone, sPhone := "+998901111111", "+998902222222"
	repo.users[tPhone] = &domain.User{ID: uuid.New(), Phone: &tPhone, Role: &teacher, IsActive: true}
	repo.users[sPhone] = &domain.User{ID: uuid.New(), Phone: &sPhone, Role: &student, IsActive: true}
	ctx := context.Background()
	code := func(err error) string {
		var domErr *domain.Error
		require.ErrorAs(t, err, &domErr)
		return domErr.Code
	}
	str := func(s string) *string { return &s }

	resp, err := svc.UpdateMe(ctx, repo.users[tPhone].ID, &auth.UpdateMeRequest{Username: str(" @Anna_Math ")})
	require.NoError(t, err)
	require.NotNil(t, resp.Username)
	assert.Equal(t, "anna_math", *resp.Username)

	_, err = svc.UpdateMe(ctx, repo.users[tPhone].ID, &auth.UpdateMeRequest{Username: str("1abc")})
	assert.Equal(t, "INVALID_USERNAME", code(err))
	_, err = svc.UpdateMe(ctx, repo.users[tPhone].ID, &auth.UpdateMeRequest{Username: str("анна")})
	assert.Equal(t, "INVALID_USERNAME", code(err))

	// Ученика (ребёнка) нельзя сделать находимым по имени.
	_, err = svc.UpdateMe(ctx, repo.users[sPhone].ID, &auth.UpdateMeRequest{Username: str("kid_one")})
	assert.Equal(t, "USERNAME_TEACHERS_ONLY", code(err))

	resp, err = svc.UpdateMe(ctx, repo.users[tPhone].ID, &auth.UpdateMeRequest{Username: str("")})
	require.NoError(t, err)
	assert.Nil(t, resp.Username)
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

// ─── Google sign-in ──────────────────────────────────────────────────────────

type fakeGoogle struct {
	id  *auth.GoogleIdentity
	err error
}

func (f fakeGoogle) Verify(_ context.Context, _ string) (*auth.GoogleIdentity, error) {
	return f.id, f.err
}

func TestGoogleSignIn_NewThenExistingUser(t *testing.T) {
	repo := newMockRepo()
	consents := &memConsents{}
	svc := buildService(repo, &mockSMS{}, allowLimiter{}).
		WithConsentRecorder(consents).
		WithGoogleVerifier(fakeGoogle{id: &auth.GoogleIdentity{Sub: "1234", Email: "a@gmail.com", Name: "Ali"}})
	req := &auth.GoogleSignInRequest{IDToken: "tok", PrivacyVersion: "2026-10-17"}

	first, err := svc.GoogleSignIn(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, first.IsNewUser)
	assert.NotEmpty(t, first.AccessToken)
	assert.Equal(t, "a@gmail.com", first.User.Email)
	assert.Equal(t, "Ali", first.User.Name)
	assert.Empty(t, first.User.Phone)
	assert.Len(t, consents.events, 1)

	second, err := svc.GoogleSignIn(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, second.IsNewUser)
	assert.Equal(t, first.User.ID, second.User.ID)
	assert.Len(t, consents.events, 1)
}

func TestGoogleSignIn_InvalidToken(t *testing.T) {
	svc := buildService(newMockRepo(), &mockSMS{}, allowLimiter{}).
		WithGoogleVerifier(fakeGoogle{err: assert.AnError})

	_, err := svc.GoogleSignIn(context.Background(), &auth.GoogleSignInRequest{IDToken: "bad"})

	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestGoogleSignIn_DisabledWithoutVerifier(t *testing.T) {
	svc := buildService(newMockRepo(), &mockSMS{}, allowLimiter{})

	_, err := svc.GoogleSignIn(context.Background(), &auth.GoogleSignInRequest{IDToken: "tok"})

	assert.Error(t, err)
}

// ─── Avatars ─────────────────────────────────────────────────────────────────

type memAvatars struct {
	objects map[string][]byte
	removed []string
}

func (m *memAvatars) PutObject(_ context.Context, key string, r io.Reader, _ int64, _ string) error {
	b, _ := io.ReadAll(r)
	m.objects[key] = b
	return nil
}

func (m *memAvatars) RemoveObject(_ context.Context, key string) error {
	delete(m.objects, key)
	m.removed = append(m.removed, key)
	return nil
}

func (m *memAvatars) PresignedGetURL(_ context.Context, key, _ string) (string, error) {
	return "https://signed/" + key, nil
}

var pngBytes = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")

func avatarSetup(role domain.Role) (*auth.Service, *mockRepo, *memAvatars, uuid.UUID) {
	repo := newMockRepo()
	phone := "+998901111111"
	id := uuid.New()
	repo.users[phone] = &domain.User{ID: id, Phone: &phone, Role: &role, IsActive: true}
	store := &memAvatars{objects: map[string][]byte{}}
	svc := buildService(repo, &mockSMS{}, allowLimiter{}).WithAvatarStorage(store)
	return svc, repo, store, id
}

func TestUploadAvatar_StudentGetsPresetsOnly(t *testing.T) {
	svc, _, store, id := avatarSetup(domain.RoleStudent)

	_, err := svc.UploadAvatar(context.Background(), id, pngBytes)

	assert.ErrorIs(t, err, domain.ErrForbidden)
	assert.Empty(t, store.objects)
}

func TestUploadAvatar_TeacherPhotoReplacesOldAndPreset(t *testing.T) {
	svc, _, store, id := avatarSetup(domain.RoleTeacher)
	ctx := context.Background()

	_, err := svc.SetAvatarPreset(ctx, id, "fox")
	require.NoError(t, err)
	first, err := svc.UploadAvatar(ctx, id, pngBytes)
	require.NoError(t, err)
	require.NotNil(t, first.AvatarURL)
	assert.Contains(t, *first.AvatarURL, "https://signed/avatars/"+id.String()+"/")
	assert.Nil(t, first.AvatarPreset)

	second, err := svc.UploadAvatar(ctx, id, pngBytes)
	require.NoError(t, err)
	assert.NotEqual(t, *first.AvatarURL, *second.AvatarURL)
	assert.Len(t, store.objects, 1, "old photo removed")
}

func TestUploadAvatar_RejectsNonImageAndTooLarge(t *testing.T) {
	svc, _, _, id := avatarSetup(domain.RoleParent)
	ctx := context.Background()

	_, err := svc.UploadAvatar(ctx, id, []byte("not an image at all"))
	assert.ErrorIs(t, err, domain.ErrValidation)

	big := append(append([]byte{}, pngBytes...), make([]byte, auth.MaxAvatarSize)...)
	_, err = svc.UploadAvatar(ctx, id, big)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestAvatarPreset_ValidatesAndDeleteClears(t *testing.T) {
	svc, _, store, id := avatarSetup(domain.RoleTeacher)
	ctx := context.Background()

	_, err := svc.SetAvatarPreset(ctx, id, "dragon")
	assert.ErrorIs(t, err, domain.ErrValidation)

	_, err = svc.UploadAvatar(ctx, id, pngBytes)
	require.NoError(t, err)
	withPreset, err := svc.SetAvatarPreset(ctx, id, "panda")
	require.NoError(t, err)
	require.NotNil(t, withPreset.AvatarPreset)
	assert.Equal(t, "panda", *withPreset.AvatarPreset)
	assert.Nil(t, withPreset.AvatarURL)
	assert.Empty(t, store.objects, "photo removed when an animal is chosen")

	cleared, err := svc.DeleteAvatar(ctx, id)
	require.NoError(t, err)
	assert.Nil(t, cleared.AvatarPreset)
	assert.Nil(t, cleared.AvatarURL)
}
