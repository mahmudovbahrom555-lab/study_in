-- ai_recommendations: persisted Rule Engine output
-- Each pending recommendation is replaced on the next insights refresh.
-- Once a teacher acts (accepted/dismissed), the row is preserved for outcome tracking.
CREATE TABLE ai_recommendations (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id      UUID        REFERENCES groups(id) ON DELETE SET NULL,
    priority      INT         NOT NULL DEFAULT 2 CHECK (priority IN (1, 2, 3)),
    action        VARCHAR(50) NOT NULL,
    topic         VARCHAR(200),
    reason        TEXT        NOT NULL,
    student_count INT         NOT NULL DEFAULT 0,
    rule_key      VARCHAR(100) NOT NULL,
    rule_data     JSONB       NOT NULL DEFAULT '{}',
    status        VARCHAR(20) NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending','accepted','dismissed','snoozed')),
    teacher_action TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    acted_at      TIMESTAMPTZ,
    expires_at    TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '7 days'
);

CREATE INDEX idx_ai_rec_teacher_group ON ai_recommendations(teacher_id, group_id, created_at DESC);
CREATE INDEX idx_ai_rec_pending       ON ai_recommendations(group_id) WHERE status = 'pending';

-- recommendation_outcomes: measures impact 7 days after a recommendation was shown.
-- Stores delta of avg mastery/confidence/consistency for the affected topic.
CREATE TABLE recommendation_outcomes (
    id                UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
    recommendation_id UUID      NOT NULL REFERENCES ai_recommendations(id) ON DELETE CASCADE,
    measured_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    mastery_delta     NUMERIC(5,3),
    confidence_delta  NUMERIC(5,3),
    consistency_delta NUMERIC(5,3),
    students_improved INT       NOT NULL DEFAULT 0,
    students_total    INT       NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX idx_rec_outcomes_unique ON recommendation_outcomes(recommendation_id);
