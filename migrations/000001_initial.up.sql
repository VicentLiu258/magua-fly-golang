-- User role enum
CREATE TYPE user_role AS ENUM ('admin', 'staff', 'instructor', 'student');

-- Users (accounts for all roles)
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(32),
    email VARCHAR(255),
    role user_role NOT NULL,
    avatar_url VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Students (extends user or standalone; responsible_instructor_id FK added after instructors)
CREATE TABLE students (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    name VARCHAR(128) NOT NULL,
    student_no VARCHAR(64) NOT NULL UNIQUE,
    contract_info JSONB,
    contract_duration_minutes INT,
    flown_minutes INT DEFAULT 0,
    training_progress TEXT,
    responsible_instructor_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Instructors (extends user)
CREATE TABLE instructors (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    display_color VARCHAR(32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE students ADD CONSTRAINT fk_students_responsible_instructor
    FOREIGN KEY (responsible_instructor_id) REFERENCES instructors(id) ON DELETE SET NULL;

-- Aircraft
CREATE TABLE aircraft (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(128),
    status VARCHAR(32) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Aircraft availability (per hour or slot)
CREATE TYPE availability_status AS ENUM ('available', 'maintenance', 'blocked');

CREATE TABLE aircraft_availability (
    id BIGSERIAL PRIMARY KEY,
    aircraft_id BIGINT NOT NULL REFERENCES aircraft(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    status availability_status NOT NULL DEFAULT 'available',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (aircraft_id, date, start_time, end_time)
);

-- Instructor schedules (duty)
CREATE TABLE instructor_schedules (
    id BIGSERIAL PRIMARY KEY,
    instructor_id BIGINT NOT NULL REFERENCES instructors(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Training plans (confirmed slot: student + instructor + aircraft)
CREATE TYPE plan_source AS ENUM ('booking', 'manual');

CREATE TABLE training_plans (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    instructor_id BIGINT NOT NULL REFERENCES instructors(id) ON DELETE CASCADE,
    aircraft_id BIGINT NOT NULL REFERENCES aircraft(id) ON DELETE CASCADE,
    plan_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    source plan_source NOT NULL DEFAULT 'manual',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Bookings (student request, needs approval)
CREATE TYPE booking_status AS ENUM ('pending', 'approved', 'rejected');

CREATE TABLE bookings (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    instructor_id BIGINT NOT NULL REFERENCES instructors(id) ON DELETE CASCADE,
    aircraft_id BIGINT NOT NULL REFERENCES aircraft(id) ON DELETE CASCADE,
    plan_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    status booking_status NOT NULL DEFAULT 'pending',
    reject_reason TEXT,
    reviewed_at TIMESTAMPTZ,
    reviewer_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Training records (flight data)
CREATE TYPE import_source AS ENUM ('manual', 'import');

CREATE TABLE training_records (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    instructor_id BIGINT NOT NULL REFERENCES instructors(id) ON DELETE CASCADE,
    aircraft_id BIGINT NOT NULL REFERENCES aircraft(id) ON DELETE CASCADE,
    training_date DATE NOT NULL,
    engine_on_time TIMESTAMPTZ,
    engine_off_time TIMESTAMPTZ,
    flight_duration_minutes INT,
    is_local BOOLEAN DEFAULT TRUE,
    import_source import_source NOT NULL DEFAULT 'manual',
    raw_import_id VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Training archives (1:1 with record, rich fields)
CREATE TABLE training_archives (
    id BIGSERIAL PRIMARY KEY,
    training_record_id BIGINT NOT NULL UNIQUE REFERENCES training_records(id) ON DELETE CASCADE,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    instructor_id BIGINT NOT NULL REFERENCES instructors(id) ON DELETE CASCADE,
    training_date DATE NOT NULL,
    stage VARCHAR(64),
    planned_duration_minutes INT,
    total_experience_minutes INT,
    remaining_minutes INT,
    engine_on_time TIMESTAMPTZ,
    engine_off_time TIMESTAMPTZ,
    flight_duration_minutes INT,
    progress_note TEXT,
    instructor_review TEXT,
    student_self_review TEXT,
    task_assignment TEXT,
    student_signed_at TIMESTAMPTZ,
    responsible_instructor_signed_at TIMESTAMPTZ,
    chief_instructor_signed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Syllabus subjects (training outline)
CREATE TABLE syllabus_subjects (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Training archive subjects (subjects + evaluation per archive)
CREATE TABLE training_archive_subjects (
    id BIGSERIAL PRIMARY KEY,
    training_archive_id BIGINT NOT NULL REFERENCES training_archives(id) ON DELETE CASCADE,
    syllabus_subject_id BIGINT NOT NULL REFERENCES syllabus_subjects(id) ON DELETE CASCADE,
    evaluation_score INT,
    is_graduated BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Next training subjects (assigned by instructor for next session)
CREATE TABLE next_training_subjects (
    id BIGSERIAL PRIMARY KEY,
    training_archive_id BIGINT NOT NULL REFERENCES training_archives(id) ON DELETE CASCADE,
    syllabus_subject_id BIGINT NOT NULL REFERENCES syllabus_subjects(id) ON DELETE CASCADE,
    assigned_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Theory documents (metadata; content in MongoDB)
CREATE TABLE theory_documents (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(128) UNIQUE,
    sort_order INT NOT NULL DEFAULT 0,
    content_ref VARCHAR(64),
    video_url VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Theory questions
CREATE TABLE theory_questions (
    id BIGSERIAL PRIMARY KEY,
    theory_document_id BIGINT NOT NULL REFERENCES theory_documents(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    options JSONB NOT NULL,
    correct_index INT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Theory attempts (quiz results)
CREATE TABLE theory_attempts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    theory_document_id BIGINT NOT NULL REFERENCES theory_documents(id) ON DELETE CASCADE,
    answers JSONB NOT NULL,
    score INT NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Subject learning (1:1 with syllabus subject, metadata)
CREATE TABLE subject_learning (
    id BIGSERIAL PRIMARY KEY,
    syllabus_subject_id BIGINT NOT NULL UNIQUE REFERENCES syllabus_subjects(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content_ref VARCHAR(64),
    video_url VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Subject questions (same shape as theory)
CREATE TABLE subject_questions (
    id BIGSERIAL PRIMARY KEY,
    subject_learning_id BIGINT NOT NULL REFERENCES subject_learning(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    options JSONB NOT NULL,
    correct_index INT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Subject attempts
CREATE TABLE subject_attempts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_learning_id BIGINT NOT NULL REFERENCES subject_learning(id) ON DELETE CASCADE,
    answers JSONB NOT NULL,
    score INT NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tasks (unified student tasks)
CREATE TYPE task_type AS ENUM ('theory', 'subject', 'other');
CREATE TYPE task_status AS ENUM ('pending', 'done', 'overdue');

CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    assigner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type task_type NOT NULL,
    ref_id BIGINT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    due_at TIMESTAMPTZ,
    status task_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Comments (optional in PG for pagination; tree in MongoDB)
CREATE TYPE comment_target_type AS ENUM ('theory_document', 'subject_learning');

CREATE TABLE comments (
    id BIGSERIAL PRIMARY KEY,
    target_type comment_target_type NOT NULL,
    target_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id BIGINT REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_target ON comments (target_type, target_id);
CREATE INDEX idx_comments_parent ON comments (parent_id);

-- Indexes for common queries
CREATE INDEX idx_students_user_id ON students(user_id);
CREATE INDEX idx_students_responsible_instructor ON students(responsible_instructor_id);
CREATE INDEX idx_aircraft_availability_aircraft_date ON aircraft_availability(aircraft_id, date);
CREATE INDEX idx_instructor_schedules_instructor_date ON instructor_schedules(instructor_id, date);
CREATE INDEX idx_training_plans_date ON training_plans(plan_date);
CREATE INDEX idx_bookings_student_status ON bookings(student_id, status);
CREATE INDEX idx_training_records_student_date ON training_records(student_id, training_date);
CREATE INDEX idx_training_archives_student ON training_archives(student_id);
CREATE INDEX idx_tasks_student_status ON tasks(student_id, status);
CREATE INDEX idx_theory_attempts_user_doc ON theory_attempts(user_id, theory_document_id);
CREATE INDEX idx_subject_attempts_user_subject ON subject_attempts(user_id, subject_learning_id);
