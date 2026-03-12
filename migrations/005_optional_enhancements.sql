-- Migration 005: 可选增强（按需执行）
-- 包含：基地表、学员进度 JSONB、训练记录扩展字段、档案签署人

-- ========== 1. bases 表（机场/基地维度，供排班天气/空域等使用）==========
CREATE TABLE IF NOT EXISTS bases (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  icao VARCHAR(10),
  city VARCHAR(100),
  region VARCHAR(100),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE bases IS '机场/基地维度，可与 aircraft、排班日配置关联';

-- 可选：aircraft 关联基地
-- ALTER TABLE aircraft ADD COLUMN IF NOT EXISTS base_id BIGINT REFERENCES bases(id);
-- CREATE INDEX IF NOT EXISTS idx_aircraft_base_id ON aircraft (base_id);

-- ========== 2. 日级运行配置（天气/空域，对应 Schedule 页 day.weather / day.airspace）==========
CREATE TABLE IF NOT EXISTS day_operational_config (
  id BIGSERIAL PRIMARY KEY,
  date DATE NOT NULL,
  base_id BIGINT REFERENCES bases(id),
  weather VARCHAR(500),
  airspace VARCHAR(500),
  notes TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_day_operational_config_date_base
  ON day_operational_config (date, COALESCE(base_id, 0));

-- ========== 3. students.training_progress 结构化（JSONB）==========
-- 新增一列，保留原 TEXT 直至数据迁移完成
ALTER TABLE students
  ADD COLUMN IF NOT EXISTS training_progress_json JSONB;

COMMENT ON COLUMN students.training_progress_json IS '结构化进度，如 {"phase":"PPL-1","completedMinutes":1200,"remainingMinutes":1200}；与 training_progress 二选一或同步';

-- ========== 4. training_records 扩展字段（与前端 Records 页对齐）==========
ALTER TABLE training_records
  ADD COLUMN IF NOT EXISTS departure VARCHAR(100);

ALTER TABLE training_records
  ADD COLUMN IF NOT EXISTS destination VARCHAR(100);

ALTER TABLE training_records
  ADD COLUMN IF NOT EXISTS flight_type VARCHAR(20);

ALTER TABLE training_records
  ADD COLUMN IF NOT EXISTS task_type VARCHAR(50);

ALTER TABLE training_records
  ADD COLUMN IF NOT EXISTS landings INT;

-- ========== 5. training_archives 签署人 ID（可选）==========
ALTER TABLE training_archives
  ADD COLUMN IF NOT EXISTS student_signer_id BIGINT REFERENCES users(id);

ALTER TABLE training_archives
  ADD COLUMN IF NOT EXISTS responsible_instructor_signer_id BIGINT REFERENCES users(id);

ALTER TABLE training_archives
  ADD COLUMN IF NOT EXISTS chief_instructor_signer_id BIGINT REFERENCES users(id);

-- ========== 6. theory_documents 分类（与前端 Category 对齐）==========
ALTER TABLE theory_documents
  ADD COLUMN IF NOT EXISTS category VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_theory_documents_category ON theory_documents (category);
