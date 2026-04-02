-- 学员权益手工调整流水（与 students.contract_info 权益字段联动）
CREATE TABLE student_benefit_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    benefit_kind VARCHAR(32) NOT NULL,
    direction VARCHAR(16) NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    reason TEXT,
    operator_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    operator_name VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    idempotency_key VARCHAR(64),
    meta JSONB,
    CONSTRAINT student_benefit_ledger_amount_positive CHECK (amount > 0),
    CONSTRAINT student_benefit_ledger_kind_chk CHECK (
        benefit_kind IN ('flight_time', 'guest_experience', 'learning_points')
    ),
    CONSTRAINT student_benefit_ledger_dir_chk CHECK (direction IN ('credit', 'debit'))
);

CREATE INDEX idx_student_benefit_ledger_student_kind_created
    ON student_benefit_ledger (student_id, benefit_kind, created_at DESC);

CREATE UNIQUE INDEX student_benefit_ledger_idempotency_uq
    ON student_benefit_ledger (idempotency_key)
    WHERE idempotency_key IS NOT NULL AND btrim(idempotency_key) <> '';

COMMENT ON TABLE student_benefit_ledger IS '学员权益调整流水：飞行时间/亲友体验/积分等手工增减';
