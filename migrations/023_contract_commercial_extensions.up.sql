-- Migration 023: 合同商业/台账扩展（与 contracts 1:1，不改现有表与业务逻辑）
-- 用途：承接 CRM 式台账字段（客户展示名、区域、负责人、开票回款、里程碑与监控 checklist 等）
-- 约定：里程碑、监控项、账务多码等放 JSONB，避免继续加列；需要强类型时可再拆子表

CREATE TABLE IF NOT EXISTS contract_commercial_extensions (
    contract_id UUID PRIMARY KEY REFERENCES contracts(id) ON DELETE CASCADE,

    external_reference_code VARCHAR(128),
    fiscal_year INT,
    project_status VARCHAR(64),
    customer_display_name VARCHAR(255),
    product_series VARCHAR(128),
    region VARCHAR(128),

    owner_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    coordinator_user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    health_status VARCHAR(32),
    quality_note TEXT,

    contract_kind VARCHAR(64),
    term_description VARCHAR(128),

    invoiced_amount NUMERIC(14, 2),
    paid_amount NUMERIC(14, 2),
    outstanding_amount NUMERIC(14, 2),

    finance_meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    milestones JSONB NOT NULL DEFAULT '[]'::jsonb,
    monitoring_checklist JSONB NOT NULL DEFAULT '{}'::jsonb,
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT contract_commercial_extensions_finance_meta_object
        CHECK (jsonb_typeof(finance_meta) = 'object'),
    CONSTRAINT contract_commercial_extensions_milestones_array
        CHECK (jsonb_typeof(milestones) = 'array'),
    CONSTRAINT contract_commercial_extensions_monitoring_object
        CHECK (jsonb_typeof(monitoring_checklist) = 'object'),
    CONSTRAINT contract_commercial_extensions_extra_object
        CHECK (jsonb_typeof(extra) = 'object')
);

COMMENT ON TABLE contract_commercial_extensions IS '合同商业台账扩展：与 contracts 一行对一行，可选填写；训练核心字段仍在 contracts/students';
COMMENT ON COLUMN contract_commercial_extensions.contract_id IS '关联合同主表，主键即外键';
COMMENT ON COLUMN contract_commercial_extensions.external_reference_code IS '外部/业务台账编号（如 HD、N- 前缀单号）';
COMMENT ON COLUMN contract_commercial_extensions.fiscal_year IS '业务年份（台账「年份」列）';
COMMENT ON COLUMN contract_commercial_extensions.project_status IS '项目/台账状态展示文案或短码（如 进行中、已完成）';
COMMENT ON COLUMN contract_commercial_extensions.customer_display_name IS '客户名称展示；未来若有 customers 表可迁走，此处可逐步弃用';
COMMENT ON COLUMN contract_commercial_extensions.product_series IS '产品型号/系列等展示标签';
COMMENT ON COLUMN contract_commercial_extensions.region IS '所属区域';
COMMENT ON COLUMN contract_commercial_extensions.owner_user_id IS '负责人（销售/教务等业务账号 users.id）';
COMMENT ON COLUMN contract_commercial_extensions.coordinator_user_id IS '配合人';
COMMENT ON COLUMN contract_commercial_extensions.health_status IS '进展质量：建议 normal / warning / risk 或中文文案';
COMMENT ON COLUMN contract_commercial_extensions.quality_note IS '进展与质量说明';
COMMENT ON COLUMN contract_commercial_extensions.contract_kind IS '项目种类：如新签、续签';
COMMENT ON COLUMN contract_commercial_extensions.term_description IS '期限描述：如一年期、分期';
COMMENT ON COLUMN contract_commercial_extensions.invoiced_amount IS '已开票金额';
COMMENT ON COLUMN contract_commercial_extensions.paid_amount IS '已回款金额';
COMMENT ON COLUMN contract_commercial_extensions.outstanding_amount IS '待收金额';
COMMENT ON COLUMN contract_commercial_extensions.finance_meta IS '账务/参考码等多列信息，如 {"codes":["..."],"refs":{}}';
COMMENT ON COLUMN contract_commercial_extensions.milestones IS '里程碑数组，如 [{"code":"acceptance","label":"验收","date":"2020-06-01","done":true}]';
COMMENT ON COLUMN contract_commercial_extensions.monitoring_checklist IS '监控项键值，如 {"contract":"ok","invoice":"pending"}，键可对应原表头';
COMMENT ON COLUMN contract_commercial_extensions.extra IS '其它扩展字段，避免再改表结构';

CREATE INDEX IF NOT EXISTS idx_contract_commercial_extensions_owner
    ON contract_commercial_extensions (owner_user_id)
    WHERE owner_user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_contract_commercial_extensions_fiscal_year
    ON contract_commercial_extensions (fiscal_year)
    WHERE fiscal_year IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_contract_commercial_extensions_region
    ON contract_commercial_extensions (region)
    WHERE region IS NOT NULL;
