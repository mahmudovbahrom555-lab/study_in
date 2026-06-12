package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	pgvector "github.com/pgvector/pgvector-go"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// ─── AIDocumentRepository ─────────────────────────────────────────────────────

type AIDocumentRepository struct{ db *sqlx.DB }

func NewAIDocumentRepository(db *sqlx.DB) *AIDocumentRepository {
	return &AIDocumentRepository{db: db}
}

func (r *AIDocumentRepository) CreateDocument(ctx context.Context, doc *domain.AIDocument) error {
	q := `INSERT INTO ai_documents
	      (id, group_id, teacher_id, title, object_key, mime_type, size_bytes, status, chunk_count, created_at)
	      VALUES (:id, :group_id, :teacher_id, :title, :object_key, :mime_type, :size_bytes, :status, :chunk_count, :created_at)`
	_, err := r.db.NamedExecContext(ctx, q, doc)
	return err
}

func (r *AIDocumentRepository) GetDocument(ctx context.Context, id uuid.UUID) (*domain.AIDocument, error) {
	var doc domain.AIDocument
	err := r.db.GetContext(ctx, &doc, `SELECT * FROM ai_documents WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &doc, err
}

func (r *AIDocumentRepository) ListDocuments(ctx context.Context, teacherID uuid.UUID, groupID *uuid.UUID) ([]*domain.AIDocument, error) {
	var docs []*domain.AIDocument
	if groupID != nil {
		err := r.db.SelectContext(ctx, &docs,
			`SELECT * FROM ai_documents WHERE teacher_id=$1 AND group_id=$2 ORDER BY created_at DESC`,
			teacherID, *groupID)
		return docs, err
	}
	err := r.db.SelectContext(ctx, &docs,
		`SELECT * FROM ai_documents WHERE teacher_id=$1 ORDER BY created_at DESC`, teacherID)
	return docs, err
}

func (r *AIDocumentRepository) UpdateDocumentStatus(ctx context.Context, id uuid.UUID, status domain.AIDocumentStatus, chunkCount int, errMsg *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE ai_documents SET status=$1, chunk_count=$2, error_msg=$3 WHERE id=$4`,
		status, chunkCount, errMsg, id)
	return err
}

func (r *AIDocumentRepository) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM ai_documents WHERE id=$1`, id)
	return err
}

func (r *AIDocumentRepository) SaveChunk(ctx context.Context, chunk *domain.AIDocumentChunk, embedding []float32) error {
	var vec *pgvector.Vector
	if len(embedding) > 0 {
		v := pgvector.NewVector(embedding)
		vec = &v
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO ai_document_chunks (id, document_id, chunk_index, content, embedding, token_count, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		chunk.ID, chunk.DocumentID, chunk.ChunkIndex, chunk.Content, vec, chunk.TokenCount, chunk.CreatedAt)
	return err
}

// SearchSimilarChunks returns chunks ordered by cosine similarity to queryEmbedding.
// Falls back to ordering by chunk_index when embedding is nil (no API key).
func (r *AIDocumentRepository) SearchSimilarChunks(ctx context.Context, documentID uuid.UUID, queryEmbedding []float32, limit int) ([]*domain.AIDocumentChunk, error) {
	if limit <= 0 {
		limit = 10
	}

	var rows []struct {
		ID         uuid.UUID `db:"id"`
		DocumentID uuid.UUID `db:"document_id"`
		ChunkIndex int       `db:"chunk_index"`
		Content    string    `db:"content"`
		TokenCount int       `db:"token_count"`
		CreatedAt  time.Time `db:"created_at"`
	}

	var err error
	if len(queryEmbedding) > 0 {
		vec := pgvector.NewVector(queryEmbedding)
		err = r.db.SelectContext(ctx, &rows,
			`SELECT id, document_id, chunk_index, content, token_count, created_at
			 FROM ai_document_chunks
			 WHERE document_id=$1 AND embedding IS NOT NULL
			 ORDER BY embedding <=> $2
			 LIMIT $3`,
			documentID, vec, limit)
	} else {
		err = r.db.SelectContext(ctx, &rows,
			`SELECT id, document_id, chunk_index, content, token_count, created_at
			 FROM ai_document_chunks
			 WHERE document_id=$1
			 ORDER BY chunk_index ASC
			 LIMIT $2`,
			documentID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("SearchSimilarChunks: %w", err)
	}

	out := make([]*domain.AIDocumentChunk, len(rows))
	for i, row := range rows {
		out[i] = &domain.AIDocumentChunk{
			ID:         row.ID,
			DocumentID: row.DocumentID,
			ChunkIndex: row.ChunkIndex,
			Content:    row.Content,
			TokenCount: row.TokenCount,
			CreatedAt:  row.CreatedAt,
		}
	}
	return out, nil
}

// ─── AIJobRepository ──────────────────────────────────────────────────────────

type AIJobRepository struct{ db *sqlx.DB }

func NewAIJobRepository(db *sqlx.DB) *AIJobRepository {
	return &AIJobRepository{db: db}
}

func (r *AIJobRepository) CreateJob(ctx context.Context, job *domain.AIJob) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO ai_jobs (id, type, user_id, ref_id, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		job.ID, job.Type, job.UserID, job.RefID, job.Status, job.CreatedAt, job.UpdatedAt)
	return err
}

func (r *AIJobRepository) GetJob(ctx context.Context, id uuid.UUID) (*domain.AIJob, error) {
	var job domain.AIJob
	err := r.db.GetContext(ctx, &job, `SELECT * FROM ai_jobs WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &job, err
}

func (r *AIJobRepository) UpdateJob(ctx context.Context, id uuid.UUID, status domain.AIJobStatus, result []byte, errMsg *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE ai_jobs SET status=$1, result=$2, error_msg=$3, updated_at=NOW() WHERE id=$4`,
		status, result, errMsg, id)
	return err
}

// ─── AISessionRepository ──────────────────────────────────────────────────────

type AISessionRepository struct{ db *sqlx.DB }

func NewAISessionRepository(db *sqlx.DB) *AISessionRepository {
	return &AISessionRepository{db: db}
}

func (r *AISessionRepository) CreateSession(ctx context.Context, s *domain.AISession) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO ai_sessions (id, user_id, session_type, document_id, subject, cefr_level, metadata, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		s.ID, s.UserID, s.SessionType, s.DocumentID, s.Subject, s.CEFRLevel,
		s.Metadata, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *AISessionRepository) GetSession(ctx context.Context, id uuid.UUID) (*domain.AISession, error) {
	var s domain.AISession
	err := r.db.GetContext(ctx, &s, `SELECT * FROM ai_sessions WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *AISessionRepository) ListUserSessions(ctx context.Context, userID uuid.UUID) ([]*domain.AISession, error) {
	var sessions []*domain.AISession
	err := r.db.SelectContext(ctx, &sessions,
		`SELECT * FROM ai_sessions WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	return sessions, err
}

func (r *AISessionRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM ai_sessions WHERE id=$1`, id)
	return err
}

func (r *AISessionRepository) AddMessage(ctx context.Context, m *domain.AIMessage) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO ai_messages (id, session_id, role, content, tokens_in, tokens_out, model, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		m.ID, m.SessionID, m.Role, m.Content, m.TokensIn, m.TokensOut, m.Model, m.CreatedAt)
	return err
}

func (r *AISessionRepository) ListMessages(ctx context.Context, sessionID uuid.UUID) ([]*domain.AIMessage, error) {
	var msgs []*domain.AIMessage
	err := r.db.SelectContext(ctx, &msgs,
		`SELECT * FROM ai_messages WHERE session_id=$1 ORDER BY created_at ASC`, sessionID)
	return msgs, err
}

// ─── AITokenRepository ────────────────────────────────────────────────────────

type AITokenRepository struct{ db *sqlx.DB }

func NewAITokenRepository(db *sqlx.DB) *AITokenRepository {
	return &AITokenRepository{db: db}
}

func (r *AITokenRepository) LogUsage(ctx context.Context, u *domain.AITokenUsage) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO ai_token_usage (id, user_id, model, tokens_in, tokens_out, feature, cost_usd, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		u.ID, u.UserID, u.Model, u.TokensIn, u.TokensOut, u.Feature, u.CostUSD, u.CreatedAt)
	return err
}

func (r *AITokenRepository) MonthlyTokens(ctx context.Context, userID uuid.UUID) (int, error) {
	var total int
	err := r.db.GetContext(ctx, &total,
		`SELECT COALESCE(SUM(tokens_in + tokens_out),0) FROM ai_token_usage
		 WHERE user_id=$1 AND created_at >= date_trunc('month', NOW())`, userID)
	return total, err
}

// ─── AIMasteryRepository ──────────────────────────────────────────────────────

type AIMasteryRepository struct{ db *sqlx.DB }

func NewAIMasteryRepository(db *sqlx.DB) *AIMasteryRepository {
	return &AIMasteryRepository{db: db}
}

func (r *AIMasteryRepository) UpsertMastery(ctx context.Context, m *domain.TopicMastery) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO topic_mastery
		    (id, student_id, topic_id, correct_count, total_count, next_review,
		     interval_days, ease_factor, confidence_score, consistency_score,
		     correct_streak, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		 ON CONFLICT (student_id, topic_id) DO UPDATE SET
		    correct_count=EXCLUDED.correct_count,
		    total_count=EXCLUDED.total_count,
		    next_review=EXCLUDED.next_review,
		    interval_days=EXCLUDED.interval_days,
		    ease_factor=EXCLUDED.ease_factor,
		    confidence_score=EXCLUDED.confidence_score,
		    consistency_score=EXCLUDED.consistency_score,
		    correct_streak=EXCLUDED.correct_streak,
		    updated_at=EXCLUDED.updated_at`,
		m.ID, m.StudentID, m.TopicID, m.CorrectCount, m.TotalCount, m.NextReview,
		m.IntervalDays, m.EaseFactor, m.ConfidenceScore, m.ConsistencyScore,
		m.CorrectStreak, m.UpdatedAt)
	return err
}

func (r *AIMasteryRepository) GetMastery(ctx context.Context, studentID, topicID uuid.UUID) (*domain.TopicMastery, error) {
	var m domain.TopicMastery
	err := r.db.GetContext(ctx, &m,
		`SELECT * FROM topic_mastery WHERE student_id=$1 AND topic_id=$2`, studentID, topicID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &m, err
}

func (r *AIMasteryRepository) ListDueMastery(ctx context.Context, studentID uuid.UUID, limit int) ([]*domain.TopicMastery, error) {
	var entries []*domain.TopicMastery
	err := r.db.SelectContext(ctx, &entries,
		`SELECT * FROM topic_mastery
		 WHERE student_id=$1 AND next_review <= NOW()
		 ORDER BY next_review ASC LIMIT $2`, studentID, limit)
	return entries, err
}

func (r *AIMasteryRepository) ListWeakTopics(ctx context.Context, studentID uuid.UUID, limit int) ([]*domain.TopicMastery, error) {
	var entries []*domain.TopicMastery
	err := r.db.SelectContext(ctx, &entries,
		`SELECT * FROM topic_mastery
		 WHERE student_id=$1 AND total_count > 0
		 ORDER BY (correct_count::float / total_count) ASC LIMIT $2`, studentID, limit)
	return entries, err
}

// ─── AIGamificationRepository ─────────────────────────────────────────────────

type AIGamificationRepository struct{ db *sqlx.DB }

func NewAIGamificationRepository(db *sqlx.DB) *AIGamificationRepository {
	return &AIGamificationRepository{db: db}
}

func (r *AIGamificationRepository) GetOrCreate(ctx context.Context, studentID uuid.UUID) (*domain.StudentGamification, error) {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO student_gamification (student_id) VALUES ($1) ON CONFLICT DO NOTHING`, studentID)
	if err != nil {
		return nil, err
	}
	var g domain.StudentGamification
	err = r.db.GetContext(ctx, &g, `SELECT * FROM student_gamification WHERE student_id=$1`, studentID)
	return &g, err
}

func (r *AIGamificationRepository) AddXP(ctx context.Context, studentID uuid.UUID, amount int, reason string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO xp_events (id, student_id, amount, reason) VALUES (gen_random_uuid(),$1,$2,$3)`,
		studentID, amount, reason)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE student_gamification SET xp_total=xp_total+$1, xp_today=xp_today+$1, updated_at=NOW()
		 WHERE student_id=$2`, amount, studentID)
	return err
}

func (r *AIGamificationRepository) UpdateStreak(ctx context.Context, studentID uuid.UUID) error {
	today := time.Now().Format("2006-01-02")
	_, err := r.db.ExecContext(ctx, `
		UPDATE student_gamification SET
		    streak_days = CASE
		        WHEN last_activity_date = $2::date - interval '1 day' THEN streak_days + 1
		        WHEN last_activity_date = $2::date THEN streak_days
		        ELSE 1
		    END,
		    longest_streak = GREATEST(longest_streak,
		        CASE
		            WHEN last_activity_date = $2::date - interval '1 day' THEN streak_days + 1
		            ELSE 1
		        END),
		    last_activity_date = $2::date,
		    updated_at = NOW()
		WHERE student_id = $1`, studentID, today)
	return err
}

// ─── AITopicRepository ────────────────────────────────────────────────────────

type AITopicRepository struct{ db *sqlx.DB }

func NewAITopicRepository(db *sqlx.DB) *AITopicRepository {
	return &AITopicRepository{db: db}
}

func (r *AITopicRepository) UpsertTopic(ctx context.Context, name string, subject *string) (*domain.TopicTag, error) {
	var tag domain.TopicTag
	err := r.db.GetContext(ctx, &tag,
		`INSERT INTO topic_tags (id, name, subject) VALUES (gen_random_uuid(),$1,$2)
		 ON CONFLICT (name, subject) DO UPDATE SET name=EXCLUDED.name
		 RETURNING *`, name, subject)
	return &tag, err
}

func (r *AITopicRepository) GetTopic(ctx context.Context, id uuid.UUID) (*domain.TopicTag, error) {
	var tag domain.TopicTag
	err := r.db.GetContext(ctx, &tag, `SELECT * FROM topic_tags WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &tag, err
}

func (r *AITopicRepository) LinkQuestionTopics(ctx context.Context, questionID uuid.UUID, topicIDs []uuid.UUID) error {
	for _, tid := range topicIDs {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO question_topics (question_id, topic_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			questionID, tid)
		if err != nil {
			return err
		}
	}
	return nil
}

