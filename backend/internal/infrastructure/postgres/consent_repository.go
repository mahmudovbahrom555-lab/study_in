package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type ConsentRepository struct{ db *sqlx.DB }

func NewConsentRepository(db *sqlx.DB) *ConsentRepository {
	return &ConsentRepository{db: db}
}

// RecordConsent appends an immutable consent event.
// Revoking consent = calling this with granted=false.
func (r *ConsentRepository) RecordConsent(ctx context.Context, e *domain.ConsentEvent) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO consent_events
		    (id, user_id, consent_type, consent_version, granted, ip_address, user_agent, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		e.ID, e.UserID, e.ConsentType, e.ConsentVersion,
		e.Granted, e.IPAddress, e.UserAgent, e.CreatedAt)
	if err != nil {
		return fmt.Errorf("RecordConsent: %w", err)
	}
	return nil
}

// LatestConsent returns the most recent event for a given user+type, or nil if none.
func (r *ConsentRepository) LatestConsent(ctx context.Context, userID interface{ String() string }, consentType string) (*domain.ConsentEvent, error) {
	var e domain.ConsentEvent
	err := r.db.GetContext(ctx, &e, `
		SELECT id, user_id, consent_type, consent_version, granted, ip_address, user_agent, created_at
		FROM consent_events
		WHERE user_id = $1 AND consent_type = $2
		ORDER BY created_at DESC
		LIMIT 1`, userID, consentType)
	if err != nil {
		return nil, nil // no consent on record = not granted
	}
	return &e, nil
}
