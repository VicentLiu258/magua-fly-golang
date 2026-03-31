DROP INDEX IF EXISTS idx_training_records_one_per_plan;
ALTER TABLE training_records DROP COLUMN IF EXISTS training_plan_id;
-- 注意：PostgreSQL 无法安全删除已使用的枚举取值 'draft' / 'schedule'，故不在 down 中回滚枚举
