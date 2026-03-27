-- 021: bookings / training_plans 必须具备 plan_date，与 internal/postgres/*_repo.go 一致。
-- 正式环境若缺该列（迁移未跑全、或从旧库恢复），ListByDateRange / ListByDate 会报 42703。

DO $$
BEGIN
  IF to_regclass('public.bookings') IS NOT NULL
     AND NOT EXISTS (
       SELECT 1 FROM information_schema.columns
       WHERE table_schema = 'public' AND table_name = 'bookings' AND column_name = 'plan_date'
     ) THEN
    ALTER TABLE public.bookings ADD COLUMN plan_date date;
    UPDATE public.bookings SET plan_date = (created_at AT TIME ZONE 'UTC')::date WHERE plan_date IS NULL;
    ALTER TABLE public.bookings ALTER COLUMN plan_date SET NOT NULL;
  END IF;
END $$;

DO $$
BEGIN
  IF to_regclass('public.training_plans') IS NOT NULL
     AND NOT EXISTS (
       SELECT 1 FROM information_schema.columns
       WHERE table_schema = 'public' AND table_name = 'training_plans' AND column_name = 'plan_date'
     ) THEN
    ALTER TABLE public.training_plans ADD COLUMN plan_date date;
    UPDATE public.training_plans SET plan_date = (created_at AT TIME ZONE 'UTC')::date WHERE plan_date IS NULL;
    ALTER TABLE public.training_plans ALTER COLUMN plan_date SET NOT NULL;
  END IF;
END $$;
