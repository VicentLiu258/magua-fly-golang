# 接口总览文档（magua-fly-golang）

本文汇总当前后端已实现的主要 HTTP 接口，按业务模块分组，统一说明：

- **认证方式**：除 `/health`、`/auth/*` 外，其余接口均需 `Authorization: Bearer <JWT>`。
- **请求格式**：`Content-Type: application/json`，方法均为 `POST`（除 `/health` 为 `GET`）。
- **统一返回结构**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {}
}
```

- `code`：业务状态码，与 HTTP 状态语义保持一致，例如 `200/400/401/500/503`。
- `msg`：简短文字描述。
- `data`：实际业务数据；错误时通常为 `null` 或携带少量上下文信息。

---

## 1. 健康检查

### 1.1 检测服务与数据库状态

- **方法**：`GET /health`
- **认证**：不需要。

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": null
}
```

**数据库异常示例**：

```json
{
  "code": 503,
  "msg": "postgres unhealthy",
  "data": { "detail": "error detail ..." }
}
```

---

## 2. 认证与用户

### 2.1 用户名密码登录

- **方法**：`POST /auth/login`
- **认证**：不需要。

**请求体**：

```json
{
  "username": "user1",
  "password": "secret"
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "token": "jwt-token",
    "user": {
      "id": 1,
      "username": "user1",
      "phone": "13800000000",
      "email": "u@example.com",
      "role": "student",
      "avatar_url": "",
      "created_at": "2026-03-10T12:00:00Z",
      "updated_at": "2026-03-10T12:00:00Z"
    }
  }
}
```

**失败示例**（账号或密码错误）：

```json
{
  "code": 401,
  "msg": "invalid credentials",
  "data": null
}
```

### 2.2 注册账号

- **方法**：`POST /auth/register`
- **认证**：不需要。

**请求体**：

```json
{
  "username": "student1",
  "password": "secret",
  "phone": "13800000000",
  "email": "s1@example.com",
  "role": "student" // admin | staff | instructor | student
}
```

**行为说明**：

- 创建一条 `users` 记录。
- 若 `role = "student"`，自动在 `students` 表创建一条记录并关联 `user_id`。
- 若 `role = "instructor"`，自动在 `instructors` 表创建一条记录并关联 `user_id`。

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": 1,
    "username": "student1",
    "phone": "13800000000",
    "email": "s1@example.com",
    "role": "student",
    "avatar_url": "",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

---

## 3. 短信登录 / 注册

详细协议见 `doc/auth_sms_api.md`，这里只给出摘要。

### 3.1 发送短信验证码

- **方法**：`POST /auth/sms/send`

**请求体**：

```json
{
  "phone": "13800000000",
  "scene": "login" // register | login | reset_pwd | 2fa
}
```

### 3.2 校验验证码并登录 / 注册

- **方法**：`POST /auth/sms/verify`

**请求体**：

```json
{
  "phone": "13800000000",
  "scene": "login",
  "code": "123456"
}
```

**成功响应（登录/注册）**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "token": "jwt-token",
    "user": { ...同登录返回... }
  }
}
```

短信注册场景下，会自动：

- 创建 `users` 记录（`role = "student"`）。
- 自动创建并关联 `students` 记录。

---

## 4. 学员、教员、飞机基础信息

### 4.1 学员列表 `POST /api/v1/students/list`

- **认证**：需要。

**请求体**（使用 query 参数解析，body 可为空）：

```http
POST /api/v1/students/list?limit=20&offset=0
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "user_id": 1,
        "name": "student1",
        "student_no": "S000001",
        "flown_minutes": 0,
        "training_progress": "",
        "created_at": "...",
        "updated_at": "..."
      }
    ]
  }
}
```

### 4.2 教员列表 `POST /api/v1/instructors/list`

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "user_id": 2,
        "name": "Instructor A",
        "display_color": "#ff0000",
        "created_at": "...",
        "updated_at": "..."
      }
    ]
  }
}
```

### 4.3 飞机列表 `POST /api/v1/aircraft/list`

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "code": "AC001",
        "name": "塞斯纳 172",
        "status": "active",
        "created_at": "...",
        "updated_at": "..."
      }
    ]
  }
}
```

### 4.4 创建飞机 `POST /api/v1/aircraft/create`

**请求体**：

```json
{
  "code": "AC001",
  "name": "塞斯纳 172",
  "status": "active"
}
```

**成功响应**同上单条飞机结构。

### 4.5 配置飞机可用时段 `POST /api/v1/aircraft/availabilities/create`

**请求体**：

```json
{
  "aircraft_id": 1,
  "date": "2026-03-15",
  "start_time": "09:00",
  "end_time": "21:00",
  "status": "available" // available | maintenance | blocked
}
```

**成功响应**：返回 `aircraft_availability` 记录详情。

---

## 5. 预约与排期

更详细的说明见 `doc/schedule_api.md`，这里给出接口摘要。

### 5.1 创建预约 `POST /api/v1/bookings`

**请求体**：

```json
{
  "student_id": 1,
  "instructor_id": 1,
  "aircraft_id": 1,
  "plan_date": "2026-03-15",
  "start_time": "10:00",
  "end_time": "11:30"
}
```

要求：

- `student_id` 必须是 `students.id`（不是 `users.id`）。
- `instructor_id` 必须是 `instructors.id`。
- `aircraft_id` 必须是 `aircraft.id`。

**成功响应**：`data` 为 `booking` 记录。

### 5.2 查询待审核预约 `POST /api/v1/bookings/pending`

返回 `status = "pending"` 的预约列表，结构见 `schedule_api.md` 第 3.2 节。

### 5.3 审核预约 `POST /api/v1/bookings/review`

**请求体**：

```json
{
  "id": 1001,
  "approve": true,
  "reason": "可选说明"
}
```

- 通过时会：
  - 将 `bookings.status` 设为 `approved`。
  - 创建对应的 `training_plans` 记录。

**响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": { "status": "approved" }
}
```

### 5.4 教员 × 时间排期网格 `POST /api/v1/schedules/instructor-grid`

详见 `doc/schedule_api.md` 第 2 节，这里只列接口：

**请求体**：

```json
{ "date": "2026-03-15" }
```

**响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "date": "2026-03-15",
    "slotMinutes": 30,
    "dayStart": "09:00",
    "dayEnd": "21:00",
    "slots": ["09:00", "09:30", "..."],
    "instructors": [
      {
        "id": 1,
        "name": "教员A",
        "displayColor": "#ff0000",
        "timeline": [
          {
            "slotIndex": 0,
            "startTime": "09:00",
            "endTime": "09:30",
            "status": "available"
          },
          {
            "slotIndex": 1,
            "startTime": "09:30",
            "endTime": "10:00",
            "status": "booked",
            "bookingId": 1001,
            "aircraftId": 1,
            "studentId": 1
          }
        ]
      }
    ]
  }
}
```

---

## 6. 统一错误示例

### 6.1 参数错误

```json
{
  "code": 400,
  "msg": "invalid start_time",
  "data": null
}
```

### 6.2 认证失败

```json
{
  "code": 401,
  "msg": "unauthorized",
  "data": null
}
```

### 6.3 外键约束 / 数据库错误

```json
{
  "code": 500,
  "msg": "ERROR: insert or update on table \"bookings\" violates foreign key constraint \"bookings_aircraft_id_fkey\" (SQLSTATE 23503)",
  "data": null
}
```

一般表示请求中的 ID 未在对应主表中存在，需要先通过创建接口或 SQL 插入基础数据。
