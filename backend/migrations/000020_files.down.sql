CREATE TABLE submission_attachments (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    object_key    TEXT NOT NULL,
    filename      TEXT NOT NULL,
    mime_type     TEXT NOT NULL DEFAULT 'application/octet-stream',
    size_bytes    BIGINT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sub_attach_submission ON submission_attachments(submission_id);

INSERT INTO submission_attachments (id, submission_id, object_key, filename, mime_type, size_bytes, created_at)
SELECT f.id, sf.submission_id, f.object_key, f.original_name, f.mime_type, f.size_bytes, f.created_at
FROM submission_files sf
JOIN files f ON f.id = sf.file_id;

DROP TABLE submission_files;
DROP TABLE files;
