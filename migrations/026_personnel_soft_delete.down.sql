DROP INDEX IF EXISTS idx_instructors_user_id_active;
DROP INDEX IF EXISTS idx_students_user_id_active;
DROP INDEX IF EXISTS idx_students_student_no_active;
DROP INDEX IF EXISTS idx_users_username_active;

ALTER TABLE instructors DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE students DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);
ALTER TABLE students ADD CONSTRAINT students_student_no_key UNIQUE (student_no);
ALTER TABLE students ADD CONSTRAINT students_user_id_key UNIQUE (user_id);
ALTER TABLE instructors ADD CONSTRAINT instructors_user_id_key UNIQUE (user_id);
