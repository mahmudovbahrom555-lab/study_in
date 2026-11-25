-- Реакции учеников на объявления ленты: «Понятно» и «Есть вопрос».
-- Без свободного текста — ученики (в том числе дети) не пишут, только отмечают.
CREATE TABLE post_reactions (
    post_id    UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind       VARCHAR(20) NOT NULL CHECK (kind IN ('understood', 'question')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);
