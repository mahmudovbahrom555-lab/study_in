package ai

import (
	"context"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// DocumentRepository manages AI documents and their chunks.
type DocumentRepository interface {
	CreateDocument(ctx context.Context, doc *domain.AIDocument) error
	GetDocument(ctx context.Context, id uuid.UUID) (*domain.AIDocument, error)
	ListDocuments(ctx context.Context, teacherID uuid.UUID, groupID *uuid.UUID) ([]*domain.AIDocument, error)
	UpdateDocumentStatus(ctx context.Context, id uuid.UUID, status domain.AIDocumentStatus, chunkCount int, errMsg *string) error
	DeleteDocument(ctx context.Context, id uuid.UUID) error

	// Chunks
	SaveChunk(ctx context.Context, chunk *domain.AIDocumentChunk, embedding []float32) error
	SearchSimilarChunks(ctx context.Context, documentID uuid.UUID, queryEmbedding []float32, limit int) ([]*domain.AIDocumentChunk, error)
}

// JobRepository tracks async generation jobs.
type JobRepository interface {
	CreateJob(ctx context.Context, job *domain.AIJob) error
	GetJob(ctx context.Context, id uuid.UUID) (*domain.AIJob, error)
	UpdateJob(ctx context.Context, id uuid.UUID, status domain.AIJobStatus, result []byte, errMsg *string) error
}

// SessionRepository stores AI mentor/writing/speaking sessions.
type SessionRepository interface {
	CreateSession(ctx context.Context, s *domain.AISession) error
	GetSession(ctx context.Context, id uuid.UUID) (*domain.AISession, error)
	ListUserSessions(ctx context.Context, userID uuid.UUID) ([]*domain.AISession, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error

	AddMessage(ctx context.Context, m *domain.AIMessage) error
	ListMessages(ctx context.Context, sessionID uuid.UUID) ([]*domain.AIMessage, error)
}

// TokenRepository logs token usage for billing and limit enforcement.
type TokenRepository interface {
	LogUsage(ctx context.Context, u *domain.AITokenUsage) error
	MonthlyTokens(ctx context.Context, userID uuid.UUID) (int, error)
}

// MasteryRepository manages spaced-repetition state per student/topic.
type MasteryRepository interface {
	UpsertMastery(ctx context.Context, m *domain.TopicMastery) error
	GetMastery(ctx context.Context, studentID, topicID uuid.UUID) (*domain.TopicMastery, error)
	ListDueMastery(ctx context.Context, studentID uuid.UUID, limit int) ([]*domain.TopicMastery, error)
	ListWeakTopics(ctx context.Context, studentID uuid.UUID, limit int) ([]*domain.TopicMastery, error)
}

// GamificationRepository manages XP and streak data.
type GamificationRepository interface {
	GetOrCreate(ctx context.Context, studentID uuid.UUID) (*domain.StudentGamification, error)
	AddXP(ctx context.Context, studentID uuid.UUID, amount int, reason string) error
	UpdateStreak(ctx context.Context, studentID uuid.UUID) error
}

// TopicRepository manages topic tag lookups/creation.
type TopicRepository interface {
	UpsertTopic(ctx context.Context, name string, subject *string) (*domain.TopicTag, error)
	GetTopic(ctx context.Context, id uuid.UUID) (*domain.TopicTag, error)
	LinkQuestionTopics(ctx context.Context, questionID uuid.UUID, topicIDs []uuid.UUID) error
}

// QuizCreator is the minimal interface the AI service needs to persist
// generated quizzes into the existing quizzes system.
type QuizCreator interface {
	CreateQuizWithQuestions(ctx context.Context, req QuizCreateRequest) (uuid.UUID, error)
}

// QuizCreateRequest carries the AI-generated quiz structure to persist.
type QuizCreateRequest struct {
	GroupID   uuid.UUID
	TeacherID uuid.UUID
	Title     string
	Questions []GeneratedQuestion
}

type GeneratedQuestion struct {
	Text        string
	Explanation string
	Difficulty  string
	TopicTags   []string
	Options     []GeneratedOption
}

type GeneratedOption struct {
	Text      string
	IsCorrect bool
}

// ObjectStore is the MinIO interface needed for document downloads.
type ObjectStore interface {
	GetObject(ctx context.Context, key string) ([]byte, error)
}

// InsightsRepository provides analytics queries for the Teacher Dashboard.
type InsightsRepository interface {
	// ClassInsights aggregates topic weaknesses, student summaries, and quiz stats for a group.
	ClassInsights(ctx context.Context, groupID uuid.UUID) (*domain.ClassInsights, error)
	// StudentProgress returns detailed per-student analytics (mastery, quiz history, skills).
	StudentProgress(ctx context.Context, groupID, studentID uuid.UUID) (*domain.StudentProgress, error)
}

// QuizFeedbackRepository stores Teacher Acceptance Rate data.
type QuizFeedbackRepository interface {
	UpsertFeedback(ctx context.Context, fb *domain.QuizQuestionFeedback) error
	AcceptanceRate(ctx context.Context, teacherID uuid.UUID) (float64, int, error) // rate, total
}
