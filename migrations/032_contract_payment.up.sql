-- 合同付款方式与分期付款计划
ALTER TABLE contracts
    ADD COLUMN IF NOT EXISTS payment_mode VARCHAR(16) NOT NULL DEFAULT 'full',
    ADD COLUMN IF NOT EXISTS payment_schedule JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE contracts
    ADD CONSTRAINT contracts_payment_mode_check
        CHECK (payment_mode IN ('full', 'installment'));

ALTER TABLE contracts
    ADD CONSTRAINT contracts_payment_schedule_is_array
        CHECK (jsonb_typeof(payment_schedule) = 'array');

COMMENT ON COLUMN contracts.payment_mode IS 'full=全款 installment=分期';
COMMENT ON COLUMN contracts.payment_schedule IS '分期计划 JSON 数组 [{"amount":n,"due_date":"YYYY-MM-DD"},...]；全款时为空数组';
