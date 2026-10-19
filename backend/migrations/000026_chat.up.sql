-- Чат в классе (docs/design/CHAT.md, DECISIONS.md 2026-10-19).

-- Сообщения со сроком жизни: удаляются физически фоновой задачей.
CREATE TABLE chat_messages (
    id          UUID PRIMARY KEY,
    group_id    UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    author_id   UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    body        TEXT NOT NULL CHECK (char_length(body) BETWEEN 1 AND 2000),
    client_id   UUID NOT NULL,            -- id от приложения: повтор отправки не создаёт дубль
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,
    UNIQUE (author_id, client_id)
);
CREATE INDEX idx_chat_group_time ON chat_messages(group_id, created_at DESC);
CREATE INDEX idx_chat_expires    ON chat_messages(expires_at);

-- Настройки чата класса: срок хранения и кто пишет.
ALTER TABLE groups
    ADD COLUMN chat_ttl_seconds INT NOT NULL DEFAULT 86400
        CHECK (chat_ttl_seconds IN (3600, 21600, 43200, 86400)),
    ADD COLUMN chat_mode VARCHAR(16) NOT NULL DEFAULT 'open'
        CHECK (chat_mode IN ('open', 'owner_only'));

-- Запрет писать: в классе (ставит репетитор) и во всех чатах (по жалобе).
ALTER TABLE group_members ADD COLUMN chat_muted BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN chat_banned BOOLEAN NOT NULL DEFAULT FALSE;

-- Заблокированные в классе не могут вступить снова.
CREATE TABLE group_bans (
    group_id   UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

-- До какого момента человек прочитал чат класса — для счётчика непрочитанных.
CREATE TABLE chat_reads (
    group_id     UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    last_read_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (group_id, user_id)
);
