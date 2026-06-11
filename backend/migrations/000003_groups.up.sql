-- Этап 2: группы, участники, статус оплаты.

CREATE TABLE groups (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name           VARCHAR(200) NOT NULL,
    subject        VARCHAR(100),
    description    TEXT,
    invite_code    VARCHAR(10) UNIQUE NOT NULL,
    is_archived    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_groups_teacher ON groups(teacher_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_groups_invite  ON groups(invite_code) WHERE deleted_at IS NULL;

-- payment_status заложен сразу (killer feature для репетиторов, см. DECISIONS.md).
CREATE TABLE group_members (
    group_id       UUID NOT NULL REFERENCES groups(id)  ON DELETE CASCADE,
    student_id     UUID NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    joined_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payment_status VARCHAR(20)  NOT NULL DEFAULT 'trial'
                       CHECK (payment_status IN ('paid', 'pending', 'trial')),
    PRIMARY KEY (group_id, student_id)
);

CREATE INDEX idx_group_members_student ON group_members(student_id);
