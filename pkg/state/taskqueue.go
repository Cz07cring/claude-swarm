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
)

// TaskQueue manages tasks using a JSON file
type TaskQueue struct {
	filePath string
	mu       sync.Mutex      // In-process synchronization
	lockFile *os.File        // Cross-process file lock
	tasks    map[string]*models.Task
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
		filePath: filePath,
		lockFile: lockFile,
		tasks:    make(map[string]*models.Task),
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

	tq.tasks[task.ID] = task
	return tq.save()
}

// ClaimTask claims a pending task for an agent (FIFO)
// This function reloads tasks from disk before claiming to reduce (but not eliminate)
// the risk of concurrent claims by multiple processes.
// Note: This is not fully atomic across processes - for production use,
// consider using a database with proper ACID guarantees.
func (tq *TaskQueue) ClaimTask(agentID string) (*models.Task, error) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	// Reload from file to get latest tasks
	// This reduces the window for race conditions but doesn't eliminate them
	if err := tq.load(); err != nil {
		// Log the error and return - don't proceed with potentially stale data
		return nil, fmt.Errorf("failed to reload task queue before claiming: %w", err)
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
	// Acquire shared lock for reading (multiple readers allowed)
	if err := syscall.Flock(int(tq.lockFile.Fd()), syscall.LOCK_SH); err != nil {
		return fmt.Errorf("failed to acquire read lock: %w", err)
	}
	defer func() {
		if err := syscall.Flock(int(tq.lockFile.Fd()), syscall.LOCK_UN); err != nil {
			// Log unlock failure - this is serious but we can't return the error
			fmt.Fprintf(os.Stderr, "⚠️  Failed to release read lock: %v\n", err)
		}
	}()

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

// save saves tasks to the JSON file using atomic write
func (tq *TaskQueue) save() error {
	// Acquire exclusive lock for writing (no other readers or writers)
	if err := syscall.Flock(int(tq.lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed to acquire write lock: %w", err)
	}
	defer func() {
		if err := syscall.Flock(int(tq.lockFile.Fd()), syscall.LOCK_UN); err != nil {
			// Log unlock failure - this is serious but we can't return the error
			fmt.Fprintf(os.Stderr, "⚠️  Failed to release write lock: %v\n", err)
		}
	}()

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
	
	// Ensure cleanup of temp file in all error cases
	defer func() {
		// Only remove if it still exists (successful rename removes it)
		if _, err := os.Stat(tmpFile); err == nil {
			os.Remove(tmpFile)
		}
	}()
	
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename (overwrites target file atomically)
	if err := os.Rename(tmpFile, tq.filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
