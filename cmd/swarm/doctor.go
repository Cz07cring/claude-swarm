package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system requirements and configuration",
	Long: `Diagnose your system to ensure Claude Swarm can run properly.

Checks:
  - Go installation
  - Git installation and version
  - Claude CLI availability
  - Project configuration
  - Environment variables`,
	Run: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) {
	fmt.Println("Claude Swarm Doctor")
	fmt.Println("==================")
	fmt.Println()

	allOk := true

	// Check Go
	fmt.Print("Checking Go... ")
	if goVersion, err := exec.Command("go", "version").Output(); err == nil {
		fmt.Printf("OK (%s)\n", strings.TrimSpace(string(goVersion)))
	} else {
		fmt.Println("NOT FOUND")
		allOk = false
	}

	// Check Git
	fmt.Print("Checking Git... ")
	if gitVersion, err := exec.Command("git", "--version").Output(); err == nil {
		version := strings.TrimSpace(string(gitVersion))
		fmt.Printf("OK (%s)\n", version)

		// Check git worktree support (2.5+)
		fmt.Print("  Worktree support... ")
		if _, err := exec.Command("git", "worktree", "list").Output(); err == nil {
			fmt.Println("OK")
		} else {
			fmt.Println("NOT AVAILABLE (need Git 2.5+)")
			allOk = false
		}
	} else {
		fmt.Println("NOT FOUND")
		allOk = false
	}

	// Check Claude CLI
	fmt.Print("Checking Claude CLI... ")
	if claudeVersion, err := exec.Command("claude", "--version").Output(); err == nil {
		fmt.Printf("OK (%s)\n", strings.TrimSpace(string(claudeVersion)))
	} else {
		fmt.Println("NOT FOUND")
		fmt.Println("  Install: https://claude.ai/code")
		allOk = false
	}

	// Check current directory
	fmt.Println()
	cwd, _ := os.Getwd()
	fmt.Printf("Current directory: %s\n", cwd)

	// Check if git repo
	fmt.Print("  Git repository... ")
	if isGitRepo(cwd) {
		fmt.Println("YES")

		// Check current branch
		if branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
			fmt.Printf("  Current branch: %s\n", strings.TrimSpace(string(branch)))
		}
	} else {
		fmt.Println("NO")
	}

	// Check project config
	fmt.Print("  Project config (.swarm/)... ")
	swarmDir := filepath.Join(cwd, ".swarm")
	if _, err := os.Stat(swarmDir); err == nil {
		fmt.Println("FOUND")

		// Check config.yaml
		configFile := filepath.Join(swarmDir, "config.yaml")
		if _, err := os.Stat(configFile); err == nil {
			fmt.Println("    config.yaml: OK")
		}

		// Check tasks.json
		tasksFile := filepath.Join(swarmDir, "tasks.json")
		if _, err := os.Stat(tasksFile); err == nil {
			fmt.Println("    tasks.json: OK")
		}
	} else {
		fmt.Println("NOT FOUND (run 'swarm init' to create)")
	}

	// Check environment variables
	fmt.Println()
	fmt.Println("Environment:")

	fmt.Print("  GEMINI_API_KEY... ")
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		fmt.Printf("SET (%s...)\n", key[:8])
	} else {
		fmt.Println("NOT SET (optional, for AI features)")
	}

	// Summary
	fmt.Println()
	fmt.Println("==================")
	if allOk {
		fmt.Println("All checks passed! Claude Swarm is ready to use.")
		fmt.Println()
		fmt.Println("Quick start:")
		fmt.Println("  swarm run \"Your task description\"")
	} else {
		fmt.Println("Some checks failed. Please fix the issues above.")
	}
}
