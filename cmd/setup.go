package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const skillContent = `---
name: 51pm-task
description: 51PM项目管理系统任务管理技能。支持查询任务列表，按状态、名称、项目、指派人、日期等条件筛选任务。当用户说"查看我的任务"、"看看任务列表"、"有哪些任务"、"查一下项目任务"、"任务进度怎么样"、"谁的任务"等需要查询51PM任务的场景时使用。
metadata:
  requires:
    bins: ["51pm"]
---

# 51PM 项目任务管理

> ` + "`51pm`" + ` 是51PM项目管理系统的命令行工具，所有操作通过执行 ` + "`51pm`" + ` 命令完成。

## 前提条件

使用前必须先完成登录：
` + "```bash" + `
51pm login --token <TOKEN>
` + "```" + `
Token 可从 51PM 前端页面登录后，在浏览器 F12 -> Application -> Local Storage -> oauthToken 获取。

如果命令返回"未登录"或"Token已过期"，提示用户重新执行 login。

## 全局参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| --base-url | 51PM API 服务地址 | http://localhost:8888 |

## 命令说明

### 查询任务列表

` + "```bash" + `
51pm task list [flags]
` + "```" + `

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

` + "```bash" + `
# 查看所有任务
51pm task list

# 按状态筛选
51pm task list --status open

# JSON 格式输出（Copilot/AI Agent 调用推荐）
51pm task list --json

# 指定 API 地址
51pm --base-url http://10.67.8.189:8888 task list --json
` + "```" + `

### 用户名称转用户ID

` + "`--assigned-to`" + ` 参数需要传入用户ID（int），但用户通常提供的是姓名。需要先通过 ` + "`user lookup`" + ` 命令将姓名转换为用户ID：

` + "```bash" + `
51pm user lookup --name <用户名称> --json
` + "```" + `

返回示例：
` + "```json" + `
[
  {
    "id": 42,
    "nickname": "张三",
    "realname": "张三",
    "account": "zhangsan"
  }
]
` + "```" + `

#### 典型工作流

用户说："查看张三的任务"

1. 先查用户ID：
` + "```bash" + `
51pm user lookup --name 张三 --json
` + "```" + `
2. 从返回结果中取 ` + "`id`" + ` 字段（如 42）
3. 再查任务：
` + "```bash" + `
51pm task list --assigned-to 42 --json
` + "```" + `

> **注意**：如果 ` + "`user lookup`" + ` 返回多个匹配用户，需展示候选列表请用户确认后再查询任务。

## 行为策略

- **AI Agent 调用时始终使用 --json 参数**，以便解析结构化数据向用户展示
- **当用户通过姓名指定指派人时，必须先调用 ` + "`51pm user lookup --name xxx --json`" + ` 获取用户ID，再传入 ` + "`--assigned-to`" + `**
- 如果返回 total > per_page，说明还有更多数据，主动告知用户
- 遇到请求失败时可重试 1 次
- 若返回 Token 过期错误，提示用户重新执行 51pm login
`

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "安装Copilot/AI Agent技能文件",
	Long:  `将51PM Skill文件安装到本机，使Copilot和其他AI Agent能够调用51pm CLI。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("获取用户目录失败: %w", err)
		}

		// 安装到 ~/.agents/skills/51pm-task/
		skillDir := filepath.Join(home, ".agents", "skills", "51pm-task")
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}

		skillPath := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
			return fmt.Errorf("写入Skill文件失败: %w", err)
		}

		fmt.Println("✅ Copilot Skill 安装成功!")
		fmt.Printf("   文件位置: %s\n", skillPath)
		fmt.Println()
		fmt.Println("现在可以在 VS Code Copilot 中直接说：")
		fmt.Println("   \"查看我的任务\"")
		fmt.Println("   \"项目任务进度怎么样\"")
		fmt.Println("   \"筛选状态为doing的任务\"")
		fmt.Println()
		fmt.Println("⚠️  使用前请确保已执行: 51pm login --token <TOKEN>")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
