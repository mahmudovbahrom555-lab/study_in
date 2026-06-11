-- Этап 4: домашние задания и загрузка файлов через MinIO.

CREATE TABLE assignments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    teacher_id  UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    title       VARCHAR(300) NOT NULL,
    description TEXT,
    due_date    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_assignments_group ON assignments(group_id, created_at DESC) WHERE deleted_at IS NULL;

-- Файловые вложения к заданию.
CREATE TABLE assignment_attachments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id UUID NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    object_key  TEXT NOT NULL,
    filename    TEXT NOT NULL,
    mime_type   TEXT NOT NULL DEFAULT 'application/octet-stream',
    size_bytes  BIGINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assign_attach_assignment ON assignment_attachments(assignment_id);

-- Сдача задания студентом.
CREATE TABLE submissions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id UUID NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    student_id    UUID NOT NULL REFERENCES users(id)       ON DELETE CASCADE,
    comment       TEXT,
    grade         SMALLINT CHECK (grade IS NULL OR (grade >= 0 AND grade <= 100)),
    teacher_note  TEXT,
    submitted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    graded_at     TIMESTAMPTZ,
    UNIQUE (assignment_id, student_id)
);

CREATE INDEX idx_submissions_assignment ON submissions(assignment_id);
CREATE INDEX idx_submissions_student    ON submissions(student_id);

-- Файловые вложения к сдаче.
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
