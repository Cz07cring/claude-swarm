package controller

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/executor"
	"github.com/yourusername/claude-swarm/pkg/git"
)

// Agent represents a single Claude Code agent
type Agent struct {
	ID         string
	Executor   *executor.ClaudeExecutor
	Status     *models.AgentStatus
	Worktree   *git.Worktree
	WorkingDir string
	mu         sync.Mutex
	version    uint64 // State version number for optimistic locking

	// Task channel for receiving tasks
	taskChan chan *models.Task

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// NewAgent creates a new agent
func NewAgent(id string, worktree *git.Worktree, workingDir string) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		ID:         id,
		Executor:   executor.NewClaudeExecutor(workingDir),
		Worktree:   worktree,
		WorkingDir: workingDir,
		taskChan:   make(chan *models.Task, 5), // Buffer for 5 tasks
		Status: &models.AgentStatus{
			AgentID:    id,
			State:      models.AgentStateIdle,
			LastUpdate: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// ExecuteTask executes a task using Claude Code CLI
func (a *Agent) ExecuteTask(task *models.Task) error {
	a.mu.Lock()
	a.Status.State = models.AgentStateWorking
	a.Status.CurrentTask = task
	a.version++
	a.mu.Unlock()

	log.Printf("🚀 Agent %s starting task: %s", a.ID, task.Description)

	// Execute with timeout
	taskCtx, cancel := context.WithTimeout(a.ctx, 10*time.Minute)
	defer cancel()

	err := a.Executor.ExecuteTask(taskCtx, task)

	a.mu.Lock()
	if err != nil {
		a.Status.State = models.AgentStateError
		log.Printf("❌ Agent %s task failed: %v", a.ID, err)
	} else {
		a.Status.State = models.AgentStateIdle
		a.Status.CurrentTask = nil
		log.Printf("✅ Agent %s task completed", a.ID)
	}
	a.version++
	a.mu.Unlock()

	return err
}

// GetStatus returns a copy of the agent status
func (a *Agent) GetStatus() *models.AgentStatus {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Deep copy
	status := &models.AgentStatus{
		AgentID:    a.Status.AgentID,
		State:      a.Status.State,
		LastUpdate: a.Status.LastUpdate,
		Output:     a.Status.Output,
	}

	if a.Status.CurrentTask != nil {
		taskCopy := *a.Status.CurrentTask
		status.CurrentTask = &taskCopy
	}

	return status
}

// Stop stops the agent
func (a *Agent) Stop() {
	a.cancel()
}

// IsIdle returns true if the agent is idle and has no task
func (a *Agent) IsIdle() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.Status.State == models.AgentStateIdle && a.Status.CurrentTask == nil
}
