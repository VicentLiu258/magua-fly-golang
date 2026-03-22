# 执行迁移 001～005 的命令

在**项目根目录** `magua-fly-golang` 下执行。  
请先确认：① 已存在基础表（执行过 `000001_initial.up.sql`）；② 数据库为 `maguafly`；③ Postgres 容器名（`docker ps` 查看，下面以 `docker-postgres-1` 为例）。

---

## 方式一：用 Docker 执行（推荐，不依赖本机 psql）

```bash
cd /Users/liuwei/magua-fly/magua-fly-golang

# 001：用户/角色唯一约束与核心索引
docker exec -i docker-postgres-1 psql -U postgres -d maguafly < ./migrations/001_constraints_and_indexes.sql

# 002：训练排班与预约约束与索引
docker exec -i docker-postgres-1 psql -U postgres -d maguafly < ./migrations/002_training_scheduling_indexes.sql

# 003：理论文档与学习单元字段扩展
docker exec -i docker-postgres-1 psql -U postgres -d maguafly < ./migrations/003_theory_and_learning_fields.sql

# 004：任务与评论扩展及索引
docker exec -i docker-postgres-1 psql -U postgres -d maguafly < ./migrations/004_tasks_and_comments_fields.sql

# 005：可选增强（bases、日配置、学员进度 JSONB、训练记录扩展、签署人、理论分类）
docker exec -i docker-postgres-1 psql -U postgres -d maguafly < ./migrations/005_optional_enhancements.sql
```

**若容器名不是 `docker-postgres-1`**，把上面每条命令里的 `docker-postgres-1` 换成你 `docker ps` 里看到的 Postgres 容器名。

---

## 方式二：本机已安装 psql 时

你当前 compose 里 Postgres 映射到本机 **5437** 端口，可这样执行：

```bash
cd /Users/liuwei/magua-fly/magua-fly-golang

export PGPASSWORD=postgres

psql -h localhost -p 5437 -U postgres -d maguafly -f ./migrations/001_constraints_and_indexes.sql
psql -h localhost -p 5437 -U postgres -d maguafly -f ./migrations/002_training_scheduling_indexes.sql
psql -h localhost -p 5437 -U postgres -d maguafly -f ./migrations/003_theory_and_learning_fields.sql
psql -h localhost -p 5437 -U postgres -d maguafly -f ./migrations/004_tasks_and_comments_fields.sql
psql -h localhost -p 5437 -U postgres -d maguafly -f ./migrations/005_optional_enhancements.sql
```

---

## 一条命令依次执行全部（方式一）

```bash
cd /Users/liuwei/magua-fly/magua-fly-golang && \
docker exec -i docker-postgres-1 psql -U postgres -d maguafly < ./migrations/001_constraints_and_indexes.sql && \
docker exec -i docker-postgres-1 psql -U postgres -d maguafly < ./migrations/002_training_scheduling_indexes.sql && \
docker exec -i docker-postgres-1 psql -U postgres -d maguafly < ./migrations/003_theory_and_learning_fields.sql && \
docker exec -i docker-postgres-1 psql -U postgres -d maguafly < ./migrations/004_tasks_and_comments_fields.sql && \
docker exec -i docker-postgres-1 psql -U postgres -d maguafly < ./migrations/005_optional_enhancements.sql && \
echo "All 5 migrations completed."
```

任一条失败会中断，便于排查。

---

## 依赖与顺序

| 文件 | 依赖 |
|------|------|
| 001 | 表 `users`, `students`, `instructors`, `comments`, `tasks` 已存在 |
| 002 | 表 `training_records`, `training_archives`, `aircraft_availability`, `instructor_schedules`, `training_plans`, `bookings` 已存在 |
| 003 | 表 `theory_documents`, `theory_questions`, `theory_attempts`, `subject_learning`, `subject_questions`, `subject_attempts` 已存在 |
| 004 | 表 `tasks`, `comments` 已存在 |
| 005 | 001～004 涉及的表已存在；会创建 `bases`, `day_operational_config` |

必须先跑完 `000001_initial.up.sql`（或等价建表），再按 001 → 002 → 003 → 004 → 005 顺序执行。

---

## 011：飞机机时与维护窗口（UUID 库）

若使用 `008_uuid_schema.up.sql` 建库，应用启动时 `golang-migrate` 会执行 `011_aircraft_maintenance.up.sql`。也可手动执行：

```bash
docker exec -i docker-postgres-1 psql -U postgres -d maguafly < ./migrations/011_aircraft_maintenance.up.sql
```

为 `aircraft` 增加机时/定检字段，新增 `aircraft_maintenance_windows`，并为 `aircraft_availability` 增加 `maintenance_window_id`。

---

## 012：教员日排班（duty_status / duty_windows）

在 `008_uuid_schema` 之后执行 `012_instructor_duty.up.sql`。应用启动时若配置了 `migrate` 会自动执行。

- 新增枚举 `instructor_duty_status`；
- `instructor_schedules` 改为每教员每自然日一行：`duty_status`、`remark`、`duty_windows`（JSONB），去掉 `start_time`/`end_time`；
- 旧数据同日多段合并为 `duty_status = custom` 与 `duty_windows` 数组。

回滚：`012_instructor_duty.down.sql`（会丢失自定义多段语义，仅保留首段时间）。

---

## 013：业务日结束时间 19:00

执行 `013_business_day_end_19h.up.sql`（应用 migrate 会自动跑）。与 `internal/usecase/schedule.go` 中业务日 **09:00–19:00** 一致；并将超出该窗口的 **pending** 预约批量标记为 `rejected`。`down` 仅恢复列注释，不撤销对 `bookings` 的更新。

---

## 014：上下午分界 14:00

执行 `014_half_day_split_14h.up.sql`：更新 `instructor_schedules.duty_windows` 列注释，与代码中 **14:00** 上/下午分界一致（`morning` / `afternoon` 推导）。业务逻辑以应用代码为准。

---

## 常见问题

- **001 报错 `relation "users_username_unique" already exists` 或 unique 冲突**  
  建表时若已对 `users(username)` 建过 UNIQUE，可删掉 001 里「ADD CONSTRAINT users_username_unique」那一段，只保留 `users_email_unique` 和后续索引。
- **容器名**  
  若 compose 项目名不同，容器名可能是 `magua-fly-golang-postgres-1` 等，用 `docker ps` 查看后替换上述命令里的容器名，或执行前设置：  
  `export POSTGRES_CONTAINER=你的容器名`  
  再运行 `./migrations/run_extra_migrations.sh`。
