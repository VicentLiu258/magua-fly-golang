-- 013: 业务日结束时间统一为 19:00（与 internal/usecase/schedule.go defaultScheduleConfig 一致）
-- 将「开始时间在 19:00 及之后」或「结束时间晚于 19:00」的待审核预约标记为已拒绝（已审核通过记录需人工改期，不在此脚本自动变更）

UPDATE bookings
SET
    status = 'rejected',
    reject_reason = '业务日可约时段结束于19:00，该预约时段已不可用',
    reviewed_at = COALESCE(reviewed_at, NOW()),
    updated_at = NOW()
WHERE status = 'pending'
  AND (start_time >= TIME '19:00' OR end_time > TIME '19:00');

COMMENT ON COLUMN instructor_schedules.duty_windows IS '可选；null 时由 duty_status 与业务日 09:00–19:00、14:00 上下午分界推导；非 null 为显式窗口 [{"start":"HH:MM","end":"HH:MM"},...]，并裁剪到业务日内';
