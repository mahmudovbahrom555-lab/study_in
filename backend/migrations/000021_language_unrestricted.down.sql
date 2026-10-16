ALTER TABLE users ADD CONSTRAINT users_language_check
    CHECK (language IN ('uz', 'uz-cyrl', 'ru', 'en'));
