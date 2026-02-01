package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version info (set by ldflags)
var (
	Version   = "2.0.0"
	BuildTime = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "swarm",
	Short: "Claude Swarm - AI-Powered Multi-Agent Development System",
	Long: `Claude Swarm is a multi-agent system powered by Claude CLI.
It uses Git worktree isolation and direct CLI execution for parallel task processing.

Quick Start:
  swarm run "Your task"           # Run single task (simplest)
  swarm init                      # Initialize project
  swarm add-task "Task 1"         # Add tasks to queue
  swarm start --agents 3          # Run with multiple agents

With AI:
  swarm orchestrate "Build API"   # AI generates tasks
  swarm start --with-brain        # AI monitors execution`,
	Version: Version,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Claude Swarm v%s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
