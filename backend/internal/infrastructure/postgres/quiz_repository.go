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
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/quizzes"
)

// QuizRepository implements quizzes.Repository.
type QuizRepository struct {
	db *sqlx.DB
}

func NewQuizRepository(db *sqlx.DB) quizzes.Repository {
	return &QuizRepository{db: db}
}

func (r *QuizRepository) CreateQuiz(ctx context.Context, q *domain.Quiz) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO quizzes (id, group_id, teacher_id, title, description, time_limit, max_attempts, open_at, close_at, is_published, created_at, updated_at)
		VALUES (:id, :group_id, :teacher_id, :title, :description, :time_limit, :max_attempts, :open_at, :close_at, :is_published, :created_at, :updated_at)`, q)
	if err != nil {
		return fmt.Errorf("QuizRepository.CreateQuiz: %w", err)
	}
	return nil
}

func (r *QuizRepository) GetQuizByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	var q domain.Quiz
	err := r.db.GetContext(ctx, &q, `SELECT * FROM quizzes WHERE id=$1 AND deleted_at IS NULL`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("QuizRepository.GetQuizByID: %w", err)
	}
	return &q, nil
}

func (r *QuizRepository) ListGroupQuizzes(ctx context.Context, groupID uuid.UUID) ([]*domain.Quiz, error) {
	var qs []*domain.Quiz
	err := r.db.SelectContext(ctx, &qs,
		`SELECT * FROM quizzes WHERE group_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC`, groupID)
	if err != nil {
		return nil, fmt.Errorf("QuizRepository.ListGroupQuizzes: %w", err)
	}
	return qs, nil
}

func (r *QuizRepository) UpdateQuiz(ctx context.Context, q *domain.Quiz) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE quizzes
		SET title=:title, description=:description, time_limit=:time_limit,
		    max_attempts=:max_attempts, open_at=:open_at, close_at=:close_at, updated_at=:updated_at
		WHERE id=:id AND deleted_at IS NULL`, q)
	if err != nil {
		return fmt.Errorf("QuizRepository.UpdateQuiz: %w", err)
	}
	return nil
}

func (r *QuizRepository) SoftDeleteQuiz(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE quizzes SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("QuizRepository.SoftDeleteQuiz: %w", err)
	}
	return nil
}

func (r *QuizRepository) PublishQuiz(ctx context.Context, id uuid.UUID, published bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE quizzes SET is_published=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
		published, id)
	if err != nil {
		return fmt.Errorf("QuizRepository.PublishQuiz: %w", err)
	}
	return nil
}

func (r *QuizRepository) CreateQuestion(ctx context.Context, q *domain.Question) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO questions (id, quiz_id, body, explanation, position, points)
		VALUES (:id, :quiz_id, :body, :explanation, :position, :points)`, q)
	if err != nil {
		return fmt.Errorf("QuizRepository.CreateQuestion: %w", err)
	}
	return nil
}

func (r *QuizRepository) GetQuestion(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	var q domain.Question
	err := r.db.GetContext(ctx, &q, `SELECT * FROM questions WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("QuizRepository.GetQuestion: %w", err)
	}
	return &q, nil
}

func (r *QuizRepository) ListQuestions(ctx context.Context, quizID uuid.UUID) ([]*domain.Question, error) {
	var qs []*domain.Question
	err := r.db.SelectContext(ctx, &qs,
		`SELECT * FROM questions WHERE quiz_id=$1 ORDER BY position ASC`, quizID)
	if err != nil {
		return nil, fmt.Errorf("QuizRepository.ListQuestions: %w", err)
	}
	return qs, nil
}

func (r *QuizRepository) UpdateQuestion(ctx context.Context, q *domain.Question) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE questions SET body=:body, explanation=:explanation, position=:position, points=:points
		WHERE id=:id`, q)
	if err != nil {
		return fmt.Errorf("QuizRepository.UpdateQuestion: %w", err)
	}
	return nil
}

func (r *QuizRepository) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM questions WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("QuizRepository.DeleteQuestion: %w", err)
	}
	return nil
}

func (r *QuizRepository) CreateOption(ctx context.Context, o *domain.Option) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO options (id, question_id, body, is_correct, position)
		VALUES (:id, :question_id, :body, :is_correct, :position)`, o)
	if err != nil {
		return fmt.Errorf("QuizRepository.CreateOption: %w", err)
	}
	return nil
}

func (r *QuizRepository) ListOptions(ctx context.Context, questionID uuid.UUID) ([]*domain.Option, error) {
	var os []*domain.Option
	err := r.db.SelectContext(ctx, &os,
		`SELECT * FROM options WHERE question_id=$1 ORDER BY position ASC`, questionID)
	if err != nil {
		return nil, fmt.Errorf("QuizRepository.ListOptions: %w", err)
	}
	return os, nil
}

func (r *QuizRepository) DeleteOption(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM options WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("QuizRepository.DeleteOption: %w", err)
	}
	return nil
}

func (r *QuizRepository) CreateAttempt(ctx context.Context, a *domain.QuizAttempt) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO quiz_attempts (id, quiz_id, student_id, started_at, max_score)
		VALUES (:id, :quiz_id, :student_id, :started_at, :max_score)`, a)
	if err != nil {
		return fmt.Errorf("QuizRepository.CreateAttempt: %w", err)
	}
	return nil
}

func (r *QuizRepository) GetAttempt(ctx context.Context, id uuid.UUID) (*domain.QuizAttempt, error) {
	var a domain.QuizAttempt
	err := r.db.GetContext(ctx, &a, `SELECT * FROM quiz_attempts WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("QuizRepository.GetAttempt: %w", err)
	}
	return &a, nil
}

func (r *QuizRepository) CountAttempts(ctx context.Context, quizID, studentID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM quiz_attempts WHERE quiz_id=$1 AND student_id=$2`, quizID, studentID)
	if err != nil {
		return 0, fmt.Errorf("QuizRepository.CountAttempts: %w", err)
	}
	return count, nil
}

func (r *QuizRepository) FinishAttempt(ctx context.Context, id uuid.UUID, score, maxScore int16) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE quiz_attempts SET finished_at=$1, score=$2, max_score=$3 WHERE id=$4`,
		now, score, maxScore, id)
	if err != nil {
		return fmt.Errorf("QuizRepository.FinishAttempt: %w", err)
	}
	return nil
}

func (r *QuizRepository) SaveAnswers(ctx context.Context, answers []*domain.StudentAnswer) error {
	for _, a := range answers {
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO student_answers (attempt_id, question_id, option_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (attempt_id, question_id) DO UPDATE SET option_id=EXCLUDED.option_id`,
			a.AttemptID, a.QuestionID, a.OptionID)
		if err != nil {
			return fmt.Errorf("QuizRepository.SaveAnswers: %w", err)
		}
	}
	return nil
}

func (r *QuizRepository) GetAnswers(ctx context.Context, attemptID uuid.UUID) ([]*domain.StudentAnswer, error) {
	var ans []*domain.StudentAnswer
	err := r.db.SelectContext(ctx, &ans,
		`SELECT * FROM student_answers WHERE attempt_id=$1`, attemptID)
	if err != nil {
		return nil, fmt.Errorf("QuizRepository.GetAnswers: %w", err)
	}
	return ans, nil
}
