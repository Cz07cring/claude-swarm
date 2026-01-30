package git

import (
	"os"
	"os/exec"
	"testing"
)

func TestNewWorktreeManager(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	t.Run("valid configuration", func(t *testing.T) {
		wm, err := NewWorktreeManager(WorktreeConfig{
			BaseRepoPath:    repoPath,
			WorktreeRootDir: ".worktrees",
			BaseBranch:      "main",
		})

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if wm == nil {
			t.Error("Expected worktree manager, got nil")
		}
	})

	t.Run("invalid repository", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "not-git-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		_, err = NewWorktreeManager(WorktreeConfig{
			BaseRepoPath: tmpDir,
		})

		if err != ErrNotGitRepo {
			t.Errorf("Expected ErrNotGitRepo, got: %v", err)
		}
	})

	t.Run("default values", func(t *testing.T) {
		wm, err := NewWorktreeManager(WorktreeConfig{
			BaseRepoPath: repoPath,
		})

		if err != nil {
			t.Fatalf("Failed to create worktree manager: %v", err)
		}

		if wm.config.WorktreeRootDir != ".worktrees" {
			t.Errorf("Expected default WorktreeRootDir '.worktrees', got: %s", wm.config.WorktreeRootDir)
		}

		if wm.config.BaseBranch != "main" {
			t.Errorf("Expected default BaseBranch 'main', got: %s", wm.config.BaseBranch)
		}
	})
}

func TestCreateWorktree(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Get the actual default branch
	repo, _ := NewRepository(repoPath)
	defaultBranch, _ := repo.GetCurrentBranch()

	wm, err := NewWorktreeManager(WorktreeConfig{
		BaseRepoPath:    repoPath,
		WorktreeRootDir: ".worktrees",
		BaseBranch:      defaultBranch,
	})
	if err != nil {
		t.Fatalf("Failed to create worktree manager: %v", err)
	}

	t.Run("create worktree", func(t *testing.T) {
		worktree, err := wm.CreateWorktree("0")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if worktree == nil {
			t.Fatal("Expected worktree, got nil")
		}

		if worktree.AgentID != "0" {
			t.Errorf("Expected AgentID '0', got: %s", worktree.AgentID)
		}

		if worktree.BranchName != "agent-0-branch" {
			t.Errorf("Expected BranchName 'agent-0-branch', got: %s", worktree.BranchName)
		}

		// Verify worktree directory exists
		if _, err := os.Stat(worktree.Path); os.IsNotExist(err) {
			t.Errorf("Worktree directory does not exist: %s", worktree.Path)
		}

		// Verify branch was created
		cmd := exec.Command("git", "-C", repoPath, "branch", "--list", worktree.BranchName)
		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Failed to list branches: %v", err)
		}
		if len(output) == 0 {
			t.Errorf("Branch %s was not created", worktree.BranchName)
		}
	})

	t.Run("duplicate worktree", func(t *testing.T) {
		_, err := wm.CreateWorktree("0")
		if err != ErrWorktreeExists {
			t.Errorf("Expected ErrWorktreeExists, got: %v", err)
		}
	})
}

func TestRemoveWorktree(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Get the actual default branch
	repo, _ := NewRepository(repoPath)
	defaultBranch, _ := repo.GetCurrentBranch()

	wm, err := NewWorktreeManager(WorktreeConfig{
		BaseRepoPath:    repoPath,
		WorktreeRootDir: ".worktrees",
		BaseBranch:      defaultBranch,
	})
	if err != nil {
		t.Fatalf("Failed to create worktree manager: %v", err)
	}

	// Create worktree
	worktree, err := wm.CreateWorktree("1")
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	t.Run("remove worktree", func(t *testing.T) {
		err := wm.RemoveWorktree("1")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify worktree directory was removed
		if _, err := os.Stat(worktree.Path); !os.IsNotExist(err) {
			t.Errorf("Worktree directory still exists: %s", worktree.Path)
		}

		// Verify branch was deleted
		cmd := exec.Command("git", "-C", repoPath, "branch", "--list", worktree.BranchName)
		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Failed to list branches: %v", err)
		}
		if len(output) != 0 {
			t.Errorf("Branch %s was not deleted", worktree.BranchName)
		}
	})

	t.Run("remove non-existent worktree", func(t *testing.T) {
		err := wm.RemoveWorktree("999")
		// Should not error on non-existent worktree
		if err != nil {
			t.Logf("Warning: Got error removing non-existent worktree: %v", err)
		}
	})
}

func TestListWorktrees(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Get the actual default branch
	repo, _ := NewRepository(repoPath)
	defaultBranch, _ := repo.GetCurrentBranch()

	wm, err := NewWorktreeManager(WorktreeConfig{
		BaseRepoPath:    repoPath,
		WorktreeRootDir: ".worktrees",
		BaseBranch:      defaultBranch,
	})
	if err != nil {
		t.Fatalf("Failed to create worktree manager: %v", err)
	}

	t.Run("list empty worktrees", func(t *testing.T) {
		worktrees, err := wm.ListWorktrees()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(worktrees) != 0 {
			t.Errorf("Expected 0 worktrees, got: %d", len(worktrees))
		}
	})

	t.Run("list multiple worktrees", func(t *testing.T) {
		// Create multiple worktrees
		_, err := wm.CreateWorktree("0")
		if err != nil {
			t.Fatalf("Failed to create worktree 0: %v", err)
		}

		_, err = wm.CreateWorktree("1")
		if err != nil {
			t.Fatalf("Failed to create worktree 1: %v", err)
		}

		worktrees, err := wm.ListWorktrees()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(worktrees) != 2 {
			t.Errorf("Expected 2 worktrees, got: %d", len(worktrees))
		}

		// Verify worktree details
		foundAgent0 := false
		foundAgent1 := false
		for _, wt := range worktrees {
			if wt.AgentID == "0" {
				foundAgent0 = true
			}
			if wt.AgentID == "1" {
				foundAgent1 = true
			}
		}

		if !foundAgent0 || !foundAgent1 {
			t.Error("Expected to find both agent-0 and agent-1 worktrees")
		}
	})
}

func TestGetWorktree(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Get the actual default branch
	repo, _ := NewRepository(repoPath)
	defaultBranch, _ := repo.GetCurrentBranch()

	wm, err := NewWorktreeManager(WorktreeConfig{
		BaseRepoPath:    repoPath,
		WorktreeRootDir: ".worktrees",
		BaseBranch:      defaultBranch,
	})
	if err != nil {
		t.Fatalf("Failed to create worktree manager: %v", err)
	}

	// Create worktree
	created, err := wm.CreateWorktree("2")
	if err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	t.Run("get existing worktree", func(t *testing.T) {
		worktree, err := wm.GetWorktree("2")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if worktree.AgentID != created.AgentID {
			t.Errorf("Expected AgentID %s, got: %s", created.AgentID, worktree.AgentID)
		}

		if worktree.BranchName != created.BranchName {
			t.Errorf("Expected BranchName %s, got: %s", created.BranchName, worktree.BranchName)
		}
	})

	t.Run("get non-existent worktree", func(t *testing.T) {
		_, err := wm.GetWorktree("999")
		if err != ErrWorktreeNotFound {
			t.Errorf("Expected ErrWorktreeNotFound, got: %v", err)
		}
	})
}
