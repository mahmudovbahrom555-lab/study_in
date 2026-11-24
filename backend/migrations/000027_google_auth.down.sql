DROP INDEX IF EXISTS idx_users_google_sub;

-- Без телефона пользователя не вернуть под NOT NULL: такие аккаунты удаляются.
DELETE FROM users WHERE phone IS NULL;
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;

ALTER TABLE users DROP COLUMN google_sub;
ALTER TABLE users DROP COLUMN email;
