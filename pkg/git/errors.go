package git

import "errors"

var (
	// ErrNotGitRepo indicates the directory is not a git repository
	ErrNotGitRepo = errors.New("not a git repository")

	// ErrWorktreeExists indicates the worktree already exists
	ErrWorktreeExists = errors.New("worktree already exists")

	// ErrWorktreeNotFound indicates the worktree was not found
	ErrWorktreeNotFound = errors.New("worktree not found")

	// ErrMergeConflict indicates a merge conflict was detected
	ErrMergeConflict = errors.New("merge conflict detected")

	// ErrDirtyWorktree indicates the worktree has uncommitted changes
	ErrDirtyWorktree = errors.New("worktree has uncommitted changes")
)
