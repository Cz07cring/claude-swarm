package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/state"
)

var addTaskCmd = &cobra.Command{
	Use:   "add-task [description]",
	Short: "添加新任务到队列",
	Long:  `添加一个新任务到任务队列，空闲的 Agent 会自动领取并执行`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		description := strings.Join(args, " ")
		runAddTask(description)
	},
}

func init() {
	rootCmd.AddCommand(addTaskCmd)
}

func runAddTask(description string) {
	queuePath := taskQueuePath
	if queuePath == "" {
		queuePath = "~/.claude-swarm/tasks.json"
	}

	// 创建任务队列
	tq, err := state.NewTaskQueue(queuePath)
	if err != nil {
		fmt.Printf("❌ 打开任务队列失败: %v\n", err)
		return
	}

	// 创建任务
	task := &models.Task{
		Description: description,
		Status:      models.TaskStatusPending,
		CreatedAt:   time.Now(),
	}

	// 添加任务
	if err := tq.AddTask(task); err != nil {
		fmt.Printf("❌ 添加任务失败: %v\n", err)
		return
	}

	fmt.Println("✓ 任务已添加")
	fmt.Printf("  ID: %s\n", task.ID)
	fmt.Printf("  描述: %s\n", task.Description)
	fmt.Printf("  状态: %s\n", task.Status)
}
