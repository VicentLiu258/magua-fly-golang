-- Migration 002: 训练主线、排班与预约约束与索引（必须/建议调整）
-- 依赖：training_records, training_archives, aircraft_availability, instructor_schedules, training_plans, bookings 已存在

-- ========== 1. training_archives 外键（若尚未存在）==========
-- 确保 training_record_id 有唯一约束且外键存在
-- 若建表时已定义可跳过；否则：
-- ALTER TABLE training_archives ADD CONSTRAINT training_archives_training_record_id_fkey
--   FOREIGN KEY (training_record_id) REFERENCES training_records(id) ON DELETE CASCADE;

-- ========== 2. bookings 关联 training_plans（建议）==========
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'bookings' AND column_name = 'training_plan_id'
  ) THEN
    ALTER TABLE bookings ADD COLUMN training_plan_id BIGINT REFERENCES training_plans(id);
  END IF;
END $$;
CREATE INDEX IF NOT EXISTS idx_bookings_training_plan_id ON bookings (training_plan_id) WHERE training_plan_id IS NOT NULL;

-- ========== 3. 飞机时段唯一性（防止同机同段重复占用）==========
-- 在 training_plans 上：同一飞机、同一天、相同时段只能有一条计划
CREATE UNIQUE INDEX IF NOT EXISTS idx_training_plans_aircraft_slot_unique
  ON training_plans (aircraft_id, plan_date, start_time, end_time);

-- 若 bookings 与 plans 是同一资源池，可在 bookings 上同样约束或二选一
-- CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_aircraft_slot_unique
--   ON bookings (aircraft_id, plan_date, start_time, end_time);

-- ========== 4. 训练与排班高频查询索引 ==========
-- aircraft
CREATE INDEX IF NOT EXISTS idx_aircraft_code ON aircraft (code);
CREATE INDEX IF NOT EXISTS idx_aircraft_status ON aircraft (status);

-- aircraft_availability
CREATE INDEX IF NOT EXISTS idx_aircraft_availability_aircraft_date
  ON aircraft_availability (aircraft_id, date);

-- instructor_schedules
CREATE INDEX IF NOT EXISTS idx_instructor_schedules_instructor_date
  ON instructor_schedules (instructor_id, date);

-- training_plans
CREATE INDEX IF NOT EXISTS idx_training_plans_plan_date ON training_plans (plan_date);
CREATE INDEX IF NOT EXISTS idx_training_plans_instructor_date ON training_plans (instructor_id, plan_date);
CREATE INDEX IF NOT EXISTS idx_training_plans_student_date ON training_plans (student_id, plan_date);

-- bookings
CREATE INDEX IF NOT EXISTS idx_bookings_plan_date ON bookings (plan_date);
CREATE INDEX IF NOT EXISTS idx_bookings_student_status ON bookings (student_id, status);
CREATE INDEX IF NOT EXISTS idx_bookings_instructor_date ON bookings (instructor_id, plan_date);
CREATE INDEX IF NOT EXISTS idx_bookings_aircraft_date ON bookings (aircraft_id, plan_date);

-- training_records
CREATE INDEX IF NOT EXISTS idx_training_records_student_date ON training_records (student_id, training_date);
CREATE INDEX IF NOT EXISTS idx_training_records_instructor_date ON training_records (instructor_id, training_date);
CREATE INDEX IF NOT EXISTS idx_training_records_aircraft_date ON training_records (aircraft_id, training_date);

-- training_archives
CREATE INDEX IF NOT EXISTS idx_training_archives_student_date ON training_archives (student_id, training_date);
CREATE INDEX IF NOT EXISTS idx_training_archives_instructor_date ON training_archives (instructor_id, training_date);
