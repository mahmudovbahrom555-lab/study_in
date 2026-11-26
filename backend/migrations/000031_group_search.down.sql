DROP INDEX IF EXISTS idx_users_username;
ALTER TABLE users DROP COLUMN IF EXISTS username;
DROP INDEX IF EXISTS idx_groups_public;
ALTER TABLE groups DROP COLUMN IF EXISTS is_public;
