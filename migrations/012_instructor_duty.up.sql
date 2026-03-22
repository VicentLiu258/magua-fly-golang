-- Migration 012: 教员日排班 — duty_status、remark、duty_windows；每教员每自然日一行

DO $$ BEGIN
    CREATE TYPE instructor_duty_status AS ENUM (
        'morning',
        'afternoon',
        'full_day',
        'leave',
        'rest',
        'custom'
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- 合并同日多段为一条，写入 duty_windows；再删除旧行并去掉 start_time/end_time
DROP TABLE IF EXISTS _instructor_schedules_merged;
CREATE TABLE _instructor_schedules_merged AS
SELECT
    instructor_id,
    date,
    jsonb_agg(
        jsonb_build_object(
            'start', to_char(start_time, 'HH24:MI'),
            'end', to_char(end_time, 'HH24:MI')
        )
        ORDER BY start_time
    ) AS duty_windows,
    min(created_at) AS created_at,
    max(updated_at) AS updated_at
FROM instructor_schedules
GROUP BY instructor_id, date;

DELETE FROM instructor_schedules;

ALTER TABLE instructor_schedules DROP COLUMN IF EXISTS start_time;
ALTER TABLE instructor_schedules DROP COLUMN IF EXISTS end_time;

ALTER TABLE instructor_schedules
    ADD COLUMN IF NOT EXISTS duty_status instructor_duty_status NOT NULL DEFAULT 'full_day',
    ADD COLUMN IF NOT EXISTS remark TEXT,
    ADD COLUMN IF NOT EXISTS duty_windows JSONB;

INSERT INTO instructor_schedules (id, instructor_id, date, duty_status, remark, duty_windows, created_at, updated_at)
SELECT
    gen_random_uuid(),
    instructor_id,
    date,
    'custom'::instructor_duty_status,
    NULL,
    duty_windows,
    created_at,
    updated_at
FROM _instructor_schedules_merged;

DROP TABLE IF EXISTS _instructor_schedules_merged;

ALTER TABLE instructor_schedules DROP CONSTRAINT IF EXISTS instructor_schedules_instructor_id_date_key;
ALTER TABLE instructor_schedules ADD CONSTRAINT instructor_schedules_instructor_id_date_key UNIQUE (instructor_id, date);

COMMENT ON COLUMN instructor_schedules.duty_status IS '出勤：morning/afternoon/full_day/leave/rest/custom（custom 表示显式 duty_windows）';
COMMENT ON COLUMN instructor_schedules.remark IS '备注';
COMMENT ON COLUMN instructor_schedules.duty_windows IS '可选；null 时由 duty_status 与业务日 09:00–19:00、14:00 上下午分界推导；非 null 为显式窗口 [{"start":"HH:MM","end":"HH:MM"},...]，并裁剪到业务日内';
