# 前端转 Go 全栈学习路线图

> 适用对象：已有 Vue / React / TypeScript / 前端工程化经验，刚学完 Go 基础语法，准备借助 AI 从前端转全栈。
>
> 核心原则：不要把主要时间花在背完所有语法和库函数上。你要训练的是需求拆解、架构判断、AI 协作、代码审查、测试验证和项目交付。

## 推荐阅读顺序

1. [学习曲线与能力目标](./fullstack-roadmap/00-learning-curve.md)
2. [AI 协作学习工作流](./fullstack-roadmap/01-ai-workflow.md)
3. [12 周执行计划](./fullstack-roadmap/02-12-week-plan.md)
4. [阶段任务清单](./fullstack-roadmap/03-stage-task-lists.md)
5. [项目 1：命令行 Todo](./fullstack-roadmap/04-project-cli-todo.md)
6. [项目 2：用户中心 API](./fullstack-roadmap/05-project-user-center.md)
7. [项目 3：CMS 内容管理系统](./fullstack-roadmap/06-project-cms.md)
8. [项目 4：任务协作系统或电商后台](./fullstack-roadmap/07-project-business-system.md)
9. [审查清单与验收标准](./fullstack-roadmap/08-review-checklists.md)

## 学习曲线总览

```text
Go 项目手感
  -> HTTP API
  -> Gin 分层项目
  -> 数据库建模与 SQL
  -> 登录认证与权限
  -> CMS 完整前后端项目
  -> Redis / 文件上传 / 异步任务
  -> 测试 / Docker / 部署
  -> 任务协作系统或电商后台
```

这条曲线刻意避免一开始就堆概念。前 3 周先建立后端服务的基本手感，第 4 到第 7 周补齐数据库、认证、权限这些真实项目底座，第 8 周开始做完整作品，第 12 周再挑战更复杂的业务系统。

## 项目顺序

| 顺序 | 项目 | 训练重点 | 不建议跳过的原因 |
| --- | --- | --- | --- |
| 1 | 命令行 Todo | Go 基础、错误处理、测试 | 先把 Go 写顺，不被 Web 框架干扰 |
| 2 | 用户中心 API | HTTP、Gin、分层、数据库、登录 | 几乎所有后台系统都复用这套能力 |
| 3 | CMS 内容管理系统 | 前后端联调、权限、上传、缓存、部署 | 适合作为第一个完整作品集 |
| 4 | 任务协作系统或电商后台 | 状态流转、事务、权限矩阵、并发 | 证明你不只是会 CRUD |

## 每周固定产出

- 一份需求拆解文档。
- 一份 API 文档或接口表。
- 一份数据库表设计。
- 一组可运行代码。
- 一组基础测试或接口测试记录。
- 一份复盘：AI 生成了什么、你改了什么、你审查出了什么问题。

## 学习方式

每个功能都按这个节奏做：

```text
1. 你先写需求拆解
2. 让 AI 审查需求拆解
3. 你补充边界条件
4. 让 AI 生成方案
5. 你审查方案
6. 让 AI 生成代码
7. 你运行、调试、测试
8. 让 AI 做 code review
9. 你最终修改并复盘
```

你可以把这个入口文件当导航页。真正执行时，按 `02-12-week-plan.md` 逐周推进，遇到具体项目时打开对应项目文档。
