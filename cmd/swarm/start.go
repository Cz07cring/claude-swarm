package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/claude-swarm/pkg/controller"
)

var (
	numAgents       int
	sessionName     string
	taskQueuePath   string
	monitorInterval int
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "启动 Agent 集群",
	Long:  `启动指定数量的 Claude Agent 并开始监控和任务调度`,
	Run: func(cmd *cobra.Command, args []string) {
		runStart()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().IntVarP(&numAgents, "agents", "n", 3, "Agent 数量")
	startCmd.Flags().StringVarP(&sessionName, "session", "s", "claude-swarm", "tmux 会话名称")
	startCmd.Flags().StringVarP(&taskQueuePath, "queue", "q", "~/.claude-swarm/tasks.json", "任务队列文件路径")
	startCmd.Flags().IntVarP(&monitorInterval, "interval", "i", 5, "监控间隔（秒）")
}

func runStart() {
	log.SetFlags(0) // 移除时间戳前缀

	fmt.Println("🚀 启动 Claude Agent Swarm...")
	fmt.Println()

	// 检查 PID 文件锁
	pidFile := getPidFilePath(sessionName)
	if err := checkPidLock(pidFile); err != nil {
		log.Fatalf("❌ %v", err)
	}

	// 创建 PID 文件
	if err := writePidFile(pidFile); err != nil {
		log.Fatalf("❌ 创建 PID 文件失败: %v", err)
	}
	defer os.Remove(pidFile)

	// 创建协调器配置
	config := controller.CoordinatorConfig{
		NumAgents:       numAgents,
		SessionName:     sessionName,
		TaskQueuePath:   taskQueuePath,
		MonitorInterval: time.Duration(monitorInterval) * time.Second,
	}

	// 创建协调器
	coord, err := controller.NewCoordinator(config)
	if err != nil {
		log.Fatalf("❌ 创建协调器失败: %v", err)
	}

	// 启动协调器
	coord.Start()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("按 Ctrl+C 停止...")
	<-sigChan

	fmt.Println("\n\n⏹️  停止中...")
	if err := coord.Stop(); err != nil {
		log.Fatalf("❌ 停止协调器失败: %v", err)
	}

	fmt.Println("✓ 已停止")
}

// getPidFilePath 返回 PID 文件路径
func getPidFilePath(sessionName string) string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".claude-swarm", fmt.Sprintf("%s.pid", sessionName))
}

// checkPidLock 检查是否已有进程在运行
func checkPidLock(pidFile string) error {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // PID 文件不存在，可以启动
		}
		return fmt.Errorf("读取 PID 文件失败: %w", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		// PID 文件损坏，删除并继续
		os.Remove(pidFile)
		return nil
	}

	// 检查进程是否存在
	process, err := os.FindProcess(pid)
	if err != nil {
		// 进程不存在，删除旧 PID 文件
		os.Remove(pidFile)
		return nil
	}

	// 尝试发送信号 0 检查进程是否真的在运行
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// 进程已死，删除旧 PID 文件
		os.Remove(pidFile)
		return nil
	}

	return fmt.Errorf("Swarm 已在运行中 (PID: %d)。请先运行 'swarm stop' 停止现有实例", pid)
}

// writePidFile 写入当前进程的 PID
func writePidFile(pidFile string) error {
	// 确保目录存在
	dir := filepath.Dir(pidFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 写入 PID
	pid := os.Getpid()
	return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}
