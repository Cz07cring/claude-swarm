package scheduler

import (
	"fmt"
	"sort"
	"sync"

	"github.com/yourusername/claude-swarm/internal/models"
)

// DAGScheduler manages task dependencies and determines which tasks are ready to execute
type DAGScheduler struct {
	tasks map[string]*models.Task  // All tasks by ID
	graph map[string][]string      // Reverse dependency graph: taskID -> dependent task IDs
	mu    sync.RWMutex             // Protect concurrent access
}

// NewDAGScheduler creates a new DAG scheduler
func NewDAGScheduler() *DAGScheduler {
	return &DAGScheduler{
		tasks: make(map[string]*models.Task),
		graph: make(map[string][]string),
	}
}

// AddTask adds a task to the scheduler
func (ds *DAGScheduler) AddTask(task *models.Task) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Check for cyclic dependencies
	if ds.hasCyclicDependency(task) {
		return fmt.Errorf("cyclic dependency detected for task %s", task.ID)
	}

	// Add task
	ds.tasks[task.ID] = task

	// Build reverse dependency graph
	for _, depID := range task.Dependencies {
		ds.graph[depID] = append(ds.graph[depID], task.ID)
	}

	return nil
}

// RemoveTask removes a task from the scheduler
func (ds *DAGScheduler) RemoveTask(taskID string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	task, exists := ds.tasks[taskID]
	if !exists {
		return
	}

	// Remove from reverse graph
	for _, depID := range task.Dependencies {
		deps := ds.graph[depID]
		for i, id := range deps {
			if id == taskID {
				ds.graph[depID] = append(deps[:i], deps[i+1:]...)
				break
			}
		}
	}

	// Remove from tasks
	delete(ds.tasks, taskID)
	delete(ds.graph, taskID)
}

// UpdateTask updates a task in the scheduler
func (ds *DAGScheduler) UpdateTask(task *models.Task) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.tasks[task.ID] = task
}

// GetReadyTasks returns all tasks that are ready to execute
// Ready tasks are those that:
// 1. Have status = pending
// 2. Have no assignee
// 3. All dependencies are completed
func (ds *DAGScheduler) GetReadyTasks() []*models.Task {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var readyTasks []*models.Task

	for _, task := range ds.tasks {
		// Check if task is pending and unassigned
		if task.Status != models.TaskStatusPending || task.AssigneeID != "" {
			continue
		}

		// Check if all dependencies are satisfied
		if ds.areDependenciesSatisfiedUnlocked(task) {
			readyTasks = append(readyTasks, task)
		}
	}

	// Sort by priority (higher priority first)
	sort.Slice(readyTasks, func(i, j int) bool {
		if readyTasks[i].Priority != readyTasks[j].Priority {
			return readyTasks[i].Priority > readyTasks[j].Priority
		}
		// If same priority, sort by creation time (earlier first)
		return readyTasks[i].CreatedAt.Before(readyTasks[j].CreatedAt)
	})

	return readyTasks
}

// GetBlockedTasks returns tasks that are blocked by dependencies
func (ds *DAGScheduler) GetBlockedTasks() []*models.Task {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var blockedTasks []*models.Task

	for _, task := range ds.tasks {
		if task.Status == models.TaskStatusPending &&
			task.AssigneeID == "" &&
			!ds.areDependenciesSatisfiedUnlocked(task) {
			blockedTasks = append(blockedTasks, task)
		}
	}

	return blockedTasks
}

// areDependenciesSatisfied checks if all dependencies of a task are completed
func (ds *DAGScheduler) AreDependenciesSatisfied(task *models.Task) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.areDependenciesSatisfiedUnlocked(task)
}

// areDependenciesSatisfiedUnlocked checks dependencies without acquiring lock
func (ds *DAGScheduler) areDependenciesSatisfiedUnlocked(task *models.Task) bool {
	for _, depID := range task.Dependencies {
		depTask, exists := ds.tasks[depID]
		if !exists {
			// Dependency doesn't exist - consider it unsatisfied
			return false
		}
		if depTask.Status != models.TaskStatusCompleted {
			// Dependency not completed
			return false
		}
	}
	return true
}

// hasCyclicDependency checks if adding this task would create a cycle
func (ds *DAGScheduler) hasCyclicDependency(newTask *models.Task) bool {
	// Use DFS to detect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(taskID string) bool
	hasCycle = func(taskID string) bool {
		visited[taskID] = true
		recStack[taskID] = true

		// Get dependencies of this task
		var deps []string
		if taskID == newTask.ID {
			deps = newTask.Dependencies
		} else if task, exists := ds.tasks[taskID]; exists {
			deps = task.Dependencies
		}

		// Check all dependencies
		for _, depID := range deps {
			if !visited[depID] {
				if hasCycle(depID) {
					return true
				}
			} else if recStack[depID] {
				// Back edge found - cycle detected
				return true
			}
		}

		recStack[taskID] = false
		return false
	}

	return hasCycle(newTask.ID)
}

// GetTaskCount returns the total number of tasks
func (ds *DAGScheduler) GetTaskCount() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return len(ds.tasks)
}

// GetTask returns a specific task by ID
func (ds *DAGScheduler) GetTask(taskID string) (*models.Task, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	task, exists := ds.tasks[taskID]
	return task, exists
}

// GetDependentTasks returns all tasks that depend on the given task
func (ds *DAGScheduler) GetDependentTasks(taskID string) []*models.Task {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var dependents []*models.Task
	for _, depID := range ds.graph[taskID] {
		if task, exists := ds.tasks[depID]; exists {
			dependents = append(dependents, task)
		}
	}

	return dependents
}
