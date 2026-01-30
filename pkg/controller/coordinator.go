package controller

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/analyzer"
	"github.com/yourusername/claude-swarm/pkg/git"
	"github.com/yourusername/claude-swarm/pkg/state"
	"github.com/yourusername/claude-swarm/pkg/tmux"
)

// Coordinator manages the agent swarm
type Coordinator struct {
	session         *tmux.Session
	agents          []*Agent
	taskQueue       *state.TaskQueue
	worktreeManager *git.WorktreeManager
	mergeManager    *git.MergeManager
	mergeMu         sync.Mutex
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	monitorInterval time.Duration
	repoPath        string
}

// Agent represents a single Claude agent
type Agent struct {
	ID         string
	Pane       *tmux.Pane
	Detector   *analyzer.Detector
	Status     *models.AgentStatus
	Worktree   *git.Worktree
	WorkingDir string
	mu         sync.Mutex
}

// CoordinatorConfig holds configuration for the coordinator
type CoordinatorConfig struct {
	NumAgents       int
	SessionName     string
	TaskQueuePath   string
	MonitorInterval time.Duration
}

// NewCoordinator creates a new coordinator
func NewCoordinator(config CoordinatorConfig) (*Coordinator, error) {
	// Create tmux session
	session, err := tmux.NewSession(config.SessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to create tmux session: %w", err)
	}

	log.Printf("✓ Created tmux session: %s", config.SessionName)

	// Create task queue
	taskQueue, err := state.NewTaskQueue(config.TaskQueuePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create task queue: %w", err)
	}

	// Get current working directory as repository path
	repoPath, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Create Worktree manager
	worktreeManager, err := git.NewWorktreeManager(git.WorktreeConfig{
		BaseRepoPath:    repoPath,
		WorktreeRootDir: ".worktrees",
		BaseBranch:      "main",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create worktree manager: %w", err)
	}

	// Create merge manager
	repo, err := git.NewRepository(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	mergeManager := git.NewMergeManager(repo)

	ctx, cancel := context.WithCancel(context.Background())

	c := &Coordinator{
		session:         session,
		agents:          make([]*Agent, 0, config.NumAgents),
		taskQueue:       taskQueue,
		worktreeManager: worktreeManager,
		mergeManager:    mergeManager,
		ctx:             ctx,
		cancel:          cancel,
		monitorInterval: config.MonitorInterval,
		repoPath:        repoPath,
	}

	// Create agents
	for i := 0; i < config.NumAgents; i++ {
		agentID := fmt.Sprintf("%d", i)

		// Create worktree
		worktree, err := worktreeManager.CreateWorktree(agentID)
		if err != nil {
			log.Printf("⚠️  Failed to create worktree for agent-%s: %v", agentID, err)
			c.cleanupWorktrees()
			return nil, fmt.Errorf("failed to create worktree: %w", err)
		}

		log.Printf("✓ Created worktree: %s (branch: %s)", worktree.Path, worktree.BranchName)

		var pane *tmux.Pane
		if i == 0 {
			// Use the first pane created with the session
			pane, err = session.GetPane(0)
		} else {
			// Split pane for additional agents
			pane, err = session.SplitPane(true) // horizontal split
		}

		if err != nil {
			c.cleanupWorktrees()
			return nil, fmt.Errorf("failed to create pane for agent-%d: %w", i, err)
		}

		agent := &Agent{
			ID:         fmt.Sprintf("agent-%s", agentID),
			Pane:       pane,
			Detector:   analyzer.NewDetector(),
			Worktree:   worktree,
			WorkingDir: worktree.Path,
			Status: &models.AgentStatus{
				AgentID:    fmt.Sprintf("agent-%s", agentID),
				State:      models.AgentStateIdle,
				LastUpdate: time.Now(),
			},
		}

		pane.AgentID = agent.ID
		c.agents = append(c.agents, agent)

		// Start claude in the worktree directory
		startCmd := fmt.Sprintf("cd %s && claude", worktree.Path)
		if err := pane.SendLine(startCmd); err != nil {
			log.Printf("Warning: failed to start claude in agent-%d: %v", i, err)
		}

		log.Printf("✓ Started %s in pane %d", agent.ID, pane.Index)

		// Give claude time to start
		time.Sleep(500 * time.Millisecond)
	}

	return c, nil
}

// Start starts the coordinator
func (c *Coordinator) Start() {
	log.Println("✓ Coordinator running...")
	log.Printf("  Monitor interval: %v", c.monitorInterval)
	log.Printf("  Agents: %d", len(c.agents))
	log.Printf("\nAttach to session: tmux attach -t %s\n", c.session.Name)

	// Log initial agent states
	for _, agent := range c.agents {
		log.Printf("  %s initial state: %s", agent.ID, agent.Status.State)
	}

	// Start monitoring each agent
	for _, agent := range c.agents {
		c.wg.Add(1)
		go c.monitorAgent(agent)
	}

	// Start scheduler
	c.wg.Add(1)
	go c.runScheduler()

	// Start rescue engine
	c.wg.Add(1)
	go c.runRescue()

	log.Println("✓ All goroutines started")
}

// Stop stops the coordinator
func (c *Coordinator) Stop() error {
	log.Println("Stopping coordinator...")
	c.cancel()
	c.wg.Wait()

	// Clean up worktrees
	log.Println("Cleaning up worktrees...")
	for _, agent := range c.agents {
		agentID := strings.TrimPrefix(agent.ID, "agent-")
		if err := c.worktreeManager.RemoveWorktree(agentID); err != nil {
			log.Printf("⚠️  Failed to remove worktree for %s: %v", agent.ID, err)
		} else {
			log.Printf("✓ Removed worktree for %s", agent.ID)
		}
	}

	// Kill tmux session
	if err := c.session.Kill(); err != nil {
		return fmt.Errorf("failed to kill session: %w", err)
	}

	log.Printf("✓ Killed tmux session: %s", c.session.Name)
	return nil
}

// monitorAgent monitors a single agent
func (c *Coordinator) monitorAgent(agent *Agent) {
	defer c.wg.Done()

	ticker := time.NewTicker(c.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// Capture pane output
			output, err := agent.Pane.Capture()
			if err != nil {
				log.Printf("Error capturing %s output: %v", agent.ID, err)
				continue
			}

			// Analyze state
			state := agent.Detector.Analyze(output)

			// Update agent status
			agent.mu.Lock()
			prevState := agent.Status.State
			agent.Status.State = state
			agent.Status.LastUpdate = time.Now()
			agent.Status.Output = agent.Detector.GetRecentOutput(10)

			// 🐛 FIX: 当agent完成任务回到idle状态时，更新任务状态为completed
			if prevState != models.AgentStateIdle && state == models.AgentStateIdle {
				if agent.Status.CurrentTask != nil {
					taskID := agent.Status.CurrentTask.ID

					// 尝试合并Agent的工作到main
					if err := c.mergeAgentWork(agent); err != nil {
						log.Printf("❌ Failed to merge work from %s: %v", agent.ID, err)
						_ = c.taskQueue.UpdateTaskStatus(taskID, models.TaskStatusFailed)
					} else {
						log.Printf("✅ Task %s completed and merged by %s", taskID, agent.ID)
						_ = c.taskQueue.UpdateTaskStatus(taskID, models.TaskStatusCompleted)
					}

					// 清空当前任务
					agent.Status.CurrentTask = nil
				}
			}

			agent.mu.Unlock()

			// Log state changes for debugging
			if prevState != state {
				log.Printf("🔄 %s state changed: %s → %s", agent.ID, prevState, state)
			}
		}
	}
}

// runScheduler runs the task scheduler
func (c *Coordinator) runScheduler() {
	defer c.wg.Done()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	log.Println("📅 Scheduler started")

	for {
		select {
		case <-c.ctx.Done():
			log.Println("📅 Scheduler stopped")
			return
		case <-ticker.C:
			// Find idle agents
			for _, agent := range c.agents {
				agent.mu.Lock()
				state := agent.Status.State
				hasTask := agent.Status.CurrentTask != nil
				isIdle := state == models.AgentStateIdle && !hasTask
				agent.mu.Unlock()

				log.Printf("📅 Scheduler check: %s state=%s hasTask=%v isIdle=%v", agent.ID, state, hasTask, isIdle)

				if isIdle {
					// Try to claim a task
					task, err := c.taskQueue.ClaimTask(agent.ID)
					if err != nil {
						log.Printf("❌ Error claiming task for %s: %v", agent.ID, err)
						continue
					}

					if task != nil {
						// Assign task to agent
						agent.mu.Lock()
						agent.Status.CurrentTask = task
						agent.Status.State = models.AgentStateWorking
						agent.mu.Unlock()

						// Send task to agent
						if err := agent.Pane.SendLine(task.Description); err != nil {
							log.Printf("❌ Error sending task to %s: %v", agent.ID, err)
							continue
						}

						log.Printf("📋 Assigned task %s to %s: %s", task.ID, agent.ID, task.Description)
					} else {
						log.Printf("📅 No pending tasks available")
					}
				}
			}
		}
	}
}

// runRescue runs the rescue engine
func (c *Coordinator) runRescue() {
	defer c.wg.Done()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			for _, agent := range c.agents {
				agent.mu.Lock()
				state := agent.Status.State
				agent.mu.Unlock()

				// Handle waiting confirmation
				if state == models.AgentStateWaitingConfirm {
					shouldConfirm, input, reason := agent.Detector.ShouldConfirm()
					if shouldConfirm {
						if err := agent.Pane.SendLine(input); err != nil {
							log.Printf("❌ Error sending confirmation to %s: %v", agent.ID, err)
						} else {
							log.Printf("✅ Auto-confirmed for %s (sent: %s, reason: %s)", agent.ID, input, reason)
						}
					} else {
						log.Printf("⚠️  %s waiting for confirmation (reason: %s)", agent.ID, reason)
						log.Printf("   请手动确认: tmux send-keys -t claude-swarm:0.%d \"1\" Enter", agent.Pane.Index)
					}
				}

				// Handle errors
				if state == models.AgentStateError {
					log.Printf("❌ %s encountered an error", agent.ID)
					// TODO: Implement error handling (retry, reassign, etc.)
				}

				// Handle stuck agents
				if state == models.AgentStateStuck {
					log.Printf("⏸️  %s appears to be stuck", agent.ID)
					// TODO: Implement stuck handling (restart, reassign, etc.)
				}
			}
		}
	}
}

// AddTask adds a task to the queue
func (c *Coordinator) AddTask(task *models.Task) error {
	return c.taskQueue.AddTask(task)
}

// GetAgentStatus returns the status of all agents
func (c *Coordinator) GetAgentStatus() []*models.AgentStatus {
	statuses := make([]*models.AgentStatus, len(c.agents))
	for i, agent := range c.agents {
		agent.mu.Lock()
		// Create a copy to avoid race conditions
		status := &models.AgentStatus{
			AgentID:     agent.Status.AgentID,
			State:       agent.Status.State,
			CurrentTask: agent.Status.CurrentTask,
			LastUpdate:  agent.Status.LastUpdate,
			Output:      agent.Status.Output,
		}
		agent.mu.Unlock()
		statuses[i] = status
	}
	return statuses
}

// GetTaskQueue returns the task queue
func (c *Coordinator) GetTaskQueue() *state.TaskQueue {
	return c.taskQueue
}

// cleanupWorktrees cleans up all worktrees on initialization failure
func (c *Coordinator) cleanupWorktrees() {
	for _, agent := range c.agents {
		if agent.Worktree != nil {
			agentID := strings.TrimPrefix(agent.ID, "agent-")
			_ = c.worktreeManager.RemoveWorktree(agentID)
		}
	}
}

// mergeAgentWork merges an agent's work into the main branch
func (c *Coordinator) mergeAgentWork(agent *Agent) error {
	c.mergeMu.Lock() // Protect main branch
	defer c.mergeMu.Unlock()

	agentID := strings.TrimPrefix(agent.ID, "agent-")
	branchName := fmt.Sprintf("agent-%s-branch", agentID)

	log.Printf("🔀 Merging %s to main...", branchName)

	// 1. Switch to main branch
	cmd := exec.Command("git", "-C", c.repoPath, "checkout", "main")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout main: %w, output: %s", err, string(output))
	}

	// 2. Pull latest main (if there's a remote)
	cmd = exec.Command("git", "-C", c.repoPath, "pull", "origin", "main")
	_ = cmd.Run() // Ignore errors

	// 3. Execute merge
	result, err := c.mergeManager.MergeBranch(branchName)
	if err != nil {
		if err == git.ErrMergeConflict {
			log.Printf("⚠️  Merge conflict detected: %v", result.Conflicts)
			log.Printf("🧠 Calling master brain to resolve conflicts...")

			// Call master brain to resolve conflicts
			if resolveErr := c.resolveMergeConflictWithMasterBrain(branchName, result.Conflicts); resolveErr != nil {
				log.Printf("❌ Master brain failed to resolve conflicts: %v", resolveErr)
				_ = c.mergeManager.AbortMerge()
				return fmt.Errorf("merge conflict could not be resolved: %v", resolveErr)
			}

			log.Printf("✅ Master brain successfully resolved conflicts")

			// Try to complete the merge
			cmd = exec.Command("git", "-C", c.repoPath, "add", ".")
			if err := cmd.Run(); err != nil {
				_ = c.mergeManager.AbortMerge()
				return fmt.Errorf("failed to stage resolved conflicts: %w", err)
			}

			cmd = exec.Command("git", "-C", c.repoPath, "commit", "--no-edit")
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to commit merge: %w, output: %s", err, string(output))
			}

			// Get current commit hash
			cmd = exec.Command("git", "-C", c.repoPath, "rev-parse", "HEAD")
			if output, err := cmd.Output(); err == nil {
				result.CommitHash = strings.TrimSpace(string(output))
			}
			result.Success = true
		} else {
			return err
		}
	}

	// 4. Log merge result
	if result.FastForward {
		log.Printf("✓ Fast-forward merge (commit: %s)", result.CommitHash[:8])
	} else {
		log.Printf("✓ Three-way merge (commit: %s)", result.CommitHash[:8])
	}

	// 5. Push to remote (optional)
	if c.shouldPushToRemote() {
		cmd = exec.Command("git", "-C", c.repoPath, "push", "origin", "main")
		if err := cmd.Run(); err != nil {
			log.Printf("⚠️  Failed to push: %v", err)
		} else {
			log.Printf("✓ Pushed to remote")
		}
	}

	// 6. Reset agent's worktree to latest main
	cmd = exec.Command("git", "-C", agent.WorkingDir, "reset", "--hard", "main")
	if err := cmd.Run(); err != nil {
		log.Printf("⚠️  Failed to reset worktree: %v", err)
	}

	return nil
}

// resolveMergeConflictWithMasterBrain uses a master brain agent to resolve merge conflicts
func (c *Coordinator) resolveMergeConflictWithMasterBrain(branchName string, conflicts []string) error {
	// Find an idle agent to act as the master brain
	var masterBrain *Agent
	for _, agent := range c.agents {
		agent.mu.Lock()
		if agent.Status.State == models.AgentStateIdle && agent.Status.CurrentTask == nil {
			masterBrain = agent
			agent.mu.Unlock()
			break
		}
		agent.mu.Unlock()
	}

	if masterBrain == nil {
		return fmt.Errorf("no idle agent available to act as master brain")
	}

	log.Printf("🧠 Using %s as master brain for conflict resolution", masterBrain.ID)

	// Build conflict resolution task
	conflictInfo := fmt.Sprintf(`Merge conflict detected when merging branch '%s' to main.

Conflicted files:
`, branchName)

	for _, file := range conflicts {
		conflictInfo += fmt.Sprintf("  - %s\n", file)
	}

	conflictInfo += `
Please resolve these conflicts by:
1. Examining the conflicted files
2. Understanding both versions of the changes
3. Manually editing the files to resolve conflicts
4. Ensuring the code compiles and tests pass
5. The changes will be automatically staged and committed

Working directory: ` + c.repoPath

	// Send conflict resolution task to master brain
	if err := masterBrain.Pane.SendLine(conflictInfo); err != nil {
		return fmt.Errorf("failed to send conflict resolution task to master brain: %w", err)
	}

	log.Printf("🧠 Master brain is analyzing conflicts...")

	// Wait for master brain to complete (poll for clean status)
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("master brain timed out resolving conflicts")
		case <-ticker.C:
			// Check if conflicts are resolved
			cmd := exec.Command("git", "-C", c.repoPath, "diff", "--name-only", "--diff-filter=U")
			output, err := cmd.Output()
			if err != nil {
				continue
			}

			remainingConflicts := strings.TrimSpace(string(output))
			if remainingConflicts == "" {
				// All conflicts resolved
				log.Printf("✅ Master brain resolved all conflicts")
				return nil
			}

			log.Printf("🧠 Master brain still working... (%d conflicts remaining)", len(strings.Split(remainingConflicts, "\n")))
		}
	}
}

// shouldPushToRemote checks if the repository has a remote configured
func (c *Coordinator) shouldPushToRemote() bool {
	cmd := exec.Command("git", "-C", c.repoPath, "remote")
	output, err := cmd.Output()
	if err != nil || len(strings.TrimSpace(string(output))) == 0 {
		return false
	}
	return true
}
