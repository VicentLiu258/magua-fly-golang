ALTER TABLE contracts DROP CONSTRAINT IF EXISTS contracts_payment_schedule_is_array;
ALTER TABLE contracts DROP CONSTRAINT IF EXISTS contracts_payment_mode_check;
ALTER TABLE contracts DROP COLUMN IF EXISTS payment_schedule;
ALTER TABLE contracts DROP COLUMN IF EXISTS payment_mode;
