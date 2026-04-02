-- 用户任职状态：与 deleted_at 软删除区分；软删用于归档/学号复用；任职状态控制登录与选人列表可见性。

DO $$ BEGIN
  CREATE TYPE employment_status AS ENUM ('active', 'resigned', 'part_time');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS employment_status employment_status NOT NULL DEFAULT 'active';

COMMENT ON COLUMN users.employment_status IS '在职 active / 离职 resigned / 兼职 part_time；非 active 时账号不可登录（admin 角色在中间件中豁免校验）';
