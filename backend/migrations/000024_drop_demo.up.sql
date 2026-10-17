-- Демо убрано (DECISIONS.md, 2026-10-15/17): демо-группы и флаг больше не нужны.
DELETE FROM groups WHERE is_demo;
DROP INDEX IF EXISTS idx_groups_teacher_demo;
ALTER TABLE groups DROP COLUMN is_demo;
