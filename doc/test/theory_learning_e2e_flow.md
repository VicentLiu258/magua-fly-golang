### 理论知识学习模块 E2E 功能测试流程（接口串联版）

> 目标：**覆盖理论模块从“教员配置内容/试题/试卷”到“学员学习、做题、出成绩、查历史”的完整闭环**，仅聚焦理论模块（不串联飞行训练档案）。  
> 建议在 Apifox 中按下面顺序组织接口，用环境变量串联 `DOCUMENT_ID`、`Q1_ID`、`Q2_ID`、`PAPER_ID`、`ATTEMPT_ID` 等。

---

### 1. 前置准备与账号

#### 1.1 教员账号登录获取 TOKEN

- **接口**：`POST /auth/login`
- **请求示例**：

```json
{
  "username": "instructor1",
  "password": "secret"
}
```

- **检查点与变量**：
  - `code = 200`
  - `INSTRUCTOR_TOKEN = data.token`
  - `INSTRUCTOR_USER_ID = data.user.id`
- **用途**：后续作为 `Authorization: Bearer {{INSTRUCTOR_TOKEN}}` 调用教员侧接口。

#### 1.2 学员账号登录获取 TOKEN

- **接口**：`POST /auth/login`
- **请求示例**：

```json
{
  "username": "student1",
  "password": "secret"
}
```

- **检查点与变量**：
  - `code = 200`
  - `STUDENT_TOKEN = data.token`
  - `STUDENT_USER_ID = data.user.id`
- **用途**：后续学员端查看文档、开始答题、交卷、查历史成绩等。

> 如无账号，可先调用 `POST /auth/register` 分别创建 `role = "instructor"` / `role = "student"` 再登录。

---

### 2. 教员端：创建理论文档（学习内容主体）

#### 2.1 创建理论文档

- **接口**：`POST /api/v1/theory/documents/create`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`
- **请求体示例**：

```json
{
  "title": "失速与尾旋理论（E2E 测试）",
  "slug": "stall-spin-theory-e2e",
  "sort_order": 10,
  "category": "空气动力学",
  "video_url": "https://video.example.com/stall-e2e.mp4"
}
```

- **检查点**：
  - `code = 200`
  - 返回 `data.title`、`data.category` 与入参一致。
- **环境变量**：
  - `DOCUMENT_ID = data.id`

#### 2.2（可选）更新理论文档元数据

- **接口**：`POST /api/v1/theory/documents/update`
- **请求示例**（根据需要微调）：

```json
{
  "id": "{{DOCUMENT_ID}}",
  "title": "失速与尾旋理论（E2E 测试 v2）",
  "slug": "stall-spin-theory-e2e",
  "sort_order": 20,
  "category": "空气动力学",
  "video_url": "https://video.example.com/stall-e2e-v2.mp4"
}
```

- **检查点**：`data.title`、`sort_order` 等字段按更新后生效。

#### 2.3 校验在列表与详情中的可见性

- **文档列表**：`POST /api/v1/theory/documents/list`

```json
{
  "keyword": "失速与尾旋理论",
  "category": "空气动力学",
  "status": "published",
  "page": 1,
  "page_size": 20
}
```

- **检查点**：
  - `code = 200`
  - `data.items` 中可找到 `id = {{DOCUMENT_ID}}` 的记录。
- **文档详情**：`POST /api/v1/theory/documents/detail`

```json
{
  "document_id": "{{DOCUMENT_ID}}"
}
```

- **检查点**：
  - `data.document.id = DOCUMENT_ID`
  - `data.content` 结构合理（有 `text/image/video_embed`，或为空数组）。

---

### 3. 教员端：录入题库题目

#### 3.1 为该文档创建若干题目（单选 + 多选）

- **接口**：`POST /api/v1/theory/questions/admin/create`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`

**示例 1：单选题**

```json
{
  "document_id": "{{DOCUMENT_ID}}",
  "question": {
    "id": "",
    "type": "single",
    "stem": "下面哪一项与失速最直接相关？",
    "options": [
      { "id": "A", "text": "飞机重量" },
      { "id": "B", "text": "迎角" },
      { "id": "C", "text": "油量" }
    ],
    "correct_option_ids": ["B"],
    "explanation": "失速本质是迎角过大导致升力丧失。",
    "tags": ["基础"],
    "score": 2
  }
}
```

**示例 2：多选题**

```json
{
  "document_id": "{{DOCUMENT_ID}}",
  "question": {
    "id": "",
    "type": "multi",
    "stem": "下列哪些因素会提高失速速度？",
    "options": [
      { "id": "A", "text": "载重增加" },
      { "id": "B", "text": "结冰" },
      { "id": "C", "text": "起落架放下" },
      { "id": "D", "text": "迎角减小" }
    ],
    "correct_option_ids": ["A", "B"],
    "explanation": "载重增加、结冰等都会提高失速速度。",
    "tags": ["进阶"],
    "score": 3
  }
}
```

- **检查点**：
  - `code = 200`
  - `data.status = "ok"`

> 可在 Apifox 的 Tests 中把每次返回的题目 `id` 收集到数组变量 `QUESTION_IDS` 里。

#### 3.2 查看题库列表，确认创建成功

- **接口**：`POST /api/v1/theory/questions/admin/list`

```json
{
  "document_id": "{{DOCUMENT_ID}}",
  "page": 1,
  "page_size": 50,
  "keyword": "",
  "tags": [],
  "type": ""
}
```

- **检查点**：
  - `data.items` 中包含刚创建的题；
  - 每题的 `stem`、`options`、`tags`、`score` 等于配置值。
- **记录典型题目的 ID**：
  - `Q1_ID`：单选题的 `id`
  - `Q2_ID`：多选题的 `id`

#### 3.3（可选）更新 / 删除题目

- **更新题目**：`POST /api/v1/theory/questions/admin/update`

```json
{
  "document_id": "{{DOCUMENT_ID}}",
  "question": {
    "id": "{{Q1_ID}}",
    "type": "single",
    "stem": "（修改后）下面哪一项与失速最直接相关？",
    "options": [
      { "id": "A", "text": "飞机重量" },
      { "id": "B", "text": "迎角（正确）" }
    ],
    "correct_option_ids": ["B"],
    "explanation": "修改后的解析……",
    "tags": ["基础"],
    "score": 2
  }
}
```

- **删除题目**：`POST /api/v1/theory/questions/admin/delete`

```json
{
  "document_id": "{{DOCUMENT_ID}}",
  "question_id": "{{Q2_ID}}"
}
```

- 再次调用 `admin/list` 验证变更生效。

---

### 4. 教员端：配置理论试卷

#### 4.1 创建一套试卷并绑定题目

- **接口**：`POST /api/v1/theory/papers/save`
- **Header**：`Authorization: Bearer {{INSTRUCTOR_TOKEN}}`

```json
{
  "id": null,
  "theory_document_id": "{{DOCUMENT_ID}}",
  "name": "失速与尾旋阶段测验（E2E）",
  "description": "E2E 全量测试试卷",
  "enabled": true,
  "questions": [
    { "question_id": "{{Q1_ID}}", "score": 2, "sort_order": 1 },
    { "question_id": "{{Q2_ID}}", "score": 3, "sort_order": 2 }
  ]
}
```

- **环境变量**：
  - `PAPER_ID = data.id`
  - `TOTAL_SCORE = data.total_score`
- **检查点**：
  - `total_score = 2 + 3 = 5`

#### 4.2 列表确认试卷与分值

- **接口**：`POST /api/v1/theory/papers/list`

```json
{
  "document_id": "{{DOCUMENT_ID}}",
  "page": 1,
  "page_size": 20
}
```

- **检查点**：
  - 能找到 `id = PAPER_ID` 的记录；
  - `enabled = true`；
  - `total_score = 5`。

#### 4.3（可选）更新 / 删除试卷

- **更新试卷**：再次调用 `papers/save`，传 `id = PAPER_ID`，调整名称或题目顺序，看 `total_score` 和排序是否按预期变化。
- **删除试卷**：`POST /api/v1/theory/papers/delete`，然后用 `papers/list` 确认不再返回该卷。

---

### 5. 学员端：查看理论学习入口

#### 5.1 学员查看理论文档列表

- **接口**：`POST /api/v1/theory/documents/list`
- **Header**：`Authorization: Bearer {{STUDENT_TOKEN}}`

```json
{
  "keyword": "失速与尾旋理论",
  "category": "空气动力学",
  "status": "published",
  "page": 1,
  "page_size": 20
}
```

- **检查点**：
  - 能在 `data.items` 中看到 `id = DOCUMENT_ID`；
  - 理想情况：`has_questions = true`（当前实现中可能尚未填充，可作为一个检查点记下）。

#### 5.2（可选）学员“我的主页”中理论概要

- **接口**：`POST /api/v1/students/my-home`
- **Header**：`Authorization: Bearer {{STUDENT_TOKEN}}`
- **请求体**：

```json
{}
```

- **检查点**：
  - 记录 `data.theory_summary` 的初始值（`completed_documents`、`avg_score`、`last_attempt_at`），供答题后对比。

---

### 6. 学员端：开始一次答题（使用试卷）

#### 6.1 使用试卷出题

- **接口**：`POST /api/v1/theory/questions/start`
- **Header**：`Authorization: Bearer {{STUDENT_TOKEN}}`

```json
{
  "document_id": "{{DOCUMENT_ID}}",
  "paper_id": "{{PAPER_ID}}"
}
```

- **返回数据**：
  - `ATTEMPT_ID = data.attempt_id`
  - `questions = data.questions`（两道题）
- **检查点**：
  - `questions.length` 与试卷中配置的题目数量一致；
  - 每题 `id`、`stem`、`options` 与后台题库对应；
  - 每题 `score` 与试卷中配置分值一致。

> 可在 Tests 中把 `data.attempt_id` 写入环境变量 `ATTEMPT_ID`，把 `data.questions[*].id` 重新映射回 `Q1_ID`、`Q2_ID`（或校验它们等于之前后台返回的 ID）。

---

### 7. 学员端：交卷并评分

> 实测后端对 `DocumentID` 做了必填校验，**建议在交卷请求中同时传入 `document_id`**，以避免校验错误。

#### 7.1 构造提交答案并交卷

- **接口**：`POST /api/v1/theory/questions/submit`
- **Header**：`Authorization: Bearer {{STUDENT_TOKEN}}`

**推荐请求体（带 document_id）：**

```json
{
  "document_id": "{{DOCUMENT_ID}}",
  "attempt_id": "{{ATTEMPT_ID}}",
  "answers": [
    {
      "question_id": "{{Q1_ID}}",
      "selected_option_ids": ["B"]
    },
    {
      "question_id": "{{Q2_ID}}",
      "selected_option_ids": ["A", "C"]
    }
  ]
}
```

- **测试场景**：
  - 场景 A：全部按正确答案作答 → 期望 `score = max_score`、`correct_count = total_count`；
  - 场景 B：故意一题答错 → 检查得分、对题数是否按题目分值计算。
- **检查点**：
  - `data.score`、`data.max_score`；
  - `data.correct_count`、`data.total_count`；
  - `data.passed` 布尔值是否按设定及格线返回。

#### 7.2 异常场景（可选）

- 使用错误的 `attempt_id`：
  - 期望返回 400/404 异常。
- 重复提交同一个 `attempt_id`：
  - 期望后端防重复或返回相应错误码。

---

### 8. 学员端：查看历史成绩与理论概要

#### 8.1 某文档下的历史答题记录

- **接口**：`POST /api/v1/theory/attempts/my-list`
- **Header**：`Authorization: Bearer {{STUDENT_TOKEN}}`

```json
{
  "document_id": "{{DOCUMENT_ID}}",
  "page": 1,
  "page_size": 20
}
```

- **检查点**：
  - `data.items` 中包含刚才提交的 `id = ATTEMPT_ID` 对应记录；
  - 该记录的 `score`、`duration_seconds`、`submitted_at` 与作答情况吻合；
  - 多次重复练习后，`total` 对应增加。

#### 8.2 学员主页中的理论统计更新

- **接口**：`POST /api/v1/students/my-home`（同 5.2）
- **检查点**：
  - 与答题前的值对比：
    - `theory_summary.completed_documents` 是否增加；
    - `avg_score` 是否更新为新平均分；
    - `last_attempt_at` 是否更新为本次作答时间。
  - 如任务系统已对接理论任务，可同时查看：
    - `data.tasks.items` 中是否还有该理论任务；
    - 若交卷后调用了 `/api/v1/tasks/update-status`，任务状态应变为 `done`。

---

### 9. 推荐在 Apifox 中的执行顺序（从上到下）

1. `POST /auth/login`（教员） → 设置 `INSTRUCTOR_TOKEN`
2. `POST /auth/login`（学员） → 设置 `STUDENT_TOKEN`
3. `POST /api/v1/theory/documents/create` → 记录 `DOCUMENT_ID`
4. `POST /api/v1/theory/documents/list` / `detail` → 校验文档
5. 多次 `POST /api/v1/theory/questions/admin/create` → 录入题库
6. `POST /api/v1/theory/questions/admin/list` → 拿到各 `QUESTION_ID`
7. `POST /api/v1/theory/papers/save` → 创建试卷，记录 `PAPER_ID`
8. `POST /api/v1/theory/papers/list` → 校验试卷信息
9. `POST /api/v1/theory/documents/list`（学员 TOKEN）→ 学员可见文档
10. `POST /api/v1/theory/questions/start`（带 `paper_id`）→ 获取 `ATTEMPT_ID` 与题目
11. `POST /api/v1/theory/questions/submit` → 交卷评分
12. `POST /api/v1/theory/attempts/my-list` → 查看历史成绩
13. `POST /api/v1/students/my-home` → 查看理论学习概要是否更新

---

### 10. 变量与串联建议

- **环境变量推荐**：
  - 账号类：`INSTRUCTOR_TOKEN`、`INSTRUCTOR_USER_ID`、`STUDENT_TOKEN`、`STUDENT_USER_ID`
  - 理论模块：`DOCUMENT_ID`、`Q1_ID`、`Q2_ID`、`QUESTION_IDS[]`、`PAPER_ID`、`ATTEMPT_ID`
- **Tests 中写变量示例（Apifox 伪代码）**：
  - 在创建文档接口 Tests 中：

    ```js
    const data = JSON.parse(response.body).data;
    apifox.setVariable("DOCUMENT_ID", data.id);
    ```

  - 在 `questions/start` 接口 Tests 中：

    ```js
    const data = JSON.parse(response.body).data;
    apifox.setVariable("ATTEMPT_ID", data.attempt_id);
    ```

这样配置后，可以在 Apifox 中从上到下依次点击，或直接运行集合，实现一次完整的**理论知识学习 E2E 功能测试（含录入试题、配置试卷、学员学习与考试、成绩与统计校验）**。

