# 接口总览文档（magua-fly-golang）

本文汇总当前后端已实现的主要 HTTP 接口，按业务模块分组，统一说明：

- **认证方式**：除 `/health`、`/auth/*` 外，其余接口均需 `Authorization: Bearer <JWT>`。
- **请求格式**：主体接口为 `POST`、`Content-Type: application/json`；`GET /health` 及少量资源接口（如理论课视频流、飞机头像、合同文件列表/下载）见对应小节。
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

### 任职状态、人员类别与鉴权（全局）

- **`users.employment_status`（任职状态）**：`active`（在职）| `resigned`（离职）| `part_time`（兼职）。新建账号默认 `active`（数据库迁移 `027_user_employment_status`）。与 **`deleted_at` 软删除**不同：软删用于归档与用户名/学号复用；任职状态用于人事状态与是否允许登录、是否出现在非管理员选人列表中。
- **规则摘要**：
  - **在职 `active`**：可登录；持有效 JWT 可调用业务接口（仍受角色与具体接口权限约束）。
  - **离职 `resigned` / 兼职 `part_time`**：不可用户名密码登录、不可短信登录（返回与密码错误一致：`401`，`msg: invalid credentials`）；若仍持有旧 JWT，后续带鉴权请求在中间件中会返回 **`403`**，Body 示例：`{"error":"employment inactive"}`（与部分业务接口的 `code`/`msg` 包装可能不同，以实际响应为准）。
  - **`role = admin`**：登录与 JWT 校验**不**按任职状态拦截；建议后台仍保持管理员账号为 `active`。
- **人员类别（不落库，由 `role` 推导）**：`student` / `instructor` → `flight`（飞行）；`maintenance` → `maintenance`（维修）；`staff` → `ground`（地面）；`admin` → `admin`。后台学员列表等接口会在有账号时返回 `personnel_category` 供展示。
- **学员 / 教员列表过滤**：`POST /api/v1/students/list`、`POST /api/v1/instructors/list` 在调用者**非** `admin` 时，仅返回关联用户为 **`active`** 的数据（学员无 `user_id` 的档案仍返回）；**`admin`** 调用时返回全部未软删记录。
- **管理员专属路由**：凡 `POST /api/v1/admin/...` 以及路径中含 **`/admin/`** 的管理接口（如大纲、理论/科目题库管理等），均需 **`role = admin`**，否则 **`403`**（如 `{"error":"forbidden"}`）。

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
      "employment_status": "active",
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

- 若账号存在但 **`employment_status` 为 `resigned` 或 `part_time`**（且非 `admin` 角色），同样返回上述 `401` / `invalid credentials`，与密码错误不做区分。

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
  "role": "student" // admin | staff | maintenance | instructor | student
}
```

**行为说明**：

- 创建一条 `users` 记录。
- 新用户 **`employment_status` 默认为 `active`**（无需在注册请求中传递）。
- 若 `role = "student"`，自动在 `students` 表创建一条记录并关联 `user_id`。
- 若 `role = "instructor"`，自动在 `instructors` 表创建一条记录并关联 `user_id`。
- 若 `role = "maintenance"`（机务），仅创建 `users`，**不**创建 `instructors`（不上机、无教员档案）；日排班与 `staff` 相同，使用 `staff_schedules`（见 §5.6）。

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
    "employment_status": "active",
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

- 创建 `users` 记录（`role = "student"`），默认 **`employment_status = active`**。
- 自动创建并关联 `students` 记录。

短信登录场景下，若手机号对应用户为 **`resigned` / `part_time`**（且非 `admin`），返回 **`401`** / `invalid credentials`，与 §2.1 一致。

---

## 4. 学员、教员、飞机基础信息

### 4.1 学员列表 `POST /api/v1/students/list`

- **认证**：需要。
- **任职状态过滤**：调用者为 **`admin`** 时列出全部未软删学员；**非 admin** 时仅列出「无关联账号」或关联用户 **`employment_status = active`** 的学员（见文首「任职状态、人员类别与鉴权」）。

**请求体**（JSON；`Content-Length` 为 0 或仅空白时视为未传，使用默认分页与「无附加筛选」）：

```json
{
  "limit": 20,
  "offset": 0,
  "keyword": "",
  "responsible_instructor_id": "",
  "flight_training_status": ""
}
```

- `limit`：默认 `20`，最大 `100`；`<=0` 时使用默认。
- `offset`：默认 `0`；`<0` 时按 `0` 处理。
- `keyword`：**可选**。非空时按学员**姓名或学号**模糊匹配（`ILIKE`）。
- `responsible_instructor_id`：**可选**。非空时按责任教员（`instructors.id`）精确过滤。
- `flight_training_status`：**可选**。非空时按学飞状态精确过滤，须为 `in_progress`（进行中）| `paused`（已暂停）| `completed`（已完结）| `terminated`（已终止）；非法取值返回 **`400`**，`msg` 形如 `invalid flight_training_status`。

**成功响应**：每条在 `students` 档案字段外，若存在 **`user_id`**，会附带关联 **`users` 账号**上的 `username`、`phone`、`email`、`employment_status`、`personnel_category`（推导规则见文首）；无账号时上述账号字段通常为空或省略。档案侧常见字段还含 **`resident_city`**（常驻城市）、**`flight_training_status`**、**`responsible_instructor`**（`{ "id", "name" }`，来自 `instructors`）、`contract_info`、`training_progress_json` 等。

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
        "username": "student1",
        "phone": "13800000000",
        "email": "s1@example.com",
        "employment_status": "active",
        "personnel_category": "flight",
        "resident_city": "上海",
        "flight_training_status": "in_progress",
        "responsible_instructor_id": "550e8400-e29b-41d4-a716-446655440002",
        "responsible_instructor": {
          "id": "550e8400-e29b-41d4-a716-446655440002",
          "name": "教员A"
        },
        "flown_minutes": 0,
        "training_progress": "",
        "created_at": "...",
        "updated_at": "..."
      }
    ]
  }
}
```

### 4.1.1 职员端学员概要 `POST /api/v1/students/staff-overview`

- **认证**：需要 JWT。
- **权限**：仅 `admin`、`staff`、`maintenance`、`instructor` 可访问；`student` 返回 `403`，`msg` 形如 `forbidden: student cannot access this resource`。
- **说明**：供教员/地面职员查看指定学员的训练状态、大纲科目进度、飞行与 `contract_info` 中的权益（学飞积分、亲友体验次数等），**不包含**学员账号的手机、邮箱、用户名。

**请求体**：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001"
}
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
      "student_no": "S000001"
    },
    "training": {
      "training_progress": "已完成单飞阶段",
      "training_progress_json": { "phase": "PPL-1" },
      "pending_task_count": 2,
      "latest_archive": {
        "id": "550e8400-e29b-41d4-a716-446655440101",
        "training_date": "2026-03-15",
        "stage": "初级训练",
        "flight_duration_minutes": 92
      },
      "upcoming_training": [
        {
          "training_date": "2026-03-20",
          "start_time": "10:00",
          "end_time": "11:30",
          "instructor_name": "教员A",
          "aircraft_name": "塞斯纳 172"
        }
      ]
    },
    "syllabus_progress": {
      "completed": 10,
      "total": 20
    },
    "entitlements": {
      "flight": {
        "contract_duration_minutes": 3600,
        "flown_minutes": 1200,
        "remaining_minutes": 2400
      },
      "learning_points": 120,
      "guest_experience": {
        "quota": 5,
        "used": 2,
        "remaining": 3
      }
    }
  }
}
```

- **`contract_info` 权益约定**（写入 `students.contract_info` JSONB，可由 `POST /api/v1/admin/students/save` 的 `contract_info` 字段维护）：

```json
{
  "learning_points": 120,
  "guest_experience": {
    "quota": 5,
    "used": 2
  }
}
```

缺字段或非法 JSON 时，解析结果按 `0` 处理；`remaining = max(0, quota - used)`。

### 4.2 教员列表 `POST /api/v1/instructors/list`

- **认证**：需要。
- **任职状态过滤**：**`admin`** 返回全部未软删教员；**非 admin** 仅返回关联用户为 **`active`** 的教员（见文首全局约定）。

**成功响应**：每条在教员档案字段外，附带关联 **`users` 账号**上的 `username`、`phone`、`email`、`employment_status`、`personnel_category`（与学员列表侧推导规则一致）；若关联用户查询失败则这些字段省略或为空。

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
        "username": "ins01",
        "phone": "13800000000",
        "email": "ins@example.com",
        "employment_status": "active",
        "personnel_category": "flight",
        "created_at": "...",
        "updated_at": "..."
      }
    ]
  }
}
```

### 4.3 飞机列表 `POST /api/v1/aircraft/list`

**说明**：

- **`status`（业务状态，库表校验）**：仅允许 `active`（现役）、`retired`（退役）、`partnership`（合作）。
- **`avatar_url`**：已上传头像时返回相对路径，客户端需拼接网关/API 基址；使用方式与 §4.4.5 一致（`GET` + `Authorization`）。未设置头像时通常省略该字段。

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
        "avatar_url": "/api/v1/aircraft/avatar?aircraft_id=550e8400-e29b-41d4-a716-446655440000",
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

- `code`：必填。
- `name`：可选。
- `status`：可选，默认 `active`；仅允许 `active` | `retired` | `partnership`（含义见 §4.3）。

**成功响应**同上单条飞机结构（含 `avatar_url` 时与列表一致；新建飞机一般无头像）。

### 4.4.1 更新飞机 `POST /api/v1/aircraft/update`

用于修改名称与业务状态（至少提供 `name` 与 `status` 之一）。

**请求体**：

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "塞斯纳 172R",
  "status": "partnership"
}
```

- `id`：必填，飞机 UUID。
- `name`、`status`：可选；`status` 取值规则同 §4.3。

**成功响应**：`data` 为更新后的单条飞机对象（结构同 §4.3 列表项）。

### 4.4.2 删除飞机 `POST /api/v1/aircraft/delete`

删除 `aircraft` 记录；关联的可用时段、维护时间窗等随外键级联删除。若该飞机在 GridFS 中存有头像文件，会先删除对象再删库表行。

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
  "data": { "deleted": true }
}
```

### 4.4.3 上传飞机头像 `POST /api/v1/aircraft/avatar/upload`

- **认证**：需要。
- **请求**：`multipart/form-data`（非 JSON）。
- **表单字段**：
  - `aircraft_id`（必填）：飞机 UUID。
  - `file`（必填）：图片文件。
- **限制**：单文件最大 **5MB**；`Content-Type` 须为 `image/jpeg`、`image/png`、`image/webp` 或 `image/gif`（未带类型时会按文件头探测，仍须落在上述类型内）。

头像二进制写入 MongoDB **GridFS**，bucket 名为 `aircraft_avatars`；PostgreSQL `aircraft.avatar_gridfs_id` 仅存文件 ID。若该飞机已有头像，上传成功后会替换为新文件并尽量删除旧 GridFS 对象。

**成功响应**：`data` 为当前飞机完整对象（同 §4.3 列表项，含 `avatar_url`）。

### 4.4.4 移除飞机头像 `POST /api/v1/aircraft/avatar/remove`

清除 `avatar_gridfs_id` 并删除 GridFS 中的对应文件。

**请求体**：

```json
{
  "aircraft_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**成功响应**：`data` 为当前飞机对象（无头像时无 `avatar_url` 或为空串，以后端实现为准）。

### 4.4.5 拉取飞机头像 `GET /api/v1/aircraft/avatar`

- **认证**：需要（请求头 `Authorization: Bearer <JWT>`），与理论课托管视频流类似，便于在 WebView/App 内带 token 拉流或使用下载接口转 blob。
- **Query**：`aircraft_id=<飞机 UUID>`（必填）。
- **响应**：成功时为图片二进制流，HTTP `Content-Type` 以上传时记录的类型为准（缺省为 `application/octet-stream`）；无飞机、未设置头像或 GridFS 中文件缺失时返回 `404` 及 JSON 错误体（与文首统一结构一致）。

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
  "task_type": "training",
  "aircraft_id": "550e8400-e29b-41d4-a716-446655440003",
  "plan_date": "2026-03-15",
  "start_time": "10:00",
  "end_time": "11:30"
}
```

要求：

- `student_id` 必须是 `students.id`（不是 `users.id`）。
- `instructor_id`：**可选**。为空或省略时，使用学员档案 `students.responsible_instructor_id` 作为默认教员（责任教员）；若档案中未设置责任教员，则须显式传 `instructor_id`，否则 `400`。非空时须为 `instructors.id`（可预约非责任教员）。
- `task_type`：**可选**，缺省为 `training`。合法取值：`training`（训练）、`experience`（体验）、`proficiency`（熟练）、`verification`（验证）、`test_flight`（试飞）、`ferry`（调机）、`internal_training`（内训）、`local_pattern`（本场带飞）。非法取值 `400`。响应中的 `booking` 含 `task_type`。
- **权限**：`role=student` 时仅允许为**本人**学员档案预约（`student_id` 须与当前登录用户关联的 `students.id` 一致），否则 `403`；**仅** `role=admin` 可为任意学员**代预约**；`staff` / `maintenance` / `instructor` 等账号**不能**调用本接口创建预约（`403`）。
- **飞行记录归属**：预约与计划中的 `instructor_id` 为所选/默认教员；实际带飞教员以训练记录 `training_records.instructor_id` 为准（录入或更新记录时可修正，与预约教员可能不一致）。
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

网格中 `available` 时段由教员**日排班**解析得到：业务日默认为 **09:00–19:00**、时间轴 **15 分钟** 一格（`slotMinutes: 15`）、上下午以 **14:00** 分界（`morning` / `afternoon` 推导规则），与 `duty_status` / `duty_windows` 的解析规则一致（见 §5.5）。教员当日**无排班记录**时，该教员时间轴全部为 `off`。`19:00` 之后无格点；历史预约若结束时间晚于 `19:00`，网格仅覆盖至当日最后一格（半开区间与 `dayEnd` 对齐）。

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
    "slotMinutes": 15,
    "dayStart": "09:00",
    "dayEnd": "19:00",
    "slots": ["09:00", "09:15", "..."],
    "instructors": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "教员A",
        "displayColor": "#ff0000",
        "phone": "13800000000",
        "email": "ins@example.com",
        "timeline": [
          {
            "slotIndex": 0,
            "startTime": "09:00",
            "endTime": "09:15",
            "status": "available"
          },
          {
            "slotIndex": 1,
            "startTime": "09:15",
            "endTime": "09:30",
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

**响应 `data.items[]`**：每条对应库中已有的一行排班，包含 `instructor_name`、`duty_status`、`remark`、`duty_windows`（库内原始值）、`resolved_windows`（解析后的展示窗口）、`flight_tasks`（由当日的 `bookings` 与 `training_plans` 聚合，相同学员/起止时刻去重，优先保留预约；其中 `source` 为 `booking` 的条目带 `booking_status`：`pending` | `approved` | `rejected`）、`ground_courses`（当前为空数组，结构与 `flight_tasks` 一致，预留地面课）。

#### 5.5.2 教员我的排班与任务 `POST /api/v1/schedules/instructor-duties/my-list`

供**已登录且 `instructors.user_id` 已绑定当前用户**的教员使用，无需传 `instructor_id`。请求体与 §5.5.1 相同（仅 `from_date`、`to_date`）。

**响应 `data`**：

- `instructor_id`、`instructor_name`：当前教员档案。
- `items`：与 §5.5.1 的 `items[]` 结构相同，且仅含该教员在区间内的排班行。
- `pending_bookings`：`status = pending` 且 `instructor_id` 为当前教员的预约列表（**不按** `from_date`/`to_date` 过滤，便于总览待审核）。

若当前用户未绑定教员档案，返回 `404`，`msg` 为「未绑定教员档案」。

#### 5.5.3 单日写入/更新 `POST /api/v1/schedules/instructor-duties/upsert`

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

#### 5.5.4 按区间批量写入 `POST /api/v1/schedules/instructor-duties/bulk`

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

面向地勤、行政 **staff** 与 **机务 maintenance** 的考勤排班；**教员**行仍读取 `instructor_schedules`（与 §5.4 规则一致），**不可预约**、无 `bookings` 覆盖层。

#### 5.6.1 全员考勤网格 `POST /api/v1/schedules/workforce-grid`

返回 **role 为 `staff`、`maintenance` 与 `instructor` 的用户**（不含 `admin`、`student`），格点定义与教员网格相同（`dayStart`/`dayEnd`/`slots`/`slotMinutes`）。用户来源为 `users` 按角色查询，且仅包含 **`employment_status = active`** 的账号（离职/兼职不出现在人力网格选人列表中）。

**请求体**：

```json
{ "date": "2026-03-15" }
```

**响应 `data`**：

- `workers[]`：每人一行。
  - `userId`、`role`（`staff` | `maintenance` | `instructor`）、`username`、`name`（教员取 `instructors.name`，staff/maintenance 取 `username`）、`phone`、`email`（来自 `users`）、`instructorId`（仅教员有）、`displayColor`（教员可选）、`timeline[]`。
  - `timeline[].status` 仅为 `off` 或 `available`（考勤在岗），**无** `booked` / `aircraft_unavailable`。
- 教员行的执勤窗口来自 **`instructor_schedules`**；staff / 机务行来自 **`staff_schedules`**（按 `users.id`）。

#### 5.6.2 staff 排班列表 `POST /api/v1/schedules/staff-duties/list`

仅列出 **`staff_schedules`** 中记录（`user_id` 对应 `role=staff` 或 `role=maintenance` 的用户）。

**请求体**：`from_date`、`to_date`、`user_ids`（可选，空表示不限）。

**响应**：`data.items[]` 含 `user_id`、`username`、`name`、`phone`、`email`（来自 `users`）、`duty_status`、`resolved_windows` 等（无带飞任务字段）。

#### 5.6.3 staff 单日 upsert `POST /api/v1/schedules/staff-duties/upsert`

**请求体**：`user_id`（须为 **staff** 或 **maintenance** 角色）、`date`、`duty_status`、`remark`、`duty_windows`（可选，语义同教员）。

#### 5.6.4 staff 批量 `POST /api/v1/schedules/staff-duties/bulk`

**请求体**：`user_id`（staff 或 maintenance）、`from_date`、`to_date`、`weekdays`、`duty_status`、`remark`、`replace_existing`（与教员 bulk 一致）。

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

**成功响应**：`data.fields` 为字段元数据数组（与图一总表及 JSON 字段对齐），包含但不限于：`record_serial_no`（只读，TR 业务编号）、`student_id`、`instructor_id`、`aircraft_id`、`training_date`、`departure`/`destination`（默认莲洲）、`is_local`、`task_type`、`flight_type`、`business_type`、`engine_on_time`/`engine_off_time`、`engine_start_meter`/`engine_stop_meter`、`landings`、`video_timeliness`、`video_operator`、`video_validity`、`flight_data_ref`、`import_source`、`status` 等；选项类字段带 `options` 中文枚举。具体以服务端 `TrainingRecordFieldsConfig` 返回为准。

### 7.2 训练记录 CRUD 与列表

#### 7.2.0 两种录入与「认领」语义说明

- **手动录入**：调用 `POST /api/v1/training-records` 创建记录，`import_source` 为 `manual`，`status` 一般为 `normal`（也可按需传其他状态）。
- **排期驱动的待补全记录（认领/补全模式）**：学员预约经后台**审核通过**后，系统会写入 `training_plans`，并**自动**插入一条 `training_records` 草稿：`import_source`=`schedule`，`status`=`draft`，并带上 `student_id`、`instructor_id`、`aircraft_id`、`training_date`（计划日）、`planned_duration_minutes`（由计划起止时刻推算）、`training_plan_id`。前端在记录管理中列出 `status=draft` 或 `import_source=schedule` 即可作为「待补全」队列；用户通过 `POST /api/v1/training-records/update` 补充开车/关车、时长、档案编号等字段后，将 `status` 改为 `normal` 即完成录入。
- **导入流水认领**（与上不同）：见 **7.3** 节 `import` + `claim`，针对 `raw_flight_imports`，**不是**排期表关联。

计划开始/结束时刻不在 `training_records` 上冗余存储，可通过 `training_plan_id` 关联 `training_plans` 查询 `plan_date`、`start_time`、`end_time`。

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
  "departure": "莲洲",
  "destination": "莲洲",
  "flight_type": "带飞",
  "task_type": "训练",
  "record_serial_no": "",
  "business_type": "学员培训",
  "engine_start_meter": 100.5,
  "engine_stop_meter": 101.6,
  "video_timeliness": "T+0",
  "video_operator": "罗俊杰",
  "video_validity": "正常",
  "flight_data_ref": "",
  "status": "normal"
}
```

**成功响应**：`data` 为完整训练记录（含 `id`、`student_no`（与 `students.student_no` 同步的快照，创建/更新时若未传则由服务端按 `student_id` 填充）、`record_serial_no`（非草稿时自动生成 `TR+YYYYMMDD+序号`）、`training_plan_id`（可空）、`created_at`、`updated_at`）。

#### 7.2.2 记录详情 `POST /api/v1/training-records/detail`

**请求体**：`{ "id": "550e8400-e29b-41d4-a716-446655440011" }`

**成功响应**：`data` 为单条训练记录。

#### 7.2.3 更新记录 `POST /api/v1/training-records/update`

**请求体**：`{ "id": "550e8400-e29b-41d4-a716-446655440011", "student_id": "550e8400-e29b-41d4-a716-446655440001", "instructor_id": "550e8400-e29b-41d4-a716-446655440002", ... }`（需传 `id`，其余为可选更新字段）。补全排期草稿时可传 `engine_on_time`、`engine_off_time`、`flight_duration_minutes`、`archive_display_no`、`landings`、`syllabus_outline`、`theory_duration_minutes`、`sim_duration_minutes`、`departure`、`destination`、`flight_type`、`task_type`、`record_serial_no`、`business_type`、`engine_start_meter`、`engine_stop_meter`、`video_timeliness`、`video_operator`、`video_validity`、`flight_data_ref` 等，并将 `status` 设为 `normal`。**首次**将草稿转为 `normal` 且尚未有业务编号时，服务端会自动分配 `record_serial_no`（若仍为空）。

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
  "import_source": "schedule",
  "sort_field": "training_date",
  "sort_asc": false,
  "limit": 20,
  "offset": 0
}
```

筛选 **待补全排期记录** 时可设 `status` 为 `draft`，或 `import_source` 为 `schedule`（可与日期、学员等条件组合）。

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [ { "id": "550e8400-e29b-41d4-a716-446655440000", "student_id": "550e8400-e29b-41d4-a716-446655440001", "instructor_id": "550e8400-e29b-41d4-a716-446655440002", "training_date": "2026-03-15", "training_plan_id": "550e8400-e29b-41d4-a716-446655440099", "import_source": "schedule", "status": "draft", ... } ],
    "total": 100
  }
}
```

#### 7.2.6 导出总表 CSV `POST /api/v1/training-records/export`

与 **7.2.5** 相同的筛选与排序参数（`student_id`、`instructor_id`、`date_from` / `date_to`、`status`、`import_source`、`sort_field`、`sort_asc`），另可选：

- `include_prior_experience`（布尔，默认 `true`）：为 `true` 时，**此前经历**列为按学员、按训练日期与开车时间升序累计的飞行小时数（与当前库内 `flight_duration_minutes` 一致，四舍五入保留一位小数）；为 `false` 时该列为空。

**请求体示例**：

```json
{
  "date_from": "2026-01-01",
  "date_to": "2026-12-31",
  "status": "normal",
  "sort_field": "training_date",
  "sort_asc": false,
  "include_prior_experience": true
}
```

**成功响应**：HTTP 200，`Content-Type: text/csv; charset=utf-8`，`Content-Disposition: attachment; filename="training_records_YYYYMMDD.csv"`，正文为 **UTF-8 BOM** + CSV。首行为表头，共 **37 列**，顺序与历史业务总表样稿一致。

**已映射列（有数据则输出）**：记录编号（优先 `record_serial_no`（TR 业务号），否则回退 `archive_display_no`）、飞行日期与星期、计划开车/关车时刻（关联 `training_plans` 的 `start_time` / `end_time`，`HH24:MI`）、开车/关车时刻、机号（飞机 `code`）、始发/目的、本场或转场、任务类型、**带飞/单飞**（`flight_type`）、**业务类型**（`business_type`）、飞行教员、客户（学员姓名）、学员阶段与备注（档案）、客户编号（学号：`COALESCE(training_records.student_no, students.student_no)`）、飞行时长（`H:MM`）/分钟数/小时数、起落架次、此前经历（见上）、**发动机小时表读数**（开车/关车）、年月（`Jan-06` 英文缩写）、**视频及时/拍摄人/视频有效**、航后讲评（档案 `instructor_review`）、**飞行数据**（`flight_data_ref` 文本）。

**当前占空列（表头保留，单元格为空，后续可接合同/计费）**：关联合同、小时价值、本次价值、校验、校验结果。

**错误**：筛选结果行数超过 **50000**（常量 `MaxTrainingRecordExportRows`）时返回 **400**，提示缩小范围后重试。

### 7.3 飞行数据导入与认领（外部数据流水，非排期草稿）

本节接口处理**机载/外部文件导入**后的待认领行，与 **7.2.0** 中 `import_source=schedule` 的排期草稿无关。

#### 7.3.0 下载空白总表模板 `POST /api/v1/training-records/import/template`

**请求体**：空 `{}` 或省略。

**成功响应**：HTTP 200，`Content-Type: text/csv; charset=utf-8`，`Content-Disposition: attachment; filename="training_records_template_YYYYMMDD.csv"`，正文为 **UTF-8 BOM** + 一行表头（与 **7.2.6** 相同的 37 列中文表头，与 `domain.TrainingRecordCSVHeaders` 一致），无数据行。

#### 7.3.0b 上传总表 CSV 批量导入 `POST /api/v1/training-records/import/csv`

**请求**：`Content-Type: multipart/form-data`，字段 **`file`**（CSV 文件）；可选字段 **`batch_id`**（批次号，默认当前 Unix 秒）。

表头必须与模板**完全一致**（列名与顺序）。数据行解析规则要点：

- **飞行日期**：`YYYY/M/D`、`YYYY-MM-DD` 等。
- **开车/关车时刻**：与导出一致的 `YYYY/M/D HH:MM`，或当日 `HH:MM`，或四位数 `1301` 表示 `13:01`。
- **本场/转场**：`本场` 或 `转场`。
- **身份解析**（同时满足则自动创建 `training_records` 并关联 `raw_flight_imports`）：
  - **机号**：先按 `aircraft.code`（首尾去空格）匹配；若无记录，再按数据库侧 `lower(trim(code))` 或 `lower(trim(name))` 与 CSV 机号列规范化后比对，且须**唯一**命中；否则报错（如「机号未找到」「飞机匹配到 N 条」）。
  - **飞行教员**：先按教员姓名**原文字符串**精确匹配；若条数不为 1，再按 `lower(trim(name))` 精确匹配，且须**唯一**命中；否则报错（如「教员姓名匹配到 N 条」）。
  - **学员**：**客户编号**列（首尾去空格）**优先按学号** `students.student_no` 匹配；仅当未命中且单元格**形如 UUID**（36 位含连字符的十六进制）时，再按学员 `id` 查询，避免学号与 UUID 同形时误判。若编号为空或两路均未命中，则用 **客户** 列姓名，按 `lower(trim(name))` 在库内**唯一**匹配。自动入库成功后，`training_records.student_no` 与 `training_archives.student_no` 会写入对应学号快照（与 `students.student_no` 一致）。若填写 **客户** 列且已通过编号解析到学员，则姓名须在首尾去空格后与库内姓名 **Unicode 大小写无关**一致（`EqualFold`），否则报错「客户姓名与库内不一致」。仅姓名路径下 0 条为「学员姓名未匹配」，多条为「学员姓名匹配到 N 条」。
- 自动入库失败时，该行仍写入 `raw_flight_imports`（`raw_data` 含 `source: "csv"` 与 `field_map`），可继续用 **7.3.2** 手工认领。

**成功响应**（JSON）：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "batch_id": "1712345678",
    "inserted_raw": 10,
    "auto_claimed": 8,
    "pending_raw": 2,
    "errors": [{ "row": 5, "msg": "教员姓名匹配到 0 条" }],
    "raw_import_ids": ["..."]
  }
}
```

单文件建议不超过约 **64MB**；总行数（含表头）不超过 **50001**（与 `MaxTrainingRecordCSVImportRows` 一致）。

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

**成功响应**：`data` 为创建的 `training_record`（及关联档案）。若 `raw_flight_imports.raw_data` 为 JSON 且含 `{"source":"csv","field_map":{...}}`（或兼容字段名），认领时会将 `field_map` 中的始发/目的、任务类型、业务类型、表读数、视频字段、`landings` 等合并写入 `training_records`，并将 `stage`、`remark`/`progress_note`、`instructor_review` 等合并入 `training_archives`（与 CSV 自动导入一致）。

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

**请求体**：

- `student_id`（可选）：传入时按档案聚合一级科目的结业状态（`is_graduated` / `status`）。
- `syllabus_outline_id`（可选）：只返回该大纲下的一级科目。
- `outline_category`（可选）：按大纲类别过滤一级科目；与 `syllabus_outline_id` 勿同时使用。

大纲类别取值（与库内一致）：`spl_local`（SPL本场）、`spl_transition`（SPL转场）、`spl_night`（SPL夜航）、`ppl_basic`（PPL基本）、`ppl_night`（PPL夜航）。

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "syllabus_outline_id": "00000001-0000-4000-8000-000000000001"
}
```

或不传筛选（返回全部一级科目）：

```json
{}
```

**成功响应**：每项为一级科目，含 `syllabus_outline_id`、嵌套 `secondaries`（按 `sort_order` 排序；无则为 `[]`）。档案与结业统计仍只关联一级科目 `id`。

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "syllabus_outline_id": "00000001-0000-4000-8000-000000000001",
        "code": "SUB-01",
        "name": "起落",
        "sort_order": 1,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z",
        "is_graduated": false,
        "status": "not_started",
        "secondaries": [
          {
            "id": "550e8400-e29b-41d4-a716-446655440010",
            "syllabus_subject_id": "550e8400-e29b-41d4-a716-446655440000",
            "seq_no": 1,
            "name": "本场起落",
            "description": "",
            "completion_standard": "连续稳定完成",
            "sort_order": 0,
            "created_at": "2026-01-01T00:00:00Z",
            "updated_at": "2026-01-01T00:00:00Z"
          }
        ]
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

#### 7.7.6 创建一级科目（后台）`POST /api/v1/syllabus-subjects/admin/create`

- **说明**：创建某大纲下的一级科目。必须二选一：指定已有 `syllabus_outline_id`，或同时传 `outline_category` + `outline_name` 以新建一条大纲记录再挂科目。
- 可选 `secondaries`：同时创建二级科目。
- 同一大纲下 `code` 唯一。

**请求体示例（已有大纲）**：

```json
{
  "syllabus_outline_id": "00000001-0000-4000-8000-000000000001",
  "code": "SUB-01",
  "name": "起落",
  "sort_order": 1,
  "secondaries": [
    {
      "seq_no": 1,
      "name": "本场起落",
      "description": "科目描述",
      "completion_standard": "完成标准",
      "sort_order": 0
    }
  ]
}
```

**请求体示例（新建大纲）**：

```json
{
  "outline_category": "spl_local",
  "outline_name": "2025 SPL本场",
  "outline_sort_order": 0,
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
    "subject": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "syllabus_outline_id": "00000001-0000-4000-8000-000000000001",
      "code": "SUB-01",
      "name": "起落",
      "sort_order": 1,
      "created_at": "2026-03-16T00:00:00Z",
      "updated_at": "2026-03-16T00:00:00Z"
    },
    "secondaries": []
  }
}
```

#### 7.7.7 更新一级科目（后台）`POST /api/v1/syllabus-subjects/admin/update`

**请求体**：

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "syllabus_outline_id": "00000001-0000-4000-8000-000000000001",
  "code": "SUB-01",
  "name": "起落（更新后）",
  "sort_order": 2
}
```

- `id`：必填。
- `syllabus_outline_id`：可选，修改所属大纲。
- 其他字段可选，未提供的字段保持原值。

**成功响应**：返回更新后的完整一级科目对象（含 `syllabus_outline_id`）。

#### 7.7.8 删除一级科目（后台）`POST /api/v1/syllabus-subjects/admin/delete`

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

#### 7.7.9 大纲列表 `POST /api/v1/syllabus-outlines/list`

返回所有训练大纲实例（`id`、`category`、`name`、`sort_order` 等）。

**请求体**：`{}`

#### 7.7.10 导入大纲树（后台）`POST /api/v1/syllabus-outlines/admin/import`

单事务新建一条大纲，并批量创建其下一级科目与二级科目（总是新增一条大纲记录，不覆盖已有数据）。

**请求体**：

```json
{
  "category": "spl_local",
  "name": "SPL本场 导入版",
  "sort_order": 0,
  "subjects": [
    {
      "code": "01",
      "name": "地面准备",
      "sort_order": 0,
      "secondaries": [
        {
          "seq_no": 1,
          "name": "检查单",
          "description": "",
          "completion_standard": "独立完成",
          "sort_order": 0
        }
      ]
    }
  ]
}
```

**成功响应**：返回新建的大纲对象。

#### 7.7.11 创建二级科目（后台）`POST /api/v1/syllabus-subject-secondaries/admin/create`

**请求体**：`syllabus_subject_id`、`seq_no`、`name` 必填；可选 `description`、`completion_standard`、`sort_order`。

#### 7.7.12 更新二级科目（后台）`POST /api/v1/syllabus-subject-secondaries/admin/update`

**请求体**：`id` 必填；其余字段为要更新的内容。

#### 7.7.13 删除二级科目（后台）`POST /api/v1/syllabus-subject-secondaries/admin/delete`

**请求体**：`{ "id": "uuid" }`

#### 7.7.14 二级科目排序（后台）`POST /api/v1/syllabus-subject-secondaries/admin/reorder`

**请求体**：`syllabus_subject_id` 与 `ordered_ids`（该一级科目下全部二级科目 id，按新顺序排列）。事务内重写 `sort_order`。

```json
{
  "syllabus_subject_id": "550e8400-e29b-41d4-a716-446655440000",
  "ordered_ids": [
    "550e8400-e29b-41d4-a716-446655440020",
    "550e8400-e29b-41d4-a716-446655440021"
  ]
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
        "my_attempt_count": 3,
        "video_playback": {
          "kind": "gridfs",
          "url": "/api/v1/theory/documents/video?document_id=550e8400-e29b-41d4-a716-446655440201",
          "requires_auth": true,
          "content_type": "video/mp4"
        }
      }
    ],
    "total": 1
  }
}
```

> `has_video`：存在**外链**（`video_url`）或**托管**（GridFS/COS）任一则为 `true`。`video_playback` 无视频时为省略；`kind` 为 `external` 时 `url` 为完整外链，`requires_auth` 为 `false`；`kind` 为 `gridfs` 时 `url` 为**相对路径**，需拼接当前 API 根（如 `https://api.example.com`）后再请求，且须带 JWT（见 §8.x.6）。

#### 8.2 理论文档详情 `POST /api/v1/theory/documents/detail`

- **说明**：返回单个理论文档的元数据与内容块（图文 + 视频占位），不包含题目。若配置了课程视频，额外返回 `video_playback`（规则同 §8.1 列表）。

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
      "slug": "stall-spin-theory",
      "sort_order": 10,
      "content_ref": "",
      "video_url": "",
      "video_storage": "gridfs",
      "video_object_id": "674a1b2c3d4e5f6789012345",
      "video_content_type": "video/mp4",
      "video_size_bytes": 12582912,
      "video_duration_seconds": 0,
      "category": "空气动力学",
      "created_at": "2026-03-16T00:00:00Z",
      "updated_at": "2026-03-16T00:00:00Z"
    },
    "video_playback": {
      "kind": "gridfs",
      "url": "/api/v1/theory/documents/video?document_id=550e8400-e29b-41d4-a716-446655440201",
      "requires_auth": true,
      "content_type": "video/mp4"
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

> 托管视频时 `video_url` 通常为空，播放地址以 `video_playback` 为准。外链视频时 `video_storage` 为 `external`（或历史数据仅填 `video_url`），`video_playback.kind` 为 `external`。

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
    "video_storage": "external",
    "video_object_id": "",
    "video_content_type": "",
    "video_size_bytes": 0,
    "video_duration_seconds": 0,
    "category": "空气动力学",
    "created_at": "2026-03-16T00:00:00Z",
    "updated_at": "2026-03-16T00:00:00Z"
  }
}
```

> 说明：返回中的 `id` 即为后续题库录入（`document_id`）、试卷管理、答题接口使用的理论文档 ID。  
> 若填写 `video_url`，后端会写入 `video_storage: "external"`（外链）。**本地上传托管视频**请使用 §8.x.4，不要用本接口覆盖托管资源。

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
    "video_storage": "external",
    "video_object_id": "",
    "video_content_type": "",
    "video_size_bytes": 0,
    "video_duration_seconds": 0,
    "category": "空气动力学",
    "created_at": "2026-03-16T00:00:00Z",
    "updated_at": "2026-03-16T01:00:00Z"
  }
}
```

> **托管视频（GridFS）**：更新标题、排序等字段时照常调用本接口；**不要**通过本接口修改已托管的视频内容。若需换片源，请先 `POST .../video/remove` 再 `POST .../video/upload`，或仅覆盖上传（上传接口会自动替换 GridFS 文件）。  
> **外链视频**：仍通过本接口的 `video_url` 维护；非托管状态下清空 `video_url` 会清除外链。

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

#### 8.x.4 上传托管视频 `POST /api/v1/theory/documents/video/upload`

- **说明**：将视频文件流式写入 MongoDB GridFS（独立 bucket：`theory_videos`），并更新 PostgreSQL `theory_documents` 的 `video_storage=gridfs`、`video_object_id`、`video_content_type`、`video_size_bytes` 等。若该文档已有 GridFS 视频，先完成新文件上传与库表更新后再删除旧 GridFS 对象。  
- **前置**：已存在理论文档（§8.x.1）；数据库需已执行迁移 `020_theory_video_assets`（新增 `video_*` 列与约束）。  
- **认证**：`Authorization: Bearer <JWT>`（与其它需登录接口相同）。  
- **Content-Type**：`multipart/form-data`  
- **表单字段**：

| 字段 | 必填 | 说明 |
|------|------|------|
| `document_id` | 是 | 理论文档 UUID |
| `file` | 是 | 视频文件 |

- **类型校验**：请求的 `Content-Type` 为 `video/*`，或文件扩展名为 `.mp4`、`.webm`、`.mov`、`.m4v`、`.mkv`（之一）。  
- **成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "status": "ok"
  }
}
```

> 运行环境：默认 `THEORY_VIDEO_STORAGE=gridfs`（或未设置）。若设为 `cos`，当前服务端会启动失败（COS 驱动待实现），部署前请保持 `gridfs` 并完成 MongoDB 与迁移。

#### 8.x.5 移除视频 `POST /api/v1/theory/documents/video/remove`

- **说明**：若当前为托管视频（GridFS），删除对象存储中的文件并清空相关列；若为外链，清空 `video_url` 与 `video_storage`。  
- **认证**：`Authorization: Bearer <JWT>`  

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
    "status": "ok"
  }
}
```

#### 8.x.6 播放托管视频 `GET /api/v1/theory/documents/video`

- **说明**：仅适用于 `video_storage=gridfs` 的文档；返回原始视频字节流，便于 `<video>` 或播放器拉流。支持 HTTP `Range` 与 `206 Partial Content`、`Accept-Ranges: bytes`（拖动进度条依赖此行为）。  
- **认证**：`Authorization: Bearer <JWT>`（Query 无法带 Body，须用 Header）。  
- **Query**：`document_id=<UUID>`（必填）。  
- **响应头**：`Content-Type` 一般为 `video/mp4`（或与上传时一致）；分段返回时含 `Content-Range`。  
- **错误**：无托管视频时 `404`；`video_storage=cos` 尚未实现时 `501`。若库表里 `video_object_id` 在 MongoDB GridFS（bucket `theory_videos`）中已不存在（库表与存储不一致、误删 Mongo、连错库等），同样返回 `404`，提示需重新上传。  
- **客户端**：列表/详情中的 `video_playback.url` 为**相对路径**（如 `/api/v1/theory/documents/video?document_id=...`），需拼接 API 根地址后再请求，并携带与业务接口相同的 JWT。

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

> 说明：本节为基于当前数据库结构的后台“学员信息管理”接口设计，实际落地时可按本规范实现到对应 Handler 中。与学员 / 教员 / 地面职员相关的**统一软删除**见 **§13.4**，**统一新建/编辑**见 **§13.5**（均仅 `admin`）。**任职状态**维护见 **§13.3**、**§13.5** 的 `employment_status` 字段。

#### 13.1 学员列表（后台）`POST /api/v1/admin/students/list`

- **认证**：需要；**仅** `role = "admin"` 可访问（与其它 `POST /api/v1/admin/*` 一致）。
- **说明**：用于后台表格展示所有学员的基础信息与训练/合同摘要（含离职/兼职学员，便于管理员维护）。

**请求体**：

```json
{
  "keyword": "",
  "responsible_instructor_id": "",
  "flight_training_status": "",
  "page": 1,
  "page_size": 20
}
```

- `keyword`、`responsible_instructor_id`、`flight_training_status`：**可选**，语义与 **§4.1** 学员列表一致（后台列表含离职/兼职学员，`onlyActiveEmployment` 为关闭）。

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
        "employment_status": "active",
        "personnel_category": "flight",
        "contract_duration_minutes": 3600,
        "flown_minutes": 1200,
        "remaining_minutes": 2400,
        "training_progress": "已完成单飞阶段",
        "resident_city": "上海",
        "flight_training_status": "in_progress",
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

- 无关联账号（学员未绑定 `user_id`）时，`username` / `phone` / `email` / `employment_status` / `personnel_category` 通常为空或省略。
- `total`：当前实现为**本页**返回条数（与 `page`/`page_size` 一致），非全表总数。

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
      "resident_city": "上海",
      "contract_duration_minutes": 3600,
      "flown_minutes": 1200,
      "remaining_minutes": 2400,
      "training_progress": "已完成单飞阶段",
      "flight_training_status": "in_progress"
    },
    "account": {
      "username": "student1",
      "phone": "13800000000",
      "email": "s1@example.com",
      "role": "student",
      "employment_status": "active",
      "personnel_category": "flight"
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

- **认证**：需要；**仅** `admin`。
- **说明**：用于后台创建或编辑学员信息；实现上与 **§13.5** `kind = "student"` 相同（本接口**无需**传 `kind`，服务端固定按学员处理）。创建时维护 `users`（`role = student`）与 `students`；**更新**时若学员已关联 `user_id`，可同时修改该账号的 `username` / `phone` / `email` / `avatar_url` / `password`（密码非空则重置）及 **`employment_status`**。

**「先录基本信息」与请求字段对照**（可与合同、进度等分步保存；未列字段见下文可选说明）：

| 业务项 | 请求 JSON 字段 | 持久化位置 |
|--------|----------------|------------|
| 姓名 | `name` | `students.name`（必填） |
| 电话 | `phone` | 关联 `users.phone`（有账号时写入） |
| 学员编号 | `student_no` | `students.student_no`（必填） |
| 责任教员 | `responsible_instructor_id` | `students.responsible_instructor_id`（UUID，可空） |
| 登录账号 | `username` | `users.username` |
| 登录密码 | `password` | `users.password_hash`（创建时可空则无法密码登录直至后续设置；更新时非空则重置） |
| 常驻城市 | `resident_city` | `students.resident_city`（见下行语义） |
| 学飞状态 | `flight_training_status` | `students.flight_training_status`（见下行枚举） |

- **`resident_city`（常驻城市）**：字符串，可空。**新建**时请求中省略或 `null` 表示空；**更新**时请求体**省略该字段**或 **`resident_city`: `null`** 表示不修改原值；显式传 `""` 可清空 city。
- **`flight_training_status`**：取值 `in_progress` | `paused` | `completed` | `terminated`。**新建**时省略默认为 `in_progress`；**更新**时省略则不修改。非法取值返回 `400`。

**请求体（示例）**：

```json
{
  "student_id": null,
  "name": "张三",
  "student_no": "S000001",
  "resident_city": "上海",
  "username": "student1",
  "phone": "13800000000",
  "email": "s1@example.com",
  "password": "secret",
  "employment_status": "active",
  "flight_training_status": "in_progress",
  "contract_duration_minutes": 3600,
  "flown_minutes": 0,
  "training_progress": "",
  "training_progress_json": { "stage": "PPL", "items": [] },
  "responsible_instructor_id": "550e8400-e29b-41d4-a716-446655440002",
  "contract_info": {
    "learning_points": 120,
    "guest_experience": { "quota": 5, "used": 0 }
  }
}
```

- 可选字段 **`employment_status`**：取值 `active` | `resigned` | `part_time`（亦接受 `part-time`，存为 `part_time`）。**新建**时省略则默认为 `active`；**更新**时省略则不修改关联账号的任职状态。非法取值返回 `400`。
- 可选字段 `contract_info`：合法 JSON 时写入 `students.contract_info`（JSONB）；请求中**未携带**该字段或值为空白时不修改（更新场景下保留原值）；**携带**且为非空 JSON 时覆盖（创建与更新均支持）。非法 JSON 返回 `400`（`msg: invalid contract_info json`）。
- 可选字段 `training_progress_json`：合法 JSON 时写入 `students.training_progress_json`（JSONB）；**未携带**或值为空白时，更新场景下**不修改**原值；语义与 `contract_info` 相同。非法 JSON 返回 `400`（`msg: invalid training_progress_json`）。
- 当 `student_id = null` 时，后端将：
  - 在 `users` 表中创建一条 `role = "student"` 的账号记录；
  - 在 `students` 表中创建一条记录并关联 `user_id`；
  - 返回新建的学员主键与账号信息。
- 当 `student_id` 非空时，更新已存在学员档案；若存在关联账号，同步更新账号字段（见上文）。

#### 13.4 人员软删除（后台）`POST /api/v1/admin/personnel/delete`

- **认证**：需要；**仅** `role = "admin"` 可调用，否则 HTTP `403`（`code: 403`，`msg: forbidden`）。
- **说明**：对学员、教员或地面职员账号做**软删除**（数据库字段 `deleted_at`），不物理删行。与 **`employment_status`** 独立：日常「离职 / 兼职」可只改任职状态保留档案；软删多用于彻底从业务列表移除并允许用户名/学号复用。软删除后：
  - 学员 / 教员列表、用户按角色列举等常规查询**不再包含**该人员；
  - 已软删除账号**无法**再通过用户名密码登录（`POST /auth/login` 查不到有效用户）；
  - 唯一约束（用户名、学号等）在「未删除」范围内生效，后续可复用相同用户名 / 学号新建档案（见迁移 `026_personnel_soft_delete`）。

**请求体**：

```json
{
  "kind": "student",
  "id": "550e8400-e29b-41d4-a716-446655440001"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `kind` | string | 是 | `student`：学员，此时 `id` 为 `students.id`；`instructor`：教员，此时 `id` 为 `instructors.id`；`staff`：地面工作人员，此时 `id` 为 `users.id`，且该用户角色须为 `staff` 或 `maintenance` |
| `id` | string (UUID) | 是 | 与 `kind` 对应的主键 |

**行为摘要**：

- `kind = "student"`：标记 `students.deleted_at`；若该学员有关联 `users`（`user_id` 非空），同时标记对应 `users.deleted_at`。
- `kind = "instructor"`：标记 `instructors.deleted_at` 及对应 `users.deleted_at`（教员账号一对一）。
- `kind = "staff"`：仅标记 `users.deleted_at`；若目标用户不是 `staff` / `maintenance`，返回 HTTP `403`，`msg` 为 `personnel kind staff requires user role staff or maintenance`。

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": null
}
```

**常见错误**：

| HTTP | `code` | 说明 |
|------|--------|------|
| `400` | `400` | `kind` 非法、JSON 绑定失败、`id` 为空等 |
| `403` | `403` | 非 `admin` 调用；或 `kind = "staff"` 但用户角色不是 `staff`/`maintenance` |
| `404` | `404` | 目标不存在或已被软删除（`msg: not found`） |

#### 13.5 人员新建 / 编辑（后台）`POST /api/v1/admin/personnel/save`

- **认证**：需要；**仅** `role = "admin"` 可调用，否则 HTTP `403`。
- **说明**：用同一套协议维护学员、教员、地面职员（`staff` / `maintenance`）档案及对应 `users` 账号字段。与 **§13.3** 学员专用接口二选一即可；本接口便于前台「人员管理」统一表单。

**请求体**：须带 `kind`；其余字段按 `kind` 选填或必填，见下表。

| `kind` | 新建 | 更新 | 必填与主键 |
|--------|------|------|------------|
| `student` | `student_id` 省略或空 | `student_id` = `students.id` | 新建：`name`、`student_no`；更新：同上。可选：`username`、`phone`、`email`、`password`、`avatar_url`（有 `user_id` 时生效）、**`employment_status`**（`active` / `resigned` / `part_time`）、**`resident_city`**（语义见 §13.3）、**`flight_training_status`**、`contract_duration_minutes`、`flown_minutes`、`training_progress`、`training_progress_json`、`responsible_instructor_id`、`contract_info`（其余语义见 §13.3） |
| `instructor` | `instructor_id` 省略或空 | `instructor_id` = `instructors.id` | 新建：`username`、`name`、`password`（非空）、可选 `phone`、`email`、`display_color`、`avatar_url`、**`employment_status`**。更新：教员 `name` / `display_color` 及账号字段（非空字段覆盖）；可选 **`employment_status`**（有值则更新） |
| `staff` | （本接口不单独创建地面账号，请用注册或其它流程） | `user_id` = `users.id` | **仅更新**：`user_id` 必填；目标用户须为 `staff` 或 `maintenance`；可改 `username`、`phone`、`email`、`avatar_url`、`password`、**`employment_status`** |

**请求示例（新建教员）**：

```json
{
  "kind": "instructor",
  "username": "ins01",
  "password": "initial-secret",
  "name": "教员A",
  "phone": "13800000000",
  "email": "ins@example.com",
  "display_color": "#3366cc",
  "avatar_url": "",
  "employment_status": "active"
}
```

**请求示例（编辑地面职员）**：

```json
{
  "kind": "staff",
  "user_id": "550e8400-e29b-41d4-a716-446655440099",
  "username": "maint01",
  "phone": "13900000000",
  "email": "m@example.com",
  "employment_status": "part_time"
}
```

**成功响应**：`data` 为对象，仅与本次 `kind` 相关的字段有值，例如：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "kind": "instructor",
    "instructor": {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "user_id": "550e8400-e29b-41d4-a716-446655440003",
      "name": "教员A",
      "display_color": "#3366cc",
      "created_at": "...",
      "updated_at": "..."
    },
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "username": "ins01",
      "phone": "13800000000",
      "email": "ins@example.com",
      "role": "instructor",
      "employment_status": "active",
      "avatar_url": "",
      "created_at": "...",
      "updated_at": "..."
    }
  }
}
```

- `kind = "student"` 时 `data.student` 必有；`user` 仅在创建学员账号或更新后需要带出账号信息时出现（与实现一致）。
- 返回中的 `user` **不含** `password_hash` 字段。

**常见错误**：

| HTTP | 说明 |
|------|------|
| `400` | `kind` 缺失或非法；新建教员未提供 `password`；学员新建缺少 `name`/`student_no`；`username` 已被他人占用（`msg: username already exists`）；`employment_status` 非法；`contract_info` / `training_progress_json` 非法 JSON |
| `403` | 非 `admin`；`kind = "staff"` 但目标用户不是 `staff`/`maintenance` |
| `404` | 更新的 `student_id` / `instructor_id` / `user_id` 不存在或已软删除 |

---

### 14. 各角色首页聚合 API

> 说明：本节与训练侧《教员 / 学员 / 管理后台主页 — 数据库与聚合接口设计（Golang）》对齐。该设计稿示例为 `GET /home/student` 等形式；**本服务**与文档文首约定一致：除 `/health` 外均为 **`POST`、`Content-Type: application/json`**，响应体为 `code` / `msg` / `data`。

#### 与设计稿路径对照

| 设计稿（示例） | 本仓库路径 |
|----------------|------------|
| `GET /home/student` | `POST /api/v1/students/my-home` |
| `GET /home/instructor` | `POST /api/v1/instructors/my-home` |
| `GET /home/admin` | `POST /api/v1/admin/home` |

设计稿中的 Query 参数（如 `posts_limit`）在本仓库中放在**请求体**字段中。

#### 通用约定

- **部分失败**：某一子块（如天气上游超时）失败时，仍返回 HTTP `200` 且 `code: 200`；对应字段可为 `null` 或空数组。可选在 `data.warnings` 中返回字符串数组（如 `["weather_upstream_timeout"]`）。客户端应容忍缺失块。
- **权限与错误码**：未登录返回 HTTP `401`，`code` 与 HTTP 状态一致、`msg` 说明原因；角色不匹配（例如学员调用教员或管理员首页接口）返回 HTTP `403`。与本文档其余接口相同，**不单独约定** `40101` / `40301` 等细粒度业务码，除非后续全项目统一。

| 接口 | 要求的 `role`（最低约定） |
|------|---------------------------|
| `POST /api/v1/students/my-home` | `student` |
| `POST /api/v1/instructors/my-home` | `instructor` |
| `POST /api/v1/admin/home` | `admin` |
| `POST /api/v1/admin/students/*`、`POST /api/v1/admin/contracts/*` | **仅** `admin` |
| `POST /api/v1/admin/personnel/delete`、`POST /api/v1/admin/personnel/save` | **仅** `admin`（见 §13.4、§13.5） |
| 路径含 **`/admin/`** 的训练/大纲/题库管理接口（如 `POST /api/v1/syllabus-subjects/admin/*`、`POST /api/v1/theory/questions/admin/*` 等） | **仅** `admin` |

未登录统一 `401`；已登录但非 `admin` 访问上表中的管理接口为 **`403`**。

---

#### 14.1 学员首页 `POST /api/v1/students/my-home`

- **认证**：需要，`role = "student"`；非学员角色返回 `403`。
- **说明**：学员端首页**一次性拉取**聚合数据，包含两类内容：
  1. **训练 / 学业块**：档案、任务、近期训练计划与档案、理论学习与科目概要（**已部分实现**，见下方实现状态）。
  2. **门户 / 运营块**：`user`、社区动态、推荐专题等已部分接入；天气、好物、通知、活动仍为空（见实现状态表）。

**请求体**

最小请求：

```json
{}
```

可选字段（服务端可忽略暂未支持的字段）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `locale` | string | 多语言内容筛选，如 `zh-CN` |
| `posts_limit` | int | `community_posts` 条数，默认 `4`，最大 `10` |
| `articles_limit` | int | `featured_articles` 条数，默认 `4` |

示例：

```json
{
  "locale": "zh-CN",
  "posts_limit": 4,
  "articles_limit": 4
}
```

**成功响应 `data` 目标结构示例**（训练块 + 门户块合并；未接入字段为 `null` 或空数组）：

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
    },
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "display_name": "张三",
      "avatar_url": "https://example.com/avatar.png"
    },
    "weather_snippet": {
      "temp_c": 24,
      "wind": "3m/s NW",
      "flight_condition": "CAVOK"
    },
    "community_posts": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440501",
        "author_name": "李教员",
        "author_avatar": null,
        "content": "今日本场训练注意事项 …",
        "created_at": "2026-03-26T02:00:00Z",
        "likes": 12,
        "comments": 3
      }
    ],
    "featured_articles": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440201",
        "title": "如何完美处理侧风着陆",
        "category": "飞行技巧",
        "cover_image_url": "https://example.com/cover.jpg"
      }
    ],
    "featured_products": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440601",
        "name": "降噪耳机",
        "price_display": "¥8,999",
        "image_url": "https://example.com/p.jpg"
      }
    ],
    "notifications": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440701",
        "title": "明日气象提醒",
        "type": "warning",
        "time_label": "今天 09:00"
      }
    ],
    "events": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440801",
        "title": "飞行开放日",
        "date_label": "03-20",
        "status": "open"
      }
    ],
    "warnings": []
  }
}
```

- **`notifications[].type`**：建议取值 `info` | `warning` | `success`（与公告严重程度一致即可）。
- **`events[].status`**：建议取值 `open` | `full` | `closed`。

**实现状态**

| 字段 | 状态 |
|------|------|
| `profile`、`tasks`、`upcoming_training`、`last_training_archive` | 已实现 |
| `theory_summary`、`subject_progress` | 已返回结构，当前后端统计可能为简化/占位，可后续对接 §8、大纲科目接口 |
| `user`、`community_posts`、`featured_articles`、`warnings` | **已实现**：`user` 来自 `users`；`community_posts` 为广场 `global` 前 N 条（见 `posts_limit`）；`featured_articles` 为理论文档列表前 N 条（`cover_image_url` 暂空）；子块失败时 `warnings` 含如 `community_posts_unavailable` |
| `weather_snippet`、`featured_products`、`notifications`、`events` | **未实现**：当前为 `null` 或空数组 |

**关联接口**

- 社区动态列表与翻页：**§17.3.2** `POST /api/v1/community/posts/list`。
- 推荐专题数据源可对齐 **§8.1** `POST /api/v1/theory/documents/list`（精选规则由聚合层或后续列表筛选参数扩展）。

**后端组装建议**：`students` / `users`；**§9.1** `POST /api/v1/tasks/my-list` 取预览条数；`training_plans` 未来 7 天；**§7.6.1** `POST /api/v1/students/training-archives/list` 取最近一条；`theory_attempts` 与 **§8** 理论模块统计；**§7.7.1** `POST /api/v1/syllabus-subjects/list`（带学员）统计科目；门户块可并行查询社区帖子、理论精选、公告/活动表或外部天气服务，失败时写入 `warnings` 并置空对应块。

---

#### 14.2 教员首页 `POST /api/v1/instructors/my-home`

- **认证**：需要，`role = "instructor"`；非教员角色返回 `403`。
- **路由**：已在 `POST /api/v1/instructors/my-home` 注册。当前实现要点：`stats` 由 `training_records` 按教员 ID 与日期范围汇总；`active_students_count` 为**近 365 天**与该教员有训练记录的去重学员数；`rank`、`passing_rate_percent`、`trend_*` 为占位或空串；`flight_time_distribution` 由本年度记录按 `flight_type`/`task_type` 等关键词粗分桶；`my_posts` 为社区中 `author_user_id = 当前用户` 的动态。

**请求体**

```json
{}
```

可选（与学员侧类似，供 `my_posts` 条数等扩展）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `posts_limit` | int | 教员本人动态 `my_posts` 条数，默认 `4`，最大 `10` |

**成功响应 `data` 结构示例**

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "profile": {
      "user_id": "550e8400-e29b-41d4-a716-446655440002",
      "display_name": "李教员",
      "avatar_url": "https://example.com/a.png",
      "title": "资深飞行教员",
      "tagline": "飞行时长 2500+ 小时",
      "rank": {
        "code": "SENIOR_FI",
        "label": "高级飞行教员",
        "progress_percent": 75,
        "hint": "距离特级教员还需 500 小时带飞时长"
      }
    },
    "stats": {
      "annual_flight_hours": 450.5,
      "monthly_flight_hours": 42.0,
      "active_students_count": 12,
      "passing_rate_percent": 95,
      "trend_annual": "+12% 较去年",
      "trend_monthly": "目标完成 70%",
      "trend_students": "3名即将结业",
      "trend_passing": "全校前 5%"
    },
    "flight_time_distribution": [
      { "type": "dual", "label": "带飞 (Dual)", "percent": 85 },
      { "type": "xc", "label": "转场 (XC)", "percent": 60 },
      { "type": "ifr", "label": "仪表 (IFR)", "percent": 40 },
      { "type": "night", "label": "夜航 (Night)", "percent": 25 }
    ],
    "my_posts": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440502",
        "content": "本场训练总结 …",
        "image_url": "https://example.com/m.jpg",
        "created_at": "2026-03-26T00:00:00Z",
        "likes": 45,
        "comments": 12
      }
    ],
    "warnings": []
  }
}
```

**数据源说明（实现参考）**：年度/月度飞行小时由 `flight_records`（教员维度、日期范围）汇总；在训人数来自教员—学员分配关系（如 `status = active`）；通过率来自考试/训练结果聚合；`flight_time_distribution` 可由分段表或月度快照预计算；`trend_*` 可为模板文案或后续改为结构化字段（如同比百分比）。教员动态可与社区帖子同源（`author_id = 当前用户`），列表扩展见 **§17.3.2**。

---

#### 14.3 管理后台首页 `POST /api/v1/admin/home`

- **认证**：需要，`role = "admin"`（与 **§13** 等管理端接口一致）；非 `admin` 返回 `403`。
- **路由**：已在 `POST /api/v1/admin/home` 注册。当前实现要点：`kpi.monthly_total_flight_hours` 为当月**正常状态**训练记录飞行分钟合计换算整数小时；`active_students_count` 为 `students` 表分页全量计数；`pending_reports_count` 固定 `0`；`system_alerts` 为空；`modules` 为内置默认模块列表。

**请求体**

```json
{}
```

**成功响应 `data` 结构示例**

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "admin_user": {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "display_name": "管理员"
    },
    "kpi": {
      "monthly_total_flight_hours": 1240,
      "monthly_total_flight_hours_trend": "+8.5%",
      "active_students_count": 156,
      "active_students_delta": "+12",
      "pending_reports_count": 2
    },
    "modules": [
      {
        "key": "student_management",
        "title": "学员管理",
        "description": "学员档案、进度、合同与权益",
        "path": "/admin/student-management",
        "icon": "graduation_cap"
      }
    ],
    "system_alerts": {
      "summary": "当前有 3 架飞机即将到达定检周期，2 名教员的执照即将到期。",
      "items": [
        {
          "type": "aircraft_inspection",
          "severity": "warning",
          "count": 3,
          "deeplink": "/admin/aircraft-schedule"
        },
        {
          "type": "instructor_license",
          "severity": "warning",
          "count": 2,
          "deeplink": "/admin/personnel-schedule"
        }
      ]
    },
    "warnings": []
  }
}
```

**说明**

- **`modules`**：可由后端下发（便于版本控制入口与路径），也可仅由前端静态配置；若后端不返回，客户端可使用本地模块表。
- **`kpi`**：优先来自定时任务写入的快照表，实时性要求高时可叠加当日增量。
- **`system_alerts`**：可由合规/定检/执照等规则聚合；`summary` 可由服务端拼接，或仅返回 `items` 由前端做文案国际化。

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

### 16. 学员合同管理 API

> 说明：基于 PostgreSQL `contracts` 表（含 `payment_mode`、`payment_schedule`），提供后台创建、更新、详情与按学员列表。与文首权限表一致，本节 **`/api/v1/admin/contracts/*`** 当前实现为 **仅 `admin`**。
>
> **付款方式**：
> - `payment_mode`：`full`（全款，默认）或 `installment`（分期）。
> - **全款**：不传或传 `payment_mode: "full"` 时，`payment_schedule` 应为空数组或省略。
> - **分期**：`payment_mode` 为 `installment` 时，`payment_schedule` 为数组，**至少 2 期、至多 24 期**；每期包含 `amount`（>0）与 `due_date`（`YYYY-MM-DD`）。**各期 `amount` 之和须与 `price_total` 一致**（允许 0.01 以内误差）。分期时必须提供 `price_total`。
>
> 校验失败时返回 `400`，`msg` 可能为：`invalid payment_mode`、`invalid payment_schedule`、`payment_schedule sum must match price_total`、`price_total required for installment contract`。

#### 16.1 创建合同 `POST /api/v1/admin/contracts/create`

- **认证**：需要（仅 `admin`）。
- **说明**：为指定学员创建一份新的训练合同。

**请求体（全款示例）**：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "contract_no": "HT-2026-001",
  "sign_date": "2026-03-20",
  "start_date": "2026-04-01",
  "end_date": "2027-04-01",
  "total_minutes": 3600,
  "price_total": 150000.0,
  "currency": "CNY",
  "status": "active",
  "payment_mode": "full",
  "payment_schedule": [],
  "remark": "私商一体训练合同"
}
```

**请求体（分期三期示例，`payment_schedule` 长度须与期数一致）**：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "contract_no": "HT-2026-002",
  "sign_date": "2026-03-20",
  "start_date": "2026-04-01",
  "end_date": "2027-04-01",
  "total_minutes": 3600,
  "price_total": 150000.0,
  "currency": "CNY",
  "status": "active",
  "payment_mode": "installment",
  "payment_schedule": [
    { "amount": 50000, "due_date": "2026-04-01" },
    { "amount": 50000, "due_date": "2026-07-01" },
    { "amount": 50000, "due_date": "2026-10-01" }
  ],
  "remark": "分期付款"
}
```

**成功响应（示例）**：`data` 为合同对象，字段包含 `payment_mode`、`payment_schedule`、`mongo_file_ids`、`created_at`、`updated_at` 等（日期字段为 `YYYY-MM-DD` 字符串，`created_at`/`updated_at` 为 RFC3339 时间）。

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
    "price_total": 150000.0,
    "currency": "CNY",
    "status": "active",
    "payment_mode": "full",
    "payment_schedule": [],
    "mongo_file_ids": [],
    "remark": "私商一体训练合同",
    "created_at": "2026-03-20T10:00:00Z",
    "updated_at": "2026-03-20T10:00:00Z"
  }
}
```

#### 16.2 更新合同 `POST /api/v1/admin/contracts/update`

- **认证**：需要（仅 `admin`）。
- **说明**：更新合同的关键信息；`id` 必填，其余字段可选，未提供的字段保持原值。

**请求体（示例）**：

```json
{
  "id": "c-uuid-1",
  "end_date": "2027-06-01",
  "total_minutes": 4000,
  "price_total": 168000.0,
  "status": "active",
  "payment_mode": "installment",
  "payment_schedule": [
    { "amount": 84000, "due_date": "2026-05-01" },
    { "amount": 84000, "due_date": "2026-11-01" }
  ],
  "remark": "延长 2 个月，追加 400 分钟，改为分期"
}
```

- 若改为分期或修改分期金额，须重新满足「期数与金额合计」规则。

**成功响应（示例）**：`data` 为更新后的完整合同对象（结构同 16.1 响应 `data`）。

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
    "price_total": 168000.0,
    "currency": "CNY",
    "status": "active",
    "payment_mode": "installment",
    "payment_schedule": [
      { "amount": 84000, "due_date": "2026-05-01" },
      { "amount": 84000, "due_date": "2026-11-01" }
    ],
    "mongo_file_ids": ["file-meta-id-1"],
    "remark": "延长 2 个月，追加 400 分钟，改为分期",
    "created_at": "2026-03-20T10:00:00Z",
    "updated_at": "2026-03-21T08:00:00Z"
  }
}
```

#### 16.3 合同详情 `POST /api/v1/admin/contracts/detail`

- **认证**：需要（仅 `admin`）。
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
      "price_total": 168000.0,
      "currency": "CNY",
      "status": "active",
      "payment_mode": "installment",
      "payment_schedule": [
        { "amount": 84000, "due_date": "2026-05-01" },
        { "amount": 84000, "due_date": "2026-11-01" }
      ],
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

> 说明：`files` 来自 `contract_files_meta` 与 `contracts.mongo_file_ids` 的关联，结构同 15.2。

#### 16.4 按学员查询合同列表 `POST /api/v1/admin/contracts/list-by-student`

- **认证**：需要（仅 `admin`）。
- **说明**：在学员详情页中列出该学员的全部合同（含历史合同），按 `created_at` 倒序。

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
        "price_total": 120000.0,
        "currency": "CNY",
        "status": "finished",
        "payment_mode": "full",
        "payment_schedule": []
      },
      {
        "id": "c-uuid-1",
        "contract_no": "HT-2026-001",
        "sign_date": "2026-03-20",
        "start_date": "2026-04-01",
        "end_date": "2027-06-01",
        "total_minutes": 4000,
        "price_total": 168000.0,
        "currency": "CNY",
        "status": "active",
        "payment_mode": "installment",
        "payment_schedule": [
          { "amount": 84000, "due_date": "2026-05-01" },
          { "amount": 84000, "due_date": "2026-11-01" }
        ]
      }
    ]
  }
}
```

> 说明：列表项含 `payment_mode`、`payment_schedule`，便于后台展示全款/分期与计划。

#### 16.5 合同状态变更（可选）`POST /api/v1/admin/contracts/change-status`

- **认证**：需要。
- **说明**：用于将合同标记为已结束、已作废等，配合 `status` 字段使用。**当前服务端未实现该路径**；可暂用 `POST /api/v1/admin/contracts/update` 传入 `id`、`status`、`remark` 达到同等效果。

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

---

## 17. 社区模块 API

本节描述**动态广场、飞行故事（发帖）、热门话题、兴趣小组、飞行挑战赛、互相关注与推荐关注**等能力。所有接口均为 **`POST`**，路径前缀 **`/api/v1/community/`**，需 **`Authorization: Bearer <JWT>`**，请求体 `Content-Type: application/json`，返回结构同文档开头约定。

**数据库**：迁移 `019_community.up.sql` 创建 `user_follows`、`community_topics`、`community_groups`、`community_group_members`、`community_posts`（含 `media_urls` JSONB）、`community_post_topics`、`community_challenges`、`community_challenge_entries`、`community_post_likes`、`community_comments` 等表；话题帖子数、小组成员数、点赞数、评论数由触发器维护。

### 17.1 领域约定

- **业务对象**：
  - **动态与故事**：底层对应同一 `community_posts` 记录；`status` 为 `published` 时对满足权限的用户可见；`status` 为 `hidden` 时仅作者可见。
  - **多维可见性**：
    - `public`：公开广场可见。
    - `followers_only`：仅作者本人的关注者（粉丝）可见。
    - `group_only`：仅特定兴趣小组成员可见，须关联有效 `group_id`。
- **角色权限**：
  - **内容运营**：创建话题 `topics/create`、管理挑战赛 `challenges/create|update` 仅限 **`admin` / `staff`**。
  - **社区自治**：任意用户可创建兴趣小组；小组组长拥有小组信息及成员管理权。
- **资源关联**：动态可同时关联多个话题及最多一个兴趣小组。

### 17.2 分页与游标机制

社区模块采用**游标分页（Cursor Pagination）**以防止列表滚动时产生重复内容。

- **控制参数**：
  - `page_size`：每页项数，默认 `20`，最大支持 `50`。
  - `next_cursor`：分页标识符。首页传空；获取下一页时，传回上一页响应中的 `next_cursor`。
- **成功响应约定**：
  - `data.items`：列表数据。
  - `data.next_cursor`：下一页游标。若无更多数据，该字段为空或不返回。
- **实现细节（透明可选）**：
  - 游标格式通常为 `RFC3339Nano|post_id`；后端自动解析为条件 `(created_at, id) < (cursor_time, cursor_id)` 进行高效检索。

### 17.3 动态（飞行故事）

#### 17.3.1 发帖 `POST /api/v1/community/posts/create`

- **认证**：需要。
- **说明**：支持选择可见性，并可关联多个话题或归属于某个兴趣小组。

**请求体**：

```json
{
  "title": "今日首航感悟",
  "body": "第一次独立起飞的感觉太棒了...",
  "media_urls": ["https://oss.example.com/p1.jpg", "https://oss.example.com/p2.jpg"],
  "visibility": "public", // public | followers_only | group_only
  "group_id": "", // 当 visibility 为 group_only 时必填
  "topic_ids": ["topic-uuid-1", "topic-uuid-2"]
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": "post-uuid-1",
    "author_user_id": "user-uuid-1",
    "title": "今日首航感悟",
    "body": "...",
    "media_urls": [...],
    "visibility": "public",
    "like_count": 0,
    "comment_count": 0,
    "created_at": "2026-03-26T10:00:00Z"
  }
}
```

#### 17.3.2 动态列表 `POST /api/v1/community/posts/list`

- **认证**：需要。

**请求体**：

```json
{
  "tab": "global", // global | following | group
  "group_id": "", // tab=group 时必填
  "topic_id": "", // 若传则忽略 tab，按话题检索
  "author_user_id": "", // 拉取特定用户的贴
  "page_size": 20,
  "next_cursor": "" // RFC3339Nano|post_id 游标
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
        "post": {
          "id": "post-uuid-1",
          "title": "...",
          "body": "...",
          "media_urls": [],
          "like_count": 10,
          "comment_count": 2,
          "liked_by_me": true,
          "topic_ids": ["..."],
          "created_at": "..."
        },
        "author": {
          "id": "user-uuid-1",
          "username": "pilot_007",
          "avatar_url": "...",
          "role": "student"
        }
      }
    ],
    "next_cursor": "2026-03-26T09:00:00Z|uuid"
  }
}
```

#### 17.3.3 点赞动态 `POST /api/v1/community/posts/like`

**请求体**：`{"post_id": "uuid"}`

**成功响应**：`{"code": 200, "msg": "ok", "data": {"like_count": 11}}`

---

### 17.4 评论

#### 17.4.1 发表评论 `POST /api/v1/community/comments/create`

**请求体**：

```json
{
  "post_id": "post-uuid-1",
  "body": "写得真好，期待下一次飞行！",
  "parent_id": "" // 可选，回复某条评论的 ID
}
```

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": "comment-uuid-1",
    "post_id": "...",
    "body": "...",
    "author_user_id": "...",
    "created_at": "..."
  }
}
```

#### 17.4.2 评论列表 `POST /api/v1/community/comments/list`

**请求体**：

```json
{
  "post_id": "post-uuid-1",
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
        "id": "comment-uuid-1",
        "body": "...",
        "author": { "id": "...", "username": "...", "avatar_url": "..." },
        "parent_id": null,
        "created_at": "..."
      }
    ]
  }
}
```

---

### 17.5 关注与推荐

#### 17.5.1 关注/取消关注 `POST /api/v1/community/follows/follow`

**说明**：内部根据当前状态切换关注/取消。

**请求体**：`{"followee_id": "user-uuid-2"}`

**成功响应**：`{"code": 200, "msg": "ok", "data": {"status": "following"}}`

#### 17.5.2 推荐关注 `POST /api/v1/community/follows/recommend`

**请求体**：`{"limit": 10}`

**成功响应**：`{"code": 200, "msg": "ok", "data": {"users": [{ "id": "uuid", "username": "...", "reason": "同组成员" }]}}`

---

### 17.6 话题

#### 17.6.1 话题列表 `POST /api/v1/community/topics/list`

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "topic-uuid-1",
        "slug": "first-flight",
        "title": "这就是飞行",
        "description": "分享你的飞行故事",
        "cover_url": "...",
        "post_count": 128
      }
    ]
  }
}
```

#### 17.6.2 创建话题 `POST /api/v1/community/topics/create`

- **权限**：仅 `admin`、`staff`。

**请求体**：`{"title": "蓝天之上", "slug": "above-sky", "description": "...", "cover_url": "..."}`

---

### 17.7 兴趣小组

#### 17.7.1 创建小组 `POST /api/v1/community/groups/create`

**请求体**：

```json
{
  "name": "塞斯纳爱好者",
  "slug": "cessna-fans",
  "description": "交流塞斯纳 172 飞行技巧",
  "cover_url": "...",
  "is_public": true
}
```

---

### 17.8 飞行挑战赛

#### 17.8.1 挑战赛列表 `POST /api/v1/community/challenges/list`

**请求体**：`{"status": "active"}` // active | all

**成功响应**：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "challenge-uuid-1",
        "title": "首飞 100 小时挑战",
        "starts_at": "2026-03-01T00:00:00Z",
        "ends_at": "2026-04-01T00:00:00Z",
        "status": "active"
      }
    ]
  }
}
```

#### 17.8.2 提交参赛作品 `POST /api/v1/community/challenges/submit`

**请求体**：`{"challenge_id": "uuid", "post_id": "uuid"}`

---

### 17.9 常见错误 HTTP 状态

与业务错误对应关系（`msg` 多为英文短句）：`400` 参数非法、`403` 无权限、`404` 资源不存在；其它服务端错误为 `500`。
