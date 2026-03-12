CREATE TYPE user_role AS ENUM ('admin', 'staff', 'instructor', 'student');
COMMENT ON TYPE user_role IS '用户角色枚举：admin（管理员）、staff（工作人员）、instructor（教员）、student（学员）';

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
COMMENT ON TABLE users IS '用户账号表，包含所有角色的登录账户';
COMMENT ON COLUMN users.id IS '主键ID';
COMMENT ON COLUMN users.username IS '用户名，唯一，用于登录';
COMMENT ON COLUMN users.password_hash IS '密码哈希值';
COMMENT ON COLUMN users.phone IS '手机号';
COMMENT ON COLUMN users.email IS '邮箱';
COMMENT ON COLUMN users.role IS '用户角色';
COMMENT ON COLUMN users.avatar_url IS '头像URL地址';
COMMENT ON COLUMN users.created_at IS '创建时间';
COMMENT ON COLUMN users.updated_at IS '更新时间';

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
COMMENT ON TABLE students IS '学员信息表，可关联用户账号';
COMMENT ON COLUMN students.id IS '主键ID';
COMMENT ON COLUMN students.user_id IS '关联的用户ID，为空表示仅学员资料';
COMMENT ON COLUMN students.name IS '学员姓名';
COMMENT ON COLUMN students.student_no IS '学员编号，唯一';
COMMENT ON COLUMN students.contract_info IS '与学员相关的合同信息（JSON）';
COMMENT ON COLUMN students.contract_duration_minutes IS '合同约定总时长（分钟）';
COMMENT ON COLUMN students.flown_minutes IS '已飞行时长（分钟）';
COMMENT ON COLUMN students.training_progress IS '训练进度描述';
COMMENT ON COLUMN students.responsible_instructor_id IS '负责该学员的教员ID';
COMMENT ON COLUMN students.created_at IS '创建时间';
COMMENT ON COLUMN students.updated_at IS '更新时间';

-- Instructors (extends user)
CREATE TABLE instructors (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    display_color VARCHAR(32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE instructors IS '教员信息表';
COMMENT ON COLUMN instructors.id IS '主键ID';
COMMENT ON COLUMN instructors.user_id IS '关联的用户ID';
COMMENT ON COLUMN instructors.name IS '教员姓名';
COMMENT ON COLUMN instructors.display_color IS '前端显示用颜色标识';
COMMENT ON COLUMN instructors.created_at IS '创建时间';
COMMENT ON COLUMN instructors.updated_at IS '更新时间';

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
COMMENT ON TABLE aircraft IS '航空器信息表';
COMMENT ON COLUMN aircraft.id IS '主键ID';
COMMENT ON COLUMN aircraft.code IS '航空器编号/注册号';
COMMENT ON COLUMN aircraft.name IS '航空器名称或型号';
COMMENT ON COLUMN aircraft.status IS '航空器状态，例如active/maintenance';
COMMENT ON COLUMN aircraft.created_at IS '创建时间';
COMMENT ON COLUMN aircraft.updated_at IS '更新时间';

-- Aircraft availability (per hour or slot)
CREATE TYPE availability_status AS ENUM ('available', 'maintenance', 'blocked');
COMMENT ON TYPE availability_status IS '航空器可用状态枚举：available（可用）、maintenance（维修）、blocked（占用/封锁）';

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
COMMENT ON TABLE aircraft_availability IS '航空器可用时段表（按日期与时间段）';
COMMENT ON COLUMN aircraft_availability.id IS '主键ID';
COMMENT ON COLUMN aircraft_availability.aircraft_id IS '航空器ID';
COMMENT ON COLUMN aircraft_availability.date IS '日期';
COMMENT ON COLUMN aircraft_availability.start_time IS '开始时间';
COMMENT ON COLUMN aircraft_availability.end_time IS '结束时间';
COMMENT ON COLUMN aircraft_availability.status IS '该时间段内航空器状态';
COMMENT ON COLUMN aircraft_availability.created_at IS '创建时间';
COMMENT ON COLUMN aircraft_availability.updated_at IS '更新时间';

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
COMMENT ON TABLE instructor_schedules IS '教员值班/排班表';
COMMENT ON COLUMN instructor_schedules.id IS '主键ID';
COMMENT ON COLUMN instructor_schedules.instructor_id IS '教员ID';
COMMENT ON COLUMN instructor_schedules.date IS '值班日期';
COMMENT ON COLUMN instructor_schedules.start_time IS '开始时间';
COMMENT ON COLUMN instructor_schedules.end_time IS '结束时间';
COMMENT ON COLUMN instructor_schedules.created_at IS '创建时间';
COMMENT ON COLUMN instructor_schedules.updated_at IS '更新时间';

-- Training plans (confirmed slot: student + instructor + aircraft)
CREATE TYPE plan_source AS ENUM ('booking', 'manual');
COMMENT ON TYPE plan_source IS '训练计划来源枚举：booking（来源于学员预约）、manual（手工创建）';

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
COMMENT ON TABLE training_plans IS '飞行训练计划表（已确认的排班）';
COMMENT ON COLUMN training_plans.id IS '主键ID';
COMMENT ON COLUMN training_plans.student_id IS '学员ID';
COMMENT ON COLUMN training_plans.instructor_id IS '教员ID';
COMMENT ON COLUMN training_plans.aircraft_id IS '航空器ID';
COMMENT ON COLUMN training_plans.plan_date IS '计划训练日期';
COMMENT ON COLUMN training_plans.start_time IS '计划开始时间';
COMMENT ON COLUMN training_plans.end_time IS '计划结束时间';
COMMENT ON COLUMN training_plans.source IS '计划来源';
COMMENT ON COLUMN training_plans.created_at IS '创建时间';
COMMENT ON COLUMN training_plans.updated_at IS '更新时间';

-- Bookings (student request, needs approval)
CREATE TYPE booking_status AS ENUM ('pending', 'approved', 'rejected');
COMMENT ON TYPE booking_status IS '预约状态枚举：pending（待审核）、approved（已通过）、rejected（已拒绝）';

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
COMMENT ON TABLE bookings IS '学员预约记录表（待审核/已审核）';
COMMENT ON COLUMN bookings.id IS '主键ID';
COMMENT ON COLUMN bookings.student_id IS '学员ID';
COMMENT ON COLUMN bookings.instructor_id IS '期望教员ID';
COMMENT ON COLUMN bookings.aircraft_id IS '期望航空器ID';
COMMENT ON COLUMN bookings.plan_date IS '预约日期';
COMMENT ON COLUMN bookings.start_time IS '预约开始时间';
COMMENT ON COLUMN bookings.end_time IS '预约结束时间';
COMMENT ON COLUMN bookings.status IS '预约状态';
COMMENT ON COLUMN bookings.reject_reason IS '预约被拒绝原因';
COMMENT ON COLUMN bookings.reviewed_at IS '审核时间';
COMMENT ON COLUMN bookings.reviewer_id IS '审核人用户ID';
COMMENT ON COLUMN bookings.created_at IS '创建时间';
COMMENT ON COLUMN bookings.updated_at IS '更新时间';

-- Training records (flight data)
CREATE TYPE import_source AS ENUM ('manual', 'import');
COMMENT ON TYPE import_source IS '记录来源枚举：manual（手工录入）、import（外部导入）';

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
COMMENT ON TABLE training_records IS '飞行训练原始记录表';
COMMENT ON COLUMN training_records.id IS '主键ID';
COMMENT ON COLUMN training_records.student_id IS '学员ID';
COMMENT ON COLUMN training_records.instructor_id IS '教员ID';
COMMENT ON COLUMN training_records.aircraft_id IS '航空器ID';
COMMENT ON COLUMN training_records.training_date IS '训练日期';
COMMENT ON COLUMN training_records.engine_on_time IS '发动机开车时间';
COMMENT ON COLUMN training_records.engine_off_time IS '发动机关车时间';
COMMENT ON COLUMN training_records.flight_duration_minutes IS '飞行时长（分钟）';
COMMENT ON COLUMN training_records.is_local IS '是否本场飞行（true为本场）';
COMMENT ON COLUMN training_records.import_source IS '记录来源';
COMMENT ON COLUMN training_records.raw_import_id IS '外部导入记录的原始ID';
COMMENT ON COLUMN training_records.created_at IS '创建时间';
COMMENT ON COLUMN training_records.updated_at IS '更新时间';

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
COMMENT ON TABLE training_archives IS '飞行训练档案表（与训练记录1:1，对训练过程进行归档与评语）';
COMMENT ON COLUMN training_archives.id IS '主键ID';
COMMENT ON COLUMN training_archives.training_record_id IS '关联的训练记录ID（唯一）';
COMMENT ON COLUMN training_archives.student_id IS '学员ID';
COMMENT ON COLUMN training_archives.instructor_id IS '教员ID';
COMMENT ON COLUMN training_archives.training_date IS '训练日期';
COMMENT ON COLUMN training_archives.stage IS '训练阶段/科目阶段';
COMMENT ON COLUMN training_archives.planned_duration_minutes IS '计划飞行时长（分钟）';
COMMENT ON COLUMN training_archives.total_experience_minutes IS '累计飞行经验时长（分钟）';
COMMENT ON COLUMN training_archives.remaining_minutes IS '剩余需完成时长（分钟）';
COMMENT ON COLUMN training_archives.engine_on_time IS '发动机开车时间';
COMMENT ON COLUMN training_archives.engine_off_time IS '发动机关车时间';
COMMENT ON COLUMN training_archives.flight_duration_minutes IS '实际飞行时长（分钟）';
COMMENT ON COLUMN training_archives.progress_note IS '训练进度记录';
COMMENT ON COLUMN training_archives.instructor_review IS '教员评语';
COMMENT ON COLUMN training_archives.student_self_review IS '学员自评';
COMMENT ON COLUMN training_archives.task_assignment IS '布置的训练任务/作业';
COMMENT ON COLUMN training_archives.student_signed_at IS '学员签字时间';
COMMENT ON COLUMN training_archives.responsible_instructor_signed_at IS '负责教员签字时间';
COMMENT ON COLUMN training_archives.chief_instructor_signed_at IS '总教签字时间';
COMMENT ON COLUMN training_archives.created_at IS '创建时间';
COMMENT ON COLUMN training_archives.updated_at IS '更新时间';

-- Syllabus subjects (training outline)
CREATE TABLE syllabus_subjects (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE syllabus_subjects IS '训练大纲科目表';
COMMENT ON COLUMN syllabus_subjects.id IS '主键ID';
COMMENT ON COLUMN syllabus_subjects.code IS '科目代码';
COMMENT ON COLUMN syllabus_subjects.name IS '科目名称';
COMMENT ON COLUMN syllabus_subjects.sort_order IS '排序序号';
COMMENT ON COLUMN syllabus_subjects.created_at IS '创建时间';
COMMENT ON COLUMN syllabus_subjects.updated_at IS '更新时间';

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
COMMENT ON TABLE training_archive_subjects IS '训练档案对应的具体科目及评定表';
COMMENT ON COLUMN training_archive_subjects.id IS '主键ID';
COMMENT ON COLUMN training_archive_subjects.training_archive_id IS '训练档案ID';
COMMENT ON COLUMN training_archive_subjects.syllabus_subject_id IS '大纲科目ID';
COMMENT ON COLUMN training_archive_subjects.evaluation_score IS '科目评价分数';
COMMENT ON COLUMN training_archive_subjects.is_graduated IS '该科目是否结业/通过';
COMMENT ON COLUMN training_archive_subjects.sort_order IS '排序序号';
COMMENT ON COLUMN training_archive_subjects.created_at IS '创建时间';
COMMENT ON COLUMN training_archive_subjects.updated_at IS '更新时间';

-- Next training subjects (assigned by instructor for next session)
CREATE TABLE next_training_subjects (
    id BIGSERIAL PRIMARY KEY,
    training_archive_id BIGINT NOT NULL REFERENCES training_archives(id) ON DELETE CASCADE,
    syllabus_subject_id BIGINT NOT NULL REFERENCES syllabus_subjects(id) ON DELETE CASCADE,
    assigned_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE next_training_subjects IS '下次训练指定科目表（由教员分配）';
COMMENT ON COLUMN next_training_subjects.id IS '主键ID';
COMMENT ON COLUMN next_training_subjects.training_archive_id IS '关联的训练档案ID';
COMMENT ON COLUMN next_training_subjects.syllabus_subject_id IS '指定的下次训练科目ID';
COMMENT ON COLUMN next_training_subjects.assigned_by IS '分配人用户ID（一般为教员）';
COMMENT ON COLUMN next_training_subjects.created_at IS '创建时间';

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
COMMENT ON TABLE theory_documents IS '理论学习文档元数据表（正文存储在MongoDB）';
COMMENT ON COLUMN theory_documents.id IS '主键ID';
COMMENT ON COLUMN theory_documents.title IS '文档标题';
COMMENT ON COLUMN theory_documents.slug IS '文档短标识（用于URL），唯一';
COMMENT ON COLUMN theory_documents.sort_order IS '排序序号';
COMMENT ON COLUMN theory_documents.content_ref IS '内容在外部存储中的引用ID';
COMMENT ON COLUMN theory_documents.video_url IS '配套视频URL';
COMMENT ON COLUMN theory_documents.created_at IS '创建时间';
COMMENT ON COLUMN theory_documents.updated_at IS '更新时间';

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
COMMENT ON TABLE theory_questions IS '理论题库问题表';
COMMENT ON COLUMN theory_questions.id IS '主键ID';
COMMENT ON COLUMN theory_questions.theory_document_id IS '所属理论文档ID';
COMMENT ON COLUMN theory_questions.question_text IS '题目正文';
COMMENT ON COLUMN theory_questions.options IS '选项列表（JSON数组）';
COMMENT ON COLUMN theory_questions.correct_index IS '正确选项下标';
COMMENT ON COLUMN theory_questions.sort_order IS '排序序号';
COMMENT ON COLUMN theory_questions.created_at IS '创建时间';
COMMENT ON COLUMN theory_questions.updated_at IS '更新时间';

-- Theory attempts (quiz results)
CREATE TABLE theory_attempts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    theory_document_id BIGINT NOT NULL REFERENCES theory_documents(id) ON DELETE CASCADE,
    answers JSONB NOT NULL,
    score INT NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE theory_attempts IS '理论考试/测验作答记录表';
COMMENT ON COLUMN theory_attempts.id IS '主键ID';
COMMENT ON COLUMN theory_attempts.user_id IS '作答人用户ID';
COMMENT ON COLUMN theory_attempts.theory_document_id IS '关联的理论文档ID';
COMMENT ON COLUMN theory_attempts.answers IS '作答内容（JSON）';
COMMENT ON COLUMN theory_attempts.score IS '得分';
COMMENT ON COLUMN theory_attempts.submitted_at IS '提交时间';

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
COMMENT ON TABLE subject_learning IS '大纲科目对应的学习内容表（1:1）';
COMMENT ON COLUMN subject_learning.id IS '主键ID';
COMMENT ON COLUMN subject_learning.syllabus_subject_id IS '大纲科目ID（唯一）';
COMMENT ON COLUMN subject_learning.title IS '学习内容标题';
COMMENT ON COLUMN subject_learning.content_ref IS '内容在外部存储中的引用ID';
COMMENT ON COLUMN subject_learning.video_url IS '配套视频URL';
COMMENT ON COLUMN subject_learning.created_at IS '创建时间';
COMMENT ON COLUMN subject_learning.updated_at IS '更新时间';

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
COMMENT ON TABLE subject_questions IS '科目题库问题表（与subject_learning关联）';
COMMENT ON COLUMN subject_questions.id IS '主键ID';
COMMENT ON COLUMN subject_questions.subject_learning_id IS '所属学习内容ID';
COMMENT ON COLUMN subject_questions.question_text IS '题目正文';
COMMENT ON COLUMN subject_questions.options IS '选项列表（JSON数组）';
COMMENT ON COLUMN subject_questions.correct_index IS '正确选项下标';
COMMENT ON COLUMN subject_questions.sort_order IS '排序序号';
COMMENT ON COLUMN subject_questions.created_at IS '创建时间';
COMMENT ON COLUMN subject_questions.updated_at IS '更新时间';

-- Subject attempts
CREATE TABLE subject_attempts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_learning_id BIGINT NOT NULL REFERENCES subject_learning(id) ON DELETE CASCADE,
    answers JSONB NOT NULL,
    score INT NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE subject_attempts IS '科目练习/测验作答记录表';
COMMENT ON COLUMN subject_attempts.id IS '主键ID';
COMMENT ON COLUMN subject_attempts.user_id IS '作答人用户ID';
COMMENT ON COLUMN subject_attempts.subject_learning_id IS '关联的学习内容ID';
COMMENT ON COLUMN subject_attempts.answers IS '作答内容（JSON）';
COMMENT ON COLUMN subject_attempts.score IS '得分';
COMMENT ON COLUMN subject_attempts.submitted_at IS '提交时间';

-- Tasks (unified student tasks)
CREATE TYPE task_type AS ENUM ('theory', 'subject', 'other');
CREATE TYPE task_status AS ENUM ('pending', 'done', 'overdue');
COMMENT ON TYPE task_type IS '任务类型枚举：theory（理论）、subject（科目）、other（其他）';
COMMENT ON TYPE task_status IS '任务状态枚举：pending（待完成）、done（已完成）、overdue（已逾期）';

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
COMMENT ON TABLE tasks IS '统一的学员任务表（理论/科目等）';
COMMENT ON COLUMN tasks.id IS '主键ID';
COMMENT ON COLUMN tasks.student_id IS '学员ID';
COMMENT ON COLUMN tasks.assigner_id IS '任务指派人用户ID';
COMMENT ON COLUMN tasks.type IS '任务类型';
COMMENT ON COLUMN tasks.ref_id IS '关联对象ID，例如理论文档ID或科目ID';
COMMENT ON COLUMN tasks.title IS '任务标题';
COMMENT ON COLUMN tasks.description IS '任务说明';
COMMENT ON COLUMN tasks.due_at IS '任务截止时间';
COMMENT ON COLUMN tasks.status IS '任务状态';
COMMENT ON COLUMN tasks.created_at IS '创建时间';
COMMENT ON COLUMN tasks.updated_at IS '更新时间';

-- Comments (optional in PG for pagination; tree in MongoDB)
CREATE TYPE comment_target_type AS ENUM ('theory_document', 'subject_learning');
COMMENT ON TYPE comment_target_type IS '评论目标类型枚举：theory_document（理论文档）、subject_learning（科目学习内容）';

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
COMMENT ON TABLE comments IS '评论记录表（树结构保存在MongoDB，此处用于分页和基础信息）';
COMMENT ON COLUMN comments.id IS '主键ID';
COMMENT ON COLUMN comments.target_type IS '评论目标类型';
COMMENT ON COLUMN comments.target_id IS '评论目标ID';
COMMENT ON COLUMN comments.user_id IS '评论人用户ID';
COMMENT ON COLUMN comments.parent_id IS '父级评论ID';
COMMENT ON COLUMN comments.content IS '评论内容';
COMMENT ON COLUMN comments.created_at IS '创建时间';
COMMENT ON COLUMN comments.updated_at IS '更新时间';

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
