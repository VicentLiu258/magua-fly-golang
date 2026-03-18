## Magua Fly 训练全流程接口执行指引（E2E）

本文档给出一套从**教员创建任务 → 学员查看并执行 → 录入训练记录 → 生成档案 → 布置下次任务 → 学员再次查看任务**的完整接口执行流程，方便在 Apifox / cURL 等工具中顺序执行，做端到端联调与回归测试。

文中所有接口均基于当前项目的实际 API 设计（见 `doc/api_overview.md`），采用统一返回结构：

```json
{
  "code": 200,
  "msg": "ok",
  "data": { }
}
```

- 除 `/health`、`/auth/*` 外，其他接口均需要 `Authorization: Bearer <JWT>`。
- 所有业务接口使用 `POST` 方法（包括「列表」「详情」类）。

---

## 1. 前置准备

### 1.1 环境与基础约定

- **本地环境**：`http://localhost:8080`
- **远程测试环境**：`https://your-dev-domain`
- 以下用 `{{BASE_URL}}` 表示当前环境地址。
- 请求通用头：
  - `Content-Type: application/json`
  - `Authorization: Bearer {{TOKEN}}`（仅登录接口不需要）

### 1.2 测试数据与变量

建议在 Apifox 中配置以下环境变量，用于串联整个流程：

- `BASE_URL`
- `INSTRUCTOR_TOKEN`（教员 JWT）
- `STUDENT_TOKEN`（学员 JWT）
- `INSTRUCTOR_USER_ID`（登录返回的 `user.id`）
- `STUDENT_USER_ID`
- `INSTRUCTOR_ID`（`instructors.id`）
- `STUDENT_ID`（`students.id`）
- `AIRCRAFT_ID`
- `BOOKING_ID`
- `TRAINING_RECORD_ID`
- `ARCHIVE_ID`

通过 Tests 脚本把每一步接口返回的关键字段写入环境变量，即可实现「一键从上到下连跑」。

---

## 2. 登录与基础信息确认

### 2.1 教员登录获取 `INSTRUCTOR_TOKEN`

- **接口**：`POST /auth/login`
- **Body 示例**：

```json
{
  "username": "instructor1",
  "password": "secret"
}
```

- **检查点**：
  - `code = 200`
  - 记录：
    - `data.token` → `INSTRUCTOR_TOKEN`
    - `data.user.id` → `INSTRUCTOR_USER_ID`
    - `data.user.role = "instructor"`

### 2.2 学员登录获取 `STUDENT_TOKEN`

- **接口**：`POST /auth/login`
- **Body 示例**：

```json
{
  "username": "student1",
  "password": "secret"
}
```

- **检查点**：
  - `code = 200`
  - 记录：
    - `data.token` → `STUDENT_TOKEN`
    - `data.user.id` → `STUDENT_USER_ID`
    - `data.user.role = "student"`

### 2.3 获取学员实体 ID：`STUDENT_ID`

- **接口**：`POST /api/v1/students/list?limit=20&offset=0`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **Body**：空 `{}` 或不传
- **检查点**：
  - 在返回的 `data.items` 中找到目标学员（可按照 `student_no` 或 `name` 筛选）。
  - 记录该条记录的：
    - `id` → `STUDENT_ID`

### 2.4 获取教员实体 ID：`INSTRUCTOR_ID`

- **接口**：`POST /api/v1/instructors/list`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **Body**：`{}`
- **检查点**：
  - 从列表中找到与当前登录教员关联的记录（`user_id = INSTRUCTOR_USER_ID`）。
  - 记录：
    - `id` → `INSTRUCTOR_ID`

### 2.5 获取一架可用飞机：`AIRCRAFT_ID`

- **接口**：`POST /api/v1/aircraft/list`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **Body**：`{}`
- **检查点**：
  - 找到 `status = "active"` 的飞机。
  - 记录：
    - `id` → `AIRCRAFT_ID`

---

## 3. 教员为学员创建一次训练计划（预约）

在当前系统中，「教员创建飞行任务」对应为学员创建一条预约（`bookings`），并经审核生成 `training_plans`。

### 3.1（可选）查看指定日期的教员排期网格

- **接口**：`POST /api/v1/schedules/instructor-grid`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **Body 示例**：

```json
{
  "date": "2026-03-20"
}
```

- **检查点**：
  - `code = 200`
  - `data.date = "2026-03-20"`
  - 在 `data.instructors` 中找到 `id = INSTRUCTOR_ID` 的记录，确认所选时间段的 `status = "available"`。

### 3.2 创建预约：`BOOKING_ID`

- **接口**：`POST /api/v1/bookings`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **Body 示例**：

```json
{
  "student_id": "{{STUDENT_ID}}",
  "instructor_id": "{{INSTRUCTOR_ID}}",
  "aircraft_id": "{{AIRCRAFT_ID}}",
  "plan_date": "2026-03-20",
  "start_time": "10:00",
  "end_time": "11:30"
}
```

- **检查点**：
  - `code = 200`
  - `data.status = "pending"`
  - 记录：
    - `data.id` → `BOOKING_ID`

### 3.3 审核预约，生成训练计划

- **接口**：`POST /api/v1/bookings/review`
- **Header**：使用具有审核权限的账号，一般可用 `INSTRUCTOR_TOKEN` 或 `admin`。
- **Body 示例**：

```json
{
  "id": {{BOOKING_ID}},
  "approve": true,
  "reason": "E2E 测试通过"
}
```

- **检查点**：
  - `code = 200`
  - `data.status = "approved"`

---

## 4. 学员查看任务与训练安排

### 4.1 学员查看「我的任务」列表

- **接口**：`POST /api/v1/tasks/my-list`
- **Header**：`Authorization: Bearer {{STUDENT_TOKEN}}`
- **Body 示例**：

```json
{
  "status": "pending",
  "page": 1,
  "page_size": 50
}
```

- **检查点**：
  - `code = 200`
  - `data.items` 非空。
  - 若飞行任务已经与任务系统打通，可在此处看到与本次预约相关的任务并记录 `task_id`。

### 4.2 学员主页聚合视图（可选）

- **接口**：`POST /api/v1/students/my-home`
- **Header**：`Authorization: Bearer {{STUDENT_TOKEN}}`
- **Body**：`{}`
- **检查点**：
  - `data.profile.student_id = STUDENT_ID`
  - `data.upcoming_training` 中存在一条 `training_date = 2026-03-20` 且教员/飞机匹配的记录。

---

## 5. 执行飞行任务并录入训练记录

结束一次实际飞行后，通过训练记录模块录入本课数据，这是后续档案与统计的基础。

### 5.1 创建训练记录：`TRAINING_RECORD_ID`

- **接口**：`POST /api/v1/training-records`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **Body 示例**：

```json
{
  "student_id": "{{STUDENT_ID}}",
  "instructor_id": "{{INSTRUCTOR_ID}}",
  "aircraft_id": "{{AIRCRAFT_ID}}",
  "training_date": "2026-03-20",
  "engine_on_time": "2026-03-20T10:05:00Z",
  "engine_off_time": "2026-03-20T11:25:00Z",
  "flight_duration_minutes": 80,
  "is_local": true,
  "planned_duration_minutes": 90,
  "theory_duration_minutes": 0,
  "sim_duration_minutes": 0,
  "landings": 5,
  "syllabus_outline": "1#",
  "archive_display_no": "1#999",
  "status": "normal"
}
```

- **检查点**：
  - `code = 200`
  - 记录：
    - `data.id` → `TRAINING_RECORD_ID`

### 5.2（可选）在任务系统中标记任务完成

如果飞行任务已经映射为学员任务，可通过任务接口更新状态：

- **接口**：`POST /api/v1/tasks/update-status`
- **Header**：`Authorization: Bearer {{STUDENT_TOKEN}}`
- **Body 示例**：

```json
{
  "task_id": "任务对应的 ID",
  "status": "done"
}
```

- **检查点**：
  - `code = 200`
  - `data.status = "ok"`

---

## 6. 基于训练记录生成训练档案

### 6.1 创建档案：`ARCHIVE_ID`

- **接口**：`POST /api/v1/training-archives`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **Body 示例**：

```json
{
  "training_record_id": "{{TRAINING_RECORD_ID}}",
  "student_id": "{{STUDENT_ID}}",
  "instructor_id": "{{INSTRUCTOR_ID}}",
  "training_date": "2026-03-20",
  "stage": "初级训练",
  "planned_duration_minutes": 90,
  "total_experience_minutes": 2635,
  "remaining_minutes": 1804,
  "engine_on_time": "2026-03-20T10:05:00Z",
  "engine_off_time": "2026-03-20T11:25:00Z",
  "flight_duration_minutes": 80,
  "progress_note": "",
  "instructor_review": "",
  "student_self_review": "",
  "task_assignment": ""
}
```

- **检查点**：
  - `code = 200`
  - 记录：
    - `data.id` → `ARCHIVE_ID`

### 6.2 教员完善评语与总结

- **接口**：`POST /api/v1/training-archives/update`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **Body 示例**：

```json
{
  "id": "{{ARCHIVE_ID}}",
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

- **检查点**：
  - `code = 200`
  - 返回的 `data` 中相关字段与提交内容一致。

### 6.3 学员签字（可选）

- **接口**：`POST /api/v1/training-archives/sign`
- **Header**：`Authorization: Bearer {{STUDENT_TOKEN}}`
- **Body 示例**：

```json
{
  "id": "{{ARCHIVE_ID}}",
  "role": "student"
}
```

- **检查点**：
  - 学员角色可以正常签字。

---

## 7. 教员基于档案布置「下次训练任务」

### 7.1 查询大纲科目 + 当前学员状态

- **接口**：`POST /api/v1/syllabus-subjects/list`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **Body 示例**：

```json
{
  "student_id": "{{STUDENT_ID}}"
}
```

- **检查点**：
  - `code = 200`
  - `data.items` 中的每个科目包含：
    - `is_graduated`
    - `status`（`completed` / `in_progress` / `not_started`）
  - 选出 1～2 个拟作为下次训练重点的 `syllabus_subject_id`，组成数组 `SYLLABUS_SUBJECT_IDS`。

### 7.2 设置本课科目（可选）

- **接口**：`POST /api/v1/training-archives/subjects/update`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **Body 示例**：

```json
{
  "training_archive_id": "{{ARCHIVE_ID}}",
  "subjects": [
    {
      "syllabus_subject_id": "某科目ID",
      "evaluation_score": 9,
      "is_graduated": false,
      "subject_content": "本场起落 5 次",
      "type": "实践",
      "sort_order": 0
    }
  ]
}
```

### 7.3 布置下次训练科目（真正意义上的「下次任务」）

- **接口**：`POST /api/v1/training-archives/next-subjects/update`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **Body 示例**：

```json
{
  "training_archive_id": "{{ARCHIVE_ID}}",
  "syllabus_subject_ids": [
    "科目ID1",
    "科目ID2"
  ]
}
```

- **检查点**：
  - `code = 200`
  - 通过下一步列表接口确认已保存。

### 7.4 查看已布置的下次科目

- **接口**：`POST /api/v1/training-archives/next-subjects/list`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **Body 示例**：

```json
{
  "training_archive_id": "{{ARCHIVE_ID}}"
}
```

- **检查点**：
  - `data.items` 中包含刚刚布置的科目。

### 7.5 责任教员签字前的业务校验

> 规则：责任教员 / 主任教员签字前，必须先布置至少一个「下次训练科目」。

1. **错误路径（可选测试）：未布置下次科目就签字**

   - 先在一个未布置下次科目的档案上调用：

   ```json
   {
     "id": "某未布置科目的 ARCHIVE_ID",
     "role": "responsible_instructor"
   }
   ```

   - 期望返回：

   ```json
   {
     "code": 400,
     "msg": "请先布置下次训练科目",
     "data": null
   }
   ```

2. **正确路径：已布置下次科目后签字**

- **接口**：`POST /api/v1/training-archives/sign`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **Body 示例**：

```json
{
  "id": "{{ARCHIVE_ID}}",
  "role": "responsible_instructor"
}
```

- **检查点**：
  - `code = 200`
  - 档案记录中责任教员签字时间已更新。

---

## 8. 学员再次查看任务与主页（验证「下次任务」可见）

### 8.1 学员查看「我的任务」中的新任务

- **接口**：`POST /api/v1/tasks/my-list`
- **Header**：`Authorization: Bearer {{STUDENT_TOKEN}}`
- **Body**：

```json
{
  "status": "pending",
  "page": 1,
  "page_size": 50
}
```

- **检查点**：
  - `data.items` 中包含与刚布置的下次科目 / 下次训练相关的任务条目（具体字段与前端展示约定有关）。

### 8.2 学员主页聚合视图确认进度

- **接口**：`POST /api/v1/students/my-home`
- **Header**：`Authorization: Bearer {{STUDENT_TOKEN}}`
- **Body**：`{}`
- **检查点**：
  - `data.last_training_archive.id = ARCHIVE_ID`
  - `data.tasks.total` 中包含新布置任务的数量变化。
  - `data.subject_progress.completed` / `total` 随训练进展发生合理变化。

---

## 9. 建议的回归执行顺序（可在 Apifox 中从上到下依次点击）

1. 认证：
   - `POST /auth/login`（教员）
   - `POST /auth/login`（学员）
2. 基础信息：
   - `POST /api/v1/students/list`
   - `POST /api/v1/instructors/list`
   - `POST /api/v1/aircraft/list`
3. 排期与预约：
   - `POST /api/v1/schedules/instructor-grid`（可选）
   - `POST /api/v1/bookings`
   - `POST /api/v1/bookings/review`
4. 学员视角：
   - `POST /api/v1/tasks/my-list`
   - `POST /api/v1/students/my-home`
5. 训练记录与档案：
   - `POST /api/v1/training-records`
   - `POST /api/v1/training-archives`
   - `POST /api/v1/training-archives/update`
   - `POST /api/v1/training-archives/sign`（学员，可选）
6. 科目与下次任务：
   - `POST /api/v1/syllabus-subjects/list`
   - `POST /api/v1/training-archives/subjects/update`（可选）
   - `POST /api/v1/training-archives/next-subjects/update`
   - `POST /api/v1/training-archives/next-subjects/list`
   - `POST /api/v1/training-archives/sign`（责任教员）
7. 学员第二次查看：
   - `POST /api/v1/tasks/my-list`
   - `POST /api/v1/students/my-home`

按以上顺序，结合环境变量和 Apifox 的测试脚本，即可完成一次从「教员创建任务」到「学员查看下次任务」的完整 E2E 流程验证。

