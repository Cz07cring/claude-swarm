// Package tui implements the terminal user interface for Claude Agent Swarm monitoring.
// It provides real-time visualization of task queues, agent status, and execution logs.
//
// The TUI is built using the Bubble Tea framework (https://github.com/charmbracelet/bubbletea)
// which follows the Elm Architecture pattern for building interactive terminal applications.
//
// Usage:
//
//	taskQueue := state.NewTaskQueue("~/.claude-swarm/tasks.json")
//	getAgents := func() []*models.AgentStatus {
//	    return state.LoadAgentStates("~/.claude-swarm/agents.json")
//	}
//	if err := tui.Run(taskQueue, getAgents); err != nil {
//	    log.Fatal(err)
//	}
package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/state"
)

// ActivePane represents which panel is currently focused in the TUI.
// The active panel is highlighted with a cyan border and can receive keyboard input.
type ActivePane int

const (
	// PaneTasks indicates the Tasks panel (left) is active
	PaneTasks ActivePane = iota
	// PaneAgents indicates the Agents panel (middle) is active
	PaneAgents
	// PaneLogs indicates the Logs panel (right) is active
	PaneLogs
)

// tickMsg is sent on every tick interval to trigger data refresh.
// The default tick interval is 2 seconds.
type tickMsg time.Time

// Dashboard is the main TUI model that implements the Bubble Tea Model interface.
// It coordinates three sub-views (Tasks, Agents, Logs) and handles user input.
//
// The Dashboard follows the Elm Architecture:
//   - Init(): Initializes the model and returns initial commands
//   - Update(msg): Processes messages (keyboard, timer) and updates state
//   - View(): Renders the current state as a string
type Dashboard struct {
	taskQueue    *state.TaskQueue
	getAgentsFn  func() []*models.AgentStatus
	taskList     *TaskListView
	agentGrid    *AgentGridView
	logViewer    *LogViewerView
	activePane   ActivePane
	width        int
	height       int
	tasks        []*models.Task
	agents       []*models.AgentStatus
	quitting     bool
	updateTicker *time.Ticker
}

// NewDashboard creates a new Dashboard instance.
//
// Parameters:
//   - taskQueue: TaskQueue instance for reading task states
//   - getAgentsFn: Function that returns current agent states
//
// The dashboard will automatically refresh every 2 seconds by calling
// taskQueue.ListTasks() and getAgentsFn() to fetch the latest state.
//
// Example:
//
//	taskQueue := state.NewTaskQueue("~/.claude-swarm/tasks.json")
//	getAgents := func() []*models.AgentStatus {
//	    return state.LoadAgentStates("~/.claude-swarm/agents.json")
//	}
//	dashboard := NewDashboard(taskQueue, getAgents)
func NewDashboard(taskQueue *state.TaskQueue, getAgentsFn func() []*models.AgentStatus) *Dashboard {
	return &Dashboard{
		taskQueue:    taskQueue,
		getAgentsFn:  getAgentsFn,
		activePane:   PaneTasks,
		tasks:        make([]*models.Task, 0),
		agents:       make([]*models.AgentStatus, 0),
		updateTicker: time.NewTicker(2 * time.Second),
	}
}

// Init initializes the dashboard and returns initial commands.
// This is part of the Bubble Tea Model interface.
//
// It performs the following:
//  1. Loads initial task and agent data
//  2. Starts the refresh timer (2 second intervals)
//  3. Enters alternate screen mode (full-screen TUI)
//
// Returns a batch of commands to be executed by the Bubble Tea runtime.
func (m *Dashboard) Init() tea.Cmd {
	// Initialize data immediately
	m.refreshData()

	return tea.Batch(
		tickCmd(),
		tea.EnterAltScreen,
	)
}

// Update processes incoming messages and updates the model state.
// This is part of the Bubble Tea Model interface.
//
// Handled message types:
//   - tea.KeyMsg: Keyboard input (navigation, quit, etc.)
//   - tea.WindowSizeMsg: Terminal window resize events
//   - tickMsg: Timer ticks for periodic data refresh
//
// Returns the updated model and any commands to execute.
func (m *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.initializeViews()
		m.refreshData()

	case tickMsg:
		if !m.quitting {
			m.refreshData()
			return m, tickCmd()
		}
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m *Dashboard) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		m.updateTicker.Stop()
		return m, tea.Quit

	case "r", "R":
		// Manual refresh
		m.refreshData()
		return m, nil

	case "tab":
		// Cycle through panes
		m.activePane = (m.activePane + 1) % 3

	case "shift+tab":
		// Cycle backwards through panes
		m.activePane = (m.activePane + 2) % 3

	case "up", "k":
		switch m.activePane {
		case PaneTasks:
			m.taskList.MoveUp()
			m.updateLogViewer()
		case PaneAgents:
			m.agentGrid.MoveUp()
			m.updateLogViewer()
		}

	case "down", "j":
		switch m.activePane {
		case PaneTasks:
			m.taskList.MoveDown()
			m.updateLogViewer()
		case PaneAgents:
			m.agentGrid.MoveDown()
			m.updateLogViewer()
		}

	case "left", "h":
		if m.activePane == PaneAgents {
			m.agentGrid.MoveLeft()
			m.updateLogViewer()
		}

	case "right", "l":
		if m.activePane == PaneAgents {
			m.agentGrid.MoveRight()
			m.updateLogViewer()
		}

	case "pageup", "pgup":
		if m.activePane == PaneLogs && m.logViewer != nil {
			m.logViewer.ScrollUp(10)
		}

	case "pagedown", "pgdown":
		if m.activePane == PaneLogs && m.logViewer != nil {
			m.logViewer.ScrollDown(10)
		}

	case "a":
		// Toggle auto-scroll in log viewer
		if m.activePane == PaneLogs && m.logViewer != nil {
			m.logViewer.ToggleAutoScroll()
		}

	case "enter":
		// Select current item and update log viewer
		m.updateLogViewer()

	case "home":
		// Jump to first item
		switch m.activePane {
		case PaneTasks:
			m.taskList.MoveToFirst()
			m.updateLogViewer()
		case PaneAgents:
			m.agentGrid.MoveToFirst()
			m.updateLogViewer()
		}

	case "end":
		// Jump to last item
		switch m.activePane {
		case PaneTasks:
			m.taskList.MoveToLast()
			m.updateLogViewer()
		case PaneAgents:
			m.agentGrid.MoveToLast()
			m.updateLogViewer()
		}
	}

	return m, nil
}

// View renders the dashboard
func (m *Dashboard) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	if m.quitting {
		return "Goodbye!\n"
	}

	// Initialize views if needed
	if m.taskList == nil {
		m.initializeViews()
	}

	// Safety check: if views are still nil, return loading message
	if m.taskList == nil || m.agentGrid == nil || m.logViewer == nil {
		return "Loading..."
	}

	// Calculate dimensions for three-pane layout
	// Layout: [Tasks | Agents | Logs]
	taskWidth := m.width / 3
	agentWidth := m.width / 3
	logWidth := m.width - taskWidth - agentWidth - 6 // Account for borders

	contentHeight := m.height - 6 // Reserve space for title, status bar and help

	// Render title and status bar
	title := titleStyle.Render("🐝 Claude Agent Swarm Monitor")
	statusBar := m.renderStatusBar()

	// Render task list pane
	taskTitle := "📋 Tasks"
	if m.activePane == PaneTasks {
		taskTitle = "📋 Tasks ◄"
	}
	taskPane := m.renderPane(taskTitle, m.taskList.Render(), taskWidth, contentHeight, m.activePane == PaneTasks)

	// Render agent grid pane
	agentTitle := "🤖 Agents"
	if m.activePane == PaneAgents {
		agentTitle = "🤖 Agents ◄"
	}
	agentPane := m.renderPane(agentTitle, m.agentGrid.Render(m.activePane == PaneAgents), agentWidth, contentHeight, m.activePane == PaneAgents)

	// Render log viewer pane
	logTitle := "📜 Logs"
	if m.activePane == PaneLogs {
		logTitle = "📜 Logs ◄"
	}
	logPane := m.renderPane(logTitle, m.logViewer.Render(), logWidth, contentHeight, m.activePane == PaneLogs)

	// Join panes horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, taskPane, agentPane, logPane)

	// Render help text based on active pane
	var helpText string
	switch m.activePane {
	case PaneLogs:
		helpText = "Tab: 切换面板 | PgUp/PgDn: 滚动日志 | a: 自动滚动 | Home/End: 顶部/底部 | q: 退出 | r: 刷新"
	case PaneAgents:
		helpText = "Tab: 切换面板 | ↑↓←→/hjkl: 导航 | Home/End: 首个/末个 | Enter: 选择 | q: 退出 | r: 刷新"
	case PaneTasks:
		helpText = "Tab: 切换面板 | ↑↓/jk: 导航 | Home/End: 首个/末个 | Enter: 选择 | q: 退出 | r: 刷新"
	default:
		helpText = "Tab: 切换面板 | ↑↓/jk: 导航 | Enter: 选择 | q: 退出 | r: 刷新"
	}
	help := helpStyle.Render(helpText)

	// Combine everything
	return lipgloss.JoinVertical(lipgloss.Left, title, statusBar, content, help)
}

// renderStatusBar renders the top status bar with cluster metrics
func (m *Dashboard) renderStatusBar() string {
	// Calculate metrics
	totalTasks := len(m.tasks)
	totalAgents := len(m.agents)

	var pendingTasks, activeTasks, completedTasks, failedTasks int
	for _, task := range m.tasks {
		switch task.Status {
		case models.TaskStatusPending:
			pendingTasks++
		case models.TaskStatusInProgress:
			activeTasks++
		case models.TaskStatusCompleted:
			completedTasks++
		case models.TaskStatusFailed:
			failedTasks++
		}
	}

	var idleAgents, workingAgents, waitingAgents, errorAgents int
	for _, agent := range m.agents {
		switch agent.State {
		case models.AgentStateIdle:
			idleAgents++
		case models.AgentStateWorking:
			workingAgents++
		case models.AgentStateWaitingConfirm:
			waitingAgents++
		case models.AgentStateError, models.AgentStateStuck:
			errorAgents++
		}
	}

	// Calculate completion rate
	completionRate := 0.0
	if totalTasks > 0 {
		completionRate = float64(completedTasks) / float64(totalTasks) * 100
	}

	// Format metrics
	agentMetric := fmt.Sprintf("%s: %s/%s 工作中 • %s 空闲",
		metricLabelStyle.Render("Agents"),
		metricValueStyle.Render(fmt.Sprintf("%d", workingAgents)),
		metricValueStyle.Render(fmt.Sprintf("%d", totalAgents)),
		metricValueStyle.Render(fmt.Sprintf("%d", idleAgents)),
	)

	taskMetric := fmt.Sprintf("%s: %s 总计 • %s 完成 • %s 进行中 • %s 待处理",
		metricLabelStyle.Render("Tasks"),
		metricValueStyle.Render(fmt.Sprintf("%d", totalTasks)),
		metricSuccessStyle.Render(fmt.Sprintf("%d", completedTasks)),
		metricValueStyle.Render(fmt.Sprintf("%d", activeTasks)),
		metricLabelStyle.Render(fmt.Sprintf("%d", pendingTasks)),
	)

	completionMetric := fmt.Sprintf("%s: %s",
		metricLabelStyle.Render("完成率"),
		metricSuccessStyle.Render(fmt.Sprintf("%.1f%%", completionRate)),
	)

	// Add warnings if any
	warnings := ""
	if errorAgents > 0 {
		warnings += " • " + metricErrorStyle.Render(fmt.Sprintf("⚠ %d 错误", errorAgents))
	}
	if waitingAgents > 0 {
		warnings += " • " + metricWarningStyle.Render(fmt.Sprintf("⏸ %d 等待", waitingAgents))
	}
	if failedTasks > 0 {
		warnings += " • " + metricErrorStyle.Render(fmt.Sprintf("✗ %d 失败", failedTasks))
	}

	statusContent := fmt.Sprintf("%s  |  %s  |  %s%s",
		agentMetric, taskMetric, completionMetric, warnings)

	return statusBarStyle.Width(m.width).Render(statusContent)
}

// renderPane renders a pane with title and content
func (m *Dashboard) renderPane(title, content string, width, height int, active bool) string {
	titleRendered := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary).
		Width(width).
		Render(title)

	var paneStyle lipgloss.Style
	if active {
		paneStyle = activePanelStyle
	} else {
		paneStyle = panelStyle
	}

	paneContent := lipgloss.JoinVertical(lipgloss.Left, titleRendered, content)
	return paneStyle.
		Width(width).
		Height(height).
		Render(paneContent)
}

// initializeViews initializes the view components with correct dimensions
func (m *Dashboard) initializeViews() {
	// Ensure we have valid dimensions
	if m.width == 0 || m.height == 0 {
		return
	}

	taskWidth := m.width / 3
	agentWidth := m.width / 3
	logWidth := m.width - taskWidth - agentWidth - 6
	contentHeight := m.height - 8 // Account for borders and padding

	if contentHeight < 5 {
		contentHeight = 5
	}

	// Ensure minimum width
	if taskWidth < 10 {
		taskWidth = 10
	}
	if agentWidth < 10 {
		agentWidth = 10
	}
	if logWidth < 10 {
		logWidth = 10
	}

	m.taskList = NewTaskListView(taskWidth-4, contentHeight)
	m.agentGrid = NewAgentGridView(agentWidth-4, contentHeight)
	m.logViewer = NewLogViewerView(logWidth-4, contentHeight)

	// Initial data refresh
	m.refreshData()
}

// refreshData refreshes task and agent data
func (m *Dashboard) refreshData() {
	// Get tasks
	if m.taskQueue != nil {
		m.tasks = m.taskQueue.ListTasks()
		if m.taskList != nil {
			m.taskList.Update(m.tasks)
		}
	}

	// Get agents
	if m.getAgentsFn != nil {
		m.agents = m.getAgentsFn()
		if m.agentGrid != nil {
			m.agentGrid.Update(m.agents)
		}
	}

	// Update log viewer with current selection
	m.updateLogViewer()
}

// updateLogViewer updates the log viewer based on active pane
func (m *Dashboard) updateLogViewer() {
	if m.logViewer == nil {
		return
	}

	switch m.activePane {
	case PaneAgents:
		// Show selected agent's logs
		agent := m.agentGrid.GetSelectedAgent()
		m.logViewer.Update(agent)

	case PaneTasks:
		// Show agent working on selected task
		task := m.taskList.GetSelectedTask()
		if task != nil && task.AssigneeID != "" {
			// Find agent by ID
			for _, agent := range m.agents {
				if agent.AgentID == task.AssigneeID {
					m.logViewer.Update(agent)
					return
				}
			}
		}
		// If no agent found, show first agent
		if len(m.agents) > 0 {
			m.logViewer.Update(m.agents[0])
		} else {
			m.logViewer.Update(nil)
		}

	default:
		// Default to first agent
		if len(m.agents) > 0 {
			m.logViewer.Update(m.agents[0])
		} else {
			m.logViewer.Update(nil)
		}
	}
}

// tickCmd returns a command that sends a tick message after 2 seconds
func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Run starts the dashboard TUI
func Run(taskQueue *state.TaskQueue, getAgentsFn func() []*models.AgentStatus) error {
	p := tea.NewProgram(
		NewDashboard(taskQueue, getAgentsFn),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}
