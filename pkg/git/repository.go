package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// NewRepository creates a new Repository instance and verifies it's a valid git repo
func NewRepository(path string) (*Repository, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify it's a git repository
	cmd := exec.Command("git", "-C", absPath, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return nil, ErrNotGitRepo
	}

	return &Repository{Path: absPath}, nil
}

// GetCurrentBranch returns the current branch name
func (r *Repository) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "-C", r.Path, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// IsClean checks if the repository has uncommitted changes
func (r *Repository) IsClean() (bool, error) {
	cmd := exec.Command("git", "-C", r.Path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check status: %w", err)
	}

	return len(strings.TrimSpace(string(output))) == 0, nil
}

// GetCurrentCommit returns the current commit hash
func (r *Repository) GetCurrentCommit() (string, error) {
	cmd := exec.Command("git", "-C", r.Path, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
