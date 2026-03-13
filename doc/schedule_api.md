# 飞行排期与预约 API 设计说明

## 1. 时间抽象与存储策略

- **时间存储**：数据库保持现有的区间模型：
  - `aircraft_availability(date, start_time, end_time, status)`
  - `instructor_schedules(date, start_time, end_time)`
  - `bookings(plan_date, start_time, end_time, status, ...)`
- **slot 抽象**：前端与接口层使用：
  - 营业时间：每天 09:00–21:00
  - 粒度：30 分钟 1 格
  - `slotIndex = 0..N-1`，仅在应用层/接口中使用，不在数据库中逐 slot 建行。

这样可以避免 “每 slot 一行” 带来的行数爆炸，同时保持与训练记录等其它时间数据结构的一致性。

## 2. Instructor × Time 网格接口

### 2.1 路由

`POST /api/v1/schedules/instructor-grid`

该接口需要登录，位于受 `GinAuth` 保护的路由组内。

### 2.2 请求参数

- **body**
  - `date` (必填): 查询的日期，格式 `YYYY-MM-DD`。

### 2.3 响应结构

统一包装：

```json
{
  "code": 0,
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
          { "slotIndex": 0, "startTime": "09:00", "endTime": "09:30", "status": "available" },
          { "slotIndex": 1, "startTime": "09:30", "endTime": "10:00", "status": "booked", "bookingId": 1001, "aircraftId": 3, "studentId": 10 }
        ]
      }
    ]
  }
}
```

其中：

- `status` 取值：
  - `off`：不在该教员排班内。
  - `available`：在排班内且没有预约。
  - `booked`：该 slot 已被预约。
  - `aircraft_unavailable`：预留给后续飞机不可用区分（当前实现未强制使用）。

### 2.4 服务端计算逻辑（概述）

以单日为例：

1. 使用固定配置 `09:00–21:00`、`30` 分钟生成 slots：
   - `slotIndex 0..23`
   - 对应 `startTime` / `endTime` 字符串。
2. 查询所有教员 `instructors` 作为横轴。
3. 查询给定日期的：
   - `instructor_schedules` → 按 `instructor_id` 分组，映射为可工作区间，并进一步映射到 `slotIndex` 区间标记为 `available`。
   - `bookings` → 按 `instructor_id` 分组，将对应区间映射到 `slotIndex`，覆盖为 `booked` 并附带 `bookingId`/`aircraftId`/`studentId`。
4. 未被上述任何区间覆盖的 slot 默认状态为 `off`。

算法完全基于时间区间，未在数据库中额外创建 slot 维度的行。

## 3. 预约接口（沿用现有 handler）

当前预约接口（均为 POST）：

- `POST /api/v1/bookings`：创建预约（学员）。
- `POST /api/v1/bookings/pending`：后台查看待审核预约。
- `POST /api/v1/bookings/review`：审核预约并在通过时生成 `training_plans`。

所有接口统一返回结构：

```json
{
  "code": 200,
  "msg": "ok",
  "data": { }
}
```

其中：

- `code`：业务状态码，与 HTTP 状态语义保持一致，例如：
  - `200`：成功
  - `400`：参数错误
  - `401`：未认证/未授权
  - `500`：服务器内部错误
- `msg`：人类可读的简要说明。
- `data`：具体业务数据；错误时通常为 `null` 或携带少量上下文信息。

### 3.1 创建预约：`POST /api/v1/bookings`

请求体示例：

```json
{
  "student_id": 123,
  "instructor_id": 1,
  "aircraft_id": 3,
  "plan_date": "2026-03-15",
  "start_time": "10:00",
  "end_time": "11:30"
}
```

成功响应示例（省略部分字段）：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": 1001,
    "student_id": 123,
    "instructor_id": 1,
    "aircraft_id": 3,
    "plan_date": "2026-03-15T00:00:00Z",
    "start_time": "0000-01-01T10:00:00Z",
    "end_time": "0000-01-01T11:30:00Z",
    "status": "pending",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

常见错误响应：

```json
{
  "code": 400,
  "msg": "invalid start_time",
  "data": null
}
```

### 3.2 查询待审核预约：`POST /api/v1/bookings/pending`

请求体：无（空 JSON 或不带 body）。

成功响应示例：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": 1001,
        "student_id": 123,
        "instructor_id": 1,
        "aircraft_id": 3,
        "plan_date": "2026-03-15T00:00:00Z",
        "start_time": "0000-01-01T10:00:00Z",
        "end_time": "0000-01-01T11:30:00Z",
        "status": "pending"
      }
    ]
  }
}
```

失败示例（数据库错误等）：

```json
{
  "code": 500,
  "msg": "database error ...",
  "data": null
}
```

### 3.3 审核预约：`POST /api/v1/bookings/review`

请求体示例：

```json
{
  "id": 1001,
  "approve": true,
  "reason": "可选的审核说明"
}
```

成功响应示例：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "status": "approved"
  }
}
```

若拒绝预约：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "status": "rejected"
  }
}
```

未认证时的响应示例：

```json
{
  "code": 401,
  "msg": "unauthorized",
  "data": null
}
```

这些接口直接使用 `bookings(plan_date, start_time, end_time)` 的区间模型，不依赖 slot 表。

## 4. Go 领域模型与 UseCase

### 4.1 核心领域结构

定义在 `internal/domain/schedule.go` 中：

- `SlotStatus`：`"off" | "available" | "booked" | "aircraft_unavailable"`
- `TimeSlot`：
  - `slotIndex`, `startTime`, `endTime`, `status`
  - 可选：`bookingId`, `aircraftId`, `studentId`
- `InstructorTimeline`：
  - `id`, `name`, `displayColor`, `timeline []TimeSlot`
- `InstructorGridResponse`：
  - `date`, `slotMinutes`, `dayStart`, `dayEnd`, `slots []string`, `instructors []InstructorTimeline`

### 4.2 UseCase

`internal/usecase/schedule.go`:

- `ScheduleUseCase`：
  - 依赖：
    - `InstructorRepository`
    - `InstructorScheduleRepository`
    - `BookingRepository`
  - 方法：
    - `GetInstructorGrid(ctx, date time.Time) (*domain.InstructorGridResponse, error)`

内部使用固定配置（后续可接 DB 配置）将 `TIME` 区间映射为 slotIndex 区间，组装出前端需要的网格数据。

