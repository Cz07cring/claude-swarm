package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/claude-swarm/pkg/config"
	"github.com/yourusername/claude-swarm/pkg/orchestrator"
	"github.com/yourusername/claude-swarm/pkg/state"
)

var orchestrateCmd = &cobra.Command{
	Use:   "orchestrate [需求描述]",
	Short: "🧠 AI主脑分析需求并自动拆分任务",
	Long: `AI主脑（Gemini）分析用户需求，智能拆分成多个可并行开发的任务。

示例：
  swarm orchestrate "实现一个用户管理系统，包括注册、登录、权限管理"
  swarm orchestrate "添加文件上传功能，支持图片预览和压缩"`,
	Args: cobra.MinimumNArgs(1),
	Run:  runOrchestrate,
}

var (
	geminiAPIKey   string
	configFilePath string
	autoStart      bool
	autoApprove    bool
	maxAgents      int
)

func init() {
	rootCmd.AddCommand(orchestrateCmd)

	orchestrateCmd.Flags().StringVarP(&geminiAPIKey, "api-key", "k", "", "Gemini API Key（或使用配置文件/环境变量）")
	orchestrateCmd.Flags().StringVarP(&configFilePath, "config", "c", "", "配置文件路径（默认: ./config.yaml 或 ~/.claude-swarm/config.yaml）")
	orchestrateCmd.Flags().BoolVar(&autoStart, "auto-start", false, "分析并审批通过后自动启动Agent集群")
	orchestrateCmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "跳过人工审批，自动创建任务")
	orchestrateCmd.Flags().IntVarP(&maxAgents, "agents", "n", 5, "Agent数量（1-10）")
}

func runOrchestrate(cmd *cobra.Command, args []string) {
	requirement := args[0]

	// 加载配置（优先级：命令行参数 > 配置文件 > 环境变量）
	cfg, err := config.Load(configFilePath)
	if err != nil {
		// 如果配置文件加载失败，使用环境变量作为后备
		log.Printf("⚠️  配置文件加载失败: %v", err)
		log.Printf("📝 尝试使用环境变量 GEMINI_API_KEY")
	}

	// 获取API Key（命令行参数最高优先级）
	apiKey := geminiAPIKey
	if apiKey == "" && cfg != nil {
		apiKey = cfg.Gemini.APIKey
	}
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		log.Fatal("❌ 请提供Gemini API Key:\n" +
			"   1. 使用 --api-key 参数\n" +
			"   2. 在 config.yaml 中配置\n" +
			"   3. 设置环境变量 GEMINI_API_KEY\n" +
			"   示例: cp config.yaml.example config.yaml && 编辑填入API Key")
	}

	fmt.Println("🧠 AI主脑启动中...")
	fmt.Printf("📝 需求: %s\n\n", requirement)

	// 初始化任务队列
	taskQueue, err := state.NewTaskQueue(taskQueuePath)
	if err != nil {
		log.Fatalf("❌ 初始化任务队列失败: %v", err)
	}

	// 创建AI主脑
	brain, err := orchestrator.NewOrchestratorBrain(apiKey, taskQueue)
	if err != nil {
		log.Fatalf("❌ 创建AI主脑失败: %v", err)
	}
	defer brain.Close()

	ctx := context.Background()

	// AI分析需求
	fmt.Println("🔍 AI分析需求中...")
	result, err := brain.AnalyzeRequirement(ctx, requirement)
	if err != nil {
		log.Fatalf("❌ AI分析失败: %v", err)
	}

	// 显示分析结果
	printAnalysisResult(result)

	// 人工审批环节（除非使用--auto-approve）
	approved := autoApprove
	if !autoApprove {
		approved = requestApproval(result)
	}

	if !approved {
		fmt.Println("\n❌ 已取消。未创建任务。")
		fmt.Println("💡 提示：您可以修改需求描述后重新运行 orchestrate")
		return
	}

	// 创建任务
	fmt.Println("\n📋 创建任务队列...")
	if err := brain.CreateTasksFromAnalysis(ctx, result); err != nil {
		log.Fatalf("❌ 创建任务失败: %v", err)
	}

	fmt.Printf("\n✅ 任务队列创建完成！共 %d 个任务\n", len(result.Tasks))

	// 提示下一步
	if autoStart {
		fmt.Println("\n🚀 自动启动Agent集群...")
		// TODO: 自动调用 start 命令
		fmt.Printf("   swarm start --agents %d\n", maxAgents)
	} else {
		fmt.Println("\n💡 下一步操作：")
		fmt.Printf("   swarm start --agents %d   # 启动%d个Agent开始工作\n", maxAgents, maxAgents)
		fmt.Println("   swarm status               # 查看任务状态")
		fmt.Println("   tmux attach -t claude-swarm  # 查看Agent实时输出")
	}
}

// requestApproval 请求用户审批AI分析结果
func requestApproval(result *orchestrator.AnalysisResult) bool {
	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Println("🔍 审批环节")
	fmt.Println(strings.Repeat("─", 60))

	fmt.Println("\n请仔细检查上述分析结果：")
	fmt.Println("  • 模块拆分是否合理？")
	fmt.Println("  • 任务描述是否清晰？")
	fmt.Println("  • 依赖关系是否正确？")
	fmt.Println("  • 预估时间是否合理？")

	fmt.Printf("\n📊 统计: %d个模块, %d个任务, 预计%s\n",
		len(result.Modules), len(result.Tasks), result.EstimatedTime)

	fmt.Println("\n选项:")
	fmt.Println("  1. ✅ 批准并创建任务队列")
	fmt.Println("  2. ❌ 拒绝（取消创建）")
	fmt.Println("  3. 📝 查看详细信息")

	for {
		fmt.Print("\n请选择 [1/2/3]: ")
		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1", "y", "Y", "yes", "Yes", "YES":
			fmt.Println("\n✅ 已批准！开始创建任务队列...")
			return true

		case "2", "n", "N", "no", "No", "NO":
			fmt.Println("\n❌ 已拒绝。")
			return false

		case "3", "d", "detail":
			printDetailedAnalysis(result)
			fmt.Println("\n返回审批选项...")
			continue

		default:
			fmt.Println("⚠️  无效选择，请输入 1、2 或 3")
			continue
		}
	}
}

// printDetailedAnalysis 打印详细的分析信息
func printDetailedAnalysis(result *orchestrator.AnalysisResult) {
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("📖 详细分析信息")
	fmt.Println(strings.Repeat("═", 60))

	// 显示每个任务的完整信息
	for i, task := range result.Tasks {
		fmt.Printf("\n【任务 %d】\n", i+1)
		fmt.Printf("  ID: %s\n", task.ID)
		fmt.Printf("  模块: %s\n", task.Module)
		fmt.Printf("  描述: %s\n", task.Description)
		fmt.Printf("  优先级: %d\n", task.Priority)
		fmt.Printf("  预计时间: %s\n", task.Estimated)

		if len(task.Dependencies) > 0 {
			fmt.Printf("  依赖: %v\n", task.Dependencies)
		} else {
			fmt.Printf("  依赖: 无（可立即执行）\n")
		}

		if len(task.Files) > 0 {
			fmt.Printf("  涉及文件: %v\n", task.Files)
		}
	}

	// 显示依赖图
	if len(result.Dependencies) > 0 {
		fmt.Println("\n" + strings.Repeat("-", 60))
		fmt.Println("🔗 完整依赖图:")
		for taskID, deps := range result.Dependencies {
			fmt.Printf("  %s 依赖于 %v\n", taskID, deps)
		}
	}
}

func printAnalysisResult(result *orchestrator.AnalysisResult) {
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("📊 AI分析结果")
	fmt.Println(strings.Repeat("═", 60))

	fmt.Printf("\n📌 需求概要: %s\n", result.Summary)
	fmt.Printf("🎯 复杂度: %s\n", result.Complexity)
	fmt.Printf("⏱️  预计时间: %s\n", result.EstimatedTime)

	// 显示模块
	fmt.Printf("\n🔧 模块拆分 (%d个模块):\n", len(result.Modules))
	for i, module := range result.Modules {
		fmt.Printf("  %d. %s\n", i+1, module.Name)
		fmt.Printf("     %s\n", module.Description)
		fmt.Printf("     优先级: %d | 文件: %v\n", module.Priority, module.Files)
	}

	// 显示任务
	fmt.Printf("\n📋 任务列表 (%d个任务):\n", len(result.Tasks))
	for i, task := range result.Tasks {
		icon := "🔵"
		if len(task.Dependencies) == 0 {
			icon = "🟢" // 无依赖，可立即执行
		}

		fmt.Printf("  %s Task-%d: %s\n", icon, i+1, task.Description)
		fmt.Printf("     模块: %s | 优先级: %d | 预计: %s\n",
			task.Module, task.Priority, task.Estimated)

		if len(task.Dependencies) > 0 {
			fmt.Printf("     依赖: %v\n", task.Dependencies)
		}
	}

	// 显示依赖关系
	if len(result.Dependencies) > 0 {
		fmt.Printf("\n🔗 依赖关系:\n")
		for taskID, deps := range result.Dependencies {
			fmt.Printf("  %s ← %v\n", taskID, deps)
		}
	}
}
