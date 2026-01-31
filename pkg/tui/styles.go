package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette
	colorPrimary   = lipgloss.Color("86")  // Cyan
	colorSuccess   = lipgloss.Color("42")  // Green
	colorWarning   = lipgloss.Color("226") // Yellow
	colorError     = lipgloss.Color("196") // Red
	colorIdle      = lipgloss.Color("240") // Gray
	colorWorking   = lipgloss.Color("39")  // Blue
	colorBorder    = lipgloss.Color("238") // Dark gray
	colorText      = lipgloss.Color("252") // Light gray
	colorHighlight = lipgloss.Color("219") // Pink

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Foreground(colorText)

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1)

	// Title styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	// Status indicator styles
	statusIdleStyle = lipgloss.NewStyle().
			Foreground(colorIdle).
			Bold(true)

	statusWorkingStyle = lipgloss.NewStyle().
				Foreground(colorWorking).
				Bold(true)

	statusWaitingStyle = lipgloss.NewStyle().
				Foreground(colorWarning).
				Bold(true)

	statusErrorStyle = lipgloss.NewStyle().
				Foreground(colorError).
				Bold(true)

	statusSuccessStyle = lipgloss.NewStyle().
				Foreground(colorSuccess).
				Bold(true)

	// Task list styles
	taskItemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			MarginBottom(0)

	taskSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Background(colorPrimary).
				Foreground(lipgloss.Color("0")).
				Bold(true)

	// Agent grid styles
	agentCellStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Width(20).
			Height(4).
			Padding(0, 1).
			Align(lipgloss.Center, lipgloss.Center)

	agentCellIdleStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorIdle).
				Width(20).
				Height(4).
				Padding(0, 1).
				Align(lipgloss.Center, lipgloss.Center)

	agentCellWorkingStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorWorking).
				Width(20).
				Height(4).
				Padding(0, 1).
				Align(lipgloss.Center, lipgloss.Center)

	agentCellWaitingStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorWarning).
				Width(20).
				Height(4).
				Padding(0, 1).
				Align(lipgloss.Center, lipgloss.Center)

	agentCellErrorStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorError).
				Width(20).
				Height(4).
				Padding(0, 1).
				Align(lipgloss.Center, lipgloss.Center)

	// Log viewer styles
	logLineStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Faint(true)

	// Help text styles
	helpStyle = lipgloss.NewStyle().
			Foreground(colorIdle).
			Faint(true).
			MarginTop(1)
)
