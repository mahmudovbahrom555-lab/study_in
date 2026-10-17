ALTER TABLE groups ADD COLUMN IF NOT EXISTS is_demo BOOLEAN NOT NULL DEFAULT FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS idx_groups_teacher_demo
    ON groups(teacher_id) WHERE is_demo = TRUE AND deleted_at IS NULL;
