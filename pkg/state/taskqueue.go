package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
)

// TaskQueue manages tasks using a JSON file
type TaskQueue struct {
	filePath string
	mu       sync.Mutex
	tasks    map[string]*models.Task
}

type taskFile struct {
	Tasks []*models.Task `json:"tasks"`
}

// NewTaskQueue creates a new task queue
func NewTaskQueue(filePath string) (*TaskQueue, error) {
	// Expand ~ to home directory
	if filePath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		filePath = filepath.Join(home, filePath[2:])
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	tq := &TaskQueue{
		filePath: filePath,
		tasks:    make(map[string]*models.Task),
	}

	// Load existing tasks
	if err := tq.load(); err != nil {
		// If file doesn't exist, create an empty one
		if os.IsNotExist(err) {
			if err := tq.save(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return tq, nil
}

// AddTask adds a new task to the queue
func (tq *TaskQueue) AddTask(task *models.Task) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	// Generate ID if not provided
	if task.ID == "" {
		// Use UnixNano for unique IDs even when tasks are added quickly
		task.ID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}

	// Set timestamps
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	task.UpdatedAt = time.Now()

	// Set status to pending if not set
	if task.Status == "" {
		task.Status = models.TaskStatusPending
	}

	tq.tasks[task.ID] = task
	return tq.save()
}

// ClaimTask claims a pending task for an agent (FIFO)
func (tq *TaskQueue) ClaimTask(agentID string) (*models.Task, error) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	// Reload from file to get latest tasks
	if err := tq.load(); err != nil {
		// If file doesn't exist or can't be read, continue with current tasks
		// This allows the system to work even if file is temporarily unavailable
	}

	// Find the oldest pending task
	var oldestTask *models.Task
	for _, task := range tq.tasks {
		if task.Status == models.TaskStatusPending {
			if oldestTask == nil || task.CreatedAt.Before(oldestTask.CreatedAt) {
				oldestTask = task
			}
		}
	}

	if oldestTask == nil {
		return nil, nil // No pending tasks
	}

	// Claim the task
	oldestTask.Status = models.TaskStatusInProgress
	oldestTask.AssigneeID = agentID
	oldestTask.UpdatedAt = time.Now()

	if err := tq.save(); err != nil {
		return nil, err
	}

	return oldestTask, nil
}

// UpdateTaskStatus updates the status of a task
func (tq *TaskQueue) UpdateTaskStatus(taskID string, status models.TaskStatus) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	task, exists := tq.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	task.Status = status
	task.UpdatedAt = time.Now()

	return tq.save()
}

// GetTask gets a task by ID
func (tq *TaskQueue) GetTask(taskID string) (*models.Task, error) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	task, exists := tq.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// ListTasks returns all tasks
func (tq *TaskQueue) ListTasks() []*models.Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	tasks := make([]*models.Task, 0, len(tq.tasks))
	for _, task := range tq.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// load loads tasks from the JSON file
func (tq *TaskQueue) load() error {
	data, err := os.ReadFile(tq.filePath)
	if err != nil {
		return err
	}

	var tf taskFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return fmt.Errorf("failed to unmarshal tasks: %w", err)
	}

	tq.tasks = make(map[string]*models.Task)
	for _, task := range tf.Tasks {
		tq.tasks[task.ID] = task
	}

	return nil
}

// save saves tasks to the JSON file
func (tq *TaskQueue) save() error {
	tasks := make([]*models.Task, 0, len(tq.tasks))
	for _, task := range tq.tasks {
		tasks = append(tasks, task)
	}

	tf := taskFile{Tasks: tasks}
	data, err := json.MarshalIndent(tf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	if err := os.WriteFile(tq.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tasks file: %w", err)
	}

	return nil
}
