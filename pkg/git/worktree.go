package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// WorktreeManager manages Git worktrees for agents
type WorktreeManager struct {
	repo   *Repository
	config WorktreeConfig
}

// NewWorktreeManager creates a new WorktreeManager
func NewWorktreeManager(config WorktreeConfig) (*WorktreeManager, error) {
	repo, err := NewRepository(config.BaseRepoPath)
	if err != nil {
		return nil, err
	}

	// Set defaults
	if config.WorktreeRootDir == "" {
		config.WorktreeRootDir = ".worktrees"
	}
	if config.BaseBranch == "" {
		config.BaseBranch = "main"
	}

	return &WorktreeManager{
		repo:   repo,
		config: config,
	}, nil
}

// CreateWorktree creates a new worktree for an agent
func (wm *WorktreeManager) CreateWorktree(agentID string) (*Worktree, error) {
	branchName := fmt.Sprintf("agent-%s-branch", agentID)
	worktreePath := filepath.Join(wm.repo.Path, wm.config.WorktreeRootDir,
		fmt.Sprintf("agent-%s", agentID))

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		return nil, ErrWorktreeExists
	}

	// Create worktree directory parent if it doesn't exist
	worktreeRoot := filepath.Join(wm.repo.Path, wm.config.WorktreeRootDir)
	if err := os.MkdirAll(worktreeRoot, 0755); err != nil {
		return nil, fmt.Errorf("failed to create worktree root directory: %w", err)
	}

	// git worktree add -b agent-X-branch .worktrees/agent-X main
	cmd := exec.Command("git", "-C", wm.repo.Path, "worktree", "add",
		"-b", branchName, worktreePath, wm.config.BaseBranch)

	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w, output: %s", err, string(output))
	}

	return &Worktree{
		Path:       worktreePath,
		BranchName: branchName,
		AgentID:    agentID,
		CreatedAt:  time.Now(),
	}, nil
}

// RemoveWorktree removes a worktree and its associated branch
func (wm *WorktreeManager) RemoveWorktree(agentID string) error {
	branchName := fmt.Sprintf("agent-%s-branch", agentID)
	worktreePath := filepath.Join(wm.repo.Path, wm.config.WorktreeRootDir,
		fmt.Sprintf("agent-%s", agentID))

	// Remove the worktree
	cmd := exec.Command("git", "-C", wm.repo.Path, "worktree", "remove", worktreePath, "--force")
	if output, err := cmd.CombinedOutput(); err != nil {
		// If worktree doesn't exist, that's fine
		if !strings.Contains(string(output), "not a working tree") {
			return fmt.Errorf("failed to remove worktree: %w, output: %s", err, string(output))
		}
	}

	// Delete the branch
	cmd = exec.Command("git", "-C", wm.repo.Path, "branch", "-D", branchName)
	if output, err := cmd.CombinedOutput(); err != nil {
		// If branch doesn't exist, that's fine
		if !strings.Contains(string(output), "not found") {
			return fmt.Errorf("failed to delete branch: %w, output: %s", err, string(output))
		}
	}

	return nil
}

// ListWorktrees lists all worktrees
func (wm *WorktreeManager) ListWorktrees() ([]*Worktree, error) {
	cmd := exec.Command("git", "-C", wm.repo.Path, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	var worktrees []*Worktree
	lines := strings.Split(string(output), "\n")
	var currentPath, currentBranch string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if currentPath != "" && currentBranch != "" {
				// Extract agent ID from path
				if strings.Contains(currentPath, wm.config.WorktreeRootDir) {
					parts := strings.Split(currentPath, string(filepath.Separator))
					for _, part := range parts {
						if strings.HasPrefix(part, "agent-") {
							agentID := strings.TrimPrefix(part, "agent-")
							worktrees = append(worktrees, &Worktree{
								Path:       currentPath,
								BranchName: currentBranch,
								AgentID:    agentID,
							})
							break
						}
					}
				}
			}
			currentPath, currentBranch = "", ""
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			currentPath = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") {
			currentBranch = strings.TrimPrefix(line, "branch refs/heads/")
		}
	}

	return worktrees, nil
}

// GetWorktree gets a specific worktree by agent ID
func (wm *WorktreeManager) GetWorktree(agentID string) (*Worktree, error) {
	worktrees, err := wm.ListWorktrees()
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if wt.AgentID == agentID {
			return wt, nil
		}
	}

	return nil, ErrWorktreeNotFound
}
