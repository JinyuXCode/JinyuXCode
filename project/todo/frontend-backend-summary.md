# Todo CLI 项目理解总结：从前端模块化到后端分层

这份文档站在 Vue / React / TypeScript 前端开发者的视角，帮助理解当前 Go Todo CLI 项目的代码组织、职责边界和后端分层思想。

## 1. 这个项目是什么

当前项目是一个基于 Go 的命令行 Todo 工具。

它不是 Web 项目，没有浏览器页面、HTTP API、数据库服务，也没有前后端分离。用户通过命令行输入命令，程序读写本地 `todo.json` 文件完成任务管理。

典型使用方式：

```bash
todo add "Buy milk"
todo list --status pending
todo done 1
todo edit 1 "Buy bread"
todo delete 1
```

从前端视角类比，它像一个“没有 UI 的 Todo 应用”。命令行参数就是用户输入，本地 JSON 文件就是持久化数据源。

## 2. 前端项目里的常见思维

前端项目通常会拆成这些模块：

```text
components/      UI 组件
pages/           页面入口
stores/          Pinia / Redux / Zustand 状态
services/        请求 API 的封装
types/           TypeScript 类型
utils/           工具函数
```

比如一个前端 Todo 应用可能是：

```text
TodoPage.vue / TodoPage.tsx
TodoList.vue / TodoList.tsx
todoStore.ts
todoApi.ts
types.ts
```

前端的核心问题通常是：

- 用户怎么操作界面
- 组件之间怎么传数据
- 状态如何更新
- 如何调用后端 API
- 页面如何响应加载、错误、空状态

## 3. 当前 Go 项目的模块对应关系

当前项目主要文件是：

```text
main.go
store.go
main_test.go
store_test.go
go.mod
prd.md
```

它看起来文件少，是因为项目规模小，而且当前只有一个入口：CLI。

可以这样类比前端：

```text
main.go   ≈ 页面入口 + 事件处理 + 表单参数解析
store.go  ≈ 状态管理 + 业务逻辑 + 本地持久化
tests     ≈ 单元测试 / 集成测试
todo.json ≈ localStorage / IndexedDB / 后端数据库
```

更具体一点：

```text
命令行输入
  -> main.go 解析命令和参数
  -> store.go 执行业务规则
  -> store.go 读写 todo.json
  -> main.go 输出结果或错误
```

## 4. 为什么现在没有拆很多包

不是后端不需要模块化，而是当前项目还比较小。

现在只有一个业务实体 `Task`，一个存储文件 `todo.json`，一个入口 `CLI`，没有 HTTP、数据库、用户系统、权限、多端同步等复杂场景。

所以当前结构是合理的轻量版本：

```text
main.go   负责输入输出
store.go  负责任务和存储
```

如果一开始就拆成很多包，可能会变成“目录很多，但抽象没有实际价值”。

后端分层的重点不是目录数量，而是职责边界清楚。

## 5. handler / service / repository 是什么

这是后端常见分层方式。

```text
handler      入口层，处理外部输入输出
service      业务层，处理业务规则和用例流程
repository   数据访问层，处理数据库、文件、缓存等存储细节
domain       领域模型，定义核心实体和业务概念
```

Web 后端里常见流程：

```text
HTTP Request
  -> handler/controller
  -> service
  -> repository
  -> database
```

当前 CLI 项目可以类比为：

```text
CLI command
  -> handler: 解析 add/list/done 参数，打印结果
  -> service: 校验标题、完成任务、编辑任务、排序筛选
  -> repository: 读写 todo.json
```

当前代码里的实际状态是：

```text
main.go   ≈ CLI handler
store.go  ≈ service + repository 合在一起
```

这在小项目里可以接受。

## 6. 和前端模块的对照

### handler 对应什么

后端的 `handler` 类似前端里的页面事件处理、表单提交逻辑。

前端例子：

```ts
async function onSubmit(title: string) {
  if (!title.trim()) {
    setError('标题不能为空')
    return
  }

  await todoApi.add(title)
  await reloadTodos()
}
```

当前 Go 项目里的类似职责在 `main.go`：

```go
todo add <title>
todo done <id>
todo list --status pending
```

它负责：

- 判断用户输入了什么命令
- 校验参数格式
- 调用 Store 执行操作
- 输出成功信息或错误信息
- 返回退出码

### service 对应什么

`service` 类似前端 store/action 里的业务规则，但更靠近后端核心业务。

前端 store 里可能会写：

```ts
function addTodo(title: string) {
  const normalized = title.trim()
  if (!normalized) throw new Error('标题不能为空')

  todos.value.push({
    id: nextId++,
    title: normalized,
    status: 'pending',
  })
}
```

当前 Go 项目里的类似逻辑在 `store.go`：

```go
Add(title)
Done(id)
Undo(id)
Edit(id, title)
Delete(id)
List(status)
```

它负责：

- 标题不能为空
- 标题不能超过 200 字符
- 标题不能包含控制字符
- `done` / `undo` 是幂等操作
- `id` 不存在时返回错误
- `list` 按 pending 在前、done 在后排序

### repository 对应什么

`repository` 类似前端里的 `todoApi.ts`，只不过它不是请求远端 API，而是读写本地文件。

前端：

```ts
export function fetchTodos() {
  return request.get('/api/todos')
}
```

当前项目：

```go
Load()
Save()
saveData()
```

它负责：

- 读取 `todo.json`
- 文件不存在时初始化
- JSON 损坏时返回明确错误
- 写入时使用临时文件 + 重命名，降低文件损坏风险

## 7. 当前代码为什么把 service 和 repository 合在一起

因为目前业务很小，`Store` 同时做了两件事：

```text
业务操作：Add / Done / Undo / Edit / Delete / List
文件存储：Load / Save / saveData
```

这不是“错误”，而是小项目里的务实做法。

什么时候应该拆开？

- 要支持 SQLite / MySQL，而不是 JSON 文件
- 要同时支持 CLI 和 HTTP API
- 业务规则变复杂，比如标签、优先级、截止时间、用户权限
- 测试时希望 mock 数据层
- 多人协作时同一个文件频繁冲突

## 8. 如果拆成后端标准分层会是什么样

一个更标准的学习版目录可以是：

```text
cmd/todo/main.go
internal/domain/task.go
internal/handler/cli.go
internal/service/todo_service.go
internal/repository/json_repository.go
```

职责如下：

```text
domain/task.go
  定义 Task、Status、业务常量

handler/cli.go
  解析命令行参数
  输出结果和错误
  不直接读写文件

service/todo_service.go
  实现 Add、Done、Undo、Edit、Delete、List
  执行业务校验
  调用 repository 保存数据

repository/json_repository.go
  只负责 todo.json 的读写
  不决定任务能不能完成
```

依赖方向应该是：

```text
handler -> service -> repository
service -> domain
repository -> domain
```

不要反过来让 `repository` 调用 `handler`，也不要让 `domain` 依赖外部输入输出。

## 9. 前端转后端时要特别注意的差异

### 前端更关注 UI 状态，后端更关注数据一致性

前端常见问题：

- loading 状态
- error toast
- 组件重渲染
- 表单校验
- 路由跳转

后端常见问题：

- 数据是否真的写入成功
- 写入失败时内存态是否污染
- 错误码是否稳定
- 数据文件损坏怎么处理
- 并发写入是否安全
- 输入是否可信

这个项目里已经体现了几个后端关注点：

- `todo.json` 损坏时要明确报错
- 写入使用临时文件 + 重命名
- 写入失败时不能提前修改内存态
- 参数错误和运行错误退出码不同

### 前端可以乐观更新，后端不能随便乐观提交

前端经常会：

```text
先更新 UI，再请求 API，失败后回滚
```

后端更常见的是：

```text
先验证输入
在副本上计算新状态
持久化成功
再提交内存态
```

当前项目修复后的 `Store` 就是这个思路。

### 前端类型更多服务于开发体验，后端类型更多服务于数据边界

前端 TS 类型常用于组件 props、接口返回值、状态结构。

Go 里的结构体不仅是类型，也直接影响 JSON 序列化：

```go
type Task struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

这里的 `json:"id"` 类似 TS 里接口字段和后端返回字段之间的约定。

## 10. 当前项目的核心业务规则

任务字段：

```text
id        自增整数，不复用
title     1~200 字符，去掉首尾空白后不能为空
status    pending / done
createdAt 创建时间
updatedAt 更新时间
```

命令规则：

```text
add     新增任务，默认 pending
done    pending -> done，重复 done 视为成功
undo    done -> pending，重复 undo 视为成功
edit    修改标题
delete  真删除
list    默认展示全部，pending 在前，done 在后
```

错误规则：

```text
退出码 0：成功
退出码 1：文件读写、JSON 解析等运行错误
退出码 2：参数错误或业务校验失败
```

## 11. 如何继续学习这个项目

建议按这个顺序理解：

1. 先看 `prd.md`，理解需求和验收标准。
2. 再看 `main.go`，理解命令从哪里进入。
3. 再看 `store.go`，理解任务如何被修改和保存。
4. 最后看测试，理解哪些行为被固定下来。

如果你想练习后端分层，下一步可以把当前项目重构为：

```text
domain
service
repository
handler
```

但建议先不要为了分层而分层。你可以先问自己：

- 如果明天把 JSON 换成 SQLite，要改多少代码？
- 如果明天加 HTTP API，业务逻辑能不能复用？
- 如果测试时不想真的写文件，能不能 mock repository？

这些问题的答案，才是是否需要分层的依据。

## 12. 一句话总结

前端模块化通常围绕“页面、组件、状态、请求”展开；后端分层通常围绕“入口、业务、数据”展开。当前 Todo 项目因为很小，所以用 `main.go + store.go` 就够了，但它已经包含了后端分层的雏形：`main.go` 处理输入输出，`store.go` 处理业务和持久化。等项目变复杂后，再把隐含边界拆成 `handler / service / repository` 会更自然。
