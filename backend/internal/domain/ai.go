package domain

import (
	"time"

	"github.com/google/uuid"
)

// AIDocumentStatus — жизненный цикл загруженного документа.
type AIDocumentStatus string

const (
	AIDocStatusPending    AIDocumentStatus = "pending"
	AIDocStatusProcessing AIDocumentStatus = "processing"
	AIDocStatusReady      AIDocumentStatus = "ready"
	AIDocStatusFailed     AIDocumentStatus = "failed"
)

type AIDocument struct {
	ID         uuid.UUID        `db:"id"`
	GroupID    *uuid.UUID       `db:"group_id"`
	TeacherID  uuid.UUID        `db:"teacher_id"`
	Title      string           `db:"title"`
	ObjectKey  string           `db:"object_key"`
	MimeType   string           `db:"mime_type"`
	SizeBytes  int64            `db:"size_bytes"`
	Status     AIDocumentStatus `db:"status"`
	ChunkCount int              `db:"chunk_count"`
	ErrorMsg   *string          `db:"error_msg"`
	CreatedAt  time.Time        `db:"created_at"`
}

type AIDocumentChunk struct {
	ID         uuid.UUID `db:"id"`
	DocumentID uuid.UUID `db:"document_id"`
	ChunkIndex int       `db:"chunk_index"`
	Content    string    `db:"content"`
	Embedding  []float32 `db:"-"` // handled manually via pgvector
	TokenCount int       `db:"token_count"`
	CreatedAt  time.Time `db:"created_at"`
}

// AIJobStatus — состояние фоновой задачи.
type AIJobStatus string

const (
	AIJobPending    AIJobStatus = "pending"
	AIJobProcessing AIJobStatus = "processing"
	AIJobDone       AIJobStatus = "done"
	AIJobFailed     AIJobStatus = "failed"
)

type AIJob struct {
	ID        uuid.UUID   `db:"id"`
	Type      string      `db:"type"`
	UserID    uuid.UUID   `db:"user_id"`
	RefID     *uuid.UUID  `db:"ref_id"`
	Status    AIJobStatus `db:"status"`
	Result    []byte      `db:"result"` // JSONB
	ErrorMsg  *string     `db:"error_msg"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`
}

type AISession struct {
	ID          uuid.UUID  `db:"id"`
	UserID      uuid.UUID  `db:"user_id"`
	SessionType string     `db:"session_type"`
	DocumentID  *uuid.UUID `db:"document_id"`
	Subject     *string    `db:"subject"`
	CEFRLevel   *string    `db:"cefr_level"`
	Metadata    []byte     `db:"metadata"` // JSONB
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type AIMessage struct {
	ID        uuid.UUID `db:"id"`
	SessionID uuid.UUID `db:"session_id"`
	Role      string    `db:"role"`
	Content   string    `db:"content"`
	TokensIn  int       `db:"tokens_in"`
	TokensOut int       `db:"tokens_out"`
	Model     string    `db:"model"`
	CreatedAt time.Time `db:"created_at"`
}

type TopicTag struct {
	ID      uuid.UUID `db:"id"`
	Name    string    `db:"name"`
	Subject *string   `db:"subject"`
}

type TopicMastery struct {
	ID           uuid.UUID `db:"id"`
	StudentID    uuid.UUID `db:"student_id"`
	TopicID      uuid.UUID `db:"topic_id"`
	CorrectCount int       `db:"correct_count"`
	TotalCount   int       `db:"total_count"`
	NextReview   time.Time `db:"next_review"`
	IntervalDays int       `db:"interval_days"`
	EaseFactor   float64   `db:"ease_factor"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type StudentGamification struct {
	StudentID        uuid.UUID  `db:"student_id"`
	XPTotal          int        `db:"xp_total"`
	XPToday          int        `db:"xp_today"`
	DailyGoalXP      int        `db:"daily_goal_xp"`
	StreakDays       int        `db:"streak_days"`
	LongestStreak    int        `db:"longest_streak"`
	LastActivityDate *time.Time `db:"last_activity_date"`
	UpdatedAt        time.Time  `db:"updated_at"`
}

type XPEvent struct {
	ID        uuid.UUID `db:"id"`
	StudentID uuid.UUID `db:"student_id"`
	Amount    int       `db:"amount"`
	Reason    string    `db:"reason"`
	CreatedAt time.Time `db:"created_at"`
}

type SkillAssessment struct {
	ID         uuid.UUID `db:"id"`
	StudentID  uuid.UUID `db:"student_id"`
	Skill      string    `db:"skill"`
	CEFRLevel  *string   `db:"cefr_level"`
	Score      *float64  `db:"score"`
	Feedback   *string   `db:"feedback"`
	AudioKey   *string   `db:"audio_key"`
	Transcript *string   `db:"transcript"`
	CreatedAt  time.Time `db:"created_at"`
}

type AITokenUsage struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Model     string    `db:"model"`
	TokensIn  int       `db:"tokens_in"`
	TokensOut int       `db:"tokens_out"`
	Feature   string    `db:"feature"`
	CostUSD   float64   `db:"cost_usd"`
	CreatedAt time.Time `db:"created_at"`
}

// XP award constants (Duolingo-style).
const (
	XPQuizCorrect       = 10
	XPWeakTopicCorrect  = 20
	XPAssignmentOnTime  = 15
	XPWritingChecked    = 25
	XPSpeakingCompleted = 30
	XPStreakBonus7Days  = 100
	XPDailyGoalReached  = 50
	XPPerfectScore      = 50
)
