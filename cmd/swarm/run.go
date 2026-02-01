package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/executor"
)

var runCmd = &cobra.Command{
	Use:   "run [task description]",
	Short: "Quick run a single task with Claude",
	Long: `Execute a single task directly without managing a task queue.

This is the simplest way to use Claude Swarm - just describe what you want
and it will be done. Perfect for quick one-off tasks.

Examples:
  swarm run "Create a README.md file"
  swarm run "Add error handling to main.go"
  swarm run "Write unit tests for user.go"

  # With timeout
  swarm run "Refactor the authentication module" --timeout 15m

  # Pipe input
  echo "Fix the bug in line 42" | swarm run`,
	Args: cobra.MaximumNArgs(1),
	Run:  runQuickTask,
}

var (
	runTimeout time.Duration
	runDryRun  bool
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().DurationVar(&runTimeout, "timeout", 10*time.Minute, "Task timeout")
	runCmd.Flags().BoolVar(&runDryRun, "dry-run", false, "Show what would be executed without running")
}

func runQuickTask(cmd *cobra.Command, args []string) {
	var taskDescription string

	// Get task from args or stdin
	if len(args) > 0 {
		taskDescription = strings.Join(args, " ")
	} else {
		// Try to read from stdin (for piping)
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped
			buf := make([]byte, 4096)
			n, err := os.Stdin.Read(buf)
			if err == nil && n > 0 {
				taskDescription = strings.TrimSpace(string(buf[:n]))
			}
		}
	}

	if taskDescription == "" {
		fmt.Println("Usage: swarm run \"task description\"")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  swarm run \"Create a hello.go file\"")
		fmt.Println("  swarm run \"Add tests for utils.go\"")
		fmt.Println("  echo \"Fix the bug\" | swarm run")
		return
	}

	// Get working directory
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Check for project config
	configPath := filepath.Join(workDir, ".swarm", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("Using project config: .swarm/config.yaml")
	}

	// Check if we're in a git repo
	if !isGitRepo(workDir) {
		fmt.Println("Note: Not in a git repository. Running without worktree isolation.")
		fmt.Println()
	}

	// Dry run mode
	if runDryRun {
		fmt.Println("Dry run mode - would execute:")
		fmt.Printf("  Task: %s\n", taskDescription)
		fmt.Printf("  Directory: %s\n", workDir)
		fmt.Printf("  Timeout: %s\n", runTimeout)
		return
	}

	fmt.Println("Running task with Claude...")
	fmt.Printf("  Task: %s\n", taskDescription)
	fmt.Printf("  Directory: %s\n", workDir)
	fmt.Println()

	// Create task
	task := &models.Task{
		ID:          fmt.Sprintf("quick-%d", time.Now().UnixNano()),
		Description: taskDescription,
		Status:      models.TaskStatusInProgress,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Setup signal handling
	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nCancelling task...")
		cancel()
	}()

	// Execute
	startTime := time.Now()
	claudeExecutor := executor.NewClaudeExecutor(workDir)

	err = claudeExecutor.ExecuteTask(ctx, task)

	elapsed := time.Since(startTime)

	fmt.Println()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("Task timed out after %s\n", runTimeout)
		} else if ctx.Err() == context.Canceled {
			fmt.Println("Task cancelled by user")
		} else {
			fmt.Printf("Task failed: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("Task completed in %.1fs\n", elapsed.Seconds())
}

// isGitRepo checks if the directory is a git repository
func isGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}
