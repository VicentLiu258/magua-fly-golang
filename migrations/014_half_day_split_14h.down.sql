-- 仅恢复列注释为 12:00 分界（不修改业务数据）

COMMENT ON COLUMN instructor_schedules.duty_windows IS '可选；null 时由 duty_status 与业务日 09:00–19:00、12:00 分界推导；非 null 为显式窗口 [{"start":"HH:MM","end":"HH:MM"},...]，并裁剪到业务日内';
