# 项目 2：用户中心 API

## 项目定位

用户中心是后端入门最重要的项目之一。几乎所有后台系统都会复用用户、登录、权限、错误码、分页、数据库连接这些能力。

这个项目分 4 个版本推进：

1. 内存版 API。
2. Gin 分层版。
3. 数据库持久化版。
4. 登录权限版。

## V1：内存版 API

目标：练 HTTP 和 RESTful。

接口：

- [ ] `GET /health`
- [ ] `GET /users`
- [ ] `GET /users/{id}`
- [ ] `POST /users`
- [ ] `PUT /users/{id}`
- [ ] `DELETE /users/{id}`

任务：

- [ ] 设计用户字段。
- [ ] 设计统一响应结构。
- [ ] 设计错误码。
- [ ] 支持参数校验。
- [ ] 写接口测试。

验收：

- [ ] 状态码合理。
- [ ] 错误响应统一。
- [ ] 前端可以直接调用。

## V2：Gin 分层版

目标：练项目结构。

推荐目录：

```text
cmd/server
internal/config
internal/handler
internal/service
internal/repository
internal/model
internal/middleware
internal/pkg/response
```

任务：

- [ ] handler 只处理 HTTP。
- [ ] service 处理业务规则。
- [ ] repository 提供数据访问接口。
- [ ] middleware 处理日志、CORS、request id。
- [ ] config 处理配置加载。

验收：

- [ ] service 不依赖 Gin。
- [ ] repository 可以替换实现。
- [ ] 错误码集中维护。

## V3：数据库持久化版

目标：练数据建模和 repository。

推荐表：

```text
users
- id
- username
- email
- phone
- password_hash
- status
- created_at
- updated_at
- deleted_at
```

任务：

- [ ] 选择 PostgreSQL 或 MySQL。
- [ ] 写迁移文件。
- [ ] 邮箱或手机号唯一。
- [ ] 支持分页查询。
- [ ] 支持关键词搜索。
- [ ] 处理唯一冲突。
- [ ] repository 方法接收 context。

验收：

- [ ] 唯一约束由数据库保证。
- [ ] 数据库配置来自环境变量。
- [ ] 查询错误不会原样暴露给前端。

## V4：登录权限版

目标：练认证鉴权。

功能：

- [ ] 注册。
- [ ] 登录。
- [ ] 获取当前用户 `GET /me`。
- [ ] 修改当前用户资料。
- [ ] 管理员查看用户列表。
- [ ] 管理员禁用用户。

任务：

- [ ] 密码哈希。
- [ ] JWT 或 Cookie Session。
- [ ] 认证 middleware。
- [ ] 角色字段或角色表。
- [ ] 管理员权限校验。

验收：

- [ ] 不保存明文密码。
- [ ] secret 不硬编码。
- [ ] 未登录不能访问受保护接口。
- [ ] 普通用户不能访问管理员接口。

## 可以让 AI 做

- [ ] 根据你的接口表生成 handler。
- [ ] 生成 Gin 项目骨架。
- [ ] 生成迁移 SQL。
- [ ] 生成 repository 代码。
- [ ] 生成 JWT middleware。
- [ ] 生成测试样例。

## 你必须审查

- [ ] 表字段是否过度或不足。
- [ ] 唯一约束是否只在代码层校验。
- [ ] token 里是否放了敏感信息。
- [ ] 错误信息是否会帮助枚举账号。
- [ ] 权限是否只在前端控制。

## 最终交付物

- [ ] 后端代码。
- [ ] API 文档。
- [ ] 数据库迁移文件。
- [ ] `.env.example`。
- [ ] 基础测试。
- [ ] README。
