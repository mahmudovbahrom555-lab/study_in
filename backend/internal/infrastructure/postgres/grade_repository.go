package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/grades"
)

type GradeRepository struct {
	db *sqlx.DB
}

func NewGradeRepository(db *sqlx.DB) grades.Repository {
	return &GradeRepository{db: db}
}

func (r *GradeRepository) Create(ctx context.Context, g *domain.Grade) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO grades (id, group_id, student_id, teacher_id, subject, value, max_value, comment, graded_at, created_at)
		VALUES (:id, :group_id, :student_id, :teacher_id, :subject, :value, :max_value, :comment, :graded_at, :created_at)`, g)
	if err != nil {
		return fmt.Errorf("GradeRepository.Create: %w", err)
	}
	return nil
}

func (r *GradeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Grade, error) {
	var g domain.Grade
	err := r.db.GetContext(ctx, &g, `SELECT * FROM grades WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GradeRepository.GetByID: %w", err)
	}
	return &g, nil
}

func (r *GradeRepository) ListByGroupStudent(ctx context.Context, groupID, studentID uuid.UUID) ([]*domain.Grade, error) {
	var gs []*domain.Grade
	err := r.db.SelectContext(ctx, &gs,
		`SELECT * FROM grades WHERE group_id=$1 AND student_id=$2 ORDER BY graded_at DESC`,
		groupID, studentID)
	if err != nil {
		return nil, fmt.Errorf("GradeRepository.ListByGroupStudent: %w", err)
	}
	return gs, nil
}

func (r *GradeRepository) ListByGroup(ctx context.Context, groupID uuid.UUID) ([]*domain.Grade, error) {
	var gs []*domain.Grade
	err := r.db.SelectContext(ctx, &gs,
		`SELECT * FROM grades WHERE group_id=$1 ORDER BY graded_at DESC`, groupID)
	if err != nil {
		return nil, fmt.Errorf("GradeRepository.ListByGroup: %w", err)
	}
	return gs, nil
}

func (r *GradeRepository) Update(ctx context.Context, g *domain.Grade) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE grades
		SET subject=:subject, value=:value, max_value=:max_value, comment=:comment, graded_at=:graded_at
		WHERE id=:id`, g)
	if err != nil {
		return fmt.Errorf("GradeRepository.Update: %w", err)
	}
	return nil
}

func (r *GradeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM grades WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("GradeRepository.Delete: %w", err)
	}
	return nil
}
