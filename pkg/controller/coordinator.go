package controller

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/analyzer"
	"github.com/yourusername/claude-swarm/pkg/state"
	"github.com/yourusername/claude-swarm/pkg/tmux"
)

// Coordinator manages the agent swarm
type Coordinator struct {
	session         *tmux.Session
	agents          []*Agent
	taskQueue       *state.TaskQueue
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	monitorInterval time.Duration
}

// Agent represents a single Claude agent
type Agent struct {
	ID       string
	Pane     *tmux.Pane
	Detector *analyzer.Detector
	Status   *models.AgentStatus
	mu       sync.Mutex
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

	ctx, cancel := context.WithCancel(context.Background())

	c := &Coordinator{
		session:         session,
		agents:          make([]*Agent, 0, config.NumAgents),
		taskQueue:       taskQueue,
		ctx:             ctx,
		cancel:          cancel,
		monitorInterval: config.MonitorInterval,
	}

	// Create agents
	for i := 0; i < config.NumAgents; i++ {
		var pane *tmux.Pane
		if i == 0 {
			// Use the first pane created with the session
			pane, err = session.GetPane(0)
		} else {
			// Split pane for additional agents
			pane, err = session.SplitPane(true) // horizontal split
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create pane for agent-%d: %w", i, err)
		}

		agent := &Agent{
			ID:       fmt.Sprintf("agent-%d", i),
			Pane:     pane,
			Detector: analyzer.NewDetector(),
			Status: &models.AgentStatus{
				AgentID:    fmt.Sprintf("agent-%d", i),
				State:      models.AgentStateIdle,
				LastUpdate: time.Now(),
			},
		}

		pane.AgentID = agent.ID
		c.agents = append(c.agents, agent)

		// Start claude in the pane
		if err := pane.SendLine("claude"); err != nil {
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
					// 更新任务状态为completed
					if err := c.taskQueue.UpdateTaskStatus(taskID, models.TaskStatusCompleted); err != nil {
						log.Printf("❌ Error updating task status for %s: %v", taskID, err)
					} else {
						log.Printf("✅ Task %s completed by %s", taskID, agent.ID)
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
