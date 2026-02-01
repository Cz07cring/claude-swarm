package controller

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/executor"
	"github.com/yourusername/claude-swarm/pkg/git"
	"github.com/yourusername/claude-swarm/pkg/retry"
	"github.com/yourusername/claude-swarm/pkg/state"
)

// Coordinator manages the swarm using direct Claude CLI execution
type Coordinator struct {
	agents           []*Agent
	taskQueue        *state.TaskQueue
	worktreeManager  *git.WorktreeManager
	retryManager     *retry.RetryManager
	mergeManager     *git.MergeManager
	mainRepo         *git.Repository
	repoPath         string
	mergeMu          sync.Mutex // Protect concurrent merge operations

	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
}

// NewCoordinator creates a new coordinator using Claude CLI execution
func NewCoordinator(repoPath string, taskQueuePath string, numAgents int) (*Coordinator, error) {
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

	// Initialize main repository and merge manager
	mainRepo, err := git.NewRepository(repoPath)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to open main repository: %w", err)
	}
	mergeManager := git.NewMergeManager(mainRepo)

	c := &Coordinator{
		agents:          make([]*Agent, 0, numAgents),
		taskQueue:       taskQueue,
		worktreeManager: worktreeManager,
		retryManager:    retryManager,
		mergeManager:    mergeManager,
		mainRepo:        mainRepo,
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
func (c *Coordinator) Start() error {
	log.Println("🚀 Starting Claude Swarm Coordinator")

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
func (c *Coordinator) runScheduler() {
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
func (c *Coordinator) runAgentWorker(agent *Agent) {
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
func (c *Coordinator) mergeAgentWork(agent *Agent) error {
	if agent.Worktree == nil {
		return fmt.Errorf("agent %s has no worktree", agent.ID)
	}

	// 1. Check if agent's worktree has any changes
	worktreeRepo, err := git.NewRepository(agent.Worktree.Path)
	if err != nil {
		return fmt.Errorf("failed to open worktree repo: %w", err)
	}

	isClean, err := worktreeRepo.IsClean()
	if err != nil {
		return fmt.Errorf("failed to check worktree status: %w", err)
	}

	// 2. If there are uncommitted changes, commit them first
	if !isClean {
		log.Printf("📝 Agent %s has uncommitted changes, committing...", agent.ID)

		// Stage all changes
		if err := c.gitCommand(agent.Worktree.Path, "add", "-A"); err != nil {
			return fmt.Errorf("failed to stage changes: %w", err)
		}

		// Commit
		commitMsg := fmt.Sprintf("Agent %s: Auto-commit task work", agent.ID)
		if err := c.gitCommand(agent.Worktree.Path, "commit", "-m", commitMsg); err != nil {
			return fmt.Errorf("failed to commit changes: %w", err)
		}
	}

	// 3. Check if there are any commits to merge (compare with main)
	hasCommits, err := c.hasNewCommits(agent.Worktree.BranchName)
	if err != nil {
		return fmt.Errorf("failed to check for new commits: %w", err)
	}

	if !hasCommits {
		log.Printf("ℹ️  Agent %s has no new commits to merge", agent.ID)
		return nil
	}

	// 4. Merge the agent's branch into main (with lock to prevent concurrent merges)
	c.mergeMu.Lock()
	defer c.mergeMu.Unlock()

	log.Printf("🔀 Merging branch %s into main...", agent.Worktree.BranchName)

	result, err := c.mergeManager.MergeBranch(agent.Worktree.BranchName)
	if err != nil {
		if err == git.ErrMergeConflict {
			log.Printf("⚠️  Merge conflict detected for %s, aborting merge", agent.ID)
			_ = c.mergeManager.AbortMerge()
			return fmt.Errorf("merge conflict: %v", result.Conflicts)
		}
		return fmt.Errorf("merge failed: %w", err)
	}

	if result.FastForward {
		log.Printf("✅ Fast-forward merge completed for %s", agent.ID)
	} else {
		log.Printf("✅ Three-way merge completed for %s (commit: %s)", agent.ID, result.CommitHash[:8])
	}

	return nil
}

// gitCommand executes a git command in the specified directory
func (c *Coordinator) gitCommand(dir string, args ...string) error {
	cmdArgs := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

// hasNewCommits checks if a branch has commits that are not in main
func (c *Coordinator) hasNewCommits(branchName string) (bool, error) {
	cmd := exec.Command("git", "-C", c.repoPath, "rev-list", "--count", "main.."+branchName)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	count := strings.TrimSpace(string(output))
	return count != "0", nil
}

// Stop stops the coordinator
func (c *Coordinator) Stop() error {
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
func (c *Coordinator) Cleanup() error {
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
func (c *Coordinator) GetAgentStatus() []*models.AgentStatus {
	statuses := make([]*models.AgentStatus, len(c.agents))
	for i, agent := range c.agents {
		statuses[i] = agent.GetStatus()
	}
	return statuses
}

// GetTaskQueue returns the task queue (for monitoring)
func (c *Coordinator) GetTaskQueue() *state.TaskQueue {
	return c.taskQueue
}

// GetMergeStatuses 获取所有 agent 分支的合并状态
func (c *Coordinator) GetMergeStatuses() []*MergeStatus {
	var statuses []*MergeStatus

	for _, agent := range c.agents {
		if agent.Worktree == nil {
			continue
		}

		status := &MergeStatus{
			Branch:  agent.Worktree.BranchName,
			AgentID: agent.ID,
		}

		// 检查是否有改动
		worktreeRepo, err := git.NewRepository(agent.Worktree.Path)
		if err != nil {
			continue
		}

		isClean, _ := worktreeRepo.IsClean()
		status.HasChanges = !isClean

		// 获取提交数
		hasCommits, _ := c.hasNewCommits(agent.Worktree.BranchName)
		if hasCommits {
			// 获取提交数量
			cmd := exec.Command("git", "-C", c.repoPath, "rev-list", "--count", "main.."+agent.Worktree.BranchName)
			output, err := cmd.Output()
			if err == nil {
				fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &status.CommitCount)
			}

			// 获取修改的文件
			cmd = exec.Command("git", "-C", c.repoPath, "diff", "--name-only", "main.."+agent.Worktree.BranchName)
			output, err = cmd.Output()
			if err == nil {
				files := strings.Split(strings.TrimSpace(string(output)), "\n")
				for _, f := range files {
					if f != "" {
						status.Files = append(status.Files, f)
					}
				}
			}
		}

		// 判断是否可以合并（有提交且 worktree 干净）
		status.ReadyToMerge = status.CommitCount > 0 && !status.HasChanges

		statuses = append(statuses, status)
	}

	return statuses
}

// MergeBranch 合并指定分支到 main（供外部调用）
func (c *Coordinator) MergeBranch(branchName string) error {
	c.mergeMu.Lock()
	defer c.mergeMu.Unlock()

	log.Printf("🔀 Merging branch %s into main...", branchName)

	result, err := c.mergeManager.MergeBranch(branchName)
	if err != nil {
		if err == git.ErrMergeConflict {
			log.Printf("⚠️  Merge conflict detected for %s", branchName)
			_ = c.mergeManager.AbortMerge()
			return fmt.Errorf("merge conflict: %v", result.Conflicts)
		}
		return fmt.Errorf("merge failed: %w", err)
	}

	if result.FastForward {
		log.Printf("✅ Fast-forward merge completed for %s", branchName)
	} else {
		log.Printf("✅ Three-way merge completed for %s (commit: %s)", branchName, result.CommitHash[:8])
	}

	return nil
}

// GetConflictDetails 获取合并冲突详情
func (c *Coordinator) GetConflictDetails(branchName string) ([]string, string, error) {
	// 尝试合并（dry-run）
	cmd := exec.Command("git", "-C", c.repoPath, "merge", "--no-commit", "--no-ff", branchName)
	output, err := cmd.CombinedOutput()

	// 获取冲突文件
	conflictCmd := exec.Command("git", "-C", c.repoPath, "diff", "--name-only", "--diff-filter=U")
	conflictOutput, _ := conflictCmd.Output()

	var conflictFiles []string
	for _, f := range strings.Split(strings.TrimSpace(string(conflictOutput)), "\n") {
		if f != "" {
			conflictFiles = append(conflictFiles, f)
		}
	}

	// 中止合并
	abortCmd := exec.Command("git", "-C", c.repoPath, "merge", "--abort")
	_ = abortCmd.Run()

	if err != nil && len(conflictFiles) > 0 {
		return conflictFiles, string(output), nil
	}

	return nil, "", err
}

// MergeStatus 合并状态（本地定义，避免循环依赖）
type MergeStatus struct {
	Branch       string   `json:"branch"`
	AgentID      string   `json:"agent_id"`
	HasChanges   bool     `json:"has_changes"`
	CommitCount  int      `json:"commit_count"`
	Files        []string `json:"files"`
	ReadyToMerge bool     `json:"ready_to_merge"`
}
