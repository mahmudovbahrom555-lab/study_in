-- Вход через Google: у такого пользователя может не быть телефона.

ALTER TABLE users ALTER COLUMN phone DROP NOT NULL;

ALTER TABLE users ADD COLUMN email      VARCHAR(255);
ALTER TABLE users ADD COLUMN google_sub VARCHAR(255);

-- sub — постоянный идентификатор Google-аккаунта; email может смениться.
CREATE UNIQUE INDEX idx_users_google_sub ON users(google_sub) WHERE google_sub IS NOT NULL;
