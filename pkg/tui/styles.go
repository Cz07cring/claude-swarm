// Package tui provides terminal user interface components for Claude Agent Swarm monitoring.
// It uses Bubble Tea framework for TUI implementation and Lipgloss for styling.
package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette defines the color scheme used across all TUI components.
	// All colors are specified using 256-color palette codes.

	// colorPrimary is the main accent color (Cyan #86) used for:
	// - Active panel borders
	// - Title text
	// - Selected agent highlights
	colorPrimary = lipgloss.Color("86") // Cyan

	// colorSuccess indicates successful or completed states (Green #42):
	// - Completed tasks (✓ icon)
	colorSuccess = lipgloss.Color("42") // Green

	// colorWarning indicates warning or pending states (Yellow #226):
	// - Agents waiting for user confirmation
	colorWarning = lipgloss.Color("226") // Yellow

	// colorError indicates error or failed states (Red #196):
	// - Failed tasks (✗ icon)
	// - Agents in error or stuck state
	colorError = lipgloss.Color("196") // Red

	// colorIdle indicates idle or pending states (Gray #240):
	// - Pending tasks (○ icon)
	// - Idle agents
	colorIdle = lipgloss.Color("240") // Gray

	// colorWorking indicates active working states (Blue #39):
	// - Tasks in progress (● icon)
	// - Working agents
	colorWorking = lipgloss.Color("39") // Blue

	// colorBorder is the default border color for inactive panels (Dark gray #238)
	colorBorder = lipgloss.Color("238") // Dark gray

	// colorText is the default text color for most content (Light gray #252)
	colorText = lipgloss.Color("252") // Light gray

	// colorHighlight is used for selected items (Pink #219):
	// - Selected agent cell borders
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
