-- Этап 5: тесты (quiz), вопросы, варианты ответов, результаты.

CREATE TABLE quizzes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    teacher_id  UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    title       VARCHAR(300) NOT NULL,
    description TEXT,
    time_limit  INT,                        -- секунды; NULL = без лимита
    max_attempts SMALLINT NOT NULL DEFAULT 1,
    open_at     TIMESTAMPTZ,               -- когда открывается студентам
    close_at    TIMESTAMPTZ,               -- когда закрывается
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_quizzes_group ON quizzes(group_id, created_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE questions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id     UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    body        TEXT NOT NULL,
    explanation TEXT,                      -- объяснение правильного ответа
    position    SMALLINT NOT NULL DEFAULT 0,
    points      SMALLINT NOT NULL DEFAULT 1
);

CREATE INDEX idx_questions_quiz ON questions(quiz_id, position);

CREATE TABLE options (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    body        TEXT NOT NULL,
    is_correct  BOOLEAN NOT NULL DEFAULT FALSE,
    position    SMALLINT NOT NULL DEFAULT 0
);

CREATE INDEX idx_options_question ON options(question_id, position);

-- Попытка студента пройти тест.
CREATE TABLE quiz_attempts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id     UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    student_id  UUID NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    score       SMALLINT,                  -- NULL пока не завершена
    max_score   SMALLINT NOT NULL DEFAULT 0
);

CREATE INDEX idx_quiz_attempts_student ON quiz_attempts(quiz_id, student_id);

-- Ответы студента на вопросы (после завершения попытки).
CREATE TABLE student_answers (
    attempt_id  UUID NOT NULL REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id)     ON DELETE CASCADE,
    option_id   UUID NOT NULL REFERENCES options(id)       ON DELETE CASCADE,
    PRIMARY KEY (attempt_id, question_id)
);
