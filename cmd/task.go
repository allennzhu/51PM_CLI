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

// Task 项目任务信息
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

// NotTask 非项目任务信息
type NotTask struct {
	Id           int     `json:"id"`
	ProjectId    int     `json:"project_id"`
	ProjectName  string  `json:"project_name"`
	Name         string  `json:"name"`
	AssignedTo   string  `json:"assigned_to"`
	Status       string  `json:"status"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	PlanHour     float64 `json:"plan_hour"`
	ConsumedHour float64 `json:"consumed_hour"`
	DoneHour     float64 `json:"done_hour"`
	OneType      int     `json:"one_type"`
	DemandName   string  `json:"demand_name"`
	Desc         string  `json:"desc"`
	Remark       string  `json:"remark"`
}

// UnifiedTask 统一任务结构（合并项目任务和非项目任务）
type UnifiedTask struct {
	Id           int     `json:"id"`
	Source       string  `json:"source"` // "project" 或 "non-project"
	ProjectId    int     `json:"project_id"`
	ProjectName  string  `json:"project_name"`
	Name         string  `json:"name"`
	AssignedTo   string  `json:"assigned_to"`
	Status       string  `json:"status"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	PlanHour     float64 `json:"plan_hour"`
	ConsumedHour float64 `json:"consumed_hour"`
	TaskProcess  float64 `json:"task_process"`
	OneType      int     `json:"one_type"`
	Desc         string  `json:"desc"`
	Remark       string  `json:"remark"`
}

// TaskListResponse 项目任务列表响应
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

// NotTaskListResponse 非项目任务列表响应
type NotTaskListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Data        []*NotTask `json:"data"`
		Total       int        `json:"total"`
		PerPage     int        `json:"per_page"`
		CurrentPage int        `json:"current_page"`
		LastPage    int        `json:"last_page"`
	} `json:"data"`
}

// UnifiedTaskListResult 统一任务列表结果
type UnifiedTaskListResult struct {
	Data            []*UnifiedTask `json:"data"`
	ProjectTotal    int            `json:"project_total"`
	NonProjectTotal int            `json:"non_project_total"`
	Total           int            `json:"total"`
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
	Long:  `根据筛选条件获取51PM项目任务和非项目任务列表（自动合并）。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		localToken, err := loadToken()
		if err != nil {
			return err
		}

		// 构建通用查询参数
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
		params.Set("page", "1")
		params.Set("limit", strconv.Itoa(taskLimit))

		// 查询项目任务
		projectTasks, projectTotal, err := fetchTaskList(localToken.AccessToken, params)
		if err != nil {
			return fmt.Errorf("查询项目任务失败: %w", err)
		}

		// 构建非项目任务查询参数
		notParams := url.Values{}
		if taskStatus != "" {
			notParams.Set("status", taskStatus)
		}
		if taskName != "" {
			notParams.Set("name", taskName)
		}
		if taskOneType > 0 {
			notParams.Set("one_type", strconv.Itoa(taskOneType))
		}
		if taskAssignedTo > 0 {
			notParams.Set("assigned_to", strconv.Itoa(taskAssignedTo))
		}
		if taskProjectId > 0 {
			notParams.Set("project_id", strconv.Itoa(taskProjectId))
		}
		if taskDeptId > 0 {
			notParams.Set("dept_id", strconv.Itoa(taskDeptId))
		}
		if taskStartDate != "" {
			notParams.Set("start_date", taskStartDate)
		}
		if taskEndDate != "" {
			notParams.Set("end_date", taskEndDate)
		}
		notParams.Set("date_type", "task")
		notParams.Set("page", "1")
		notParams.Set("limit", strconv.Itoa(taskLimit))

		// 查询非项目任务
		notProjectTasks, notProjectTotal, err := fetchNotTaskList(localToken.AccessToken, notParams)
		if err != nil {
			return fmt.Errorf("查询非项目任务失败: %w", err)
		}

		// 合并任务
		var unified []*UnifiedTask
		for _, t := range projectTasks {
			unified = append(unified, &UnifiedTask{
				Id:           t.Id,
				Source:       "project",
				ProjectId:    t.ProjectId,
				ProjectName:  t.ProjectName,
				Name:         t.Name,
				AssignedTo:   t.AssignedTo,
				Status:       t.Status,
				StartDate:    t.StartDate,
				EndDate:      t.EndDate,
				PlanHour:     t.StandardHour,
				ConsumedHour: t.ConsumedHour,
				TaskProcess:  t.TaskProcess,
				OneType:      t.OneType,
				Desc:         t.Desc,
				Remark:       t.Remark,
			})
		}
		for _, t := range notProjectTasks {
			consumed := t.ConsumedHour
			if consumed == 0 {
				consumed = t.DoneHour
			}
			unified = append(unified, &UnifiedTask{
				Id:           t.Id,
				Source:       "non-project",
				ProjectId:    t.ProjectId,
				ProjectName:  t.ProjectName,
				Name:         t.Name,
				AssignedTo:   t.AssignedTo,
				Status:       t.Status,
				StartDate:    t.StartDate,
				EndDate:      t.EndDate,
				PlanHour:     t.PlanHour,
				ConsumedHour: consumed,
				OneType:      t.OneType,
				Desc:         t.Desc,
				Remark:       t.Remark,
			})
		}

		totalCount := projectTotal + notProjectTotal
		if len(unified) == 0 {
			fmt.Println("暂无任务数据")
			return nil
		}

		// JSON 格式输出
		if taskOutputJSON {
			result := UnifiedTaskListResult{
				Data:            unified,
				ProjectTotal:    projectTotal,
				NonProjectTotal: notProjectTotal,
				Total:           totalCount,
			}
			output, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(output))
			return nil
		}

		// 表格输出
		printUnifiedTaskTable(unified, projectTotal, notProjectTotal)
		return nil
	},
}

func fetchTaskList(token string, params url.Values) ([]*Task, int, error) {
	reqURL := fmt.Sprintf("%s/manage_api/task/get_task_list?%s", baseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, 0, fmt.Errorf("Token已过期，请重新执行 login 命令")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result TaskListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, 0, fmt.Errorf("解析响应失败: %w", err)
	}
	if result.Code != 0 {
		return nil, 0, fmt.Errorf("请求失败[%d]: %s", result.Code, result.Message)
	}
	return result.Data.Data, result.Data.Total, nil
}

func fetchNotTaskList(token string, params url.Values) ([]*NotTask, int, error) {
	reqURL := fmt.Sprintf("%s/manage_api/task/get_not_task_list?%s", baseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, 0, fmt.Errorf("Token已过期，请重新执行 login 命令")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result NotTaskListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, 0, fmt.Errorf("解析响应失败: %w", err)
	}
	if result.Code != 0 {
		return nil, 0, fmt.Errorf("请求失败[%d]: %s", result.Code, result.Message)
	}
	return result.Data.Data, result.Data.Total, nil
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

func printUnifiedTaskTable(tasks []*UnifiedTask, projectTotal, notProjectTotal int) {
	// 表头
	fmt.Println()
	fmt.Printf("\033[1m\033[37m  %-6s  %-6s  %-14s  %-20s  %-8s  %-8s  %-22s  %-6s  %-6s\033[0m\n",
		"ID", "类型", "项目", "任务名称", "指派给", "状态", "进度", "计划", "已耗")
	fmt.Printf("  %s\n", strings.Repeat("─", 110))

	// 数据行
	for _, t := range tasks {
		sourceTag := "\033[34m项目\033[0m"
		if t.Source == "non-project" {
			sourceTag = "\033[35m非项目\033[0m"
		}
		fmt.Printf("  %-6d  %-6s  %-14s  %-20s  %-8s  %-8s  %-22s  %-6.1f  %-6.1f\n",
			t.Id,
			sourceTag,
			truncate(t.ProjectName, 12),
			truncate(t.Name, 18),
			truncate(t.AssignedTo, 6),
			statusStyle(t.Status),
			progressBar(t.TaskProcess, 10),
			t.PlanHour,
			t.ConsumedHour,
		)
	}

	// 汇总信息
	fmt.Printf("  %s\n", strings.Repeat("─", 110))
	fmt.Printf("  \033[90m项目任务 %d 条 + 非项目任务 %d 条 = 共 %d 条\033[0m\n\n",
		projectTotal, notProjectTotal, projectTotal+notProjectTotal)
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
