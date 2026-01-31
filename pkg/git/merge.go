package git

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// MergeManager handles merge operations
type MergeManager struct {
	repo       *Repository
	mu         sync.Mutex  // Protect concurrent merge operations
	inProgress bool        // Track if a merge is in progress
}

// NewMergeManager creates a new MergeManager
func NewMergeManager(repo *Repository) *MergeManager {
	return &MergeManager{repo: repo}
}

// MergeBranch merges a branch into the current branch
func (mm *MergeManager) MergeBranch(branchName string) (*MergeResult, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Check if another merge is in progress
	if mm.inProgress {
		return nil, fmt.Errorf("merge already in progress")
	}

	mm.inProgress = true
	defer func() { mm.inProgress = false }()

	result := &MergeResult{}

	// Try fast-forward merge first
	cmd := exec.Command("git", "-C", mm.repo.Path, "merge", "--ff-only", branchName)
	output, err := cmd.CombinedOutput()

	if err == nil {
		result.Success = true
		result.FastForward = true
		result.CommitHash, _ = mm.getCurrentCommit()
		return result, nil
	}

	// Fast-forward failed, try three-way merge
	cmd = exec.Command("git", "-C", mm.repo.Path, "merge", "--no-ff",
		"-m", fmt.Sprintf("Merge branch '%s'", branchName), branchName)
	output, err = cmd.CombinedOutput()

	if err != nil {
		if strings.Contains(string(output), "CONFLICT") {
			conflicts, _ := mm.getConflicts()
			result.Success = false
			result.Conflicts = conflicts
			return result, ErrMergeConflict
		}
		return nil, fmt.Errorf("merge failed: %w, output: %s", err, string(output))
	}

	result.Success = true
	result.FastForward = false
	result.CommitHash, _ = mm.getCurrentCommit()
	return result, nil
}

// AbortMerge aborts an in-progress merge
func (mm *MergeManager) AbortMerge() error {
	cmd := exec.Command("git", "-C", mm.repo.Path, "merge", "--abort")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to abort merge: %w, output: %s", err, string(output))
	}
	return nil
}

// CanFastForward checks if a branch can be fast-forward merged
func (mm *MergeManager) CanFastForward(branchName string) (bool, error) {
	// Check if current branch is ancestor of target branch
	cmd := exec.Command("git", "-C", mm.repo.Path, "merge-base", "--is-ancestor", "HEAD", branchName)
	err := cmd.Run()
	return err == nil, nil
}

// getCurrentCommit gets the current commit hash
func (mm *MergeManager) getCurrentCommit() (string, error) {
	return mm.repo.GetCurrentCommit()
}

// getConflicts returns the list of conflicted files
func (mm *MergeManager) getConflicts() ([]string, error) {
	cmd := exec.Command("git", "-C", mm.repo.Path, "diff", "--name-only", "--diff-filter=U")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get conflicts: %w", err)
	}

	conflicts := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, conflict := range conflicts {
		if conflict != "" {
			result = append(result, conflict)
		}
	}

	return result, nil
}
