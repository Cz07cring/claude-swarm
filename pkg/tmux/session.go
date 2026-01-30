package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

// NewSession creates a new tmux session
func NewSession(name string) (*Session, error) {
	// Check if session already exists
	cmd := exec.Command("tmux", "has-session", "-t", name)
	if err := cmd.Run(); err == nil {
		// Session exists, kill it first
		killCmd := exec.Command("tmux", "kill-session", "-t", name)
		_ = killCmd.Run()
	}

	// Create new session (detached)
	cmd = exec.Command("tmux", "new-session", "-d", "-s", name)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	session := &Session{
		Name:  name,
		Panes: []*Pane{},
	}

	// Get the first pane (created automatically)
	pane, err := session.getPaneByIndex(0)
	if err != nil {
		return nil, err
	}
	session.Panes = append(session.Panes, pane)

	return session, nil
}

// SplitPane splits the current pane
func (s *Session) SplitPane(horizontal bool) (*Pane, error) {
	flag := "-v" // vertical split (side by side)
	if !horizontal {
		flag = "-h" // horizontal split (top/bottom)
	}

	cmd := exec.Command("tmux", "split-window", flag, "-t", s.Name)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to split pane: %w", err)
	}

	// Get the newly created pane
	paneIndex := len(s.Panes)
	pane, err := s.getPaneByIndex(paneIndex)
	if err != nil {
		return nil, err
	}

	s.Panes = append(s.Panes, pane)
	return pane, nil
}

// getPaneByIndex gets a pane by its index
func (s *Session) getPaneByIndex(index int) (*Pane, error) {
	// Get pane ID using tmux list-panes
	cmd := exec.Command("tmux", "list-panes", "-t", s.Name, "-F", "#{pane_id}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list panes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if index >= len(lines) {
		return nil, fmt.Errorf("pane index %d out of range", index)
	}

	paneID := strings.TrimSpace(lines[index])
	return &Pane{
		ID:    paneID,
		Index: index,
	}, nil
}

// Kill terminates the tmux session
func (s *Session) Kill() error {
	cmd := exec.Command("tmux", "kill-session", "-t", s.Name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill session: %w", err)
	}
	return nil
}

// GetPaneCount returns the number of panes in the session
func (s *Session) GetPaneCount() int {
	return len(s.Panes)
}

// GetPane returns a pane by index
func (s *Session) GetPane(index int) (*Pane, error) {
	if index < 0 || index >= len(s.Panes) {
		return nil, fmt.Errorf("pane index %d out of range", index)
	}
	return s.Panes[index], nil
}
