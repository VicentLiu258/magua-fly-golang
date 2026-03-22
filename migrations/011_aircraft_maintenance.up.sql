-- Migration 011: 飞机机时/定检字段、维护窗口、availability 关联

ALTER TABLE aircraft ADD COLUMN IF NOT EXISTS total_airframe_hours NUMERIC(12, 2);
ALTER TABLE aircraft ADD COLUMN IF NOT EXISTS last_inspection_hours NUMERIC(12, 2);
ALTER TABLE aircraft ADD COLUMN IF NOT EXISTS inspection_interval_hours INT NOT NULL DEFAULT 50;

COMMENT ON COLUMN aircraft.total_airframe_hours IS '机体/计时器总飞行小时（业务口径与 Hobbs 等一致）';
COMMENT ON COLUMN aircraft.last_inspection_hours IS '上次完成同级别定检时的总小时';
COMMENT ON COLUMN aircraft.inspection_interval_hours IS '定检间隔小时数，默认 50';

DO $$ BEGIN
    CREATE TYPE maintenance_window_status AS ENUM ('planned', 'in_progress', 'completed', 'cancelled');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS aircraft_maintenance_windows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aircraft_id UUID NOT NULL REFERENCES aircraft(id) ON DELETE CASCADE,
    kind VARCHAR(64) NOT NULL DEFAULT 'periodic_inspection',
    scheduled_start_at TIMESTAMPTZ NOT NULL,
    scheduled_end_at TIMESTAMPTZ NOT NULL,
    status maintenance_window_status NOT NULL DEFAULT 'planned',
    hangar_location VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_aircraft_maintenance_windows_aircraft_time
    ON aircraft_maintenance_windows (aircraft_id, scheduled_start_at);

COMMENT ON TABLE aircraft_maintenance_windows IS '飞机定检/进机库等维护时间窗';
COMMENT ON COLUMN aircraft_maintenance_windows.kind IS '例如 periodic_inspection、annual 等';

ALTER TABLE aircraft_availability
    ADD COLUMN IF NOT EXISTS maintenance_window_id UUID REFERENCES aircraft_maintenance_windows(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_aircraft_availability_maintenance_window
    ON aircraft_availability (maintenance_window_id);

COMMENT ON COLUMN aircraft_availability.maintenance_window_id IS '若由维护计划生成，关联维护窗口';
