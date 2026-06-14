-- Parent ROI reports: monthly snapshots per student per group.
CREATE TABLE parent_reports (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id                UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    period_start            DATE NOT NULL,
    period_end              DATE NOT NULL,
    attendance_pct          NUMERIC(5,2) NOT NULL DEFAULT 0,
    quiz_score_avg          NUMERIC(5,2) NOT NULL DEFAULT 0,
    quiz_score_prev_avg     NUMERIC(5,2) NOT NULL DEFAULT 0,
    mastery_avg             NUMERIC(5,3) NOT NULL DEFAULT 0,
    mastery_prev_avg        NUMERIC(5,3) NOT NULL DEFAULT 0,
    homework_completion_pct NUMERIC(5,2) NOT NULL DEFAULT 0,
    quiz_attempts_count     INT NOT NULL DEFAULT 0,
    ai_summary              TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (student_id, group_id, period_start)
);
CREATE INDEX idx_parent_reports_student ON parent_reports(student_id, created_at DESC);
CREATE INDEX idx_parent_reports_group   ON parent_reports(group_id, period_start DESC);

-- Owner risk alerts: churn-risk signals per student detected weekly.
CREATE TABLE owner_risk_alerts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id       UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    risk_level     VARCHAR(10) NOT NULL CHECK (risk_level IN ('high', 'medium', 'low')),
    trigger_reason TEXT NOT NULL,
    recommendation TEXT NOT NULL,
    status         VARCHAR(20) NOT NULL DEFAULT 'new'
                   CHECK (status IN ('new', 'in_progress', 'resolved')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_risk_alerts_teacher ON owner_risk_alerts(teacher_id, status, created_at DESC);

-- Skill logs: granular per-answer trace for Academic Twin foundation.
CREATE TABLE skill_logs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    skill_tag        VARCHAR(100) NOT NULL,
    source           VARCHAR(30) NOT NULL CHECK (source IN ('quiz', 'assignment', 'manual')),
    correct          BOOLEAN NOT NULL,
    mastery_estimate NUMERIC(4,3) NOT NULL DEFAULT 0.500,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_skill_logs_student ON skill_logs(student_id, skill_tag, created_at DESC);
