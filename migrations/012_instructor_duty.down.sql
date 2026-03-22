-- Rollback 012: 恢复 start_time/end_time（取 duty_windows 首段或默认 09:00–19:00）
-- 若库从未执行过 012 up（无 duty_windows 列），则跳过 JSON 回填，仅保证存在 start_time/end_time。

ALTER TABLE instructor_schedules DROP CONSTRAINT IF EXISTS instructor_schedules_instructor_id_date_key;

ALTER TABLE instructor_schedules
    ADD COLUMN IF NOT EXISTS start_time TIME NOT NULL DEFAULT TIME '09:00',
    ADD COLUMN IF NOT EXISTS end_time TIME NOT NULL DEFAULT TIME '19:00';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'instructor_schedules'
          AND column_name = 'duty_windows'
    ) THEN
        UPDATE instructor_schedules s
        SET
            start_time = COALESCE(
                (SELECT (elem->>'start')::TIME
                 FROM jsonb_array_elements(COALESCE(s.duty_windows, '[]'::jsonb)) AS elem
                 LIMIT 1),
                TIME '09:00'
            ),
            end_time = COALESCE(
                (SELECT (elem->>'end')::TIME
                 FROM jsonb_array_elements(COALESCE(s.duty_windows, '[]'::jsonb)) AS elem
                 LIMIT 1),
                TIME '19:00'
            );
    END IF;
END $$;

ALTER TABLE instructor_schedules DROP COLUMN IF EXISTS duty_status;
ALTER TABLE instructor_schedules DROP COLUMN IF EXISTS remark;
ALTER TABLE instructor_schedules DROP COLUMN IF EXISTS duty_windows;

DROP TYPE IF EXISTS instructor_duty_status;
