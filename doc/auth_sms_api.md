## 鉴权与短信验证码接口说明

- **说明**：本系统未特别声明时，接口统一使用 **POST** 方法。
- **Base URL（本地开发）**：`http://localhost:8080`

---

## 1. 用户名密码登录

- **接口地址**：`/auth/login`
- **请求方式**：POST
- **请求体（JSON）**：

```json
{
  "username": "admin",
  "password": "password123"
}
```

- **响应体（200 OK）**：

```json
{
  "token": "jwt-token-string",
  "user": {
    "id": 1,
    "username": "admin",
    "phone": "13800138000",
    "email": "admin@example.com",
    "role": "admin",
    "avatar_url": "",
    "created_at": "2026-03-13T12:00:00Z",
    "updated_at": "2026-03-13T12:00:00Z"
  }
}
```

- **错误示例**：
  - 400：请求体格式错误
  - 401：`{ "error": "invalid credentials" }`

---

## 2. 用户注册（用户名密码）

- **接口地址**：`/auth/register`
- **请求方式**：POST
- **请求体（JSON）**：

```json
{
  "username": "new_user",
  "password": "password123",
  "phone": "13800138000",
  "email": "user@example.com",
  "role": "student"
}
```

- **字段说明**：
  - `role`：可选值 `admin | staff | instructor | student`

- **响应体（201 Created）**：

```json
{
  "id": 2,
  "username": "new_user",
  "phone": "13800138000",
  "email": "user@example.com",
  "role": "student",
  "avatar_url": "",
  "created_at": "2026-03-13T12:00:00Z",
  "updated_at": "2026-03-13T12:00:00Z"
}
```

---

## 3. 发送短信验证码

- **接口地址**：`/auth/sms/send`
- **请求方式**：POST
- **请求体（JSON）**：

```json
{
  "phone": "13800138000",
  "scene": "login"
}
```

- **字段说明**：
  - `phone`：手机号（后端默认按中国大陆处理，未带 `+` 前缀时自动补全为 `+86`）
  - `scene`：
    - `"login"`：短信登录
    - `"register"`：短信注册 + 登录一体
    - `"reset_pwd"`：找回/重置密码
    - `"2fa"`：高风险操作二次验证

- **成功响应（200 OK）**：

```json
{
  "status": "ok"
}
```

- **错误情况**：
  - 400：参数错误或发送过于频繁，例如：

```json
{ "error": "too many requests, please try later" }
```

> 频率限制：同一 `phone + scene` 在 60 秒内只能发送一次验证码。

---

## 4. 校验短信验证码（登录 / 注册 / 其它场景）

- **接口地址**：`/auth/sms/verify`
- **请求方式**：POST
- **请求体（JSON）**：

```json
{
  "phone": "13800138000",
  "scene": "login",
  "code": "123456"
}
```

- **响应（场景说明）**：

### 4.1 短信登录：`scene = "login"`

- **成功（200 OK）**：

```json
{
  "token": "jwt-token-string",
  "user": {
    "id": 1,
    "username": "13800138000",
    "phone": "13800138000",
    "email": "",
    "role": "student",
    "avatar_url": "",
    "created_at": "2026-03-13T12:00:00Z",
    "updated_at": "2026-03-13T12:00:00Z"
  }
}
```

- 要求：该手机号已存在用户，否则返回 401。

### 4.2 短信注册 + 登录：`scene = "register"`

- 若手机号不存在用户：
  - 自动创建一个 `student` 角色用户（`username = phone`，密码为空，只能通过短信登录）。
  - 并返回与登录场景相同的 `token + user`。
- 若手机号已存在用户：
  - 仅校验验证码并返回该用户的 `token + user`。

### 4.3 重置密码：`scene = "reset_pwd"`

- 当前实现：仅校验验证码是否正确、未过期、未使用。
- **成功（200 OK）**：

```json
{ "status": "ok" }
```

- 建议：在校验成功后，由前端携带手机号 + 新密码调用单独的重置密码接口（可后续扩展）。

### 4.4 二次验证（2FA）：`scene = "2fa"`

- 当前实现：仅做验证码校验并返回：

```json
{ "status": "ok" }
```

- 建议：在需要 2FA 的高风险操作前，先调用本接口校验成功，再在后续 handler 中检查相应标记（可按业务需要扩展）。

- **通用错误**：
  - 400：参数错误
  - 401：验证码错误或已过期，例如：

```json
{ "error": "invalid code" }
```

