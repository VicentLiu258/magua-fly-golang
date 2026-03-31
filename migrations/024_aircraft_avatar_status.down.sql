ALTER TABLE aircraft DROP CONSTRAINT IF EXISTS chk_aircraft_status;
ALTER TABLE aircraft DROP COLUMN IF EXISTS avatar_gridfs_id;
