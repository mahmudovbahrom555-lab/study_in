DROP INDEX IF EXISTS idx_join_requests_parent_pending;
ALTER TABLE group_join_requests DROP COLUMN IF EXISTS approved_by_parent;
DELETE FROM group_join_requests WHERE status = 'parent_pending';
ALTER TABLE group_join_requests DROP CONSTRAINT group_join_requests_status_check;
ALTER TABLE group_join_requests ALTER COLUMN status TYPE VARCHAR(10);
ALTER TABLE group_join_requests ADD CONSTRAINT group_join_requests_status_check
    CHECK (status IN ('pending', 'accepted', 'declined'));
ALTER TABLE users DROP COLUMN IF EXISTS is_child;
