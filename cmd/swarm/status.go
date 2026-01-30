package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/state"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看集群状态",
	Long:  `显示当前 Agent 集群和任务队列的状态`,
	Run: func(cmd *cobra.Command, args []string) {
		runStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus() {
	session := "claude-swarm"
	if sessionName != "" {
		session = sessionName
	}

	fmt.Println("📊 Claude Agent Swarm 状态")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// 检查 tmux 会话
	checkCmd := exec.Command("tmux", "has-session", "-t", session)
	if err := checkCmd.Run(); err != nil {
		fmt.Printf("⚠️  会话 %s 未运行\n", session)
		fmt.Println()
	} else {
		fmt.Printf("✓ 会话: %s (运行中)\n", session)
		fmt.Println()

		// 获取窗格数量
		listCmd := exec.Command("tmux", "list-panes", "-t", session)
		output, err := listCmd.Output()
		if err == nil {
			paneCount := len(strings.Split(strings.TrimSpace(string(output)), "\n"))
			fmt.Printf("  窗格数量: %d\n", paneCount)
		}
		fmt.Println()
	}

	// 显示任务队列
	queuePath := taskQueuePath
	if queuePath == "" {
		queuePath = "~/.claude-swarm/tasks.json"
	}

	tq, err := state.NewTaskQueue(queuePath)
	if err != nil {
		fmt.Printf("❌ 读取任务队列失败: %v\n", err)
		return
	}

	tasks := tq.ListTasks()
	if len(tasks) == 0 {
		fmt.Println("📋 任务队列: 无任务")
	} else {
		fmt.Printf("📋 任务队列: %d 个任务\n\n", len(tasks))

		// 统计任务状态
		statusCount := make(map[models.TaskStatus]int)
		for _, task := range tasks {
			statusCount[task.Status]++
		}

		// 显示统计
		fmt.Println("  状态统计:")
		if count := statusCount[models.TaskStatusPending]; count > 0 {
			fmt.Printf("    待处理: %d\n", count)
		}
		if count := statusCount[models.TaskStatusInProgress]; count > 0 {
			fmt.Printf("    进行中: %d\n", count)
		}
		if count := statusCount[models.TaskStatusCompleted]; count > 0 {
			fmt.Printf("    已完成: %d\n", count)
		}
		if count := statusCount[models.TaskStatusFailed]; count > 0 {
			fmt.Printf("    失败: %d\n", count)
		}
		fmt.Println()

		// 显示最近的任务
		fmt.Println("  最近任务:")
		displayCount := 5
		if len(tasks) < displayCount {
			displayCount = len(tasks)
		}

		for i := 0; i < displayCount; i++ {
			task := tasks[len(tasks)-1-i] // 从最新开始
			statusIcon := getStatusIcon(task.Status)
			fmt.Printf("    %s %s | %s\n", statusIcon, task.ID, truncate(task.Description, 50))
			fmt.Printf("      状态: %s", task.Status)
			if task.AssigneeID != "" {
				fmt.Printf(" | Agent: %s", task.AssigneeID)
			}
			fmt.Printf(" | %s\n", formatTime(task.CreatedAt))
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("💡 提示:")
	fmt.Println("  - 查看实时输出: tmux attach -t", session)
	fmt.Println("  - 添加任务: swarm add-task \"任务描述\"")
	fmt.Println("  - 停止集群: swarm stop")
}

func getStatusIcon(status models.TaskStatus) string {
	switch status {
	case models.TaskStatusPending:
		return "⏳"
	case models.TaskStatusInProgress:
		return "🔄"
	case models.TaskStatusCompleted:
		return "✅"
	case models.TaskStatusFailed:
		return "❌"
	default:
		return "❓"
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "刚刚"
	} else if diff < time.Hour {
		return fmt.Sprintf("%d 分钟前", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%d 小时前", int(diff.Hours()))
	} else {
		return t.Format("2006-01-02 15:04")
	}
}
