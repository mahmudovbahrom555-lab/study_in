-- Ребёнок младше 13 (по нейтральному возрастному экрану в приложении):
-- его заявку в группу сначала разрешает привязанный родитель. Храним только
-- признак, без года рождения.
ALTER TABLE users ADD COLUMN is_child BOOLEAN NOT NULL DEFAULT FALSE;

-- parent_pending — заявка ребёнка ждёт родителя; репетитор её ещё не видит.
ALTER TABLE group_join_requests DROP CONSTRAINT group_join_requests_status_check;
ALTER TABLE group_join_requests ALTER COLUMN status TYPE VARCHAR(16);
ALTER TABLE group_join_requests ADD CONSTRAINT group_join_requests_status_check
    CHECK (status IN ('parent_pending', 'pending', 'accepted', 'declined'));
-- Кто из родителей разрешил — репетитор видит это в заявке.
ALTER TABLE group_join_requests
    ADD COLUMN approved_by_parent UUID REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX idx_join_requests_parent_pending ON group_join_requests(student_id)
    WHERE status = 'parent_pending';
