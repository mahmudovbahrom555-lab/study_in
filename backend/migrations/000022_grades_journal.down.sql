DELETE FROM grades WHERE kind = 'homework';
DROP INDEX IF EXISTS uq_grades_assignment_student;
ALTER TABLE grades DROP COLUMN assignment_id, DROP COLUMN kind;
