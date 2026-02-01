package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/state"
	"github.com/yourusername/claude-swarm/pkg/tui"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "启动 TUI 监控面板",
	Long:  "启动交互式 TUI 监控面板，实时查看 agents 和任务状态",
	Run:   runMonitor,
}

var (
	monitorTaskFile string
)

func init() {
	rootCmd.AddCommand(monitorCmd)

	monitorCmd.Flags().StringVar(&monitorTaskFile, "tasks", "~/.claude-swarm/tasks.json", "任务队列文件路径")
}

func runMonitor(cmd *cobra.Command, args []string) {
	// Expand task file path
	if len(monitorTaskFile) >= 2 && monitorTaskFile[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		monitorTaskFile = filepath.Join(home, monitorTaskFile[2:])
	}

	// Initialize task queue
	taskQueue, err := state.NewTaskQueue(monitorTaskFile)
	if err != nil {
		log.Fatalf("Failed to open task queue: %v", err)
	}
	defer taskQueue.Close()

	// Create agent status loader function
	getAgentsFn := func() []*models.AgentStatus {
		return loadAgentStatuses(taskQueue)
	}

	// Start TUI using the Run helper
	if err := tui.Run(taskQueue, getAgentsFn); err != nil {
		fmt.Printf("Error running monitor: %v\n", err)
		os.Exit(1)
	}
}

// loadAgentStatuses loads agent statuses from task queue
func loadAgentStatuses(taskQueue *state.TaskQueue) []*models.AgentStatus {
	tasks := taskQueue.ListTasks()

	// Extract unique agents from tasks
	agentMap := make(map[string]*models.AgentStatus)

	for _, task := range tasks {
		if task.AssigneeID != "" {
			if _, exists := agentMap[task.AssigneeID]; !exists {
				state := models.AgentStateIdle
				var currentTask *models.Task

				if task.Status == models.TaskStatusInProgress {
					state = models.AgentStateWorking
					currentTask = task
				}

				agentMap[task.AssigneeID] = &models.AgentStatus{
					AgentID:     task.AssigneeID,
					State:       state,
					CurrentTask: currentTask,
					LastUpdate:  task.UpdatedAt,
				}
			}
		}
	}

	// Convert map to slice
	statuses := make([]*models.AgentStatus, 0, len(agentMap))
	for _, status := range agentMap {
		statuses = append(statuses, status)
	}

	return statuses
}

