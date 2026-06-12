-- One demo group per teacher (non-deleted). Prevents duplicate demo groups
-- when a teacher calls POST /ai/me/demo more than once.
CREATE UNIQUE INDEX idx_groups_teacher_demo
    ON groups(teacher_id)
    WHERE is_demo = TRUE AND deleted_at IS NULL;
