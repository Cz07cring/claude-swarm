package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/claude-swarm/internal/models"
)

// LogViewerView displays logs for a selected agent
type LogViewerView struct {
	agent  *models.AgentStatus
	width  int
	height int
}

// NewLogViewerView creates a new log viewer
func NewLogViewerView(width, height int) *LogViewerView {
	return &LogViewerView{
		width:  width,
		height: height,
	}
}

// Update updates the displayed agent
func (v *LogViewerView) Update(agent *models.AgentStatus) {
	v.agent = agent
}

// Render renders the log viewer
func (v *LogViewerView) Render() string {
	if v.agent == nil {
		return lipgloss.NewStyle().
			Width(v.width).
			Height(v.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Select an agent to view logs")
	}

	var content strings.Builder

	// Agent header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary).
		Render(v.agent.AgentID)
	content.WriteString(header)
	content.WriteString("\n")

	// State
	var stateStyle lipgloss.Style
	switch v.agent.State {
	case models.AgentStateIdle:
		stateStyle = statusIdleStyle
	case models.AgentStateWorking:
		stateStyle = statusWorkingStyle
	case models.AgentStateWaitingConfirm:
		stateStyle = statusWaitingStyle
	case models.AgentStateError, models.AgentStateStuck:
		stateStyle = statusErrorStyle
	default:
		stateStyle = baseStyle
	}
	content.WriteString(stateStyle.Render(string(v.agent.State)))
	content.WriteString("\n\n")

	// Current task
	if v.agent.CurrentTask != nil {
		content.WriteString(lipgloss.NewStyle().Foreground(colorText).Bold(true).Render("Task:"))
		content.WriteString("\n")
		content.WriteString(v.agent.CurrentTask.Description)
		content.WriteString("\n\n")
	}

	// Output/logs
	if v.agent.Output != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(colorText).Bold(true).Render("Output:"))
		content.WriteString("\n")

		// Split output into lines and show last N lines that fit
		lines := strings.Split(v.agent.Output, "\n")

		// Calculate how many lines we can show
		headerLines := 4 // agent header + state + blank line
		if v.agent.CurrentTask != nil {
			// Count lines in task description
			taskLines := len(strings.Split(v.agent.CurrentTask.Description, "\n"))
			headerLines += 2 + taskLines // "Task:" + description + blank line
		}
		headerLines += 2 // "Output:" + blank line

		availableLines := v.height - headerLines
		if availableLines < 1 {
			availableLines = 1
		}

		// Show last N lines
		start := len(lines) - availableLines
		if start < 0 {
			start = 0
		}

		for _, line := range lines[start:] {
			// Truncate long lines
			if len(line) > v.width-4 {
				line = line[:v.width-7] + "..."
			}
			content.WriteString(logLineStyle.Render(line))
			content.WriteString("\n")
		}
	} else {
		content.WriteString(lipgloss.NewStyle().
			Faint(true).
			Render("No output available"))
	}

	return content.String()
}
