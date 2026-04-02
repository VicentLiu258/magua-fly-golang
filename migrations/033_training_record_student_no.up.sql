-- training_records / training_archives：冗余存储学员学号，与 students.student_no 对应；便于导入导出与避免将学号误判为 UUID。
ALTER TABLE training_records
    ADD COLUMN IF NOT EXISTS student_no VARCHAR(64);
ALTER TABLE training_archives
    ADD COLUMN IF NOT EXISTS student_no VARCHAR(64);

COMMENT ON COLUMN training_records.student_no IS '学员学号快照，与 students.student_no 对应';
COMMENT ON COLUMN training_archives.student_no IS '学员学号快照，与 students.student_no 对应';

UPDATE training_records tr
SET student_no = s.student_no
FROM students s
WHERE s.id = tr.student_id
  AND (tr.student_no IS NULL OR TRIM(tr.student_no) = '');

UPDATE training_archives ta
SET student_no = s.student_no
FROM students s
WHERE s.id = ta.student_id
  AND (ta.student_no IS NULL OR TRIM(ta.student_no) = '');
