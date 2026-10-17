-- Удаление аккаунта (DECISIONS.md, 2026-10-17): все ссылки на users каскадные.
-- grades.teacher_id и attendance.teacher_id были NO ACTION — их проверка
-- срабатывала раньше каскада через группы, и аккаунт репетитора не удалялся.
ALTER TABLE grades DROP CONSTRAINT grades_teacher_id_fkey,
    ADD CONSTRAINT grades_teacher_id_fkey
        FOREIGN KEY (teacher_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE attendance DROP CONSTRAINT attendance_teacher_id_fkey,
    ADD CONSTRAINT attendance_teacher_id_fkey
        FOREIGN KEY (teacher_id) REFERENCES users(id) ON DELETE CASCADE;
