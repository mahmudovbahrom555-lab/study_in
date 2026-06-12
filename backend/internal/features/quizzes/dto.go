package quizzes

import (
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// --- Requests ---

type CreateQuizRequest struct {
	Title       string     `json:"title"        validate:"required,min=2,max=300"`
	Description *string    `json:"description"`
	TimeLimit   *int       `json:"time_limit"`
	MaxAttempts int16      `json:"max_attempts" validate:"min=1,max=10"`
	OpenAt      *time.Time `json:"open_at"`
	CloseAt     *time.Time `json:"close_at"`
}

type UpdateQuizRequest struct {
	Title       *string    `json:"title"        validate:"omitempty,min=2,max=300"`
	Description *string    `json:"description"`
	TimeLimit   *int       `json:"time_limit"`
	MaxAttempts *int16     `json:"max_attempts" validate:"omitempty,min=1,max=10"`
	OpenAt      *time.Time `json:"open_at"`
	CloseAt     *time.Time `json:"close_at"`
}

type CreateQuestionRequest struct {
	Body        string  `json:"body"     validate:"required,min=1,max=2000"`
	Explanation *string `json:"explanation"`
	Position    int16   `json:"position"`
	Points      int16   `json:"points"   validate:"min=1,max=100"`
}

type CreateOptionRequest struct {
	Body      string `json:"body"       validate:"required,min=1,max=500"`
	IsCorrect bool   `json:"is_correct"`
	Position  int16  `json:"position"`
}

// SubmitAnswersRequest maps question_id → option_id.
type SubmitAnswersRequest struct {
	Answers map[string]string `json:"answers" validate:"required"` // question_id → option_id (string UUIDs)
}

// QuizResultItem carries per-question feedback returned after submission.
type QuizResultItem struct {
	QuestionID       uuid.UUID  `json:"question_id"`
	QuestionBody     string     `json:"question_body"`
	Points           int16      `json:"points"`
	Explanation      *string    `json:"explanation,omitempty"`
	SelectedOptionID *uuid.UUID `json:"selected_option_id,omitempty"`
	SelectedBody     *string    `json:"selected_body,omitempty"`
	CorrectOptionID  uuid.UUID  `json:"correct_option_id"`
	CorrectBody      string     `json:"correct_body"`
	IsCorrect        bool       `json:"is_correct"`
}

// SubmitAttemptFullResponse wraps the attempt result + per-question feedback.
type SubmitAttemptFullResponse struct {
	Attempt         AttemptResponse  `json:"attempt"`
	QuestionResults []QuizResultItem `json:"question_results"`
}

// SubmitAttemptResult is returned by the service after scoring.
type SubmitAttemptResult struct {
	Attempt         *domain.QuizAttempt
	QuestionResults []QuizResultItem
}

// --- Responses ---

type OptionResponse struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	Position  int16     `json:"position"`
	IsCorrect *bool     `json:"is_correct,omitempty"` // nil for students during active attempt
}

type QuestionResponse struct {
	ID          uuid.UUID        `json:"id"`
	Body        string           `json:"body"`
	Explanation *string          `json:"explanation,omitempty"`
	Position    int16            `json:"position"`
	Points      int16            `json:"points"`
	Options     []OptionResponse `json:"options"`
}

type QuizResponse struct {
	ID          uuid.UUID          `json:"id"`
	GroupID     uuid.UUID          `json:"group_id"`
	TeacherID   uuid.UUID          `json:"teacher_id"`
	Title       string             `json:"title"`
	Description *string            `json:"description"`
	TimeLimit   *int               `json:"time_limit"`
	MaxAttempts int16              `json:"max_attempts"`
	OpenAt      *time.Time         `json:"open_at"`
	CloseAt     *time.Time         `json:"close_at"`
	IsPublished bool               `json:"is_published"`
	Questions   []QuestionResponse `json:"questions,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
}

type AttemptResponse struct {
	ID         uuid.UUID  `json:"id"`
	QuizID     uuid.UUID  `json:"quiz_id"`
	StudentID  uuid.UUID  `json:"student_id"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	Score      *int16     `json:"score"`
	MaxScore   int16      `json:"max_score"`
}

func quizToResponse(q *domain.Quiz) QuizResponse {
	return QuizResponse{
		ID:          q.ID,
		GroupID:     q.GroupID,
		TeacherID:   q.TeacherID,
		Title:       q.Title,
		Description: q.Description,
		TimeLimit:   q.TimeLimit,
		MaxAttempts: q.MaxAttempts,
		OpenAt:      q.OpenAt,
		CloseAt:     q.CloseAt,
		IsPublished: q.IsPublished,
		CreatedAt:   q.CreatedAt,
	}
}

func attemptToResponse(a *domain.QuizAttempt) AttemptResponse {
	return AttemptResponse{
		ID:         a.ID,
		QuizID:     a.QuizID,
		StudentID:  a.StudentID,
		StartedAt:  a.StartedAt,
		FinishedAt: a.FinishedAt,
		Score:      a.Score,
		MaxScore:   a.MaxScore,
	}
}
