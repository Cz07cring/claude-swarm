package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Configure git
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create initial commit
	testFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create test file: %v", err)
	}

	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	cmd = exec.Command("git", "-C", tmpDir, "commit", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestNewRepository(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	t.Run("valid repository", func(t *testing.T) {
		repo, err := NewRepository(repoPath)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if repo == nil {
			t.Error("Expected repository, got nil")
		}
		if repo.Path != repoPath {
			t.Errorf("Expected path %s, got %s", repoPath, repo.Path)
		}
	})

	t.Run("invalid repository", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "not-git-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		_, err = NewRepository(tmpDir)
		if err != ErrNotGitRepo {
			t.Errorf("Expected ErrNotGitRepo, got: %v", err)
		}
	})
}

func TestGetCurrentBranch(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := NewRepository(repoPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	branch, err := repo.GetCurrentBranch()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Default branch could be "master" or "main"
	if branch != "master" && branch != "main" {
		t.Errorf("Expected master or main, got: %s", branch)
	}
}

func TestIsClean(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := NewRepository(repoPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	t.Run("clean repository", func(t *testing.T) {
		clean, err := repo.IsClean()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !clean {
			t.Error("Expected clean repository")
		}
	})

	t.Run("dirty repository", func(t *testing.T) {
		// Create untracked file
		testFile := filepath.Join(repoPath, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		clean, err := repo.IsClean()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if clean {
			t.Error("Expected dirty repository")
		}
	})
}

func TestGetCurrentCommit(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	repo, err := NewRepository(repoPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	commit, err := repo.GetCurrentCommit()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Commit hash should be 40 characters (SHA-1)
	if len(commit) != 40 {
		t.Errorf("Expected 40 character commit hash, got %d: %s", len(commit), commit)
	}
}
