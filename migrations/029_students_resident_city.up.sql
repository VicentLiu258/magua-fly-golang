-- 学员常驻城市（档案展示与筛选用）
ALTER TABLE students
    ADD COLUMN IF NOT EXISTS resident_city VARCHAR(128);

COMMENT ON COLUMN students.resident_city IS '常驻城市';
