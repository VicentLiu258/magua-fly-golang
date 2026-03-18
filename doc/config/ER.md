erDiagram
users {
UUID id PK
VARCHAR username
VARCHAR password_hash
VARCHAR phone
VARCHAR email
user_role role
VARCHAR avatar_url
TIMESTAMPTZ created_at
TIMESTAMPTZ updated_at
}

    students {
        UUID id PK
        UUID user_id FK
        VARCHAR name
        VARCHAR student_no
        JSONB contract_info
        INT contract_duration_minutes
        INT flown_minutes
        TEXT training_progress
        UUID responsible_instructor_id FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    instructors {
        UUID id PK
        UUID user_id FK
        VARCHAR name
        VARCHAR display_color
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    aircraft {
        UUID id PK
        VARCHAR code
        VARCHAR name
        VARCHAR status
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    aircraft_availability {
        UUID id PK
        UUID aircraft_id FK
        DATE date
        TIME start_time
        TIME end_time
        availability_status status
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    instructor_schedules {
        UUID id PK
        UUID instructor_id FK
        DATE date
        TIME start_time
        TIME end_time
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    training_plans {
        UUID id PK
        UUID student_id FK
        UUID instructor_id FK
        UUID aircraft_id FK
        DATE plan_date
        TIME start_time
        TIME end_time
        plan_source source
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    bookings {
        UUID id PK
        UUID student_id FK
        UUID instructor_id FK
        UUID aircraft_id FK
        DATE plan_date
        TIME start_time
        TIME end_time
        booking_status status
        TEXT reject_reason
        TIMESTAMPTZ reviewed_at
        UUID reviewer_id FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    training_records {
        UUID id PK
        UUID student_id FK
        UUID instructor_id FK
        UUID aircraft_id FK
        DATE training_date
        TIMESTAMPTZ engine_on_time
        TIMESTAMPTZ engine_off_time
        INT flight_duration_minutes
        BOOLEAN is_local
        import_source import_source
        VARCHAR raw_import_id
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    training_archives {
        UUID id PK
        UUID training_record_id FK UNIQUE
        UUID student_id FK
        UUID instructor_id FK
        DATE training_date
        VARCHAR stage
        INT planned_duration_minutes
        INT total_experience_minutes
        INT remaining_minutes
        TIMESTAMPTZ engine_on_time
        TIMESTAMPTZ engine_off_time
        INT flight_duration_minutes
        TEXT progress_note
        TEXT instructor_review
        TEXT student_self_review
        TEXT task_assignment
        TIMESTAMPTZ student_signed_at
        TIMESTAMPTZ responsible_instructor_signed_at
        TIMESTAMPTZ chief_instructor_signed_at
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    syllabus_subjects {
        UUID id PK
        VARCHAR code
        VARCHAR name
        INT sort_order
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    training_archive_subjects {
        UUID id PK
        UUID training_archive_id FK
        UUID syllabus_subject_id FK
        INT evaluation_score
        BOOLEAN is_graduated
        INT sort_order
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    next_training_subjects {
        UUID id PK
        UUID training_archive_id FK
        UUID syllabus_subject_id FK
        UUID assigned_by FK
        TIMESTAMPTZ created_at
    }

    theory_documents {
        UUID id PK
        VARCHAR title
        VARCHAR slug
        INT sort_order
        VARCHAR content_ref
        VARCHAR video_url
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    theory_questions {
        UUID id PK
        UUID theory_document_id FK
        TEXT question_text
        JSONB options
        INT correct_index
        INT sort_order
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    theory_attempts {
        UUID id PK
        UUID user_id FK
        UUID theory_document_id FK
        JSONB answers
        INT score
        TIMESTAMPTZ submitted_at
    }

    subject_learning {
        UUID id PK
        UUID syllabus_subject_id FK UNIQUE
        VARCHAR title
        VARCHAR content_ref
        VARCHAR video_url
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    subject_questions {
        UUID id PK
        UUID subject_learning_id FK
        TEXT question_text
        JSONB options
        INT correct_index
        INT sort_order
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    subject_attempts {
        UUID id PK
        UUID user_id FK
        UUID subject_learning_id FK
        JSONB answers
        INT score
        TIMESTAMPTZ submitted_at
    }

    tasks {
        UUID id PK
        UUID student_id FK
        UUID assigner_id FK
        task_type type
        UUID ref_id
        VARCHAR title
        TEXT description
        TIMESTAMPTZ due_at
        task_status status
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    comments {
        UUID id PK
        comment_target_type target_type
        UUID target_id
        UUID user_id FK
        UUID parent_id FK
        TEXT content
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    %% 关系

    users ||--o{ students : "user_id"
    users ||--o{ instructors : "user_id"
    instructors ||--o{ students : "responsible_instructor_id"

    aircraft ||--o{ aircraft_availability : "aircraft_id"
    aircraft ||--o{ training_plans : "aircraft_id"
    aircraft ||--o{ bookings : "aircraft_id"
    aircraft ||--o{ training_records : "aircraft_id"

    instructors ||--o{ instructor_schedules : "instructor_id"
    instructors ||--o{ training_plans : "instructor_id"
    instructors ||--o{ bookings : "instructor_id"
    instructors ||--o{ training_records : "instructor_id"
    instructors ||--o{ training_archives : "instructor_id"

    students ||--o{ training_plans : "student_id"
    students ||--o{ bookings : "student_id"
    students ||--o{ training_records : "student_id"
    students ||--o{ training_archives : "student_id"
    students ||--o{ tasks : "student_id"

    training_records ||--|| training_archives : "training_record_id"

    syllabus_subjects ||--o{ training_archive_subjects : "syllabus_subject_id"
    syllabus_subjects ||--|| subject_learning : "syllabus_subject_id"

    training_archives ||--o{ training_archive_subjects : "training_archive_id"
    training_archives ||--o{ next_training_subjects : "training_archive_id"

    users ||--o{ next_training_subjects : "assigned_by"

    theory_documents ||--o{ theory_questions : "theory_document_id"
    users ||--o{ theory_attempts : "user_id"
    theory_documents ||--o{ theory_attempts : "theory_document_id"

    subject_learning ||--o{ subject_questions : "subject_learning_id"
    users ||--o{ subject_attempts : "user_id"
    subject_learning ||--o{ subject_attempts : "subject_learning_id"

    users ||--o{ tasks : "assigner_id"

    users ||--o{ comments : "user_id"
    comments ||--o{ comments : "parent_id"
