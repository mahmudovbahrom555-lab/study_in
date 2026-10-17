package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// AccountRepository implements account.Repository.
type AccountRepository struct{ db *sqlx.DB }

func NewAccountRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// teacherSubmissionFiles — файлы учеников в группах репетитора $1. Их строки
// files принадлежат ученикам и каскадом не удалятся, а без группы они никому
// не нужны — удаляем вместе с группой.
const teacherSubmissionFiles = `
	SELECT sf.file_id
	FROM submission_files sf
	JOIN submissions s ON s.id = sf.submission_id
	JOIN assignments a ON a.id = s.assignment_id
	JOIN groups g ON g.id = a.group_id
	WHERE g.teacher_id = $1`

// DeleteUser удаляет пользователя одной транзакцией и возвращает ключи
// объектов хранилища, оставшихся без владельца.
//
// Всё остальное удаляют каскады БД (DECISIONS.md, 2026-10-17): у репетитора —
// группы со всем содержимым, у ученика — участие, сдачи, оценки, посещаемость,
// у всех — токены, уведомления, связи родитель–ребёнок, согласия.
func (r *AccountRepository) DeleteUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("AccountRepository.DeleteUser begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // после Commit — no-op

	var keys []string
	if err := tx.SelectContext(ctx, &keys, `
		SELECT object_key FROM files WHERE owner_id = $1
		UNION
		SELECT object_key FROM files WHERE id IN (`+teacherSubmissionFiles+`)
		UNION
		SELECT aa.object_key FROM assignment_attachments aa
		JOIN assignments a ON a.id = aa.assignment_id
		JOIN groups g ON g.id = a.group_id
		WHERE g.teacher_id = $1
		UNION
		SELECT pa.object_key FROM post_attachments pa
		JOIN posts p ON p.id = pa.post_id
		JOIN groups g ON g.id = p.group_id
		WHERE p.author_id = $1 OR g.teacher_id = $1
		UNION
		SELECT object_key FROM ai_documents WHERE teacher_id = $1`, userID); err != nil {
		return nil, fmt.Errorf("AccountRepository.DeleteUser keys: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM files WHERE id IN (`+teacherSubmissionFiles+`)`, userID); err != nil {
		return nil, fmt.Errorf("AccountRepository.DeleteUser student files: %w", err)
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("AccountRepository.DeleteUser: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, domain.ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("AccountRepository.DeleteUser commit: %w", err)
	}
	return keys, nil
}
