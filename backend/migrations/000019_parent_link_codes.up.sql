-- Короткие коды, которыми ученик подтверждает привязку родителя.
-- Ученик выдаёт код → родитель вводит его → создаётся parent_links.
-- Заменяет ввод UUID ученика: тот был неудобен и не требовал согласия ученика.
CREATE TABLE parent_link_codes (
    code       VARCHAR(8)  PRIMARY KEY,
    student_id UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_parent_link_codes_student ON parent_link_codes(student_id);
