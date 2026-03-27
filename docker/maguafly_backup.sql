--
-- PostgreSQL database dump
--

\restrict 1DanrRl7n90YVoUokwmY1sE56xm6r4cGPfABkgQPYhgtw7C2axc3qtco9QZS5de

-- Dumped from database version 15.17
-- Dumped by pg_dump version 15.17

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: availability_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.availability_status AS ENUM (
    'available',
    'maintenance',
    'blocked'
);


ALTER TYPE public.availability_status OWNER TO postgres;

--
-- Name: TYPE availability_status; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TYPE public.availability_status IS '航空器可用状态枚举：available（可用）、maintenance（维修）、blocked（占用/封锁）';


--
-- Name: booking_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.booking_status AS ENUM (
    'pending',
    'approved',
    'rejected'
);


ALTER TYPE public.booking_status OWNER TO postgres;

--
-- Name: TYPE booking_status; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TYPE public.booking_status IS '预约状态枚举：pending（待审核）、approved（已通过）、rejected（已拒绝）';


--
-- Name: comment_target_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.comment_target_type AS ENUM (
    'theory_document',
    'subject_learning'
);


ALTER TYPE public.comment_target_type OWNER TO postgres;

--
-- Name: TYPE comment_target_type; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TYPE public.comment_target_type IS '评论目标类型枚举：theory_document（理论文档）、subject_learning（科目学习内容）';


--
-- Name: import_source; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.import_source AS ENUM (
    'manual',
    'import'
);


ALTER TYPE public.import_source OWNER TO postgres;

--
-- Name: TYPE import_source; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TYPE public.import_source IS '记录来源枚举：manual（手工录入）、import（外部导入）';


--
-- Name: instructor_duty_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.instructor_duty_status AS ENUM (
    'morning',
    'afternoon',
    'full_day',
    'leave',
    'rest',
    'custom'
);


ALTER TYPE public.instructor_duty_status OWNER TO postgres;

--
-- Name: maintenance_window_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.maintenance_window_status AS ENUM (
    'planned',
    'in_progress',
    'completed',
    'cancelled'
);


ALTER TYPE public.maintenance_window_status OWNER TO postgres;

--
-- Name: plan_source; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.plan_source AS ENUM (
    'booking',
    'manual'
);


ALTER TYPE public.plan_source OWNER TO postgres;

--
-- Name: TYPE plan_source; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TYPE public.plan_source IS '训练计划来源枚举：booking（来源于学员预约）、manual（手工创建）';


--
-- Name: sms_scene; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.sms_scene AS ENUM (
    'register',
    'login',
    'reset_pwd',
    '2fa'
);


ALTER TYPE public.sms_scene OWNER TO postgres;

--
-- Name: task_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.task_status AS ENUM (
    'pending',
    'done',
    'overdue'
);


ALTER TYPE public.task_status OWNER TO postgres;

--
-- Name: TYPE task_status; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TYPE public.task_status IS '任务状态枚举：pending（待完成）、done（已完成）、overdue（已逾期）';


--
-- Name: task_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.task_type AS ENUM (
    'theory',
    'subject',
    'other',
    'flight'
);


ALTER TYPE public.task_type OWNER TO postgres;

--
-- Name: TYPE task_type; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TYPE public.task_type IS '任务类型枚举：theory（理论）、subject（科目）、other（其他）';


--
-- Name: training_record_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.training_record_status AS ENUM (
    'normal',
    'cancelled',
    'deleted',
    'expired'
);


ALTER TYPE public.training_record_status OWNER TO postgres;

--
-- Name: TYPE training_record_status; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TYPE public.training_record_status IS '训练记录状态：normal 正常、cancelled 取消、deleted 删除、expired 已过期';


--
-- Name: user_role; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.user_role AS ENUM (
    'admin',
    'staff',
    'instructor',
    'student',
    'maintenance'
);


ALTER TYPE public.user_role OWNER TO postgres;

--
-- Name: TYPE user_role; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TYPE public.user_role IS '用户角色枚举：admin（管理员）、staff（工作人员）、maintenance（机务）、instructor（教员）、student（学员）';


--
-- Name: community_groups_bump_member_count(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.community_groups_bump_member_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE community_groups SET member_count = member_count + 1, updated_at = NOW() WHERE id = NEW.group_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE community_groups SET member_count = GREATEST(0, member_count - 1), updated_at = NOW() WHERE id = OLD.group_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.community_groups_bump_member_count() OWNER TO postgres;

--
-- Name: community_posts_bump_comment_count(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.community_posts_bump_comment_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE community_posts SET comment_count = comment_count + 1, updated_at = NOW() WHERE id = NEW.post_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE community_posts SET comment_count = GREATEST(0, comment_count - 1), updated_at = NOW() WHERE id = OLD.post_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.community_posts_bump_comment_count() OWNER TO postgres;

--
-- Name: community_posts_bump_like_count(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.community_posts_bump_like_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE community_posts SET like_count = like_count + 1, updated_at = NOW() WHERE id = NEW.post_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE community_posts SET like_count = GREATEST(0, like_count - 1), updated_at = NOW() WHERE id = OLD.post_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.community_posts_bump_like_count() OWNER TO postgres;

--
-- Name: community_topics_bump_post_count(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.community_topics_bump_post_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE community_topics SET post_count = post_count + 1, updated_at = NOW() WHERE id = NEW.topic_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE community_topics SET post_count = GREATEST(0, post_count - 1), updated_at = NOW() WHERE id = OLD.topic_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.community_topics_bump_post_count() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: aircraft; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.aircraft (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code character varying(64) NOT NULL,
    name character varying(128),
    status character varying(32) DEFAULT 'active'::character varying,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    total_airframe_hours numeric(12,2),
    last_inspection_hours numeric(12,2),
    inspection_interval_hours integer DEFAULT 50 NOT NULL
);


ALTER TABLE public.aircraft OWNER TO postgres;

--
-- Name: TABLE aircraft; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.aircraft IS '飞机表：飞机编号、名称、状态（如 active）';


--
-- Name: COLUMN aircraft.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft.id IS '主键 UUID';


--
-- Name: COLUMN aircraft.code; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft.code IS '飞机编号/呼号';


--
-- Name: COLUMN aircraft.name; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft.name IS '飞机名称';


--
-- Name: COLUMN aircraft.status; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft.status IS '状态，如 active';


--
-- Name: COLUMN aircraft.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft.created_at IS '创建时间';


--
-- Name: COLUMN aircraft.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft.updated_at IS '更新时间';


--
-- Name: COLUMN aircraft.total_airframe_hours; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft.total_airframe_hours IS '机体/计时器总飞行小时（业务口径与 Hobbs 等一致）';


--
-- Name: COLUMN aircraft.last_inspection_hours; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft.last_inspection_hours IS '上次完成同级别定检时的总小时';


--
-- Name: COLUMN aircraft.inspection_interval_hours; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft.inspection_interval_hours IS '定检间隔小时数，默认 50';


--
-- Name: aircraft_availability; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.aircraft_availability (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    aircraft_id uuid NOT NULL,
    date date NOT NULL,
    start_time time without time zone NOT NULL,
    end_time time without time zone NOT NULL,
    status public.availability_status DEFAULT 'available'::public.availability_status NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    maintenance_window_id uuid
);


ALTER TABLE public.aircraft_availability OWNER TO postgres;

--
-- Name: TABLE aircraft_availability; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.aircraft_availability IS '飞机可用性：某日某时段飞机的可用/维护/占用状态';


--
-- Name: COLUMN aircraft_availability.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft_availability.id IS '主键 UUID';


--
-- Name: COLUMN aircraft_availability.aircraft_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft_availability.aircraft_id IS '关联飞机';


--
-- Name: COLUMN aircraft_availability.date; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft_availability.date IS '日期';


--
-- Name: COLUMN aircraft_availability.start_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft_availability.start_time IS '时段开始时间';


--
-- Name: COLUMN aircraft_availability.end_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft_availability.end_time IS '时段结束时间';


--
-- Name: COLUMN aircraft_availability.status; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft_availability.status IS 'available/maintenance/blocked';


--
-- Name: COLUMN aircraft_availability.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft_availability.created_at IS '创建时间';


--
-- Name: COLUMN aircraft_availability.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft_availability.updated_at IS '更新时间';


--
-- Name: COLUMN aircraft_availability.maintenance_window_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft_availability.maintenance_window_id IS '若由维护计划生成，关联维护窗口';


--
-- Name: aircraft_maintenance_windows; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.aircraft_maintenance_windows (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    aircraft_id uuid NOT NULL,
    kind character varying(64) DEFAULT 'periodic_inspection'::character varying NOT NULL,
    scheduled_start_at timestamp with time zone NOT NULL,
    scheduled_end_at timestamp with time zone NOT NULL,
    status public.maintenance_window_status DEFAULT 'planned'::public.maintenance_window_status NOT NULL,
    hangar_location character varying(255),
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.aircraft_maintenance_windows OWNER TO postgres;

--
-- Name: TABLE aircraft_maintenance_windows; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.aircraft_maintenance_windows IS '飞机定检/进机库等维护时间窗';


--
-- Name: COLUMN aircraft_maintenance_windows.kind; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.aircraft_maintenance_windows.kind IS '例如 periodic_inspection、annual 等';


--
-- Name: bases; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.bases (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    icao character varying(10),
    city character varying(100),
    region character varying(100),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.bases OWNER TO postgres;

--
-- Name: TABLE bases; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.bases IS '基地表：机场/基地名称、ICAO、城市、区域';


--
-- Name: COLUMN bases.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bases.id IS '主键 UUID';


--
-- Name: COLUMN bases.name; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bases.name IS '基地/机场名称';


--
-- Name: COLUMN bases.icao; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bases.icao IS 'ICAO 四字码';


--
-- Name: COLUMN bases.city; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bases.city IS '城市';


--
-- Name: COLUMN bases.region; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bases.region IS '区域';


--
-- Name: COLUMN bases.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bases.created_at IS '创建时间';


--
-- Name: COLUMN bases.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bases.updated_at IS '更新时间';


--
-- Name: bookings; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.bookings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    student_id uuid NOT NULL,
    instructor_id uuid NOT NULL,
    aircraft_id uuid NOT NULL,
    plan_date date NOT NULL,
    start_time time without time zone NOT NULL,
    end_time time without time zone NOT NULL,
    status public.booking_status DEFAULT 'pending'::public.booking_status NOT NULL,
    reject_reason text,
    reviewed_at timestamp with time zone,
    reviewer_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.bookings OWNER TO postgres;

--
-- Name: TABLE bookings; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.bookings IS '预约表：学员预约某日某时段，状态待审批/通过/拒绝，可填拒绝原因';


--
-- Name: COLUMN bookings.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bookings.id IS '主键 UUID';


--
-- Name: COLUMN bookings.student_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bookings.student_id IS '预约学员';


--
-- Name: COLUMN bookings.instructor_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bookings.instructor_id IS '预约教员';


--
-- Name: COLUMN bookings.aircraft_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bookings.aircraft_id IS '预约飞机';


--
-- Name: COLUMN bookings.plan_date; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bookings.plan_date IS '预约日期';


--
-- Name: COLUMN bookings.start_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bookings.start_time IS '开始时间';


--
-- Name: COLUMN bookings.end_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bookings.end_time IS '结束时间';


--
-- Name: COLUMN bookings.status; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bookings.status IS 'pending/approved/rejected';


--
-- Name: COLUMN bookings.reject_reason; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bookings.reject_reason IS '拒绝原因（拒绝时填写）';


--
-- Name: COLUMN bookings.reviewed_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bookings.reviewed_at IS '审批时间';


--
-- Name: COLUMN bookings.reviewer_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bookings.reviewer_id IS '审批人';


--
-- Name: COLUMN bookings.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bookings.created_at IS '创建时间';


--
-- Name: COLUMN bookings.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.bookings.updated_at IS '更新时间';


--
-- Name: comments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.comments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    target_type public.comment_target_type NOT NULL,
    target_id uuid NOT NULL,
    user_id uuid NOT NULL,
    parent_id uuid,
    content text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.comments OWNER TO postgres;

--
-- Name: TABLE comments; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.comments IS '评论表：对理论文档或科目学习的评论，支持父子级回复';


--
-- Name: COLUMN comments.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.comments.id IS '主键 UUID';


--
-- Name: COLUMN comments.target_type; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.comments.target_type IS '评论对象类型：theory_document / subject_learning';


--
-- Name: COLUMN comments.target_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.comments.target_id IS '评论对象 ID（理论文档或科目学习）';


--
-- Name: COLUMN comments.user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.comments.user_id IS '评论人';


--
-- Name: COLUMN comments.parent_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.comments.parent_id IS '父评论 ID，为空则为顶级评论';


--
-- Name: COLUMN comments.content; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.comments.content IS '评论内容';


--
-- Name: COLUMN comments.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.comments.created_at IS '创建时间';


--
-- Name: COLUMN comments.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.comments.updated_at IS '更新时间';


--
-- Name: community_challenge_entries; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.community_challenge_entries (
    challenge_id uuid NOT NULL,
    user_id uuid NOT NULL,
    post_id uuid NOT NULL,
    submitted_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.community_challenge_entries OWNER TO postgres;

--
-- Name: TABLE community_challenge_entries; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.community_challenge_entries IS '挑战赛作品（关联动态）';


--
-- Name: community_challenges; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.community_challenges (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title character varying(255) NOT NULL,
    description text,
    rules text,
    cover_url text,
    starts_at timestamp with time zone NOT NULL,
    ends_at timestamp with time zone NOT NULL,
    status character varying(16) DEFAULT 'draft'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT community_challenges_check CHECK ((ends_at >= starts_at)),
    CONSTRAINT community_challenges_status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'active'::character varying, 'ended'::character varying])::text[])))
);


ALTER TABLE public.community_challenges OWNER TO postgres;

--
-- Name: TABLE community_challenges; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.community_challenges IS '飞行挑战赛';


--
-- Name: community_comments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.community_comments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    post_id uuid NOT NULL,
    author_user_id uuid NOT NULL,
    body text NOT NULL,
    parent_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.community_comments OWNER TO postgres;

--
-- Name: TABLE community_comments; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.community_comments IS '动态评论';


--
-- Name: community_group_members; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.community_group_members (
    group_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role character varying(16) DEFAULT 'member'::character varying NOT NULL,
    joined_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT community_group_members_role_check CHECK (((role)::text = ANY ((ARRAY['member'::character varying, 'admin'::character varying])::text[])))
);


ALTER TABLE public.community_group_members OWNER TO postgres;

--
-- Name: TABLE community_group_members; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.community_group_members IS '小组成员';


--
-- Name: community_groups; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.community_groups (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(128) NOT NULL,
    description text,
    cover_url text,
    owner_user_id uuid NOT NULL,
    is_public boolean DEFAULT true NOT NULL,
    member_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.community_groups OWNER TO postgres;

--
-- Name: TABLE community_groups; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.community_groups IS '兴趣小组';


--
-- Name: community_post_likes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.community_post_likes (
    post_id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.community_post_likes OWNER TO postgres;

--
-- Name: TABLE community_post_likes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.community_post_likes IS '动态点赞';


--
-- Name: community_post_topics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.community_post_topics (
    post_id uuid NOT NULL,
    topic_id uuid NOT NULL
);


ALTER TABLE public.community_post_topics OWNER TO postgres;

--
-- Name: TABLE community_post_topics; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.community_post_topics IS '动态关联话题';


--
-- Name: community_posts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.community_posts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    author_user_id uuid NOT NULL,
    title character varying(500),
    body text NOT NULL,
    visibility character varying(32) DEFAULT 'public'::character varying NOT NULL,
    group_id uuid,
    status character varying(16) DEFAULT 'published'::character varying NOT NULL,
    media_urls jsonb DEFAULT '[]'::jsonb NOT NULL,
    like_count integer DEFAULT 0 NOT NULL,
    comment_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT community_posts_check CHECK ((((visibility)::text <> 'group_only'::text) OR (group_id IS NOT NULL))),
    CONSTRAINT community_posts_status_check CHECK (((status)::text = ANY ((ARRAY['published'::character varying, 'hidden'::character varying])::text[]))),
    CONSTRAINT community_posts_visibility_check CHECK (((visibility)::text = ANY ((ARRAY['public'::character varying, 'followers_only'::character varying, 'group_only'::character varying])::text[])))
);


ALTER TABLE public.community_posts OWNER TO postgres;

--
-- Name: TABLE community_posts; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.community_posts IS '社区动态/飞行故事';


--
-- Name: community_topics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.community_topics (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    slug character varying(128) NOT NULL,
    title character varying(255) NOT NULL,
    description text,
    cover_url text,
    post_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.community_topics OWNER TO postgres;

--
-- Name: TABLE community_topics; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.community_topics IS '社区话题';


--
-- Name: contracts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.contracts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    student_id uuid NOT NULL,
    contract_no character varying(64),
    sign_date date,
    start_date date,
    end_date date,
    total_minutes integer,
    price_total numeric(12,2),
    currency character varying(8),
    status character varying(32) DEFAULT 'active'::character varying NOT NULL,
    mongo_file_ids text[],
    remark text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.contracts OWNER TO postgres;

--
-- Name: TABLE contracts; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.contracts IS '学员合同表（结构化记录合同信息）';


--
-- Name: COLUMN contracts.student_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.contracts.student_id IS '关联学员 ID（students.id）';


--
-- Name: COLUMN contracts.contract_no; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.contracts.contract_no IS '合同编号';


--
-- Name: COLUMN contracts.sign_date; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.contracts.sign_date IS '合同签署日期';


--
-- Name: COLUMN contracts.start_date; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.contracts.start_date IS '合同生效日期';


--
-- Name: COLUMN contracts.end_date; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.contracts.end_date IS '合同结束日期';


--
-- Name: COLUMN contracts.total_minutes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.contracts.total_minutes IS '合同约定总训练时长（分钟）';


--
-- Name: COLUMN contracts.price_total; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.contracts.price_total IS '合同总金额';


--
-- Name: COLUMN contracts.currency; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.contracts.currency IS '币种，如 CNY/USD';


--
-- Name: COLUMN contracts.status; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.contracts.status IS '合同状态：active/finished/cancelled 等';


--
-- Name: COLUMN contracts.mongo_file_ids; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.contracts.mongo_file_ids IS '关联的合同扫描件/附件在 MongoDB 中的文件 ID 列表';


--
-- Name: day_operational_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.day_operational_config (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    date date NOT NULL,
    base_id uuid,
    weather character varying(500),
    airspace character varying(500),
    notes text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.day_operational_config OWNER TO postgres;

--
-- Name: TABLE day_operational_config; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.day_operational_config IS '每日运行配置：某日（可选某基地）的天气、空域、备注';


--
-- Name: COLUMN day_operational_config.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.day_operational_config.id IS '主键 UUID';


--
-- Name: COLUMN day_operational_config.date; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.day_operational_config.date IS '日期';


--
-- Name: COLUMN day_operational_config.base_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.day_operational_config.base_id IS '关联基地，可空表示全局';


--
-- Name: COLUMN day_operational_config.weather; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.day_operational_config.weather IS '天气信息';


--
-- Name: COLUMN day_operational_config.airspace; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.day_operational_config.airspace IS '空域信息';


--
-- Name: COLUMN day_operational_config.notes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.day_operational_config.notes IS '备注';


--
-- Name: COLUMN day_operational_config.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.day_operational_config.created_at IS '创建时间';


--
-- Name: COLUMN day_operational_config.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.day_operational_config.updated_at IS '更新时间';


--
-- Name: instructor_schedules; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.instructor_schedules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    instructor_id uuid NOT NULL,
    date date NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    duty_status public.instructor_duty_status DEFAULT 'full_day'::public.instructor_duty_status NOT NULL,
    remark text,
    duty_windows jsonb
);


ALTER TABLE public.instructor_schedules OWNER TO postgres;

--
-- Name: TABLE instructor_schedules; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.instructor_schedules IS '教员排班：某日某时段教员的可飞时段';


--
-- Name: COLUMN instructor_schedules.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructor_schedules.id IS '主键 UUID';


--
-- Name: COLUMN instructor_schedules.instructor_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructor_schedules.instructor_id IS '关联教员';


--
-- Name: COLUMN instructor_schedules.date; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructor_schedules.date IS '日期';


--
-- Name: COLUMN instructor_schedules.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructor_schedules.created_at IS '创建时间';


--
-- Name: COLUMN instructor_schedules.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructor_schedules.updated_at IS '更新时间';


--
-- Name: COLUMN instructor_schedules.duty_status; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructor_schedules.duty_status IS '出勤：morning/afternoon/full_day/leave/rest/custom（custom 表示显式 duty_windows）';


--
-- Name: COLUMN instructor_schedules.remark; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructor_schedules.remark IS '备注';


--
-- Name: COLUMN instructor_schedules.duty_windows; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructor_schedules.duty_windows IS '可选；null 时由 duty_status 与业务日 09:00–19:00、14:00 上下午分界推导；非 null 为显式窗口 [{"start":"HH:MM","end":"HH:MM"},...]，并裁剪到业务日内';


--
-- Name: instructors; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.instructors (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    name character varying(128) NOT NULL,
    display_color character varying(32),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.instructors OWNER TO postgres;

--
-- Name: TABLE instructors; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.instructors IS '教员表：关联用户，含姓名、显示颜色，用于排班与训练';


--
-- Name: COLUMN instructors.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructors.id IS '主键 UUID';


--
-- Name: COLUMN instructors.user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructors.user_id IS '关联用户表，一对一';


--
-- Name: COLUMN instructors.name; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructors.name IS '教员姓名';


--
-- Name: COLUMN instructors.display_color; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructors.display_color IS '日历/排班显示颜色';


--
-- Name: COLUMN instructors.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructors.created_at IS '创建时间';


--
-- Name: COLUMN instructors.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.instructors.updated_at IS '更新时间';


--
-- Name: next_training_subjects; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.next_training_subjects (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    training_archive_id uuid NOT NULL,
    syllabus_subject_id uuid NOT NULL,
    assigned_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.next_training_subjects OWNER TO postgres;

--
-- Name: TABLE next_training_subjects; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.next_training_subjects IS '下次训练科目：为某次训练档案指定的下次要训练的科目及指定人';


--
-- Name: COLUMN next_training_subjects.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.next_training_subjects.id IS '主键 UUID';


--
-- Name: COLUMN next_training_subjects.training_archive_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.next_training_subjects.training_archive_id IS '关联训练档案';


--
-- Name: COLUMN next_training_subjects.syllabus_subject_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.next_training_subjects.syllabus_subject_id IS '下次要练的大纲科目';


--
-- Name: COLUMN next_training_subjects.assigned_by; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.next_training_subjects.assigned_by IS '指定人（用户 ID）';


--
-- Name: COLUMN next_training_subjects.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.next_training_subjects.created_at IS '创建时间';


--
-- Name: raw_flight_imports; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.raw_flight_imports (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    batch_id character varying(64) NOT NULL,
    engine_on_time timestamp with time zone,
    engine_off_time timestamp with time zone,
    flight_duration_minutes integer,
    is_local boolean DEFAULT true,
    raw_data jsonb,
    training_record_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.raw_flight_imports OWNER TO postgres;

--
-- Name: TABLE raw_flight_imports; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.raw_flight_imports IS '原始飞行导入：批量导入的原始飞行数据，可关联到某条训练记录';


--
-- Name: COLUMN raw_flight_imports.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.raw_flight_imports.id IS '主键 UUID';


--
-- Name: COLUMN raw_flight_imports.batch_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.raw_flight_imports.batch_id IS '导入批次 ID';


--
-- Name: COLUMN raw_flight_imports.engine_on_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.raw_flight_imports.engine_on_time IS '开车时间';


--
-- Name: COLUMN raw_flight_imports.engine_off_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.raw_flight_imports.engine_off_time IS '关车时间';


--
-- Name: COLUMN raw_flight_imports.flight_duration_minutes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.raw_flight_imports.flight_duration_minutes IS '飞行时长（分钟）';


--
-- Name: COLUMN raw_flight_imports.is_local; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.raw_flight_imports.is_local IS '是否本场';


--
-- Name: COLUMN raw_flight_imports.raw_data; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.raw_flight_imports.raw_data IS '原始数据 JSON';


--
-- Name: COLUMN raw_flight_imports.training_record_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.raw_flight_imports.training_record_id IS '关联到的训练记录（匹配后填入）';


--
-- Name: COLUMN raw_flight_imports.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.raw_flight_imports.created_at IS '创建时间';


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO postgres;

--
-- Name: sms_verification_codes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sms_verification_codes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    phone character varying(32) NOT NULL,
    scene public.sms_scene NOT NULL,
    code character varying(16) NOT NULL,
    ip character varying(64),
    expired_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.sms_verification_codes OWNER TO postgres;

--
-- Name: TABLE sms_verification_codes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.sms_verification_codes IS '短信验证码：按手机号与场景存储验证码、过期时间、使用时间、IP';


--
-- Name: COLUMN sms_verification_codes.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.sms_verification_codes.id IS '主键 UUID';


--
-- Name: COLUMN sms_verification_codes.phone; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.sms_verification_codes.phone IS '手机号';


--
-- Name: COLUMN sms_verification_codes.scene; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.sms_verification_codes.scene IS '场景：register/login/reset_pwd/2fa';


--
-- Name: COLUMN sms_verification_codes.code; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.sms_verification_codes.code IS '验证码';


--
-- Name: COLUMN sms_verification_codes.ip; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.sms_verification_codes.ip IS '请求 IP';


--
-- Name: COLUMN sms_verification_codes.expired_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.sms_verification_codes.expired_at IS '过期时间';


--
-- Name: COLUMN sms_verification_codes.used_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.sms_verification_codes.used_at IS '使用时间（验证通过时）';


--
-- Name: COLUMN sms_verification_codes.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.sms_verification_codes.created_at IS '创建时间';


--
-- Name: staff_schedules; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.staff_schedules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    date date NOT NULL,
    duty_status public.instructor_duty_status DEFAULT 'full_day'::public.instructor_duty_status NOT NULL,
    remark text,
    duty_windows jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.staff_schedules OWNER TO postgres;

--
-- Name: TABLE staff_schedules; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.staff_schedules IS '工作人员日排班（考勤）：staff 账号，不可预约';


--
-- Name: COLUMN staff_schedules.user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.staff_schedules.user_id IS 'users.id，角色应为 staff';


--
-- Name: COLUMN staff_schedules.duty_status; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.staff_schedules.duty_status IS '与 instructor_schedules 相同枚举';


--
-- Name: COLUMN staff_schedules.duty_windows; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.staff_schedules.duty_windows IS '与 instructor_schedules.duty_windows 一致';


--
-- Name: students; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.students (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    name character varying(128) NOT NULL,
    student_no character varying(64) NOT NULL,
    contract_info jsonb,
    contract_duration_minutes integer,
    flown_minutes integer DEFAULT 0,
    training_progress text,
    training_progress_json jsonb,
    responsible_instructor_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.students OWNER TO postgres;

--
-- Name: TABLE students; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.students IS '学员表：关联用户，含学号、合同、已飞时长、训练进度、负责教员';


--
-- Name: COLUMN students.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.students.id IS '主键 UUID';


--
-- Name: COLUMN students.user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.students.user_id IS '关联用户表，可空（未绑定账号时）';


--
-- Name: COLUMN students.name; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.students.name IS '学员姓名';


--
-- Name: COLUMN students.student_no; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.students.student_no IS '学号，唯一';


--
-- Name: COLUMN students.contract_info; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.students.contract_info IS '合同信息 JSON';


--
-- Name: COLUMN students.contract_duration_minutes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.students.contract_duration_minutes IS '合同约定飞行时长（分钟）';


--
-- Name: COLUMN students.flown_minutes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.students.flown_minutes IS '已飞时长（分钟）';


--
-- Name: COLUMN students.training_progress; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.students.training_progress IS '训练进度文本（兼容旧版）';


--
-- Name: COLUMN students.training_progress_json; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.students.training_progress_json IS '训练进度结构化 JSON';


--
-- Name: COLUMN students.responsible_instructor_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.students.responsible_instructor_id IS '负责教员';


--
-- Name: COLUMN students.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.students.created_at IS '创建时间';


--
-- Name: COLUMN students.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.students.updated_at IS '更新时间';


--
-- Name: subject_attempts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.subject_attempts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    subject_learning_id uuid NOT NULL,
    answers jsonb NOT NULL,
    score integer NOT NULL,
    submitted_at timestamp with time zone DEFAULT now() NOT NULL,
    duration_seconds integer,
    passed boolean,
    max_score integer,
    correct_count integer,
    total_count integer
);


ALTER TABLE public.subject_attempts OWNER TO postgres;

--
-- Name: TABLE subject_attempts; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.subject_attempts IS '科目作答记录：用户对某科目学习的答题、得分与提交时间';


--
-- Name: COLUMN subject_attempts.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_attempts.id IS '主键 UUID';


--
-- Name: COLUMN subject_attempts.user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_attempts.user_id IS '答题用户';


--
-- Name: COLUMN subject_attempts.subject_learning_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_attempts.subject_learning_id IS '科目学习';


--
-- Name: COLUMN subject_attempts.answers; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_attempts.answers IS '用户答案 JSON';


--
-- Name: COLUMN subject_attempts.score; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_attempts.score IS '得分';


--
-- Name: COLUMN subject_attempts.submitted_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_attempts.submitted_at IS '提交时间';


--
-- Name: COLUMN subject_attempts.max_score; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_attempts.max_score IS '试卷满分（与 Mongo 题库一致）';


--
-- Name: COLUMN subject_attempts.correct_count; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_attempts.correct_count IS '答对题数';


--
-- Name: COLUMN subject_attempts.total_count; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_attempts.total_count IS '本次提交答题数';


--
-- Name: subject_learning; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.subject_learning (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    syllabus_subject_id uuid NOT NULL,
    title character varying(255) NOT NULL,
    content_ref character varying(64),
    video_url character varying(512),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.subject_learning OWNER TO postgres;

--
-- Name: TABLE subject_learning; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.subject_learning IS '科目学习：某大纲科目对应的学习材料（标题、内容引用、视频），一对一';


--
-- Name: COLUMN subject_learning.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_learning.id IS '主键 UUID';


--
-- Name: COLUMN subject_learning.syllabus_subject_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_learning.syllabus_subject_id IS '关联大纲科目，一对一';


--
-- Name: COLUMN subject_learning.title; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_learning.title IS '学习标题';


--
-- Name: COLUMN subject_learning.content_ref; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_learning.content_ref IS '内容引用/存储标识';


--
-- Name: COLUMN subject_learning.video_url; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_learning.video_url IS '视频 URL';


--
-- Name: COLUMN subject_learning.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_learning.created_at IS '创建时间';


--
-- Name: COLUMN subject_learning.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_learning.updated_at IS '更新时间';


--
-- Name: subject_questions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.subject_questions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    subject_learning_id uuid NOT NULL,
    question_text text NOT NULL,
    options jsonb NOT NULL,
    correct_index integer NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.subject_questions OWNER TO postgres;

--
-- Name: TABLE subject_questions; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.subject_questions IS '科目题目：某科目学习下的题目文本、选项、正确答案索引、排序';


--
-- Name: COLUMN subject_questions.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_questions.id IS '主键 UUID';


--
-- Name: COLUMN subject_questions.subject_learning_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_questions.subject_learning_id IS '所属科目学习';


--
-- Name: COLUMN subject_questions.question_text; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_questions.question_text IS '题目内容';


--
-- Name: COLUMN subject_questions.options; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_questions.options IS '选项 JSON';


--
-- Name: COLUMN subject_questions.correct_index; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_questions.correct_index IS '正确选项索引（从 0 起）';


--
-- Name: COLUMN subject_questions.sort_order; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_questions.sort_order IS '排序序号';


--
-- Name: COLUMN subject_questions.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_questions.created_at IS '创建时间';


--
-- Name: COLUMN subject_questions.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subject_questions.updated_at IS '更新时间';


--
-- Name: syllabus_outlines; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.syllabus_outlines (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    category character varying(32) NOT NULL,
    name character varying(255) NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT syllabus_outlines_category_check CHECK (((category)::text = ANY ((ARRAY['spl_local'::character varying, 'spl_transition'::character varying, 'spl_night'::character varying, 'ppl_basic'::character varying, 'ppl_night'::character varying])::text[])))
);


ALTER TABLE public.syllabus_outlines OWNER TO postgres;

--
-- Name: TABLE syllabus_outlines; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.syllabus_outlines IS '训练大纲实例：类别 + 名称，下挂一级科目';


--
-- Name: COLUMN syllabus_outlines.category; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.syllabus_outlines.category IS 'spl_local|SPL本场 spl_transition|SPL转场 spl_night|SPL夜航 ppl_basic|PPL基本 ppl_night|PPL夜航';


--
-- Name: syllabus_subject_secondaries; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.syllabus_subject_secondaries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    syllabus_subject_id uuid NOT NULL,
    seq_no integer DEFAULT 0 NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    completion_standard text,
    sort_order integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.syllabus_subject_secondaries OWNER TO postgres;

--
-- Name: TABLE syllabus_subject_secondaries; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.syllabus_subject_secondaries IS '一级科目下的二级科目：序号、名称、描述、完成标准、排序';


--
-- Name: syllabus_subjects; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.syllabus_subjects (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code character varying(64) NOT NULL,
    name character varying(255) NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    syllabus_outline_id uuid NOT NULL
);


ALTER TABLE public.syllabus_subjects OWNER TO postgres;

--
-- Name: TABLE syllabus_subjects; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.syllabus_subjects IS '大纲科目：训练大纲中的科目编码、名称与排序';


--
-- Name: COLUMN syllabus_subjects.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.syllabus_subjects.id IS '主键 UUID';


--
-- Name: COLUMN syllabus_subjects.code; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.syllabus_subjects.code IS '科目编码';


--
-- Name: COLUMN syllabus_subjects.name; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.syllabus_subjects.name IS '科目名称';


--
-- Name: COLUMN syllabus_subjects.sort_order; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.syllabus_subjects.sort_order IS '排序序号';


--
-- Name: COLUMN syllabus_subjects.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.syllabus_subjects.created_at IS '创建时间';


--
-- Name: COLUMN syllabus_subjects.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.syllabus_subjects.updated_at IS '更新时间';


--
-- Name: tasks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tasks (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    student_id uuid NOT NULL,
    assigner_id uuid NOT NULL,
    type public.task_type NOT NULL,
    ref_id uuid,
    title character varying(255) NOT NULL,
    description text,
    due_at timestamp with time zone,
    status public.task_status DEFAULT 'pending'::public.task_status NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.tasks OWNER TO postgres;

--
-- Name: TABLE tasks; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.tasks IS '任务表：分配给学员的理论/科目/其他任务，含截止时间与状态';


--
-- Name: COLUMN tasks.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.tasks.id IS '主键 UUID';


--
-- Name: COLUMN tasks.student_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.tasks.student_id IS '任务所属学员';


--
-- Name: COLUMN tasks.assigner_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.tasks.assigner_id IS '布置人';


--
-- Name: COLUMN tasks.type; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.tasks.type IS 'theory 理论 / subject 科目 / other 其他';


--
-- Name: COLUMN tasks.ref_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.tasks.ref_id IS '关联 ID（如理论文档或科目学习 ID）';


--
-- Name: COLUMN tasks.title; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.tasks.title IS '任务标题';


--
-- Name: COLUMN tasks.description; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.tasks.description IS '任务描述';


--
-- Name: COLUMN tasks.due_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.tasks.due_at IS '截止时间';


--
-- Name: COLUMN tasks.status; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.tasks.status IS 'pending/done/overdue';


--
-- Name: COLUMN tasks.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.tasks.created_at IS '创建时间';


--
-- Name: COLUMN tasks.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.tasks.updated_at IS '更新时间';


--
-- Name: theory_attempts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.theory_attempts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    theory_document_id uuid NOT NULL,
    answers jsonb NOT NULL,
    score integer NOT NULL,
    submitted_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.theory_attempts OWNER TO postgres;

--
-- Name: TABLE theory_attempts; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.theory_attempts IS '理论作答记录：用户对某理论文档的答题、得分与提交时间';


--
-- Name: COLUMN theory_attempts.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_attempts.id IS '主键 UUID';


--
-- Name: COLUMN theory_attempts.user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_attempts.user_id IS '答题用户';


--
-- Name: COLUMN theory_attempts.theory_document_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_attempts.theory_document_id IS '理论文档';


--
-- Name: COLUMN theory_attempts.answers; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_attempts.answers IS '用户答案 JSON，如 [0,1,2]';


--
-- Name: COLUMN theory_attempts.score; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_attempts.score IS '得分';


--
-- Name: COLUMN theory_attempts.submitted_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_attempts.submitted_at IS '提交时间';


--
-- Name: theory_documents; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.theory_documents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title character varying(255) NOT NULL,
    slug character varying(128),
    sort_order integer DEFAULT 0 NOT NULL,
    content_ref character varying(64),
    video_url character varying(512),
    category character varying(50),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    video_storage character varying(16),
    video_object_id character varying(256),
    video_content_type character varying(128),
    video_size_bytes bigint,
    video_duration_seconds integer,
    CONSTRAINT theory_documents_video_storage_ck CHECK (((video_storage IS NULL) OR (((video_storage)::text = 'external'::text) AND (NULLIF(btrim((COALESCE(video_url, ''::character varying))::text), ''::text) IS NOT NULL) AND (NULLIF(btrim((COALESCE(video_object_id, ''::character varying))::text), ''::text) IS NULL)) OR (((video_storage)::text = ANY ((ARRAY['gridfs'::character varying, 'cos'::character varying])::text[])) AND (NULLIF(btrim((COALESCE(video_object_id, ''::character varying))::text), ''::text) IS NOT NULL))))
);


ALTER TABLE public.theory_documents OWNER TO postgres;

--
-- Name: TABLE theory_documents; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.theory_documents IS '理论文档：理论课文档标题、slug、排序、内容引用、视频、分类';


--
-- Name: COLUMN theory_documents.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.id IS '主键 UUID';


--
-- Name: COLUMN theory_documents.title; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.title IS '文档标题';


--
-- Name: COLUMN theory_documents.slug; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.slug IS 'URL 友好标识，唯一';


--
-- Name: COLUMN theory_documents.sort_order; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.sort_order IS '排序序号';


--
-- Name: COLUMN theory_documents.content_ref; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.content_ref IS '内容引用/存储标识';


--
-- Name: COLUMN theory_documents.video_url; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.video_url IS '视频 URL';


--
-- Name: COLUMN theory_documents.category; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.category IS '分类';


--
-- Name: COLUMN theory_documents.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.created_at IS '创建时间';


--
-- Name: COLUMN theory_documents.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.updated_at IS '更新时间';


--
-- Name: COLUMN theory_documents.video_storage; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.video_storage IS '视频存储：external 外链 / gridfs / cos';


--
-- Name: COLUMN theory_documents.video_object_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.video_object_id IS '托管对象 ID（GridFS 文件 _id 或 COS object key）';


--
-- Name: COLUMN theory_documents.video_content_type; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.video_content_type IS '视频 MIME，如 video/mp4';


--
-- Name: COLUMN theory_documents.video_size_bytes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.video_size_bytes IS '视频字节大小';


--
-- Name: COLUMN theory_documents.video_duration_seconds; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_documents.video_duration_seconds IS '时长（秒），可后续异步填充';


--
-- Name: theory_paper_questions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.theory_paper_questions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    paper_id uuid NOT NULL,
    question_id character varying(255) NOT NULL,
    score integer NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.theory_paper_questions OWNER TO postgres;

--
-- Name: TABLE theory_paper_questions; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.theory_paper_questions IS '理论试卷题目关联：包含题目ID、分值与排序';


--
-- Name: COLUMN theory_paper_questions.paper_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_paper_questions.paper_id IS '所属试卷ID';


--
-- Name: COLUMN theory_paper_questions.question_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_paper_questions.question_id IS '题目ID（Mongo 题库中的 questions.id）';


--
-- Name: theory_papers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.theory_papers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    theory_document_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    total_score integer,
    enabled boolean DEFAULT true NOT NULL,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.theory_papers OWNER TO postgres;

--
-- Name: TABLE theory_papers; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.theory_papers IS '理论试卷：隶属于某个理论文档，可配置多套卷';


--
-- Name: COLUMN theory_papers.theory_document_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_papers.theory_document_id IS '关联的理论文档ID';


--
-- Name: COLUMN theory_papers.total_score; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_papers.total_score IS '试卷总分，来自所有题目分值之和';


--
-- Name: COLUMN theory_papers.enabled; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_papers.enabled IS '是否启用，对学员可见';


--
-- Name: theory_questions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.theory_questions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    theory_document_id uuid NOT NULL,
    question_text text NOT NULL,
    options jsonb NOT NULL,
    correct_index integer NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.theory_questions OWNER TO postgres;

--
-- Name: TABLE theory_questions; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.theory_questions IS '理论题目：某理论文档下的题目文本、选项、正确答案索引、排序';


--
-- Name: COLUMN theory_questions.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_questions.id IS '主键 UUID';


--
-- Name: COLUMN theory_questions.theory_document_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_questions.theory_document_id IS '所属理论文档';


--
-- Name: COLUMN theory_questions.question_text; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_questions.question_text IS '题目内容';


--
-- Name: COLUMN theory_questions.options; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_questions.options IS '选项 JSON，如 ["A","B","C"]';


--
-- Name: COLUMN theory_questions.correct_index; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_questions.correct_index IS '正确选项索引（从 0 起）';


--
-- Name: COLUMN theory_questions.sort_order; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_questions.sort_order IS '排序序号';


--
-- Name: COLUMN theory_questions.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_questions.created_at IS '创建时间';


--
-- Name: COLUMN theory_questions.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.theory_questions.updated_at IS '更新时间';


--
-- Name: training_archive_subjects; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.training_archive_subjects (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    training_archive_id uuid NOT NULL,
    syllabus_subject_id uuid NOT NULL,
    evaluation_score integer,
    is_graduated boolean DEFAULT false NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    subject_content text,
    type character varying(32),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.training_archive_subjects OWNER TO postgres;

--
-- Name: TABLE training_archive_subjects; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.training_archive_subjects IS '档案科目：某次训练档案下各大纲科目的评估分、是否结业、内容与类型';


--
-- Name: COLUMN training_archive_subjects.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archive_subjects.id IS '主键 UUID';


--
-- Name: COLUMN training_archive_subjects.training_archive_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archive_subjects.training_archive_id IS '关联训练档案';


--
-- Name: COLUMN training_archive_subjects.syllabus_subject_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archive_subjects.syllabus_subject_id IS '关联大纲科目';


--
-- Name: COLUMN training_archive_subjects.evaluation_score; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archive_subjects.evaluation_score IS '评估分数';


--
-- Name: COLUMN training_archive_subjects.is_graduated; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archive_subjects.is_graduated IS '该科目是否已结业';


--
-- Name: COLUMN training_archive_subjects.sort_order; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archive_subjects.sort_order IS '排序序号';


--
-- Name: COLUMN training_archive_subjects.subject_content; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archive_subjects.subject_content IS '科目内容/讲评摘要';


--
-- Name: COLUMN training_archive_subjects.type; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archive_subjects.type IS '科目类型';


--
-- Name: COLUMN training_archive_subjects.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archive_subjects.created_at IS '创建时间';


--
-- Name: COLUMN training_archive_subjects.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archive_subjects.updated_at IS '更新时间';


--
-- Name: training_archives; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.training_archives (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    training_record_id uuid NOT NULL,
    student_id uuid NOT NULL,
    instructor_id uuid NOT NULL,
    training_date date NOT NULL,
    stage character varying(64),
    planned_duration_minutes integer,
    total_experience_minutes integer,
    remaining_minutes integer,
    engine_on_time timestamp with time zone,
    engine_off_time timestamp with time zone,
    flight_duration_minutes integer,
    progress_note text,
    instructor_review text,
    student_self_review text,
    task_assignment text,
    student_signed_at timestamp with time zone,
    responsible_instructor_signed_at timestamp with time zone,
    chief_instructor_signed_at timestamp with time zone,
    student_signer_id uuid,
    responsible_instructor_signer_id uuid,
    chief_instructor_signer_id uuid,
    responsible_instructor_id uuid,
    chief_instructor_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.training_archives OWNER TO postgres;

--
-- Name: TABLE training_archives; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.training_archives IS '训练档案：与单次训练记录一一对应，含阶段、进度、签字、负责人与主任教员';


--
-- Name: COLUMN training_archives.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.id IS '主键 UUID';


--
-- Name: COLUMN training_archives.training_record_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.training_record_id IS '关联训练记录，一对一';


--
-- Name: COLUMN training_archives.student_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.student_id IS '学员';


--
-- Name: COLUMN training_archives.instructor_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.instructor_id IS '教员';


--
-- Name: COLUMN training_archives.training_date; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.training_date IS '训练日期';


--
-- Name: COLUMN training_archives.stage; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.stage IS '训练阶段';


--
-- Name: COLUMN training_archives.planned_duration_minutes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.planned_duration_minutes IS '计划时长（分钟）';


--
-- Name: COLUMN training_archives.total_experience_minutes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.total_experience_minutes IS '累计经历时长（分钟）';


--
-- Name: COLUMN training_archives.remaining_minutes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.remaining_minutes IS '剩余时长（分钟）';


--
-- Name: COLUMN training_archives.engine_on_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.engine_on_time IS '开车时间';


--
-- Name: COLUMN training_archives.engine_off_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.engine_off_time IS '关车时间';


--
-- Name: COLUMN training_archives.flight_duration_minutes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.flight_duration_minutes IS '飞行时长（分钟）';


--
-- Name: COLUMN training_archives.progress_note; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.progress_note IS '进度备注';


--
-- Name: COLUMN training_archives.instructor_review; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.instructor_review IS '教员讲评';


--
-- Name: COLUMN training_archives.student_self_review; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.student_self_review IS '学员自评';


--
-- Name: COLUMN training_archives.task_assignment; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.task_assignment IS '下次任务布置';


--
-- Name: COLUMN training_archives.student_signed_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.student_signed_at IS '学员签字时间';


--
-- Name: COLUMN training_archives.responsible_instructor_signed_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.responsible_instructor_signed_at IS '负责教员签字时间';


--
-- Name: COLUMN training_archives.chief_instructor_signed_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.chief_instructor_signed_at IS '主任教员签字时间';


--
-- Name: COLUMN training_archives.student_signer_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.student_signer_id IS '学员签字人';


--
-- Name: COLUMN training_archives.responsible_instructor_signer_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.responsible_instructor_signer_id IS '负责教员签字人';


--
-- Name: COLUMN training_archives.chief_instructor_signer_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.chief_instructor_signer_id IS '主任教员签字人';


--
-- Name: COLUMN training_archives.responsible_instructor_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.responsible_instructor_id IS '负责教员';


--
-- Name: COLUMN training_archives.chief_instructor_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.chief_instructor_id IS '主任教员';


--
-- Name: COLUMN training_archives.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.created_at IS '创建时间';


--
-- Name: COLUMN training_archives.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_archives.updated_at IS '更新时间';


--
-- Name: training_plans; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.training_plans (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    student_id uuid NOT NULL,
    instructor_id uuid NOT NULL,
    aircraft_id uuid NOT NULL,
    plan_date date NOT NULL,
    start_time time without time zone NOT NULL,
    end_time time without time zone NOT NULL,
    source public.plan_source DEFAULT 'manual'::public.plan_source NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.training_plans OWNER TO postgres;

--
-- Name: TABLE training_plans; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.training_plans IS '训练计划：计划某日某时段学员-教员-飞机的安排，来源可为预约或手工';


--
-- Name: COLUMN training_plans.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_plans.id IS '主键 UUID';


--
-- Name: COLUMN training_plans.student_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_plans.student_id IS '学员';


--
-- Name: COLUMN training_plans.instructor_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_plans.instructor_id IS '教员';


--
-- Name: COLUMN training_plans.aircraft_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_plans.aircraft_id IS '飞机';


--
-- Name: COLUMN training_plans.plan_date; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_plans.plan_date IS '计划日期';


--
-- Name: COLUMN training_plans.start_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_plans.start_time IS '开始时间';


--
-- Name: COLUMN training_plans.end_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_plans.end_time IS '结束时间';


--
-- Name: COLUMN training_plans.source; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_plans.source IS '来源：booking 预约 / manual 手工';


--
-- Name: COLUMN training_plans.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_plans.created_at IS '创建时间';


--
-- Name: COLUMN training_plans.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_plans.updated_at IS '更新时间';


--
-- Name: training_record_views; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.training_record_views (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    name character varying(128) NOT NULL,
    filter_json jsonb,
    sort_json jsonb,
    group_by character varying(64),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.training_record_views OWNER TO postgres;

--
-- Name: TABLE training_record_views; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.training_record_views IS '训练记录视图：用户保存的筛选、排序、分组配置';


--
-- Name: COLUMN training_record_views.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_record_views.id IS '主键 UUID';


--
-- Name: COLUMN training_record_views.user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_record_views.user_id IS '所属用户';


--
-- Name: COLUMN training_record_views.name; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_record_views.name IS '视图名称';


--
-- Name: COLUMN training_record_views.filter_json; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_record_views.filter_json IS '筛选条件 JSON';


--
-- Name: COLUMN training_record_views.sort_json; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_record_views.sort_json IS '排序配置 JSON';


--
-- Name: COLUMN training_record_views.group_by; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_record_views.group_by IS '分组字段';


--
-- Name: COLUMN training_record_views.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_record_views.created_at IS '创建时间';


--
-- Name: COLUMN training_record_views.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_record_views.updated_at IS '更新时间';


--
-- Name: training_records; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.training_records (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    student_id uuid NOT NULL,
    instructor_id uuid NOT NULL,
    aircraft_id uuid NOT NULL,
    training_date date NOT NULL,
    engine_on_time timestamp with time zone,
    engine_off_time timestamp with time zone,
    flight_duration_minutes integer,
    is_local boolean DEFAULT true,
    import_source public.import_source DEFAULT 'manual'::public.import_source NOT NULL,
    raw_import_id character varying(64),
    status public.training_record_status DEFAULT 'normal'::public.training_record_status NOT NULL,
    planned_duration_minutes integer,
    theory_duration_minutes integer,
    sim_duration_minutes integer,
    syllabus_outline character varying(64),
    archive_display_no character varying(64),
    departure character varying(100),
    destination character varying(100),
    flight_type character varying(20),
    task_type character varying(50),
    landings integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.training_records OWNER TO postgres;

--
-- Name: TABLE training_records; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.training_records IS '飞行训练记录：单次飞行的学员/教员/飞机、起落时间、时长、状态、导入来源等';


--
-- Name: COLUMN training_records.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.id IS '主键 UUID';


--
-- Name: COLUMN training_records.student_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.student_id IS '学员';


--
-- Name: COLUMN training_records.instructor_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.instructor_id IS '教员';


--
-- Name: COLUMN training_records.aircraft_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.aircraft_id IS '飞机';


--
-- Name: COLUMN training_records.training_date; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.training_date IS '训练日期';


--
-- Name: COLUMN training_records.engine_on_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.engine_on_time IS '开车时间';


--
-- Name: COLUMN training_records.engine_off_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.engine_off_time IS '关车时间';


--
-- Name: COLUMN training_records.flight_duration_minutes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.flight_duration_minutes IS '飞行时长（分钟）';


--
-- Name: COLUMN training_records.is_local; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.is_local IS '是否本场';


--
-- Name: COLUMN training_records.import_source; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.import_source IS 'manual 手工 / import 导入';


--
-- Name: COLUMN training_records.raw_import_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.raw_import_id IS '原始导入批次或记录标识';


--
-- Name: COLUMN training_records.status; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.status IS 'normal/cancelled/deleted/expired';


--
-- Name: COLUMN training_records.planned_duration_minutes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.planned_duration_minutes IS '计划时长（分钟）';


--
-- Name: COLUMN training_records.theory_duration_minutes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.theory_duration_minutes IS '理论时长（分钟）';


--
-- Name: COLUMN training_records.sim_duration_minutes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.sim_duration_minutes IS '模拟机时长（分钟）';


--
-- Name: COLUMN training_records.syllabus_outline; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.syllabus_outline IS '大纲/阶段标识';


--
-- Name: COLUMN training_records.archive_display_no; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.archive_display_no IS '档案显示编号';


--
-- Name: COLUMN training_records.departure; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.departure IS '起飞地';


--
-- Name: COLUMN training_records.destination; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.destination IS '目的地';


--
-- Name: COLUMN training_records.flight_type; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.flight_type IS '飞行类型';


--
-- Name: COLUMN training_records.task_type; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.task_type IS '任务类型';


--
-- Name: COLUMN training_records.landings; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.landings IS '起落次数';


--
-- Name: COLUMN training_records.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.created_at IS '创建时间';


--
-- Name: COLUMN training_records.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.training_records.updated_at IS '更新时间';


--
-- Name: user_follows; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_follows (
    follower_id uuid NOT NULL,
    followee_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT user_follows_check CHECK ((follower_id <> followee_id))
);


ALTER TABLE public.user_follows OWNER TO postgres;

--
-- Name: TABLE user_follows; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.user_follows IS '用户关注关系';


--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    username character varying(64) NOT NULL,
    password_hash character varying(255) NOT NULL,
    phone character varying(32),
    email character varying(255),
    role public.user_role NOT NULL,
    avatar_url character varying(512),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: TABLE users; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.users IS '用户表：系统账号，含用户名、密码、手机、邮箱、角色、头像等';


--
-- Name: COLUMN users.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.id IS '主键 UUID';


--
-- Name: COLUMN users.username; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.username IS '登录用户名，唯一';


--
-- Name: COLUMN users.password_hash; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.password_hash IS '密码哈希';


--
-- Name: COLUMN users.phone; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.phone IS '手机号';


--
-- Name: COLUMN users.email; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.email IS '邮箱';


--
-- Name: COLUMN users.role; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.role IS '角色：admin/staff/instructor/student';


--
-- Name: COLUMN users.avatar_url; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.avatar_url IS '头像 URL';


--
-- Name: COLUMN users.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.created_at IS '创建时间';


--
-- Name: COLUMN users.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.updated_at IS '更新时间';


--
-- Data for Name: aircraft; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.aircraft (id, code, name, status, created_at, updated_at, total_airframe_hours, last_inspection_hours, inspection_interval_hours) FROM stdin;
b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	AC002	飞书1	active	2026-03-20 08:58:30.065324+00	2026-03-20 08:58:30.065324+00	\N	\N	50
eb5bd27f-009f-4267-b7c8-68a41ae77a23	AC001	塞斯纳 172	active	2026-03-16 08:47:56.919174+00	2026-03-20 11:32:59.464765+00	121.00	100.00	50
\.


--
-- Data for Name: aircraft_availability; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.aircraft_availability (id, aircraft_id, date, start_time, end_time, status, created_at, updated_at, maintenance_window_id) FROM stdin;
3cc130f1-f709-4c8f-9a8b-2e025be16b92	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-20	09:00:00	17:00:00	available	2026-03-20 10:54:58.263933+00	2026-03-20 10:54:58.263933+00	\N
b5258e89-761e-422d-ace2-79c76c99d840	b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	2026-03-20	09:00:00	17:00:00	available	2026-03-20 11:41:06.071893+00	2026-03-20 11:41:06.071893+00	\N
840e70f3-24a6-43d0-9de4-1598155c00b4	b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	2026-03-23	09:00:00	17:00:00	available	2026-03-22 13:08:46.650523+00	2026-03-22 13:08:46.650523+00	\N
30ec2c31-9b60-4f34-8630-42d66c799036	b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	2026-03-23	12:00:00	18:00:00	maintenance	2026-03-23 13:18:40.335027+00	2026-03-23 13:18:40.335027+00	ed0af527-ce30-42f1-baee-23223eac752a
\.


--
-- Data for Name: aircraft_maintenance_windows; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.aircraft_maintenance_windows (id, aircraft_id, kind, scheduled_start_at, scheduled_end_at, status, hangar_location, notes, created_at, updated_at) FROM stdin;
07d0390f-b09d-4b6d-9b48-c256bf4bfb85	eb5bd27f-009f-4267-b7c8-68a41ae77a23	periodic_inspection	2026-03-21 00:00:00+00	2026-03-21 10:00:00+00	cancelled	\N	\N	2026-03-20 11:35:42.851815+00	2026-03-20 11:38:23.359949+00
ed0af527-ce30-42f1-baee-23223eac752a	b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	manual	2026-03-23 04:00:00+00	2026-03-23 10:00:00+00	planned	\N	部分时段维护	2026-03-23 13:18:40.335027+00	2026-03-23 13:18:40.335027+00
\.


--
-- Data for Name: bases; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.bases (id, name, icao, city, region, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: bookings; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.bookings (id, student_id, instructor_id, aircraft_id, plan_date, start_time, end_time, status, reject_reason, reviewed_at, reviewer_id, created_at, updated_at) FROM stdin;
5a4abbc1-c9f0-4e02-ab56-98a70eb753fe	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-18	10:00:00	11:30:00	approved	可选的审核说明	2026-03-17 02:16:58.551984+00	c13764be-834f-4634-bbac-c1442a1a5532	2026-03-17 02:12:35.9865+00	2026-03-17 02:16:58.551984+00
b72e8a54-8744-46ab-8292-38627b8bd202	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-18	12:00:00	13:30:00	approved	同意	2026-03-17 02:35:35.415466+00	c13764be-834f-4634-bbac-c1442a1a5532	2026-03-17 02:35:09.404589+00	2026-03-17 02:35:35.415466+00
009237cb-6c91-463e-93f0-0ffa5f0d84bc	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-18	12:00:00	13:30:00	approved	同意	2026-03-17 12:45:09.894211+00	16c9ed24-bd34-4e70-a681-303903910781	2026-03-17 12:44:04.08566+00	2026-03-17 12:45:09.894211+00
91b6e3f0-3785-446b-9482-604489f29ab0	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-19	12:00:00	13:30:00	approved	同意	2026-03-17 12:54:14.529949+00	16c9ed24-bd34-4e70-a681-303903910781	2026-03-17 12:53:45.739464+00	2026-03-17 12:54:14.529949+00
06ecd346-8eed-4579-9ac4-6858b8691320	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-19	10:00:00	11:30:00	approved	同意	2026-03-17 12:57:09.175191+00	16c9ed24-bd34-4e70-a681-303903910781	2026-03-17 12:56:54.384607+00	2026-03-17 12:57:09.175191+00
14a45a1a-bec2-4c6e-bc0c-6f0bb9e5ef39	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	2026-03-24	09:00:00	10:00:00	pending	\N	\N	\N	2026-03-23 07:24:05.944482+00	2026-03-23 07:24:05.944482+00
28ff1188-7f7e-4856-b56f-7a48fdd19adb	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	2026-03-23	09:00:00	10:00:00	pending	\N	\N	\N	2026-03-23 07:31:52.06057+00	2026-03-23 07:31:52.06057+00
9e7cb64f-d5e5-4316-916f-26bf95c085c2	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	2026-03-23	09:00:00	10:15:00	pending	\N	\N	\N	2026-03-23 08:15:28.800378+00	2026-03-23 08:15:28.800378+00
827e06d9-e140-4c48-98bf-d1d7f8419516	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-23	10:15:00	11:15:00	pending	\N	\N	\N	2026-03-23 08:16:04.658395+00	2026-03-23 08:16:04.658395+00
fbef4e10-5159-4e4b-aa35-eec316323f88	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-23	12:15:00	13:15:00	pending	\N	\N	\N	2026-03-23 08:16:19.16589+00	2026-03-23 08:16:19.16589+00
b05ad25d-b792-4b5f-ba39-9e737231e918	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	2026-03-23	10:45:00	12:15:00	rejected	飞机维护/定检，时段已取消	2026-03-23 13:18:40.335027+00	8d56b96f-d5da-4016-803a-638715c2b515	2026-03-23 08:01:11.0915+00	2026-03-23 13:18:40.335027+00
ac4ea82b-7028-40fe-b164-fb05b1075713	05626980-d27f-4248-8072-c83e7569b3d8	34a5aa64-d7dd-493d-bed3-c45ff961dcf4	b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	2026-03-23	14:15:00	15:15:00	rejected	飞机维护/定检，时段已取消	2026-03-23 13:18:40.335027+00	8d56b96f-d5da-4016-803a-638715c2b515	2026-03-23 08:15:52.599125+00	2026-03-23 13:18:40.335027+00
ef9dd4f9-7b71-47f0-9018-e433c2d62579	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	2026-03-26	09:15:00	10:15:00	approved		2026-03-25 08:25:53.164064+00	16c9ed24-bd34-4e70-a681-303903910781	2026-03-25 07:53:09.752102+00	2026-03-25 08:25:53.164064+00
92ac06ad-72ba-4512-9c19-38799fa0f203	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	2026-03-27	09:15:00	10:15:00	pending	\N	\N	\N	2026-03-25 08:29:43.835049+00	2026-03-25 08:29:43.835049+00
\.


--
-- Data for Name: comments; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.comments (id, target_type, target_id, user_id, parent_id, content, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: community_challenge_entries; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.community_challenge_entries (challenge_id, user_id, post_id, submitted_at) FROM stdin;
\.


--
-- Data for Name: community_challenges; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.community_challenges (id, title, description, rules, cover_url, starts_at, ends_at, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: community_comments; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.community_comments (id, post_id, author_user_id, body, parent_id, created_at) FROM stdin;
b1e3ae88-d441-40aa-8921-e962b003b81a	6b174fbe-50bb-4489-a8a6-ba6bf889cfb0	c13764be-834f-4634-bbac-c1442a1a5532	22	\N	2026-03-26 07:18:33.70266+00
\.


--
-- Data for Name: community_group_members; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.community_group_members (group_id, user_id, role, joined_at) FROM stdin;
943db4b9-30d8-462f-8804-63f1872a5240	c13764be-834f-4634-bbac-c1442a1a5532	admin	2026-03-26 07:10:31.957537+00
\.


--
-- Data for Name: community_groups; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.community_groups (id, name, slug, description, cover_url, owner_user_id, is_public, member_count, created_at, updated_at) FROM stdin;
943db4b9-30d8-462f-8804-63f1872a5240	GrouP2	slug0ee	desc	http://	c13764be-834f-4634-bbac-c1442a1a5532	t	1	2026-03-26 07:10:31.949004+00	2026-03-26 07:10:31.957537+00
\.


--
-- Data for Name: community_post_likes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.community_post_likes (post_id, user_id, created_at) FROM stdin;
\.


--
-- Data for Name: community_post_topics; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.community_post_topics (post_id, topic_id) FROM stdin;
\.


--
-- Data for Name: community_posts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.community_posts (id, author_user_id, title, body, visibility, group_id, status, media_urls, like_count, comment_count, created_at, updated_at) FROM stdin;
b88cf6c4-b32a-4098-a44c-93927e52fa40	8d56b96f-d5da-4016-803a-638715c2b515	1	2	public	\N	published	["333"]	0	0	2026-03-26 07:09:00.950094+00	2026-03-26 07:09:00.950094+00
95bc0480-5ca6-4e24-b023-39cbc1c4d91e	8d56b96f-d5da-4016-803a-638715c2b515	1	2	public	\N	published	["333"]	0	0	2026-03-26 07:09:06.033327+00	2026-03-26 07:09:06.033327+00
590fbccf-ca45-4ff7-950e-d72c5d046a5d	c13764be-834f-4634-bbac-c1442a1a5532	学员的故事	咳	public	\N	published	["222"]	0	0	2026-03-26 07:10:02.250111+00	2026-03-26 07:10:02.250111+00
f193debb-1c52-4352-b0af-557f4bd09430	c13764be-834f-4634-bbac-c1442a1a5532	11	222	public	\N	published	["333"]	0	0	2026-03-26 07:18:19.311408+00	2026-03-26 07:18:19.311408+00
6b174fbe-50bb-4489-a8a6-ba6bf889cfb0	c13764be-834f-4634-bbac-c1442a1a5532	Go	Fly over	public	\N	published	["http://djj"]	0	1	2026-03-26 07:15:37.743071+00	2026-03-26 07:18:33.70266+00
\.


--
-- Data for Name: community_topics; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.community_topics (id, slug, title, description, cover_url, post_count, created_at, updated_at) FROM stdin;
ecf54691-02da-4fa9-9180-46acdb204c79	slug	话题1	这里是描述	\N	0	2026-03-26 07:08:36.034479+00	2026-03-26 07:08:36.034479+00
\.


--
-- Data for Name: contracts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.contracts (id, student_id, contract_no, sign_date, start_date, end_date, total_minutes, price_total, currency, status, mongo_file_ids, remark, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: day_operational_config; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.day_operational_config (id, date, base_id, weather, airspace, notes, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: instructor_schedules; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.instructor_schedules (id, instructor_id, date, created_at, updated_at, duty_status, remark, duty_windows) FROM stdin;
fbdf0faa-32ba-4e40-b102-104068c274c1	878dda87-3aeb-4fa8-819b-7e3ec65d8212	2026-03-30	2026-03-22 12:00:52.491909+00	2026-03-22 12:00:52.491909+00	full_day	\N	\N
3ed63f20-9c3d-4cdd-a201-a9d0a7552eca	878dda87-3aeb-4fa8-819b-7e3ec65d8212	2026-03-31	2026-03-22 12:00:52.492417+00	2026-03-22 12:00:52.492417+00	full_day	\N	\N
e06fbea6-aa6a-47b0-b387-392aa0e8bc4f	34a5aa64-d7dd-493d-bed3-c45ff961dcf4	2026-03-23	2026-03-22 12:02:50.801872+00	2026-03-22 12:02:50.801872+00	afternoon	\N	\N
5fb911ad-8613-40d4-829a-d659f09f79bc	34a5aa64-d7dd-493d-bed3-c45ff961dcf4	2026-03-24	2026-03-22 12:02:50.814021+00	2026-03-22 12:02:50.814021+00	afternoon	\N	\N
36919f3d-9f6f-47ae-87e1-5edcd7a76f50	34a5aa64-d7dd-493d-bed3-c45ff961dcf4	2026-03-25	2026-03-22 12:02:50.821412+00	2026-03-22 12:02:50.821412+00	afternoon	\N	\N
5a9d10bb-7ed2-4771-b172-69f8da895c0f	34a5aa64-d7dd-493d-bed3-c45ff961dcf4	2026-03-26	2026-03-22 12:02:50.823957+00	2026-03-22 12:02:50.823957+00	afternoon	\N	\N
0793fba7-8cc9-4329-935c-71df9c4d1076	34a5aa64-d7dd-493d-bed3-c45ff961dcf4	2026-03-27	2026-03-22 12:02:50.828634+00	2026-03-22 12:02:50.828634+00	afternoon	\N	\N
4894d870-ce93-467e-8b72-9d0e50867b37	34a5aa64-d7dd-493d-bed3-c45ff961dcf4	2026-03-30	2026-03-22 12:02:50.830261+00	2026-03-22 12:02:50.830261+00	afternoon	\N	\N
2abe5f35-c691-4141-9003-7174dd288060	34a5aa64-d7dd-493d-bed3-c45ff961dcf4	2026-03-31	2026-03-22 12:02:50.832368+00	2026-03-22 12:02:50.832368+00	afternoon	\N	\N
e13715a5-3720-486e-aeee-1af6ba8178e8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	2026-03-23	2026-03-22 11:58:33.444347+00	2026-03-22 13:04:45.538247+00	full_day	\N	\N
ee37346a-a68b-40e6-aa9d-768118b2b538	878dda87-3aeb-4fa8-819b-7e3ec65d8212	2026-03-24	2026-03-22 12:00:52.488124+00	2026-03-22 13:04:45.547578+00	full_day	\N	\N
1c56df23-09c4-4b44-8bd3-bd2f3fc533be	878dda87-3aeb-4fa8-819b-7e3ec65d8212	2026-03-25	2026-03-22 12:00:52.489889+00	2026-03-22 13:04:45.548523+00	full_day	\N	\N
06048b3c-92f5-4370-a554-dcb88c9c073a	878dda87-3aeb-4fa8-819b-7e3ec65d8212	2026-03-26	2026-03-22 12:00:52.490661+00	2026-03-22 13:04:45.549152+00	full_day	\N	\N
1ef0313a-c870-46f5-875c-8e2b03a8008a	878dda87-3aeb-4fa8-819b-7e3ec65d8212	2026-03-27	2026-03-22 12:00:52.491259+00	2026-03-22 13:04:45.549812+00	full_day	\N	\N
\.


--
-- Data for Name: instructors; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.instructors (id, user_id, name, display_color, created_at, updated_at) FROM stdin;
878dda87-3aeb-4fa8-819b-7e3ec65d8212	16c9ed24-bd34-4e70-a681-303903910781	instructor_vicent		2026-03-15 14:31:49.607747+00	2026-03-15 14:31:49.607747+00
34a5aa64-d7dd-493d-bed3-c45ff961dcf4	242f6832-a633-48d3-87fd-6041e68cffed	instructor1		2026-03-17 13:58:26.898349+00	2026-03-17 13:58:26.898349+00
3771420b-c55c-4698-9c9f-adab6f590c08	e5d9d7dc-c090-4b56-a870-6659b86ff01c	coach_v		2026-03-23 02:34:39.937349+00	2026-03-23 02:34:39.937349+00
\.


--
-- Data for Name: next_training_subjects; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.next_training_subjects (id, training_archive_id, syllabus_subject_id, assigned_by, created_at) FROM stdin;
4ed89393-fb52-480a-b1e6-499a4e64f02e	351b722a-a326-4def-b390-0a7dbc1e78a4	ab157b7c-d793-47c2-b6eb-c25a4cbe9368	16c9ed24-bd34-4e70-a681-303903910781	2026-03-25 06:40:11.621462+00
\.


--
-- Data for Name: raw_flight_imports; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.raw_flight_imports (id, batch_id, engine_on_time, engine_off_time, flight_duration_minutes, is_local, raw_data, training_record_id, created_at) FROM stdin;
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.schema_migrations (version, dirty) FROM stdin;
1	f
\.


--
-- Data for Name: sms_verification_codes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sms_verification_codes (id, phone, scene, code, ip, expired_at, used_at, created_at) FROM stdin;
\.


--
-- Data for Name: staff_schedules; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.staff_schedules (id, user_id, date, duty_status, remark, duty_windows, created_at, updated_at) FROM stdin;
e7dde365-7248-43af-ba63-c275317c5a3c	275f243c-8392-4738-b850-888c4e5246d8	2026-03-24	full_day	\N	\N	2026-03-22 12:37:26.530544+00	2026-03-22 12:37:26.530544+00
eea52ef5-049a-44f1-9c95-2170cf7738f2	275f243c-8392-4738-b850-888c4e5246d8	2026-03-25	full_day	\N	\N	2026-03-22 12:37:26.536473+00	2026-03-22 12:37:26.536473+00
936a0522-f9f8-49ba-9ef2-d9f7818f4d2d	275f243c-8392-4738-b850-888c4e5246d8	2026-03-26	full_day	\N	\N	2026-03-22 12:37:26.537803+00	2026-03-22 12:37:26.537803+00
794f8505-28d1-439c-9048-2068e8f1376e	275f243c-8392-4738-b850-888c4e5246d8	2026-03-27	full_day	\N	\N	2026-03-22 12:37:26.538687+00	2026-03-22 12:37:26.538687+00
a0360b8c-2356-485c-9ab6-0246422e67e3	275f243c-8392-4738-b850-888c4e5246d8	2026-03-30	full_day	\N	\N	2026-03-22 12:37:26.564626+00	2026-03-22 12:37:26.564626+00
66554111-5583-4da5-8cd0-9b9546e12474	275f243c-8392-4738-b850-888c4e5246d8	2026-03-31	full_day	\N	\N	2026-03-22 12:37:26.570688+00	2026-03-22 12:37:26.570688+00
d515bdae-b07b-479f-965a-0d7ce58bf3f7	275f243c-8392-4738-b850-888c4e5246d8	2026-03-23	full_day	\N	\N	2026-03-22 12:35:18.674254+00	2026-03-22 13:06:16.749756+00
d677b621-29e8-4ed5-bbfd-6a58c18abd89	afd0713a-a9a5-499a-98f2-63a1b9352602	2026-03-24	full_day	\N	\N	2026-03-23 13:30:45.725697+00	2026-03-23 13:30:45.725697+00
\.


--
-- Data for Name: students; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.students (id, user_id, name, student_no, contract_info, contract_duration_minutes, flown_minutes, training_progress, training_progress_json, responsible_instructor_id, created_at, updated_at) FROM stdin;
05626980-d27f-4248-8072-c83e7569b3d8	c13764be-834f-4634-bbac-c1442a1a5532	stu_vicent	S-c13764be-834f-4634-bbac-c1442a1a5532	\N	\N	0		\N	\N	2026-03-15 14:33:27.803608+00	2026-03-15 14:33:27.803608+00
37e727d6-869b-43a7-b8f8-35d82f55b64f	3a1b07e3-102b-4ebb-9e83-3b9b9c8c3b68	student2	S000001	\N	3600	0		\N	878dda87-3aeb-4fa8-819b-7e3ec65d8212	2026-03-16 11:38:47.067473+00	2026-03-16 11:38:47.067473+00
991c6deb-6170-4ac0-a3fe-b4948001522c	98cf2e2d-f9fe-40e1-9df8-6f1f4a4b9845	student1	S-98cf2e2d-f9fe-40e1-9df8-6f1f4a4b9845	\N	\N	0		\N	\N	2026-03-17 13:58:26.998778+00	2026-03-17 13:58:26.998778+00
\.


--
-- Data for Name: subject_attempts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.subject_attempts (id, user_id, subject_learning_id, answers, score, submitted_at, duration_seconds, passed, max_score, correct_count, total_count) FROM stdin;
\.


--
-- Data for Name: subject_learning; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.subject_learning (id, syllabus_subject_id, title, content_ref, video_url, created_at, updated_at, sort_order) FROM stdin;
\.


--
-- Data for Name: subject_questions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.subject_questions (id, subject_learning_id, question_text, options, correct_index, sort_order, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: syllabus_outlines; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.syllabus_outlines (id, category, name, sort_order, created_at, updated_at) FROM stdin;
00000001-0000-4000-8000-000000000001	ppl_basic	历史数据（迁移默认大纲）	0	2026-03-24 03:19:59.745619+00	2026-03-24 03:19:59.745619+00
4568e006-51a5-41af-b525-63b95fae73de	spl_local	SPL本场 导入版	0	2026-03-24 06:00:29.844127+00	2026-03-24 06:00:29.844127+00
6018e4f6-fbd0-4df9-b7e9-68ba109e6844	ppl_night	2025PPL夜航	0	2026-03-24 06:36:39.792614+00	2026-03-24 06:36:39.792614+00
\.


--
-- Data for Name: syllabus_subject_secondaries; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.syllabus_subject_secondaries (id, syllabus_subject_id, seq_no, name, description, completion_standard, sort_order, created_at, updated_at) FROM stdin;
e8e99e2d-0c24-497b-b695-724ecb24f3f6	f28e896c-caee-4f04-8c44-513b165213db	2	检查单22	这里是描述	独立完成	0	2026-03-24 06:08:43.536242+00	2026-03-24 06:13:17.005105+00
48cd1563-4a60-4077-9dab-89537eb4f484	f28e896c-caee-4f04-8c44-513b165213db	1	检查单	\N	独立完成	1	2026-03-24 06:00:29.844127+00	2026-03-24 06:13:17.005105+00
c9dc583a-4a27-49f6-8095-b6f5e68a2dc5	ab157b7c-d793-47c2-b6eb-c25a4cbe9368	1	检查单	\N	独立完成	0	2026-03-24 06:36:39.792614+00	2026-03-24 06:36:39.792614+00
ae0f6e6f-195e-497c-8ef8-b3a33e8ebe32	92210a28-6680-4c91-b135-f4130d775ec8	0	新二级科目2	\N	\N	0	2026-03-24 06:37:10.579441+00	2026-03-24 06:37:18.691474+00
b8c295a0-41bd-4a80-8f3a-d12289bb23c6	92210a28-6680-4c91-b135-f4130d775ec8	0	新二级科目	11111	go	1	2026-03-24 06:36:49.49711+00	2026-03-24 06:37:18.691474+00
\.


--
-- Data for Name: syllabus_subjects; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.syllabus_subjects (id, code, name, sort_order, created_at, updated_at, syllabus_outline_id) FROM stdin;
92210a28-6680-4c91-b135-f4130d775ec8	SUB-01	起落（更新后）	2	2026-03-16 09:05:47.464779+00	2026-03-16 09:07:15.93191+00	00000001-0000-4000-8000-000000000001
f633f473-7fce-4fad-8ea1-28eaebaa4e38	SUB-02	高飞	2	2026-03-16 09:09:37.448622+00	2026-03-16 09:09:37.448622+00	00000001-0000-4000-8000-000000000001
9988b22b-8d2b-454e-8dd5-863138600b57	SUB-03	平飞	3	2026-03-17 13:30:29.982361+00	2026-03-17 13:30:29.982361+00	00000001-0000-4000-8000-000000000001
fcf1176e-149d-4012-9fe5-b76cf6e8284a	SUB-04	俯冲	4	2026-03-17 13:30:55.733749+00	2026-03-17 13:30:55.733749+00	00000001-0000-4000-8000-000000000001
f28e896c-caee-4f04-8c44-513b165213db	01	地面准备	0	2026-03-24 06:00:29.844127+00	2026-03-24 06:00:29.844127+00	4568e006-51a5-41af-b525-63b95fae73de
ab157b7c-d793-47c2-b6eb-c25a4cbe9368	01	地面准备	0	2026-03-24 06:36:39.792614+00	2026-03-24 06:36:39.792614+00	6018e4f6-fbd0-4df9-b7e9-68ba109e6844
\.


--
-- Data for Name: tasks; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tasks (id, student_id, assigner_id, type, ref_id, title, description, due_at, status, created_at, updated_at) FROM stdin;
cc9c52a4-43c9-4b5d-9631-2a2310fb4545	05626980-d27f-4248-8072-c83e7569b3d8	16c9ed24-bd34-4e70-a681-303903910781	flight	06ecd346-8eed-4579-9ac4-6858b8691320	飞行训练：2026-03-19 10:00-11:30	预约飞行任务，教员ID=878dda87-3aeb-4fa8-819b-7e3ec65d8212，飞机ID=eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-19 10:00:00+00	done	2026-03-17 12:57:09.182618+00	2026-03-17 13:06:59.690363+00
8bf9961d-877b-43f8-b573-15f6da8cbfc3	05626980-d27f-4248-8072-c83e7569b3d8	16c9ed24-bd34-4e70-a681-303903910781	flight	ef9dd4f9-7b71-47f0-9018-e433c2d62579	飞行训练：2026-03-26 09:15-10:15	预约飞行任务，教员ID=878dda87-3aeb-4fa8-819b-7e3ec65d8212，飞机ID=b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	2026-03-26 09:15:00+00	pending	2026-03-25 08:25:53.179446+00	2026-03-25 08:25:53.179446+00
\.


--
-- Data for Name: theory_attempts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.theory_attempts (id, user_id, theory_document_id, answers, score, submitted_at) FROM stdin;
47145425-3559-449b-842f-b3d6a5d53457	16c9ed24-bd34-4e70-a681-303903910781	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-16 06:35:32.949451+00
1c6902e4-be9f-4db7-bb56-bd9de7b14d60	16c9ed24-bd34-4e70-a681-303903910781	550e8400-e29b-41d4-a716-446655440201	[{"question_id": "q1", "selected_option_ids": ["A"]}]	0	2026-03-16 06:36:56.49958+00
7f9182cb-9240-4689-a578-ed7d66175afe	98cf2e2d-f9fe-40e1-9df8-6f1f4a4b9845	8aebc6fb-09fb-43bc-918c-d7c0a277b5f0	[]	0	2026-03-17 13:59:24.71556+00
d85b5feb-e94c-4d20-8674-e0e6b0a348a9	c13764be-834f-4634-bbac-c1442a1a5532	0ba15a9e-56b9-4269-aa2a-e58cd7667f18	[{"question_id": "Q001", "selected_option_ids": ["B"]}, {"question_id": "Q002", "selected_option_ids": ["A"]}, {"question_id": "Q003", "selected_option_ids": ["C"]}, {"question_id": "Q004", "selected_option_ids": ["D"]}]	15	2026-03-18 14:09:43.153635+00
95b687a0-10a0-4b13-abad-5e2fe5876f82	c13764be-834f-4634-bbac-c1442a1a5532	0ba15a9e-56b9-4269-aa2a-e58cd7667f18	[]	0	2026-03-18 14:34:09.956662+00
aea3c4d6-6071-48ce-a8e7-548aa957d82d	c13764be-834f-4634-bbac-c1442a1a5532	0ba15a9e-56b9-4269-aa2a-e58cd7667f18	[]	0	2026-03-18 14:34:37.06158+00
38305f86-70ae-4bd1-8fd4-99a39b837418	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-26 09:12:02.815328+00
bc6f6180-edc8-44da-a715-9071137f5d71	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-26 09:12:02.815875+00
9bb15a3e-ffa3-4e20-b867-224c75270391	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-26 09:56:39.095238+00
9255e341-d47f-4d25-849f-07d066b36de6	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-26 09:56:39.095341+00
34d8db19-215c-4c68-adfe-be295887f6a1	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-26 09:57:27.719432+00
fd89e5c6-6e57-4e00-8c2a-b781d4d2fc55	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-26 09:57:27.71942+00
beb6c3c4-a148-41eb-8d93-68c347dc2591	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-26 10:06:34.010565+00
bbe4ea51-3fc9-4cc4-8c02-e53eb8b08ce9	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-26 10:06:34.00901+00
62d738ea-18b8-4f0d-86ed-2e48c0579ec9	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-26 10:07:14.601741+00
150d7110-4896-4fe9-8529-39b3c03f6888	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-26 10:07:14.601276+00
8ae118fb-36f3-442b-a8c1-98b5af2926e8	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-26 10:10:42.494556+00
897ef499-84b7-442c-9ea4-39de4c5fe75a	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-26 10:19:21.234816+00
f7e27dc7-2a3c-49b8-82ae-9abfa1e1b623	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[]	0	2026-03-26 10:19:44.296303+00
c2722fdf-3cec-4af7-90f1-b42e8c46e9d7	c13764be-834f-4634-bbac-c1442a1a5532	550e8400-e29b-41d4-a716-446655440201	[{"question_id": "q_1774515471475152000", "selected_option_ids": ["C"]}]	1	2026-03-26 10:22:57.805032+00
\.


--
-- Data for Name: theory_documents; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.theory_documents (id, title, slug, sort_order, content_ref, video_url, category, created_at, updated_at, video_storage, video_object_id, video_content_type, video_size_bytes, video_duration_seconds) FROM stdin;
906a25c3-cf47-4e9c-9f2f-aaecc73f8e66	失速与尾旋理论（更新后）	stall-spin-theory	20		https://video.example.com/stall-v2.mp4	空气动力学	2026-03-16 07:23:27.893726+00	2026-03-16 07:25:34.390793+00	\N	\N	\N	\N	\N
8aebc6fb-09fb-43bc-918c-d7c0a277b5f0	失速与尾旋理论（E2E 测试）	stall-spin-theory-e2e	10		https://video.example.com/stall-e2e.mp4	空气动力学	2026-03-17 13:58:42.411941+00	2026-03-24 02:39:34.120401+00	\N	\N	\N	\N	\N
c423371d-6734-4ce1-8c99-66e756056f35	新增个章节0324	0324	30			领航学	2026-03-24 02:40:24.4869+00	2026-03-24 02:40:24.4869+00	\N	\N	\N	\N	\N
0ba15a9e-56b9-4269-aa2a-e58cd7667f18	航空法与运行程序（更新后）	Air Law & Operational Procedures	10			法规	2026-03-18 08:51:43.436706+00	2026-03-26 08:13:13.899532+00	gridfs	69c4ea990449777255fefda7	video/quicktime	14798421	\N
550e8400-e29b-41d4-a716-446655440201	示例理论文档	sample-theory-doc	1	550e8400-e29b-41d4-a716-446655440201		气象学	2026-03-16 06:26:37.063391+00	2026-03-26 08:32:49.663307+00	gridfs	69c4ef180449777255fefe1b	video/quicktime	5861769	\N
\.


--
-- Data for Name: theory_paper_questions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.theory_paper_questions (id, paper_id, question_id, score, sort_order) FROM stdin;
99bf397d-78d2-47b5-a176-f89c810cbda9	d6051c4b-b917-436e-b33e-509ee211c3e9	q_1773755940012322000	2	1
4af10747-19e6-4c93-a1ea-83cebe36f942	d6051c4b-b917-436e-b33e-509ee211c3e9	q_1773755940047095000	3	2
19fa0ec9-01a0-482d-acd9-adc5d350b62b	8b7b2d0e-1efd-455f-b979-d6e8c822e011	Q001	5	1
5085e2ba-46f7-4f5d-99b3-19b323bf54be	8b7b2d0e-1efd-455f-b979-d6e8c822e011	Q002	5	2
\.


--
-- Data for Name: theory_papers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.theory_papers (id, theory_document_id, name, description, total_score, enabled, created_by, created_at, updated_at) FROM stdin;
d6051c4b-b917-436e-b33e-509ee211c3e9	8aebc6fb-09fb-43bc-918c-d7c0a277b5f0	失速与尾旋阶段测验（E2E）	E2E 全量测试试卷	5	t	242f6832-a633-48d3-87fd-6041e68cffed	2026-03-17 13:59:13.90733+00	2026-03-17 13:59:13.90733+00
8b7b2d0e-1efd-455f-b979-d6e8c822e011	0ba15a9e-56b9-4269-aa2a-e58cd7667f18	试卷001	第一张正式试卷	10	t	16c9ed24-bd34-4e70-a681-303903910781	2026-03-18 13:34:28.814967+00	2026-03-18 13:34:28.814967+00
\.


--
-- Data for Name: theory_questions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.theory_questions (id, theory_document_id, question_text, options, correct_index, sort_order, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: training_archive_subjects; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.training_archive_subjects (id, training_archive_id, syllabus_subject_id, evaluation_score, is_graduated, sort_order, subject_content, type, created_at, updated_at) FROM stdin;
132e2e0e-134d-4f8a-93eb-0d5b5190634c	916c8939-92f6-4689-9483-d079202354ad	92210a28-6680-4c91-b135-f4130d775ec8	9	f	0	本场起落 5 次	实践	2026-03-16 09:09:55.449249+00	2026-03-16 09:09:55.449249+00
1fd63a4c-4ab4-4767-9f1a-4e4e2eb85804	916c8939-92f6-4689-9483-d079202354ad	f633f473-7fce-4fad-8ea1-28eaebaa4e38	8	f	1	带飞航线	实践	2026-03-16 09:09:55.449249+00	2026-03-16 09:09:55.449249+00
592f7906-694e-44b4-8a76-1e12e00654db	351b722a-a326-4def-b390-0a7dbc1e78a4	92210a28-6680-4c91-b135-f4130d775ec8	20	f	0	本场起落 5 次	实践	2026-03-17 13:28:22.685445+00	2026-03-17 13:28:22.685445+00
32c8babd-28e7-486e-bc25-0f03873afef8	351b722a-a326-4def-b390-0a7dbc1e78a4	f633f473-7fce-4fad-8ea1-28eaebaa4e38	10	f	1	高飞9次	实践	2026-03-17 13:28:22.685445+00	2026-03-17 13:28:22.685445+00
\.


--
-- Data for Name: training_archives; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.training_archives (id, training_record_id, student_id, instructor_id, training_date, stage, planned_duration_minutes, total_experience_minutes, remaining_minutes, engine_on_time, engine_off_time, flight_duration_minutes, progress_note, instructor_review, student_self_review, task_assignment, student_signed_at, responsible_instructor_signed_at, chief_instructor_signed_at, student_signer_id, responsible_instructor_signer_id, chief_instructor_signer_id, responsible_instructor_id, chief_instructor_id, created_at, updated_at) FROM stdin;
916c8939-92f6-4689-9483-d079202354ad	a2cfbe63-5ebb-4717-ab87-bc83b219859b	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	2026-03-15	初级训练	90	2635	1804	2026-03-15 10:00:00+00	2026-03-15 11:32:00+00	92					\N	\N	\N	\N	\N	\N	\N	\N	2026-03-16 08:51:49.170919+00	2026-03-16 08:51:49.170919+00
351b722a-a326-4def-b390-0a7dbc1e78a4	ea42776e-041f-4cc7-a689-bbbbea52fe46	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	2026-03-18	初级训练	90	2635	1804	2026-03-18 10:00:00+00	2026-03-18 11:32:00+00	92	本课完成起落 5 次。	起飞姿态有进步，落地需注意收油门时机。	对五边速度控制更有把握。	下次重点练习短五边与侧风落地。	2026-03-17 13:22:44.135167+00	\N	\N	\N	\N	\N	\N	\N	2026-03-17 13:18:44.890228+00	2026-03-17 13:22:44.134609+00
\.


--
-- Data for Name: training_plans; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.training_plans (id, student_id, instructor_id, aircraft_id, plan_date, start_time, end_time, source, created_at, updated_at) FROM stdin;
60c72cde-bcee-4dc2-b51f-415ac737d50a	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-18	10:00:00	11:30:00	booking	2026-03-17 02:16:58.55587+00	2026-03-17 02:16:58.55587+00
f6315fb3-f022-4ca3-89d3-ad539cfe35b9	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-18	12:00:00	13:30:00	booking	2026-03-17 02:35:35.41945+00	2026-03-17 02:35:35.41945+00
13f9616b-a100-45a3-a844-396bdffc4afc	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-18	12:00:00	13:30:00	booking	2026-03-17 12:45:09.901289+00	2026-03-17 12:45:09.901289+00
9b037bfe-6f45-4256-bffa-df673171102e	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-19	12:00:00	13:30:00	booking	2026-03-17 12:54:14.533998+00	2026-03-17 12:54:14.533998+00
6b0891af-eeda-47dd-a8e8-0f84a773ebb5	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-19	10:00:00	11:30:00	booking	2026-03-17 12:57:09.177231+00	2026-03-17 12:57:09.177231+00
17b3e538-6794-400c-b143-6b040c6c331b	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	2026-03-26	09:15:00	10:15:00	booking	2026-03-25 08:25:53.172638+00	2026-03-25 08:25:53.172638+00
\.


--
-- Data for Name: training_record_views; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.training_record_views (id, user_id, name, filter_json, sort_json, group_by, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: training_records; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.training_records (id, student_id, instructor_id, aircraft_id, training_date, engine_on_time, engine_off_time, flight_duration_minutes, is_local, import_source, raw_import_id, status, planned_duration_minutes, theory_duration_minutes, sim_duration_minutes, syllabus_outline, archive_display_no, departure, destination, flight_type, task_type, landings, created_at, updated_at) FROM stdin;
a2cfbe63-5ebb-4717-ab87-bc83b219859b	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-15	2026-03-15 10:00:00+00	2026-03-15 11:30:00+00	90	t	manual	\N	normal	90	0	0	1#	1#672	\N	\N	\N	\N	5	2026-03-16 08:48:38.06301+00	2026-03-16 08:48:38.06301+00
ea42776e-041f-4cc7-a689-bbbbea52fe46	05626980-d27f-4248-8072-c83e7569b3d8	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-18	2026-03-18 10:00:00+00	2026-03-18 11:30:00+00	90	t	manual	\N	normal	90	0	0	1#	1#672	\N	\N	\N	\N	5	2026-03-17 13:03:41.687355+00	2026-03-17 13:03:41.687355+00
fb545555-cdbe-43c7-9a64-bc7c57cef0a1	05626980-d27f-4248-8072-c83e7569b3d8	34a5aa64-d7dd-493d-bed3-c45ff961dcf4	b98f4ff0-c8ff-4d11-94d1-8e60c7fc9e28	2026-03-24	2026-03-24 08:24:00+00	2026-03-24 10:23:00+00	119	f	manual	\N	normal	119	0	0	\N	\N	\N	\N	\N	\N	0	2026-03-24 07:20:15.291071+00	2026-03-24 07:20:15.291071+00
6d12a93d-ee95-439a-a377-3ecbd0dcd603	37e727d6-869b-43a7-b8f8-35d82f55b64f	878dda87-3aeb-4fa8-819b-7e3ec65d8212	eb5bd27f-009f-4267-b7c8-68a41ae77a23	2026-03-25	2026-03-25 01:10:00+00	2026-03-25 02:07:00+00	57	t	manual	\N	normal	57	0	0	\N	\N	\N	\N	\N	\N	0	2026-03-24 07:50:39.875007+00	2026-03-24 07:50:39.875007+00
\.


--
-- Data for Name: user_follows; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_follows (follower_id, followee_id, created_at) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, username, password_hash, phone, email, role, avatar_url, created_at, updated_at) FROM stdin;
8d56b96f-d5da-4016-803a-638715c2b515	admin	$2a$10$LsWUktg3FMppbLKSR8WUSee9Dp2gVcbRkNDaSslQz0gHdCzJRBe6a	18800008888	admin@magua.com	admin		2026-03-15 14:30:06.074014+00	2026-03-15 14:30:06.074014+00
16c9ed24-bd34-4e70-a681-303903910781	instructor_vicent	$2a$10$bcjkr.U4gJjCNJ2u1FY8Vu3wzkyxCk6wsTLJHv2iVGaYOgQQEDmXC	17700007777	instructor_vicent@magua.com	instructor		2026-03-15 14:31:49.604537+00	2026-03-15 14:31:49.604537+00
275f243c-8392-4738-b850-888c4e5246d8	staff_vicent	$2a$10$HTWSjG2MByVxi.yE832x3uFsRmq4uGT7hdQ58336FLm.oX462dKki	17700007777	staff_vicent@magua.com	staff		2026-03-15 14:32:35.78887+00	2026-03-15 14:32:35.78887+00
c13764be-834f-4634-bbac-c1442a1a5532	stu_vicent	$2a$10$YFpM8N534aYGInE/zCemoOhMKXIjFtr6jWXPzwiNGZVEl497YeFX6	13800001234	stu_vicent@magua.com	student		2026-03-15 14:33:27.797253+00	2026-03-15 14:33:27.797253+00
3a1b07e3-102b-4ebb-9e83-3b9b9c8c3b68	student2		13800000000	s1@example.com	student		2026-03-16 11:38:47.062463+00	2026-03-16 11:38:47.062463+00
242f6832-a633-48d3-87fd-6041e68cffed	instructor1	$2a$10$V1awWITJALGdbR6NjpnNa.tHEV/wPszYp2HXluZDbCzD/mPsz7bYG	13800000001	i1@example.com	instructor		2026-03-17 13:58:26.89435+00	2026-03-17 13:58:26.89435+00
98cf2e2d-f9fe-40e1-9df8-6f1f4a4b9845	student1	$2a$10$WjCtdLNq2LTIEUkdykXXreUhvoy8DNDLbs4lQK1Y5y1aCy68n.9/.	13800000002	s1@example.com	student		2026-03-17 13:58:26.994312+00	2026-03-17 13:58:26.994312+00
e5d9d7dc-c090-4b56-a870-6659b86ff01c	coach_v	$2a$10$ZaXnoEBkyH3cbwhxUFFwlez7oi9fElHtgnEwQeFoazbNbfKhidJoy	17713582108	coach_v@magua.com	instructor		2026-03-23 02:34:39.930829+00	2026-03-23 02:34:39.930829+00
afd0713a-a9a5-499a-98f2-63a1b9352602	jiwu_vicent	$2a$10$0tlRnA4boY13tbFvMVSM9u5jGy712Jh195VOl4bdFyc39XIVAIQTe	911	jiwu_vicent@magua.com	maintenance		2026-03-23 12:22:49.864892+00	2026-03-23 12:22:49.864892+00
7050b3bf-d35e-458a-8250-65b4d931ce94	dimian_vicent	$2a$10$sfdO1Dhcyei2ShxKYjVVq.zH6xTL52D3ybnrL5V96pdbw6jsqs3lC	912	dimian_vicent@magua.com	staff		2026-03-23 12:23:46.290234+00	2026-03-23 12:23:46.290234+00
\.


--
-- Name: aircraft_availability aircraft_availability_aircraft_id_date_start_time_end_time_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.aircraft_availability
    ADD CONSTRAINT aircraft_availability_aircraft_id_date_start_time_end_time_key UNIQUE (aircraft_id, date, start_time, end_time);


--
-- Name: aircraft_availability aircraft_availability_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.aircraft_availability
    ADD CONSTRAINT aircraft_availability_pkey PRIMARY KEY (id);


--
-- Name: aircraft_maintenance_windows aircraft_maintenance_windows_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.aircraft_maintenance_windows
    ADD CONSTRAINT aircraft_maintenance_windows_pkey PRIMARY KEY (id);


--
-- Name: aircraft aircraft_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.aircraft
    ADD CONSTRAINT aircraft_pkey PRIMARY KEY (id);


--
-- Name: bases bases_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bases
    ADD CONSTRAINT bases_pkey PRIMARY KEY (id);


--
-- Name: bookings bookings_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_pkey PRIMARY KEY (id);


--
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);


--
-- Name: community_challenge_entries community_challenge_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_challenge_entries
    ADD CONSTRAINT community_challenge_entries_pkey PRIMARY KEY (challenge_id, user_id);


--
-- Name: community_challenges community_challenges_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_challenges
    ADD CONSTRAINT community_challenges_pkey PRIMARY KEY (id);


--
-- Name: community_comments community_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_comments
    ADD CONSTRAINT community_comments_pkey PRIMARY KEY (id);


--
-- Name: community_group_members community_group_members_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_group_members
    ADD CONSTRAINT community_group_members_pkey PRIMARY KEY (group_id, user_id);


--
-- Name: community_groups community_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_groups
    ADD CONSTRAINT community_groups_pkey PRIMARY KEY (id);


--
-- Name: community_groups community_groups_slug_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_groups
    ADD CONSTRAINT community_groups_slug_key UNIQUE (slug);


--
-- Name: community_post_likes community_post_likes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_post_likes
    ADD CONSTRAINT community_post_likes_pkey PRIMARY KEY (post_id, user_id);


--
-- Name: community_post_topics community_post_topics_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_post_topics
    ADD CONSTRAINT community_post_topics_pkey PRIMARY KEY (post_id, topic_id);


--
-- Name: community_posts community_posts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_posts
    ADD CONSTRAINT community_posts_pkey PRIMARY KEY (id);


--
-- Name: community_topics community_topics_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_topics
    ADD CONSTRAINT community_topics_pkey PRIMARY KEY (id);


--
-- Name: community_topics community_topics_slug_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_topics
    ADD CONSTRAINT community_topics_slug_key UNIQUE (slug);


--
-- Name: contracts contracts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.contracts
    ADD CONSTRAINT contracts_pkey PRIMARY KEY (id);


--
-- Name: day_operational_config day_operational_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.day_operational_config
    ADD CONSTRAINT day_operational_config_pkey PRIMARY KEY (id);


--
-- Name: instructor_schedules instructor_schedules_instructor_id_date_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.instructor_schedules
    ADD CONSTRAINT instructor_schedules_instructor_id_date_key UNIQUE (instructor_id, date);


--
-- Name: instructor_schedules instructor_schedules_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.instructor_schedules
    ADD CONSTRAINT instructor_schedules_pkey PRIMARY KEY (id);


--
-- Name: instructors instructors_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.instructors
    ADD CONSTRAINT instructors_pkey PRIMARY KEY (id);


--
-- Name: instructors instructors_user_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.instructors
    ADD CONSTRAINT instructors_user_id_key UNIQUE (user_id);


--
-- Name: next_training_subjects next_training_subjects_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.next_training_subjects
    ADD CONSTRAINT next_training_subjects_pkey PRIMARY KEY (id);


--
-- Name: raw_flight_imports raw_flight_imports_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.raw_flight_imports
    ADD CONSTRAINT raw_flight_imports_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: sms_verification_codes sms_verification_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sms_verification_codes
    ADD CONSTRAINT sms_verification_codes_pkey PRIMARY KEY (id);


--
-- Name: staff_schedules staff_schedules_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.staff_schedules
    ADD CONSTRAINT staff_schedules_pkey PRIMARY KEY (id);


--
-- Name: staff_schedules staff_schedules_user_id_date_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.staff_schedules
    ADD CONSTRAINT staff_schedules_user_id_date_key UNIQUE (user_id, date);


--
-- Name: students students_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.students
    ADD CONSTRAINT students_pkey PRIMARY KEY (id);


--
-- Name: students students_student_no_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.students
    ADD CONSTRAINT students_student_no_key UNIQUE (student_no);


--
-- Name: students students_user_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.students
    ADD CONSTRAINT students_user_id_key UNIQUE (user_id);


--
-- Name: subject_attempts subject_attempts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subject_attempts
    ADD CONSTRAINT subject_attempts_pkey PRIMARY KEY (id);


--
-- Name: subject_learning subject_learning_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subject_learning
    ADD CONSTRAINT subject_learning_pkey PRIMARY KEY (id);


--
-- Name: subject_learning subject_learning_syllabus_subject_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subject_learning
    ADD CONSTRAINT subject_learning_syllabus_subject_id_key UNIQUE (syllabus_subject_id);


--
-- Name: subject_questions subject_questions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subject_questions
    ADD CONSTRAINT subject_questions_pkey PRIMARY KEY (id);


--
-- Name: syllabus_outlines syllabus_outlines_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.syllabus_outlines
    ADD CONSTRAINT syllabus_outlines_pkey PRIMARY KEY (id);


--
-- Name: syllabus_subject_secondaries syllabus_subject_secondaries_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.syllabus_subject_secondaries
    ADD CONSTRAINT syllabus_subject_secondaries_pkey PRIMARY KEY (id);


--
-- Name: syllabus_subjects syllabus_subjects_outline_code_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.syllabus_subjects
    ADD CONSTRAINT syllabus_subjects_outline_code_unique UNIQUE (syllabus_outline_id, code);


--
-- Name: syllabus_subjects syllabus_subjects_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.syllabus_subjects
    ADD CONSTRAINT syllabus_subjects_pkey PRIMARY KEY (id);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- Name: theory_attempts theory_attempts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.theory_attempts
    ADD CONSTRAINT theory_attempts_pkey PRIMARY KEY (id);


--
-- Name: theory_documents theory_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.theory_documents
    ADD CONSTRAINT theory_documents_pkey PRIMARY KEY (id);


--
-- Name: theory_documents theory_documents_slug_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.theory_documents
    ADD CONSTRAINT theory_documents_slug_key UNIQUE (slug);


--
-- Name: theory_paper_questions theory_paper_questions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.theory_paper_questions
    ADD CONSTRAINT theory_paper_questions_pkey PRIMARY KEY (id);


--
-- Name: theory_papers theory_papers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.theory_papers
    ADD CONSTRAINT theory_papers_pkey PRIMARY KEY (id);


--
-- Name: theory_questions theory_questions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.theory_questions
    ADD CONSTRAINT theory_questions_pkey PRIMARY KEY (id);


--
-- Name: training_archive_subjects training_archive_subjects_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_archive_subjects
    ADD CONSTRAINT training_archive_subjects_pkey PRIMARY KEY (id);


--
-- Name: training_archives training_archives_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_archives
    ADD CONSTRAINT training_archives_pkey PRIMARY KEY (id);


--
-- Name: training_archives training_archives_training_record_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_archives
    ADD CONSTRAINT training_archives_training_record_id_key UNIQUE (training_record_id);


--
-- Name: training_plans training_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_plans
    ADD CONSTRAINT training_plans_pkey PRIMARY KEY (id);


--
-- Name: training_record_views training_record_views_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_record_views
    ADD CONSTRAINT training_record_views_pkey PRIMARY KEY (id);


--
-- Name: training_records training_records_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_records
    ADD CONSTRAINT training_records_pkey PRIMARY KEY (id);


--
-- Name: user_follows user_follows_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_follows
    ADD CONSTRAINT user_follows_pkey PRIMARY KEY (follower_id, followee_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: idx_aircraft_availability_aircraft_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_aircraft_availability_aircraft_date ON public.aircraft_availability USING btree (aircraft_id, date);


--
-- Name: idx_aircraft_availability_maintenance_window; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_aircraft_availability_maintenance_window ON public.aircraft_availability USING btree (maintenance_window_id);


--
-- Name: idx_aircraft_maintenance_windows_aircraft_time; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_aircraft_maintenance_windows_aircraft_time ON public.aircraft_maintenance_windows USING btree (aircraft_id, scheduled_start_at);


--
-- Name: idx_bookings_student_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_bookings_student_status ON public.bookings USING btree (student_id, status);


--
-- Name: idx_comments_parent; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_comments_parent ON public.comments USING btree (parent_id);


--
-- Name: idx_comments_target; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_comments_target ON public.comments USING btree (target_type, target_id);


--
-- Name: idx_community_challenge_entries_post; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_community_challenge_entries_post ON public.community_challenge_entries USING btree (post_id);


--
-- Name: idx_community_challenges_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_community_challenges_status ON public.community_challenges USING btree (status, starts_at DESC);


--
-- Name: idx_community_comments_post_created; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_community_comments_post_created ON public.community_comments USING btree (post_id, created_at);


--
-- Name: idx_community_group_members_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_community_group_members_user ON public.community_group_members USING btree (user_id);


--
-- Name: idx_community_post_likes_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_community_post_likes_user ON public.community_post_likes USING btree (user_id);


--
-- Name: idx_community_posts_author_created; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_community_posts_author_created ON public.community_posts USING btree (author_user_id, created_at DESC);


--
-- Name: idx_community_posts_created; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_community_posts_created ON public.community_posts USING btree (created_at DESC);


--
-- Name: idx_community_posts_group_created; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_community_posts_group_created ON public.community_posts USING btree (group_id, created_at DESC) WHERE (group_id IS NOT NULL);


--
-- Name: idx_community_topics_post_count; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_community_topics_post_count ON public.community_topics USING btree (post_count DESC);


--
-- Name: idx_contracts_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_contracts_status ON public.contracts USING btree (status);


--
-- Name: idx_contracts_student_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_contracts_student_id ON public.contracts USING btree (student_id);


--
-- Name: idx_day_operational_config_date_base; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_day_operational_config_date_base ON public.day_operational_config USING btree (date, COALESCE((base_id)::text, ''::text));


--
-- Name: idx_instructor_schedules_instructor_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_instructor_schedules_instructor_date ON public.instructor_schedules USING btree (instructor_id, date);


--
-- Name: idx_raw_flight_imports_batch; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_raw_flight_imports_batch ON public.raw_flight_imports USING btree (batch_id);


--
-- Name: idx_sms_verification_phone_scene_created_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_sms_verification_phone_scene_created_at ON public.sms_verification_codes USING btree (phone, scene, created_at DESC);


--
-- Name: idx_staff_schedules_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_staff_schedules_date ON public.staff_schedules USING btree (date);


--
-- Name: idx_staff_schedules_user_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_staff_schedules_user_date ON public.staff_schedules USING btree (user_id, date);


--
-- Name: idx_students_responsible_instructor; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_students_responsible_instructor ON public.students USING btree (responsible_instructor_id);


--
-- Name: idx_students_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_students_user_id ON public.students USING btree (user_id);


--
-- Name: idx_subject_attempts_user_subject; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_subject_attempts_user_subject ON public.subject_attempts USING btree (user_id, subject_learning_id);


--
-- Name: idx_syllabus_subject_secondaries_subject_sort; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_syllabus_subject_secondaries_subject_sort ON public.syllabus_subject_secondaries USING btree (syllabus_subject_id, sort_order, id);


--
-- Name: idx_tasks_student_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tasks_student_status ON public.tasks USING btree (student_id, status);


--
-- Name: idx_theory_attempts_user_doc; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_theory_attempts_user_doc ON public.theory_attempts USING btree (user_id, theory_document_id);


--
-- Name: idx_theory_paper_questions_paper; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_theory_paper_questions_paper ON public.theory_paper_questions USING btree (paper_id);


--
-- Name: idx_theory_papers_doc; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_theory_papers_doc ON public.theory_papers USING btree (theory_document_id);


--
-- Name: idx_training_archives_student; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_training_archives_student ON public.training_archives USING btree (student_id);


--
-- Name: idx_training_archives_student_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_training_archives_student_date ON public.training_archives USING btree (student_id, training_date);


--
-- Name: idx_training_plans_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_training_plans_date ON public.training_plans USING btree (plan_date);


--
-- Name: idx_training_record_views_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_training_record_views_user ON public.training_record_views USING btree (user_id);


--
-- Name: idx_training_records_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_training_records_status ON public.training_records USING btree (status);


--
-- Name: idx_training_records_student_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_training_records_student_date ON public.training_records USING btree (student_id, training_date);


--
-- Name: idx_user_follows_followee; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_follows_followee ON public.user_follows USING btree (followee_id);


--
-- Name: community_comments tr_community_comments_ad; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER tr_community_comments_ad AFTER DELETE ON public.community_comments FOR EACH ROW EXECUTE FUNCTION public.community_posts_bump_comment_count();


--
-- Name: community_comments tr_community_comments_ai; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER tr_community_comments_ai AFTER INSERT ON public.community_comments FOR EACH ROW EXECUTE FUNCTION public.community_posts_bump_comment_count();


--
-- Name: community_group_members tr_community_group_members_ad; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER tr_community_group_members_ad AFTER DELETE ON public.community_group_members FOR EACH ROW EXECUTE FUNCTION public.community_groups_bump_member_count();


--
-- Name: community_group_members tr_community_group_members_ai; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER tr_community_group_members_ai AFTER INSERT ON public.community_group_members FOR EACH ROW EXECUTE FUNCTION public.community_groups_bump_member_count();


--
-- Name: community_post_likes tr_community_post_likes_ad; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER tr_community_post_likes_ad AFTER DELETE ON public.community_post_likes FOR EACH ROW EXECUTE FUNCTION public.community_posts_bump_like_count();


--
-- Name: community_post_likes tr_community_post_likes_ai; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER tr_community_post_likes_ai AFTER INSERT ON public.community_post_likes FOR EACH ROW EXECUTE FUNCTION public.community_posts_bump_like_count();


--
-- Name: community_post_topics tr_community_post_topics_ad; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER tr_community_post_topics_ad AFTER DELETE ON public.community_post_topics FOR EACH ROW EXECUTE FUNCTION public.community_topics_bump_post_count();


--
-- Name: community_post_topics tr_community_post_topics_ai; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER tr_community_post_topics_ai AFTER INSERT ON public.community_post_topics FOR EACH ROW EXECUTE FUNCTION public.community_topics_bump_post_count();


--
-- Name: aircraft_availability aircraft_availability_aircraft_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.aircraft_availability
    ADD CONSTRAINT aircraft_availability_aircraft_id_fkey FOREIGN KEY (aircraft_id) REFERENCES public.aircraft(id) ON DELETE CASCADE;


--
-- Name: aircraft_availability aircraft_availability_maintenance_window_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.aircraft_availability
    ADD CONSTRAINT aircraft_availability_maintenance_window_id_fkey FOREIGN KEY (maintenance_window_id) REFERENCES public.aircraft_maintenance_windows(id) ON DELETE CASCADE;


--
-- Name: aircraft_maintenance_windows aircraft_maintenance_windows_aircraft_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.aircraft_maintenance_windows
    ADD CONSTRAINT aircraft_maintenance_windows_aircraft_id_fkey FOREIGN KEY (aircraft_id) REFERENCES public.aircraft(id) ON DELETE CASCADE;


--
-- Name: bookings bookings_aircraft_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_aircraft_id_fkey FOREIGN KEY (aircraft_id) REFERENCES public.aircraft(id) ON DELETE CASCADE;


--
-- Name: bookings bookings_instructor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_instructor_id_fkey FOREIGN KEY (instructor_id) REFERENCES public.instructors(id) ON DELETE CASCADE;


--
-- Name: bookings bookings_reviewer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_reviewer_id_fkey FOREIGN KEY (reviewer_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: bookings bookings_student_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_student_id_fkey FOREIGN KEY (student_id) REFERENCES public.students(id) ON DELETE CASCADE;


--
-- Name: comments comments_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.comments(id) ON DELETE CASCADE;


--
-- Name: comments comments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: community_challenge_entries community_challenge_entries_challenge_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_challenge_entries
    ADD CONSTRAINT community_challenge_entries_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES public.community_challenges(id) ON DELETE CASCADE;


--
-- Name: community_challenge_entries community_challenge_entries_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_challenge_entries
    ADD CONSTRAINT community_challenge_entries_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.community_posts(id) ON DELETE CASCADE;


--
-- Name: community_challenge_entries community_challenge_entries_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_challenge_entries
    ADD CONSTRAINT community_challenge_entries_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: community_comments community_comments_author_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_comments
    ADD CONSTRAINT community_comments_author_user_id_fkey FOREIGN KEY (author_user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: community_comments community_comments_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_comments
    ADD CONSTRAINT community_comments_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.community_comments(id) ON DELETE CASCADE;


--
-- Name: community_comments community_comments_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_comments
    ADD CONSTRAINT community_comments_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.community_posts(id) ON DELETE CASCADE;


--
-- Name: community_group_members community_group_members_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_group_members
    ADD CONSTRAINT community_group_members_group_id_fkey FOREIGN KEY (group_id) REFERENCES public.community_groups(id) ON DELETE CASCADE;


--
-- Name: community_group_members community_group_members_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_group_members
    ADD CONSTRAINT community_group_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: community_groups community_groups_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_groups
    ADD CONSTRAINT community_groups_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: community_post_likes community_post_likes_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_post_likes
    ADD CONSTRAINT community_post_likes_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.community_posts(id) ON DELETE CASCADE;


--
-- Name: community_post_likes community_post_likes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_post_likes
    ADD CONSTRAINT community_post_likes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: community_post_topics community_post_topics_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_post_topics
    ADD CONSTRAINT community_post_topics_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.community_posts(id) ON DELETE CASCADE;


--
-- Name: community_post_topics community_post_topics_topic_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_post_topics
    ADD CONSTRAINT community_post_topics_topic_id_fkey FOREIGN KEY (topic_id) REFERENCES public.community_topics(id) ON DELETE CASCADE;


--
-- Name: community_posts community_posts_author_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_posts
    ADD CONSTRAINT community_posts_author_user_id_fkey FOREIGN KEY (author_user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: community_posts community_posts_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.community_posts
    ADD CONSTRAINT community_posts_group_id_fkey FOREIGN KEY (group_id) REFERENCES public.community_groups(id) ON DELETE SET NULL;


--
-- Name: contracts contracts_student_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.contracts
    ADD CONSTRAINT contracts_student_id_fkey FOREIGN KEY (student_id) REFERENCES public.students(id) ON DELETE CASCADE;


--
-- Name: day_operational_config day_operational_config_base_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.day_operational_config
    ADD CONSTRAINT day_operational_config_base_id_fkey FOREIGN KEY (base_id) REFERENCES public.bases(id);


--
-- Name: instructor_schedules instructor_schedules_instructor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.instructor_schedules
    ADD CONSTRAINT instructor_schedules_instructor_id_fkey FOREIGN KEY (instructor_id) REFERENCES public.instructors(id) ON DELETE CASCADE;


--
-- Name: instructors instructors_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.instructors
    ADD CONSTRAINT instructors_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: next_training_subjects next_training_subjects_assigned_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.next_training_subjects
    ADD CONSTRAINT next_training_subjects_assigned_by_fkey FOREIGN KEY (assigned_by) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: next_training_subjects next_training_subjects_syllabus_subject_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.next_training_subjects
    ADD CONSTRAINT next_training_subjects_syllabus_subject_id_fkey FOREIGN KEY (syllabus_subject_id) REFERENCES public.syllabus_subjects(id) ON DELETE CASCADE;


--
-- Name: next_training_subjects next_training_subjects_training_archive_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.next_training_subjects
    ADD CONSTRAINT next_training_subjects_training_archive_id_fkey FOREIGN KEY (training_archive_id) REFERENCES public.training_archives(id) ON DELETE CASCADE;


--
-- Name: raw_flight_imports raw_flight_imports_training_record_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.raw_flight_imports
    ADD CONSTRAINT raw_flight_imports_training_record_id_fkey FOREIGN KEY (training_record_id) REFERENCES public.training_records(id) ON DELETE SET NULL;


--
-- Name: staff_schedules staff_schedules_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.staff_schedules
    ADD CONSTRAINT staff_schedules_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: students students_responsible_instructor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.students
    ADD CONSTRAINT students_responsible_instructor_id_fkey FOREIGN KEY (responsible_instructor_id) REFERENCES public.instructors(id) ON DELETE SET NULL;


--
-- Name: students students_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.students
    ADD CONSTRAINT students_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: subject_attempts subject_attempts_subject_learning_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subject_attempts
    ADD CONSTRAINT subject_attempts_subject_learning_id_fkey FOREIGN KEY (subject_learning_id) REFERENCES public.subject_learning(id) ON DELETE CASCADE;


--
-- Name: subject_attempts subject_attempts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subject_attempts
    ADD CONSTRAINT subject_attempts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: subject_learning subject_learning_syllabus_subject_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subject_learning
    ADD CONSTRAINT subject_learning_syllabus_subject_id_fkey FOREIGN KEY (syllabus_subject_id) REFERENCES public.syllabus_subjects(id) ON DELETE CASCADE;


--
-- Name: subject_questions subject_questions_subject_learning_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subject_questions
    ADD CONSTRAINT subject_questions_subject_learning_id_fkey FOREIGN KEY (subject_learning_id) REFERENCES public.subject_learning(id) ON DELETE CASCADE;


--
-- Name: syllabus_subject_secondaries syllabus_subject_secondaries_syllabus_subject_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.syllabus_subject_secondaries
    ADD CONSTRAINT syllabus_subject_secondaries_syllabus_subject_id_fkey FOREIGN KEY (syllabus_subject_id) REFERENCES public.syllabus_subjects(id) ON DELETE CASCADE;


--
-- Name: syllabus_subjects syllabus_subjects_syllabus_outline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.syllabus_subjects
    ADD CONSTRAINT syllabus_subjects_syllabus_outline_id_fkey FOREIGN KEY (syllabus_outline_id) REFERENCES public.syllabus_outlines(id) ON DELETE CASCADE;


--
-- Name: tasks tasks_assigner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_assigner_id_fkey FOREIGN KEY (assigner_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: tasks tasks_student_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_student_id_fkey FOREIGN KEY (student_id) REFERENCES public.students(id) ON DELETE CASCADE;


--
-- Name: theory_attempts theory_attempts_theory_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.theory_attempts
    ADD CONSTRAINT theory_attempts_theory_document_id_fkey FOREIGN KEY (theory_document_id) REFERENCES public.theory_documents(id) ON DELETE CASCADE;


--
-- Name: theory_attempts theory_attempts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.theory_attempts
    ADD CONSTRAINT theory_attempts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: theory_paper_questions theory_paper_questions_paper_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.theory_paper_questions
    ADD CONSTRAINT theory_paper_questions_paper_id_fkey FOREIGN KEY (paper_id) REFERENCES public.theory_papers(id) ON DELETE CASCADE;


--
-- Name: theory_papers theory_papers_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.theory_papers
    ADD CONSTRAINT theory_papers_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: theory_papers theory_papers_theory_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.theory_papers
    ADD CONSTRAINT theory_papers_theory_document_id_fkey FOREIGN KEY (theory_document_id) REFERENCES public.theory_documents(id) ON DELETE CASCADE;


--
-- Name: theory_questions theory_questions_theory_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.theory_questions
    ADD CONSTRAINT theory_questions_theory_document_id_fkey FOREIGN KEY (theory_document_id) REFERENCES public.theory_documents(id) ON DELETE CASCADE;


--
-- Name: training_archive_subjects training_archive_subjects_syllabus_subject_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_archive_subjects
    ADD CONSTRAINT training_archive_subjects_syllabus_subject_id_fkey FOREIGN KEY (syllabus_subject_id) REFERENCES public.syllabus_subjects(id) ON DELETE CASCADE;


--
-- Name: training_archive_subjects training_archive_subjects_training_archive_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_archive_subjects
    ADD CONSTRAINT training_archive_subjects_training_archive_id_fkey FOREIGN KEY (training_archive_id) REFERENCES public.training_archives(id) ON DELETE CASCADE;


--
-- Name: training_archives training_archives_chief_instructor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_archives
    ADD CONSTRAINT training_archives_chief_instructor_id_fkey FOREIGN KEY (chief_instructor_id) REFERENCES public.instructors(id) ON DELETE SET NULL;


--
-- Name: training_archives training_archives_chief_instructor_signer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_archives
    ADD CONSTRAINT training_archives_chief_instructor_signer_id_fkey FOREIGN KEY (chief_instructor_signer_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: training_archives training_archives_instructor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_archives
    ADD CONSTRAINT training_archives_instructor_id_fkey FOREIGN KEY (instructor_id) REFERENCES public.instructors(id) ON DELETE CASCADE;


--
-- Name: training_archives training_archives_responsible_instructor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_archives
    ADD CONSTRAINT training_archives_responsible_instructor_id_fkey FOREIGN KEY (responsible_instructor_id) REFERENCES public.instructors(id) ON DELETE SET NULL;


--
-- Name: training_archives training_archives_responsible_instructor_signer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_archives
    ADD CONSTRAINT training_archives_responsible_instructor_signer_id_fkey FOREIGN KEY (responsible_instructor_signer_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: training_archives training_archives_student_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_archives
    ADD CONSTRAINT training_archives_student_id_fkey FOREIGN KEY (student_id) REFERENCES public.students(id) ON DELETE CASCADE;


--
-- Name: training_archives training_archives_student_signer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_archives
    ADD CONSTRAINT training_archives_student_signer_id_fkey FOREIGN KEY (student_signer_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: training_archives training_archives_training_record_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_archives
    ADD CONSTRAINT training_archives_training_record_id_fkey FOREIGN KEY (training_record_id) REFERENCES public.training_records(id) ON DELETE CASCADE;


--
-- Name: training_plans training_plans_aircraft_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_plans
    ADD CONSTRAINT training_plans_aircraft_id_fkey FOREIGN KEY (aircraft_id) REFERENCES public.aircraft(id) ON DELETE CASCADE;


--
-- Name: training_plans training_plans_instructor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_plans
    ADD CONSTRAINT training_plans_instructor_id_fkey FOREIGN KEY (instructor_id) REFERENCES public.instructors(id) ON DELETE CASCADE;


--
-- Name: training_plans training_plans_student_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_plans
    ADD CONSTRAINT training_plans_student_id_fkey FOREIGN KEY (student_id) REFERENCES public.students(id) ON DELETE CASCADE;


--
-- Name: training_record_views training_record_views_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_record_views
    ADD CONSTRAINT training_record_views_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: training_records training_records_aircraft_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_records
    ADD CONSTRAINT training_records_aircraft_id_fkey FOREIGN KEY (aircraft_id) REFERENCES public.aircraft(id) ON DELETE CASCADE;


--
-- Name: training_records training_records_instructor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_records
    ADD CONSTRAINT training_records_instructor_id_fkey FOREIGN KEY (instructor_id) REFERENCES public.instructors(id) ON DELETE CASCADE;


--
-- Name: training_records training_records_student_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.training_records
    ADD CONSTRAINT training_records_student_id_fkey FOREIGN KEY (student_id) REFERENCES public.students(id) ON DELETE CASCADE;


--
-- Name: user_follows user_follows_followee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_follows
    ADD CONSTRAINT user_follows_followee_id_fkey FOREIGN KEY (followee_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_follows user_follows_follower_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_follows
    ADD CONSTRAINT user_follows_follower_id_fkey FOREIGN KEY (follower_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict 1DanrRl7n90YVoUokwmY1sE56xm6r4cGPfABkgQPYhgtw7C2axc3qtco9QZS5de

