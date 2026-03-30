-- 022: bookings 审核相关列与 008_uuid_schema / booking_repo 一致。
-- 正式库若缺列，ListPending / ListByDateRange 等会报 42703。

ALTER TABLE public.bookings ADD COLUMN IF NOT EXISTS reject_reason text;
ALTER TABLE public.bookings ADD COLUMN IF NOT EXISTS reviewed_at timestamptz;
ALTER TABLE public.bookings ADD COLUMN IF NOT EXISTS reviewer_id uuid;

-- 尽量补外键（已存在同名列时 IF NOT EXISTS 不会重复添加列，本语句仅在新加列时生效）
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'bookings' AND column_name = 'reviewer_id'
  ) AND NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'bookings_reviewer_id_fkey'
  ) AND to_regclass('public.users') IS NOT NULL THEN
    ALTER TABLE public.bookings
      ADD CONSTRAINT bookings_reviewer_id_fkey
      FOREIGN KEY (reviewer_id) REFERENCES public.users (id) ON DELETE SET NULL;
  END IF;
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
