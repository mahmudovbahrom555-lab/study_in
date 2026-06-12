package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// DemoRepository implements ai.DemoGroupCreator.
type DemoRepository struct{ db *sqlx.DB }

func NewDemoRepository(db *sqlx.DB) *DemoRepository {
	return &DemoRepository{db: db}
}

// CreateDemoGroup inserts a demo group for the given teacher and returns its ID.
// The group is marked is_demo=true and uses CEFR level B2 to give a realistic
// Course Coverage display in the hardcoded DemoInsights response.
func (r *DemoRepository) CreateDemoGroup(ctx context.Context, teacherID uuid.UUID) (uuid.UUID, error) {
	id := uuid.New()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO groups (id, teacher_id, name, subject, description, invite_code, is_demo, cefr_level)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE, 'B2')`,
		id, teacherID,
		"Demo группа — English B2",
		"English",
		"Демонстрационная группа. Здесь вы увидите, как платформа анализирует успеваемость.",
		generateInviteCode(),
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("CreateDemoGroup: %w", err)
	}
	return id, nil
}

// generateInviteCode produces a random 6-character code using a safe alphabet.
// Re-uses the same logic as groups/service.go to keep codes consistent.
func generateInviteCode() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = alphabet[uuid.New().ID()%uint32(len(alphabet))]
	}
	return string(b)
}
