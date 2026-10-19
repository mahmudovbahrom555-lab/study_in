package chat

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// ─── fakes ────────────────────────────────────────────────────────────────────

type memRepo struct {
	access   map[uuid.UUID]*domain.ChatAccess // userID → права в единственном классе
	msgs     map[uuid.UUID]*domain.ChatMessage
	settings struct {
		ttl  int
		mode domain.ChatMode
	}
	banned map[uuid.UUID]bool
}

func (r *memRepo) Access(_ context.Context, _, userID uuid.UUID) (*domain.ChatAccess, error) {
	a, ok := r.access[userID]
	if !ok {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}

func (r *memRepo) Insert(_ context.Context, m *domain.ChatMessage) (*domain.ChatMessage, error) {
	for _, x := range r.msgs {
		if x.AuthorID == m.AuthorID && x.ClientID == m.ClientID {
			return x, nil
		}
	}
	cp := *m
	r.msgs[m.ID] = &cp
	return &cp, nil
}

func (r *memRepo) List(_ context.Context, groupID uuid.UUID, _ *uuid.UUID, limit int, now time.Time) ([]*domain.ChatMessage, error) {
	var out []*domain.ChatMessage
	for _, m := range r.msgs {
		if m.GroupID == groupID && m.ExpiresAt.After(now) {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *memRepo) Get(_ context.Context, id uuid.UUID) (*domain.ChatMessage, error) { return r.msgs[id], nil }
func (r *memRepo) Delete(_ context.Context, id uuid.UUID) error            { delete(r.msgs, id); return nil }
func (r *memRepo) DeleteExpired(_ context.Context, now time.Time, _ int) (int, error) {
	n := 0
	for id, m := range r.msgs {
		if !m.ExpiresAt.After(now) {
			delete(r.msgs, id)
			n++
		}
	}
	return n, nil
}
func (r *memRepo) MarkRead(context.Context, uuid.UUID, uuid.UUID, time.Time) error { return nil }
func (r *memRepo) UnreadCounts(context.Context, uuid.UUID, []uuid.UUID, time.Time) (map[uuid.UUID]int, error) {
	return nil, nil
}
func (r *memRepo) UpdateSettings(_ context.Context, _ uuid.UUID, ttl int, mode domain.ChatMode) error {
	r.settings.ttl, r.settings.mode = ttl, mode
	return nil
}
func (r *memRepo) SetMuted(_ context.Context, _, userID uuid.UUID, muted bool) error {
	r.access[userID].Muted = muted
	return nil
}
func (r *memRepo) Ban(_ context.Context, _, userID uuid.UUID) error {
	r.banned[userID] = true
	delete(r.access, userID)
	return nil
}

type allow struct{ deny bool }

func (a allow) Allow(context.Context, string, int, time.Duration) error {
	if a.deny {
		return domain.ErrRateLimit
	}
	return nil
}

type recPub struct{ events []Event }

func (p *recPub) Publish(_ context.Context, _ uuid.UUID, e Event) { p.events = append(p.events, e) }

// ─── fixture ──────────────────────────────────────────────────────────────────

var (
	groupID   = uuid.New()
	ownerID   = uuid.New()
	studentID = uuid.New()
	otherID   = uuid.New() // пользователь не из класса
	now       = time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
)

func setup() (*Service, *memRepo, *recPub) {
	base := domain.ChatAccess{TeacherID: ownerID, TTLSeconds: 86400, Mode: domain.ChatModeOpen, IsActive: true}
	owner, student, other := base, base, base
	student.IsMember = true
	repo := &memRepo{
		access: map[uuid.UUID]*domain.ChatAccess{ownerID: &owner, studentID: &student, otherID: &other},
		msgs:   map[uuid.UUID]*domain.ChatMessage{},
		banned: map[uuid.UUID]bool{},
	}
	pub := &recPub{}
	svc := NewService(repo, allow{}).WithPublisher(pub)
	svc.now = func() time.Time { return now }
	return svc, repo, pub
}

func code(t *testing.T, err error) string {
	t.Helper()
	var de *domain.Error
	if !errors.As(err, &de) {
		t.Fatalf("want domain.Error, got %v", err)
	}
	return de.Code
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestSend_MemberGetsTTLAndEvent(t *testing.T) {
	svc, _, pub := setup()
	m, err := svc.Send(context.Background(), groupID, studentID, uuid.New(), "  Hello  ")
	if err != nil {
		t.Fatal(err)
	}
	if m.Body != "Hello" || !m.ExpiresAt.Equal(now.Add(24*time.Hour)) || m.FromOwner {
		t.Fatalf("unexpected message: %+v", m)
	}
	if len(pub.events) != 1 || pub.events[0].Type != "message.created" {
		t.Fatalf("event not published: %+v", pub.events)
	}
}

func TestSend_OwnerMessageMarkedFromOwner(t *testing.T) {
	svc, _, _ := setup()
	m, err := svc.Send(context.Background(), groupID, ownerID, uuid.New(), "https://example.com/article")
	if err != nil || !m.FromOwner {
		t.Fatalf("owner message must be from_owner: %+v %v", m, err)
	}
}

func TestSend_NotMemberForbidden(t *testing.T) {
	svc, _, _ := setup()
	_, err := svc.Send(context.Background(), groupID, otherID, uuid.New(), "hi")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("want forbidden, got %v", err)
	}
}

func TestSend_Restrictions(t *testing.T) {
	cases := []struct {
		name  string
		setup func(r *memRepo)
		user  uuid.UUID
		want  string
	}{
		{"muted in class", func(r *memRepo) { r.access[studentID].Muted = true }, studentID, "CHAT_MUTED"},
		{"owner only mode", func(r *memRepo) { r.access[studentID].Mode = domain.ChatModeOwnerOnly }, studentID, "CHAT_READ_ONLY"},
		{"banned in all chats", func(r *memRepo) { r.access[studentID].ChatBanned = true }, studentID, "CHAT_BANNED"},
		{"blocked account", func(r *memRepo) { r.access[studentID].IsActive = false }, studentID, "CHAT_BANNED"},
		{"banned owner too", func(r *memRepo) { r.access[ownerID].ChatBanned = true }, ownerID, "CHAT_BANNED"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			svc, repo, _ := setup()
			c.setup(repo)
			_, err := svc.Send(context.Background(), groupID, c.user, uuid.New(), "hi")
			if got := code(t, err); got != c.want {
				t.Fatalf("want %s, got %s", c.want, got)
			}
		})
	}
}

func TestSend_OwnerWritesInOwnerOnlyMode(t *testing.T) {
	svc, repo, _ := setup()
	repo.access[ownerID].Mode = domain.ChatModeOwnerOnly
	if _, err := svc.Send(context.Background(), groupID, ownerID, uuid.New(), "announcement"); err != nil {
		t.Fatal(err)
	}
}

func TestSend_Validation(t *testing.T) {
	svc, _, _ := setup()
	if got := code(t, mustErr(svc.Send(context.Background(), groupID, studentID, uuid.New(), "   "))); got != "MESSAGE_EMPTY" {
		t.Fatal(got)
	}
	long := strings.Repeat("я", 2001)
	if got := code(t, mustErr(svc.Send(context.Background(), groupID, studentID, uuid.New(), long))); got != "MESSAGE_TOO_LONG" {
		t.Fatal(got)
	}
}

func mustErr(_ *MessageResponse, err error) error { return err }

func TestSend_RateLimited(t *testing.T) {
	svc, _, _ := setup()
	svc.limiter = allow{deny: true}
	_, err := svc.Send(context.Background(), groupID, studentID, uuid.New(), "hi")
	if !errors.Is(err, domain.ErrRateLimit) {
		t.Fatalf("want rate limit, got %v", err)
	}
}

func TestSend_RetryWithSameClientIDNoDuplicate(t *testing.T) {
	svc, repo, _ := setup()
	cid := uuid.New()
	a, _ := svc.Send(context.Background(), groupID, studentID, cid, "once")
	b, _ := svc.Send(context.Background(), groupID, studentID, cid, "once")
	if a.ID != b.ID || len(repo.msgs) != 1 {
		t.Fatalf("retry created a duplicate: %v %v (%d)", a.ID, b.ID, len(repo.msgs))
	}
}

func TestList_HidesExpired(t *testing.T) {
	svc, repo, _ := setup()
	_, _ = svc.Send(context.Background(), groupID, studentID, uuid.New(), "alive")
	old := uuid.New()
	repo.msgs[old] = &domain.ChatMessage{ID: old, GroupID: groupID, AuthorID: studentID, Body: "old", ExpiresAt: now.Add(-time.Minute)}

	ms, err := svc.List(context.Background(), groupID, studentID, nil, 0)
	if err != nil || len(ms) != 1 || ms[0].Body != "alive" {
		t.Fatalf("expired message visible: %+v %v", ms, err)
	}
	if n, _ := svc.DeleteExpired(context.Background()); n != 1 {
		t.Fatalf("cleanup removed %d, want 1", n)
	}
}

func TestDelete_Permissions(t *testing.T) {
	svc, _, _ := setup()
	m, _ := svc.Send(context.Background(), groupID, studentID, uuid.New(), "mine")
	own, _ := svc.Send(context.Background(), groupID, ownerID, uuid.New(), "tutor")

	if err := svc.Delete(context.Background(), groupID, own.ID, studentID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("student must not delete tutor's message: %v", err)
	}
	if err := svc.Delete(context.Background(), groupID, m.ID, ownerID); err != nil {
		t.Fatalf("owner deletes any message: %v", err)
	}
	if err := svc.Delete(context.Background(), groupID, own.ID, ownerID); err != nil {
		t.Fatal(err)
	}
}

func TestSettings_OwnerOnlyAndValidated(t *testing.T) {
	svc, repo, _ := setup()
	if err := svc.UpdateSettings(context.Background(), groupID, studentID, 3600, domain.ChatModeOpen); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("student changed settings: %v", err)
	}
	if got := code(t, svc.UpdateSettings(context.Background(), groupID, ownerID, 7*86400, domain.ChatModeOpen)); got != "CHAT_TTL_INVALID" {
		t.Fatal(got)
	}
	if err := svc.UpdateSettings(context.Background(), groupID, ownerID, 3600, domain.ChatModeOwnerOnly); err != nil {
		t.Fatal(err)
	}
	if repo.settings.ttl != 3600 || repo.settings.mode != domain.ChatModeOwnerOnly {
		t.Fatalf("settings not saved: %+v", repo.settings)
	}
}

func TestMuteAndBan(t *testing.T) {
	svc, repo, _ := setup()
	if got := code(t, svc.SetMuted(context.Background(), groupID, ownerID, otherID, true)); got != "NOT_GROUP_MEMBER" {
		t.Fatal(got)
	}
	if err := svc.SetMuted(context.Background(), groupID, ownerID, studentID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Send(context.Background(), groupID, studentID, uuid.New(), "hi"); code(t, err) != "CHAT_MUTED" {
		t.Fatal("muted student could write")
	}
	if err := svc.Ban(context.Background(), groupID, studentID, studentID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("student banned someone: %v", err)
	}
	if err := svc.Ban(context.Background(), groupID, ownerID, studentID); err != nil || !repo.banned[studentID] {
		t.Fatalf("ban failed: %v", err)
	}
}
