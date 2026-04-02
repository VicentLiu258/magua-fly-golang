ALTER TABLE students
    ADD COLUMN IF NOT EXISTS english_name VARCHAR(256);

COMMENT ON COLUMN students.english_name IS '英文名';
