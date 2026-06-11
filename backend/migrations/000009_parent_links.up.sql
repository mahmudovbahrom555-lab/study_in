-- parent_links: connects a parent user to a student user
CREATE TABLE parent_links (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (parent_id, student_id)
);

CREATE INDEX idx_parent_links_parent  ON parent_links(parent_id);
CREATE INDEX idx_parent_links_student ON parent_links(student_id);
