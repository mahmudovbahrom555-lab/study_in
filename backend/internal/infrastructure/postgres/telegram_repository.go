package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type TelegramRepository struct{ db *sqlx.DB }

func NewTelegramRepository(db *sqlx.DB) *TelegramRepository { return &TelegramRepository{db: db} }

// SaveChat: номер, переподтверждённый с другого аккаунта Telegram,
// переходит к нему — коды получает тот, у кого номер сейчас.
func (r *TelegramRepository) SaveChat(ctx context.Context, phone string, chatID, tgUserID int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO telegram_chats (phone, chat_id, tg_user_id) VALUES ($1, $2, $3)
		ON CONFLICT (phone) DO UPDATE
		    SET chat_id = EXCLUDED.chat_id, tg_user_id = EXCLUDED.tg_user_id, updated_at = NOW()`,
		phone, chatID, tgUserID)
	if err != nil {
		return fmt.Errorf("TelegramRepository.SaveChat: %w", err)
	}
	return nil
}

func (r *TelegramRepository) ChatByPhone(ctx context.Context, phone string) (int64, bool, error) {
	var chatID int64
	err := r.db.GetContext(ctx, &chatID, `SELECT chat_id FROM telegram_chats WHERE phone = $1`, phone)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("TelegramRepository.ChatByPhone: %w", err)
	}
	return chatID, true, nil
}
