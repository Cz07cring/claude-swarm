package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/claude-swarm/internal/models"
)

// LogViewerView displays logs for a selected agent
type LogViewerView struct {
	agent        *models.AgentStatus
	width        int
	height       int
	scrollOffset int  // Current scroll position
	autoScroll   bool // Auto-scroll to bottom
}

// NewLogViewerView creates a new log viewer
func NewLogViewerView(width, height int) *LogViewerView {
	return &LogViewerView{
		width:        width,
		height:       height,
		scrollOffset: 0,
		autoScroll:   true, // Auto-scroll enabled by default
	}
}

// Update updates the displayed agent
func (v *LogViewerView) Update(agent *models.AgentStatus) {
	v.agent = agent
	// Reset scroll when switching agents
	v.scrollOffset = 0
}

// ScrollUp scrolls the log view up
func (v *LogViewerView) ScrollUp(lines int) {
	v.scrollOffset -= lines
	if v.scrollOffset < 0 {
		v.scrollOffset = 0
	}
	v.autoScroll = false
}

// ScrollDown scrolls the log view down
func (v *LogViewerView) ScrollDown(lines int) {
	v.scrollOffset += lines
	v.autoScroll = false
}

// ScrollToTop scrolls to the top of logs
func (v *LogViewerView) ScrollToTop() {
	v.scrollOffset = 0
	v.autoScroll = false
}

// ScrollToBottom scrolls to the bottom of logs
func (v *LogViewerView) ScrollToBottom() {
	v.scrollOffset = 999999 // Will be clamped in Render
	v.autoScroll = true
}

// ToggleAutoScroll toggles auto-scroll mode
func (v *LogViewerView) ToggleAutoScroll() {
	v.autoScroll = !v.autoScroll
	if v.autoScroll {
		v.scrollOffset = 999999
	}
}

// Render renders the log viewer with scroll support
func (v *LogViewerView) Render() string {
	if v.agent == nil {
		return lipgloss.NewStyle().
			Width(v.width).
			Height(v.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("选择一个 Agent 查看日志")
	}

	var content strings.Builder

	// Agent header with status indicator
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary).
		Render(fmt.Sprintf("📌 %s", v.agent.AgentID))
	content.WriteString(header)
	content.WriteString(" ")

	// State badge
	var stateStyle lipgloss.Style
	var stateBadge string
	switch v.agent.State {
	case models.AgentStateIdle:
		stateStyle = badgePendingStyle
		stateBadge = "空闲"
	case models.AgentStateWorking:
		stateStyle = badgeWorkingStyle
		stateBadge = "工作中"
	case models.AgentStateWaitingConfirm:
		stateStyle = badgeStyle.Copy().Background(colorWarning).Foreground(lipgloss.Color("0"))
		stateBadge = "等待"
	case models.AgentStateError, models.AgentStateStuck:
		stateStyle = badgeErrorStyle
		stateBadge = "错误"
	default:
		stateStyle = badgeStyle
		stateBadge = string(v.agent.State)
	}
	content.WriteString(stateStyle.Render(stateBadge))
	content.WriteString("\n")
	content.WriteString(strings.Repeat("─", v.width))
	content.WriteString("\n")

	headerLines := 3 // Header + state + separator

	// Current task
	if v.agent.CurrentTask != nil {
		taskLabel := lipgloss.NewStyle().Foreground(colorInfo).Bold(true).Render("📋 任务:")
		content.WriteString(taskLabel)
		content.WriteString("\n")

		taskDesc := v.agent.CurrentTask.Description
		// Wrap long task descriptions
		wrappedDesc := wrapText(taskDesc, v.width-2)
		content.WriteString(lipgloss.NewStyle().Foreground(colorText).Render(wrappedDesc))
		content.WriteString("\n")
		content.WriteString(strings.Repeat("─", v.width))
		content.WriteString("\n")

		headerLines += 3 + strings.Count(wrappedDesc, "\n")
	}

	// Output/logs with scroll support
	if v.agent.Output != "" {
		logLabel := lipgloss.NewStyle().Foreground(colorInfo).Bold(true).Render("📜 输出:")
		scrollIndicator := ""
		if !v.autoScroll {
			scrollIndicator = lipgloss.NewStyle().Foreground(colorMuted).Render(" [滚动模式]")
		}
		content.WriteString(logLabel)
		content.WriteString(scrollIndicator)
		content.WriteString("\n")
		headerLines++

		// Split output into lines
		lines := strings.Split(v.agent.Output, "\n")
		totalLines := len(lines)

		// Calculate available lines for logs
		availableLines := v.height - headerLines - 1
		if availableLines < 1 {
			availableLines = 1
		}

		// Apply scroll offset with auto-scroll
		start := v.scrollOffset
		if v.autoScroll {
			start = totalLines - availableLines
			if start < 0 {
				start = 0
			}
			v.scrollOffset = start
		} else {
			// Clamp scroll offset
			maxScroll := totalLines - availableLines
			if maxScroll < 0 {
				maxScroll = 0
			}
			if start > maxScroll {
				start = maxScroll
				v.scrollOffset = start
			}
		}

		end := start + availableLines
		if end > totalLines {
			end = totalLines
		}

		// Render visible log lines
		for i, line := range lines[start:end] {
			lineNum := start + i + 1
			lineNumStr := lipgloss.NewStyle().
				Foreground(colorMuted).
				Render(fmt.Sprintf("%4d│ ", lineNum))

			// Truncate long lines
			maxLineLen := v.width - 8
			if maxLineLen > 11 && len(line) > maxLineLen {
				line = line[:v.width-11] + "..."
			} else if maxLineLen > 0 && len(line) > maxLineLen {
				line = line[:maxLineLen]
			}

			content.WriteString(lineNumStr)
			content.WriteString(logLineStyle.Render(line))
			content.WriteString("\n")
		}

		// Show scroll position indicator
		if totalLines > availableLines {
			scrollPos := lipgloss.NewStyle().
				Foreground(colorMuted).
				Faint(true).
				Render(fmt.Sprintf("─ %d-%d / %d 行 ─", start+1, end, totalLines))
			content.WriteString(scrollPos)
		}
	} else {
		content.WriteString(lipgloss.NewStyle().
			Foreground(colorMuted).
			Faint(true).
			Render("暂无输出"))
	}

	return content.String()
}

// wrapText wraps text to fit within the specified width
func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 <= width {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		} else {
			if currentLine != "" {
				result.WriteString(currentLine)
				result.WriteString("\n")
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		result.WriteString(currentLine)
	}

	return result.String()
}
