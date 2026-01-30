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
