package controller

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/analyzer"
	"github.com/yourusername/claude-swarm/pkg/git"
	"github.com/yourusername/claude-swarm/pkg/retry"
	"github.com/yourusername/claude-swarm/pkg/state"
	"github.com/yourusername/claude-swarm/pkg/tmux"
	"github.com/yourusername/claude-swarm/pkg/utils"
)

// Coordinator manages the agent swarm
type Coordinator struct {
	session         *tmux.Session
	agents          []*Agent
	taskQueue       *state.TaskQueue
	agentStateMgr   *state.AgentStateManager
	worktreeManager *git.WorktreeManager
	mergeManager    *git.MergeManager
	retryManager    *retry.RetryManager
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
	version    uint64  // State version number for detecting concurrent modifications
}

// CoordinatorConfig holds configuration for the coordinator
type CoordinatorConfig struct {
	NumAgents        int
	SessionName      string
	TaskQueuePath    string
	AgentStatePath   string
	MonitorInterval  time.Duration
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

	// Create agent state manager
	agentStatePath := config.AgentStatePath
	if agentStatePath == "" {
		agentStatePath = "~/.claude-swarm/agents.json"
	}
	agentStateMgr, err := state.NewAgentStateManager(agentStatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent state manager: %w", err)
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

	// Create retry manager
	retryManager := retry.NewRetryManager(retry.DefaultRetryConfig())

	ctx, cancel := context.WithCancel(context.Background())

	c := &Coordinator{
		session:         session,
		agents:          make([]*Agent, 0, config.NumAgents),
		taskQueue:       taskQueue,
		agentStateMgr:   agentStateMgr,
		worktreeManager: worktreeManager,
		mergeManager:    mergeManager,
		retryManager:    retryManager,
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

	// Start agent state persister
	c.wg.Add(1)
	go c.runStatePersister()

	log.Println("✓ All goroutines started")
}

// Stop stops the coordinator
func (c *Coordinator) Stop() error {
	log.Println("Stopping coordinator...")
	c.cancel()

	// 🔧 FIX: 使用带超时的等待，避免无限阻塞
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("✓ All goroutines stopped gracefully")
	case <-time.After(30 * time.Second):
		log.Println("⚠️  Timeout waiting for goroutines to stop (30s), forcing shutdown")
		// 继续执行清理，即使 goroutine 可能还在运行
	}

	// Save final agent state
	if c.agentStateMgr != nil {
		statuses := c.GetAgentStatus()
		if err := c.agentStateMgr.UpdateAgents(statuses); err != nil {
			log.Printf("⚠️  Failed to save final agent state: %v", err)
		}
		c.agentStateMgr.Close()
	}

	// 🔧 Reset all in_progress tasks to pending (cleanup orphaned tasks)
	log.Println("Resetting orphaned tasks...")
	if c.taskQueue != nil {
		tasks := c.taskQueue.ListTasks()
		resetCount := 0
		for _, task := range tasks {
			if task.Status == models.TaskStatusInProgress {
				if err := c.taskQueue.UpdateTaskStatus(task.ID, models.TaskStatusPending); err != nil {
					log.Printf("⚠️  Failed to reset task %s: %v", task.ID, err)
				} else {
					resetCount++
				}
			}
		}
		if resetCount > 0 {
			log.Printf("✓ Reset %d orphaned tasks to pending", resetCount)
		}
	}

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

// isTmuxSessionAlive checks if the tmux session is still running
func (c *Coordinator) isTmuxSessionAlive() bool {
	cmd := exec.Command("tmux", "has-session", "-t", c.session.Name)
	err := cmd.Run()
	return err == nil
}

// monitorAgent monitors a single agent
func (c *Coordinator) monitorAgent(agent *Agent) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ PANIC in monitorAgent for %s: %v\nStack trace will be logged by runtime", agent.ID, r)
		}
		c.wg.Done()
	}()

	ticker := time.NewTicker(c.monitorInterval)
	defer ticker.Stop()

	sessionDeadCount := 0
	const maxSessionDeadChecks = 3

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// 🔧 FIX: Check if tmux session is still alive
			if !c.isTmuxSessionAlive() {
				sessionDeadCount++
				log.Printf("⚠️  tmux 会话 '%s' 不可用 (检查 %d/%d)", c.session.Name, sessionDeadCount, maxSessionDeadChecks)

				if sessionDeadCount >= maxSessionDeadChecks {
					log.Printf("❌ tmux 会话 '%s' 已终止，停止监控 %s", c.session.Name, agent.ID)
					log.Printf("⚠️  coordinator 将在所有 agent 监控停止后退出")
					c.cancel() // 触发所有 goroutine 退出
					return
				}
				continue
			}
			sessionDeadCount = 0 // 重置计数器

			// Capture pane output
			output, err := agent.Pane.Capture()
			if err != nil {
				log.Printf("Error capturing %s output: %v", agent.ID, err)
				continue
			}

			// Analyze state
			detectedState := agent.Detector.Analyze(output)

			// Update agent status
			agent.mu.Lock()
			prevState := agent.Status.State
			currentTask := agent.Status.CurrentTask

			// 智能状态更新：如果有任务且检测到idle或waiting_confirm，说明任务刚完成
			taskCompleted := currentTask != nil &&
				prevState != models.AgentStateIdle &&
				prevState != models.AgentStateWaitingConfirm &&
				(detectedState == models.AgentStateIdle || detectedState == models.AgentStateWaitingConfirm)

			if taskCompleted {
				log.Printf("📊 %s: 检测到任务完成 (task: %s, state: %s)", agent.ID, currentTask.ID, detectedState)

				// 保存必要信息在锁外执行合并
				taskID := currentTask.ID
				currentVersion := agent.version

				// 临时释放锁以执行合并（避免死锁）
				agent.mu.Unlock()

				log.Printf("🔀 开始合并 %s 的工作到 main 分支...", agent.ID)
				mergeErr := c.mergeAgentWork(agent)

				// 重新获取锁，验证状态是否仍然有效
				agent.mu.Lock()

				// 使用版本号验证状态未被修改
				if agent.version == currentVersion &&
					agent.Status.CurrentTask != nil &&
					agent.Status.CurrentTask.ID == taskID {
					// 状态仍然有效，可以安全更新
					// 递增版本号表示状态已修改
					agent.version++

					if mergeErr != nil {
						log.Printf("❌ 合并失败 - %s: %v", agent.ID, mergeErr)
						_ = c.taskQueue.UpdateTaskStatus(taskID, models.TaskStatusFailed)
					} else {
						log.Printf("✅ 合并成功 - 任务 %s 已完成", taskID)
						_ = c.taskQueue.UpdateTaskStatus(taskID, models.TaskStatusCompleted)
					}

					// 清空任务
					agent.Status.CurrentTask = nil
					agent.Status.State = models.AgentStateIdle
					agent.Status.LastUpdate = time.Now()
					agent.Status.Output = agent.Detector.GetRecentOutput(10)

					log.Printf("🔄 %s state changed: %s → %s (task completed)", agent.ID, prevState, models.AgentStateIdle)
				} else {
					// 状态已被其他 goroutine 修改，记录警告
					log.Printf("⚠️  %s: 任务状态在合并过程中已变更 (version: %d → %d, expected task: %s, current: %v)",
						agent.ID, currentVersion, agent.version, taskID, agent.Status.CurrentTask)

					// 仍然更新任务队列中的状态
					if mergeErr != nil {
						_ = c.taskQueue.UpdateTaskStatus(taskID, models.TaskStatusFailed)
					} else {
						_ = c.taskQueue.UpdateTaskStatus(taskID, models.TaskStatusCompleted)
					}
				}

				agent.mu.Unlock()

				// 跳过后续的正常状态更新逻辑
				continue
			} else {
				// 正常状态更新
				if prevState != detectedState {
					agent.version++  // Increment version on state change
				}
				agent.Status.State = detectedState
				agent.Status.LastUpdate = time.Now()
				agent.Status.Output = agent.Detector.GetRecentOutput(10)
				agent.mu.Unlock()

				// Log state changes for debugging
				if prevState != detectedState {
					hasTaskStr := "no task"
					if currentTask != nil {
						hasTaskStr = fmt.Sprintf("task: %s", currentTask.ID)
					}
					log.Printf("🔄 %s state changed: %s → %s (%s)", agent.ID, prevState, detectedState, hasTaskStr)
				}
			}
		}
	}
}

// tryAssignTask atomically attempts to assign a task to an agent
// Returns true if assignment was successful, false otherwise
func (c *Coordinator) tryAssignTask(agent *Agent) bool {
	agent.mu.Lock()
	defer agent.mu.Unlock()

	// Check if agent is still idle and has no task
	if agent.Status.State != models.AgentStateIdle || agent.Status.CurrentTask != nil {
		return false
	}

	// Try to claim a task
	task, err := c.taskQueue.ClaimTask(agent.ID)
	if err != nil {
		log.Printf("❌ Error claiming task for %s: %v", agent.ID, err)
		return false
	}

	if task == nil {
		// No tasks available
		return false
	}

	// Assign task (version will be incremented when state changes)
	agent.Status.CurrentTask = task

	// Send task to agent (still holding lock to prevent state changes)
	if err := agent.Pane.SendLine(task.Description); err != nil {
		log.Printf("❌ Error sending task to %s: %v", agent.ID, err)
		// Task send failed, rollback
		agent.Status.CurrentTask = nil
		_ = c.taskQueue.UpdateTaskStatus(task.ID, models.TaskStatusPending)
		return false
	}

	log.Printf("📋 已分配任务 %s 给 %s: %s", task.ID, agent.ID, task.Description)
	return true
}

// runScheduler runs the task scheduler
func (c *Coordinator) runScheduler() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ PANIC in runScheduler: %v\nStack trace will be logged by runtime", r)
		}
		c.wg.Done()
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	log.Println("📅 Scheduler started")

	for {
		select {
		case <-c.ctx.Done():
			log.Println("📅 Scheduler stopped")
			return
		case <-ticker.C:
			// Find idle agents and try to assign tasks atomically
			hasIdleAgents := false
			for _, agent := range c.agents {
				// Quick check without lock first
				agent.mu.Lock()
				state := agent.Status.State
				hasTask := agent.Status.CurrentTask != nil

				// 🔧 FIX: 检测并修复状态不一致
				// 如果 Agent 状态不是 idle/waiting_confirm，但没有任务，说明状态不一致
				if !hasTask && state != models.AgentStateIdle && state != models.AgentStateWaitingConfirm {
					log.Printf("⚠️  %s 状态不一致：state=%s but hasTask=false，重置为 idle", agent.ID, state)
					agent.Status.State = models.AgentStateIdle
					agent.version++
					state = models.AgentStateIdle
				}

				isIdle := state == models.AgentStateIdle && !hasTask
				agent.mu.Unlock()

				log.Printf("📅 Scheduler check: %s state=%s hasTask=%v isIdle=%v", agent.ID, state, hasTask, isIdle)

				if isIdle {
					hasIdleAgents = true
					// Atomically try to assign a task
					assigned := c.tryAssignTask(agent)
					if !assigned {
						// Either no tasks available or assignment failed
						// Continue to check other agents
					}
				}
			}

			if hasIdleAgents {
				log.Printf("📅 Idle agents found, tasks may have been assigned")
			}
		}
	}
}

// runRescue runs the rescue engine
func (c *Coordinator) runRescue() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ PANIC in runRescue: %v\nStack trace will be logged by runtime", r)
		}
		c.wg.Done()
	}()

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
					agent.mu.Lock()
					currentTask := agent.Status.CurrentTask
					agent.mu.Unlock()

					if currentTask != nil {
						c.handleTaskError(agent, currentTask)
					} else {
						log.Printf("❌ %s encountered an error but has no current task", agent.ID)
					}
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

// validateMergePrerequisites 验证合并前置条件
func (c *Coordinator) validateMergePrerequisites() error {
	// 检查仓库状态
	cmd := exec.Command("git", "-C", c.repoPath, "status", "--porcelain")
	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("无法访问git仓库: %w", err)
	}

	// 检查当前分支
	cmd = exec.Command("git", "-C", c.repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("无法获取当前分支: %w", err)
	}

	currentBranch := strings.TrimSpace(string(output))
	if currentBranch != "main" {
		log.Printf("⚠️  当前不在main分支，将切换到main")
	}

	return nil
}

// handleTaskError handles task errors with intelligent retry logic
func (c *Coordinator) handleTaskError(agent *Agent, task *models.Task) {
	// Get recent output for error analysis
	output := agent.Detector.GetRecentOutput(50)

	// Analyze the error
	errorDetails := agent.Detector.AnalyzeError(output)

	log.Printf("❌ %s task %s encountered error: %s (Type: %v)",
		agent.ID, task.ID, errorDetails.Message, errorDetails.Type)

	// Determine if we should retry
	if c.retryManager.ShouldRetry(task, errorDetails) {
		// Record the retry
		c.retryManager.RecordRetry(task, errorDetails)

		// Calculate delay
		delay := c.retryManager.CalculateDelay(task.RetryCount - 1) // -1 because we already incremented

		log.Printf("🔄 Will retry task %s after %v (retry %d/%d)",
			task.ID, delay, task.RetryCount, task.MaxRetries)

		// Clear agent's current task
		agent.mu.Lock()
		agent.Status.CurrentTask = nil
		agent.Status.State = models.AgentStateIdle
		agent.version++
		agent.mu.Unlock()

		// Schedule retry by resetting task status after delay
		go func() {
			time.Sleep(delay)

			// Reset task to pending for retry
			if err := c.taskQueue.UpdateTaskStatus(task.ID, models.TaskStatusPending); err != nil {
				log.Printf("❌ Failed to reset task %s for retry: %v", task.ID, err)
			} else {
				log.Printf("✅ Task %s reset to pending for retry", task.ID)
			}
		}()
	} else {
		// Don't retry - mark as failed
		log.Printf("❌ Task %s failed permanently: %s", task.ID, errorDetails.Message)

		agent.mu.Lock()
		agent.Status.CurrentTask = nil
		agent.Status.State = models.AgentStateIdle
		agent.version++
		agent.mu.Unlock()

		// Update task with error details
		task.LastError = errorDetails.Message
		if err := c.taskQueue.UpdateTaskStatus(task.ID, models.TaskStatusFailed); err != nil {
			log.Printf("❌ Failed to update task %s status to failed: %v", task.ID, err)
		}
	}
}

// mergeAgentWork merges an agent's work into the main branch
func (c *Coordinator) mergeAgentWork(agent *Agent) error {
	c.mergeMu.Lock() // Protect main branch
	defer c.mergeMu.Unlock()

	agentID := strings.TrimPrefix(agent.ID, "agent-")
	branchName := fmt.Sprintf("agent-%s-branch", agentID)
	worktreePath := filepath.Join(c.repoPath, ".worktrees", fmt.Sprintf("agent-%s", agentID))

	log.Printf("🔀 Merging %s to main...", branchName)

	// Check disk space before merging (require at least 100MB free)
	requiredSpace := uint64(100 * 1024 * 1024) // 100MB
	if err := utils.CheckDiskSpace(c.repoPath, requiredSpace); err != nil {
		available, _ := utils.GetAvailableDiskSpace(c.repoPath)
		return fmt.Errorf("磁盘空间不足: %s 可用, 需要 %s: %w",
			utils.FormatBytes(available),
			utils.FormatBytes(requiredSpace),
			err)
	}

	// 验证前置条件
	if err := c.validateMergePrerequisites(); err != nil {
		return fmt.Errorf("合并前置条件验证失败: %w", err)
	}

	// 0. Commit any uncommitted changes in agent's worktree
	cmd := exec.Command("git", "-C", worktreePath, "add", ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("无法暂存agent工作区的更改: %w, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "-C", worktreePath, "diff-index", "--quiet", "HEAD")
	if err := cmd.Run(); err != nil {
		// There are changes to commit
		commitMsg := fmt.Sprintf("Agent %s: 自动提交任务完成的更改", agentID)
		cmd = exec.Command("git", "-C", worktreePath, "commit", "-m", commitMsg)
		if output, err := cmd.CombinedOutput(); err != nil {
			// Commit failed - 这可能是严重问题（磁盘满、权限等）
			return fmt.Errorf("无法提交agent工作区的更改: %w, output: %s", err, string(output))
		}
		log.Printf("✓ Auto-committed changes in %s", branchName)
	}

	// 1. Switch to main branch
	cmd = exec.Command("git", "-C", c.repoPath, "checkout", "main")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("无法切换到main分支: %w, output: %s", err, string(output))
	}

	// 2. Pull latest main (if there's a remote)
	// 检查是否有远程仓库
	cmd = exec.Command("git", "-C", c.repoPath, "remote", "get-url", "origin")
	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		// 有远程仓库，尝试pull
		cmd = exec.Command("git", "-C", c.repoPath, "pull", "origin", "main")
		if output, err := cmd.CombinedOutput(); err != nil {
			// Pull失败 - 可能是网络问题或冲突
			log.Printf("⚠️  Pull失败（将继续本地合并）: %v, output: %s", err, string(output))
			// 不返回错误，继续本地合并
		} else {
			log.Printf("✓ Pulled latest main from remote")
		}
	}

	// 保存当前HEAD，用于回滚
	cmd = exec.Command("git", "-C", c.repoPath, "rev-parse", "HEAD")
	originalHead, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("无法获取当前HEAD: %w", err)
	}
	originalHeadStr := strings.TrimSpace(string(originalHead))

	// 3. Execute merge
	result, err := c.mergeManager.MergeBranch(branchName)
	if err != nil {
		if err == git.ErrMergeConflict {
			log.Printf("⚠️  检测到合并冲突: %v", result.Conflicts)
			log.Printf("🧠 调用主控脑解决冲突...")

			// Call master brain to resolve conflicts
			if resolveErr := c.resolveMergeConflictWithMasterBrain(branchName, result.Conflicts); resolveErr != nil {
				log.Printf("❌ 主控脑无法解决冲突: %v", resolveErr)

				// 中止合并
				if abortErr := c.mergeManager.AbortMerge(); abortErr != nil {
					log.Printf("⚠️  中止合并失败: %v", abortErr)
				}

				return fmt.Errorf("合并冲突无法解决: %v", resolveErr)
			}

			log.Printf("✅ 主控脑成功解决冲突")

			// Try to complete the merge
			cmd = exec.Command("git", "-C", c.repoPath, "add", ".")
			if err := cmd.Run(); err != nil {
				_ = c.mergeManager.AbortMerge()
				return fmt.Errorf("无法暂存已解决的冲突: %w", err)
			}

			cmd = exec.Command("git", "-C", c.repoPath, "commit", "--no-edit")
			if output, err := cmd.CombinedOutput(); err != nil {
				// Commit failed - 回滚到原始状态
				log.Printf("❌ 提交合并失败，回滚到 %s", originalHeadStr[:8])
				rollbackCmd := exec.Command("git", "-C", c.repoPath, "reset", "--hard", originalHeadStr)
				if rollbackErr := rollbackCmd.Run(); rollbackErr != nil {
					log.Printf("⚠️  回滚失败: %v", rollbackErr)
				}
				return fmt.Errorf("无法提交合并: %w, output: %s", err, string(output))
			}

			// Get current commit hash
			cmd = exec.Command("git", "-C", c.repoPath, "rev-parse", "HEAD")
			if output, err := cmd.Output(); err == nil {
				result.CommitHash = strings.TrimSpace(string(output))
			}
			result.Success = true
		} else {
			// 其他合并错误
			return fmt.Errorf("合并失败: %w", err)
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
	// Use context with timeout to respect coordinator shutdown
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Minute)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Check why context was cancelled
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("master brain timed out resolving conflicts (5 minutes)")
			}
			// Coordinator is shutting down
			return fmt.Errorf("conflict resolution cancelled: coordinator shutting down")

		case <-ticker.C:
			// Check if conflicts are resolved
			cmd := exec.CommandContext(ctx, "git", "-C", c.repoPath, "diff", "--name-only", "--diff-filter=U")
			output, err := cmd.Output()
			if err != nil {
				// Command failed, continue polling
				continue
			}

			remainingConflicts := strings.TrimSpace(string(output))
			if remainingConflicts == "" {
				// All conflicts resolved
				log.Printf("✅ Master brain resolved all conflicts")
				return nil
			}

			conflictCount := len(strings.Split(remainingConflicts, "\n"))
			log.Printf("🧠 Master brain still working... (%d conflicts remaining)", conflictCount)
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

// runStatePersister periodically saves agent state to file
func (c *Coordinator) runStatePersister() {
	defer c.wg.Done()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			statuses := c.GetAgentStatus()
			if err := c.agentStateMgr.UpdateAgents(statuses); err != nil {
				log.Printf("⚠️  Failed to save agent state: %v", err)
			}
		}
	}
}
