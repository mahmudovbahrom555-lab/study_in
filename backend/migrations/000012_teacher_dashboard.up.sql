-- Teacher Acceptance Rate: teachers mark AI-generated questions as accepted/rejected.
-- This is the key quality metric for quiz generation.
CREATE TABLE quiz_question_feedback (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    teacher_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    accepted    BOOLEAN NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (question_id, teacher_id)
);
CREATE INDEX idx_question_feedback_teacher ON quiz_question_feedback(teacher_id);
CREATE INDEX idx_question_feedback_question ON quiz_question_feedback(question_id);
