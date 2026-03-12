## 飞校平台后端 API 文档（可用于 Apifox 导入）

> 说明：本文件采用 Markdown 编写，并内嵌一份 OpenAPI 3.0 的 `yaml` 描述。  
> - 你可以复制下面的 `yaml` 代码块保存为 `openapi.yaml`，在 Apifox 中选择「从 OpenAPI 导入」。  
> - 也可以直接把本文件作为接口说明文档，后续持续更新。

---

## 一、OpenAPI 3.0 描述（用于 Apifox 导入）

将下面代码块内容复制到单独文件（例如 `doc/openapi.yaml`），在 Apifox 中选择「OpenAPI 3.0」导入。

```yaml
openapi: 3.0.0
info:
  title: Magua Fly API
  version: 1.0.0
  description: 飞校平台后端接口
servers:
  - url: http://localhost:8080
    description: 本地开发环境（Docker / 本机）

paths:
  /health:
    get:
      summary: 健康检查
      tags: [System]
      responses:
        '200':
          description: 服务健康
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    example: ok

  /auth/login:
    post:
      summary: 用户登录
      tags: [Auth]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                  description: 用户名
                  example: admin
                password:
                  type: string
                  format: password
                  description: 密码
                  example: "123456"
              required: [username, password]
      responses:
        '200':
          description: 登录成功
          content:
            application/json:
              schema:
                type: object
                properties:
                  token:
                    type: string
                    description: JWT 访问令牌
                  user:
                    $ref: '#/components/schemas/User'
        '401':
          description: 账号或密码错误

  /auth/register:
    post:
      summary: 用户注册
      tags: [Auth]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                password:
                  type: string
                  format: password
                phone:
                  type: string
                email:
                  type: string
                  format: email
              required: [username, password]
      responses:
        '201':
          description: 注册成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'

  /students:
    get:
      summary: 学员列表
      description: |
        返回当前系统中的学员列表。  
        需要登录并携带 `Authorization: Bearer <token>`。
      tags: [Student]
      security:
        - bearerAuth: []
      parameters:
        - in: query
          name: page
          schema:
            type: integer
            minimum: 1
          description: 页码（可选）
        - in: query
          name: page_size
          schema:
            type: integer
            minimum: 1
          description: 每页数量（可选）
      responses:
        '200':
          description: 学员列表
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Student'

  /instructors:
    get:
      summary: 教员列表
      tags: [Instructor]
      security:
        - bearerAuth: []
      responses:
        '200':
          description: 教员列表
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Instructor'

  /aircraft:
    get:
      summary: 航空器列表
      tags: [Aircraft]
      security:
        - bearerAuth: []
      responses:
        '200':
          description: 航空器列表
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Aircraft'

  /bookings/pending:
    get:
      summary: 待审核预约列表
      tags: [Booking]
      security:
        - bearerAuth: []
      responses:
        '200':
          description: 待审核的预约列表
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Booking'

  /bookings:
    post:
      summary: 创建预约
      tags: [Booking]
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                student_id:
                  type: integer
                  description: 学员 ID
                instructor_id:
                  type: integer
                  description: 教员 ID
                aircraft_id:
                  type: integer
                  description: 航空器 ID
                plan_date:
                  type: string
                  format: date
                  description: 预约日期（YYYY-MM-DD）
                start_time:
                  type: string
                  format: time
                  description: 开始时间（HH:MM）
                end_time:
                  type: string
                  format: time
                  description: 结束时间（HH:MM）
              required:
                [student_id, instructor_id, aircraft_id, plan_date, start_time, end_time]
      responses:
        '201':
          description: 创建成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Booking'

  /bookings/{id}/review:
    post:
      summary: 审核预约
      tags: [Booking]
      security:
        - bearerAuth: []
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: integer
          description: 预约 ID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                status:
                  type: string
                  enum: [approved, rejected]
                  description: 审核结果
                reject_reason:
                  type: string
                  description: 拒绝原因（当 status = rejected 时必填）
      responses:
        '200':
          description: 审核结果
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Booking'

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    User:
      type: object
      properties:
        id:
          type: integer
        username:
          type: string
        phone:
          type: string
        email:
          type: string
        role:
          type: string
          description: 用户角色：admin/staff/instructor/student

    Student:
      type: object
      properties:
        id:
          type: integer
        user_id:
          type: integer
          nullable: true
        name:
          type: string
        student_no:
          type: string
        flown_minutes:
          type: integer
          description: 已飞行分钟数
        training_progress:
          type: string
          nullable: true
        responsible_instructor_id:
          type: integer
          nullable: true

    Instructor:
      type: object
      properties:
        id:
          type: integer
        user_id:
          type: integer
        name:
          type: string
        display_color:
          type: string
          nullable: true

    Aircraft:
      type: object
      properties:
        id:
          type: integer
        code:
          type: string
        name:
          type: string
          nullable: true
        status:
          type: string
          description: 航空器状态（active/maintenance/blocked 等）

    Booking:
      type: object
      properties:
        id:
          type: integer
        student_id:
          type: integer
        instructor_id:
          type: integer
        aircraft_id:
          type: integer
        plan_date:
          type: string
          format: date
        start_time:
          type: string
          format: time
        end_time:
          type: string
          format: time
        status:
          type: string
          description: 预约状态（pending/approved/rejected）
        reject_reason:
          type: string
          nullable: true
        reviewed_at:
          type: string
          format: date-time
          nullable: true
        reviewer_id:
          type: integer
          nullable: true
```

---

## 二、接口一览表

> 下面是便于人工阅读与维护的接口说明概要，建议以后新增接口时同步更新本表和上面的 OpenAPI 配置。

### 1. 系统相关

- **GET `/health`**  
  - **说明**：健康检查，确认服务是否正常运行。  
  - **鉴权**：否  
  - **返回示例**：
    ```json
    { "status": "ok" }
    ```

### 2. 认证相关

- **POST `/auth/login`**  
  - **说明**：用户登录，获取 JWT。  
  - **鉴权**：否  
  - **请求体**：
    ```json
    {
      "username": "admin",
      "password": "123456"
    }
    ```
  - **返回**：
    ```json
    {
      "token": "jwt-token-string",
      "user": {
        "id": 1,
        "username": "admin",
        "phone": "13800000000",
        "email": "admin@example.com",
        "role": "admin"
      }
    }
    ```

- **POST `/auth/register`**  
  - **说明**：新用户注册。  
  - **鉴权**：否  
  - **请求体**：
    ```json
    {
      "username": "student1",
      "password": "123456",
      "phone": "13800000000",
      "email": "student1@example.com"
    }
    ```
  - **返回**：创建成功的用户信息。

### 3. 学员、教员与航空器

- **GET `/students`**  
  - **说明**：学员列表。  
  - **鉴权**：是（需要 `Authorization: Bearer <token>`）  
  - **返回**：`Student` 数组。

- **GET `/instructors`**  
  - **说明**：教员列表。  
  - **鉴权**：是  
  - **返回**：`Instructor` 数组。

- **GET `/aircraft`**  
  - **说明**：航空器列表。  
  - **鉴权**：是  
  - **返回**：`Aircraft` 数组。

### 4. 预约相关

- **GET `/bookings/pending`**  
  - **说明**：待审核的预约列表（通常给教员或管理员使用）。  
  - **鉴权**：是  
  - **返回**：`Booking` 数组。

- **POST `/bookings`**  
  - **说明**：学员创建飞行预约。  
  - **鉴权**：是  
  - **请求体示例**：
    ```json
    {
      "student_id": 1,
      "instructor_id": 2,
      "aircraft_id": 3,
      "plan_date": "2026-03-12",
      "start_time": "09:00",
      "end_time": "10:00"
    }
    ```

- **POST `/bookings/:id/review`**  
  - **说明**：教员/管理员审核预约。  
  - **鉴权**：是  
  - **路径参数**：`id` – 预约 ID  
  - **请求体示例（通过）**：
    ```json
    { "status": "approved" }
    ```
  - **请求体示例（拒绝）**：
    ```json
    {
      "status": "rejected",
      "reject_reason": "飞行器维护中"
    }
    ```

---

## 三、后续更新建议

- 新增接口时：  
  - 在 `paths` 下补充新的路径及方法；  
  - 如有新数据结构，在 `components.schemas` 下补充对应 Schema；  
  - 在「接口一览表」部分增加一小节说明，便于快速浏览。

