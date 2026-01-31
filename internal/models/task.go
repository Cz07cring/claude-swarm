package models

import "time"

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending     TaskStatus = "pending"
	TaskStatusInProgress  TaskStatus = "in_progress"
	TaskStatusCompleted   TaskStatus = "completed"
	TaskStatusFailed      TaskStatus = "failed"
)

// Task represents a task to be executed by an agent
type Task struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	AssigneeID  string     `json:"assignee_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// DAG scheduling support
	Dependencies []string `json:"dependencies,omitempty"` // IDs of tasks that must complete before this task
	Priority     int      `json:"priority"`               // Priority 1-10 (higher = more important)
	RetryCount   int      `json:"retry_count"`            // Number of times this task has been retried
	MaxRetries   int      `json:"max_retries"`            // Maximum number of retries allowed
	LastError    string   `json:"last_error,omitempty"`   // Last error message if task failed
}

// AgentState represents the state of an agent
type AgentState string

const (
	AgentStateIdle            AgentState = "idle"
	AgentStateWorking         AgentState = "working"
	AgentStateWaitingConfirm  AgentState = "waiting_confirm"
	AgentStateError           AgentState = "error"
	AgentStateStuck           AgentState = "stuck"
)

// AgentStatus represents the current status of an agent
type AgentStatus struct {
	AgentID     string     `json:"agent_id"`
	State       AgentState `json:"state"`
	CurrentTask *Task      `json:"current_task,omitempty"`
	LastUpdate  time.Time  `json:"last_update"`
	Output      string     `json:"output,omitempty"`
}
