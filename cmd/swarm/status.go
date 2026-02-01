package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/state"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看任务队列状态",
	Long: `查看任务队列的当前状态，包括统计信息和任务详情。

示例:
  # 基础状态查看
  swarm status

  # 详细模式
  swarm status --verbose

  # 仅显示特定状态
  swarm status --filter pending
  swarm status --filter failed`,
	Run: runStatus,
}

var (
	statusVerbose bool
	statusFilter  string
)

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().BoolVarP(&statusVerbose, "verbose", "v", false, "显示详细信息")
	statusCmd.Flags().StringVarP(&statusFilter, "filter", "f", "", "过滤任务状态 (pending/in_progress/completed/failed)")
	statusCmd.Flags().StringVar(&taskQueuePath, "queue", "~/.claude-swarm/tasks.json", "任务队列文件路径")
}

func runStatus(cmd *cobra.Command, args []string) {
	// 1. 初始化任务队列
	taskQueue, err := state.NewTaskQueue(expandPath(taskQueuePath))
	if err != nil {
		log.Fatalf("❌ 无法打开任务队列: %v", err)
	}
	defer taskQueue.Close()

	// 2. 获取任务列表
	tasks := taskQueue.ListTasks()

	if len(tasks) == 0 {
		fmt.Println("📭 任务队列为空")
		fmt.Println()
		fmt.Println("提示：使用 'swarm add-task' 添加新任务，或使用 'swarm orchestrate' 从需求生成任务")
		return
	}

	// 3. 统计
	stats := calculateStats(tasks)

	// 4. 打印标题
	fmt.Println("📊 Claude Swarm V2 任务状态")
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println()

	// 5. 打印统计
	printStats(stats)

	// 6. 打印任务详情
	printTaskList(tasks, statusFilter, statusVerbose, taskQueue)
}

// TaskStats holds task statistics
type TaskStats struct {
	Completed  int
	InProgress int
	Pending    int
	Failed     int
}

// calculateStats calculates task statistics
func calculateStats(tasks []*models.Task) *TaskStats {
	stats := &TaskStats{}

	for _, task := range tasks {
		switch task.Status {
		case models.TaskStatusCompleted:
			stats.Completed++
		case models.TaskStatusInProgress:
			stats.InProgress++
		case models.TaskStatusPending:
			stats.Pending++
		case models.TaskStatusFailed:
			stats.Failed++
		}
	}

	return stats
}

// printStats prints task statistics
func printStats(stats *TaskStats) {
	total := stats.Completed + stats.InProgress + stats.Pending + stats.Failed
	percentage := 0
	if total > 0 {
		percentage = (stats.Completed * 100) / total
	}

	fmt.Println("📈 统计:")
	fmt.Printf("  ✅ 已完成: %d / %d (%d%%)\n", stats.Completed, total, percentage)
	fmt.Printf("  🔄 进行中: %d\n", stats.InProgress)
	fmt.Printf("  ⏳ 待执行: %d\n", stats.Pending)
	fmt.Printf("  ❌ 失败: %d\n", stats.Failed)
	fmt.Println()

	// 进度条
	printProgressBar(stats.Completed, total, 40)
	fmt.Println()
}

// printProgressBar prints a progress bar
func printProgressBar(completed, total, width int) {
	if total == 0 {
		return
	}

	filled := (completed * width) / total
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	fmt.Printf("  [%s] %d%%\n", bar, (completed*100)/total)
}

// printTaskList prints the task list
func printTaskList(tasks []*models.Task, filter string, verbose bool, taskQueue *state.TaskQueue) {
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println("📋 任务详情:")
	fmt.Println()

	// 按优先级排序（高优先级在前）
	sort.Slice(tasks, func(i, j int) bool {
		// 首先按状态排序：in_progress > pending > failed > completed
		statusPriority := map[models.TaskStatus]int{
			models.TaskStatusInProgress: 4,
			models.TaskStatusPending:    3,
			models.TaskStatusFailed:     2,
			models.TaskStatusCompleted:  1,
		}
		if statusPriority[tasks[i].Status] != statusPriority[tasks[j].Status] {
			return statusPriority[tasks[i].Status] > statusPriority[tasks[j].Status]
		}
		// 然后按优先级排序
		return tasks[i].Priority > tasks[j].Priority
	})

	for _, task := range tasks {
		// 过滤
		if filter != "" && string(task.Status) != filter {
			continue
		}

		// 状态图标
		icon := getStatusIcon(task.Status)

		// 基本信息
		fmt.Printf("[%s] %s (优先级: %d", task.Status, task.ID, task.Priority)
		if task.AssigneeID != "" {
			fmt.Printf(", 分配给: %s", task.AssigneeID)
		}
		fmt.Println(")")
		fmt.Printf("  %s %s\n", icon, task.Description)

		// 依赖信息
		if len(task.Dependencies) > 0 {
			satisfied := areDependenciesSatisfied(task, tasks)
			if satisfied {
				fmt.Printf("  依赖: %v ✓\n", task.Dependencies)
			} else {
				fmt.Printf("  依赖: %v ⚠️  未满足\n", task.Dependencies)
			}
		}

		// 失败信息
		if task.Status == models.TaskStatusFailed {
			if task.LastError != "" {
				fmt.Printf("  错误: %s\n", task.LastError)
			}
			if task.RetryCount > 0 {
				fmt.Printf("  重试: %d/%d\n", task.RetryCount, task.MaxRetries)
			}
		}

		// 详细模式
		if verbose {
			fmt.Printf("  创建时间: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  更新时间: %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05"))
			if task.RetryCount > 0 && task.Status != models.TaskStatusFailed {
				fmt.Printf("  重试次数: %d/%d\n", task.RetryCount, task.MaxRetries)
			}
		}

		fmt.Println()
	}

	// 显示被阻塞的任务统计
	if filter == "" || filter == "pending" {
		blockedTasks := taskQueue.GetBlockedTasks()
		if len(blockedTasks) > 0 {
			fmt.Println(strings.Repeat("━", 60))
			fmt.Printf("⚠️  有 %d 个任务因依赖未满足而被阻塞\n", len(blockedTasks))
			fmt.Println()
		}
	}
}

// getStatusIcon returns an icon for the task status
func getStatusIcon(status models.TaskStatus) string {
	switch status {
	case models.TaskStatusCompleted:
		return "✅"
	case models.TaskStatusInProgress:
		return "🔄"
	case models.TaskStatusPending:
		return "⏳"
	case models.TaskStatusFailed:
		return "❌"
	default:
		return "❓"
	}
}

// areDependenciesSatisfied checks if all dependencies are satisfied
func areDependenciesSatisfied(task *models.Task, allTasks []*models.Task) bool {
	if len(task.Dependencies) == 0 {
		return true
	}

	taskMap := make(map[string]*models.Task)
	for _, t := range allTasks {
		taskMap[t.ID] = t
	}

	for _, depID := range task.Dependencies {
		depTask, exists := taskMap[depID]
		if !exists {
			return false // 依赖的任务不存在
		}
		if depTask.Status != models.TaskStatusCompleted {
			return false // 依赖的任务未完成
		}
	}

	return true
}
