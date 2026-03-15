-- Migration 007: 训练数据录入与管理
-- 包含：training_records 状态与扩展、raw_flight_imports、training_record_views、training_archives 责任/主任教员

-- ========== 1. training_record_status 枚举 ==========
DO $$ BEGIN
  CREATE TYPE training_record_status AS ENUM ('normal', 'cancelled', 'deleted', 'expired');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
COMMENT ON TYPE training_record_status IS '训练记录状态：normal 正常、cancelled 取消、deleted 删除、expired 已过期';

-- ========== 2. training_records 扩展字段 ==========
ALTER TABLE training_records
  ADD COLUMN IF NOT EXISTS status training_record_status NOT NULL DEFAULT 'normal';

ALTER TABLE training_records
  ADD COLUMN IF NOT EXISTS planned_duration_minutes INT;

ALTER TABLE training_records
  ADD COLUMN IF NOT EXISTS theory_duration_minutes INT;

ALTER TABLE training_records
  ADD COLUMN IF NOT EXISTS sim_duration_minutes INT;

ALTER TABLE training_records
  ADD COLUMN IF NOT EXISTS syllabus_outline VARCHAR(64);

ALTER TABLE training_records
  ADD COLUMN IF NOT EXISTS archive_display_no VARCHAR(64);

COMMENT ON COLUMN training_records.status IS '记录状态';
COMMENT ON COLUMN training_records.planned_duration_minutes IS '计划时长（分钟）';
COMMENT ON COLUMN training_records.theory_duration_minutes IS '理论时长（分钟）';
COMMENT ON COLUMN training_records.sim_duration_minutes IS '模拟时长（分钟）';
COMMENT ON COLUMN training_records.syllabus_outline IS '训练大纲（如 1#、2#）';
COMMENT ON COLUMN training_records.archive_display_no IS '训练档案编号（如 1#672）';

CREATE INDEX IF NOT EXISTS idx_training_records_status ON training_records (status);
CREATE INDEX IF NOT EXISTS idx_training_records_training_date ON training_records (training_date);

-- ========== 3. raw_flight_imports 表（待认领导入）==========
CREATE TABLE IF NOT EXISTS raw_flight_imports (
  id BIGSERIAL PRIMARY KEY,
  batch_id VARCHAR(64) NOT NULL,
  engine_on_time TIMESTAMPTZ,
  engine_off_time TIMESTAMPTZ,
  flight_duration_minutes INT,
  is_local BOOLEAN DEFAULT TRUE,
  raw_data JSONB,
  training_record_id BIGINT REFERENCES training_records(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE raw_flight_imports IS '飞行数据导入待认领表';
COMMENT ON COLUMN raw_flight_imports.batch_id IS '导入批次ID';
COMMENT ON COLUMN raw_flight_imports.engine_on_time IS '开车时间';
COMMENT ON COLUMN raw_flight_imports.engine_off_time IS '关车时间';
COMMENT ON COLUMN raw_flight_imports.flight_duration_minutes IS '飞行时长（分钟）';
COMMENT ON COLUMN raw_flight_imports.is_local IS '是否本场';
COMMENT ON COLUMN raw_flight_imports.raw_data IS '原始导入数据';
COMMENT ON COLUMN raw_flight_imports.training_record_id IS '认领后关联的训练记录ID';

CREATE INDEX IF NOT EXISTS idx_raw_flight_imports_batch ON raw_flight_imports (batch_id);
CREATE INDEX IF NOT EXISTS idx_raw_flight_imports_claimed ON raw_flight_imports (training_record_id) WHERE training_record_id IS NULL;

-- ========== 4. training_record_views 表（自定义视图）==========
CREATE TABLE IF NOT EXISTS training_record_views (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name VARCHAR(128) NOT NULL,
  filter_json JSONB,
  sort_json JSONB,
  group_by VARCHAR(64),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE training_record_views IS '训练记录自定义视图';
COMMENT ON COLUMN training_record_views.filter_json IS '筛选条件 JSON';
COMMENT ON COLUMN training_record_views.sort_json IS '排序 JSON';
COMMENT ON COLUMN training_record_views.group_by IS '分组字段';

CREATE INDEX IF NOT EXISTS idx_training_record_views_user ON training_record_views (user_id);

-- ========== 5. training_archives 责任/主任教员 ==========
ALTER TABLE training_archives
  ADD COLUMN IF NOT EXISTS responsible_instructor_id BIGINT REFERENCES instructors(id) ON DELETE SET NULL;

ALTER TABLE training_archives
  ADD COLUMN IF NOT EXISTS chief_instructor_id BIGINT REFERENCES instructors(id) ON DELETE SET NULL;

COMMENT ON COLUMN training_archives.responsible_instructor_id IS '责任教员ID';
COMMENT ON COLUMN training_archives.chief_instructor_id IS '主任教员ID';

-- ========== 6. training_archive_subjects 科目内容与类型（可选）==========
ALTER TABLE training_archive_subjects
  ADD COLUMN IF NOT EXISTS subject_content TEXT;

ALTER TABLE training_archive_subjects
  ADD COLUMN IF NOT EXISTS type VARCHAR(32);

COMMENT ON COLUMN training_archive_subjects.subject_content IS '科目内容';
COMMENT ON COLUMN training_archive_subjects.type IS '科目类型';

CREATE INDEX IF NOT EXISTS idx_training_archives_student_date ON training_archives (student_id, training_date);
