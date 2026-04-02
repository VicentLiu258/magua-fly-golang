-- 学员训练课程阶段（SPL / PPL / FD），与训练记录汇总展示配合使用
ALTER TABLE students
    ADD COLUMN IF NOT EXISTS training_course VARCHAR(8);

ALTER TABLE students
    DROP CONSTRAINT IF EXISTS students_training_course_check;
ALTER TABLE students
    ADD CONSTRAINT students_training_course_check
        CHECK (training_course IS NULL OR training_course IN ('SPL', 'PPL', 'FD'));

COMMENT ON COLUMN students.training_course IS '训练课程阶段：SPL | PPL | FD';
