package controller

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/executor"
	"github.com/yourusername/claude-swarm/pkg/git"
	"github.com/yourusername/claude-swarm/pkg/retry"
	"github.com/yourusername/claude-swarm/pkg/state"
)

// CoordinatorV2 manages the swarm using direct Claude CLI execution
type CoordinatorV2 struct {
	agents           []*Agent
	taskQueue        *state.TaskQueue
	worktreeManager  *git.WorktreeManager
	retryManager     *retry.RetryManager
	repoPath         string

	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
}

// NewCoordinatorV2 creates a new coordinator using Claude CLI execution
func NewCoordinatorV2(repoPath string, taskQueuePath string, numAgents int) (*CoordinatorV2, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize task queue
	taskQueue, err := state.NewTaskQueue(taskQueuePath)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create task queue: %w", err)
	}

	// Initialize worktree manager
	worktreeManager, err := git.NewWorktreeManager(git.WorktreeConfig{
		BaseRepoPath:    repoPath,
		WorktreeRootDir: ".worktrees",
		BaseBranch:      "main",
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create worktree manager: %w", err)
	}

	// Initialize retry manager
	retryManager := retry.NewRetryManager(retry.RetryConfig{
		MaxRetries:    3,
		InitialDelay:  5 * time.Second,
		MaxDelay:      5 * time.Minute,
		BackoffFactor: 2.0,
	})

	c := &CoordinatorV2{
		agents:          make([]*Agent, 0, numAgents),
		taskQueue:       taskQueue,
		worktreeManager: worktreeManager,
		retryManager:    retryManager,
		repoPath:        repoPath,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Create agents
	for i := 0; i < numAgents; i++ {
		agentID := fmt.Sprintf("agent-%d", i)

		// Create worktree for agent
		worktree, err := worktreeManager.CreateWorktree(fmt.Sprintf("%d", i))
		if err != nil {
			c.Cleanup()
			return nil, fmt.Errorf("failed to create worktree for %s: %w", agentID, err)
		}

		// Create agent
		agent := NewAgent(agentID, worktree, worktree.Path)
		c.agents = append(c.agents, agent)

		log.Printf("✓ Created agent: %s (worktree: %s)", agentID, worktree.Path)
	}

	return c, nil
}

// Start starts the coordinator
func (c *CoordinatorV2) Start() error {
	log.Println("🚀 Starting Claude Swarm Coordinator V2")

	// Start scheduler
	c.wg.Add(1)
	go c.runScheduler()

	// Start agent workers
	for _, agent := range c.agents {
		c.wg.Add(1)
		go c.runAgentWorker(agent)
	}

	log.Printf("✓ Started %d agents", len(c.agents))
	return nil
}

// runScheduler runs the task scheduling loop
func (c *CoordinatorV2) runScheduler() {
	defer c.wg.Done()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	log.Println("📅 Scheduler started")

	for {
		select {
		case <-c.ctx.Done():
			log.Println("📅 Scheduler stopped")
			return

		case <-ticker.C:
			// Check for idle agents and assign tasks
			for _, agent := range c.agents {
				if agent.IsIdle() {
					// Try to claim a task from the queue
					task, err := c.taskQueue.ClaimTask(agent.ID)
					if err != nil {
						log.Printf("⚠️  Failed to claim task for %s: %v", agent.ID, err)
						continue
					}

					if task != nil {
						// Send task to agent's work channel
						select {
						case agent.taskChan <- task:
							log.Printf("📋 Assigned task %s to %s", task.ID, agent.ID)
						default:
							// Channel full, task will be retried next cycle
							_ = c.taskQueue.UpdateTaskStatus(task.ID, models.TaskStatusPending)
						}
					}
				}
			}
		}
	}
}

// runAgentWorker runs a worker loop for an agent
func (c *CoordinatorV2) runAgentWorker(agent *Agent) {
	defer c.wg.Done()

	log.Printf("👷 Worker started for %s", agent.ID)

	for {
		select {
		case <-c.ctx.Done():
			log.Printf("👷 Worker stopped for %s", agent.ID)
			return

		case task := <-agent.taskChan:
			// Execute task
			err := agent.ExecuteTask(task)

			if err != nil {
				// Check if error is retryable
				if _, ok := err.(*executor.RetryableError); ok {
					// Update task for retry
					task.RetryCount++
					task.LastError = err.Error()

					if c.retryManager.ShouldRetry(task, nil) {
						delay := c.retryManager.CalculateDelay(task.RetryCount - 1)
						log.Printf("🔄 Task %s will retry in %s (attempt %d/%d)",
							task.ID, delay, task.RetryCount, task.MaxRetries)

						// Schedule retry
						go func() {
							time.Sleep(delay)
							_ = c.taskQueue.UpdateTaskStatus(task.ID, models.TaskStatusPending)
						}()
					} else {
						// Max retries reached
						log.Printf("❌ Task %s failed after %d retries", task.ID, task.RetryCount)
						_ = c.taskQueue.UpdateTaskStatus(task.ID, models.TaskStatusFailed)
					}
				} else {
					// Non-retryable error
					log.Printf("❌ Task %s failed: %v", task.ID, err)
					_ = c.taskQueue.UpdateTaskStatus(task.ID, models.TaskStatusFailed)
				}
			} else {
				// Task completed successfully
				log.Printf("✅ Task %s completed by %s", task.ID, agent.ID)

				// Update task status
				_ = c.taskQueue.UpdateTaskStatus(task.ID, models.TaskStatusCompleted)

				// Merge agent's work back to main
				if err := c.mergeAgentWork(agent); err != nil {
					log.Printf("⚠️  Failed to merge work from %s: %v", agent.ID, err)
				} else {
					log.Printf("🔀 Merged work from %s to main", agent.ID)
				}
			}
		}
	}
}

// mergeAgentWork merges an agent's work back to the main branch
func (c *CoordinatorV2) mergeAgentWork(agent *Agent) error {
	// TODO: Implement git merge logic
	// For now, just a placeholder
	return nil
}

// Stop stops the coordinator
func (c *CoordinatorV2) Stop() error {
	log.Println("🛑 Stopping coordinator...")

	// Cancel context to stop all goroutines
	c.cancel()

	// Wait for all goroutines to finish
	c.wg.Wait()

	// Reset orphaned tasks
	log.Println("Resetting orphaned tasks...")
	tasks := c.taskQueue.ListTasks()
	resetCount := 0
	for _, task := range tasks {
		if task.Status == models.TaskStatusInProgress {
			if err := c.taskQueue.ResetOrphanedTask(task.ID); err != nil {
				log.Printf("⚠️  Failed to reset task %s: %v", task.ID, err)
			} else {
				resetCount++
			}
		}
	}
	if resetCount > 0 {
		log.Printf("✓ Reset %d orphaned tasks", resetCount)
	}

	log.Println("✓ Coordinator stopped")
	return nil
}

// Cleanup cleans up resources
func (c *CoordinatorV2) Cleanup() error {
	log.Println("🧹 Cleaning up...")

	// Stop agents
	for _, agent := range c.agents {
		agent.Stop()
	}

	// Clean up worktrees
	for _, agent := range c.agents {
		agentNum := agent.ID[len("agent-"):]
		if err := c.worktreeManager.RemoveWorktree(agentNum); err != nil {
			log.Printf("⚠️  Failed to remove worktree for %s: %v", agent.ID, err)
		} else {
			log.Printf("✓ Removed worktree for %s", agent.ID)
		}
	}

	// Close task queue
	if c.taskQueue != nil {
		c.taskQueue.Close()
	}

	log.Println("✓ Cleanup complete")
	return nil
}

// GetAgentStatus returns status of all agents
func (c *CoordinatorV2) GetAgentStatus() []*models.AgentStatus {
	statuses := make([]*models.AgentStatus, len(c.agents))
	for i, agent := range c.agents {
		statuses[i] = agent.GetStatus()
	}
	return statuses
}

// GetTaskQueue returns the task queue (for monitoring)
func (c *CoordinatorV2) GetTaskQueue() *state.TaskQueue {
	return c.taskQueue
}
