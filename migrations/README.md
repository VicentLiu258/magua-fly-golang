# 数据库迁移说明

本目录包含基于 [ER 分析与调整方案](.cursor/plans/analyze-flight-training-db_e6567e32.plan.md) 的 PostgreSQL 迁移脚本，按优先级分步执行。

## 执行顺序

| 顺序 | 文件 | 说明 |
|------|------|------|
| 1 | `001_constraints_and_indexes.sql` | 用户/学生/教员唯一约束，comments、tasks 等核心索引 |
| 2 | `002_training_scheduling_indexes.sql` | 训练/排班/预约索引，bookings.training_plan_id，飞机时段唯一 |
| 3 | `003_theory_and_learning_fields.sql` | 理论文档 content_blocks、题库多选与统计字段 |
| 4 | `004_tasks_and_comments_fields.sql` | 任务 priority/completed_at，评论软删除与可见性 |
| 5 | `005_optional_enhancements.sql` | 可选：bases、day_operational_config、进度 JSONB、签署人 ID 等 |

## 环境要求

- PostgreSQL 11+（003、005 中使用了 `ADD COLUMN IF NOT EXISTS`；若使用 9.x 可将新增列改为 DO 块判断后执行）。

## 使用方式

- 确保数据库已按 ER 图建好表结构（含枚举类型：`user_role`, `availability_status`, `booking_status`, `plan_source`, `import_source`, `task_type`, `task_status`, `comment_target_type` 等）。
- 在项目根目录或任意可连接数据库的环境执行：

```bash
# 单次执行（按顺序）
psql -U your_user -d your_db -f db/migrations/001_constraints_and_indexes.sql
psql -U your_user -d your_db -f db/migrations/002_training_scheduling_indexes.sql
# ...
```

- 若使用迁移工具（如 node-pg-migrate、Flyway、atlas），可将上述 SQL 拆成对应格式并按顺序注册。

## 注意事项

1. **001**：若 `comments` 表中目标类型列名为 `comment_target_type` 而非 `target_type`，请将索引 `idx_comments_target` 改为 `(comment_target_type, target_id)`，并同步修改 **004** 中的 `idx_comments_target_not_deleted`。
2. **002**：`training_plans` 上的飞机时段唯一索引会阻止同一飞机同一时段插入多条计划；若业务允许“候选多学员”等，可改为部分唯一或去掉该唯一索引。
3. **005**：为可选增强，按需执行；新增的 `bases`、`day_operational_config` 需在应用层写入与查询。

## 枚举约定（建议与前端一致）

- `user_role`: student, instructor, admin, pilot
- `booking_status`: pending, approved, rejected, cancelled, completed
- `task_type`: training_record_review, training_archive_sign, prepare_theory_doc, student_evaluation, generic
- `comment_target_type`: theory_document, video, training_record, training_archive, forum_post, subject_learning, task
