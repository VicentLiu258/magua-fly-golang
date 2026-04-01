-- 人员（用户 / 学员 / 教员）软删除：deleted_at；唯一约束改为「未删除」时生效，便于学号/用户名复用

ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE students ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE instructors ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

COMMENT ON COLUMN users.deleted_at IS '软删除时间，NULL 表示有效账号';
COMMENT ON COLUMN students.deleted_at IS '软删除时间，NULL 表示有效学员';
COMMENT ON COLUMN instructors.deleted_at IS '软删除时间，NULL 表示有效教员';

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_active ON users (username) WHERE deleted_at IS NULL;

ALTER TABLE students DROP CONSTRAINT IF EXISTS students_student_no_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_students_student_no_active ON students (student_no) WHERE deleted_at IS NULL;

ALTER TABLE students DROP CONSTRAINT IF EXISTS students_user_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_students_user_id_active ON students (user_id) WHERE deleted_at IS NULL AND user_id IS NOT NULL;

ALTER TABLE instructors DROP CONSTRAINT IF EXISTS instructors_user_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_instructors_user_id_active ON instructors (user_id) WHERE deleted_at IS NULL;
