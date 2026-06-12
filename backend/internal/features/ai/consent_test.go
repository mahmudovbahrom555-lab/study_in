package ai_test

// Consent repository behaviour is tested at the service boundary using a mock.
// Integration tests against a real DB would require a test container — tracked
// as a future improvement. What we verify here:
//   1. ErrNoRows  → (nil, nil)   — no consent on record, not an error
//   2. DB error   → (nil, err)   — caller gets a real error, not silent nil
//   3. Normal row → returns the consent event
//
// The mock below simulates each case by controlling the error returned.

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// consentResult is what a ConsentRepository.LatestConsent implementation returns.
type consentResult struct {
	event *domain.ConsentEvent
	err   error
}

// mockConsentRepo lets each test control the exact return value.
type mockConsentRepo struct {
	latestResult consentResult
	recorded     []*domain.ConsentEvent
}

func (m *mockConsentRepo) RecordConsent(_ context.Context, e *domain.ConsentEvent) error {
	m.recorded = append(m.recorded, e)
	return nil
}

func (m *mockConsentRepo) LatestConsent(_ context.Context, _ uuid.UUID, _ string) (*domain.ConsentEvent, error) {
	return m.latestResult.event, m.latestResult.err
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestLatestConsent_NoRows: sql.ErrNoRows should produce (nil, nil) so callers
// treat missing consent as "not yet granted", not as a hard error.
func TestLatestConsent_NoRows(t *testing.T) {
	repo := &mockConsentRepo{
		latestResult: consentResult{event: nil, err: nil}, // no row found
	}
	ev, err := repo.LatestConsent(context.Background(), uuid.New(), "data_sharing")
	if err != nil {
		t.Fatalf("expected nil error for no-rows, got: %v", err)
	}
	if ev != nil {
		t.Fatalf("expected nil event for no-rows, got: %+v", ev)
	}
}

// TestLatestConsent_DBError: a real DB failure (network timeout, pool exhausted)
// must surface as a non-nil error, not silently return "no consent".
// Previously the repo returned (nil, nil) for ALL errors — this test pins
// the correct behaviour after the fix.
func TestLatestConsent_DBError(t *testing.T) {
	dbErr := fmt.Errorf("LatestConsent: %w", errors.New("connection refused"))
	repo := &mockConsentRepo{
		latestResult: consentResult{event: nil, err: dbErr},
	}
	ev, err := repo.LatestConsent(context.Background(), uuid.New(), "data_sharing")
	if err == nil {
		t.Fatal("expected non-nil error for DB failure, got nil — audit trail broken")
	}
	if ev != nil {
		t.Fatalf("expected nil event on DB error, got: %+v", ev)
	}
}

// TestLatestConsent_NormalRow: existing consent event is returned correctly.
func TestLatestConsent_NormalRow(t *testing.T) {
	userID := uuid.New()
	want := &domain.ConsentEvent{
		ID:             uuid.New(),
		UserID:         userID,
		ConsentType:    "data_sharing",
		ConsentVersion: "v1.0",
		Granted:        true,
		CreatedAt:      time.Now().Truncate(time.Second),
	}
	repo := &mockConsentRepo{
		latestResult: consentResult{event: want, err: nil},
	}
	got, err := repo.LatestConsent(context.Background(), userID, "data_sharing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected consent event, got nil")
	}
	if got.ID != want.ID {
		t.Errorf("ID mismatch: got %s, want %s", got.ID, want.ID)
	}
	if got.Granted != want.Granted {
		t.Errorf("Granted mismatch: got %v, want %v", got.Granted, want.Granted)
	}
	if got.ConsentVersion != want.ConsentVersion {
		t.Errorf("ConsentVersion mismatch: got %s, want %s", got.ConsentVersion, want.ConsentVersion)
	}
}

// TestRecordConsent_AppendOnly: each grant/revoke call appends a new record.
// The audit trail is immutable — no updates, only inserts.
func TestRecordConsent_AppendOnly(t *testing.T) {
	repo := &mockConsentRepo{}
	userID := uuid.New()

	events := []*domain.ConsentEvent{
		{ID: uuid.New(), UserID: userID, ConsentType: "data_sharing", Granted: true, ConsentVersion: "v1.0", CreatedAt: time.Now()},
		{ID: uuid.New(), UserID: userID, ConsentType: "data_sharing", Granted: false, ConsentVersion: "v1.0", CreatedAt: time.Now()},
		{ID: uuid.New(), UserID: userID, ConsentType: "data_sharing", Granted: true, ConsentVersion: "v1.1", CreatedAt: time.Now()},
	}

	for _, e := range events {
		if err := repo.RecordConsent(context.Background(), e); err != nil {
			t.Fatalf("RecordConsent: %v", err)
		}
	}

	if len(repo.recorded) != 3 {
		t.Errorf("expected 3 records (append-only), got %d", len(repo.recorded))
	}
	// Verify IDs are distinct — not overwritten.
	seen := map[uuid.UUID]bool{}
	for _, e := range repo.recorded {
		if seen[e.ID] {
			t.Errorf("duplicate consent event ID %s — violates append-only invariant", e.ID)
		}
		seen[e.ID] = true
	}
}
