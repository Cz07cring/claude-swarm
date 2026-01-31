package main

import (
	"fmt"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/scheduler"
)

func main() {
	dag := scheduler.NewDAGScheduler()

	fmt.Println("测试 DAG 循环检测...")

	// 创建有循环依赖的任务
	task1 := &models.Task{
		ID:           "task1",
		Description:  "Task 1",
		Status:       models.TaskStatusPending,
		Dependencies: []string{"task2"}, // 依赖 task2
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Priority:     5,
	}

	task2 := &models.Task{
		ID:           "task2",
		Description:  "Task 2",
		Status:       models.TaskStatusPending,
		Dependencies: []string{"task1"}, // 依赖 task1 - 形成循环！
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Priority:     5,
	}

	// 添加 task1
	if err := dag.AddTask(task1); err != nil {
		fmt.Printf("✗ Task1 添加失败: %v\n", err)
		return
	}
	fmt.Println("✓ Task1 添加成功")

	// 尝试添加 task2（应该失败）
	if err := dag.AddTask(task2); err != nil {
		fmt.Printf("✓ 循环依赖检测正常: %v\n", err)
	} else {
		fmt.Println("✗ 循环依赖检测失败 - 应该拒绝循环依赖")
	}

	// 测试正常的依赖链
	fmt.Println("\n测试正常依赖链...")

	dag2 := scheduler.NewDAGScheduler()

	taskA := &models.Task{
		ID:           "taskA",
		Description:  "Task A (no deps)",
		Status:       models.TaskStatusPending,
		Dependencies: []string{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Priority:     1,
	}

	taskB := &models.Task{
		ID:           "taskB",
		Description:  "Task B (depends on A)",
		Status:       models.TaskStatusPending,
		Dependencies: []string{"taskA"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Priority:     2,
	}

	taskC := &models.Task{
		ID:           "taskC",
		Description:  "Task C (depends on B)",
		Status:       models.TaskStatusPending,
		Dependencies: []string{"taskB"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Priority:     3,
	}

	dag2.AddTask(taskA)
	dag2.AddTask(taskB)
	dag2.AddTask(taskC)

	readyTasks := dag2.GetReadyTasks()
	fmt.Printf("✓ 可执行任务数: %d (期望: 1, 应该只有 taskA)\n", len(readyTasks))
	if len(readyTasks) > 0 {
		fmt.Printf("  第一个可执行任务: %s\n", readyTasks[0].ID)
	}

	// 标记 taskA 完成
	taskA.Status = models.TaskStatusCompleted
	dag2.UpdateTask(taskA)

	readyTasks = dag2.GetReadyTasks()
	fmt.Printf("✓ taskA完成后可执行任务数: %d (期望: 1, 应该是 taskB)\n", len(readyTasks))
	if len(readyTasks) > 0 {
		fmt.Printf("  下一个可执行任务: %s\n", readyTasks[0].ID)
	}
}
