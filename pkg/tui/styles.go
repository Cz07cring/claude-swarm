// Package tui provides terminal user interface components for Claude Agent Swarm monitoring.
// It uses Bubble Tea framework for TUI implementation and Lipgloss for styling.
package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette defines the color scheme used across all TUI components.
	// All colors are specified using 256-color palette codes.

	// colorPrimary is the main accent color (Bright Cyan #51) used for:
	// - Active panel borders
	// - Title text
	// - Selected agent highlights
	colorPrimary = lipgloss.Color("51") // Bright Cyan

	// colorSuccess indicates successful or completed states (Bright Green #46):
	// - Completed tasks (✓ icon)
	colorSuccess = lipgloss.Color("46") // Bright Green

	// colorWarning indicates warning or pending states (Bright Yellow #226):
	// - Agents waiting for user confirmation
	colorWarning = lipgloss.Color("226") // Yellow

	// colorError indicates error or failed states (Bright Red #196):
	// - Failed tasks (✗ icon)
	// - Agents in error or stuck state
	colorError = lipgloss.Color("196") // Red

	// colorIdle indicates idle or pending states (Gray #243):
	// - Pending tasks (○ icon)
	// - Idle agents
	colorIdle = lipgloss.Color("243") // Light Gray

	// colorWorking indicates active working states (Bright Blue #33):
	// - Tasks in progress (● icon)
	// - Working agents
	colorWorking = lipgloss.Color("33") // Bright Blue

	// colorBorder is the default border color for inactive panels (Dark gray #238)
	colorBorder = lipgloss.Color("238") // Dark gray

	// colorText is the default text color for most content (Light gray #252)
	colorText = lipgloss.Color("252") // Light gray

	// colorHighlight is used for selected items (Bright Magenta #201):
	// - Selected agent cell borders
	colorHighlight = lipgloss.Color("201") // Bright Magenta

	// colorInfo is used for informational text (Soft blue #111)
	colorInfo = lipgloss.Color("111") // Soft Blue

	// colorMuted is used for less important text (Dark gray #245)
	colorMuted = lipgloss.Color("245") // Dark Gray

	// Base styles

	// baseStyle is the default style applied to most text elements.
	// It provides consistent text color across the UI.
	baseStyle = lipgloss.NewStyle().
		Foreground(colorText)

	// Panel styles define the appearance of the three main panels (Tasks, Agents, Logs)

	// panelStyle is used for inactive panels with a dark gray border
	panelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1)

	// activePanelStyle is used for the currently focused panel with cyan border.
	// The active panel is indicated by the ◄ symbol in its title.
	activePanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(0, 1)

	// Title styles

	// titleStyle is used for the main dashboard title at the top of the screen
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary).
		MarginBottom(1)

	// Status indicator styles define the appearance of status icons and text

	// statusIdleStyle is used for idle/pending states (gray)
	// Applied to: ○ icon for pending tasks, idle agent states
	statusIdleStyle = lipgloss.NewStyle().
		Foreground(colorIdle).
		Bold(true)

	// statusWorkingStyle is used for active/working states (blue)
	// Applied to: ● icon for in-progress tasks, working agent states
	statusWorkingStyle = lipgloss.NewStyle().
		Foreground(colorWorking).
		Bold(true)

	// statusWaitingStyle is used for waiting/warning states (yellow)
	// Applied to: ? icon for agents waiting for user confirmation
	statusWaitingStyle = lipgloss.NewStyle().
		Foreground(colorWarning).
		Bold(true)

	// statusErrorStyle is used for error/failed states (red)
	// Applied to: ✗ icon for failed tasks, error/stuck agent states
	statusErrorStyle = lipgloss.NewStyle().
		Foreground(colorError).
		Bold(true)

	// statusSuccessStyle is used for completed/success states (green)
	// Applied to: ✓ icon for completed tasks
	statusSuccessStyle = lipgloss.NewStyle().
		Foreground(colorSuccess).
		Bold(true)

	// Task list styles define the appearance of task items in the left panel

	// taskItemStyle is the default style for unselected task items
	taskItemStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		MarginBottom(0)

	// taskSelectedStyle is used for the currently selected task item.
	// The selected task has a cyan background and is displayed in bold.
	taskSelectedStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		Background(colorPrimary).
		Foreground(lipgloss.Color("0")).
		Bold(true)

	// Agent grid styles define the appearance of agent cells in the 3x3 grid

	// agentCellStyle is the base style for agent cells (neutral state)
	agentCellStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		Width(20).
		Height(4).
		Padding(0, 1).
		Align(lipgloss.Center, lipgloss.Center)

	// agentCellIdleStyle is used for idle agents (gray border)
	// Indicates the agent is not currently assigned to any task
	agentCellIdleStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorIdle).
		Width(20).
		Height(4).
		Padding(0, 1).
		Align(lipgloss.Center, lipgloss.Center)

	// agentCellWorkingStyle is used for working agents (blue border)
	// Indicates the agent is actively executing a task
	agentCellWorkingStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorWorking).
		Width(20).
		Height(4).
		Padding(0, 1).
		Align(lipgloss.Center, lipgloss.Center)

	// agentCellWaitingStyle is used for agents waiting for confirmation (yellow border)
	// Indicates the agent needs user input to proceed
	agentCellWaitingStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorWarning).
		Width(20).
		Height(4).
		Padding(0, 1).
		Align(lipgloss.Center, lipgloss.Center)

	// agentCellErrorStyle is used for agents in error or stuck state (red border)
	// Indicates the agent encountered an error or is unresponsive
	agentCellErrorStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorError).
		Width(20).
		Height(4).
		Padding(0, 1).
		Align(lipgloss.Center, lipgloss.Center)

	// Log viewer styles

	// logLineStyle is used for individual log lines in the right panel.
	// The faint attribute makes logs visually distinct from headers.
	logLineStyle = lipgloss.NewStyle().
		Foreground(colorText).
		Faint(true)

	// Help text styles

	// helpStyle is used for the help text bar at the bottom of the screen.
	// Displays keyboard shortcuts and usage hints.
	helpStyle = lipgloss.NewStyle().
		Foreground(colorMuted).
		Faint(true).
		MarginTop(1)

	// Status bar styles

	// statusBarStyle is used for the top status bar showing cluster metrics
	statusBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Foreground(colorText).
		Padding(0, 1).
		MarginBottom(1)

	// metricLabelStyle is used for metric labels in the status bar
	metricLabelStyle = lipgloss.NewStyle().
		Foreground(colorMuted).
		Bold(false)

	// metricValueStyle is used for metric values in the status bar
	metricValueStyle = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	// metricSuccessStyle is used for positive metrics
	metricSuccessStyle = lipgloss.NewStyle().
		Foreground(colorSuccess).
		Bold(true)

	// metricWarningStyle is used for warning metrics
	metricWarningStyle = lipgloss.NewStyle().
		Foreground(colorWarning).
		Bold(true)

	// metricErrorStyle is used for error metrics
	metricErrorStyle = lipgloss.NewStyle().
		Foreground(colorError).
		Bold(true)

	// Progress indicator styles

	// progressBarBg is the background style for progress bars
	progressBarBg = lipgloss.NewStyle().
		Foreground(colorBorder)

	// progressBarFg is the foreground style for progress bars
	progressBarFg = lipgloss.NewStyle().
		Foreground(colorWorking)

	// Badge styles

	// badgeStyle is the base style for status badges
	badgeStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true)

	// badgePendingStyle is for pending status badges
	badgePendingStyle = badgeStyle.Copy().
		Background(colorIdle).
		Foreground(lipgloss.Color("0"))

	// badgeWorkingStyle is for working status badges
	badgeWorkingStyle = badgeStyle.Copy().
		Background(colorWorking).
		Foreground(lipgloss.Color("0"))

	// badgeSuccessStyle is for success status badges
	badgeSuccessStyle = badgeStyle.Copy().
		Background(colorSuccess).
		Foreground(lipgloss.Color("0"))

	// badgeErrorStyle is for error status badges
	badgeErrorStyle = badgeStyle.Copy().
		Background(colorError).
		Foreground(lipgloss.Color("0"))
)
