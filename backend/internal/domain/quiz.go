package domain

import (
	"time"

	"github.com/google/uuid"
)

type Quiz struct {
	ID          uuid.UUID  `db:"id"`
	GroupID     uuid.UUID  `db:"group_id"`
	TeacherID   uuid.UUID  `db:"teacher_id"`
	Title       string     `db:"title"`
	Description *string    `db:"description"`
	TimeLimit   *int       `db:"time_limit"`
	MaxAttempts int16      `db:"max_attempts"`
	OpenAt      *time.Time `db:"open_at"`
	CloseAt     *time.Time `db:"close_at"`
	IsPublished bool       `db:"is_published"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

type Question struct {
	ID          uuid.UUID `db:"id"`
	QuizID      uuid.UUID `db:"quiz_id"`
	Body        string    `db:"body"`
	Explanation *string   `db:"explanation"`
	Position    int16     `db:"position"`
	Points      int16     `db:"points"`
}

type Option struct {
	ID         uuid.UUID `db:"id"`
	QuestionID uuid.UUID `db:"question_id"`
	Body       string    `db:"body"`
	IsCorrect  bool      `db:"is_correct"`
	Position   int16     `db:"position"`
}

// OptionForStudent — вариант без IsCorrect (показывается студентам во время теста).
type OptionForStudent struct {
	ID       uuid.UUID `db:"id"`
	Body     string    `db:"body"`
	Position int16     `db:"position"`
}

type QuizAttempt struct {
	ID         uuid.UUID  `db:"id"`
	QuizID     uuid.UUID  `db:"quiz_id"`
	StudentID  uuid.UUID  `db:"student_id"`
	StartedAt  time.Time  `db:"started_at"`
	FinishedAt *time.Time `db:"finished_at"`
	Score      *int16     `db:"score"`
	MaxScore   int16      `db:"max_score"`
}

type StudentAnswer struct {
	AttemptID  uuid.UUID `db:"attempt_id"`
	QuestionID uuid.UUID `db:"question_id"`
	OptionID   uuid.UUID `db:"option_id"`
}
