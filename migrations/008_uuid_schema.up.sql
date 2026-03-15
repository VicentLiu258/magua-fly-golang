-- Migration 008: 全库 ID 统一为 UUID（清空并重建）
-- 执行前会删除下列所有表，执行后所有主键与外键均为 UUID。

-- ========== 1. 删除所有表（逆依赖顺序）==========
DROP TABLE IF EXISTS comments CASCADE;
DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS subject_attempts CASCADE;
DROP TABLE IF EXISTS subject_questions CASCADE;
DROP TABLE IF EXISTS subject_learning CASCADE;
DROP TABLE IF EXISTS theory_attempts CASCADE;
DROP TABLE IF EXISTS theory_questions CASCADE;
DROP TABLE IF EXISTS theory_documents CASCADE;
DROP TABLE IF EXISTS next_training_subjects CASCADE;
DROP TABLE IF EXISTS training_archive_subjects CASCADE;
DROP TABLE IF EXISTS syllabus_subjects CASCADE;
DROP TABLE IF EXISTS training_archives CASCADE;
DROP TABLE IF EXISTS raw_flight_imports CASCADE;
DROP TABLE IF EXISTS training_record_views CASCADE;
DROP TABLE IF EXISTS training_records CASCADE;
DROP TABLE IF EXISTS bookings CASCADE;
DROP TABLE IF EXISTS training_plans CASCADE;
DROP TABLE IF EXISTS instructor_schedules CASCADE;
DROP TABLE IF EXISTS aircraft_availability CASCADE;
DROP TABLE IF EXISTS aircraft CASCADE;
DROP TABLE IF EXISTS day_operational_config CASCADE;
DROP TABLE IF EXISTS bases CASCADE;
DROP TABLE IF EXISTS instructors CASCADE;
DROP TABLE IF EXISTS students CASCADE;
DROP TABLE IF EXISTS sms_verification_codes CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- ========== 2. 枚举类型（若不存在则创建）==========
DO $$ BEGIN CREATE TYPE user_role AS ENUM ('admin', 'staff', 'instructor', 'student'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE availability_status AS ENUM ('available', 'maintenance', 'blocked'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE plan_source AS ENUM ('booking', 'manual'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE booking_status AS ENUM ('pending', 'approved', 'rejected'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE import_source AS ENUM ('manual', 'import'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE training_record_status AS ENUM ('normal', 'cancelled', 'deleted', 'expired'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE task_type AS ENUM ('theory', 'subject', 'other'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE task_status AS ENUM ('pending', 'done', 'overdue'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE comment_target_type AS ENUM ('theory_document', 'subject_learning'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE sms_scene AS ENUM ('register', 'login', 'reset_pwd', '2fa'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ========== 3. 建表（UUID 主键与外键）==========
-- 用户表：系统账号，含用户名、密码、手机、邮箱、角色、头像等
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(32),
    email VARCHAR(255),
    role user_role NOT NULL,
    avatar_url VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 教员表：关联用户，含姓名、显示颜色，用于排班与训练
CREATE TABLE instructors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    name VARCHAR(128) NOT NULL,
    display_color VARCHAR(32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 学员表：关联用户，含学号、合同、已飞时长、训练进度、负责教员
CREATE TABLE students (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    student_no VARCHAR(64) NOT NULL UNIQUE,
    contract_info JSONB,
    contract_duration_minutes INT,
    flown_minutes INT DEFAULT 0,
    training_progress TEXT,
    training_progress_json JSONB,
    responsible_instructor_id UUID REFERENCES instructors(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 飞机表：飞机编号、名称、状态（如 active）
CREATE TABLE aircraft (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(64) NOT NULL,
    name VARCHAR(128),
    status VARCHAR(32) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 飞机可用性：某日某时段飞机的可用/维护/占用状态
CREATE TABLE aircraft_availability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aircraft_id UUID NOT NULL REFERENCES aircraft(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    status availability_status NOT NULL DEFAULT 'available',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (aircraft_id, date, start_time, end_time)
);

-- 教员排班：某日某时段教员的可飞时段
CREATE TABLE instructor_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instructor_id UUID NOT NULL REFERENCES instructors(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 训练计划：计划某日某时段学员-教员-飞机的安排，来源可为预约或手工
CREATE TABLE training_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    instructor_id UUID NOT NULL REFERENCES instructors(id) ON DELETE CASCADE,
    aircraft_id UUID NOT NULL REFERENCES aircraft(id) ON DELETE CASCADE,
    plan_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    source plan_source NOT NULL DEFAULT 'manual',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 预约表：学员预约某日某时段，状态待审批/通过/拒绝，可填拒绝原因
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    instructor_id UUID NOT NULL REFERENCES instructors(id) ON DELETE CASCADE,
    aircraft_id UUID NOT NULL REFERENCES aircraft(id) ON DELETE CASCADE,
    plan_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    status booking_status NOT NULL DEFAULT 'pending',
    reject_reason TEXT,
    reviewed_at TIMESTAMPTZ,
    reviewer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 飞行训练记录：单次飞行的学员/教员/飞机、起落时间、时长、状态、导入来源等
CREATE TABLE training_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    instructor_id UUID NOT NULL REFERENCES instructors(id) ON DELETE CASCADE,
    aircraft_id UUID NOT NULL REFERENCES aircraft(id) ON DELETE CASCADE,
    training_date DATE NOT NULL,
    engine_on_time TIMESTAMPTZ,
    engine_off_time TIMESTAMPTZ,
    flight_duration_minutes INT,
    is_local BOOLEAN DEFAULT TRUE,
    import_source import_source NOT NULL DEFAULT 'manual',
    raw_import_id VARCHAR(64),
    status training_record_status NOT NULL DEFAULT 'normal',
    planned_duration_minutes INT,
    theory_duration_minutes INT,
    sim_duration_minutes INT,
    syllabus_outline VARCHAR(64),
    archive_display_no VARCHAR(64),
    departure VARCHAR(100),
    destination VARCHAR(100),
    flight_type VARCHAR(20),
    task_type VARCHAR(50),
    landings INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 训练档案：与单次训练记录一一对应，含阶段、进度、签字、负责人与主任教员
CREATE TABLE training_archives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    training_record_id UUID NOT NULL UNIQUE REFERENCES training_records(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    instructor_id UUID NOT NULL REFERENCES instructors(id) ON DELETE CASCADE,
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
    student_signer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    responsible_instructor_signer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    chief_instructor_signer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    responsible_instructor_id UUID REFERENCES instructors(id) ON DELETE SET NULL,
    chief_instructor_id UUID REFERENCES instructors(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 大纲科目：训练大纲中的科目编码、名称与排序
CREATE TABLE syllabus_subjects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 档案科目：某次训练档案下各大纲科目的评估分、是否结业、内容与类型
CREATE TABLE training_archive_subjects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    training_archive_id UUID NOT NULL REFERENCES training_archives(id) ON DELETE CASCADE,
    syllabus_subject_id UUID NOT NULL REFERENCES syllabus_subjects(id) ON DELETE CASCADE,
    evaluation_score INT,
    is_graduated BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    subject_content TEXT,
    type VARCHAR(32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 下次训练科目：为某次训练档案指定的下次要训练的科目及指定人
CREATE TABLE next_training_subjects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    training_archive_id UUID NOT NULL REFERENCES training_archives(id) ON DELETE CASCADE,
    syllabus_subject_id UUID NOT NULL REFERENCES syllabus_subjects(id) ON DELETE CASCADE,
    assigned_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 原始飞行导入：批量导入的原始飞行数据，可关联到某条训练记录
CREATE TABLE raw_flight_imports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id VARCHAR(64) NOT NULL,
    engine_on_time TIMESTAMPTZ,
    engine_off_time TIMESTAMPTZ,
    flight_duration_minutes INT,
    is_local BOOLEAN DEFAULT TRUE,
    raw_data JSONB,
    training_record_id UUID REFERENCES training_records(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 训练记录视图：用户保存的筛选、排序、分组配置（视图名、filter_json、sort_json 等）
CREATE TABLE training_record_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    filter_json JSONB,
    sort_json JSONB,
    group_by VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 理论文档：理论课文档标题、slug、排序、内容引用、视频、分类
CREATE TABLE theory_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(128) UNIQUE,
    sort_order INT NOT NULL DEFAULT 0,
    content_ref VARCHAR(64),
    video_url VARCHAR(512),
    category VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 理论题目：某理论文档下的题目文本、选项、正确答案索引、排序
CREATE TABLE theory_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    theory_document_id UUID NOT NULL REFERENCES theory_documents(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    options JSONB NOT NULL,
    correct_index INT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 理论作答记录：用户对某理论文档的答题（answers）、得分与提交时间
CREATE TABLE theory_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    theory_document_id UUID NOT NULL REFERENCES theory_documents(id) ON DELETE CASCADE,
    answers JSONB NOT NULL,
    score INT NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 科目学习：某大纲科目对应的学习材料（标题、内容引用、视频），一对一
CREATE TABLE subject_learning (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    syllabus_subject_id UUID NOT NULL UNIQUE REFERENCES syllabus_subjects(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content_ref VARCHAR(64),
    video_url VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 科目题目：某科目学习下的题目文本、选项、正确答案索引、排序
CREATE TABLE subject_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_learning_id UUID NOT NULL REFERENCES subject_learning(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    options JSONB NOT NULL,
    correct_index INT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 科目作答记录：用户对某科目学习的答题、得分与提交时间
CREATE TABLE subject_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_learning_id UUID NOT NULL REFERENCES subject_learning(id) ON DELETE CASCADE,
    answers JSONB NOT NULL,
    score INT NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 任务表：分配给学员的理论/科目/其他任务，含截止时间与状态
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    assigner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type task_type NOT NULL,
    ref_id UUID,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    due_at TIMESTAMPTZ,
    status task_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 评论表：对理论文档或科目学习的评论，支持父子级回复
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type comment_target_type NOT NULL,
    target_id UUID NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 基地表：机场/基地名称、ICAO、城市、区域
CREATE TABLE bases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    icao VARCHAR(10),
    city VARCHAR(100),
    region VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 每日运行配置：某日（可选某基地）的天气、空域、备注
CREATE TABLE day_operational_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL,
    base_id UUID REFERENCES bases(id),
    weather VARCHAR(500),
    airspace VARCHAR(500),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 短信验证码：按手机号与场景存储验证码、过期时间、使用时间、IP
CREATE TABLE sms_verification_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone VARCHAR(32) NOT NULL,
    scene sms_scene NOT NULL,
    code VARCHAR(16) NOT NULL,
    ip VARCHAR(64),
    expired_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ========== 4. 表注释（写入 pg_description，便于客户端与 \d+ 查看）==========
COMMENT ON TABLE users IS '用户表：系统账号，含用户名、密码、手机、邮箱、角色、头像等';
COMMENT ON TABLE instructors IS '教员表：关联用户，含姓名、显示颜色，用于排班与训练';
COMMENT ON TABLE students IS '学员表：关联用户，含学号、合同、已飞时长、训练进度、负责教员';
COMMENT ON TABLE aircraft IS '飞机表：飞机编号、名称、状态（如 active）';
COMMENT ON TABLE aircraft_availability IS '飞机可用性：某日某时段飞机的可用/维护/占用状态';
COMMENT ON TABLE instructor_schedules IS '教员排班：某日某时段教员的可飞时段';
COMMENT ON TABLE training_plans IS '训练计划：计划某日某时段学员-教员-飞机的安排，来源可为预约或手工';
COMMENT ON TABLE bookings IS '预约表：学员预约某日某时段，状态待审批/通过/拒绝，可填拒绝原因';
COMMENT ON TABLE training_records IS '飞行训练记录：单次飞行的学员/教员/飞机、起落时间、时长、状态、导入来源等';
COMMENT ON TABLE training_archives IS '训练档案：与单次训练记录一一对应，含阶段、进度、签字、负责人与主任教员';
COMMENT ON TABLE syllabus_subjects IS '大纲科目：训练大纲中的科目编码、名称与排序';
COMMENT ON TABLE training_archive_subjects IS '档案科目：某次训练档案下各大纲科目的评估分、是否结业、内容与类型';
COMMENT ON TABLE next_training_subjects IS '下次训练科目：为某次训练档案指定的下次要训练的科目及指定人';
COMMENT ON TABLE raw_flight_imports IS '原始飞行导入：批量导入的原始飞行数据，可关联到某条训练记录';
COMMENT ON TABLE training_record_views IS '训练记录视图：用户保存的筛选、排序、分组配置';
COMMENT ON TABLE theory_documents IS '理论文档：理论课文档标题、slug、排序、内容引用、视频、分类';
COMMENT ON TABLE theory_questions IS '理论题目：某理论文档下的题目文本、选项、正确答案索引、排序';
COMMENT ON TABLE theory_attempts IS '理论作答记录：用户对某理论文档的答题、得分与提交时间';
COMMENT ON TABLE subject_learning IS '科目学习：某大纲科目对应的学习材料（标题、内容引用、视频），一对一';
COMMENT ON TABLE subject_questions IS '科目题目：某科目学习下的题目文本、选项、正确答案索引、排序';
COMMENT ON TABLE subject_attempts IS '科目作答记录：用户对某科目学习的答题、得分与提交时间';
COMMENT ON TABLE tasks IS '任务表：分配给学员的理论/科目/其他任务，含截止时间与状态';
COMMENT ON TABLE comments IS '评论表：对理论文档或科目学习的评论，支持父子级回复';
COMMENT ON TABLE bases IS '基地表：机场/基地名称、ICAO、城市、区域';
COMMENT ON TABLE day_operational_config IS '每日运行配置：某日（可选某基地）的天气、空域、备注';
COMMENT ON TABLE sms_verification_codes IS '短信验证码：按手机号与场景存储验证码、过期时间、使用时间、IP';

-- ========== 4.1 列注释（写入 pg_description，\d+ 表名 可查看）==========
-- users
COMMENT ON COLUMN users.id IS '主键 UUID';
COMMENT ON COLUMN users.username IS '登录用户名，唯一';
COMMENT ON COLUMN users.password_hash IS '密码哈希';
COMMENT ON COLUMN users.phone IS '手机号';
COMMENT ON COLUMN users.email IS '邮箱';
COMMENT ON COLUMN users.role IS '角色：admin/staff/instructor/student';
COMMENT ON COLUMN users.avatar_url IS '头像 URL';
COMMENT ON COLUMN users.created_at IS '创建时间';
COMMENT ON COLUMN users.updated_at IS '更新时间';
-- instructors
COMMENT ON COLUMN instructors.id IS '主键 UUID';
COMMENT ON COLUMN instructors.user_id IS '关联用户表，一对一';
COMMENT ON COLUMN instructors.name IS '教员姓名';
COMMENT ON COLUMN instructors.display_color IS '日历/排班显示颜色';
COMMENT ON COLUMN instructors.created_at IS '创建时间';
COMMENT ON COLUMN instructors.updated_at IS '更新时间';
-- students
COMMENT ON COLUMN students.id IS '主键 UUID';
COMMENT ON COLUMN students.user_id IS '关联用户表，可空（未绑定账号时）';
COMMENT ON COLUMN students.name IS '学员姓名';
COMMENT ON COLUMN students.student_no IS '学号，唯一';
COMMENT ON COLUMN students.contract_info IS '合同信息 JSON';
COMMENT ON COLUMN students.contract_duration_minutes IS '合同约定飞行时长（分钟）';
COMMENT ON COLUMN students.flown_minutes IS '已飞时长（分钟）';
COMMENT ON COLUMN students.training_progress IS '训练进度文本（兼容旧版）';
COMMENT ON COLUMN students.training_progress_json IS '训练进度结构化 JSON';
COMMENT ON COLUMN students.responsible_instructor_id IS '负责教员';
COMMENT ON COLUMN students.created_at IS '创建时间';
COMMENT ON COLUMN students.updated_at IS '更新时间';
-- aircraft
COMMENT ON COLUMN aircraft.id IS '主键 UUID';
COMMENT ON COLUMN aircraft.code IS '飞机编号/呼号';
COMMENT ON COLUMN aircraft.name IS '飞机名称';
COMMENT ON COLUMN aircraft.status IS '状态，如 active';
COMMENT ON COLUMN aircraft.created_at IS '创建时间';
COMMENT ON COLUMN aircraft.updated_at IS '更新时间';
-- aircraft_availability
COMMENT ON COLUMN aircraft_availability.id IS '主键 UUID';
COMMENT ON COLUMN aircraft_availability.aircraft_id IS '关联飞机';
COMMENT ON COLUMN aircraft_availability.date IS '日期';
COMMENT ON COLUMN aircraft_availability.start_time IS '时段开始时间';
COMMENT ON COLUMN aircraft_availability.end_time IS '时段结束时间';
COMMENT ON COLUMN aircraft_availability.status IS 'available/maintenance/blocked';
COMMENT ON COLUMN aircraft_availability.created_at IS '创建时间';
COMMENT ON COLUMN aircraft_availability.updated_at IS '更新时间';
-- instructor_schedules
COMMENT ON COLUMN instructor_schedules.id IS '主键 UUID';
COMMENT ON COLUMN instructor_schedules.instructor_id IS '关联教员';
COMMENT ON COLUMN instructor_schedules.date IS '日期';
COMMENT ON COLUMN instructor_schedules.start_time IS '可飞开始时间';
COMMENT ON COLUMN instructor_schedules.end_time IS '可飞结束时间';
COMMENT ON COLUMN instructor_schedules.created_at IS '创建时间';
COMMENT ON COLUMN instructor_schedules.updated_at IS '更新时间';
-- training_plans
COMMENT ON COLUMN training_plans.id IS '主键 UUID';
COMMENT ON COLUMN training_plans.student_id IS '学员';
COMMENT ON COLUMN training_plans.instructor_id IS '教员';
COMMENT ON COLUMN training_plans.aircraft_id IS '飞机';
COMMENT ON COLUMN training_plans.plan_date IS '计划日期';
COMMENT ON COLUMN training_plans.start_time IS '开始时间';
COMMENT ON COLUMN training_plans.end_time IS '结束时间';
COMMENT ON COLUMN training_plans.source IS '来源：booking 预约 / manual 手工';
COMMENT ON COLUMN training_plans.created_at IS '创建时间';
COMMENT ON COLUMN training_plans.updated_at IS '更新时间';
-- bookings
COMMENT ON COLUMN bookings.id IS '主键 UUID';
COMMENT ON COLUMN bookings.student_id IS '预约学员';
COMMENT ON COLUMN bookings.instructor_id IS '预约教员';
COMMENT ON COLUMN bookings.aircraft_id IS '预约飞机';
COMMENT ON COLUMN bookings.plan_date IS '预约日期';
COMMENT ON COLUMN bookings.start_time IS '开始时间';
COMMENT ON COLUMN bookings.end_time IS '结束时间';
COMMENT ON COLUMN bookings.status IS 'pending/approved/rejected';
COMMENT ON COLUMN bookings.reject_reason IS '拒绝原因（拒绝时填写）';
COMMENT ON COLUMN bookings.reviewed_at IS '审批时间';
COMMENT ON COLUMN bookings.reviewer_id IS '审批人';
COMMENT ON COLUMN bookings.created_at IS '创建时间';
COMMENT ON COLUMN bookings.updated_at IS '更新时间';
-- training_records
COMMENT ON COLUMN training_records.id IS '主键 UUID';
COMMENT ON COLUMN training_records.student_id IS '学员';
COMMENT ON COLUMN training_records.instructor_id IS '教员';
COMMENT ON COLUMN training_records.aircraft_id IS '飞机';
COMMENT ON COLUMN training_records.training_date IS '训练日期';
COMMENT ON COLUMN training_records.engine_on_time IS '开车时间';
COMMENT ON COLUMN training_records.engine_off_time IS '关车时间';
COMMENT ON COLUMN training_records.flight_duration_minutes IS '飞行时长（分钟）';
COMMENT ON COLUMN training_records.is_local IS '是否本场';
COMMENT ON COLUMN training_records.import_source IS 'manual 手工 / import 导入';
COMMENT ON COLUMN training_records.raw_import_id IS '原始导入批次或记录标识';
COMMENT ON COLUMN training_records.status IS 'normal/cancelled/deleted/expired';
COMMENT ON COLUMN training_records.planned_duration_minutes IS '计划时长（分钟）';
COMMENT ON COLUMN training_records.theory_duration_minutes IS '理论时长（分钟）';
COMMENT ON COLUMN training_records.sim_duration_minutes IS '模拟机时长（分钟）';
COMMENT ON COLUMN training_records.syllabus_outline IS '大纲/阶段标识';
COMMENT ON COLUMN training_records.archive_display_no IS '档案显示编号';
COMMENT ON COLUMN training_records.departure IS '起飞地';
COMMENT ON COLUMN training_records.destination IS '目的地';
COMMENT ON COLUMN training_records.flight_type IS '飞行类型';
COMMENT ON COLUMN training_records.task_type IS '任务类型';
COMMENT ON COLUMN training_records.landings IS '起落次数';
COMMENT ON COLUMN training_records.created_at IS '创建时间';
COMMENT ON COLUMN training_records.updated_at IS '更新时间';
-- training_archives
COMMENT ON COLUMN training_archives.id IS '主键 UUID';
COMMENT ON COLUMN training_archives.training_record_id IS '关联训练记录，一对一';
COMMENT ON COLUMN training_archives.student_id IS '学员';
COMMENT ON COLUMN training_archives.instructor_id IS '教员';
COMMENT ON COLUMN training_archives.training_date IS '训练日期';
COMMENT ON COLUMN training_archives.stage IS '训练阶段';
COMMENT ON COLUMN training_archives.planned_duration_minutes IS '计划时长（分钟）';
COMMENT ON COLUMN training_archives.total_experience_minutes IS '累计经历时长（分钟）';
COMMENT ON COLUMN training_archives.remaining_minutes IS '剩余时长（分钟）';
COMMENT ON COLUMN training_archives.engine_on_time IS '开车时间';
COMMENT ON COLUMN training_archives.engine_off_time IS '关车时间';
COMMENT ON COLUMN training_archives.flight_duration_minutes IS '飞行时长（分钟）';
COMMENT ON COLUMN training_archives.progress_note IS '进度备注';
COMMENT ON COLUMN training_archives.instructor_review IS '教员讲评';
COMMENT ON COLUMN training_archives.student_self_review IS '学员自评';
COMMENT ON COLUMN training_archives.task_assignment IS '下次任务布置';
COMMENT ON COLUMN training_archives.student_signed_at IS '学员签字时间';
COMMENT ON COLUMN training_archives.responsible_instructor_signed_at IS '负责教员签字时间';
COMMENT ON COLUMN training_archives.chief_instructor_signed_at IS '主任教员签字时间';
COMMENT ON COLUMN training_archives.student_signer_id IS '学员签字人';
COMMENT ON COLUMN training_archives.responsible_instructor_signer_id IS '负责教员签字人';
COMMENT ON COLUMN training_archives.chief_instructor_signer_id IS '主任教员签字人';
COMMENT ON COLUMN training_archives.responsible_instructor_id IS '负责教员';
COMMENT ON COLUMN training_archives.chief_instructor_id IS '主任教员';
COMMENT ON COLUMN training_archives.created_at IS '创建时间';
COMMENT ON COLUMN training_archives.updated_at IS '更新时间';
-- syllabus_subjects
COMMENT ON COLUMN syllabus_subjects.id IS '主键 UUID';
COMMENT ON COLUMN syllabus_subjects.code IS '科目编码';
COMMENT ON COLUMN syllabus_subjects.name IS '科目名称';
COMMENT ON COLUMN syllabus_subjects.sort_order IS '排序序号';
COMMENT ON COLUMN syllabus_subjects.created_at IS '创建时间';
COMMENT ON COLUMN syllabus_subjects.updated_at IS '更新时间';
-- training_archive_subjects
COMMENT ON COLUMN training_archive_subjects.id IS '主键 UUID';
COMMENT ON COLUMN training_archive_subjects.training_archive_id IS '关联训练档案';
COMMENT ON COLUMN training_archive_subjects.syllabus_subject_id IS '关联大纲科目';
COMMENT ON COLUMN training_archive_subjects.evaluation_score IS '评估分数';
COMMENT ON COLUMN training_archive_subjects.is_graduated IS '该科目是否已结业';
COMMENT ON COLUMN training_archive_subjects.sort_order IS '排序序号';
COMMENT ON COLUMN training_archive_subjects.subject_content IS '科目内容/讲评摘要';
COMMENT ON COLUMN training_archive_subjects.type IS '科目类型';
COMMENT ON COLUMN training_archive_subjects.created_at IS '创建时间';
COMMENT ON COLUMN training_archive_subjects.updated_at IS '更新时间';
-- next_training_subjects
COMMENT ON COLUMN next_training_subjects.id IS '主键 UUID';
COMMENT ON COLUMN next_training_subjects.training_archive_id IS '关联训练档案';
COMMENT ON COLUMN next_training_subjects.syllabus_subject_id IS '下次要练的大纲科目';
COMMENT ON COLUMN next_training_subjects.assigned_by IS '指定人（用户 ID）';
COMMENT ON COLUMN next_training_subjects.created_at IS '创建时间';
-- raw_flight_imports
COMMENT ON COLUMN raw_flight_imports.id IS '主键 UUID';
COMMENT ON COLUMN raw_flight_imports.batch_id IS '导入批次 ID';
COMMENT ON COLUMN raw_flight_imports.engine_on_time IS '开车时间';
COMMENT ON COLUMN raw_flight_imports.engine_off_time IS '关车时间';
COMMENT ON COLUMN raw_flight_imports.flight_duration_minutes IS '飞行时长（分钟）';
COMMENT ON COLUMN raw_flight_imports.is_local IS '是否本场';
COMMENT ON COLUMN raw_flight_imports.raw_data IS '原始数据 JSON';
COMMENT ON COLUMN raw_flight_imports.training_record_id IS '关联到的训练记录（匹配后填入）';
COMMENT ON COLUMN raw_flight_imports.created_at IS '创建时间';
-- training_record_views
COMMENT ON COLUMN training_record_views.id IS '主键 UUID';
COMMENT ON COLUMN training_record_views.user_id IS '所属用户';
COMMENT ON COLUMN training_record_views.name IS '视图名称';
COMMENT ON COLUMN training_record_views.filter_json IS '筛选条件 JSON';
COMMENT ON COLUMN training_record_views.sort_json IS '排序配置 JSON';
COMMENT ON COLUMN training_record_views.group_by IS '分组字段';
COMMENT ON COLUMN training_record_views.created_at IS '创建时间';
COMMENT ON COLUMN training_record_views.updated_at IS '更新时间';
-- theory_documents
COMMENT ON COLUMN theory_documents.id IS '主键 UUID';
COMMENT ON COLUMN theory_documents.title IS '文档标题';
COMMENT ON COLUMN theory_documents.slug IS 'URL 友好标识，唯一';
COMMENT ON COLUMN theory_documents.sort_order IS '排序序号';
COMMENT ON COLUMN theory_documents.content_ref IS '内容引用/存储标识';
COMMENT ON COLUMN theory_documents.video_url IS '视频 URL';
COMMENT ON COLUMN theory_documents.category IS '分类';
COMMENT ON COLUMN theory_documents.created_at IS '创建时间';
COMMENT ON COLUMN theory_documents.updated_at IS '更新时间';
-- theory_questions
COMMENT ON COLUMN theory_questions.id IS '主键 UUID';
COMMENT ON COLUMN theory_questions.theory_document_id IS '所属理论文档';
COMMENT ON COLUMN theory_questions.question_text IS '题目内容';
COMMENT ON COLUMN theory_questions.options IS '选项 JSON，如 ["A","B","C"]';
COMMENT ON COLUMN theory_questions.correct_index IS '正确选项索引（从 0 起）';
COMMENT ON COLUMN theory_questions.sort_order IS '排序序号';
COMMENT ON COLUMN theory_questions.created_at IS '创建时间';
COMMENT ON COLUMN theory_questions.updated_at IS '更新时间';
-- theory_attempts
COMMENT ON COLUMN theory_attempts.id IS '主键 UUID';
COMMENT ON COLUMN theory_attempts.user_id IS '答题用户';
COMMENT ON COLUMN theory_attempts.theory_document_id IS '理论文档';
COMMENT ON COLUMN theory_attempts.answers IS '用户答案 JSON，如 [0,1,2]';
COMMENT ON COLUMN theory_attempts.score IS '得分';
COMMENT ON COLUMN theory_attempts.submitted_at IS '提交时间';
-- subject_learning
COMMENT ON COLUMN subject_learning.id IS '主键 UUID';
COMMENT ON COLUMN subject_learning.syllabus_subject_id IS '关联大纲科目，一对一';
COMMENT ON COLUMN subject_learning.title IS '学习标题';
COMMENT ON COLUMN subject_learning.content_ref IS '内容引用/存储标识';
COMMENT ON COLUMN subject_learning.video_url IS '视频 URL';
COMMENT ON COLUMN subject_learning.created_at IS '创建时间';
COMMENT ON COLUMN subject_learning.updated_at IS '更新时间';
-- subject_questions
COMMENT ON COLUMN subject_questions.id IS '主键 UUID';
COMMENT ON COLUMN subject_questions.subject_learning_id IS '所属科目学习';
COMMENT ON COLUMN subject_questions.question_text IS '题目内容';
COMMENT ON COLUMN subject_questions.options IS '选项 JSON';
COMMENT ON COLUMN subject_questions.correct_index IS '正确选项索引（从 0 起）';
COMMENT ON COLUMN subject_questions.sort_order IS '排序序号';
COMMENT ON COLUMN subject_questions.created_at IS '创建时间';
COMMENT ON COLUMN subject_questions.updated_at IS '更新时间';
-- subject_attempts
COMMENT ON COLUMN subject_attempts.id IS '主键 UUID';
COMMENT ON COLUMN subject_attempts.user_id IS '答题用户';
COMMENT ON COLUMN subject_attempts.subject_learning_id IS '科目学习';
COMMENT ON COLUMN subject_attempts.answers IS '用户答案 JSON';
COMMENT ON COLUMN subject_attempts.score IS '得分';
COMMENT ON COLUMN subject_attempts.submitted_at IS '提交时间';
-- tasks
COMMENT ON COLUMN tasks.id IS '主键 UUID';
COMMENT ON COLUMN tasks.student_id IS '任务所属学员';
COMMENT ON COLUMN tasks.assigner_id IS '布置人';
COMMENT ON COLUMN tasks.type IS 'theory 理论 / subject 科目 / other 其他';
COMMENT ON COLUMN tasks.ref_id IS '关联 ID（如理论文档或科目学习 ID）';
COMMENT ON COLUMN tasks.title IS '任务标题';
COMMENT ON COLUMN tasks.description IS '任务描述';
COMMENT ON COLUMN tasks.due_at IS '截止时间';
COMMENT ON COLUMN tasks.status IS 'pending/done/overdue';
COMMENT ON COLUMN tasks.created_at IS '创建时间';
COMMENT ON COLUMN tasks.updated_at IS '更新时间';
-- comments
COMMENT ON COLUMN comments.id IS '主键 UUID';
COMMENT ON COLUMN comments.target_type IS '评论对象类型：theory_document / subject_learning';
COMMENT ON COLUMN comments.target_id IS '评论对象 ID（理论文档或科目学习）';
COMMENT ON COLUMN comments.user_id IS '评论人';
COMMENT ON COLUMN comments.parent_id IS '父评论 ID，为空则为顶级评论';
COMMENT ON COLUMN comments.content IS '评论内容';
COMMENT ON COLUMN comments.created_at IS '创建时间';
COMMENT ON COLUMN comments.updated_at IS '更新时间';
-- bases
COMMENT ON COLUMN bases.id IS '主键 UUID';
COMMENT ON COLUMN bases.name IS '基地/机场名称';
COMMENT ON COLUMN bases.icao IS 'ICAO 四字码';
COMMENT ON COLUMN bases.city IS '城市';
COMMENT ON COLUMN bases.region IS '区域';
COMMENT ON COLUMN bases.created_at IS '创建时间';
COMMENT ON COLUMN bases.updated_at IS '更新时间';
-- day_operational_config
COMMENT ON COLUMN day_operational_config.id IS '主键 UUID';
COMMENT ON COLUMN day_operational_config.date IS '日期';
COMMENT ON COLUMN day_operational_config.base_id IS '关联基地，可空表示全局';
COMMENT ON COLUMN day_operational_config.weather IS '天气信息';
COMMENT ON COLUMN day_operational_config.airspace IS '空域信息';
COMMENT ON COLUMN day_operational_config.notes IS '备注';
COMMENT ON COLUMN day_operational_config.created_at IS '创建时间';
COMMENT ON COLUMN day_operational_config.updated_at IS '更新时间';
-- sms_verification_codes
COMMENT ON COLUMN sms_verification_codes.id IS '主键 UUID';
COMMENT ON COLUMN sms_verification_codes.phone IS '手机号';
COMMENT ON COLUMN sms_verification_codes.scene IS '场景：register/login/reset_pwd/2fa';
COMMENT ON COLUMN sms_verification_codes.code IS '验证码';
COMMENT ON COLUMN sms_verification_codes.ip IS '请求 IP';
COMMENT ON COLUMN sms_verification_codes.expired_at IS '过期时间';
COMMENT ON COLUMN sms_verification_codes.used_at IS '使用时间（验证通过时）';
COMMENT ON COLUMN sms_verification_codes.created_at IS '创建时间';

-- ========== 5. 索引 ==========
CREATE INDEX idx_students_user_id ON students(user_id);
CREATE INDEX idx_students_responsible_instructor ON students(responsible_instructor_id);
CREATE INDEX idx_aircraft_availability_aircraft_date ON aircraft_availability(aircraft_id, date);
CREATE INDEX idx_instructor_schedules_instructor_date ON instructor_schedules(instructor_id, date);
CREATE INDEX idx_training_plans_date ON training_plans(plan_date);
CREATE INDEX idx_bookings_student_status ON bookings(student_id, status);
CREATE INDEX idx_training_records_student_date ON training_records(student_id, training_date);
CREATE INDEX idx_training_records_status ON training_records(status);
CREATE INDEX idx_training_archives_student ON training_archives(student_id);
CREATE INDEX idx_training_archives_student_date ON training_archives(student_id, training_date);
CREATE INDEX idx_tasks_student_status ON tasks(student_id, status);
CREATE INDEX idx_theory_attempts_user_doc ON theory_attempts(user_id, theory_document_id);
CREATE INDEX idx_subject_attempts_user_subject ON subject_attempts(user_id, subject_learning_id);
CREATE INDEX idx_comments_target ON comments (target_type, target_id);
CREATE INDEX idx_comments_parent ON comments (parent_id);
CREATE INDEX idx_raw_flight_imports_batch ON raw_flight_imports (batch_id);
CREATE INDEX idx_training_record_views_user ON training_record_views (user_id);
CREATE UNIQUE INDEX idx_day_operational_config_date_base ON day_operational_config (date, COALESCE(base_id::text, ''));
CREATE INDEX idx_sms_verification_phone_scene_created_at ON sms_verification_codes(phone, scene, created_at DESC);
