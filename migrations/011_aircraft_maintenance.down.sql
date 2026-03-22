-- Rollback 011

ALTER TABLE aircraft_availability DROP COLUMN IF EXISTS maintenance_window_id;

DROP TABLE IF EXISTS aircraft_maintenance_windows;

DROP TYPE IF EXISTS maintenance_window_status;

ALTER TABLE aircraft DROP COLUMN IF EXISTS total_airframe_hours;
ALTER TABLE aircraft DROP COLUMN IF EXISTS last_inspection_hours;
ALTER TABLE aircraft DROP COLUMN IF EXISTS inspection_interval_hours;
