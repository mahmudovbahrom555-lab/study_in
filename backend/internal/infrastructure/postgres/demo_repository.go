package postgres

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)


// DemoRepository implements ai.DemoGroupCreator.
type DemoRepository struct{ db *sqlx.DB }

func NewDemoRepository(db *sqlx.DB) *DemoRepository {
	return &DemoRepository{db: db}
}

// CreateDemoGroup is idempotent: returns the existing demo group ID if one
// already exists for this teacher, otherwise creates a new one.
// The unique partial index idx_groups_teacher_demo (migration 000017) enforces
// the one-demo-per-teacher invariant at the DB level.
func (r *DemoRepository) CreateDemoGroup(ctx context.Context, teacherID uuid.UUID) (uuid.UUID, error) {
	// Return existing demo group if already created.
	var existing uuid.UUID
	err := r.db.GetContext(ctx, &existing,
		`SELECT id FROM groups WHERE teacher_id = $1 AND is_demo = TRUE AND deleted_at IS NULL`,
		teacherID)
	if err == nil {
		return existing, nil
	}

	id := uuid.New()
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO groups (id, teacher_id, name, subject, description, invite_code, is_demo, cefr_level)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE, 'B2')`,
		id, teacherID,
		"Demo группа — English B2",
		"English",
		"Демонстрационная группа. Здесь вы увидите, как платформа анализирует успеваемость.",
		generateInviteCode(),
	)
	if err != nil {
		// Unique index violation = concurrent request won the race; fetch their ID.
		var finalID uuid.UUID
		if ferr := r.db.GetContext(ctx, &finalID,
			`SELECT id FROM groups WHERE teacher_id = $1 AND is_demo = TRUE AND deleted_at IS NULL`,
			teacherID); ferr != nil {
			return uuid.Nil, fmt.Errorf("CreateDemoGroup insert: %w", err)
		}
		return finalID, nil
	}
	return id, nil
}

// generateInviteCode produces a cryptographically random 6-character code.
// Alphabet excludes visually ambiguous chars (0/O, l/1/I).
func generateInviteCode() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	alphabetLen := big.NewInt(int64(len(alphabet)))
	b := make([]byte, 6)
	for i := range b {
		n, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			panic("crypto/rand unavailable: " + err.Error())
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b)
}
