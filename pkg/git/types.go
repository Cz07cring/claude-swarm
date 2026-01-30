package git

import "time"

// Repository represents a Git repository
type Repository struct {
	Path string // 主仓库路径
}

// Worktree represents a Git worktree
type Worktree struct {
	Path       string    // Worktree绝对路径
	BranchName string    // 对应的分支名
	AgentID    string    // 所属Agent ID
	CreatedAt  time.Time
}

// WorktreeConfig contains configuration for worktree management
type WorktreeConfig struct {
	BaseRepoPath    string // 主仓库路径
	WorktreeRootDir string // 默认：.worktrees
	BaseBranch      string // 默认：main
}

// MergeResult contains the result of a merge operation
type MergeResult struct {
	Success     bool
	FastForward bool
	Conflicts   []string
	CommitHash  string
}
