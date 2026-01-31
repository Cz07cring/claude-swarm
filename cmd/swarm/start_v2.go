package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/yourusername/claude-swarm/pkg/controller"
)

var startV2Cmd = &cobra.Command{
	Use:   "start-v2",
	Short: "Start the swarm (V2 - direct Claude CLI execution)",
	Long:  "Starts the Claude agent swarm using V2 architecture with Git worktree isolation and free Claude CLI execution",
	Run:   runStartV2,
}

var (
	v2NumAgents int
	v2TaskFile  string
)

func init() {
	rootCmd.AddCommand(startV2Cmd)

	startV2Cmd.Flags().IntVar(&v2NumAgents, "agents", 3, "Number of agents to start")
	startV2Cmd.Flags().StringVar(&v2TaskFile, "tasks", "~/.claude-swarm/tasks.json", "Path to tasks file")
}

func runStartV2(cmd *cobra.Command, args []string) {
	log.SetFlags(log.Ltime)

	fmt.Println("🚀 启动 Claude Agent Swarm V2...")
	fmt.Println()

	// Get current directory as repo path
	repoPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Expand task file path
	if len(v2TaskFile) >= 2 && v2TaskFile[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		v2TaskFile = filepath.Join(home, v2TaskFile[2:])
	}

	// Create coordinator
	coord, err := controller.NewCoordinatorV2(repoPath, v2TaskFile, v2NumAgents)
	if err != nil {
		log.Fatalf("Failed to create coordinator: %v", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start coordinator
	if err := coord.Start(); err != nil {
		log.Fatalf("Failed to start coordinator: %v", err)
	}

	fmt.Println()
	fmt.Printf("✓ Swarm started with %d agents\n", v2NumAgents)
	fmt.Printf("✓ Task queue: %s\n", v2TaskFile)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop...")
	fmt.Println()

	// Wait for signal
	<-sigChan

	fmt.Println()
	fmt.Println("🛑 Stopping swarm...")

	// Stop coordinator
	if err := coord.Stop(); err != nil {
		log.Printf("Error stopping coordinator: %v", err)
	}

	// Cleanup
	if err := coord.Cleanup(); err != nil {
		log.Printf("Error during cleanup: %v", err)
	}

	fmt.Println("✓ Swarm stopped")
}
