### MongoDB 配置说明（理论学习与科目学习）

- **连接配置**：
  - 通过 `internal/config` 中的 `MongoURI`、`MongoDBName` 加载。
  - 在 `cmd/server/main.go` 中使用：
    - `mongodb.New(ctx, cfg.MongoURI, cfg.MongoDBName)` 建立连接并 `Ping`。
  - 默认数据库名：`maguafly`（当 `MongoDBName` 为空时使用）。

- **集合一：理论内容 `theory_content`**
  - 对应常量：`internal/mongodb/mongodb.go` 中的 `CollTheoryContent`。
  - 文档结构（`internal/repository/mongo_repo.go`）：
    - `document_id`：对应 `theory_documents.id`（UUID 字符串）。
    - `blocks[]`：内容块数组，每块结构：
      - `type`：`text | image | video_embed` 等。
      - `content`：文本/HTML/Markdown。
      - `url`：图片等资源地址。
      - `caption`：说明文字。
      - `video_id`：视频媒体 ID（与 PG 中 `media_files` 关联）。
      - `position`：在文档中的位置，用于控制插入顺序。
    - `updated_at`：Unix 时间戳（秒）。
  - 访问接口：
    - `internal/repository.TheoryContentRepository`
    - 实现：`internal/mongodb/theory_content_repo.go`

- **集合二：理论题库 `theory_question_sets`**
  - 对应常量：`CollTheoryQuestionSets`。
  - 文档结构（`TheoryQuestionSetDoc`）：
    - `document_id`：`theory_documents.id`。
    - `questions[]`：题目数组，每题结构：
      - `id`：题目本地唯一 ID。
      - `type`：`single | multi`。
      - `stem`：题干。
      - `options[]`：选项数组，包含 `id`（如 `A/B/C`）与 `text`。
      - `correct_option_ids[]`：正确选项 ID 集合，仅在服务端使用，不返回给前端。
      - `explanation`：解析，仅在服务端使用。
      - `tags[]`：标签（难度、知识点等）。
      - `score`：该题分值。
    - `updated_at`：Unix 时间戳。
  - 访问接口：
    - `internal/repository.TheoryQuestionSetRepository`
    - 实现：`internal/mongodb/theory_question_repo.go`

- **集合三：理论作答明细 `theory_attempt_details`**
  - 对应常量：`CollTheoryAttemptDetails`。
  - 文档结构（`TheoryAttemptDetailDoc`）：
    - `theory_attempt_id`：PostgreSQL `theory_attempts.id`（UUID）。
    - `answers[]`：作答记录数组，每条结构：
      - `question_id`：对应题目 ID。
      - `selected_option_ids[]`：学员选择的选项 ID。
      - `is_correct`：是否答对。
      - `score`：该题得分。
    - `auto_graded`：是否由系统自动判分。
    - `created_at`：Unix 时间戳。
  - 访问接口：
    - `internal/repository.TheoryAttemptDetailRepository`
    - 实现：`internal/mongodb/theory_attempt_detail_repo.go`

- **集合四：科目学习内容 `subject_content`（预留）**
  - 对应常量：`CollSubjectContent`。
  - 推荐结构与 `theory_content` 保持一致，只是将 `document_id` 字段替换为：
    - `subject_learning_id`：对应 PG 表 `subject_learning.id`。
  - 后续可按理论模块相同模式实现：
    - `SubjectContentRepository`
    - 以及对应 Mongo repo。

- **索引与性能建议**：
  - 为高频查询字段建立索引（建议在 Mongo shell 或迁移脚本中维护）：
    - `theory_content`：`{ document_id: 1 }`。
    - `theory_question_sets`：`{ document_id: 1 }`。
    - `theory_attempt_details`：`{ theory_attempt_id: 1 }`。
    - `subject_content`：`{ subject_learning_id: 1 }`（预留）。
  - 多租户或多基地扩展时，可在文档内增加 `tenant_id / org_id` 字段，并相应扩展复合索引。

- **与 PostgreSQL 的关系映射**：
  - `theory_documents.id`：由 PostgreSQL 通过 `DEFAULT gen_random_uuid()` 自动生成，作为理论文档的全局唯一标识。
  - `theory_content.document_id` / `theory_question_sets.document_id` / `theory_papers.theory_document_id`：均引用 `theory_documents.id`。
  - `theory_documents.content_ref`：可保存 `document_id` 或 Mongo `_id` 的字符串形式，用于定位 `theory_content` 文档。
  - `theory_attempts.id` ↔ `theory_attempt_details.theory_attempt_id`：一对一映射，PG 负责统计字段，Mongo 负责详细作答记录。
  - `subject_learning.content_ref` ↔ `subject_content.subject_learning_id`：预留同构关系，用于科目学习内容。

