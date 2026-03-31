-- 飞机头像（GridFS 文件 ID）与业务状态（现役/退役/合作）
ALTER TABLE aircraft ADD COLUMN IF NOT EXISTS avatar_gridfs_id VARCHAR(64);

COMMENT ON COLUMN aircraft.avatar_gridfs_id IS '头像文件在 MongoDB GridFS 中的文件 ID（ObjectId 字符串）';

UPDATE aircraft SET status = 'active'
WHERE status IS NULL OR TRIM(status) = ''
   OR status NOT IN ('active', 'retired', 'partnership');

ALTER TABLE aircraft DROP CONSTRAINT IF EXISTS chk_aircraft_status;
ALTER TABLE aircraft ADD CONSTRAINT chk_aircraft_status
    CHECK (status IN ('active', 'retired', 'partnership'));

COMMENT ON COLUMN aircraft.status IS '业务状态：active=现役，retired=退役，partnership=合作';
