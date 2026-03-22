-- 015: 工作人员（地勤/行政等 staff 角色）日排班，与教员排班字段语义一致，按 user_id 关联；不可预约，仅考勤

CREATE TABLE IF NOT EXISTS staff_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    duty_status instructor_duty_status NOT NULL DEFAULT 'full_day',
    remark TEXT,
    duty_windows JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT staff_schedules_user_id_date_key UNIQUE (user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_staff_schedules_user_date ON staff_schedules(user_id, date);
CREATE INDEX IF NOT EXISTS idx_staff_schedules_date ON staff_schedules(date);

COMMENT ON TABLE staff_schedules IS '工作人员日排班（考勤）：staff 账号，不可预约';
COMMENT ON COLUMN staff_schedules.user_id IS 'users.id，角色应为 staff';
COMMENT ON COLUMN staff_schedules.duty_status IS '与 instructor_schedules 相同枚举';
COMMENT ON COLUMN staff_schedules.duty_windows IS '与 instructor_schedules.duty_windows 一致';
