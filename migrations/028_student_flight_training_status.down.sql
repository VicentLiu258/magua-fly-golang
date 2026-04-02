DROP INDEX IF EXISTS idx_students_flight_training_status;
ALTER TABLE students DROP COLUMN IF EXISTS flight_training_status;
DROP TYPE IF EXISTS student_flight_training_status;
