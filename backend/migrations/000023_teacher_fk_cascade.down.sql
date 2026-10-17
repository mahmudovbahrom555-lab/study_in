ALTER TABLE grades DROP CONSTRAINT grades_teacher_id_fkey,
    ADD CONSTRAINT grades_teacher_id_fkey
        FOREIGN KEY (teacher_id) REFERENCES users(id);

ALTER TABLE attendance DROP CONSTRAINT attendance_teacher_id_fkey,
    ADD CONSTRAINT attendance_teacher_id_fkey
        FOREIGN KEY (teacher_id) REFERENCES users(id);
