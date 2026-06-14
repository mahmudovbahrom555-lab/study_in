package quizzes

import (
	"context"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// Repository — хранилище тестов, вопросов и попыток.
type Repository interface {
	// Quizzes
	CreateQuiz(ctx context.Context, q *domain.Quiz) error
	GetQuizByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error)
	ListGroupQuizzes(ctx context.Context, groupID uuid.UUID) ([]*domain.Quiz, error)
	UpdateQuiz(ctx context.Context, q *domain.Quiz) error
	SoftDeleteQuiz(ctx context.Context, id uuid.UUID) error
	PublishQuiz(ctx context.Context, id uuid.UUID, published bool) error

	// Questions
	CreateQuestion(ctx context.Context, q *domain.Question) error
	GetQuestion(ctx context.Context, id uuid.UUID) (*domain.Question, error)
	ListQuestions(ctx context.Context, quizID uuid.UUID) ([]*domain.Question, error)
	UpdateQuestion(ctx context.Context, q *domain.Question) error
	DeleteQuestion(ctx context.Context, id uuid.UUID) error

	// Options
	CreateOption(ctx context.Context, o *domain.Option) error
	ListOptions(ctx context.Context, questionID uuid.UUID) ([]*domain.Option, error)
	DeleteOption(ctx context.Context, id uuid.UUID) error

	// Attempts
	CreateAttempt(ctx context.Context, a *domain.QuizAttempt) error
	GetAttempt(ctx context.Context, id uuid.UUID) (*domain.QuizAttempt, error)
	CountAttempts(ctx context.Context, quizID, studentID uuid.UUID) (int, error)
	FinishAttempt(ctx context.Context, id uuid.UUID, score, maxScore int16) error

	// Answers
	SaveAnswers(ctx context.Context, answers []*domain.StudentAnswer) error
	GetAnswers(ctx context.Context, attemptID uuid.UUID) ([]*domain.StudentAnswer, error)
}

// GroupChecker — минимальный интерфейс для проверки доступа к группе.
type GroupChecker interface {
	GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.Group, error)
	GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.GroupMember, error)
}

// AnswerObserver receives quiz answer events so the AI layer can update
// spaced-repetition state (SM-2) and award XP without coupling the two packages.
// Called fire-and-forget after each submission — errors are logged, not returned.
type AnswerObserver interface {
	OnAnswer(ctx context.Context, studentID, questionID uuid.UUID, correct bool)
}
