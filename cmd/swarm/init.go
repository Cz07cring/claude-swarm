package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Claude Swarm in current project",
	Long: `Initialize Claude Swarm configuration in the current project directory.

This creates:
  - .swarm/config.yaml   - Project configuration
  - .swarm/tasks.json    - Task queue file
  - .gitignore update    - Ignore worktrees and logs

Example:
  cd your-project
  swarm init
  swarm run "Add user authentication"`,
	Run: runInit,
}

var (
	initForce bool
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing configuration")
}

func runInit(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	swarmDir := filepath.Join(cwd, ".swarm")
	configFile := filepath.Join(swarmDir, "config.yaml")
	tasksFile := filepath.Join(swarmDir, "tasks.json")

	// Check if already initialized
	if _, err := os.Stat(swarmDir); err == nil && !initForce {
		fmt.Println("Project already initialized. Use --force to reinitialize.")
		return
	}

	fmt.Println("Initializing Claude Swarm...")

	// Create .swarm directory
	if err := os.MkdirAll(swarmDir, 0755); err != nil {
		log.Fatalf("Failed to create .swarm directory: %v", err)
	}

	// Create config.yaml
	configContent := `# Claude Swarm Configuration
# Project: ` + filepath.Base(cwd) + `

# Agent settings
agents:
  count: 3                    # Number of parallel agents (1-10)
  timeout: 10m                # Task timeout

# Task queue
tasks:
  file: .swarm/tasks.json     # Task queue file location
  max_retries: 3              # Max retry attempts per task

# Git settings
git:
  worktree_dir: .worktrees    # Worktree directory
  base_branch: main           # Base branch for merging
  auto_merge: true            # Auto-merge completed work

# AI Brain (optional - requires GEMINI_API_KEY)
brain:
  enabled: false              # Enable AI orchestration
  # api_key: ""               # Or use GEMINI_API_KEY env var

# Logging
logging:
  level: info                 # debug, info, warn, error
  file: .swarm/swarm.log      # Log file location
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		log.Fatalf("Failed to create config.yaml: %v", err)
	}
	fmt.Println("  Created .swarm/config.yaml")

	// Create empty tasks.json
	tasksContent := `{
  "tasks": []
}
`
	if err := os.WriteFile(tasksFile, []byte(tasksContent), 0644); err != nil {
		log.Fatalf("Failed to create tasks.json: %v", err)
	}
	fmt.Println("  Created .swarm/tasks.json")

	// Update .gitignore
	gitignorePath := filepath.Join(cwd, ".gitignore")
	gitignoreEntries := `
# Claude Swarm
.worktrees/
.swarm/tasks.json
.swarm/swarm.log
`

	// Check if .gitignore exists and append, otherwise create
	if _, err := os.Stat(gitignorePath); err == nil {
		// Append to existing
		f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			f.WriteString(gitignoreEntries)
			fmt.Println("  Updated .gitignore")
		}
	} else {
		// Create new
		if err := os.WriteFile(gitignorePath, []byte(gitignoreEntries), 0644); err == nil {
			fmt.Println("  Created .gitignore")
		}
	}

	fmt.Println()
	fmt.Println("Claude Swarm initialized successfully!")
	fmt.Println()
	fmt.Println("Quick start:")
	fmt.Println("  swarm run \"Your task description\"     # Run single task")
	fmt.Println("  swarm add-task \"Task 1\"               # Add to queue")
	fmt.Println("  swarm add-task \"Task 2\"")
	fmt.Println("  swarm start --agents 3                # Run queue")
	fmt.Println()
	fmt.Println("With AI orchestration:")
	fmt.Println("  export GEMINI_API_KEY=your-key")
	fmt.Println("  swarm orchestrate \"Build a REST API\"  # AI generates tasks")
	fmt.Println("  swarm start --agents 5 --with-brain   # Run with AI monitoring")
}
