-- AI layer: documents, chunks (pgvector), sessions, messages,
-- topic mastery (SM-2), gamification (XP/streak), skill assessments, token billing.

CREATE EXTENSION IF NOT EXISTS vector;

-- ── Documents ────────────────────────────────────────────────────────────────

CREATE TABLE ai_documents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID REFERENCES groups(id) ON DELETE CASCADE,
    teacher_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(300) NOT NULL,
    object_key  TEXT NOT NULL,
    mime_type   VARCHAR(100) NOT NULL,
    size_bytes  BIGINT NOT NULL,
    status      VARCHAR(30) NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending','processing','ready','failed')),
    chunk_count INT NOT NULL DEFAULT 0,
    error_msg   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ai_documents_group   ON ai_documents(group_id);
CREATE INDEX idx_ai_documents_teacher ON ai_documents(teacher_id);

-- ── Chunks + embeddings (pgvector) ───────────────────────────────────────────

CREATE TABLE ai_document_chunks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES ai_documents(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    content     TEXT NOT NULL,
    embedding   vector(1536),
    token_count INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_chunks_document ON ai_document_chunks(document_id);
-- HNSW for fast cosine similarity search
CREATE INDEX idx_chunks_embedding ON ai_document_chunks
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- ── Async generation jobs ─────────────────────────────────────────────────────

CREATE TABLE ai_jobs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type        VARCHAR(50) NOT NULL,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ref_id      UUID,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending','processing','done','failed')),
    result      JSONB,
    error_msg   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ai_jobs_user ON ai_jobs(user_id, created_at DESC);

-- ── Conversation sessions ─────────────────────────────────────────────────────

CREATE TABLE ai_sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_type VARCHAR(30) NOT NULL
                 CHECK (session_type IN ('mentor','writing','speaking','reading','quiz_gen')),
    document_id  UUID REFERENCES ai_documents(id) ON DELETE SET NULL,
    subject      VARCHAR(100),
    cefr_level   VARCHAR(5),
    metadata     JSONB NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ai_sessions_user ON ai_sessions(user_id, created_at DESC);

CREATE TABLE ai_messages (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES ai_sessions(id) ON DELETE CASCADE,
    role       VARCHAR(15) NOT NULL CHECK (role IN ('user','assistant','system')),
    content    TEXT NOT NULL,
    tokens_in  INT NOT NULL DEFAULT 0,
    tokens_out INT NOT NULL DEFAULT 0,
    model      VARCHAR(50) NOT NULL DEFAULT 'gpt-4o-mini',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ai_messages_session ON ai_messages(session_id, created_at);

-- ── Topic tags (for weakness tracking) ───────────────────────────────────────

CREATE TABLE topic_tags (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name    VARCHAR(200) NOT NULL,
    subject VARCHAR(100),
    UNIQUE (name, subject)
);

-- Link quiz questions → topic tags (populated during AI quiz generation)
CREATE TABLE question_topics (
    question_id UUID REFERENCES questions(id) ON DELETE CASCADE,
    topic_id    UUID REFERENCES topic_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (question_id, topic_id)
);

-- ── Spaced repetition (SM-2) ──────────────────────────────────────────────────

CREATE TABLE topic_mastery (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id      UUID NOT NULL REFERENCES topic_tags(id) ON DELETE CASCADE,
    correct_count INT NOT NULL DEFAULT 0,
    total_count   INT NOT NULL DEFAULT 0,
    next_review   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    interval_days INT NOT NULL DEFAULT 1,
    ease_factor   NUMERIC(4,2) NOT NULL DEFAULT 2.5,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (student_id, topic_id)
);
CREATE INDEX idx_mastery_student ON topic_mastery(student_id);
CREATE INDEX idx_mastery_review  ON topic_mastery(student_id, next_review);

-- ── Gamification (XP + streak) ────────────────────────────────────────────────

CREATE TABLE student_gamification (
    student_id         UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    xp_total           INT NOT NULL DEFAULT 0,
    xp_today           INT NOT NULL DEFAULT 0,
    daily_goal_xp      INT NOT NULL DEFAULT 50,
    streak_days        INT NOT NULL DEFAULT 0,
    longest_streak     INT NOT NULL DEFAULT 0,
    last_activity_date DATE,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE xp_events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount     INT NOT NULL,
    reason     VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_xp_events_student ON xp_events(student_id, created_at DESC);

-- ── Skill assessments (reading/writing/speaking/listening) ───────────────────

CREATE TABLE skill_assessments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    skill      VARCHAR(20) NOT NULL
               CHECK (skill IN ('reading','writing','speaking','listening')),
    cefr_level VARCHAR(5),
    score      NUMERIC(5,2),
    feedback   TEXT,
    audio_key  TEXT,
    transcript TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_assessments_student ON skill_assessments(student_id, skill, created_at DESC);

-- ── Token usage / billing ─────────────────────────────────────────────────────

CREATE TABLE ai_token_usage (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    model      VARCHAR(50) NOT NULL,
    tokens_in  INT NOT NULL,
    tokens_out INT NOT NULL,
    feature    VARCHAR(50) NOT NULL,
    cost_usd   NUMERIC(10,6) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_token_usage_user_month ON ai_token_usage(user_id, created_at);
