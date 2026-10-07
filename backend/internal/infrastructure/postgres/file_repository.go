package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// fileColumns — явный список вместо SELECT *: новая колонка в files не сломает
// маппинг (docs/backend-architecture-methodology.md, P1).
const fileColumns = `id, owner_id, purpose, object_key, original_name, mime_type,
	size_bytes, sha256, width, height, status, created_at`

type FileRepository struct {
	db *sqlx.DB
}

func NewFileRepository(db *sqlx.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, f *domain.File) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO files (`+fileColumns+`)
		VALUES (:id, :owner_id, :purpose, :object_key, :original_name, :mime_type,
		        :size_bytes, :sha256, :width, :height, :status, :created_at)`, f)
	if err != nil {
		return fmt.Errorf("FileRepository.Create: %w", err)
	}
	return nil
}

// GetByIDs возвращает найденные файлы; отсутствующие ID просто не попадают в ответ.
func (r *FileRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domain.File, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In(`SELECT `+fileColumns+` FROM files WHERE id IN (?)`, ids)
	if err != nil {
		return nil, fmt.Errorf("FileRepository.GetByIDs build: %w", err)
	}
	var out []*domain.File
	if err := r.db.SelectContext(ctx, &out, r.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("FileRepository.GetByIDs: %w", err)
	}
	return out, nil
}

func (r *FileRepository) ListStalePending(ctx context.Context, before time.Time, limit int) ([]*domain.File, error) {
	var out []*domain.File
	err := r.db.SelectContext(ctx, &out, `
		SELECT `+fileColumns+` FROM files
		WHERE status = 'pending' AND created_at < $1
		ORDER BY created_at
		LIMIT $2`, before, limit)
	if err != nil {
		return nil, fmt.Errorf("FileRepository.ListStalePending: %w", err)
	}
	return out, nil
}

func (r *FileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM files WHERE id = $1`, id); err != nil {
		return fmt.Errorf("FileRepository.Delete: %w", err)
	}
	return nil
}
