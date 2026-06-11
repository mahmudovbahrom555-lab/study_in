-- Этап 3: лента объявлений.
-- Учитель публикует посты в группу; студенты/родители читают.

CREATE TABLE posts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    author_id   UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    body        TEXT NOT NULL,
    pinned      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_posts_group ON posts(group_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_posts_pinned ON posts(group_id, pinned) WHERE deleted_at IS NULL AND pinned = TRUE;

-- Файловые вложения к посту (MinIO object keys).
CREATE TABLE post_attachments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    object_key  TEXT NOT NULL,
    filename    TEXT NOT NULL,
    mime_type   TEXT NOT NULL DEFAULT 'application/octet-stream',
    size_bytes  BIGINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_post_attachments_post ON post_attachments(post_id);
