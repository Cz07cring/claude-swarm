package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/claude-swarm/internal/models"
)

// TaskListView renders a list of tasks with status indicators
type TaskListView struct {
	tasks         []*models.Task
	selectedIndex int
	height        int
	width         int
}

// NewTaskListView creates a new task list view
func NewTaskListView(width, height int) *TaskListView {
	return &TaskListView{
		tasks:         make([]*models.Task, 0),
		selectedIndex: 0,
		height:        height,
		width:         width,
	}
}

// Update updates the task list
func (v *TaskListView) Update(tasks []*models.Task) {
	v.tasks = tasks
	// Ensure selected index is valid
	if v.selectedIndex >= len(v.tasks) && len(v.tasks) > 0 {
		v.selectedIndex = len(v.tasks) - 1
	}
	if v.selectedIndex < 0 {
		v.selectedIndex = 0
	}
}

// MoveUp moves selection up
func (v *TaskListView) MoveUp() {
	if v.selectedIndex > 0 {
		v.selectedIndex--
	}
}

// MoveDown moves selection down
func (v *TaskListView) MoveDown() {
	if v.selectedIndex < len(v.tasks)-1 {
		v.selectedIndex++
	}
}

// MoveToFirst moves selection to the first task
func (v *TaskListView) MoveToFirst() {
	if len(v.tasks) > 0 {
		v.selectedIndex = 0
	}
}

// MoveToLast moves selection to the last task
func (v *TaskListView) MoveToLast() {
	if len(v.tasks) > 0 {
		v.selectedIndex = len(v.tasks) - 1
	}
}

// GetSelectedTask returns the currently selected task
func (v *TaskListView) GetSelectedTask() *models.Task {
	if v.selectedIndex >= 0 && v.selectedIndex < len(v.tasks) {
		return v.tasks[v.selectedIndex]
	}
	return nil
}

// Render renders the task list view
func (v *TaskListView) Render() string {
	if len(v.tasks) == 0 {
		return lipgloss.NewStyle().
			Width(v.width).
			Height(v.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No tasks")
	}

	var lines []string
	for i, task := range v.tasks {
		line := v.renderTask(task, i == v.selectedIndex)
		lines = append(lines, line)
	}

	// Truncate to height
	if len(lines) > v.height {
		// Show tasks around selected index
		start := v.selectedIndex - v.height/2
		if start < 0 {
			start = 0
		}
		end := start + v.height
		if end > len(lines) {
			end = len(lines)
			start = end - v.height
			if start < 0 {
				start = 0
			}
		}
		lines = lines[start:end]
	}

	return strings.Join(lines, "\n")
}

// renderTask renders a single task line
func (v *TaskListView) renderTask(task *models.Task, selected bool) string {
	// Status indicator with emoji
	var statusIcon string
	var statusStyle lipgloss.Style
	switch task.Status {
	case models.TaskStatusPending:
		statusIcon = "⏳"
		statusStyle = statusIdleStyle
	case models.TaskStatusInProgress:
		statusIcon = "🔄"
		statusStyle = statusWorkingStyle
	case models.TaskStatusCompleted:
		statusIcon = "✅"
		statusStyle = statusSuccessStyle
	case models.TaskStatusFailed:
		statusIcon = "❌"
		statusStyle = statusErrorStyle
	default:
		statusIcon = "❓"
		statusStyle = statusIdleStyle
	}

	// Format task info
	statusStr := statusStyle.Render(statusIcon)

	// Truncate description if too long
	desc := task.Description
	maxDescLen := v.width - 18 // Reserve space for icons and info
	if len(desc) > maxDescLen {
		desc = desc[:maxDescLen-3] + "..."
	}

	// Format assignee with better styling
	assignee := ""
	if task.AssigneeID != "" {
		agentID := task.AssigneeID
		if len(agentID) > 8 {
			agentID = agentID[:8]
		}
		assignee = lipgloss.NewStyle().
			Foreground(colorInfo).
			Render(fmt.Sprintf("[%s]", agentID))
	}

	// Time since created/updated
	timeSince := lipgloss.NewStyle().
		Foreground(colorMuted).
		Render(formatTimeSince(task.UpdatedAt))

	line := fmt.Sprintf("%s %s %s %s", statusStr, desc, assignee, timeSince)

	// Apply selection style
	if selected {
		return taskSelectedStyle.Width(v.width).Render(line)
	}
	return taskItemStyle.Width(v.width).Render(line)
}

// formatTimeSince formats a time duration in a human-readable way
func formatTimeSince(t time.Time) string {
	dur := time.Since(t)
	if dur < time.Minute {
		return fmt.Sprintf("%ds", int(dur.Seconds()))
	}
	if dur < time.Hour {
		return fmt.Sprintf("%dm", int(dur.Minutes()))
	}
	if dur < 24*time.Hour {
		return fmt.Sprintf("%dh", int(dur.Hours()))
	}
	return fmt.Sprintf("%dd", int(dur.Hours()/24))
}
