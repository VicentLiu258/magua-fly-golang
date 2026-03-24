DROP TABLE IF EXISTS syllabus_subject_secondaries;

ALTER TABLE syllabus_subjects DROP CONSTRAINT IF EXISTS syllabus_subjects_outline_code_unique;
ALTER TABLE syllabus_subjects DROP COLUMN IF EXISTS syllabus_outline_id;

DROP TABLE IF EXISTS syllabus_outlines;
