package notifications_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/notifications"
)

// ─── Mocks ───────────────────────────────────────────────────────────────────

type mockRepo struct {
	notifications []*domain.Notification
	tokens        map[string][]string // userID → tokens
}

func newMockRepo() *mockRepo {
	return &mockRepo{tokens: map[string][]string{}}
}

func (m *mockRepo) Create(_ context.Context, n *domain.Notification) error {
	n.ID = uuid.New()
	n.CreatedAt = time.Now()
	m.notifications = append(m.notifications, n)
	return nil
}

func (m *mockRepo) ListUnread(_ context.Context, userID uuid.UUID, limit int) ([]*domain.Notification, error) {
	var out []*domain.Notification
	for _, n := range m.notifications {
		if n.UserID == userID && !n.IsRead {
			out = append(out, n)
		}
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m *mockRepo) MarkRead(_ context.Context, id, userID uuid.UUID) error {
	for _, n := range m.notifications {
		if n.ID == id && n.UserID == userID {
			n.IsRead = true
		}
	}
	return nil
}

func (m *mockRepo) MarkAllRead(_ context.Context, userID uuid.UUID) error {
	for _, n := range m.notifications {
		if n.UserID == userID {
			n.IsRead = true
		}
	}
	return nil
}

func (m *mockRepo) CountUnread(_ context.Context, userID uuid.UUID) (int, error) {
	count := 0
	for _, n := range m.notifications {
		if n.UserID == userID && !n.IsRead {
			count++
		}
	}
	return count, nil
}

func (m *mockRepo) UpsertDeviceToken(_ context.Context, userID uuid.UUID, token, _ string) error {
	uid := userID.String()
	for _, t := range m.tokens[uid] {
		if t == token {
			return nil
		}
	}
	m.tokens[uid] = append(m.tokens[uid], token)
	return nil
}

func (m *mockRepo) DeleteDeviceToken(_ context.Context, token string) error {
	for uid, ts := range m.tokens {
		filtered := ts[:0]
		for _, t := range ts {
			if t != token {
				filtered = append(filtered, t)
			}
		}
		m.tokens[uid] = filtered
	}
	return nil
}

func (m *mockRepo) ListDeviceTokens(_ context.Context, userID uuid.UUID) ([]string, error) {
	return m.tokens[userID.String()], nil
}

type mockEnqueuer struct{ sent []string }

func (e *mockEnqueuer) EnqueuePush(_ context.Context, token, _, _ string, _ any) error {
	e.sent = append(e.sent, token)
	return nil
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestSend_CreatesNotificationAndEnqueuesPush(t *testing.T) {
	repo := newMockRepo()
	enq := &mockEnqueuer{}
	svc := notifications.NewService(repo, enq)

	userID := uuid.New()
	_ = repo.UpsertDeviceToken(context.Background(), userID, "tok-abc", "android")

	if err := svc.Send(context.Background(), userID, "New grade", "You got 95/100", nil); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if len(repo.notifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(repo.notifications))
	}
	if len(enq.sent) != 1 || enq.sent[0] != "tok-abc" {
		t.Fatalf("expected push to tok-abc, got %v", enq.sent)
	}
}

func TestList_ReturnsOnlyUnread(t *testing.T) {
	repo := newMockRepo()
	svc := notifications.NewService(repo, &mockEnqueuer{})

	userID := uuid.New()
	_ = svc.Send(context.Background(), userID, "A", "body", nil)
	_ = svc.Send(context.Background(), userID, "B", "body", nil)

	// Mark first as read
	_ = repo.MarkRead(context.Background(), repo.notifications[0].ID, userID)

	ns, err := svc.List(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(ns) != 1 || ns[0].Title != "B" {
		t.Fatalf("expected 1 unread notification 'B', got %v", ns)
	}
}

func TestMarkAllRead(t *testing.T) {
	repo := newMockRepo()
	svc := notifications.NewService(repo, &mockEnqueuer{})

	userID := uuid.New()
	_ = svc.Send(context.Background(), userID, "A", "body", nil)
	_ = svc.Send(context.Background(), userID, "B", "body", nil)

	if err := svc.MarkAllRead(context.Background(), userID); err != nil {
		t.Fatal(err)
	}

	count, _ := svc.UnreadCount(context.Background(), userID)
	if count != 0 {
		t.Fatalf("expected 0 unread, got %d", count)
	}
}

func TestRegisterDevice_DeduplicatesTokens(t *testing.T) {
	repo := newMockRepo()
	svc := notifications.NewService(repo, &mockEnqueuer{})

	userID := uuid.New()
	_ = svc.RegisterDevice(context.Background(), userID, "tok-xyz", "ios")
	_ = svc.RegisterDevice(context.Background(), userID, "tok-xyz", "ios")

	tokens, _ := repo.ListDeviceTokens(context.Background(), userID)
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token after duplicate register, got %d", len(tokens))
	}
}

func TestUnregisterDevice(t *testing.T) {
	repo := newMockRepo()
	svc := notifications.NewService(repo, &mockEnqueuer{})

	userID := uuid.New()
	_ = svc.RegisterDevice(context.Background(), userID, "tok-del", "android")
	_ = svc.UnregisterDevice(context.Background(), "tok-del")

	tokens, _ := repo.ListDeviceTokens(context.Background(), userID)
	if len(tokens) != 0 {
		t.Fatalf("expected 0 tokens after unregister, got %d", len(tokens))
	}
}
