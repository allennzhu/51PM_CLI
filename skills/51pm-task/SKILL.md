---
name: 51pm-task
description: 51PM项目管理系统任务管理技能。支持查询任务列表，按状态、名称、项目、指派人、日期等条件筛选任务。当用户说"查看我的任务"、"看看任务列表"、"有哪些任务"、"查一下项目任务"、"任务进度怎么样"、"谁的任务"等需要查询51PM任务的场景时使用。
metadata:
  requires:
    bins: ["51pm"]
  cliHelp: "51pm --help"
---

# 51PM 项目任务管理

> `51pm` 是51PM项目管理系统的命令行工具，所有操作通过执行 `51pm` 命令完成。

> **重要**：`task list` 命令会自动同时查询**项目任务**和**非项目任务**两种数据并合并输出，无需分别调用。

## 前提条件

使用前必须先完成登录：
```bash
51pm login --token <TOKEN>
```
Token 可从 51PM 前端页面登录后，在浏览器 F12 -> Application -> Local Storage -> oauthToken 获取。

如果命令返回"未登录"或"Token已过期"，提示用户重新执行 login。

## 全局参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| --base-url | 51PM API 服务地址 | http://localhost:8888 |

## 命令说明

### 查询任务列表

```bash
51pm task list [flags]
```

#### 参数

| 参数 | 类型 | 说明 |
|------|------|------|
| --status | string | 任务状态筛选（如 open、doing、done、closed） |
| --name | string | 任务名称模糊查询 |
| --type | int | 任务类型 |
| --assigned-to | int | 指派给（用户ID） |
| --project-id | int | 项目ID |
| --dept-id | int | 部门ID |
| --start-date | string | 开始日期（如 2026-01-01） |
| --end-date | string | 结束日期（如 2026-12-31） |
| --page | int | 页码（默认 1） |
| --limit | int | 每页条数（默认 10） |
| --json | bool | 以 JSON 格式输出（用于程序解析） |

#### 使用示例

```bash
# 查看所有任务
51pm task list

# 按状态筛选
51pm task list --status open

# JSON 格式输出（Copilot/AI Agent 调用推荐）
51pm task list --json

# 指定 API 地址
51pm --base-url http://10.67.8.189:8888 task list --json
```

#### JSON 输出结构

`--json` 输出的统一结构：
```json
{
  "data": [
    {
      "id": 123,
      "source": "project",
      "project_name": "XX项目",
      "name": "任务名称",
      "assigned_to": "474",
      "status": "done",
      "start_date": "2026-01-01",
      "end_date": "2026-01-05",
      "plan_hour": 20,
      "consumed_hour": 15,
      "task_process": 100
    },
    {
      "id": 456,
      "source": "non-project",
      "project_name": "XX非项目",
      "name": "非项目任务名称",
      "assigned_to": "474",
      "status": "done",
      "start_date": "2026-01-06",
      "end_date": "2026-01-06",
      "plan_hour": 8,
      "consumed_hour": 8,
      "task_process": 0
    }
  ],
  "project_total": 5,
  "non_project_total": 3,
  "total": 8
}
```

- `source` 字段区分任务来源：`"project"`（项目任务）或 `"non-project"`（非项目任务）
- `project_total` 和 `non_project_total` 分别统计两种任务的总数

### 用户名称转用户ID

`--assigned-to` 参数需要传入用户ID（int），但用户通常提供的是姓名。需要先通过 `user lookup` 命令将姓名转换为用户ID：

```bash
51pm user lookup --name <用户名称> --json
```

返回示例：
```json
[
  {
    "id": 42,
    "nickname": "张三",
    "realname": "张三",
    "account": "zhangsan"
  }
]
```

#### 典型工作流

用户说："查看张三的任务"

1. 先查用户ID：
```bash
51pm user lookup --name 张三 --json
```
2. 从返回结果中取 `id` 字段（如 42）
3. 再查任务：
```bash
51pm task list --assigned-to 42 --json
```

> **注意**：如果 `user lookup` 返回多个匹配用户，需展示候选列表请用户确认后再查询任务。

## 行为策略

- **AI Agent 调用时始终使用 --json 参数**，以便解析结构化数据向用户展示
- **当用户通过姓名指定指派人时，必须先调用 `51pm user lookup --name xxx --json` 获取用户ID，再传入 `--assigned-to`**
- 如果返回 total > per_page，说明还有更多数据，主动告知用户
- 遇到请求失败时可重试 1 次
- 若返回 Token 过期错误，提示用户重新执行 51pm login
