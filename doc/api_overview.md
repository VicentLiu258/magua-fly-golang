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
      "id": "550e8400-e29b-41d4-a716-446655440000",
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
    "id": "550e8400-e29b-41d4-a716-446655440000",
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
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "user_id": "550e8400-e29b-41d4-a716-446655440001",
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
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "user_id": "550e8400-e29b-41d4-a716-446655440002",
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
        "id": "550e8400-e29b-41d4-a716-446655440000",
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
  "aircraft_id": "550e8400-e29b-41d4-a716-446655440003",
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
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "instructor_id": "550e8400-e29b-41d4-a716-446655440002",
  "aircraft_id": "550e8400-e29b-41d4-a716-446655440003",
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
  "id": "550e8400-e29b-41d4-a716-446655440010",
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
        "id": "550e8400-e29b-41d4-a716-446655440000",
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
            "bookingId": "550e8400-e29b-41d4-a716-446655440010",
            "aircraftId": "550e8400-e29b-41d4-a716-446655440003",
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

---

## 7. 训练数据录入与管理

本模块接口**均为 POST**，需 `Authorization: Bearer <JWT>`，请求体 `Content-Type: application/json`，返回结构同上。

### 7.1 录入配置

#### 7.1.1 训练记录字段配置 `POST /api/v1/config/training-record-fields`

**请求体**：空 `{}` 或省略。

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "fields": [
      {
        "key": "student_id",
        "label": "学员",
        "type": "select",
        "required": true
      },
      {
        "key": "instructor_id",
        "label": "带飞教员",
        "type": "select",
        "required": true
      },
      {
        "key": "training_date",
        "label": "训练日期",
        "type": "date",
        "required": true
      },
      { "key": "engine_on_time", "label": "开车时间", "type": "time" },
      {
        "key": "status",
        "label": "状态",
        "type": "select",
        "options": ["normal", "cancelled", "deleted", "expired"]
      }
    ]
  }
}
```

### 7.2 训练记录 CRUD 与列表

#### 7.2.1 创建记录 `POST /api/v1/training-records`

**请求体**：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "instructor_id": "550e8400-e29b-41d4-a716-446655440002",
  "aircraft_id": "550e8400-e29b-41d4-a716-446655440003",
  "training_date": "2026-03-15",
  "engine_on_time": "2026-03-15T10:00:00Z",
  "engine_off_time": "2026-03-15T11:30:00Z",
  "flight_duration_minutes": 90,
  "is_local": true,
  "planned_duration_minutes": 90,
  "theory_duration_minutes": 0,
  "sim_duration_minutes": 0,
  "landings": 5,
  "syllabus_outline": "1#",
  "archive_display_no": "1#672",
  "status": "normal"
}
```

**成功响应**：`data` 为完整训练记录（含 `id`、`created_at`、`updated_at`）。

#### 7.2.2 记录详情 `POST /api/v1/training-records/detail`

**请求体**：`{ "id": "550e8400-e29b-41d4-a716-446655440011" }`

**成功响应**：`data` 为单条训练记录。

#### 7.2.3 更新记录 `POST /api/v1/training-records/update`

**请求体**：`{ "id": "550e8400-e29b-41d4-a716-446655440011", "student_id": "550e8400-e29b-41d4-a716-446655440001", "instructor_id": "550e8400-e29b-41d4-a716-446655440002", ... }`（需传 `id`，其余为可选更新字段）。

**成功响应**：`data` 为更新后的记录。

#### 7.2.4 删除记录（软删）`POST /api/v1/training-records/delete`

**请求体**：`{ "id": "550e8400-e29b-41d4-a716-446655440011" }`

**成功响应**：`data`: `{ "status": "deleted" }`。

#### 7.2.5 记录列表（多维筛选）`POST /api/v1/training-records/list`

**请求体**：

```json
{
  "view_id": null,
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "instructor_id": null,
  "date_from": "2026-01-01",
  "date_to": "2026-12-31",
  "status": "normal",
  "import_source": null,
  "sort_field": "training_date",
  "sort_asc": false,
  "limit": 20,
  "offset": 0
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [ { "id": "550e8400-e29b-41d4-a716-446655440000", "student_id": "550e8400-e29b-41d4-a716-446655440001", "instructor_id": "550e8400-e29b-41d4-a716-446655440002", "training_date": "2026-03-15", ... } ],
    "total": 100
  }
}
```

### 7.3 飞行数据导入与认领

#### 7.3.1 批量导入 `POST /api/v1/training-records/import`

**请求体**：

```json
{
  "batch_id": "batch-001",
  "rows": [
    {
      "engine_on_time": "2026-03-15T10:00:00Z",
      "engine_off_time": "2026-03-15T10:54:00Z",
      "flight_duration_minutes": 54,
      "is_local": true
    }
  ]
}
```

**成功响应**：`data`: `{ "batch_id": "batch-001", "items": [ { "id": "550e8400-e29b-41d4-a716-446655440000", "batch_id": "batch-001", ... } ] }`。

#### 7.3.2 认领导入记录 `POST /api/v1/training-records/claim`

**请求体**：

```json
{
  "raw_import_id": "550e8400-e29b-41d4-a716-446655440020",
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "instructor_id": "550e8400-e29b-41d4-a716-446655440002",
  "aircraft_id": "550e8400-e29b-41d4-a716-446655440003",
  "training_date": "2026-03-15"
}
```

**成功响应**：`data` 为创建的 `training_record`（及关联档案）。

#### 7.3.3 待认领列表 `POST /api/v1/training-records/import/pending`

**请求体**：`{ "limit": 50, "offset": 0 }` 或空。

**成功响应**：`data`: `{ "items": [ ... ] }`。

### 7.4 统计与仪表盘

#### 7.4.1 记录统计 `POST /api/v1/training-records/stats`

**请求体**：`{ "date_from": "2026-01-01", "date_to": "2026-12-31", "dimensions": [] }`

**成功响应**：`data`: `{ "count": 100, "total_flight_minutes": 9000 }` 等聚合结果。

#### 7.4.2 仪表盘概览 `POST /api/v1/dashboard/summary`

**请求体**：空或 `{}`。

**成功响应**：`data` 含今日/总计等概览字段。

### 7.5 自定义视图

- **保存视图** `POST /api/v1/training-views`  
  请求体：`{ "name": "我的视图", "filter_json": {}, "sort_json": {}, "group_by": "" }`。返回创建的视图。

- **视图列表** `POST /api/v1/training-views/list`  
  请求体：空。返回当前用户的 `items` 列表。

- **更新视图** `POST /api/v1/training-views/update`  
  请求体：`{ "id": "550e8400-e29b-41d4-a716-446655440000", "name": "...", "filter_json": {}, "sort_json": {}, "group_by": "" }`。

- **删除视图** `POST /api/v1/training-views/delete`  
  请求体：`{ "id": "550e8400-e29b-41d4-a716-446655440012" }`。

### 7.6 训练档案

#### 7.6.1 学员档案列表 `POST /api/v1/students/training-archives/list`

**请求体**：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "limit": 20,
  "offset": 0
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440101",
        "training_record_id": "550e8400-e29b-41d4-a716-446655440201",
        "student_id": "550e8400-e29b-41d4-a716-446655440001",
        "instructor_id": "550e8400-e29b-41d4-a716-446655440002",
        "training_date": "2026-03-15",
        "stage": "初级训练",
        "planned_duration_minutes": 90,
        "total_experience_minutes": 2635,
        "remaining_minutes": 1804,
        "engine_on_time": "2026-03-15T10:00:00Z",
        "engine_off_time": "2026-03-15T11:32:00Z",
        "flight_duration_minutes": 92,
        "progress_note": "",
        "instructor_review": "",
        "student_self_review": "",
        "task_assignment": "",
        "student_signed_at": null,
        "responsible_instructor_signed_at": null,
        "chief_instructor_signed_at": null,
        "created_at": "2026-03-15T12:00:00Z",
        "updated_at": "2026-03-15T12:00:00Z"
      }
    ]
  }
}
```

#### 7.6.2 档案详情 `POST /api/v1/training-archives/detail`

**请求体**：

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440101"
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440101",
    "training_record_id": "550e8400-e29b-41d4-a716-446655440201",
    "student_id": "550e8400-e29b-41d4-a716-446655440001",
    "instructor_id": "550e8400-e29b-41d4-a716-446655440002",
    "training_date": "2026-03-15",
    "stage": "初级训练",
    "planned_duration_minutes": 90,
    "total_experience_minutes": 2635,
    "remaining_minutes": 1804,
    "engine_on_time": "2026-03-15T10:00:00Z",
    "engine_off_time": "2026-03-15T11:32:00Z",
    "flight_duration_minutes": 92,
    "progress_note": "本课完成起落 5 次，航线保持良好。",
    "instructor_review": "起飞姿态有进步，落地需注意收油门时机。",
    "student_self_review": "对五边速度控制更有把握。",
    "task_assignment": "下次重点练习短五边与侧风落地。",
    "student_signed_at": "2026-03-15T14:00:00Z",
    "responsible_instructor_signed_at": "2026-03-15T14:05:00Z",
    "chief_instructor_signed_at": null,
    "responsible_instructor_id": "550e8400-e29b-41d4-a716-446655440002",
    "chief_instructor_id": null,
    "created_at": "2026-03-15T12:00:00Z",
    "updated_at": "2026-03-15T14:05:00Z"
  }
}
```

#### 7.6.3 创建档案 `POST /api/v1/training-archives`

**请求体**：

```json
{
  "training_record_id": "550e8400-e29b-41d4-a716-446655440201",
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "instructor_id": "550e8400-e29b-41d4-a716-446655440002",
  "training_date": "2026-03-15",
  "stage": "初级训练",
  "planned_duration_minutes": 90,
  "total_experience_minutes": 2635,
  "remaining_minutes": 1804,
  "engine_on_time": "2026-03-15T10:00:00Z",
  "engine_off_time": "2026-03-15T11:32:00Z",
  "flight_duration_minutes": 92,
  "progress_note": "",
  "instructor_review": "",
  "student_self_review": "",
  "task_assignment": ""
}
```

**成功响应**：`data` 为创建的档案对象（含 `id`、`created_at`、`updated_at`）。

#### 7.6.4 更新档案 `POST /api/v1/training-archives/update`

**请求体**：

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440101",
  "stage": "初级训练",
  "progress_note": "本课完成起落 5 次。",
  "instructor_review": "起飞姿态有进步，落地需注意收油门时机。",
  "student_self_review": "对五边速度控制更有把握。",
  "task_assignment": "下次重点练习短五边与侧风落地。",
  "planned_duration_minutes": 90,
  "total_experience_minutes": 2635,
  "remaining_minutes": 1804
}
```

**成功响应**：`data` 为更新后的完整档案对象。

#### 7.6.5 签字 `POST /api/v1/training-archives/sign`

**请求体**：

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440101",
  "role": "student"
}
```

`role` 取值：`student`（学员签字）、`responsible_instructor`（责任教员签字）、`chief_instructor`（主任教员签字）。

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "status": "signed"
  }
}
```

---

### 7.7 训练科目

#### 7.7.1 大纲科目列表 `POST /api/v1/syllabus-subjects/list`

**请求体**（可选带学员 ID，用于标记该学员已结业科目）：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

或不传 `student_id`：

```json
{}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "code": "SUB-01",
        "name": "起落",
        "sort_order": 1,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z",
        "is_graduated": false
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655440002",
        "code": "SUB-02",
        "name": "带飞",
        "sort_order": 2,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z",
        "is_graduated": true
      }
    ]
  }
}
```

#### 7.7.2 本课科目列表 `POST /api/v1/training-archives/subjects/list`

**请求体**：

```json
{
  "training_archive_id": "550e8400-e29b-41d4-a716-446655440101"
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440010",
        "training_archive_id": "550e8400-e29b-41d4-a716-446655440101",
        "syllabus_subject_id": "550e8400-e29b-41d4-a716-446655440300",
        "evaluation_score": 9,
        "is_graduated": false,
        "sort_order": 0,
        "subject_content": "本场起落 5 次",
        "type": "实践",
        "created_at": "2026-03-15T12:00:00Z",
        "updated_at": "2026-03-15T12:00:00Z"
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655440102",
        "training_archive_id": "550e8400-e29b-41d4-a716-446655440101",
        "syllabus_subject_id": "550e8400-e29b-41d4-a716-446655440301",
        "evaluation_score": 8,
        "is_graduated": false,
        "sort_order": 1,
        "subject_content": "带飞航线",
        "type": "实践",
        "created_at": "2026-03-15T12:00:00Z",
        "updated_at": "2026-03-15T12:00:00Z"
      }
    ]
  }
}
```

#### 7.7.3 设置本课科目 `POST /api/v1/training-archives/subjects/update`

**请求体**：

```json
{
  "training_archive_id": "550e8400-e29b-41d4-a716-446655440101",
  "subjects": [
    {
      "syllabus_subject_id": "550e8400-e29b-41d4-a716-446655440300",
      "evaluation_score": 9,
      "is_graduated": false,
      "subject_content": "本场起落 5 次",
      "type": "实践",
      "sort_order": 0
    },
    {
      "syllabus_subject_id": "550e8400-e29b-41d4-a716-446655440301",
      "evaluation_score": 8,
      "is_graduated": false,
      "subject_content": "带飞航线",
      "type": "实践",
      "sort_order": 1
    }
  ]
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "status": "ok"
  }
}
```

#### 7.7.4 下次布置科目列表 `POST /api/v1/training-archives/next-subjects/list`

**请求体**：

```json
{
  "training_archive_id": "550e8400-e29b-41d4-a716-446655440101"
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440501",
        "training_archive_id": "550e8400-e29b-41d4-a716-446655440101",
        "syllabus_subject_id": "550e8400-e29b-41d4-a716-446655440302",
        "assigned_by": "550e8400-e29b-41d4-a716-446655440002",
        "created_at": "2026-03-15T14:00:00Z"
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655440502",
        "training_archive_id": "550e8400-e29b-41d4-a716-446655440101",
        "syllabus_subject_id": "550e8400-e29b-41d4-a716-446655440303",
        "assigned_by": "550e8400-e29b-41d4-a716-446655440002",
        "created_at": "2026-03-15T14:00:00Z"
      }
    ]
  }
}
```

#### 7.7.5 布置下次科目 `POST /api/v1/training-archives/next-subjects/update`

需登录，以当前用户为分配人（`assigned_by`）。

**请求体**：

```json
{
  "training_archive_id": "550e8400-e29b-41d4-a716-446655440101",
  "syllabus_subject_ids": [ 3, 4, 5 ]
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "status": "ok"
  }
}
```

---

### 8. 理论学习模块 API

#### 8.1 理论文档列表 `POST /api/v1/theory/documents/list`

- **说明**：返回当前系统内可用的理论文档列表，可按关键字、分类筛选。

**请求体**：

```json
{
  "keyword": "",
  "category": "",
  "status": "published",
  "page": 1,
  "page_size": 20
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440201",
        "title": "失速与尾旋理论",
        "category": "空气动力学",
        "sort_order": 10,
        "status": "published",
        "has_questions": true,
        "has_video": true,
        "my_latest_score": 86,
        "my_attempt_count": 3
      }
    ],
    "total": 1
  }
}
```

#### 8.2 理论文档详情 `POST /api/v1/theory/documents/detail`

- **说明**：返回单个理论文档的元数据与内容块（图文 + 视频占位），不包含题目。

**请求体**：

```json
{
  "document_id": "550e8400-e29b-41d4-a716-446655440201"
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "document": {
      "id": "550e8400-e29b-41d4-a716-446655440201",
      "title": "失速与尾旋理论",
      "category": "空气动力学",
      "status": "published",
      "created_by": "550e8400-e29b-41d4-a716-446655440002",
      "created_at": "2026-03-16T00:00:00Z"
    },
    "content": [
      {
        "type": "text",
        "content": "富文本内容..."
      },
      {
        "type": "image",
        "url": "https://cdn.example.com/images/stall.png",
        "caption": "失速示意图"
      },
      {
        "type": "video_embed",
        "video_id": "550e8400-e29b-41d4-a716-446655440901",
        "position": 0
      }
    ]
  }
}
```

#### 8.3 开始答题 `POST /api/v1/theory/questions/start`

- **说明**：点击“答题”后调用，初始化一次答题尝试并返回题目列表。

**请求体**：

```json
{
  "document_id": "550e8400-e29b-41d4-a716-446655440201"
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "attempt_id": "550e8400-e29b-41d4-a716-446655441001",
    "time_limit_seconds": 1800,
    "questions": [
      {
        "id": "q1",
        "type": "single",
        "stem": "下列关于失速的说法，正确的是？",
        "options": [
          { "id": "A", "text": "与迎角无关" },
          { "id": "B", "text": "与迎角有关" }
        ],
        "score": 2
      }
    ]
  }
}
```

#### 8.4 交卷并评分 `POST /api/v1/theory/questions/submit`

- **说明**：提交本次答题的所有题目选项，系统自动评分。

**请求体**：

```json
{
  "attempt_id": "550e8400-e29b-41d4-a716-446655441001",
  "answers": [
    {
      "question_id": "q1",
      "selected_option_ids": ["B"]
    }
  ]
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "score": 86,
    "max_score": 100,
    "correct_count": 43,
    "total_count": 50,
    "passed": true
  }
}
```

#### 8.5 历史成绩列表 `POST /api/v1/theory/attempts/my-list`

- **说明**：返回当前登录学员在某理论文档下的历史答题记录。

**请求体**：

```json
{
  "document_id": "550e8400-e29b-41d4-a716-446655440201",
  "page": 1,
  "page_size": 20
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "550e8400-e29b-41d4-a716-446655441001",
        "user_id": "550e8400-e29b-41d4-a716-446655440003",
        "theory_document_id": "550e8400-e29b-41d4-a716-446655440201",
        "score": 86,
        "submitted_at": "2026-03-16T10:00:00Z",
        "duration_seconds": 900,
        "max_score": 100,
        "correct_count": 43,
        "total_count": 50,
        "passed": true
      }
    ],
    "total": 3
  }
}
```

---

### 8.x 理论文档管理 API

#### 8.x.1 创建文档 `POST /api/v1/theory/documents/create`

- **说明**：教员/管理员创建理论文档，`id` 由后端自动生成 UUID。

**请求体**：

```json
{
  "title": "失速与尾旋理论",
  "slug": "stall-spin-theory",
  "sort_order": 10,
  "category": "空气动力学",
  "video_url": "https://video.example.com/stall.mp4"
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440201",
    "title": "失速与尾旋理论",
    "slug": "stall-spin-theory",
    "sort_order": 10,
    "content_ref": "",
    "video_url": "https://video.example.com/stall.mp4",
    "category": "空气动力学",
    "created_at": "2026-03-16T00:00:00Z",
    "updated_at": "2026-03-16T00:00:00Z"
  }
}
```

> 说明：返回中的 `id` 即为后续题库录入（`document_id`）、试卷管理、答题接口使用的理论文档 ID。

#### 8.x.2 更新文档 `POST /api/v1/theory/documents/update`

**请求体**：

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440201",
  "title": "失速与尾旋理论（更新后）",
  "slug": "stall-spin-theory",
  "sort_order": 20,
  "category": "空气动力学",
  "video_url": "https://video.example.com/stall-v2.mp4"
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440201",
    "title": "失速与尾旋理论（更新后）",
    "slug": "stall-spin-theory",
    "sort_order": 20,
    "content_ref": "",
    "video_url": "https://video.example.com/stall-v2.mp4",
    "category": "空气动力学",
    "created_at": "2026-03-16T00:00:00Z",
    "updated_at": "2026-03-16T01:00:00Z"
  }
}
```

#### 8.x.3 删除文档 `POST /api/v1/theory/documents/delete`

**请求体**：

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440201"
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "status": "ok"
  }
}
```

---

### 9. 学员任务模块 API

#### 9.1 我的任务列表 `POST /api/v1/tasks/my-list`

- **说明**：学员在“我的任务”界面查看自己的所有任务（理论、科目、其他）。

**请求体**：

```json
{
  "status": "pending",
  "page": 1,
  "page_size": 50
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "550e8400-e29b-41d4-a716-446655442001",
        "type": "theory",
        "ref_id": "550e8400-e29b-41d4-a716-446655440201",
        "title": "失速与尾旋理论学习",
        "description": "阅读文档并完成随堂测试。",
        "due_at": "2026-03-20T00:00:00Z",
        "status": "pending",
        "related_entity": {
          "kind": "theory_doc",
          "display_title": "失速与尾旋理论"
        }
      }
    ],
    "total": 1
  }
}
```

#### 9.2 更新任务状态 `POST /api/v1/tasks/update-status`

- **说明**：学员执行任务后将状态标记为进行中/已完成，也可由系统在完成答题后自动调用。

**请求体**：

```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655442001",
  "status": "done"
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "status": "ok"
  }
}
```

---

### 10. 题库管理 API

#### 10.1 列出题目（后台）`POST /api/v1/theory/questions/admin/list`

- **说明**：教员/管理员查看某理论文档下的全部题目，用于后台题库维护。

**请求体**：

```json
{
  "document_id": "<UUID>",
  "page": 1,
  "page_size": 50,
  "keyword": "",
  "tags": [],
  "type": ""
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "q1",
        "type": "single",
        "stem": "下面哪一项与失速最直接相关？",
        "options": [
          { "id": "A", "text": "飞机重量" },
          { "id": "B", "text": "迎角" }
        ],
        "tags": ["基础"],
        "score": 2
      }
    ],
    "total": 1
  }
}
```

#### 10.2 创建题目 `POST /api/v1/theory/questions/admin/create`

**请求体**：

```json
{
  "document_id": "<UUID>",
  "question": {
    "id": "",                      // 可留空，由后端生成
    "type": "single",              // single | multi
    "stem": "下面哪一项与失速最直接相关？",
    "options": [
      { "id": "A", "text": "飞机重量" },
      { "id": "B", "text": "迎角" },
      { "id": "C", "text": "油量" }
    ],
    "correct_option_ids": ["B"],
    "explanation": "失速本质上是迎角过大导致升力丧失。",
    "tags": ["基础"],
    "score": 2
  }
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": { "status": "ok" }
}
```

#### 10.3 更新题目 `POST /api/v1/theory/questions/admin/update`

**请求体**：

```json
{
  "document_id": "<UUID>",
  "question": {
    "id": "q1",
    "type": "single",
    "stem": "修改后的题干...",
    "options": [
      { "id": "A", "text": "修改后的选项A" },
      { "id": "B", "text": "修改后的选项B(正确)" }
    ],
    "correct_option_ids": ["B"],
    "explanation": "修改后的解析...",
    "tags": ["进阶"],
    "score": 3
  }
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": { "status": "ok" }
}
```

#### 10.4 删除题目 `POST /api/v1/theory/questions/admin/delete`

**请求体**：

```json
{
  "document_id": "<UUID>",
  "question_id": "q1"
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": { "status": "ok" }
}
```

---

### 11. 试卷管理 API

#### 11.1 列出试卷 `POST /api/v1/theory/papers/list`

- **说明**：查看某理论文档下已配置的试卷列表，用于后台维护和前端出卷选择。

**请求体**：

```json
{
  "document_id": "<UUID>",
  "page": 1,
  "page_size": 20
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "paper-uuid-1",
        "theory_document_id": "<UUID>",
        "name": "阶段测验 1",
        "description": "用于阶段性考核",
        "total_score": 100,
        "enabled": true,
        "created_by": "<user_id>",
        "created_at": "2026-03-16T00:00:00Z",
        "updated_at": "2026-03-16T00:00:00Z"
      }
    ],
    "total": 1
  }
}
```

#### 11.2 创建/更新试卷 `POST /api/v1/theory/papers/save`

**请求体**：

```json
{
  "id": null,                        // 为空表示创建；非空表示更新
  "theory_document_id": "<UUID>",
  "name": "阶段测验 1",
  "description": "用于阶段性考核",
  "enabled": true,
  "questions": [
    { "question_id": "q1", "score": 2, "sort_order": 1 },
    { "question_id": "q2", "score": 3, "sort_order": 2 }
  ]
}
```

**成功响应**（返回试卷基本信息）：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": "paper-uuid-1",
    "theory_document_id": "<UUID>",
    "name": "阶段测验 1",
    "description": "用于阶段性考核",
    "total_score": 100,
    "enabled": true,
    "created_by": "<user_id>",
    "created_at": "2026-03-16T00:00:00Z",
    "updated_at": "2026-03-16T00:00:00Z"
  }
}
```

#### 11.3 删除试卷 `POST /api/v1/theory/papers/delete`

**请求体**：

```json
{
  "paper_id": "<UUID>"
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": { "status": "ok" }
}
```

#### 11.4 使用试卷出题：扩展 `/api/v1/theory/questions/start`

- **说明**：在原有 `document_id` 的基础上，增加可选字段 `paper_id`；若提供，则按照试卷配置出题。

**请求体**：

```json
{
  "document_id": "<UUID>",
  "paper_id": "paper-uuid-1"
}
```

**行为差异**：

- 若携带 `paper_id`：\n  - 从 `theory_paper_questions` 中按 `sort_order` 取出该卷的题目 ID 和分值。\n  - 在 Mongo 题库 `theory_question_sets` 中按 `question_id` 拉取题干与选项，并将 `score` 覆盖为试卷中配置的分值。\n- 若不携带 `paper_id`：\n  - 保持原有行为，直接使用该文档下的全部题目作为一套卷。


