-- Attendance: teacher marks presence/absence per student per lesson date
CREATE TYPE attendance_status AS ENUM ('present', 'absent', 'late', 'excused');

CREATE TABLE attendance (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    student_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    teacher_id  UUID NOT NULL REFERENCES users(id),
    lesson_date DATE NOT NULL,
    status      attendance_status NOT NULL DEFAULT 'present',
    note        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, student_id, lesson_date)
);

CREATE INDEX idx_attendance_group_date    ON attendance(group_id, lesson_date DESC);
CREATE INDEX idx_attendance_group_student ON attendance(group_id, student_id);
