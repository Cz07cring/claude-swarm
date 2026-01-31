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
	rows           int
	width          int
	height         int
}

// NewAgentGridView creates a new agent grid view
func NewAgentGridView(width, height int) *AgentGridView {
	return &AgentGridView{
		agents:        make([]*models.AgentStatus, 0),
		selectedIndex: 0,
		cols:          3, // Default 3 columns
		rows:          3, // Default 3 rows
		width:         width,
		height:        height,
	}
}

// calculateOptimalGrid calculates the optimal grid dimensions based on agent count
func (v *AgentGridView) calculateOptimalGrid() (cols, rows int) {
	numAgents := len(v.agents)
	if numAgents == 0 {
		return 3, 3
	}

	// Determine optimal grid size based on agent count
	switch {
	case numAgents <= 4:
		return 2, 2
	case numAgents <= 9:
		return 3, 3
	case numAgents <= 12:
		return 4, 3
	case numAgents <= 16:
		return 4, 4
	case numAgents <= 20:
		return 5, 4
	default:
		// For more than 20 agents, calculate dynamically
		cols = 5
		rows = (numAgents + cols - 1) / cols
		return cols, rows
	}
}

// Update updates the agent list
func (v *AgentGridView) Update(agents []*models.AgentStatus) {
	v.agents = agents

	// Recalculate optimal grid dimensions
	v.cols, v.rows = v.calculateOptimalGrid()

	// Ensure selected index is valid
	if v.selectedIndex >= len(v.agents) && len(v.agents) > 0 {
		v.selectedIndex = len(v.agents) - 1
	}
	if v.selectedIndex < 0 {
		v.selectedIndex = 0
	}
}

// MoveUp moves selection up in the grid with wrap-around
func (v *AgentGridView) MoveUp() {
	if len(v.agents) == 0 {
		return
	}
	newIndex := v.selectedIndex - v.cols
	if newIndex >= 0 {
		v.selectedIndex = newIndex
	} else {
		// Wrap to bottom
		col := v.selectedIndex % v.cols
		lastRow := (len(v.agents) - 1) / v.cols
		newIndex = lastRow*v.cols + col
		if newIndex >= len(v.agents) {
			// If wrapped position doesn't exist, go to last agent in that column
			newIndex = ((len(v.agents)-1)/v.cols)*v.cols + col
			if newIndex >= len(v.agents) {
				newIndex = len(v.agents) - 1
			}
		}
		v.selectedIndex = newIndex
	}
}

// MoveDown moves selection down in the grid with wrap-around
func (v *AgentGridView) MoveDown() {
	if len(v.agents) == 0 {
		return
	}
	newIndex := v.selectedIndex + v.cols
	if newIndex < len(v.agents) {
		v.selectedIndex = newIndex
	} else {
		// Wrap to top of same column
		col := v.selectedIndex % v.cols
		v.selectedIndex = col
	}
}

// MoveLeft moves selection left in the grid with wrap-around
func (v *AgentGridView) MoveLeft() {
	if len(v.agents) == 0 {
		return
	}
	if v.selectedIndex%v.cols > 0 {
		v.selectedIndex--
	} else {
		// Wrap to end of row
		rowStart := (v.selectedIndex / v.cols) * v.cols
		rowEnd := rowStart + v.cols - 1
		if rowEnd >= len(v.agents) {
			rowEnd = len(v.agents) - 1
		}
		v.selectedIndex = rowEnd
	}
}

// MoveRight moves selection right in the grid with wrap-around
func (v *AgentGridView) MoveRight() {
	if len(v.agents) == 0 {
		return
	}
	if v.selectedIndex%v.cols < v.cols-1 && v.selectedIndex+1 < len(v.agents) {
		v.selectedIndex++
	} else {
		// Wrap to start of row
		rowStart := (v.selectedIndex / v.cols) * v.cols
		v.selectedIndex = rowStart
	}
}

// MoveToFirst moves selection to the first agent (Home key)
func (v *AgentGridView) MoveToFirst() {
	if len(v.agents) > 0 {
		v.selectedIndex = 0
	}
}

// MoveToLast moves selection to the last agent (End key)
func (v *AgentGridView) MoveToLast() {
	if len(v.agents) > 0 {
		v.selectedIndex = len(v.agents) - 1
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
			Render("暂无 Agent")
	}

	// Calculate cell dimensions dynamically
	cellWidth := (v.width - (v.cols - 1)) / v.cols
	if cellWidth < 15 {
		cellWidth = 15 // Minimum width
	}

	// Determine if we should use compact mode
	compactMode := len(v.agents) > 12 || cellWidth < 20

	var rows []string
	for i := 0; i < len(v.agents); i += v.cols {
		var cells []string
		for j := 0; j < v.cols && i+j < len(v.agents); j++ {
			agentIdx := i + j
			agent := v.agents[agentIdx]
			selected := isActive && agentIdx == v.selectedIndex
			cell := v.renderAgentCell(agent, selected, cellWidth, compactMode)
			cells = append(cells, cell)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// renderAgentCell renders a single agent cell with dynamic sizing
func (v *AgentGridView) renderAgentCell(agent *models.AgentStatus, selected bool, cellWidth int, compact bool) string {
	// State indicator with emoji
	var stateIcon string
	var stateText string
	var cellStyle lipgloss.Style
	switch agent.State {
	case models.AgentStateIdle:
		stateIcon = "😴"
		stateText = "空闲"
		cellStyle = agentCellIdleStyle
	case models.AgentStateWorking:
		stateIcon = "🚀"
		stateText = "工作中"
		cellStyle = agentCellWorkingStyle
	case models.AgentStateWaitingConfirm:
		stateIcon = "⏸️"
		stateText = "等待"
		cellStyle = agentCellWaitingStyle
	case models.AgentStateError:
		stateIcon = "❌"
		stateText = "错误"
		cellStyle = agentCellErrorStyle
	case models.AgentStateStuck:
		stateIcon = "⚠️"
		stateText = "卡住"
		cellStyle = agentCellErrorStyle
	default:
		stateIcon = "❓"
		stateText = "未知"
		cellStyle = agentCellStyle
	}

	// Format agent ID based on available width
	agentID := agent.AgentID
	maxIDLen := cellWidth - 6
	if maxIDLen < 6 {
		maxIDLen = 6
	}
	if len(agentID) > maxIDLen {
		agentID = agentID[:maxIDLen]
	}

	// Format agent info with better layout
	var info strings.Builder

	if compact {
		// Compact mode: single line
		info.WriteString(fmt.Sprintf("%s %s", stateIcon, agentID))
	} else {
		// Full mode: multi-line
		info.WriteString(fmt.Sprintf("%s %s\n", stateIcon, agentID))
		info.WriteString(lipgloss.NewStyle().Faint(true).Render(stateText))

		// Add task info if working
		if agent.CurrentTask != nil {
			info.WriteString("\n")
			taskDesc := agent.CurrentTask.Description
			maxTaskLen := cellWidth - 4
			if len(taskDesc) > maxTaskLen {
				taskDesc = taskDesc[:maxTaskLen-3] + "..."
			}
			info.WriteString(lipgloss.NewStyle().
				Foreground(colorInfo).
				Faint(true).
				Render(taskDesc))
		}
	}

	// Calculate cell height based on compact mode
	cellHeight := 3
	if !compact {
		cellHeight = 5
	}

	// Apply cell dimensions
	cellStyle = cellStyle.Width(cellWidth).Height(cellHeight)

	// Highlight selected agent with special border
	if selected {
		cellStyle = cellStyle.
			BorderForeground(colorHighlight).
			BorderStyle(lipgloss.DoubleBorder()).
			Bold(true)
	}

	return cellStyle.Render(info.String())
}
