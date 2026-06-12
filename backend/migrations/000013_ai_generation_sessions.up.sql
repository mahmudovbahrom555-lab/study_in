-- AI Generation Sessions: tracks every quiz-generation event.
-- Core dataset for TAR analysis, model improvement, and content quality trends.
CREATE TABLE ai_generation_sessions (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_document_id UUID        REFERENCES ai_documents(id) ON DELETE SET NULL,
    group_id           UUID        REFERENCES groups(id) ON DELETE SET NULL,
    quiz_id            UUID        REFERENCES quizzes(id) ON DELETE SET NULL,
    generated_count    INT         NOT NULL DEFAULT 0,
    accepted_count     INT         NOT NULL DEFAULT 0,  -- updated lazily from feedback
    edited_count       INT         NOT NULL DEFAULT 0,  -- future: track question edits
    rejected_count     INT         NOT NULL DEFAULT 0,  -- updated lazily from feedback
    cefr_level         VARCHAR(10),
    subject            VARCHAR(100),
    model_used         VARCHAR(50),                     -- which GPT model generated this
    prompt_tokens      INT         NOT NULL DEFAULT 0,
    completion_tokens  INT         NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at       TIMESTAMPTZ
);
CREATE INDEX idx_ai_gen_sessions_teacher   ON ai_generation_sessions(teacher_id);
CREATE INDEX idx_ai_gen_sessions_group     ON ai_generation_sessions(group_id)   WHERE group_id IS NOT NULL;
CREATE INDEX idx_ai_gen_sessions_quiz      ON ai_generation_sessions(quiz_id)    WHERE quiz_id IS NOT NULL;
CREATE INDEX idx_ai_gen_sessions_created   ON ai_generation_sessions(created_at  DESC);
