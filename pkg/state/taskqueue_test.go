package state

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
)

func TestTaskQueue_UpdateTask(t *testing.T) {
	// 创建临时任务队列
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "tasks.json")

	tq, err := NewTaskQueue(taskFile)
	if err != nil {
		t.Fatalf("Failed to create task queue: %v", err)
	}
	defer tq.Close()

	// 添加初始任务
	task := &models.Task{
		ID:          "test-1",
		Description: "Test task",
		Status:      models.TaskStatusPending,
		Priority:    5,
		CreatedAt:   time.Now(),
	}

	if err := tq.AddTask(task); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// 更新任务
	task.Priority = 10
	task.Dependencies = []string{"dep-1"}

	if err := tq.UpdateTask(task); err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// 验证更新
	updated, err := tq.GetTask("test-1")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if updated.Priority != 10 {
		t.Errorf("Expected priority 10, got %d", updated.Priority)
	}

	if len(updated.Dependencies) != 1 || updated.Dependencies[0] != "dep-1" {
		t.Errorf("Expected dependencies [dep-1], got %v", updated.Dependencies)
	}
}

func TestTaskQueue_RemoveTask(t *testing.T) {
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "tasks.json")

	tq, err := NewTaskQueue(taskFile)
	if err != nil {
		t.Fatalf("Failed to create task queue: %v", err)
	}
	defer tq.Close()

	// 添加任务
	task := &models.Task{
		ID:          "test-1",
		Description: "Test task",
		Status:      models.TaskStatusPending,
	}

	if err := tq.AddTask(task); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// 移除任务
	if err := tq.RemoveTask("test-1"); err != nil {
		t.Fatalf("Failed to remove task: %v", err)
	}

	// 验证移除
	_, err = tq.GetTask("test-1")
	if err == nil {
		t.Error("Expected task to be removed, but it still exists")
	}
}

func TestTaskQueue_ClearCompleted(t *testing.T) {
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "tasks.json")

	tq, err := NewTaskQueue(taskFile)
	if err != nil {
		t.Fatalf("Failed to create task queue: %v", err)
	}
	defer tq.Close()

	// 添加多个任务
	tasks := []*models.Task{
		{ID: "completed-1", Description: "Done", Status: models.TaskStatusCompleted},
		{ID: "completed-2", Description: "Done", Status: models.TaskStatusCompleted},
		{ID: "pending-1", Description: "Pending", Status: models.TaskStatusPending},
	}

	for _, task := range tasks {
		if err := tq.AddTask(task); err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
	}

	// 清理已完成的任务
	count, err := tq.ClearCompleted()
	if err != nil {
		t.Fatalf("Failed to clear completed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected to clear 2 tasks, cleared %d", count)
	}

	// 验证
	allTasks := tq.ListTasks()
	if len(allTasks) != 1 {
		t.Errorf("Expected 1 task remaining, got %d", len(allTasks))
	}

	if allTasks[0].ID != "pending-1" {
		t.Errorf("Expected pending-1 to remain, got %s", allTasks[0].ID)
	}
}

func TestTaskQueue_ClearFailed(t *testing.T) {
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "tasks.json")

	tq, err := NewTaskQueue(taskFile)
	if err != nil {
		t.Fatalf("Failed to create task queue: %v", err)
	}
	defer tq.Close()

	// 添加多个任务
	tasks := []*models.Task{
		{ID: "failed-1", Description: "Failed", Status: models.TaskStatusFailed},
		{ID: "pending-1", Description: "Pending", Status: models.TaskStatusPending},
	}

	for _, task := range tasks {
		if err := tq.AddTask(task); err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
	}

	// 清理失败的任务
	count, err := tq.ClearFailed()
	if err != nil {
		t.Fatalf("Failed to clear failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected to clear 1 task, cleared %d", count)
	}

	// 验证
	allTasks := tq.ListTasks()
	if len(allTasks) != 1 {
		t.Errorf("Expected 1 task remaining, got %d", len(allTasks))
	}
}

func TestTaskQueue_ClearAll(t *testing.T) {
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "tasks.json")

	tq, err := NewTaskQueue(taskFile)
	if err != nil {
		t.Fatalf("Failed to create task queue: %v", err)
	}
	defer tq.Close()

	// 添加多个任务
	for i := 0; i < 5; i++ {
		task := &models.Task{
			ID:          string(rune('A' + i)),
			Description: "Test",
			Status:      models.TaskStatusPending,
		}
		if err := tq.AddTask(task); err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
	}

	// 清理所有任务
	count, err := tq.ClearAll()
	if err != nil {
		t.Fatalf("Failed to clear all: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected to clear 5 tasks, cleared %d", count)
	}

	// 验证
	allTasks := tq.ListTasks()
	if len(allTasks) != 0 {
		t.Errorf("Expected 0 tasks remaining, got %d", len(allTasks))
	}
}

func TestTaskQueue_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "tasks.json")

	// 创建任务队列并添加任务
	tq1, err := NewTaskQueue(taskFile)
	if err != nil {
		t.Fatalf("Failed to create task queue: %v", err)
	}

	task := &models.Task{
		ID:          "persist-1",
		Description: "Persistent task",
		Status:      models.TaskStatusPending,
		Priority:    7,
		Dependencies: []string{"dep-1"},
	}

	if err := tq1.AddTask(task); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	tq1.Close()

	// 重新打开并验证
	tq2, err := NewTaskQueue(taskFile)
	if err != nil {
		t.Fatalf("Failed to reopen task queue: %v", err)
	}
	defer tq2.Close()

	loaded, err := tq2.GetTask("persist-1")
	if err != nil {
		t.Fatalf("Failed to get persisted task: %v", err)
	}

	if loaded.Description != "Persistent task" {
		t.Errorf("Expected description 'Persistent task', got %s", loaded.Description)
	}

	if loaded.Priority != 7 {
		t.Errorf("Expected priority 7, got %d", loaded.Priority)
	}

	if len(loaded.Dependencies) != 1 || loaded.Dependencies[0] != "dep-1" {
		t.Errorf("Expected dependencies [dep-1], got %v", loaded.Dependencies)
	}
}

func TestTaskQueue_UpdateTaskNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "tasks.json")

	tq, err := NewTaskQueue(taskFile)
	if err != nil {
		t.Fatalf("Failed to create task queue: %v", err)
	}
	defer tq.Close()

	task := &models.Task{
		ID:          "nonexistent",
		Description: "Test",
	}

	err = tq.UpdateTask(task)
	if err == nil {
		t.Error("Expected error when updating nonexistent task, got nil")
	}
}

func TestTaskQueue_RemoveTaskNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "tasks.json")

	tq, err := NewTaskQueue(taskFile)
	if err != nil {
		t.Fatalf("Failed to create task queue: %v", err)
	}
	defer tq.Close()

	err = tq.RemoveTask("nonexistent")
	if err == nil {
		t.Error("Expected error when removing nonexistent task, got nil")
	}
}
