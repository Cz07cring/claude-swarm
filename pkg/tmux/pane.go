package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

// Capture captures the pane output
func (p *Pane) Capture() (string, error) {
	cmd := exec.Command("tmux", "capture-pane", "-p", "-t", p.ID, "-S", "-100")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to capture pane: %w", err)
	}
	return string(output), nil
}

// SendKeys sends keys to the pane
func (p *Pane) SendKeys(keys string) error {
	p.sendMu.Lock()
	defer p.sendMu.Unlock()

	cmd := exec.Command("tmux", "send-keys", "-t", p.ID, keys)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys: %w", err)
	}
	return nil
}

// SendLine sends a line (with Enter) to the pane
// Uses a single tmux command to send both text and Enter atomically
func (p *Pane) SendLine(line string) error {
	p.sendMu.Lock()
	defer p.sendMu.Unlock()

	// Use single tmux command to send text and Enter atomically
	cmd := exec.Command("tmux", "send-keys", "-t", p.ID, line, "Enter")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send line: %w", err)
	}
	return nil
}

// Clear clears the pane
func (p *Pane) Clear() error {
	cmd := exec.Command("tmux", "send-keys", "-t", p.ID, "C-l")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clear pane: %w", err)
	}
	return nil
}

// GetLastLine returns the last non-empty line from the pane
func (p *Pane) GetLastLine() (string, error) {
	output, err := p.Capture()
	if err != nil {
		return "", err
	}

	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line, nil
		}
	}

	return "", nil
}
