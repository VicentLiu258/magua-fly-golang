-- 014: 与代码一致：上/下午执勤分界由 12:00 调整为 14:00（见 internal/usecase/schedule.go defaultHalfDaySplitMin）

COMMENT ON COLUMN instructor_schedules.duty_windows IS '可选；null 时由 duty_status 与业务日 09:00–19:00、14:00 上下午分界推导；非 null 为显式窗口 [{"start":"HH:MM","end":"HH:MM"},...]，并裁剪到业务日内';
