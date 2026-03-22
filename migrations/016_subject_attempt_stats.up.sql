-- subject_learning.sort_order（003 曾加，008 重建表后缺失）
ALTER TABLE subject_learning ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0;

-- Align subject_attempts summary columns with theory_attempts for API parity.
-- (UUID schema 008 重建表后未包含 003 中曾加的 duration_seconds / passed，此处一并补齐。)
ALTER TABLE subject_attempts ADD COLUMN IF NOT EXISTS duration_seconds INT;
ALTER TABLE subject_attempts ADD COLUMN IF NOT EXISTS passed BOOLEAN;
ALTER TABLE subject_attempts ADD COLUMN IF NOT EXISTS max_score INT;
ALTER TABLE subject_attempts ADD COLUMN IF NOT EXISTS correct_count INT;
ALTER TABLE subject_attempts ADD COLUMN IF NOT EXISTS total_count INT;

COMMENT ON COLUMN subject_attempts.max_score IS '试卷满分（与 Mongo 题库一致）';
COMMENT ON COLUMN subject_attempts.correct_count IS '答对题数';
COMMENT ON COLUMN subject_attempts.total_count IS '本次提交答题数';
