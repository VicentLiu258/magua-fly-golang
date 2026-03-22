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
        "total_airframe_hours": 120.5,
        "last_inspection_hours": 100.0,
        "inspection_interval_hours": 50,
        "created_at": "...",
        "updated_at": "..."
      }
    ]
  }
}
```

`total_airframe_hours`、`last_inspection_hours` 可为 `null`（未录入时）；`inspection_interval_hours` 默认 `50`。

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
  "end_time": "19:00",
  "status": "available" // available | maintenance | blocked
}
```

**成功响应**：返回 `aircraft_availability` 记录详情（若由维护计划生成，可能含 `maintenance_window_id`）。

#### 检修排期 vs 机时参考（约定）

- **检修占用时段**：以人工创建/维护的 **维护时间窗**（§4.8～4.11）及其生成的 `aircraft_availability`（`status=maintenance`）为**唯一排期依据**；排班网格与预约冲突处理均以此为标准（见 §5.4）。
- **机时与定检间隔**（§4.6～4.7）：`total_airframe_hours`、`last_inspection_hours`、`inspection_interval_hours` 等字段**仅作参考与提醒**（`hours_reference_role` 恒为 `advisory`），**不会**根据机时自动生成维护窗，也**不会**因机时“到期”而自动拒绝预约；具体检修时段须由运维在维护窗接口中配置。

### 4.6 机时与定检查询 `POST /api/v1/aircraft/airworthiness/get`

**请求体**：

```json
{ "aircraft_id": "550e8400-e29b-41d4-a716-446655440003" }
```

**成功响应** `data`：

```json
{
  "hours_reference_role": "advisory",
  "aircraft_id": "550e8400-e29b-41d4-a716-446655440003",
  "total_airframe_hours": 120.5,
  "last_inspection_hours": 100.0,
  "inspection_interval_hours": 50,
  "next_due_hours": 150.0,
  "remaining_hours": 29.5,
  "reminder_level": "info",
  "reminder_hint": "机时定检参考：剩余约 29.5 小时"
}
```

- `hours_reference_role`：固定为 `advisory`，表示机时相关字段**不驱动**自动排期，仅供展示与提醒。
- `next_due_hours` = `last_inspection_hours` + `inspection_interval_hours`（需已录入上次定检小时）；`remaining_hours` = `next_due_hours` - `total_airframe_hours`。若缺少必要字段，对应派生字段可为 `null`。
- `reminder_level`：`none` | `info` | `warning` | `overdue`。在能计算 `remaining_hours` 时，按剩余小时分段：`>15` 为 `none`；`(5,15]` 为 `info`；`(0,5]` 为 `warning`；`<=0` 为 `overdue`。无 `remaining_hours` 时为 `none`。
- `reminder_hint`：与级别配套的中文短句，便于前端直接展示；`reminder_level` 为 `none` 时通常为空。

### 4.7 机时与定检更新 `POST /api/v1/aircraft/airworthiness/update`

**请求体**（字段均可选，仅提交需要修改的项）：

```json
{
  "aircraft_id": "550e8400-e29b-41d4-a716-446655440003",
  "total_airframe_hours": 121.0,
  "last_inspection_hours": 100.0,
  "inspection_interval_hours": 50
}
```

**成功响应**：`data` 同 4.6 查询结构（更新后的派生信息）。

### 4.8 维护时间窗列表 `POST /api/v1/aircraft/maintenance-windows/list`

**请求体**：

```json
{
  "aircraft_id": "550e8400-e29b-41d4-a716-446655440003",
  "from_date": "2026-03-01",
  "to_date": "2026-03-31",
  "status": "planned"
}
```

`from_date`、`to_date`、`status` 均可选。`status`：`planned` | `in_progress` | `completed` | `cancelled`。

**成功响应**：`data.items` 为维护窗数组，字段含 `id`、`aircraft_id`、`kind`、`scheduled_start_at`、`scheduled_end_at`（RFC3339）、`status`、`hangar_location`、`notes`、`created_at`、`updated_at`。

`kind` 为业务分类字符串：**新建时若省略，默认为 `manual`**（人工排定检修时段）。常见取值示例：`manual`（默认）、`periodic_inspection`（周期定检类备注）、其它由业务自定义；**不改变**「排期由人工维护窗决定」的规则。

### 4.9 创建维护时间窗 `POST /api/v1/aircraft/maintenance-windows/create`

**检修时段以本接口人工配置为准**：写入 `aircraft_maintenance_windows`，按自然日拆分写入 `aircraft_availability`（`status=maintenance`，并关联 `maintenance_window_id`）。机时定检参考信息（§4.6）仅可在 `notes` 中作为**备注**说明，例如「参考 50 小时定检窗口」，**系统不会**根据机时自动创建或修改本记录。

**请求体**：

```json
{
  "aircraft_id": "550e8400-e29b-41d4-a716-446655440003",
  "kind": "manual",
  "scheduled_start_at": "2026-03-21T08:00:00+08:00",
  "scheduled_end_at": "2026-03-21T19:00:00+08:00",
  "hangar_location": "一号机库",
  "notes": "明日进机库检修（机时定检参考见机务记录）",
  "sync_schedule": true
}
```

- `kind`：可选；省略时默认为 `manual`。
- `scheduled_start_at`：**必填**，RFC3339。
- `scheduled_end_at`：可选；省略则取开始日 **当地** 23:59:59。
- `sync_schedule`：为 `true` 时，将与该时间窗重叠的 `pending`/`approved` 预约置为 `rejected`（原因：飞机维护/定检），并对 `approved` 预约删除对应 `training_plans` 中与时段匹配的一条记录。

**成功响应** `data`：

```json
{
  "window": { "...": "..." },
  "affected_booking_ids": ["..."],
  "cancelled_training_plan_ids": ["..."]
}
```

### 4.10 更新维护时间窗 `POST /api/v1/aircraft/maintenance-windows/update`

**请求体**：

```json
{
  "id": "维护窗 UUID",
  "aircraft_id": "",
  "kind": "manual",
  "scheduled_start_at": "2026-03-21T08:00:00+08:00",
  "scheduled_end_at": "2026-03-21T18:00:00+08:00",
  "hangar_location": "",
  "notes": "",
  "sync_schedule": true
}
```

`aircraft_id` 可留空（表示不修改所属飞机，须与原有记录一致）。`kind` 可留空，留空时按 **`manual`** 写入。行为：删除该窗下原 `aircraft_availability` 行后按新时间重写，并在 `sync_schedule` 为 `true` 时再次执行与 4.9 相同的预约同步逻辑。

### 4.11 取消维护时间窗 `POST /api/v1/aircraft/maintenance-windows/cancel`

**请求体**：

```json
{ "id": "维护窗 UUID" }
```

将状态置为 `cancelled`，并删除关联的 `aircraft_availability` 记录（不自动恢复已拒绝的预约）。

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
- 教员当日须已有排班，且 `[start_time, end_time)` 完全落在某一执勤窗口内（与 §5.5 规则一致）；否则 `400`。

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

网格中 `available` 时段由教员**日排班**解析得到：业务日默认为 **09:00–19:00**、上下午以 **14:00** 分界（`morning` / `afternoon` 推导规则），与 `duty_status` / `duty_windows` 的解析规则一致（见 §5.5）。教员当日**无排班记录**时，该教员时间轴全部为 `off`。`19:00` 之后无格点；历史预约若结束时间晚于 `19:00`，网格仅覆盖至当日最后一格（半开区间与 `dayEnd` 对齐）。

**补充**：当某日某飞机存在 `aircraft_availability.status = maintenance` 且与预约时段重叠时，对应格子的 `status` 为 `aircraft_unavailable`（仍带原 `bookingId` / `aircraftId` / `studentId`，表示该预约与维护冲突需在业务上处理）。维护时段来源见 §4.8～4.11（人工维护窗），**与** §4.6 **机时参考提醒无关**。

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
    "dayEnd": "19:00",
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

### 5.5 教员日排班

`instructor_schedules` 每教员每自然日**最多一行**。`duty_status` 取值：`morning`（上午出勤）、`afternoon`（下午出勤）、`full_day`（全天）、`leave`（请假）、`rest`（休息）、`custom`（自定义多段，须提供 `duty_windows`）。`duty_windows` 为 `null` 时，由 `duty_status` 与业务日 **09:00–19:00**、**14:00** 上下午分界推导执勤窗口；非空时为 JSON 数组 `[{"start":"HH:MM","end":"HH:MM"},...]`，并裁剪到业务日内。

学员创建预约 `POST /api/v1/bookings` 时，预约时段须**完全落在**教员当日某一执勤半开区间内；否则返回 `400`。教员当日无排班记录时不可预约。

#### 5.5.1 排班列表 `POST /api/v1/schedules/instructor-duties/list`

**请求体**：

```json
{
  "from_date": "2026-03-01",
  "to_date": "2026-03-31",
  "instructor_ids": []
}
```

`instructor_ids` 省略或空数组表示不限教员。

**响应 `data.items[]`**：每条对应库中已有的一行排班，包含 `instructor_name`、`duty_status`、`remark`、`duty_windows`（库内原始值）、`resolved_windows`（解析后的展示窗口）、`flight_tasks`（由当日的 `bookings` 与 `training_plans` 聚合，相同学员/起止时刻去重，优先保留预约）、`ground_courses`（当前为空数组，结构与 `flight_tasks` 一致，预留地面课）。

#### 5.5.2 单日写入/更新 `POST /api/v1/schedules/instructor-duties/upsert`

**请求体**：

```json
{
  "instructor_id": "550e8400-e29b-41d4-a716-446655440000",
  "date": "2026-03-15",
  "duty_status": "full_day",
  "remark": "可选",
  "duty_windows": [{ "start": "09:00", "end": "14:00" }]
}
```

`duty_windows` 可省略；`duty_status` 为 `custom` 时须提供非空 `duty_windows`。

#### 5.5.3 按区间批量写入 `POST /api/v1/schedules/instructor-duties/bulk`

**请求体**：

```json
{
  "instructor_id": "550e8400-e29b-41d4-a716-446655440000",
  "from_date": "2026-03-01",
  "to_date": "2026-03-31",
  "weekdays": [1, 2, 3, 4, 5],
  "duty_status": "full_day",
  "remark": "",
  "replace_existing": true
}
```

`weekdays` 使用 **Go `time.Weekday`**：**0=周日 … 6=周六**。`replace_existing` 为 `false` 时跳过已有排班的日期；为 `true` 时覆盖。

**响应**：`data.upserted` 为实际写入的天数。

### 5.6 工作人员考勤网格与 staff 排班

面向地勤、行政等 **staff** 账号的考勤排班；**教员**行仍读取 `instructor_schedules`（与 §5.4 规则一致），**不可预约**、无 `bookings` 覆盖层。

#### 5.6.1 全员考勤网格 `POST /api/v1/schedules/workforce-grid`

返回 **role 为 `staff` 与 `instructor` 的用户**（不含 `admin`、`student`），格点定义与教员网格相同（`dayStart`/`dayEnd`/`slots`/`slotMinutes`）。

**请求体**：

```json
{ "date": "2026-03-15" }
```

**响应 `data`**：

- `workers[]`：每人一行。
  - `userId`、`role`（`staff` | `instructor`）、`username`、`name`（教员取 `instructors.name`，staff 取 `username`）、`instructorId`（仅教员有）、`displayColor`（教员可选）、`timeline[]`。
  - `timeline[].status` 仅为 `off` 或 `available`（考勤在岗），**无** `booked` / `aircraft_unavailable`。
- 教员行的执勤窗口来自 **`instructor_schedules`**；纯 staff 行来自 **`staff_schedules`**（按 `users.id`）。

#### 5.6.2 staff 排班列表 `POST /api/v1/schedules/staff-duties/list`

仅列出 **`staff_schedules`** 中记录（`user_id` 对应 `role=staff` 的用户）。

**请求体**：`from_date`、`to_date`、`user_ids`（可选，空表示不限）。

**响应**：`data.items[]` 含 `user_id`、`username`、`name`、`duty_status`、`resolved_windows` 等（无带飞任务字段）。

#### 5.6.3 staff 单日 upsert `POST /api/v1/schedules/staff-duties/upsert`

**请求体**：`user_id`（须为 **staff** 角色）、`date`、`duty_status`、`remark`、`duty_windows`（可选，语义同教员）。

#### 5.6.4 staff 批量 `POST /api/v1/schedules/staff-duties/bulk`

**请求体**：`user_id`（staff）、`from_date`、`to_date`、`weekdays`、`duty_status`、`remark`、`replace_existing`（与教员 bulk 一致）。

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
  "syllabus_subject_ids": [3, 4, 5]
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

#### 7.7.6 创建大纲科目（后台）`POST /api/v1/syllabus-subjects/admin/create`

- **说明**：用于后台或教员创建新的训练大纲科目。

**请求体**：

```json
{
  "code": "SUB-01",
  "name": "起落",
  "sort_order": 1
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "SUB-01",
    "name": "起落",
    "sort_order": 1,
    "created_at": "2026-03-16T00:00:00Z",
    "updated_at": "2026-03-16T00:00:00Z"
  }
}
```

#### 7.7.7 更新大纲科目（后台）`POST /api/v1/syllabus-subjects/admin/update`

**请求体**：

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "code": "SUB-01",
  "name": "起落（更新后）",
  "sort_order": 2
}
```

- `id`：必填，待更新科目的 UUID。
- 其他字段可选，未提供的字段保持原值。

**成功响应**：返回更新后的完整科目对象，结构同创建接口。

#### 7.7.8 删除大纲科目（后台）`POST /api/v1/syllabus-subjects/admin/delete`

**请求体**：

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "status": "deleted"
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

**说明**：

- 当只传 `document_id` 时，系统会使用该文档下的**全部题目**生成一次练习/测试。
- 当同时传入 `paper_id` 时，系统会根据对应试卷配置（`theory_papers` + `theory_paper_questions`）抽取题目子集并设置每题分值，实现“不同试卷”的出题方案。

**请求体（不指定试卷，整套题库）**：

```json
{
  "document_id": "550e8400-e29b-41d4-a716-446655440201"
}
```

**请求体（指定某张试卷）**：

```json
{
  "document_id": "550e8400-e29b-41d4-a716-446655440201",
  "paper_id": "paper-uuid-1"
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
        "tags": ["基础"],
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

#### 8.6 训练科目：开始答题 `POST /api/v1/subject-learning/questions/start`

- **说明**：与理论 `POST /api/v1/theory/questions/start` 同构，针对 PostgreSQL `subject_learning` / Mongo `subject_question_sets`。当前版本**不按试卷抽题**，始终返回该科目下全部题目。
- **前置**：`subject_learning.id` 需已存在；题库写入 Mongo `subject_question_sets`（`subject_learning_id` 为键）。

**请求体**：

```json
{
  "subject_learning_id": "550e8400-e29b-41d4-a716-446655440301"
}
```

**成功响应**：与理论 8.3 相同结构（`attempt_id`、`time_limit_seconds`、`questions[]`）。

#### 8.7 训练科目：交卷 `POST /api/v1/subject-learning/questions/submit`

**请求体**：

```json
{
  "subject_learning_id": "550e8400-e29b-41d4-a716-446655440301",
  "attempt_id": "550e8400-e29b-41d4-a716-446655441101",
  "answers": [
    {
      "question_id": "q1",
      "selected_option_ids": ["B"]
    }
  ]
}
```

**成功响应**：与理论 8.4 相同（`score`、`max_score`、`correct_count`、`total_count`、`passed`）。

#### 8.8 训练科目：历史成绩 `POST /api/v1/subject-learning/attempts/my-list`

**请求体**：

```json
{
  "subject_learning_id": "550e8400-e29b-41d4-a716-446655440301",
  "page": 1,
  "page_size": 20
}
```

**成功响应**：`data.items[]` 为 `SubjectAttempt` 摘要（`subject_learning_id`、`score`、`submitted_at`、`max_score`、`correct_count`、`total_count`、`passed` 等）。

#### 8.9 训练科目：题库管理（后台）

与理论 10.1–10.4 同构，路径与字段替换如下：

- `POST /api/v1/subject-learning/questions/admin/list`：请求体使用 `subject_learning_id`（可选 `page` / `page_size`）。
- `POST /api/v1/subject-learning/questions/admin/create|update`：使用 `subject_learning_id` + `question`（结构与理论 `question` 相同）。
- `POST /api/v1/subject-learning/questions/admin/delete`：`subject_learning_id` + `question_id`。

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
    "id": "", // 可留空，由后端生成
    "type": "single", // single | multi
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

- **说明**：查看某理论文档下已配置的试卷列表，用于后台维护和前端出卷选择（例如 A/B 卷、模拟考试卷、强化训练卷等）。

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
  "id": null, // 为空表示创建；非空表示更新
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

> 前端通常不直接调用该接口，而是在后台管理系统中使用，用于配置不同的出卷方案；学员端只需要通过 `POST /api/v1/theory/papers/list` 获取可选试卷，并在 `POST /api/v1/theory/questions/start` 时携带所选 `paper_id`。

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

- 若携带 `paper_id`：
  - 从 `theory_paper_questions` 中按 `sort_order` 取出该卷的题目 ID 和分值。
  - 在 Mongo 题库 `theory_question_sets` 中按 `question_id` 拉取题干与选项，并将 `score` 覆盖为试卷中配置的分值。
- 若不携带 `paper_id`：
  - 保持原有行为，直接使用该文档下的全部题目作为一套卷。

---

### 12. 教员任务发布与管理 API

本节描述教员在“训练档案 + 科目 + 下次任务”一体化界面下需要使用的接口。所有接口均需登录，且通常要求当前用户角色为 `instructor`。

#### 12.1 本课科目状态视图（扩展）`POST /api/v1/syllabus-subjects/list`

- **说明**：在原有返回 `is_graduated` 的基础上，当请求体中携带 `student_id` 时，后端会为每个科目计算并返回一个派生字段 `status`，用于前端科目管理视角展示。

**请求体**：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**成功响应**（在原有基础上增加 `status` 字段，仅当传入 `student_id` 时计算）：

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
        "is_graduated": true,
        "status": "completed"
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655440002",
        "code": "SUB-02",
        "name": "带飞",
        "sort_order": 2,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z",
        "is_graduated": false,
        "status": "in_progress"
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655440003",
        "code": "SUB-03",
        "name": "仪表飞行",
        "sort_order": 3,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z",
        "is_graduated": false,
        "status": "not_started"
      }
    ]
  }
}
```

- **状态计算规则**（基于当前学员的所有 `training_archives` 与 `training_archive_subjects` 记录）：
  - `completed`：该科目在任意一条档案中出现，且至少有一条记录 `is_graduated = true`。
  - `in_progress`：该科目在档案中出现过，但所有记录均为 `is_graduated = false`。
  - `not_started`：该科目从未在任何档案中出现。

#### 12.2 布置下次训练科目（已有）`POST /api/v1/training-archives/next-subjects/update`

- **说明**：教员在当前训练档案页面中勾选若干大纲科目作为“下次训练任务”，前端将所选科目的 `syllabus_subject_ids` 作为数组提交。后端会用新集合替换原有数据（幂等），并记录当前登录用户为 `assigned_by`。

**请求体**：

```json
{
  "training_archive_id": "550e8400-e29b-41d4-a716-446655440101",
  "syllabus_subject_ids": [
    "550e8400-e29b-41d4-a716-446655440302",
    "550e8400-e29b-41d4-a716-446655440303"
  ]
}
```

**成功响应**同 7.7.5。

#### 12.3 教员签字前的业务约束（扩展）`POST /api/v1/training-archives/sign`

- **说明**：在已有签字逻辑基础上，增加**责任教员 / 主任教员签字前必须先布置下次科目**的校验。

**请求体**（保持不变）：

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440101",
  "role": "responsible_instructor"
}
```

**新增行为说明**：

- 当 `role` 为 `responsible_instructor` 或 `chief_instructor` 时，后端会在签字前检查当前档案在 `next_training_subjects` 表中是否存在至少一条记录：
  - 若不存在下次训练科目，则返回 `400` 错误，示例：

```json
{
  "code": 400,
  "msg": "请先布置下次训练科目",
  "data": null
}
```

- 学员签字（`role = "student"`）不受该约束影响。

> 通过该约束实现“每次训练结束前必须先安排下次训练任务”的业务规则。

---

### 13. 学员信息管理（后台）API（设计稿）

> 说明：本节为基于当前数据库结构的后台“学员信息管理”接口设计，实际落地时可按本规范实现到对应 Handler 中。

#### 13.1 学员列表（后台）`POST /api/v1/admin/students/list`

- **认证**：需要，推荐仅 `admin` / `staff` 角色可访问。
- **说明**：用于后台表格展示所有学员的基础信息与训练/合同摘要。

**请求体**：

```json
{
  "keyword": "",
  "responsible_instructor_id": "",
  "page": 1,
  "page_size": 20
}
```

**成功响应（示例）**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "student_id": "550e8400-e29b-41d4-a716-446655440001",
        "name": "张三",
        "student_no": "S000001",
        "username": "student1",
        "phone": "13800000000",
        "email": "s1@example.com",
        "contract_duration_minutes": 3600,
        "flown_minutes": 1200,
        "remaining_minutes": 2400,
        "training_progress": "已完成单飞阶段",
        "responsible_instructor": {
          "id": "550e8400-e29b-41d4-a716-446655440002",
          "name": "教员A"
        }
      }
    ],
    "total": 1
  }
}
```

#### 13.2 学员详情（后台）`POST /api/v1/admin/students/detail`

**请求体**：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**成功响应（示例）**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "basic": {
      "student_id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "张三",
      "student_no": "S000001",
      "contract_duration_minutes": 3600,
      "flown_minutes": 1200,
      "remaining_minutes": 2400,
      "training_progress": "已完成单飞阶段"
    },
    "account": {
      "username": "student1",
      "phone": "13800000000",
      "email": "s1@example.com",
      "role": "student"
    },
    "responsible_instructor": {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "name": "教员A"
    },
    "latest_archive": {
      "id": "550e8400-e29b-41d4-a716-446655440101",
      "training_date": "2026-03-15",
      "stage": "初级训练",
      "flight_duration_minutes": 92
    },
    "pending_task_count": 3
  }
}
```

#### 13.3 新建 / 更新学员（后台）`POST /api/v1/admin/students/save`

- **说明**：用于后台创建或编辑学员信息，同时维护 `users` 与 `students` 两张表的字段。

**请求体（示例）**：

```json
{
  "student_id": null,
  "name": "张三",
  "student_no": "S000001",
  "username": "student1",
  "phone": "13800000000",
  "email": "s1@example.com",
  "password": "secret",
  "contract_duration_minutes": 3600,
  "flown_minutes": 0,
  "training_progress": "",
  "responsible_instructor_id": "550e8400-e29b-41d4-a716-446655440002"
}
```

- 当 `student_id = null` 时，后端将：
  - 在 `users` 表中创建一条 `role = "student"` 的账号记录；
  - 在 `students` 表中创建一条记录并关联 `user_id`；
  - 返回新建的学员主键与账号信息。
- 当 `student_id` 非空时，仅更新已存在学员的基础信息与账号（密码可选字段 `new_password` 单独处理）。

---

### 14. 学员“我的主页”聚合 API（设计稿）

#### 14.1 学员主页 `POST /api/v1/students/my-home`

- **认证**：需要，当前登录用户 `role = "student"`。
- **说明**：为学员端“我的主页”提供一个聚合接口，汇总学员档案、任务、训练计划与理论学习进度等多源信息。

**请求体**：

```json
{}
```

**成功响应（结构示例）**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "profile": {
      "student_id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "张三",
      "student_no": "S000001",
      "contract_duration_minutes": 3600,
      "flown_minutes": 1200,
      "remaining_minutes": 2400,
      "training_progress": "已完成单飞阶段"
    },
    "tasks": {
      "items": [
        {
          "id": "550e8400-e29b-41d4-a716-446655442001",
          "type": "theory",
          "title": "失速与尾旋理论学习",
          "due_at": "2026-03-20T00:00:00Z",
          "status": "pending"
        }
      ],
      "total": 3
    },
    "upcoming_training": [
      {
        "training_date": "2026-03-20",
        "start_time": "10:00",
        "end_time": "11:30",
        "instructor_name": "教员A",
        "aircraft_name": "塞斯纳 172"
      }
    ],
    "last_training_archive": {
      "id": "550e8400-e29b-41d4-a716-446655440101",
      "training_date": "2026-03-15",
      "stage": "初级训练",
      "flight_duration_minutes": 92
    },
    "theory_summary": {
      "completed_documents": 5,
      "avg_score": 86,
      "last_attempt_at": "2026-03-16T10:00:00Z"
    },
    "subject_progress": {
      "completed": 10,
      "total": 20
    }
  }
}
```

- **实现建议**：后端在实现时可复用以下现有能力：
  - `students` / `users`：查询学员基本信息与账号；
  - `POST /api/v1/tasks/my-list`：获取当前学员任务列表，取前若干条作为预览；
  - `training_plans`：按未来一段时间（如 7 天）查询训练计划；
  - `POST /api/v1/students/training-archives/list`：获取最近一条训练档案；
  - `theory_attempts` 与理论模块接口：统计理论学习进度；
  - `POST /api/v1/syllabus-subjects/list`（带 `student_id`）：统计科目完成数量。

---

### 15. 合同文件管理 API

> 说明：本节接口基于 PostgreSQL `contracts` 表与 MongoDB GridFS + `contract_files_meta` 集合，实现合同扫描件/附件的上传、列表与下载。所有接口均需登录，并按照业务权限校验学员与合同关系。

#### 15.1 上传合同文件 `POST /api/v1/contracts/files/upload`

- **认证**：需要，建议仅 `admin` / `staff` / 具有相应学员访问权限的教员角色调用。
- **请求方式**：`multipart/form-data`

**表单字段**：

- `contract_id`：合同 ID（`contracts.id`，UUID 字符串）。
- `student_id`：合同所属学员 ID（`students.id`）。
- `file`：实际上传的文件（PDF、图片等）。

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": "file-meta-id-1",
    "filename": "contract-signed.pdf",
    "content_type": "application/pdf",
    "size": 123456,
    "tags": [],
    "uploaded_at": "2026-03-20T10:00:00Z"
  }
}
```

**行为说明**：

- 将文件二进制内容写入 MongoDB GridFS（Bucket 名称如 `contracts_files`）。
- 在 `contract_files_meta` 集合中写入一条元数据文档，记录 `student_id`、`contract_id`、`filename`、`content_type`、`size`、`gridfs_id` 等。
- 将该元数据文档的 `_id`（字符串形式）追加到 PG `contracts.mongo_file_ids` 数组字段中，形成 PG ↔ Mongo 的关联。

#### 15.2 合同文件列表 `GET /api/v1/contracts/files/list`

- **认证**：需要。
- **请求参数**（Query）：

```http
GET /api/v1/contracts/files/list?contract_id=<UUID>
```

- `contract_id`：目标合同 ID（必填）。

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "file-meta-id-1",
        "filename": "contract-signed.pdf",
        "content_type": "application/pdf",
        "size": 123456,
        "tags": ["signed"],
        "uploaded_at": "2026-03-20T10:00:00Z"
      }
    ]
  }
}
```

> 说明：该接口仅返回元数据列表，不包含文件内容。前端通常在学员详情页或合同详情页中使用，展示已有合同扫描件清单。

#### 15.3 下载合同文件 `GET /api/v1/contracts/files/download`

- **认证**：需要。
- **请求参数**（Query）：

```http
GET /api/v1/contracts/files/download?id=<FILE_META_ID>
```

- `id`：合同文件元数据 ID（即 `contract_files_meta._id`，由上传接口返回）。

**成功行为**：

- 根据 `id` 从 `contract_files_meta` 获取对应记录，读出内部 `gridfs_id`。
- 通过 GridFS 以流方式读取文件内容。
- 按元数据中的 `content_type` 设置 `Content-Type` 响应头，返回文件二进制内容。

**响应示例**：

- HTTP 头部：

```http
HTTP/1.1 200 OK
Content-Type: application/pdf
Content-Disposition: attachment
```

- Body：文件二进制流（不再包裹统一的 JSON 结构）。

---

### 16. 学员合同管理 API（设计稿）

> 说明：本节为基于当前 `contracts` 表结构的后台“学员合同管理”接口设计，包括合同的创建、更新、详情与按学员查询等，接口均建议由 `admin` / `staff` 角色或具有相应权限的教员调用。

#### 16.1 创建合同 `POST /api/v1/admin/contracts/create`

- **认证**：需要，推荐仅 `admin` / `staff` 角色可访问。
- **说明**：为指定学员创建一份新的训练合同。

**请求体（示例）**：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "contract_no": "HT-2026-001",
  "sign_date": "2026-03-20",
  "start_date": "2026-04-01",
  "end_date": "2027-04-01",
  "total_minutes": 3600,
  "price_total": 150000.00,
  "currency": "CNY",
  "status": "active",
  "remark": "私商一体训练合同"
}
```

**成功响应（示例）**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": "c-uuid-1",
    "student_id": "550e8400-e29b-41d4-a716-446655440001",
    "contract_no": "HT-2026-001",
    "sign_date": "2026-03-20",
    "start_date": "2026-04-01",
    "end_date": "2027-04-01",
    "total_minutes": 3600,
    "price_total": 150000.00,
    "currency": "CNY",
    "status": "active",
    "mongo_file_ids": [],
    "remark": "私商一体训练合同",
    "created_at": "2026-03-20T10:00:00Z",
    "updated_at": "2026-03-20T10:00:00Z"
  }
}
```

#### 16.2 更新合同 `POST /api/v1/admin/contracts/update`

- **认证**：需要。
- **说明**：更新合同的关键信息，例如结束日期、总时长、金额、状态等。

**请求体（示例）**：

```json
{
  "id": "c-uuid-1",
  "end_date": "2027-06-01",
  "total_minutes": 4000,
  "price_total": 168000.00,
  "status": "active",
  "remark": "延长 2 个月，追加 400 分钟"
}
```

- `id`：必填，其余字段可选，未提供的字段保持原值。

**成功响应（示例）**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": "c-uuid-1",
    "student_id": "550e8400-e29b-41d4-a716-446655440001",
    "contract_no": "HT-2026-001",
    "sign_date": "2026-03-20",
    "start_date": "2026-04-01",
    "end_date": "2027-06-01",
    "total_minutes": 4000,
    "price_total": 168000.00,
    "currency": "CNY",
    "status": "active",
    "mongo_file_ids": ["file-meta-id-1"],
    "remark": "延长 2 个月，追加 400 分钟",
    "created_at": "2026-03-20T10:00:00Z",
    "updated_at": "2026-03-21T08:00:00Z"
  }
}
```

#### 16.3 合同详情 `POST /api/v1/admin/contracts/detail`

- **认证**：需要。
- **说明**：查看单份合同的详细信息及其绑定的文件列表（可结合 15.2/15.3 使用）。

**请求体**：

```json
{
  "id": "c-uuid-1"
}
```

**成功响应（示例）**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "contract": {
      "id": "c-uuid-1",
      "student_id": "550e8400-e29b-41d4-a716-446655440001",
      "contract_no": "HT-2026-001",
      "sign_date": "2026-03-20",
      "start_date": "2026-04-01",
      "end_date": "2027-06-01",
      "total_minutes": 4000,
      "price_total": 168000.00,
      "currency": "CNY",
      "status": "active",
      "mongo_file_ids": ["file-meta-id-1"],
      "remark": "延长 2 个月，追加 400 分钟",
      "created_at": "2026-03-20T10:00:00Z",
      "updated_at": "2026-03-21T08:00:00Z"
    },
    "files": [
      {
        "id": "file-meta-id-1",
        "filename": "contract-signed.pdf",
        "content_type": "application/pdf",
        "size": 123456,
        "tags": ["signed"],
        "uploaded_at": "2026-03-20T10:00:00Z"
      }
    ]
  }
}
```

> 说明：`files` 数组可直接复用 15.2 中的文件元数据结构，实际实现时可通过 `contracts.mongo_file_ids` 与 `contract_files_meta` 关联查询。

#### 16.4 按学员查询合同列表 `POST /api/v1/admin/contracts/list-by-student`

- **认证**：需要。
- **说明**：在学员详情页中列出该学员的全部合同（含历史合同）。

**请求体**：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**成功响应（示例）**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "c-uuid-2",
        "contract_no": "HT-2025-001",
        "sign_date": "2025-03-01",
        "start_date": "2025-04-01",
        "end_date": "2026-03-31",
        "total_minutes": 3000,
        "price_total": 120000.00,
        "currency": "CNY",
        "status": "finished"
      },
      {
        "id": "c-uuid-1",
        "contract_no": "HT-2026-001",
        "sign_date": "2026-03-20",
        "start_date": "2026-04-01",
        "end_date": "2027-06-01",
        "total_minutes": 4000,
        "price_total": 168000.00,
        "currency": "CNY",
        "status": "active"
      }
    ]
  }
}
```

> 建议：按 `sign_date` 或 `created_at` 倒序排序，方便后台快速看到最新合同。

#### 16.5 合同状态变更（可选）`POST /api/v1/admin/contracts/change-status`

- **认证**：需要。
- **说明**：用于将合同标记为已结束、已作废等，配合 `status` 字段使用。

**请求体（示例）**：

```json
{
  "id": "c-uuid-1",
  "status": "finished",
  "remark": "学员提前结业，合同结束"
}
```

**成功响应（示例）**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "status": "ok"
  }
}
```

