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
	ID               uuid.UUID `db:"id"`
	StudentID        uuid.UUID `db:"student_id"`
	TopicID          uuid.UUID `db:"topic_id"`
	CorrectCount     int       `db:"correct_count"`
	TotalCount       int       `db:"total_count"`
	NextReview       time.Time `db:"next_review"`
	IntervalDays     int       `db:"interval_days"`
	EaseFactor       float64   `db:"ease_factor"`
	ConfidenceScore  float64   `db:"confidence_score"`  // [0,1] EMA of answer confidence
	ConsistencyScore float64   `db:"consistency_score"` // [0,1] variance-based reliability
	CorrectStreak    int       `db:"correct_streak"`
	UpdatedAt        time.Time `db:"updated_at"`
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

// AIGenerationSession records one quiz-generation event end-to-end.
// Persisted at generation time; accepted/rejected counts updated lazily from feedback.
// Over time becomes the dataset for model quality analysis and TAR trend tracking.
type AIGenerationSession struct {
	ID               uuid.UUID  `db:"id"`
	TeacherID        uuid.UUID  `db:"teacher_id"`
	SourceDocumentID *uuid.UUID `db:"source_document_id"`
	GroupID          *uuid.UUID `db:"group_id"`
	QuizID           *uuid.UUID `db:"quiz_id"`
	GeneratedCount   int        `db:"generated_count"`
	AcceptedCount    int        `db:"accepted_count"`
	EditedCount      int        `db:"edited_count"`
	RejectedCount    int        `db:"rejected_count"`
	CEFRLevel        *string    `db:"cefr_level"`
	Subject          *string    `db:"subject"`
	ModelUsed        string     `db:"model_used"`
	PromptTokens     int        `db:"prompt_tokens"`
	CompletionTokens int        `db:"completion_tokens"`
	CreatedAt        time.Time  `db:"created_at"`
	CompletedAt      *time.Time `db:"completed_at"`
}

// TeacherGenerationStats aggregates all generation sessions for one teacher.
type TeacherGenerationStats struct {
	TotalSessions  int     `json:"total_sessions"`
	TotalGenerated int     `json:"total_generated"`
	TotalAccepted  int     `json:"total_accepted"`
	TotalEdited    int     `json:"total_edited"`
	TotalRejected  int     `json:"total_rejected"`
	AcceptanceRate float64 `json:"acceptance_rate"`
}

// QuizQuestionFeedback tracks teacher acceptance/rejection of AI-generated questions.
// Used to compute Teacher Acceptance Rate (TAR) — the primary quality metric.
type QuizQuestionFeedback struct {
	ID         uuid.UUID `db:"id"`
	QuestionID uuid.UUID `db:"question_id"`
	TeacherID  uuid.UUID `db:"teacher_id"`
	Accepted   bool      `db:"accepted"`
	CreatedAt  time.Time `db:"created_at"`
}

// ─── Analytics DTOs (read-only, not persisted directly) ──────────────────────

// TeacherRecommendation is a "Next Best Action" item shown on the Teacher Dashboard.
// Persisted in ai_recommendations; ID lets the teacher accept, dismiss, or explain it.
type TeacherRecommendation struct {
	ID           uuid.UUID `json:"id"`
	Priority     int       `json:"priority"`      // 1=high, 2=medium, 3=low
	Action       string    `json:"action"`        // create_quiz | schedule_review | check_students | rate_questions | celebrate
	Topic        string    `json:"topic,omitempty"`
	Reason       string    `json:"reason"`
	StudentCount int       `json:"student_count,omitempty"`
}

// AIRecommendation is the persisted form stored in ai_recommendations.
type AIRecommendation struct {
	ID           uuid.UUID  `db:"id"`
	TeacherID    uuid.UUID  `db:"teacher_id"`
	GroupID      *uuid.UUID `db:"group_id"`
	Priority     int        `db:"priority"`
	Action       string     `db:"action"`
	Topic        string     `db:"topic"`
	Reason       string     `db:"reason"`
	StudentCount int        `db:"student_count"`
	RuleKey      string     `db:"rule_key"`
	RuleData     []byte     `db:"rule_data"` // JSONB snapshot of metrics at creation time
	Status       string     `db:"status"`
	CreatedAt    time.Time  `db:"created_at"`
	ActedAt      *time.Time `db:"acted_at"`
	ExpiresAt    time.Time  `db:"expires_at"`
}

// RecommendationOutcome records the measurable impact 7 days after a recommendation.
type RecommendationOutcome struct {
	ID               uuid.UUID `db:"id"`
	RecommendationID uuid.UUID `db:"recommendation_id"`
	MeasuredAt       time.Time `db:"measured_at"`
	MasteryDelta     *float64  `db:"mastery_delta"`
	ConfidenceDelta  *float64  `db:"confidence_delta"`
	ConsistencyDelta *float64  `db:"consistency_delta"`
	StudentsImproved int       `db:"students_improved"`
	StudentsTotal    int       `db:"students_total"`
}

type ClassInsights struct {
	GroupID         uuid.UUID               `json:"group_id"`
	Period          string                  `json:"period"`
	StudentCount    int                     `json:"student_count"`
	ClassWeakness   []TopicWeakness         `json:"class_weakness"`
	Students        []StudentSummary        `json:"students"`
	QuizStats       QuizStats               `json:"quiz_stats"`
	Recommendations []TeacherRecommendation `json:"recommendations"`
}

type TopicWeakness struct {
	Topic              string  `json:"topic"`
	AvgAccuracy        float64 `json:"avg_accuracy"`
	StudentsStruggling int     `json:"students_struggling"`
	TotalStudents      int     `json:"total_students"`
	AvgConfidence      float64 `json:"avg_confidence"`
	AvgConsistency     float64 `json:"avg_consistency"`
}

type StudentSummary struct {
	StudentID    uuid.UUID `json:"student_id"`
	Name         string    `json:"name"`
	XPTotal      int       `json:"xp"`
	StreakDays   int       `json:"streak"`
	TopWeakness  string    `json:"top_weakness"`
	LastActive   *string   `json:"last_active"`
	IsAtRisk     bool      `json:"is_at_risk"` // no activity for 3+ days
}

type QuizStats struct {
	Generated      int     `json:"generated"`
	AcceptanceRate float64 `json:"acceptance_rate"`
	TotalAttempts  int     `json:"total_attempts"`
	AvgScore       float64 `json:"avg_score"`
}

type StudentProgress struct {
	StudentID    uuid.UUID             `json:"student_id"`
	Name         string                `json:"name"`
	Gamification *StudentGamification  `json:"gamification"`
	WeakTopics   []StudentTopicDetail  `json:"weak_topics"`
	QuizHistory  []QuizAttemptSummary  `json:"quiz_history"`
	SkillLevels  map[string]SkillLevel `json:"skill_levels"`
}

type StudentTopicDetail struct {
	TopicID          uuid.UUID `json:"topic_id"`
	TopicName        string    `json:"topic_name"`
	Accuracy         float64   `json:"accuracy"`
	TotalAnswers     int       `json:"total_answers"`
	NextReview       time.Time `json:"next_review"`
	IntervalDays     int       `json:"interval_days"`
	ConfidenceScore  float64   `json:"confidence_score"`
	ConsistencyScore float64   `json:"consistency_score"`
}

type QuizAttemptSummary struct {
	QuizTitle  string    `json:"quiz_title"`
	Score      int16     `json:"score"`
	MaxScore   int16     `json:"max_score"`
	Percentage float64   `json:"percentage"`
	FinishedAt time.Time `json:"finished_at"`
}

type SkillLevel struct {
	CEFR         string    `json:"cefr"`
	Score        float64   `json:"score"`
	LastAssessed time.Time `json:"last_assessed"`
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
