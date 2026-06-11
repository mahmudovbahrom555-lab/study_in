-- Grades journal: teacher assigns numeric grades to students per group
CREATE TABLE grades (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id     UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    student_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    teacher_id   UUID NOT NULL REFERENCES users(id),
    subject      VARCHAR(100) NOT NULL DEFAULT '',
    value        NUMERIC(5,2) NOT NULL CHECK (value >= 0),
    max_value    NUMERIC(5,2) NOT NULL DEFAULT 100 CHECK (max_value > 0),
    comment      TEXT,
    graded_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_grades_group_student ON grades(group_id, student_id);
CREATE INDEX idx_grades_group_at      ON grades(group_id, graded_at DESC);
