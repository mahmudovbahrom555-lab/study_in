-- Вступление в группу — по заявке, репетитор одобряет (решение владельцев,
-- 2026-11-26): и по коду, и по ссылке. Одна строка на ученика и группу;
-- после отказа ученик может подать заявку снова.
CREATE TABLE group_join_requests (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id   UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    status     VARCHAR(10) NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending', 'accepted', 'declined')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    decided_at TIMESTAMPTZ,
    UNIQUE (group_id, student_id)
);

CREATE INDEX idx_join_requests_pending ON group_join_requests(group_id) WHERE status = 'pending';
CREATE INDEX idx_join_requests_student ON group_join_requests(student_id) WHERE status = 'pending';
