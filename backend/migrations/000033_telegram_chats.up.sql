-- Вход через Telegram-бота: человек делится в боте своим номером (Telegram
-- подтверждает, что номер его), и бот может присылать ему коды входа.
-- Код всё равно вводится в приложении — как SMS: так чужая ссылка на бота
-- не даёт войти в чужой аккаунт.
CREATE TABLE telegram_chats (
    phone       VARCHAR(20) PRIMARY KEY,
    chat_id     BIGINT      NOT NULL,
    tg_user_id  BIGINT      NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
