-- Этап 1 rollback: удаление схемы авторизации.

DROP TABLE IF EXISTS device_tokens;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS verification_codes;
DROP TABLE IF EXISTS users;
