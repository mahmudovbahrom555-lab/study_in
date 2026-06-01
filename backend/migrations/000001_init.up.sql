-- Этап 0: базовая инициализация схемы.
-- Расширение для генерации UUID.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Эта миграция — скелет. Полная схема будет добавлена на Этапе 1.
-- Здесь только метатаблица для отслеживания развития схемы.

CREATE TABLE IF NOT EXISTS schema_info (
    key VARCHAR(50) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO schema_info (key, value) VALUES ('initialized_at', NOW()::TEXT)
ON CONFLICT (key) DO NOTHING;
