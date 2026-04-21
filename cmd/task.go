package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// Task 任务信息
type Task struct {
	Id           int     `json:"id"`
	ProjectId    int     `json:"project_id"`
	ProjectName  string  `json:"project_name"`
	Name         string  `json:"name"`
	AssignedTo   string  `json:"assigned_to"`
	Status       string  `json:"status"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	StandardHour float64 `json:"standard_hour"`
	EstimateHour float64 `json:"estimate_hour"`
	ConsumedHour float64 `json:"consumed_hour"`
	LeftHour     float64 `json:"left_hour"`
	TaskProcess  float64 `json:"task_process"`
	OneType      int     `json:"one_type"`
	IsIncome     int     `json:"is_income"`
	Desc         string  `json:"desc"`
	Remark       string  `json:"remark"`
}

// TaskListResponse 任务列表响应
type TaskListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Data        []*Task `json:"data"`
		Total       int     `json:"total"`
		PerPage     int     `json:"per_page"`
		CurrentPage int     `json:"current_page"`
		LastPage    int     `json:"last_page"`
	} `json:"data"`
}

var (
	taskStatus     string
	taskName       string
	taskOneType    int
	taskAssignedTo int
	taskProjectId  int
	taskDeptId     int
	taskStartDate  string
	taskEndDate    string
	taskPage       int
	taskLimit      int
	taskOutputJSON bool
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "任务管理",
	Long:  `51PM任务管理相关命令。`,
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "获取任务列表",
	Long:  `根据筛选条件获取51PM项目任务列表。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		localToken, err := loadToken()
		if err != nil {
			return err
		}

		// 构建查询参数
		params := url.Values{}
		if taskStatus != "" {
			params.Set("status", taskStatus)
		}
		if taskName != "" {
			params.Set("name", taskName)
		}
		if taskOneType > 0 {
			params.Set("one_type", strconv.Itoa(taskOneType))
		}
		if taskAssignedTo > 0 {
			params.Set("assigned_to", strconv.Itoa(taskAssignedTo))
		}
		if taskProjectId > 0 {
			params.Set("project_id", strconv.Itoa(taskProjectId))
		}
		if taskDeptId > 0 {
			params.Set("dept_id", strconv.Itoa(taskDeptId))
		}
		if taskStartDate != "" {
			params.Set("start_date", taskStartDate)
		}
		if taskEndDate != "" {
			params.Set("end_date", taskEndDate)
		}
		params.Set("page", strconv.Itoa(taskPage))
		params.Set("limit", strconv.Itoa(taskLimit))

		reqURL := fmt.Sprintf("%s/manage_api/task/get_task_list?%s", baseURL, params.Encode())
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			return fmt.Errorf("构造请求失败: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+localToken.AccessToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("请求失败: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取响应失败: %w", err)
		}

		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("Token已过期，请重新执行 login 命令")
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}

		var result TaskListResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}

		if result.Code != 0 {
			return fmt.Errorf("请求失败[%d]: %s", result.Code, result.Message)
		}

		tasks := result.Data.Data
		if len(tasks) == 0 {
			fmt.Println("暂无任务数据")
			return nil
		}

		// JSON 格式输出
		if taskOutputJSON {
			output, _ := json.MarshalIndent(result.Data, "", "  ")
			fmt.Println(string(output))
			return nil
		}

		// 飞书风格表格输出
		printTaskTable(tasks, result.Data.Total, result.Data.CurrentPage, result.Data.LastPage)
		return nil
	},
}

// 状态颜色映射
func statusStyle(status string) string {
	switch strings.ToLower(status) {
	case "open", "wait":
		return fmt.Sprintf("\033[33m%s\033[0m", status) // 黄色
	case "doing", "started":
		return fmt.Sprintf("\033[36m%s\033[0m", status) // 青色
	case "done", "closed":
		return fmt.Sprintf("\033[32m%s\033[0m", status) // 绿色
	case "cancel", "pause":
		return fmt.Sprintf("\033[90m%s\033[0m", status) // 灰色
	default:
		return status
	}
}

// 进度条
func progressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	color := "\033[32m" // 绿色
	if percent < 30 {
		color = "\033[31m" // 红色
	} else if percent < 70 {
		color = "\033[33m" // 黄色
	}
	return fmt.Sprintf("%s%s\033[0m %3.0f%%", color, bar, percent)
}

// 截断过长字符串
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

func printTaskTable(tasks []*Task, total, currentPage, lastPage int) {
	// 表头
	fmt.Println()
	fmt.Printf("\033[1m\033[37m  %-6s  %-14s  %-20s  %-8s  %-8s  %-22s  %-6s  %-6s\033[0m\n",
		"ID", "项目", "任务名称", "指派给", "状态", "进度", "标准", "已耗")
	fmt.Printf("  %s\n", strings.Repeat("─", 100))

	// 数据行
	for _, t := range tasks {
		fmt.Printf("  %-6d  %-14s  %-20s  %-8s  %-8s  %-22s  %-6.1f  %-6.1f\n",
			t.Id,
			truncate(t.ProjectName, 12),
			truncate(t.Name, 18),
			truncate(t.AssignedTo, 6),
			statusStyle(t.Status),
			progressBar(t.TaskProcess, 10),
			t.StandardHour,
			t.ConsumedHour,
		)
	}

	// 分页信息
	fmt.Printf("  %s\n", strings.Repeat("─", 100))
	fmt.Printf("  \033[90m共 %d 条  第 %d/%d 页\033[0m\n\n", total, currentPage, lastPage)
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskListCmd)

	taskListCmd.Flags().StringVar(&taskStatus, "status", "", "任务状态筛选")
	taskListCmd.Flags().StringVar(&taskName, "name", "", "任务名称模糊查询")
	taskListCmd.Flags().IntVar(&taskOneType, "type", 0, "任务类型")
	taskListCmd.Flags().IntVar(&taskAssignedTo, "assigned-to", 0, "指派给（用户ID）")
	taskListCmd.Flags().IntVar(&taskProjectId, "project-id", 0, "项目ID")
	taskListCmd.Flags().IntVar(&taskDeptId, "dept-id", 0, "部门ID")
	taskListCmd.Flags().StringVar(&taskStartDate, "start-date", "", "开始日期（如 2026-01-01）")
	taskListCmd.Flags().StringVar(&taskEndDate, "end-date", "", "结束日期（如 2026-12-31）")
	taskListCmd.Flags().IntVar(&taskPage, "page", 1, "页码")
	taskListCmd.Flags().IntVar(&taskLimit, "limit", 10, "每页条数")
	taskListCmd.Flags().BoolVar(&taskOutputJSON, "json", false, "以JSON格式输出")
}
