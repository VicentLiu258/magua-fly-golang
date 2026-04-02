-- 030: 预约任务类型（训练/体验/熟练/验证/试飞/调机/内训/本场带飞）

DO $$ BEGIN
    CREATE TYPE booking_task_type AS ENUM (
        'training',
        'experience',
        'proficiency',
        'verification',
        'test_flight',
        'ferry',
        'internal_training',
        'local_pattern'
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS task_type booking_task_type NOT NULL DEFAULT 'training';

COMMENT ON COLUMN bookings.task_type IS '预约任务类型：训练/体验/熟练/验证/试飞/调机/内训/本场带飞';
