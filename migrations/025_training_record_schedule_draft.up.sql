-- 排期驱动的待补全飞行记录：状态 draft、来源 schedule、关联 training_plans

DO $block$
BEGIN
    ALTER TYPE training_record_status ADD VALUE 'draft';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END
$block$;

DO $block$
BEGIN
    ALTER TYPE import_source ADD VALUE 'schedule';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END
$block$;

COMMENT ON TYPE training_record_status IS '训练记录状态：normal 正常、cancelled 取消、deleted 删除、expired 已过期、draft 待补全（来自排期）';

ALTER TABLE training_records
    ADD COLUMN IF NOT EXISTS training_plan_id UUID REFERENCES training_plans (id) ON DELETE SET NULL;

COMMENT ON COLUMN training_records.training_plan_id IS '关联训练计划（预约通过后生成）；一条计划最多对应一条训练记录草稿';

CREATE UNIQUE INDEX IF NOT EXISTS idx_training_records_one_per_plan
    ON training_records (training_plan_id)
    WHERE training_plan_id IS NOT NULL;
