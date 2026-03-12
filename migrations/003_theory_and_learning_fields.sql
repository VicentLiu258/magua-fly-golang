-- Migration 003: 理论文档与学习单元字段扩展（建议调整）
-- 依赖：theory_documents, theory_questions, theory_attempts, subject_learning, subject_questions, subject_attempts 已存在

-- ========== 1. theory_documents：富文本块存储（与前端 DocBlock[] 对齐）==========
ALTER TABLE theory_documents
  ADD COLUMN IF NOT EXISTS content_blocks JSONB DEFAULT '[]';

COMMENT ON COLUMN theory_documents.content_blocks IS '富文本块数组，与前端 DocBlock[] 结构一致；content_ref 可保留为外部存储引用';

-- ========== 2. theory_questions：多选与扩展 ==========
ALTER TABLE theory_questions
  ADD COLUMN IF NOT EXISTS is_multi_select BOOLEAN DEFAULT FALSE;

ALTER TABLE theory_questions
  ADD COLUMN IF NOT EXISTS correct_indexes INT[];

COMMENT ON COLUMN theory_questions.correct_indexes IS '多选题正确选项下标数组；单选可继续用 correct_index';

ALTER TABLE theory_questions
  ADD COLUMN IF NOT EXISTS difficulty SMALLINT;

ALTER TABLE theory_questions
  ADD COLUMN IF NOT EXISTS tags TEXT[];

-- ========== 3. theory_attempts：统计与通过状态 ==========
ALTER TABLE theory_attempts
  ADD COLUMN IF NOT EXISTS duration_seconds INT;

ALTER TABLE theory_attempts
  ADD COLUMN IF NOT EXISTS max_score INT;

ALTER TABLE theory_attempts
  ADD COLUMN IF NOT EXISTS correct_count INT;

ALTER TABLE theory_attempts
  ADD COLUMN IF NOT EXISTS total_count INT;

ALTER TABLE theory_attempts
  ADD COLUMN IF NOT EXISTS passed BOOLEAN;

CREATE INDEX IF NOT EXISTS idx_theory_attempts_user_doc_submitted
  ON theory_attempts (user_id, theory_document_id, submitted_at);

-- ========== 4. subject_learning：保留 UNIQUE(syllabus_subject_id)，可选 sort_order ==========
ALTER TABLE subject_learning
  ADD COLUMN IF NOT EXISTS sort_order INT DEFAULT 0;

-- ========== 5. subject_questions：与 theory_questions 对齐的扩展（可选）==========
ALTER TABLE subject_questions
  ADD COLUMN IF NOT EXISTS is_multi_select BOOLEAN DEFAULT FALSE;

ALTER TABLE subject_questions
  ADD COLUMN IF NOT EXISTS correct_indexes INT[];

-- ========== 6. subject_attempts：统计字段 ==========
ALTER TABLE subject_attempts
  ADD COLUMN IF NOT EXISTS duration_seconds INT;

ALTER TABLE subject_attempts
  ADD COLUMN IF NOT EXISTS passed BOOLEAN;

CREATE INDEX IF NOT EXISTS idx_subject_attempts_user_subject_submitted
  ON subject_attempts (user_id, subject_learning_id, submitted_at);
