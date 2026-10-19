package chat

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

const (
	maxBody         = 2000
	defaultPageSize = 50
	maxPageSize     = 100
	sendLimit       = 20 // сообщений в минуту от человека в классе
	sendWindow      = time.Minute
)

var (
	errMuted     = domain.NewError("CHAT_MUTED", "you can't write in this chat", domain.ErrForbidden)
	errReadOnly  = domain.NewError("CHAT_READ_ONLY", "only the tutor writes in this chat", domain.ErrForbidden)
	errBanned    = domain.NewError("CHAT_BANNED", "writing in chats is disabled for this account", domain.ErrForbidden)
	errEmpty     = domain.NewError("MESSAGE_EMPTY", "message is empty", domain.ErrValidation)
	errTooLong   = domain.NewError("MESSAGE_TOO_LONG", "message is longer than 2000 characters", domain.ErrValidation)
	errBadTTL    = domain.NewError("CHAT_TTL_INVALID", "unsupported message lifetime", domain.ErrValidation)
	errBadMode   = domain.NewError("CHAT_MODE_INVALID", "unsupported chat mode", domain.ErrValidation)
	errNotMember = domain.NewError("NOT_GROUP_MEMBER", "user is not in this class", domain.ErrValidation)
)

type Service struct {
	repo    Repository
	limiter Limiter
	pub     Publisher
	now     func() time.Time
}

func NewService(repo Repository, limiter Limiter) *Service {
	return &Service{repo: repo, limiter: limiter, now: time.Now}
}

// WithPublisher включает рассылку событий подключённым клиентам.
func (s *Service) WithPublisher(p Publisher) *Service {
	s.pub = p
	return s
}

func (s *Service) publish(ctx context.Context, groupID uuid.UUID, typ string, data any) {
	if s.pub != nil {
		s.pub.Publish(ctx, groupID, Event{Type: typ, GroupID: groupID.String(), Data: data})
	}
}

// access проверяет, что класс существует и пользователь в нём (владелец или участник).
func (s *Service) access(ctx context.Context, groupID, userID uuid.UUID) (*domain.ChatAccess, error) {
	a, err := s.repo.Access(ctx, groupID, userID)
	if err != nil {
		return nil, fmt.Errorf("chat.access: %w", err)
	}
	if a == nil {
		return nil, domain.ErrNotFound
	}
	if a.TeacherID != userID && !a.IsMember {
		return nil, domain.ErrForbidden
	}
	return a, nil
}

// Send отправляет сообщение. Срок жизни — по настройке класса на момент отправки.
func (s *Service) Send(ctx context.Context, groupID, userID, clientID uuid.UUID, body string) (*MessageResponse, error) {
	a, err := s.access(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	isOwner := a.TeacherID == userID
	switch {
	case !a.IsActive, a.ChatBanned:
		return nil, errBanned
	case a.IsArchived:
		return nil, domain.ErrForbidden
	case !isOwner && a.Mode == domain.ChatModeOwnerOnly:
		return nil, errReadOnly
	case !isOwner && a.Muted:
		return nil, errMuted
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return nil, errEmpty
	}
	if utf8.RuneCountInString(body) > maxBody {
		return nil, errTooLong
	}
	if err := s.limiter.Allow(ctx, fmt.Sprintf("chat:%s:%s", groupID, userID), sendLimit, sendWindow); err != nil {
		return nil, err
	}

	now := s.now()
	m, err := s.repo.Insert(ctx, &domain.ChatMessage{
		ID:        uuid.New(),
		GroupID:   groupID,
		AuthorID:  userID,
		Body:      body,
		ClientID:  clientID,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(a.TTLSeconds) * time.Second),
	})
	if err != nil {
		return nil, fmt.Errorf("chat.Send: %w", err)
	}
	resp := toResponse(m, a.TeacherID)
	s.publish(ctx, groupID, "message.created", resp)
	return &resp, nil
}

// List — живые сообщения класса страницами назад, новые первыми.
func (s *Service) List(ctx context.Context, groupID, userID uuid.UUID, before *uuid.UUID, limit int) ([]MessageResponse, error) {
	a, err := s.access(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}
	ms, err := s.repo.List(ctx, groupID, before, limit, s.now())
	if err != nil {
		return nil, fmt.Errorf("chat.List: %w", err)
	}
	out := make([]MessageResponse, len(ms))
	for i, m := range ms {
		out[i] = toResponse(m, a.TeacherID)
	}
	return out, nil
}

// Delete удаляет сообщение: своё — автор, любое — владелец класса.
func (s *Service) Delete(ctx context.Context, groupID, messageID, userID uuid.UUID) error {
	m, err := s.repo.Get(ctx, messageID)
	if err != nil {
		return fmt.Errorf("chat.Delete fetch: %w", err)
	}
	if m == nil || m.GroupID != groupID {
		return domain.ErrNotFound
	}
	a, err := s.access(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if m.AuthorID != userID && a.TeacherID != userID {
		return domain.ErrForbidden
	}
	if err := s.repo.Delete(ctx, messageID); err != nil {
		return fmt.Errorf("chat.Delete: %w", err)
	}
	s.publish(ctx, groupID, "message.deleted", map[string]string{"id": messageID.String()})
	return nil
}

// MarkRead отмечает чат прочитанным до момента at (не позже текущего).
func (s *Service) MarkRead(ctx context.Context, groupID, userID uuid.UUID, at time.Time) error {
	if _, err := s.access(ctx, groupID, userID); err != nil {
		return err
	}
	if now := s.now(); at.IsZero() || at.After(now) {
		at = now
	}
	return s.repo.MarkRead(ctx, groupID, userID, at)
}

// UnreadCounts — для списка классов; без проверки прав: классы уже свои.
func (s *Service) UnreadCounts(ctx context.Context, userID uuid.UUID, groupIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(groupIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}
	return s.repo.UnreadCounts(ctx, userID, groupIDs, s.now())
}

// ─── Модерация (только владелец класса) ──────────────────────────────────────

func (s *Service) owner(ctx context.Context, groupID, userID uuid.UUID) (*domain.ChatAccess, error) {
	a, err := s.access(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if a.TeacherID != userID {
		return nil, domain.ErrForbidden
	}
	return a, nil
}

// UpdateSettings меняет срок хранения (только для новых сообщений) и режим.
func (s *Service) UpdateSettings(ctx context.Context, groupID, userID uuid.UUID, ttlSeconds int, mode domain.ChatMode) error {
	if _, err := s.owner(ctx, groupID, userID); err != nil {
		return err
	}
	if !domain.ValidChatTTL(ttlSeconds) {
		return errBadTTL
	}
	if mode != domain.ChatModeOpen && mode != domain.ChatModeOwnerOnly {
		return errBadMode
	}
	if err := s.repo.UpdateSettings(ctx, groupID, ttlSeconds, mode); err != nil {
		return fmt.Errorf("chat.UpdateSettings: %w", err)
	}
	s.publish(ctx, groupID, "chat.settings_changed", SettingsResponse{TTLSeconds: ttlSeconds, Mode: string(mode)})
	return nil
}

// SetMuted запрещает или разрешает участнику писать в чат класса.
func (s *Service) SetMuted(ctx context.Context, groupID, ownerID, memberID uuid.UUID, muted bool) error {
	if err := s.ownerAndMember(ctx, groupID, ownerID, memberID); err != nil {
		return err
	}
	if err := s.repo.SetMuted(ctx, groupID, memberID, muted); err != nil {
		return fmt.Errorf("chat.SetMuted: %w", err)
	}
	typ := "member.unmuted"
	if muted {
		typ = "member.muted"
	}
	s.publish(ctx, groupID, typ, map[string]string{"user_id": memberID.String()})
	return nil
}

// Ban удаляет участника из класса и не даёт вступить снова.
func (s *Service) Ban(ctx context.Context, groupID, ownerID, memberID uuid.UUID) error {
	if err := s.ownerAndMember(ctx, groupID, ownerID, memberID); err != nil {
		return err
	}
	if err := s.repo.Ban(ctx, groupID, memberID); err != nil {
		return fmt.Errorf("chat.Ban: %w", err)
	}
	s.publish(ctx, groupID, "member.removed", map[string]string{"user_id": memberID.String()})
	return nil
}

func (s *Service) ownerAndMember(ctx context.Context, groupID, ownerID, memberID uuid.UUID) error {
	if _, err := s.owner(ctx, groupID, ownerID); err != nil {
		return err
	}
	m, err := s.repo.Access(ctx, groupID, memberID)
	if err != nil {
		return fmt.Errorf("chat.member: %w", err)
	}
	if m == nil || !m.IsMember {
		return errNotMember
	}
	return nil
}

// DeleteExpired — фоновая очистка истёкших сообщений.
func (s *Service) DeleteExpired(ctx context.Context) (int, error) {
	total := 0
	for {
		n, err := s.repo.DeleteExpired(ctx, s.now(), 1000)
		total += n
		if err != nil || n < 1000 {
			return total, err
		}
	}
}
