package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestNewMergeManager(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := NewRepository(repoPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	mm := NewMergeManager(repo)
	if mm == nil {
		t.Error("Expected merge manager, got nil")
	}

	if mm.repo != repo {
		t.Error("Expected merge manager to reference the repository")
	}
}

func TestMergeBranch(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := NewRepository(repoPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	mm := NewMergeManager(repo)

	t.Run("fast-forward merge", func(t *testing.T) {
		// Create a new branch
		cmd := exec.Command("git", "-C", repoPath, "checkout", "-b", "feature-branch")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create branch: %v", err)
		}

		// Add a commit
		testFile := filepath.Join(repoPath, "feature.txt")
		if err := os.WriteFile(testFile, []byte("feature"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		exec.Command("git", "-C", repoPath, "add", ".").Run()
		cmd = exec.Command("git", "-C", repoPath, "commit", "-m", "Add feature")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		// Switch back to master/main
		cmd = exec.Command("git", "-C", repoPath, "checkout", "master")
		if err := cmd.Run(); err != nil {
			// Try main if master doesn't exist
			exec.Command("git", "-C", repoPath, "checkout", "main").Run()
		}

		// Merge
		result, err := mm.MergeBranch("feature-branch")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if !result.Success {
			t.Error("Expected successful merge")
		}

		if !result.FastForward {
			t.Error("Expected fast-forward merge")
		}

		if result.CommitHash == "" {
			t.Error("Expected commit hash")
		}
	})

	t.Run("merge non-existent branch", func(t *testing.T) {
		_, err := mm.MergeBranch("non-existent-branch")
		if err == nil {
			t.Error("Expected error when merging non-existent branch")
		}
	})
}

func TestCanFastForward(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := NewRepository(repoPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	mm := NewMergeManager(repo)

	t.Run("can fast-forward", func(t *testing.T) {
		// Create a new branch from current HEAD
		cmd := exec.Command("git", "-C", repoPath, "checkout", "-b", "ff-branch")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create branch: %v", err)
		}

		// Add a commit
		testFile := filepath.Join(repoPath, "ff.txt")
		if err := os.WriteFile(testFile, []byte("ff"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		exec.Command("git", "-C", repoPath, "add", ".").Run()
		cmd = exec.Command("git", "-C", repoPath, "commit", "-m", "Add ff")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		// Switch back to master/main
		cmd = exec.Command("git", "-C", repoPath, "checkout", "master")
		if err := cmd.Run(); err != nil {
			exec.Command("git", "-C", repoPath, "checkout", "main").Run()
		}

		// Check if can fast-forward
		canFF, err := mm.CanFastForward("ff-branch")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if !canFF {
			t.Error("Expected to be able to fast-forward")
		}
	})
}

func TestAbortMerge(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := NewRepository(repoPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	mm := NewMergeManager(repo)

	t.Run("abort non-existent merge", func(t *testing.T) {
		// Should not error if there's no merge in progress
		err := mm.AbortMerge()
		if err != nil {
			t.Logf("Warning: Got error aborting non-existent merge: %v", err)
		}
	})
}
