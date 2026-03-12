# 飞校平台 (Magua Fly)

Golang 后端：PostgreSQL + MongoDB，Docker 部署。

## 技术栈

- **运行时**: Go 1.21
- **数据库**: PostgreSQL 15, MongoDB 7
- **HTTP**: Gin
- **认证**: JWT
- **迁移**: golang-migrate

## 本地开发

```bash
# 依赖
go mod tidy

# 需本地已启动 Postgres、Mongo（或使用 docker-compose 只起数据库）
export POSTGRES_DSN="postgres://postgres:postgres@localhost:5432/maguafly?sslmode=disable"
export MONGODB_URI="mongodb://localhost:27017"
export JWT_SECRET=dev-secret

# 运行
go run ./cmd/server
```

服务默认 `http://localhost:8080`。

## Docker 部署

```bash
cd docker
docker compose up -d
```

- 应用: http://localhost:8080
- Postgres: localhost:5432
- MongoDB: localhost:27017

环境变量见 `docker/docker-compose.yml`，生产请修改 `JWT_SECRET` 与数据库密码。

## API 概览

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /health | 健康检查 |
| POST | /auth/login | 登录 |
| POST | /auth/register | 注册 |
| GET | /students | 学员列表（需鉴权） |
| GET | /instructors | 教员列表（需鉴权） |
| GET | /aircraft | 飞机列表（需鉴权） |
| GET | /bookings/pending | 待审核预约（需鉴权） |
| POST | /bookings | 创建预约（需鉴权） |
| POST | /bookings/:id/review | 审核预约（需鉴权） |

鉴权：请求头 `Authorization: Bearer <token>`。

## 数据库

- **PostgreSQL**: 用户、学员、教员、飞机、排期、预约、训练记录与档案、大纲科目、任务、理论文档元数据等，见 `migrations/`.
- **MongoDB**: 理论文档富文本、评论、原始导入数据等，见方案文档。

## 项目结构

```
cmd/server/          # 入口
internal/
  config/            # 配置
  domain/            # 领域模型
  handler/           # HTTP 处理与路由
  middleware/        # 鉴权、CORS、日志
  mongodb/           # MongoDB 连接与仓储
  postgres/          # PostgreSQL 连接、迁移、仓储
  repository/        # 仓储接口
  usecase/           # 业务用例
  pkg/               # 密码、JWT
migrations/          # SQL 迁移
docker/              # Dockerfile、docker-compose
```
