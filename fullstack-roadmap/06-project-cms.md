# 项目 3：CMS 内容管理系统

## 项目定位

CMS 是第一个完整作品集项目。它连接了前端页面、后端 API、数据库、认证、权限、上传、缓存和部署。

目标不是做一个功能庞大的博客平台，而是做一个结构完整、能展示全栈能力的系统。

## 角色

- 访客：浏览文章。
- 作者：创建和编辑自己的文章。
- 管理员：管理文章、分类、标签、用户。

## 核心功能

管理端：

- [ ] 登录。
- [ ] 文章列表。
- [ ] 创建文章。
- [ ] 编辑文章。
- [ ] 草稿、发布、下架。
- [ ] 分类管理。
- [ ] 标签管理。
- [ ] 图片上传。
- [ ] 用户管理。

前台：

- [ ] 文章列表。
- [ ] 文章详情。
- [ ] 分类筛选。
- [ ] 标签筛选。
- [ ] 关键词搜索。

## 推荐数据表

```text
users
roles
permissions
articles
categories
tags
article_tags
media_files
operation_logs
```

## 文章状态流转

```text
draft -> published -> archived
published -> draft
```

需要禁止：

- [ ] 已删除文章继续发布。
- [ ] 普通作者修改别人的文章。
- [ ] 前台看到 draft 或 archived。

## API 模块

认证：

- [ ] `POST /auth/register`
- [ ] `POST /auth/login`
- [ ] `GET /auth/me`
- [ ] `POST /auth/logout`

管理端文章：

- [ ] `GET /admin/articles`
- [ ] `POST /admin/articles`
- [ ] `GET /admin/articles/{id}`
- [ ] `PUT /admin/articles/{id}`
- [ ] `POST /admin/articles/{id}/publish`
- [ ] `POST /admin/articles/{id}/archive`
- [ ] `DELETE /admin/articles/{id}`

前台文章：

- [ ] `GET /articles`
- [ ] `GET /articles/{slug}`

媒体：

- [ ] `POST /admin/media`
- [ ] `GET /admin/media`
- [ ] `DELETE /admin/media/{id}`

## 需求拆解任务

- [ ] 设计文章字段。
- [ ] 设计分类和标签关系。
- [ ] 设计 slug 规则。
- [ ] 设计文章状态流转。
- [ ] 设计前台和后台接口差异。
- [ ] 设计上传限制。
- [ ] 设计缓存策略。
- [ ] 设计权限规则。

## 可以让 AI 做

- [ ] 生成 ERD 草稿。
- [ ] 生成迁移文件。
- [ ] 生成 API 文档。
- [ ] 生成 CRUD 初版代码。
- [ ] 生成前端页面骨架。
- [ ] 生成 Dockerfile 和 docker-compose 草稿。

## 你必须审查

- [ ] 多对多表是否合理。
- [ ] 草稿文章是否会被前台看到。
- [ ] 文章 slug 是否唯一。
- [ ] 作者权限是否正确。
- [ ] 上传文件是否限制大小和类型。
- [ ] 缓存是否会读到旧文章。
- [ ] 删除文章是否影响标签和媒体。

## 缓存任务

- [ ] 缓存文章详情。
- [ ] 缓存热门文章。
- [ ] 文章更新后删除缓存。
- [ ] Redis 异常时降级到数据库。

## 测试任务

- [ ] 登录成功。
- [ ] 登录失败。
- [ ] 未登录访问管理端失败。
- [ ] 作者不能编辑别人的文章。
- [ ] 草稿不出现在前台列表。
- [ ] 发布后前台可见。
- [ ] 更新文章后缓存失效。

## 部署任务

- [ ] 提供 `.env.example`。
- [ ] 写 Dockerfile。
- [ ] 写 docker-compose。
- [ ] 包含 Go 服务、数据库、Redis。
- [ ] 提供数据库迁移命令。
- [ ] 写 README。

## 最小验收标准

- [ ] 管理端可以登录。
- [ ] 可以创建、编辑、发布文章。
- [ ] 前台可以浏览已发布文章。
- [ ] 图片可以上传。
- [ ] 数据保存在数据库。
- [ ] 项目可以通过 Docker 启动。

## 作品集说明建议

README 里要写清楚：

- 为什么这样设计表。
- 为什么前台和后台接口分开。
- 权限如何校验。
- 缓存如何失效。
- AI 生成了哪些代码。
- 你审查和修改了哪些问题。
