DROP TABLE IF EXISTS chat_reads;
DROP TABLE IF EXISTS group_bans;
ALTER TABLE users DROP COLUMN IF EXISTS chat_banned;
ALTER TABLE group_members DROP COLUMN IF EXISTS chat_muted;
ALTER TABLE groups DROP COLUMN IF EXISTS chat_mode, DROP COLUMN IF EXISTS chat_ttl_seconds;
DROP TABLE IF EXISTS chat_messages;
