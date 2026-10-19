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
)

// ChatRepository implements chat.Repository.
type ChatRepository struct{ db *sqlx.DB }

func NewChatRepository(db *sqlx.DB) *ChatRepository { return &ChatRepository{db: db} }

const chatMessageSelect = `
	SELECT m.id, m.group_id, m.author_id, u.name AS author_name, m.body, m.client_id,
	       m.created_at, m.expires_at
	FROM chat_messages m
	JOIN users u ON u.id = m.author_id`

func (r *ChatRepository) Access(ctx context.Context, groupID, userID uuid.UUID) (*domain.ChatAccess, error) {
	var a domain.ChatAccess
	err := r.db.GetContext(ctx, &a, `
		SELECT g.teacher_id, g.is_archived, g.chat_ttl_seconds, g.chat_mode,
		       (m.student_id IS NOT NULL)       AS is_member,
		       COALESCE(m.chat_muted, FALSE)     AS chat_muted,
		       u.chat_banned, u.is_active
		FROM groups g
		JOIN users u ON u.id = $2
		LEFT JOIN group_members m ON m.group_id = g.id AND m.student_id = $2
		WHERE g.id = $1 AND g.deleted_at IS NULL`, groupID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ChatRepository.Access: %w", err)
	}
	return &a, nil
}

func (r *ChatRepository) Insert(ctx context.Context, m *domain.ChatMessage) (*domain.ChatMessage, error) {
	if _, err := r.db.NamedExecContext(ctx, `
		INSERT INTO chat_messages (id, group_id, author_id, body, client_id, created_at, expires_at)
		VALUES (:id, :group_id, :author_id, :body, :client_id, :created_at, :expires_at)
		ON CONFLICT (author_id, client_id) DO NOTHING`, m); err != nil {
		return nil, fmt.Errorf("ChatRepository.Insert: %w", err)
	}
	// Повтор отправки вернёт уже сохранённое сообщение — дубля не будет.
	var saved domain.ChatMessage
	if err := r.db.GetContext(ctx, &saved,
		chatMessageSelect+` WHERE m.author_id = $1 AND m.client_id = $2`, m.AuthorID, m.ClientID); err != nil {
		return nil, fmt.Errorf("ChatRepository.Insert read back: %w", err)
	}
	return &saved, nil
}

func (r *ChatRepository) List(ctx context.Context, groupID uuid.UUID, before *uuid.UUID, limit int, now time.Time) ([]*domain.ChatMessage, error) {
	var ms []*domain.ChatMessage
	err := r.db.SelectContext(ctx, &ms, chatMessageSelect+`
		WHERE m.group_id = $1 AND m.expires_at > $2
		  AND ($3::uuid IS NULL OR (m.created_at, m.id) <
		       (SELECT created_at, id FROM chat_messages WHERE id = $3))
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT $4`, groupID, now, before, limit)
	if err != nil {
		return nil, fmt.Errorf("ChatRepository.List: %w", err)
	}
	return ms, nil
}

func (r *ChatRepository) Get(ctx context.Context, id uuid.UUID) (*domain.ChatMessage, error) {
	var m domain.ChatMessage
	err := r.db.GetContext(ctx, &m, chatMessageSelect+` WHERE m.id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ChatRepository.Get: %w", err)
	}
	return &m, nil
}

func (r *ChatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM chat_messages WHERE id = $1`, id); err != nil {
		return fmt.Errorf("ChatRepository.Delete: %w", err)
	}
	return nil
}

func (r *ChatRepository) DeleteExpired(ctx context.Context, now time.Time, batch int) (int, error) {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM chat_messages
		WHERE id IN (SELECT id FROM chat_messages WHERE expires_at <= $1 LIMIT $2)`, now, batch)
	if err != nil {
		return 0, fmt.Errorf("ChatRepository.DeleteExpired: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (r *ChatRepository) MarkRead(ctx context.Context, groupID, userID uuid.UUID, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO chat_reads (group_id, user_id, last_read_at) VALUES ($1, $2, $3)
		ON CONFLICT (group_id, user_id)
		DO UPDATE SET last_read_at = GREATEST(chat_reads.last_read_at, EXCLUDED.last_read_at)`,
		groupID, userID, at)
	if err != nil {
		return fmt.Errorf("ChatRepository.MarkRead: %w", err)
	}
	return nil
}

func (r *ChatRepository) UnreadCounts(ctx context.Context, userID uuid.UUID, groupIDs []uuid.UUID, now time.Time) (map[uuid.UUID]int, error) {
	q, args, err := sqlx.In(`
		SELECT m.group_id, COUNT(*) AS n
		FROM chat_messages m
		LEFT JOIN chat_reads r ON r.group_id = m.group_id AND r.user_id = ?
		WHERE m.group_id IN (?) AND m.expires_at > ? AND m.author_id <> ?
		  AND (r.last_read_at IS NULL OR m.created_at > r.last_read_at)
		GROUP BY m.group_id`, userID, groupIDs, now, userID)
	if err != nil {
		return nil, fmt.Errorf("ChatRepository.UnreadCounts build: %w", err)
	}
	var rows []struct {
		GroupID uuid.UUID `db:"group_id"`
		N       int       `db:"n"`
	}
	if err := r.db.SelectContext(ctx, &rows, r.db.Rebind(q), args...); err != nil {
		return nil, fmt.Errorf("ChatRepository.UnreadCounts: %w", err)
	}
	out := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		out[row.GroupID] = row.N
	}
	return out, nil
}

func (r *ChatRepository) UpdateSettings(ctx context.Context, groupID uuid.UUID, ttlSeconds int, mode domain.ChatMode) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE groups SET chat_ttl_seconds = $2, chat_mode = $3, updated_at = NOW() WHERE id = $1`,
		groupID, ttlSeconds, mode)
	if err != nil {
		return fmt.Errorf("ChatRepository.UpdateSettings: %w", err)
	}
	return nil
}

func (r *ChatRepository) SetMuted(ctx context.Context, groupID, userID uuid.UUID, muted bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE group_members SET chat_muted = $3 WHERE group_id = $1 AND student_id = $2`,
		groupID, userID, muted)
	if err != nil {
		return fmt.Errorf("ChatRepository.SetMuted: %w", err)
	}
	return nil
}

func (r *ChatRepository) Ban(ctx context.Context, groupID, userID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ChatRepository.Ban begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // после Commit — no-op

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM group_members WHERE group_id = $1 AND student_id = $2`, groupID, userID); err != nil {
		return fmt.Errorf("ChatRepository.Ban remove: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO group_bans (group_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		groupID, userID); err != nil {
		return fmt.Errorf("ChatRepository.Ban insert: %w", err)
	}
	return tx.Commit()
}
