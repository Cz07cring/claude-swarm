package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/scheduler"
)

// TaskQueue manages tasks using a JSON file
type TaskQueue struct {
	filePath  string
	mu        sync.Mutex              // In-process synchronization
	lockFile  *os.File                // Cross-process file lock
	tasks     map[string]*models.Task
	scheduler *scheduler.DAGScheduler // DAG scheduler for dependency management
}

type taskFile struct {
	Tasks []*models.Task `json:"tasks"`
}

// NewTaskQueue creates a new task queue
func NewTaskQueue(filePath string) (*TaskQueue, error) {
	// Validate and expand file path
	if filePath == "" {
		return nil, fmt.Errorf("filePath cannot be empty")
	}

	// Expand ~ to home directory
	if len(filePath) >= 2 && filePath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		filePath = filepath.Join(home, filePath[2:])
	} else if filePath == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		filePath = home
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create or open lock file
	lockFilePath := filePath + ".lock"
	lockFile, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create lock file: %w", err)
	}

	tq := &TaskQueue{
		filePath:  filePath,
		lockFile:  lockFile,
		tasks:     make(map[string]*models.Task),
		scheduler: scheduler.NewDAGScheduler(),
	}

	// Load existing tasks
	if err := tq.load(); err != nil {
		// If file doesn't exist, create an empty one
		if os.IsNotExist(err) {
			if err := tq.save(); err != nil {
				lockFile.Close()
				return nil, err
			}
		} else {
			lockFile.Close()
			return nil, err
		}
	}

	return tq, nil
}

// Close closes the task queue and releases the file lock
func (tq *TaskQueue) Close() error {
	if tq.lockFile != nil {
		return tq.lockFile.Close()
	}
	return nil
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

	// Set default max retries if not specified
	if task.MaxRetries == 0 {
		task.MaxRetries = 3
	}

	tq.tasks[task.ID] = task

	// Add to DAG scheduler
	if err := tq.scheduler.AddTask(task); err != nil {
		return fmt.Errorf("failed to add task to scheduler: %w", err)
	}

	return tq.save()
}

// ClaimTask claims a pending task for an agent using DAG scheduling
func (tq *TaskQueue) ClaimTask(agentID string) (*models.Task, error) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	// Reload from file to get latest tasks
	if err := tq.load(); err != nil {
		// If file doesn't exist or can't be read, continue with current tasks
		// This allows the system to work even if file is temporarily unavailable
	}

	// Get ready tasks from DAG scheduler (already sorted by priority)
	readyTasks := tq.scheduler.GetReadyTasks()

	if len(readyTasks) == 0 {
		return nil, nil // No ready tasks
	}

	// Select the first ready task (highest priority)
	selectedTask := readyTasks[0]

	// Claim the task
	selectedTask.Status = models.TaskStatusInProgress
	selectedTask.AssigneeID = agentID
	selectedTask.UpdatedAt = time.Now()

	// Update in scheduler
	tq.scheduler.UpdateTask(selectedTask)

	if err := tq.save(); err != nil {
		return nil, err
	}

	return selectedTask, nil
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

// GetReadyTasks returns all tasks that are ready to execute
func (tq *TaskQueue) GetReadyTasks() []*models.Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	return tq.scheduler.GetReadyTasks()
}

// GetBlockedTasks returns tasks that are blocked by dependencies
func (tq *TaskQueue) GetBlockedTasks() []*models.Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	return tq.scheduler.GetBlockedTasks()
}

// GetDependentTasks returns all tasks that depend on the given task
func (tq *TaskQueue) GetDependentTasks(taskID string) []*models.Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	return tq.scheduler.GetDependentTasks(taskID)
}

// load loads tasks from the JSON file
func (tq *TaskQueue) load() error {
	// Acquire shared lock for reading (multiple readers allowed)
	if err := syscall.Flock(int(tq.lockFile.Fd()), syscall.LOCK_SH); err != nil {
		return fmt.Errorf("failed to acquire read lock: %w", err)
	}
	defer syscall.Flock(int(tq.lockFile.Fd()), syscall.LOCK_UN)

	data, err := os.ReadFile(tq.filePath)
	if err != nil {
		return err
	}

	var tf taskFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return fmt.Errorf("failed to unmarshal tasks: %w", err)
	}

	tq.tasks = make(map[string]*models.Task)

	// Recreate scheduler with loaded tasks
	tq.scheduler = scheduler.NewDAGScheduler()

	for _, task := range tf.Tasks {
		tq.tasks[task.ID] = task
		// Add task to scheduler (ignore errors for now as tasks may already exist)
		_ = tq.scheduler.AddTask(task)
	}

	return nil
}

// save saves tasks to the JSON file using atomic write
func (tq *TaskQueue) save() error {
	// Acquire exclusive lock for writing (no other readers or writers)
	if err := syscall.Flock(int(tq.lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed to acquire write lock: %w", err)
	}
	defer syscall.Flock(int(tq.lockFile.Fd()), syscall.LOCK_UN)

	tasks := make([]*models.Task, 0, len(tq.tasks))
	for _, task := range tq.tasks {
		tasks = append(tasks, task)
	}

	tf := taskFile{Tasks: tasks}
	data, err := json.MarshalIndent(tf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	// Atomic write: write to temp file then rename
	tmpFile := tq.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename (overwrites target file atomically)
	if err := os.Rename(tmpFile, tq.filePath); err != nil {
		os.Remove(tmpFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
