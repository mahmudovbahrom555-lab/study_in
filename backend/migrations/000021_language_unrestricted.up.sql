-- Список языков живёт в коде (domain.SupportedLanguages): новый язык
-- добавляется без миграции.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_language_check;
