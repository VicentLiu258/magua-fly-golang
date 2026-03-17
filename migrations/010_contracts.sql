-- Migration 010: 学员合同信息与合同文档关联
-- 1. contracts 表：结构化合同信息（UUID 主键）
-- 2. contracts 与 students 的 UUID 外键关联
-- 3. 预留 MongoDB 合同文档 GridFS 关联字段

-- ========== 1. contracts 表 ==========
CREATE TABLE IF NOT EXISTS contracts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  contract_no VARCHAR(64),
  sign_date DATE,
  start_date DATE,
  end_date DATE,
  total_minutes INT,
  price_total NUMERIC(12,2),
  currency VARCHAR(8),
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  mongo_file_ids TEXT[], -- 关联 MongoDB GridFS / 元数据文档 ID（字符串形式）
  remark TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE contracts IS '学员合同表（结构化记录合同信息）';
COMMENT ON COLUMN contracts.student_id IS '关联学员 ID（students.id）';
COMMENT ON COLUMN contracts.contract_no IS '合同编号';
COMMENT ON COLUMN contracts.sign_date IS '合同签署日期';
COMMENT ON COLUMN contracts.start_date IS '合同生效日期';
COMMENT ON COLUMN contracts.end_date IS '合同结束日期';
COMMENT ON COLUMN contracts.total_minutes IS '合同约定总训练时长（分钟）';
COMMENT ON COLUMN contracts.price_total IS '合同总金额';
COMMENT ON COLUMN contracts.currency IS '币种，如 CNY/USD';
COMMENT ON COLUMN contracts.status IS '合同状态：active/finished/cancelled 等';
COMMENT ON COLUMN contracts.mongo_file_ids IS '关联的合同扫描件/附件在 MongoDB 中的文件 ID 列表';

-- ========== 2. 常用索引 ==========
CREATE INDEX IF NOT EXISTS idx_contracts_student_id ON contracts (student_id);
CREATE INDEX IF NOT EXISTS idx_contracts_status ON contracts (status);

