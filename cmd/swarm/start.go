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
	"github.com/yourusername/claude-swarm/pkg/controller"
	"github.com/yourusername/claude-swarm/pkg/orchestrator"
	"github.com/yourusername/claude-swarm/pkg/state"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the swarm with direct Claude CLI execution",
	Long:  "Starts the Claude agent swarm with Git worktree isolation and free Claude CLI execution",
	Run:   runStart,
}

var (
	numAgents   int
	taskFile    string
	withBrain   bool
	brainAPIKey string
)

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().IntVar(&numAgents, "agents", 3, "Number of agents to start")
	startCmd.Flags().StringVar(&taskFile, "tasks", "~/.claude-swarm/tasks.json", "Path to tasks file")
	startCmd.Flags().BoolVar(&withBrain, "with-brain", false, "启用AI主脑监控和智能决策")
	startCmd.Flags().StringVar(&brainAPIKey, "brain-api-key", "", "Gemini API Key for AI brain (or use GEMINI_API_KEY env var)")
}

func runStart(cmd *cobra.Command, args []string) {
	log.SetFlags(log.Ltime)

	fmt.Println("🚀 启动 Claude Agent Swarm...")
	fmt.Println()

	// Get current directory as repo path
	repoPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Expand task file path
	if len(taskFile) >= 2 && taskFile[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		taskFile = filepath.Join(home, taskFile[2:])
	}

	// Create coordinator
	coord, err := controller.NewCoordinator(repoPath, taskFile, numAgents)
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
	fmt.Printf("✓ Swarm started with %d agents\n", numAgents)
	fmt.Printf("✓ Task queue: %s\n", taskFile)
	fmt.Println()

	// 启动AI主脑监控（可选）
	var brainCancelFunc context.CancelFunc
	if withBrain {
		ctx, cancel := context.WithCancel(context.Background())
		brainCancelFunc = cancel

		if err := startBrainMonitor(ctx, taskFile, coord); err != nil {
			log.Printf("⚠️  AI主脑启动失败: %v", err)
			log.Println("继续运行（无主脑监控）...")
		} else {
			fmt.Println("✓ AI主脑监控已启动")
			fmt.Println()
		}
	}

	fmt.Println("Press Ctrl+C to stop...")
	fmt.Println()

	// Wait for signal
	<-sigChan

	// 停止主脑监控
	if brainCancelFunc != nil {
		brainCancelFunc()
	}

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

// startBrainMonitor 启动AI主脑监控循环
func startBrainMonitor(ctx context.Context, taskFilePath string, coord *controller.Coordinator) error {
	// 获取API Key
	apiKey := brainAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		return fmt.Errorf("需要Gemini API Key: 使用 --brain-api-key 或设置 GEMINI_API_KEY 环境变量")
	}

	// 初始化任务队列
	taskQueue, err := state.NewTaskQueue(taskFilePath)
	if err != nil {
		return fmt.Errorf("初始化任务队列失败: %w", err)
	}

	// 创建AI主脑
	brain, err := orchestrator.NewOrchestratorBrain(apiKey, taskQueue)
	if err != nil {
		return fmt.Errorf("创建AI主脑失败: %w", err)
	}

	// 启动监控协程
	go func() {
		defer brain.Close()
		defer taskQueue.Close()

		ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
		defer ticker.Stop()

		log.Println("🧠 AI主脑监控循环启动...")

		for {
			select {
			case <-ticker.C:
				// 收集Agent状态（这里简化处理，实际应从coordinator获取）
				agents := collectAgentStatus(taskQueue)

				// AI监控进度
				progress, err := brain.MonitorProgress(ctx, agents)
				if err != nil {
					log.Printf("⚠️  主脑监控失败: %v", err)
					continue
				}

				// 打印进度摘要
				log.Printf("📊 进度: %d/%d 完成 (%.1f%%), %d进行中, %d失败",
					progress.CompletedTasks, progress.TotalTasks, progress.OverallProgress,
					progress.InProgressTasks, progress.FailedTasks)

				// AI决策下一步行动
				action, err := brain.DecideNextAction(ctx, progress)
				if err != nil {
					log.Printf("⚠️  主脑决策失败: %v", err)
					continue
				}

				// 执行行动
				if action.Type != orchestrator.ActionWait {
					executeAction(action, taskQueue)
				}

				// 检查是否需要智能合并（每完成一批任务后）
				if progress.CompletedTasks > 0 && progress.InProgressTasks == 0 {
					checkAndMerge(ctx, brain, coord)
				}

			case <-ctx.Done():
				log.Println("🛑 AI主脑监控循环停止")
				return
			}
		}
	}()

	return nil
}

// collectAgentStatus 收集Agent状态（简化版）
func collectAgentStatus(taskQueue *state.TaskQueue) []*models.AgentStatus {
	tasks := taskQueue.ListTasks()
	agents := make(map[string]*models.AgentStatus)

	// 根据任务状态推断Agent状态
	for _, task := range tasks {
		if task.Status == models.TaskStatusInProgress && task.AssigneeID != "" {
			if agent, exists := agents[task.AssigneeID]; !exists {
				agents[task.AssigneeID] = &models.AgentStatus{
					AgentID:     task.AssigneeID,
					State:       models.AgentStateWorking,
					CurrentTask: task,
					LastUpdate:  task.UpdatedAt,
				}
			} else {
				// 检查是否卡住（3分钟无更新）
				if time.Since(agent.LastUpdate) > 3*time.Minute {
					agent.State = models.AgentStateStuck
				}
			}
		}
	}

	// 转换为切片
	result := make([]*models.AgentStatus, 0, len(agents))
	for _, agent := range agents {
		result = append(result, agent)
	}

	return result
}

// executeAction 执行主脑决策的行动
func executeAction(action *orchestrator.Action, taskQueue *state.TaskQueue) {
	switch action.Type {
	case orchestrator.ActionHelpAgent:
		log.Printf("🆘 主脑介入帮助Agent: %s", action.Reason)
		if action.Command != "" {
			log.Printf("   提示: %s", action.Command)
		}
		// TODO: 实际发送提示给Agent（需要IPC机制）

	case orchestrator.ActionReassignTask:
		log.Printf("🔄 主脑重新分配任务: %s", action.Reason)
		if action.TaskID != "" {
			// 重置任务为pending状态
			if err := taskQueue.ResetOrphanedTask(action.TaskID); err != nil {
				log.Printf("⚠️  重置任务失败: %v", err)
			} else {
				log.Printf("   ✓ 任务 %s 已重置为待执行", action.TaskID)
			}
		}

	case orchestrator.ActionRestartAgent:
		log.Printf("♻️  主脑建议重启Agent: %s", action.Reason)
		// TODO: 实现Agent重启逻辑

	case orchestrator.ActionAssignTask:
		log.Printf("📌 主脑建议分配任务: %s", action.Reason)
		// 任务分配由coordinator自动处理

	case orchestrator.ActionMergeBranch:
		log.Printf("🔀 主脑决策合并分支: %s", action.Reason)
		// 合并逻辑在 checkAndMerge 中处理

	case orchestrator.ActionWait:
		// 静默等待
		return

	default:
		log.Printf("⚠️  未知的主脑行动类型: %s", action.Type)
	}
}

// checkAndMerge 使用AI智能决策合并
func checkAndMerge(ctx context.Context, brain *orchestrator.OrchestratorBrain, coord *controller.Coordinator) {
	// 获取所有分支的合并状态
	coordStatuses := coord.GetMergeStatuses()
	if len(coordStatuses) == 0 {
		return
	}

	// 转换为 orchestrator.MergeStatus
	mergeStatuses := make([]*orchestrator.MergeStatus, len(coordStatuses))
	for i, s := range coordStatuses {
		mergeStatuses[i] = &orchestrator.MergeStatus{
			Branch:       s.Branch,
			AgentID:      s.AgentID,
			HasChanges:   s.HasChanges,
			CommitCount:  s.CommitCount,
			Files:        s.Files,
			ReadyToMerge: s.ReadyToMerge,
		}
	}

	// 检查是否有可合并的分支
	hasReadyBranches := false
	for _, s := range mergeStatuses {
		if s.ReadyToMerge {
			hasReadyBranches = true
			break
		}
	}

	if !hasReadyBranches {
		return
	}

	log.Println("🧠 检测到可合并的分支，请求AI决策...")

	// 让AI决定合并策略
	decision, err := brain.DecideMergeStrategy(ctx, mergeStatuses)
	if err != nil {
		log.Printf("⚠️  AI合并决策失败: %v", err)
		return
	}

	if !decision.ShouldMerge {
		log.Printf("ℹ️  AI决策暂不合并: %s", decision.Reason)
		return
	}

	log.Printf("🔀 AI决策合并顺序: %v", decision.MergeOrder)
	if len(decision.PotentialIssues) > 0 {
		log.Printf("⚠️  潜在问题: %v", decision.PotentialIssues)
	}

	// 按顺序执行合并
	for _, branch := range decision.MergeOrder {
		log.Printf("🔀 合并分支: %s", branch)

		err := coord.MergeBranch(branch)
		if err != nil {
			// 检查是否是冲突
			if strings.Contains(err.Error(), "conflict") {
				log.Printf("⚠️  合并冲突: %s", branch)

				// 获取冲突详情
				conflictFiles, conflictContent, _ := coord.GetConflictDetails(branch)
				if len(conflictFiles) > 0 {
					// 让AI分析冲突
					resolution, err := brain.ResolveConflict(ctx, branch, conflictFiles, conflictContent)
					if err != nil {
						log.Printf("⚠️  AI冲突分析失败: %v", err)
					} else {
						log.Printf("🧠 AI冲突分析: %s", resolution.Resolution)
						if resolution.NeedsHumanReview {
							log.Printf("⚠️  需要人工审核冲突")
						}
					}
				}
			} else {
				log.Printf("❌ 合并失败: %v", err)
			}
			continue
		}

		log.Printf("✅ 成功合并: %s", branch)
	}
}
