-- Единый модуль файлов (G10, DECISIONS.md 2026-10-07).
-- Все загрузки хранятся в одной таблице, разделы ссылаются на file_id —
-- так будущий AI-обработчик получает любой файл одинаково, по ID.

CREATE TABLE files (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purpose       VARCHAR(32) NOT NULL CHECK (purpose IN ('submission')),
    object_key    TEXT        NOT NULL UNIQUE,
    original_name TEXT        NOT NULL,
    mime_type     TEXT        NOT NULL,   -- определён по содержимому, не по имени
    size_bytes    BIGINT      NOT NULL,
    sha256        CHAR(64),               -- NULL только у перенесённых старых файлов
    width         INT,                    -- только для изображений
    height        INT,
    -- pending: загружен, но ещё не прикреплён (удаляется фоновой очисткой через сутки)
    -- attached: прикреплён к сущности (сдаче ДЗ)
    status        VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'attached')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_files_owner   ON files(owner_id, created_at DESC);
CREATE INDEX idx_files_pending ON files(created_at) WHERE status = 'pending';

CREATE TABLE submission_files (
    submission_id UUID     NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    file_id       UUID     NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    position      SMALLINT NOT NULL,
    PRIMARY KEY (submission_id, file_id)
);

CREATE INDEX idx_submission_files_file ON submission_files(file_id);

-- Перенос существующих вложений сдач.
INSERT INTO files (id, owner_id, purpose, object_key, original_name, mime_type, size_bytes, status, created_at)
SELECT sa.id, s.student_id, 'submission', sa.object_key, sa.filename, sa.mime_type, sa.size_bytes, 'attached', sa.created_at
FROM submission_attachments sa
JOIN submissions s ON s.id = sa.submission_id;

INSERT INTO submission_files (submission_id, file_id, position)
SELECT submission_id, id,
       (ROW_NUMBER() OVER (PARTITION BY submission_id ORDER BY created_at))::SMALLINT
FROM submission_attachments;

DROP TABLE submission_attachments;
