package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/state"
	"github.com/yourusername/claude-swarm/pkg/tui"
)

var (
	monitorQueuePath string
	monitorStatePath string
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "启动 TUI 监控面板",
	Long:  `启动一个交互式终端界面来监控 Agent 集群和任务队列的实时状态`,
	Run: func(cmd *cobra.Command, args []string) {
		runMonitor()
	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)

	monitorCmd.Flags().StringVarP(&monitorQueuePath, "queue", "q", "~/.claude-swarm/tasks.json", "任务队列文件路径")
	monitorCmd.Flags().StringVarP(&monitorStatePath, "state", "s", "~/.claude-swarm/agents.json", "Agent 状态文件路径")
}

func runMonitor() {
	// Create task queue
	taskQueue, err := state.NewTaskQueue(monitorQueuePath)
	if err != nil {
		log.Fatalf("❌ 打开任务队列失败: %v\n提示: 请确保 swarm 已启动或任务队列文件存在", err)
	}
	defer taskQueue.Close()

	// Create agent state manager
	agentStateMgr, err := state.NewAgentStateManager(monitorStatePath)
	if err != nil {
		log.Fatalf("❌ 打开 Agent 状态文件失败: %v\n提示: 请确保 swarm 已启动或状态文件存在", err)
	}
	defer agentStateMgr.Close()

	// Create function to get agent status
	getAgentsFn := func() []*models.AgentStatus {
		agents, err := agentStateMgr.GetAgents()
		if err != nil {
			// Return empty list on error to avoid crashing the TUI
			return []*models.AgentStatus{}
		}
		return agents
	}

	// Run TUI
	fmt.Println("🚀 启动 Claude Agent Swarm 监控面板...")
	fmt.Println("   提示: 使用 Tab 切换面板，↑↓/jk 导航，q 退出")
	fmt.Println()

	if err := tui.Run(taskQueue, getAgentsFn); err != nil {
		log.Fatalf("❌ TUI 运行失败: %v", err)
	}
}
