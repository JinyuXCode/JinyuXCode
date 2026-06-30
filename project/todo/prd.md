# Todo List

## 业务目标

实现一个基于 Go 的命令行 Todo 管理工具，支持新增、完成、删除、查看和筛选任务。数据使用本地 JSON 文件持久化，程序重启后任务数据仍然保留。

## 使用角色

- Windows 单机用户
- 单用户、单实例本地使用，不提供账号体系和远程访问

## 核心实体

```bash
id: 自增整数
title: 字符串
status: pending/done
createdAt: 创建时间
updatedAt: 最近更新时间
```

## 业务规则

- Task 只有两种状态：`pending` 和 `done`
- `add` 创建的任务默认状态为 `pending`
- `done` 仅将任务状态从 `pending` 切换为 `done`
- `undo` 仅将任务状态从 `done` 恢复为 `pending`
- `delete` 采用真删除，删除后任务不再出现在列表中，也不保留在 JSON 文件内
- `list` 默认展示全部任务，默认排序为 `pending` 在前、`done` 在后，同状态内按 `id` 升序
- `updatedAt` 在任务创建、完成、恢复、编辑时更新
- `createdAt` 在任务首次创建时写入，之后不变
- `id` 只递增不复用，删除后也不回收

## 字段设计

```bash
id: int，自增主键
title: string，1~200 字符，去掉首尾空白后不能为空
status: string，仅允许 pending/done
createdAt: time.Time，RFC3339 格式
updatedAt: time.Time，RFC3339 格式
```

## 命令设计

```bash
todo add <title>
todo list [--status pending|done|all]
todo done <id>
todo undo <id>
todo edit <id> <new-title>
todo delete <id>
todo help
```

## 命令行为

- `todo add <title>`：创建任务，返回新任务 `id`
- `todo list`：展示任务列表，默认 `--status all`
- `todo list --status pending`：仅展示待办任务
- `todo list --status done`：仅展示已完成任务
- `todo done <id>`：将指定任务标记为完成
- `todo undo <id>`：将指定任务恢复为待办
- `todo edit <id> <new-title>`：修改任务标题
- `todo delete <id>`：删除指定任务
- `todo help`：输出命令用法和参数说明

## 结果约定

- `done` 对已完成任务执行时，视为幂等成功
- `undo` 对待办任务执行时，视为幂等成功
- `delete` 对不存在任务执行时，返回错误
- `done`、`undo`、`edit`、`delete` 对不存在的 `id` 都应返回错误
- `add` 的标题为空、仅空白、或超过长度上限时应返回错误

## 存储设计

JSON 文件采用单文件存储，建议结构如下：

```json
{
  "version": 1,
  "nextId": 3,
  "tasks": [
    {
      "id": 1,
      "title": "Buy milk",
      "status": "pending",
      "createdAt": "2026-06-30T10:00:00+08:00",
      "updatedAt": "2026-06-30T10:00:00+08:00"
    }
  ]
}
```

- `version` 用于后续结构升级
- `nextId` 用于保证 id 单调递增
- `tasks` 保存任务数组，数组内元素按 `id` 升序保存

## 文件规则

- 数据文件默认存放在当前程序工作目录下，文件名建议为 `todo.json`
- 如果文件不存在，程序应自动创建
- 如果文件损坏、无法解析或无法写入，应给出明确错误信息
- 写入时必须采用临时文件 + 原子重命名的方式，避免程序异常退出后文件损坏
- 程序仅保证单实例本地使用，不保证多进程并发写入安全

## 异常场景

- JSON 文件不存在：自动创建并初始化为空任务列表
- JSON 文件损坏：返回解析失败错误，并提示用户备份或重建文件
- 读取权限不足：返回文件不可读错误
- 写入权限不足：返回文件不可写错误
- `id` 不存在：返回“任务不存在”错误
- `title` 为空：返回“标题不能为空”错误
- 参数缺失或参数格式错误：返回帮助信息和错误码

## 权限规则

- 当前版本不提供登录、注册、角色鉴权
- 默认认为本地 Windows 单用户环境可信
- 不提供跨用户共享和远程访问能力

## 缓存规则

- 可在内存中缓存已加载的任务列表
- 每次写入后必须同步刷新内存态与文件态，避免读写不一致
- 如果不做缓存，也必须保证所有命令读取到的是最新文件内容

## 退出码

- `0`：执行成功
- `1`：通用运行错误，如文件读写失败、解析失败
- `2`：参数错误或业务校验失败，如 `id` 不存在、标题为空

## 验收标准

- 能新增、完成、恢复、删除、编辑和展示任务
- 程序重启后任务仍然存在
- `list` 能按状态筛选并按约定排序
- 空标题、无效 `id`、损坏 JSON 文件等场景有明确提示
- 文件写入采用原子更新方式，避免中断后文件损坏

## 测试用例

- 新增任务后列表可见
- 完成任务后状态变为 `done`
- 恢复任务后状态变回 `pending`
- 删除任务后列表中不再出现
- 重启程序后任务数据仍然存在
- 空标题、无效 id、损坏 JSON 文件等异常路径可正确返回错误
