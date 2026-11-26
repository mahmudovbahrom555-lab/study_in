-- Открытые группы видны в поиске; закрытые (по умолчанию) — только по ссылке или коду.
-- Вступление в любую — через заявку, которую принимает репетитор (000030).
ALTER TABLE groups ADD COLUMN is_public BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX idx_groups_public ON groups(created_at DESC)
    WHERE is_public AND NOT is_archived AND deleted_at IS NULL;

-- @username — только у репетиторов: по нему ученик находит открытые группы.
-- Учеников в поиске нет.
ALTER TABLE users ADD COLUMN username VARCHAR(32);
CREATE UNIQUE INDEX idx_users_username ON users(username) WHERE username IS NOT NULL;
