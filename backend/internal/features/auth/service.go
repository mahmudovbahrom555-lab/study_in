package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"net/http"
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
	repo     Repository
	sms      sms.Sender
	jwt      *jwt.Manager
	limiter  Limiter
	devCode  string // если задан — вместо случайного кода (только dev, см. config.validate)
	consents ConsentRecorder
	google   GoogleVerifier // nil — вход через Google выключен
	avatars  AvatarStorage  // nil — загрузка фото выключена
	telegram TelegramSender // nil — вход через Telegram-бота выключен
	botURL   string         // https://t.me/<бот> — куда отправить, если номер не подтверждён
}

// TelegramSender присылает код входа через бота (модуль tgbot). false —
// номер ещё не подтверждён в боте; код придёт, когда человек им поделится.
type TelegramSender interface {
	SendCode(ctx context.Context, phone, code string) (bool, error)
}

// WithTelegram включает коды входа через Telegram-бота (бесплатно, вместо SMS).
func (s *Service) WithTelegram(t TelegramSender, botURL string) *Service {
	s.telegram, s.botURL = t, botURL
	return s
}

// WithAvatarStorage включает фото профиля.
func (s *Service) WithAvatarStorage(a AvatarStorage) *Service {
	s.avatars = a
	return s
}

// WithGoogleVerifier включает вход через Google.
func (s *Service) WithGoogleVerifier(v GoogleVerifier) *Service {
	s.google = v
	return s
}

// ConsentRecorder пишет журнал согласий (consent_events).
type ConsentRecorder interface {
	RecordConsent(ctx context.Context, e *domain.ConsentEvent) error
}

// ConsentPrivacyPolicy — согласие с политикой конфиденциальности при регистрации.
const ConsentPrivacyPolicy = "privacy_policy"

// WithDevCode включает фиксированный код подтверждения для разработки.
func (s *Service) WithDevCode(code string) *Service {
	s.devCode = code
	return s
}

// WithConsentRecorder включает запись согласия с политикой при регистрации.
func (s *Service) WithConsentRecorder(c ConsentRecorder) *Service {
	s.consents = c
	return s
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
func (s *Service) SendCode(ctx context.Context, req *SendCodeRequest) (*SendCodeResponse, error) {
	viaTelegram := req.Channel == ChannelTelegram
	if viaTelegram && s.telegram == nil {
		return nil, domain.NewError("TELEGRAM_UNAVAILABLE", "вход через Telegram не настроен", domain.ErrValidation)
	}
	key := fmt.Sprintf("sms:%s", req.Phone)
	if err := s.limiter.Allow(ctx, key, smsRateLimit, smsRateWindow); err != nil {
		return nil, err
	}

	code := s.devCode
	if code == "" {
		var err error
		if code, err = generateCode(codeLength); err != nil {
			return nil, fmt.Errorf("generate code: %w", err)
		}
	}

	expiresAt := time.Now().Add(codeTTL)
	if err := s.repo.CreateVerificationCode(ctx, req.Phone, code, expiresAt); err != nil {
		return nil, fmt.Errorf("save code: %w", err)
	}

	if viaTelegram {
		delivered, err := s.telegram.SendCode(ctx, req.Phone, code)
		if err != nil {
			return nil, fmt.Errorf("send telegram code: %w", err)
		}
		// Не доставлен — человек откроет бота, поделится номером, и бот
		// пришлёт этот же код (он уже сохранён).
		return &SendCodeResponse{Channel: ChannelTelegram, Delivered: delivered, BotURL: s.botURL}, nil
	}

	message := verificationMessage(code)
	if err := s.sms.Send(ctx, req.Phone, message); err != nil {
		return nil, fmt.Errorf("send sms: %w", err)
	}
	return &SendCodeResponse{Channel: ChannelSMS, Delivered: true}, nil
}

// Verify проверяет код и выдаёт токены. Создаёт пользователя при первом входе.
func (s *Service) Verify(ctx context.Context, req *VerifyRequest) (*Response, error) {
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
		phone := req.Phone
		user = &domain.User{
			ID:        uuid.New(),
			Phone:     &phone,
			Name:      "",
			Language:  domain.LanguageUz,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
		s.recordPrivacyConsent(ctx, user.ID, req.PrivacyVersion, req.IP, req.UserAgent)
	}

	return s.signIn(ctx, user, req.DeviceInfo, isNewUser)
}

// GoogleSignIn проверяет Google ID-токен и выдаёт токены. Создаёт пользователя
// при первом входе; аккаунт связывается по sub, а не по email.
func (s *Service) GoogleSignIn(ctx context.Context, req *GoogleSignInRequest) (*Response, error) {
	if s.google == nil {
		return nil, domain.NewError("GOOGLE_DISABLED", "Вход через Google не настроен", domain.ErrValidation)
	}
	id, err := s.google.Verify(ctx, req.IDToken)
	if err != nil {
		return nil, domain.NewError("GOOGLE_TOKEN_INVALID", "Недействительный токен Google", domain.ErrUnauthorized)
	}

	user, err := s.repo.GetUserByGoogleSub(ctx, id.Sub)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	isNewUser := user == nil
	if isNewUser {
		sub := id.Sub
		user = &domain.User{
			ID:        uuid.New(),
			GoogleSub: &sub,
			Name:      truncateRunes(id.Name, 100),
			Language:  domain.LanguageUz,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if id.Email != "" {
			email := id.Email
			user.Email = &email
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
		s.recordPrivacyConsent(ctx, user.ID, req.PrivacyVersion, req.IP, req.UserAgent)
	}
	if !user.IsActive {
		return nil, domain.ErrUnauthorized
	}

	return s.signIn(ctx, user, req.DeviceInfo, isNewUser)
}

func (s *Service) signIn(ctx context.Context, user *domain.User, deviceInfo string, isNewUser bool) (*Response, error) {
	accessToken, refreshToken, err := s.issueTokenPair(ctx, user, deviceInfo)
	if err != nil {
		return nil, err
	}
	return &Response{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         s.userResponse(ctx, user),
		IsNewUser:    isNewUser,
	}, nil
}

// recordPrivacyConsent сохраняет, с какой версией политики человек согласился
// при регистрации. Сбой записи не мешает входу — пишем в лог.
func (s *Service) recordPrivacyConsent(ctx context.Context, userID uuid.UUID, version, ip, userAgent string) {
	if s.consents == nil || version == "" {
		return
	}
	e := &domain.ConsentEvent{
		ID:             uuid.New(),
		UserID:         userID,
		ConsentType:    ConsentPrivacyPolicy,
		ConsentVersion: version,
		Granted:        true,
		CreatedAt:      time.Now(),
	}
	if parsed := net.ParseIP(ip); parsed != nil {
		v := parsed.String()
		e.IPAddress = &v
	}
	if userAgent != "" {
		ua := userAgent
		e.UserAgent = &ua
	}
	if err := s.consents.RecordConsent(ctx, e); err != nil {
		slog.Error("record privacy consent", slog.String("err", err.Error()))
	}
}

// Refresh обменивает refresh-токен на новую пару токенов.
func (s *Service) Refresh(ctx context.Context, req *RefreshRequest) (*Response, error) {
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

	return &Response{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         s.userResponse(ctx, user),
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

	resp := s.userResponse(ctx, user)
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
	resp := s.userResponse(ctx, user)
	return &resp, nil
}

// UpdateMe обновляет имя и/или язык профиля.
func (s *Service) UpdateMe(ctx context.Context, userID uuid.UUID, req *UpdateMeRequest) (*UserResponse, error) {
	if req.Language != nil && !req.Language.Valid() {
		return nil, domain.NewError("UNSUPPORTED_LANGUAGE", "unsupported language", domain.ErrValidation)
	}
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
	if req.Username != nil {
		if err := setUsername(user, *req.Username); err != nil {
			return nil, err
		}
	}
	// Только в одну сторону: иначе ребёнок сам убрал бы согласие родителя.
	if req.IsChild != nil && *req.IsChild {
		if user.Role == nil || *user.Role != domain.RoleStudent {
			return nil, domain.NewError("CHILD_STUDENTS_ONLY", "only a student account can be marked as a child", domain.ErrForbidden)
		}
		user.IsChild = true
	}
	user.UpdatedAt = time.Now()

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	resp := s.userResponse(ctx, user)
	return &resp, nil
}

// setUsername: username бывает только у репетитора — учеников не находят
// по имени (дети), родителей искать незачем.
func setUsername(user *domain.User, raw string) error {
	u := domain.NormalizeUsername(raw)
	if u == "" {
		user.Username = nil
		return nil
	}
	if user.Role == nil || *user.Role != domain.RoleTeacher {
		return domain.NewError("USERNAME_TEACHERS_ONLY", "only tutors can have a username", domain.ErrForbidden)
	}
	if !domain.ValidUsername(u) {
		return domain.NewError("INVALID_USERNAME", "4-32 latin letters, digits or _, starting with a letter", domain.ErrValidation)
	}
	user.Username = &u
	return nil
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

// verificationMessage — текст SMS на двух языках: сначала английский, затем русский.
// Язык пользователя на этапе входа ещё неизвестен. Кириллица переводит SMS в UCS-2,
// где лимит одного сообщения — 70 символов; текст подобран ровно под него,
// чтобы не платить за второй сегмент (проверяется в тесте).
func verificationMessage(code string) string {
	return fmt.Sprintf("%s — MaMu Learn code. Don't share it.\nНикому не сообщайте код.", code)
}

// truncateRunes обрезает имя под VARCHAR(100), не разрывая символ UTF-8.
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
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

// MaxAvatarSize — лимит фото профиля.
const MaxAvatarSize = 2 << 20

var avatarTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// UploadAvatar сохраняет фото профиля. Только репетитор и родитель: ученики
// (среди них дети) выбирают животное — фото детей мы не храним.
func (s *Service) UploadAvatar(ctx context.Context, userID uuid.UUID, data []byte) (*UserResponse, error) {
	if s.avatars == nil {
		return nil, domain.NewError("STORAGE_DISABLED", "Загрузка фото недоступна", domain.ErrValidation)
	}
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrNotFound
	}
	if user.Role == nil || *user.Role == domain.RoleStudent {
		return nil, domain.NewError("AVATAR_PRESET_ONLY", "Ученики выбирают аватар из набора", domain.ErrForbidden)
	}
	if len(data) > MaxAvatarSize {
		return nil, domain.NewError("AVATAR_TOO_LARGE", "Фото больше 2 МБ", domain.ErrValidation)
	}
	contentType := http.DetectContentType(data)
	ext, ok := avatarTypes[contentType]
	if !ok {
		return nil, domain.NewError("AVATAR_BAD_TYPE", "Нужна картинка JPEG, PNG или WebP", domain.ErrValidation)
	}

	key := fmt.Sprintf("avatars/%s/%s%s", userID, uuid.New(), ext)
	if err := s.avatars.PutObject(ctx, key, bytes.NewReader(data), int64(len(data)), contentType); err != nil {
		return nil, fmt.Errorf("put avatar: %w", err)
	}
	old := user.AvatarURL
	user.AvatarURL = &key
	user.AvatarPreset = nil
	if err := s.saveAvatar(ctx, user); err != nil {
		s.removeAvatarObject(ctx, &key)
		return nil, err
	}
	s.removeAvatarObject(ctx, old)
	resp := s.userResponse(ctx, user)
	return &resp, nil
}

// SetAvatarPreset выбирает аватар-животное (любая роль); фото удаляется.
func (s *Service) SetAvatarPreset(ctx context.Context, userID uuid.UUID, preset string) (*UserResponse, error) {
	if !domain.ValidAvatarPreset(preset) {
		return nil, domain.NewError("AVATAR_PRESET_UNKNOWN", "Нет такого аватара", domain.ErrValidation)
	}
	return s.replaceAvatar(ctx, userID, &preset)
}

// DeleteAvatar убирает и фото, и животное.
func (s *Service) DeleteAvatar(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	return s.replaceAvatar(ctx, userID, nil)
}

func (s *Service) replaceAvatar(ctx context.Context, userID uuid.UUID, preset *string) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrNotFound
	}
	old := user.AvatarURL
	user.AvatarURL = nil
	user.AvatarPreset = preset
	if err := s.saveAvatar(ctx, user); err != nil {
		return nil, err
	}
	s.removeAvatarObject(ctx, old)
	resp := s.userResponse(ctx, user)
	return &resp, nil
}

func (s *Service) saveAvatar(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// removeAvatarObject удаляет прежнее фото; сбой не мешает — пишем в лог.
func (s *Service) removeAvatarObject(ctx context.Context, key *string) {
	if key == nil || s.avatars == nil {
		return
	}
	if err := s.avatars.RemoveObject(ctx, *key); err != nil {
		slog.Warn("remove avatar", slog.String("err", err.Error()))
	}
}

// userResponse — ответ с подписанной ссылкой на фото (ключ наружу не отдаём).
func (s *Service) userResponse(ctx context.Context, u *domain.User) UserResponse {
	resp := userToResponse(u)
	if u.AvatarURL != nil && s.avatars != nil {
		if url, err := s.avatars.PresignedGetURL(ctx, *u.AvatarURL, ""); err == nil {
			resp.AvatarURL = &url
		}
	}
	return resp
}
