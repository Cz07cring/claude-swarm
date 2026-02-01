package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/state"
)

var addTaskCmd = &cobra.Command{
	Use:   "add-task [description]",
	Short: "快速添加单个任务到队列",
	Long: `快速添加单个任务到任务队列。支持设置优先级、依赖关系和重试次数。

示例:
  # 基础使用
  swarm add-task "创建 README.md 文件"

  # 高级用法（支持 V2 新特性）
  swarm add-task "编写单元测试" \
    --priority 8 \
    --dependencies task-1,task-2 \
    --max-retries 5 \
    --id custom-task-id`,
	Args: cobra.MinimumNArgs(1),
	Run:  runAddTask,
}

var (
	taskPriority     int
	taskDependencies []string
	taskMaxRetries   int
	taskID           string
)

func init() {
	rootCmd.AddCommand(addTaskCmd)

	addTaskCmd.Flags().IntVarP(&taskPriority, "priority", "p", 5, "任务优先级 (1-10)")
	addTaskCmd.Flags().StringSliceVarP(&taskDependencies, "dependencies", "d", nil, "依赖的任务ID（逗号分隔）")
	addTaskCmd.Flags().IntVar(&taskMaxRetries, "max-retries", 3, "最大重试次数")
	addTaskCmd.Flags().StringVar(&taskID, "id", "", "自定义任务ID（留空自动生成）")
	addTaskCmd.Flags().StringVar(&taskQueuePath, "queue", "~/.claude-swarm/tasks.json", "任务队列文件路径")
}

func runAddTask(cmd *cobra.Command, args []string) {
	// 1. 验证输入
	description := strings.Join(args, " ")
	if err := validateTaskDescription(description); err != nil {
		log.Fatalf("❌ 无效的任务描述: %v", err)
	}

	// 验证优先级
	if taskPriority < 1 || taskPriority > 10 {
		log.Fatalf("❌ 优先级必须在 1-10 之间，当前值: %d", taskPriority)
	}

	// 验证重试次数
	if taskMaxRetries < 0 {
		log.Fatalf("❌ 最大重试次数不能为负数，当前值: %d", taskMaxRetries)
	}

	// 2. 初始化任务队列
	taskQueue, err := state.NewTaskQueue(expandPath(taskQueuePath))
	if err != nil {
		log.Fatalf("❌ 无法打开任务队列: %v", err)
	}
	defer taskQueue.Close()

	// 3. 创建任务
	task := &models.Task{
		ID:           generateTaskID(taskID),
		Description:  description,
		Status:       models.TaskStatusPending,
		Priority:     taskPriority,
		Dependencies: taskDependencies,
		MaxRetries:   taskMaxRetries,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 4. 验证依赖是否存在
	if len(taskDependencies) > 0 {
		if err := validateDependencies(taskQueue, taskDependencies); err != nil {
			log.Fatalf("❌ 依赖验证失败: %v", err)
		}
	}

	// 5. 添加到队列
	if err := taskQueue.AddTask(task); err != nil {
		log.Fatalf("❌ 添加任务失败: %v", err)
	}

	// 6. 输出成功信息
	fmt.Println("✅ 任务已添加")
	fmt.Printf("   ID: %s\n", task.ID)
	fmt.Printf("   描述: %s\n", description)
	fmt.Printf("   优先级: %d\n", taskPriority)
	if len(taskDependencies) > 0 {
		fmt.Printf("   依赖: %v\n", taskDependencies)
	}
	fmt.Printf("   最大重试: %d\n", taskMaxRetries)
}

// validateTaskDescription validates the task description
func validateTaskDescription(description string) error {
	description = strings.TrimSpace(description)

	if description == "" {
		return fmt.Errorf("任务描述不能为空")
	}

	if len(description) > 500 {
		return fmt.Errorf("任务描述过长（最多 500 字符），当前长度: %d", len(description))
	}

	return nil
}

// generateTaskID generates a task ID (custom or auto-generated)
func generateTaskID(customID string) string {
	if customID != "" {
		return customID
	}
	// Use UnixNano for unique IDs even when tasks are added quickly
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	} else if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	return path
}

// validateDependencies validates that all dependencies exist
func validateDependencies(taskQueue *state.TaskQueue, dependencies []string) error {
	tasks := taskQueue.ListTasks()
	taskMap := make(map[string]bool)
	for _, task := range tasks {
		taskMap[task.ID] = true
	}

	for _, depID := range dependencies {
		if !taskMap[depID] {
			return fmt.Errorf("依赖的任务不存在: %s", depID)
		}
	}

	return nil
}
