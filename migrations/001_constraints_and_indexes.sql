-- Migration 001: 用户/角色唯一约束与核心索引（必须调整）
-- 依赖：表 users, students, instructors 已存在
-- 执行顺序：在建表之后首次运行

-- ========== 1. users 唯一约束 ==========
ALTER TABLE users
  ADD CONSTRAINT users_username_unique UNIQUE (username);

ALTER TABLE users
  ADD CONSTRAINT users_email_unique UNIQUE (email);

-- phone 可选唯一（若业务允许多用户共填同一电话可省略）
-- ALTER TABLE users ADD CONSTRAINT users_phone_unique UNIQUE (phone);

-- ========== 2. students / instructors 与 user 一对一 ==========
ALTER TABLE students
  ADD CONSTRAINT students_user_id_unique UNIQUE (user_id);

ALTER TABLE instructors
  ADD CONSTRAINT instructors_user_id_unique UNIQUE (user_id);

-- ========== 3. 核心查询索引 ==========
-- users：按登录与角色
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);
CREATE INDEX IF NOT EXISTS idx_users_username_lower ON users (LOWER(username));

-- students：按责任教员、学员编号
CREATE INDEX IF NOT EXISTS idx_students_responsible_instructor ON students (responsible_instructor_id);
CREATE INDEX IF NOT EXISTS idx_students_student_no ON students (student_no);

-- instructors：按 user 查
CREATE INDEX IF NOT EXISTS idx_instructors_user_id ON instructors (user_id);

-- comments：按目标与用户
CREATE INDEX IF NOT EXISTS idx_comments_target ON comments (target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments (user_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments (parent_id);

-- tasks：按学员、指派人、状态
CREATE INDEX IF NOT EXISTS idx_tasks_student_status ON tasks (student_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_assigner_status ON tasks (assigner_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_type_ref ON tasks (type, ref_id);
