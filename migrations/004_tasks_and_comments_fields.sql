-- Migration 004: 任务与评论扩展及索引（建议调整）
-- 依赖：tasks, comments 已存在

-- ========== 1. tasks：优先级与完成时间 ==========
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'tasks' AND column_name = 'priority'
  ) THEN
    ALTER TABLE tasks ADD COLUMN priority VARCHAR(20) DEFAULT 'medium';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'tasks' AND column_name = 'completed_at'
  ) THEN
    ALTER TABLE tasks ADD COLUMN completed_at TIMESTAMPTZ;
  END IF;
END $$;

COMMENT ON COLUMN tasks.priority IS 'low | medium | high，与前端 InstructorDashboard 待办优先级对应';

-- 可选：标签（若需可启用）
-- ALTER TABLE tasks ADD COLUMN IF NOT EXISTS tags TEXT[];

-- ========== 2. comments：软删除与可见性 ==========
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'comments' AND column_name = 'deleted_at'
  ) THEN
    ALTER TABLE comments ADD COLUMN deleted_at TIMESTAMPTZ;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'comments' AND column_name = 'visibility'
  ) THEN
    ALTER TABLE comments ADD COLUMN visibility VARCHAR(20) DEFAULT 'public';
  END IF;
END $$;

COMMENT ON COLUMN comments.deleted_at IS '软删除时间，NULL 表示未删除';
COMMENT ON COLUMN comments.visibility IS 'public | student_only | instructor_only | internal';

-- 评论父级外键（若建表时未定义）
-- ALTER TABLE comments
--   ADD CONSTRAINT comments_parent_id_fkey
--   FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE;

-- 索引：按目标查询时排除已删除
CREATE INDEX IF NOT EXISTS idx_comments_target_not_deleted
  ON comments (target_type, target_id) WHERE deleted_at IS NULL;
