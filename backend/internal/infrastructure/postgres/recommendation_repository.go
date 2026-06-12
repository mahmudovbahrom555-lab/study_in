package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type RecommendationRepository struct{ db *sqlx.DB }

func NewRecommendationRepository(db *sqlx.DB) *RecommendationRepository {
	return &RecommendationRepository{db: db}
}

// ReplaceForGroup atomically deletes pending recs for the group and inserts fresh ones.
// Already-acted recommendations (accepted/dismissed/snoozed) are preserved.
func (r *RecommendationRepository) ReplaceForGroup(ctx context.Context, groupID uuid.UUID, recs []*domain.AIRecommendation) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ReplaceForGroup begin: %w", err)
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx,
		`DELETE FROM ai_recommendations WHERE group_id=$1 AND status='pending'`, groupID); err != nil {
		return fmt.Errorf("ReplaceForGroup delete: %w", err)
	}

	for _, rec := range recs {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO ai_recommendations
			    (id, teacher_id, group_id, priority, action, topic, reason,
			     student_count, rule_key, rule_data, status, created_at, expires_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			rec.ID, rec.TeacherID, rec.GroupID, rec.Priority, rec.Action,
			rec.Topic, rec.Reason, rec.StudentCount, rec.RuleKey,
			rec.RuleData, rec.Status, rec.CreatedAt, rec.ExpiresAt,
		); err != nil {
			return fmt.Errorf("ReplaceForGroup insert %s: %w", rec.RuleKey, err)
		}
	}

	return tx.Commit()
}

// GetRecommendation fetches one recommendation by primary key.
func (r *RecommendationRepository) GetRecommendation(ctx context.Context, id uuid.UUID) (*domain.AIRecommendation, error) {
	var rec domain.AIRecommendation
	err := r.db.GetContext(ctx, &rec,
		`SELECT id,teacher_id,group_id,priority,action,topic,reason,student_count,
		        rule_key,rule_data,status,created_at,acted_at,expires_at
		 FROM ai_recommendations WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetRecommendation: %w", err)
	}
	return &rec, nil
}

// RecordAction updates status and records when the teacher acted.
func (r *RecommendationRepository) RecordAction(ctx context.Context, id uuid.UUID, status, teacherAction string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE ai_recommendations
		 SET status=$1, teacher_action=$2, acted_at=NOW()
		 WHERE id=$3 AND status='pending'`,
		status, teacherAction, id)
	return err
}

// PendingForOutcome returns acted-on recs that are 7+ days old and have no outcome measured.
func (r *RecommendationRepository) PendingForOutcome(ctx context.Context, groupID uuid.UUID) ([]*domain.AIRecommendation, error) {
	var recs []*domain.AIRecommendation
	err := r.db.SelectContext(ctx, &recs, `
		SELECT ar.id, ar.teacher_id, ar.group_id, ar.priority, ar.action,
		       ar.topic, ar.reason, ar.student_count, ar.rule_key,
		       ar.rule_data, ar.status, ar.created_at, ar.acted_at, ar.expires_at
		FROM ai_recommendations ar
		LEFT JOIN recommendation_outcomes ro ON ro.recommendation_id = ar.id
		WHERE ar.group_id = $1
		  AND ar.status IN ('accepted','dismissed','pending')
		  AND ar.created_at <= NOW() - INTERVAL '7 days'
		  AND ro.id IS NULL`, groupID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("PendingForOutcome: %w", err)
	}
	return recs, nil
}

// SaveOutcome inserts a measured outcome; safe to call once per recommendation.
func (r *RecommendationRepository) SaveOutcome(ctx context.Context, o *domain.RecommendationOutcome) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO recommendation_outcomes
		    (id, recommendation_id, measured_at, mastery_delta, confidence_delta,
		     consistency_delta, students_improved, students_total)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (recommendation_id) DO NOTHING`,
		o.ID, o.RecommendationID, o.MeasuredAt,
		o.MasteryDelta, o.ConfidenceDelta, o.ConsistencyDelta,
		o.StudentsImproved, o.StudentsTotal)
	return err
}

// MeasureOutcomeForTopic implements RecommendationRepository.
// Computes how a topic's avg accuracy changed relative to snapshotAccuracy.
func (r *RecommendationRepository) MeasureOutcomeForTopic(ctx context.Context, groupID uuid.UUID, topic string, snapshotAccuracy float64) *domain.RecommendationOutcome {
	type row struct {
		AvgAccuracy      float64 `db:"avg_accuracy"`
		StudentsImproved int     `db:"students_improved"`
		Total            int     `db:"total"`
	}
	var res row
	err := r.db.GetContext(ctx, &res, `
		SELECT
		    AVG(CASE WHEN tm.total_count > 0
		            THEN tm.correct_count::float / tm.total_count ELSE 0 END) AS avg_accuracy,
		    COUNT(*) FILTER (WHERE tm.total_count > 0 AND
		        tm.correct_count::float / tm.total_count > $3 + 0.05)        AS students_improved,
		    COUNT(*) AS total
		FROM group_members gm
		JOIN topic_mastery tm ON tm.student_id = gm.student_id
		JOIN topic_tags tt ON tt.id = tm.topic_id
		WHERE gm.group_id = $1 AND tt.name = $2`, groupID, topic, snapshotAccuracy)
	if err != nil || res.Total == 0 {
		return nil
	}
	delta := res.AvgAccuracy - snapshotAccuracy
	o := &domain.RecommendationOutcome{
		ID:               uuid.New(),
		MeasuredAt:       time.Now(),
		StudentsImproved: res.StudentsImproved,
		StudentsTotal:    res.Total,
	}
	o.MasteryDelta = &delta
	return o
}
