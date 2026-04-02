-- 学员学飞状态：进行中、已暂停、已完结、已终止（API/库内码见注释）
DO $$ BEGIN
    CREATE TYPE student_flight_training_status AS ENUM (
        'in_progress',
        'paused',
        'completed',
        'terminated'
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE students
    ADD COLUMN IF NOT EXISTS flight_training_status student_flight_training_status NOT NULL DEFAULT 'in_progress';

COMMENT ON COLUMN students.flight_training_status IS '学飞状态：in_progress 进行中 | paused 已暂停 | completed 已完结 | terminated 已终止';

CREATE INDEX IF NOT EXISTS idx_students_flight_training_status
    ON students (flight_training_status)
    WHERE deleted_at IS NULL;
