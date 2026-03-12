#!/usr/bin/env bash
# 在项目根目录执行：./migrations/run_extra_migrations.sh
# 或：bash migrations/run_extra_migrations.sh
# 执行 001～005 五个迁移（需先跑完 000001_initial.up.sql）

set -e
cd "$(dirname "$0")/.."
DB_CONTAINER="${POSTGRES_CONTAINER:-docker-postgres-1}"
DB_NAME="${POSTGRES_DB:-maguafly}"
DB_USER="${POSTGRES_USER:-postgres}"

echo "Using container: $DB_CONTAINER, database: $DB_NAME, user: $DB_USER"
for f in migrations/001_constraints_and_indexes.sql \
         migrations/002_training_scheduling_indexes.sql \
         migrations/003_theory_and_learning_fields.sql \
         migrations/004_tasks_and_comments_fields.sql \
         migrations/005_optional_enhancements.sql; do
  echo "Running $f ..."
  docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" < "$f"
done
echo "All 5 migrations completed."
