package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/parents"
)

type ParentRepository struct {
	db *sqlx.DB
}

func NewParentRepository(db *sqlx.DB) parents.Repository {
	return &ParentRepository{db: db}
}

func (r *ParentRepository) Link(ctx context.Context, l *domain.ParentLink) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO parent_links (id, parent_id, student_id, created_at)
		VALUES (:id, :parent_id, :student_id, :created_at)`, l)
	if err != nil {
		return fmt.Errorf("ParentRepository.Link: %w", err)
	}
	return nil
}

func (r *ParentRepository) Unlink(ctx context.Context, parentID, studentID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM parent_links WHERE parent_id=$1 AND student_id=$2`, parentID, studentID)
	if err != nil {
		return fmt.Errorf("ParentRepository.Unlink: %w", err)
	}
	return nil
}

func (r *ParentRepository) GetLink(ctx context.Context, parentID, studentID uuid.UUID) (*domain.ParentLink, error) {
	var l domain.ParentLink
	err := r.db.GetContext(ctx, &l,
		`SELECT * FROM parent_links WHERE parent_id=$1 AND student_id=$2`, parentID, studentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ParentRepository.GetLink: %w", err)
	}
	return &l, nil
}

func (r *ParentRepository) ListChildren(ctx context.Context, parentID uuid.UUID) ([]*domain.ParentLink, error) {
	var links []*domain.ParentLink
	err := r.db.SelectContext(ctx, &links,
		`SELECT * FROM parent_links WHERE parent_id=$1 ORDER BY created_at DESC`, parentID)
	if err != nil {
		return nil, fmt.Errorf("ParentRepository.ListChildren: %w", err)
	}
	return links, nil
}

func (r *ParentRepository) ListParents(ctx context.Context, studentID uuid.UUID) ([]*domain.ParentLink, error) {
	var links []*domain.ParentLink
	err := r.db.SelectContext(ctx, &links,
		`SELECT * FROM parent_links WHERE student_id=$1 ORDER BY created_at DESC`, studentID)
	if err != nil {
		return nil, fmt.Errorf("ParentRepository.ListParents: %w", err)
	}
	return links, nil
}
