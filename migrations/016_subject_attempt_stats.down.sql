ALTER TABLE subject_attempts DROP COLUMN IF EXISTS total_count;
ALTER TABLE subject_attempts DROP COLUMN IF EXISTS correct_count;
ALTER TABLE subject_attempts DROP COLUMN IF EXISTS max_score;
ALTER TABLE subject_attempts DROP COLUMN IF EXISTS passed;
ALTER TABLE subject_attempts DROP COLUMN IF EXISTS duration_seconds;

ALTER TABLE subject_learning DROP COLUMN IF EXISTS sort_order;
