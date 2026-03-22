-- 无法可靠撤销 013 中对 bookings 的批量拒绝，此处仅恢复列注释（与 21:00 口径一致的历史说明）

COMMENT ON COLUMN instructor_schedules.duty_windows IS '可选；null 时由 duty_status 与业务日 09:00–21:00、14:00 上下午分界推导；非 null 为显式窗口 [{"start":"HH:MM","end":"HH:MM"},...]';
