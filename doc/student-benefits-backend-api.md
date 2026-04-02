# 学员权益概览 — 后台接口设计（Golang 实现）

本文档对齐前端 `legacy_pages/admin/StudentManagement.tsx` 中 **「权益信息」** Tab（学员权益概览、三类权益卡片、预约限制、其他权益、权益额度调整与变更历史），并与现有 `doc/api_overview.md` 中的 **`contract_info`**、学员端 **`entitlements`**、**`POST /api/v1/students/training-summary/detail`** 等约定衔接，供在 Golang 服务中实现。

---

## 1. 范围与原则

| 项目 | 说明 |
|------|------|
| 权限 | 与 `training-summary/detail` 一致：`admin` / `staff` / `maintenance` / `instructor` 可查；**调整类写操作**建议仅 `admin`，若业务允许职员调额度可再开放 `staff`（需审计）。 |
| 传输风格 | 与现有工程一致：**`POST` + JSON body**，统一 `ApiEnvelope`（`code` / `msg` / `data`）。 |
| 飞行时间 | **既有台账**：`students.contract_duration_minutes`、`flown_minutes`、剩余分钟等已在列表与训练汇总中体现；权益 Tab 的「飞行时间权益」应对齐同一套数字，**手工增减**必须落库流水，避免与前端本地 `setState` 不一致。 |
| 亲友体验 / 积分 | 与 `contract_info.guest_experience`、`contract_info.learning_points` 对齐；**used / remaining** 的推导规则见 `api_overview.md`（`remaining = max(0, quota - used)`）。 |
| 预约限制 / 视频 / 拍摄 | 当前界面为展示项，建议落在 **`contract_info` 扩展字段**或独立 JSON 列，由本聚合接口一并返回。 |

---

## 2. 与现有文档的映射

| 界面区块 | 建议数据源 |
|----------|------------|
| 飞行时间：总额 / 已用 / 剩余 | `contract_duration_minutes`、`flown_minutes`、派生 `remaining_minutes`（与 §4.1.2 `duration` 一致）；若有「赠送课时」等，可用 `contract_info.flight_time_adjust_minutes` 或流水汇总修正展示。 |
| 亲友体验 | `contract_info.guest_experience`：`quota`、`used`（`remaining` 计算） |
| 学员积分 | `contract_info.learning_points` 需区分 **总额度 / 已消耗 / 当前余额**；若历史仅存单一数值，迁移时定义为「当前余额」，总额度用流水倒推或新增字段。 |
| 最后同步时间 | 服务端聚合完成时刻 **`computed_at`**（各源表 `updated_at` 的最大值或快照时间） |
| 变更历史 | **独立流水表**（见 §3），禁止仅依赖前端内存 |

---

## 3. 推荐数据模型

### 3.1 `students.contract_info`（JSONB）扩展形状（示例）

在现有 `learning_points`、`guest_experience` 基础上扩展，便于少建表、快速上线；若审计要求高，可将下列标量迁到关系表。

```json
{
  "learning_points": {
    "total": 500,
    "used": 120,
    "remaining": 380
  },
  "guest_experience": {
    "quota": 10,
    "used": 3,
    "remaining": 7
  },
  "booking_limits": {
    "two_week_rule": "无",
    "three_month_rule": "无",
    "weekday_only": "无",
    "morning_only": "无"
  },
  "extras": {
    "training_video_hours": 0,
    "custom_shoot_count": 0
  }
}
```

> **兼容**：旧结构 `learning_points: 120`、`guest_experience: { "quota", "used" }` 在读时归一化为上表；写回时由服务层统一格式。

### 3.2 `student_benefit_ledger`（流水表，推荐）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | PK |
| student_id | UUID | FK → students |
| benefit_kind | VARCHAR | `flight_time` \| `guest_experience` \| `learning_points` |
| direction | VARCHAR | `credit`（增加额度/回补） \| `debit`（扣减/消耗） |
| amount | DECIMAL(12,2) | 数量；飞行时间可存**小时**或**分钟**（全库统一一种，建议分钟） |
| reason | TEXT | 原因 |
| operator_user_id | UUID | 可空（系统任务） |
| operator_name | VARCHAR | 冗余展示名 |
| created_at | TIMESTAMPTZ | |
| idempotency_key | VARCHAR(64) | 可选，防重复提交 |

**约束**：`amount > 0`；扣减时业务层校验剩余充足；与训练记录自动扣时可在 `reason` / 扩展 JSON 中标明来源 `source: training_record`。

---

## 4. 接口列表

| 接口 | 说明 |
|------|------|
| `POST /api/v1/admin/students/benefits/overview` | 权益概览一次性聚合（Tab 主数据） |
| `POST /api/v1/admin/students/benefits/adjust` | 手工增减某一类权益（对应「确认增加/减少」） |
| `POST /api/v1/admin/students/benefits/logs` | 分页查询某一类权益的变更历史（弹窗表格） |
| `POST /api/v1/admin/students/benefits/extras/save` | 可选：单独保存预约限制与其它权益（若不想用 `save` 大表单） |

与 `POST /api/v1/admin/students/save` 的关系：**档案保存**仍可整体提交 `contract_info`；本组接口用于权益 Tab **专用高频操作**与**流水只读列表**，减少全量 `save` 冲突。

---

## 5. 接口说明

### 5.1 `POST /api/v1/admin/students/benefits/overview`

**权限**：`admin` | `staff` | `maintenance` | `instructor`（与学员详情策略一致）。

**请求体**：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**成功 `data` 示例**（字段名建议 snake_case，与现有 API 一致）：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "computed_at": "2026-04-02T09:00:00+08:00",
  "flight_time": {
    "unit": "hour",
    "total": 60.0,
    "used": 20.5,
    "remaining": 39.5
  },
  "guest_experience": {
    "unit": "times",
    "total": 10,
    "used": 3,
    "remaining": 7
  },
  "learning_points": {
    "unit": "points",
    "total": 500,
    "used": 120,
    "remaining": 380
  },
  "booking_limits": {
    "two_week_rule": "无",
    "three_month_rule": "无",
    "weekday_only": "无",
    "morning_only": "无"
  },
  "extras": {
    "training_video_hours": 0,
    "custom_shoot_count": 0
  },
  "has_active_contract": true
}
```

**业务规则**：

- `flight_time` 与 §4.1.2、`/admin/students/list` 中分钟字段一致时，可在此处以 **小时** 返回（保留 1 位小数）以减少前端换算。
- `has_active_contract`：`contracts.status` 存在非 `cancelled` 合同时为 `true`；前端可用其禁用「管理权益」按钮（与现 UI「请先添加合同」一致）。

**错误**：`404` 学员不存在；`403` 权限不足。

---

### 5.2 `POST /api/v1/admin/students/benefits/adjust`

**权限**：建议 **仅 `admin`**（或 `admin` + `staff`）。

**请求体**：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "benefit_kind": "learning_points",
  "direction": "credit",
  "amount": 50,
  "reason": "活动赠送",
  "idempotency_key": "optional-uuid-from-client"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| benefit_kind | string | `flight_time` \| `guest_experience` \| `learning_points` |
| direction | string | `credit` 增加可用量；`debit` 扣减（对应界面「增加/减少」） |
| amount | number | 正数；飞行时间若用分钟，与 `overview` 的 `unit` 约定一致并在文档写死 |

**成功 `data`**：返回 **更新后的该类权益快照**（结构与 `overview` 中对应块相同），便于前端局部刷新。

**事务**：

1. 插入 `student_benefit_ledger`；
2. 更新 `students` 台账或 `contract_info` 中对应字段；
3. 飞行时间若只改「合同外赠送分钟」，勿覆盖 `contract_duration_minutes` 合同本意，建议单独字段或流水汇总。

**错误**：`400` 参数非法、`debit` 超额；`409` 幂等键冲突；`404` 学员不存在。

---

### 5.3 `POST /api/v1/admin/students/benefits/logs`

**权限**：与 `overview` 相同。

**请求体**：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "benefit_kind": "guest_experience",
  "page": 1,
  "page_size": 20
}
```

**成功 `data`**：

```json
{
  "items": [
    {
      "id": "…",
      "date": "2026-04-01",
      "direction": "credit",
      "amount": 2,
      "reason": "补发",
      "operator": "管理员A",
      "created_at": "2026-04-01T15:00:00+08:00"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

**说明**：`overview` 可内嵌 **最近 N 条** 日志避免二次请求；弹窗「变更历史」用本接口分页拉全量。

---

### 5.4 `POST /api/v1/admin/students/benefits/extras/save`（可选）

**请求体**：

```json
{
  "student_id": "550e8400-e29b-41d4-a716-446655440001",
  "booking_limits": {
    "two_week_rule": "最多 2 次",
    "three_month_rule": "无",
    "weekday_only": "是",
    "morning_only": "无"
  },
  "extras": {
    "training_video_hours": 10,
    "custom_shoot_count": 2
  }
}
```

仅更新 `contract_info` 对应子对象，**不动**权益额度与流水。

---

## 6. 与训练记录 / 合同的联动（实现要点）

1. **自动扣减**：亲友体验、积分若在业务上随某次训练/核销动作扣减，应在同一事务写入 `student_benefit_ledger`（`direction=debit`，`reason` 带记录 ID），并更新 `contract_info`。
2. **飞行时间**：正常消耗以 `training_records` 汇总与 `flown_minutes` 为准；**adjust** 仅处理赠送、纠错、人工调账，避免双写冲突。
3. **computed_at**：取 `MAX(students.updated_at, contracts.updated_at, ledger MAX(created_at))` 或查询时刻的 `NOW()` 作为快照时间。

---

## 7. 错误码与 HTTP 状态（建议）

| code | HTTP | 含义 |
|------|------|------|
| 200 | 200 | 成功 |
| 400 | 400 | 参数错误、超额扣减 |
| 403 | 403 | 角色不允许 |
| 404 | 404 | 学员不存在 |
| 409 | 409 | 幂等冲突 |

---

## 8. Golang 实现提示

- **Handler**：与现有 `students` 模块同路由前缀；JWT 中解析 `operator_user_id` 写入流水。
- **ORM**：流水表 + `contract_info` 用 `jsonb` 序列化结构体；读时做版本迁移（旧 JSON → 新形状）。
- **并发**：`adjust` 对同一 `student_id` 可用行锁（`SELECT … FOR UPDATE`）或单字段原子更新 + 流水插入。
- **测试**：覆盖「扣减超额」「幂等重放」「旧 contract_info 兼容解析」。

---

## 9. 前端对接说明（后续改 TS）

- 将 `StudentManagement` 中 `handleAdjustBenefit` 的本地 `setState` 改为调用 `benefits/adjust`，成功后 `overview` 刷新或直接用返回快照。
- 「最后同步时间」绑定 `computed_at`。
- `benefit_kind` 与界面 `flightTime` / `familyExperience` / `points` 做一层 camelCase ↔ snake_case 映射即可。

---

**文档版本**：与仓库 `StudentManagement` 权益 Tab（2026-04）对齐；若 `api_overview.md` 后续增加正式章节，可将本节编号并入该总文档。
