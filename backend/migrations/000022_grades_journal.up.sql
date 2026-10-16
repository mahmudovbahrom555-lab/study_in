-- Единый журнал оценок (DECISIONS.md, 2026-10-16): оценки за ДЗ попадают
-- в grades. kind различает классную работу и ДЗ — по нему журнал можно
-- позже разделить на вкладки без переноса данных.
ALTER TABLE grades
    ADD COLUMN kind VARCHAR(20) NOT NULL DEFAULT 'class'
        CHECK (kind IN ('class', 'homework')),
    ADD COLUMN assignment_id UUID REFERENCES assignments(id) ON DELETE CASCADE;

-- Одна оценка за задание на ученика: повторная проверка обновляет её.
CREATE UNIQUE INDEX uq_grades_assignment_student
    ON grades(assignment_id, student_id) WHERE assignment_id IS NOT NULL;

-- Уже поставленные оценки за ДЗ — в журнал.
INSERT INTO grades (group_id, student_id, teacher_id, subject, value, max_value,
                    comment, graded_at, created_at, kind, assignment_id)
SELECT a.group_id, s.student_id, a.teacher_id, LEFT(a.title, 100), s.grade, 100,
       NULLIF(s.teacher_note, ''), COALESCE(s.graded_at, NOW()), COALESCE(s.graded_at, NOW()),
       'homework', a.id
FROM submissions s
JOIN assignments a ON a.id = s.assignment_id
WHERE s.grade IS NOT NULL;
