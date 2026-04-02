DROP INDEX IF EXISTS idx_training_records_record_serial_no;

ALTER TABLE training_records
    DROP COLUMN IF EXISTS flight_data_ref,
    DROP COLUMN IF EXISTS video_validity,
    DROP COLUMN IF EXISTS video_operator,
    DROP COLUMN IF EXISTS video_timeliness,
    DROP COLUMN IF EXISTS engine_stop_meter,
    DROP COLUMN IF EXISTS engine_start_meter,
    DROP COLUMN IF EXISTS business_type,
    DROP COLUMN IF EXISTS record_serial_no;

DROP TABLE IF EXISTS training_record_daily_seq;
