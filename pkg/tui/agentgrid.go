package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/claude-swarm/internal/models"
)

// AgentGridView renders agents in a grid layout
type AgentGridView struct {
	agents         []*models.AgentStatus
	selectedIndex  int
	cols           int
	width          int
	height         int
}

// NewAgentGridView creates a new agent grid view
func NewAgentGridView(width, height int) *AgentGridView {
	return &AgentGridView{
		agents:        make([]*models.AgentStatus, 0),
		selectedIndex: 0,
		cols:          3, // 3x3 grid
		width:         width,
		height:        height,
	}
}

// Update updates the agent list
func (v *AgentGridView) Update(agents []*models.AgentStatus) {
	v.agents = agents
	// Ensure selected index is valid
	if v.selectedIndex >= len(v.agents) && len(v.agents) > 0 {
		v.selectedIndex = len(v.agents) - 1
	}
	if v.selectedIndex < 0 {
		v.selectedIndex = 0
	}
}

// MoveUp moves selection up in the grid
func (v *AgentGridView) MoveUp() {
	newIndex := v.selectedIndex - v.cols
	if newIndex >= 0 {
		v.selectedIndex = newIndex
	}
}

// MoveDown moves selection down in the grid
func (v *AgentGridView) MoveDown() {
	newIndex := v.selectedIndex + v.cols
	if newIndex < len(v.agents) {
		v.selectedIndex = newIndex
	}
}

// MoveLeft moves selection left in the grid
func (v *AgentGridView) MoveLeft() {
	if v.selectedIndex%v.cols > 0 {
		v.selectedIndex--
	}
}

// MoveRight moves selection right in the grid
func (v *AgentGridView) MoveRight() {
	if v.selectedIndex%v.cols < v.cols-1 && v.selectedIndex+1 < len(v.agents) {
		v.selectedIndex++
	}
}

// GetSelectedAgent returns the currently selected agent
func (v *AgentGridView) GetSelectedAgent() *models.AgentStatus {
	if v.selectedIndex >= 0 && v.selectedIndex < len(v.agents) {
		return v.agents[v.selectedIndex]
	}
	return nil
}

// Render renders the agent grid view
func (v *AgentGridView) Render(isActive bool) string {
	if len(v.agents) == 0 {
		return lipgloss.NewStyle().
			Width(v.width).
			Height(v.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No agents")
	}

	var rows []string
	for i := 0; i < len(v.agents); i += v.cols {
		var cells []string
		for j := 0; j < v.cols && i+j < len(v.agents); j++ {
			agentIdx := i + j
			agent := v.agents[agentIdx]
			selected := isActive && agentIdx == v.selectedIndex
			cell := v.renderAgentCell(agent, selected)
			cells = append(cells, cell)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// renderAgentCell renders a single agent cell
func (v *AgentGridView) renderAgentCell(agent *models.AgentStatus, selected bool) string {
	// State indicator
	var stateIcon string
	var cellStyle lipgloss.Style
	switch agent.State {
	case models.AgentStateIdle:
		stateIcon = "○"
		cellStyle = agentCellIdleStyle
	case models.AgentStateWorking:
		stateIcon = "●"
		cellStyle = agentCellWorkingStyle
	case models.AgentStateWaitingConfirm:
		stateIcon = "?"
		cellStyle = agentCellWaitingStyle
	case models.AgentStateError:
		stateIcon = "✗"
		cellStyle = agentCellErrorStyle
	case models.AgentStateStuck:
		stateIcon = "⏸"
		cellStyle = agentCellErrorStyle
	default:
		stateIcon = "?"
		cellStyle = agentCellStyle
	}

	// Format agent info
	info := fmt.Sprintf("%s %s\n%s", stateIcon, agent.AgentID, string(agent.State))

	// Add task info if working
	if agent.CurrentTask != nil {
		taskID := agent.CurrentTask.ID
		if len(taskID) > 10 {
			taskID = taskID[:10] + "..."
		}
		info += fmt.Sprintf("\n%s", taskID)
	}

	// Highlight selected agent
	if selected {
		cellStyle = cellStyle.BorderForeground(colorHighlight).Bold(true)
	}

	return cellStyle.Render(info)
}
