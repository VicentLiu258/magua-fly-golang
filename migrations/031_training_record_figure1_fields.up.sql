-- 飞行记录图一扩展：业务记录号、业务类型、小时表读数、视频元数据、飞行数据引用、按日序号

CREATE TABLE IF NOT EXISTS training_record_daily_seq (
    seq_date DATE PRIMARY KEY,
    last_seq INT NOT NULL DEFAULT 0
);

COMMENT ON TABLE training_record_daily_seq IS '飞行记录业务编号 TR+日期+序号 的按日递增';

ALTER TABLE training_records
    ADD COLUMN IF NOT EXISTS record_serial_no VARCHAR(32),
    ADD COLUMN IF NOT EXISTS business_type VARCHAR(128),
    ADD COLUMN IF NOT EXISTS engine_start_meter DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS engine_stop_meter DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS video_timeliness VARCHAR(32),
    ADD COLUMN IF NOT EXISTS video_operator VARCHAR(128),
    ADD COLUMN IF NOT EXISTS video_validity VARCHAR(32),
    ADD COLUMN IF NOT EXISTS flight_data_ref VARCHAR(512);

COMMENT ON COLUMN training_records.record_serial_no IS '业务记录编号 TR+YYYYMMDD+三位序号';
COMMENT ON COLUMN training_records.business_type IS '业务类型（展示/导入中文或规范 code）';
COMMENT ON COLUMN training_records.engine_start_meter IS '发动机开车表读数（小时表）';
COMMENT ON COLUMN training_records.engine_stop_meter IS '发动机关车表读数';
COMMENT ON COLUMN training_records.video_timeliness IS '视频及时 T+0/T+1 等';
COMMENT ON COLUMN training_records.video_operator IS '拍摄人';
COMMENT ON COLUMN training_records.video_validity IS '视频有效';
COMMENT ON COLUMN training_records.flight_data_ref IS '飞行数据附件 URL 或对象标识';

CREATE UNIQUE INDEX IF NOT EXISTS idx_training_records_record_serial_no
    ON training_records (record_serial_no)
    WHERE record_serial_no IS NOT NULL AND btrim(record_serial_no) <> '';
