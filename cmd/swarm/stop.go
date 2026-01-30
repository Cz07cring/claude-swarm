package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "停止 Agent 集群",
	Long:  `停止正在运行的 Claude Agent 集群并终止 tmux 会话`,
	Run: func(cmd *cobra.Command, args []string) {
		runStop()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop() {
	session := "claude-swarm"
	if sessionName != "" {
		session = sessionName
	}

	fmt.Printf("🛑 停止 tmux 会话: %s...\n", session)

	// 检查会话是否存在
	checkCmd := exec.Command("tmux", "has-session", "-t", session)
	if err := checkCmd.Run(); err != nil {
		fmt.Printf("⚠️  会话 %s 不存在或已停止\n", session)
		// 即使会话不存在，也尝试清理 worktrees
		cleanupWorktrees()
		cleanupPidFile(session)
		return
	}

	// 先清理 worktrees，再终止会话
	cleanupWorktrees()

	// 终止会话
	killCmd := exec.Command("tmux", "kill-session", "-t", session)
	if err := killCmd.Run(); err != nil {
		fmt.Printf("❌ 停止会话失败: %v\n", err)
		cleanupPidFile(session)
		return
	}

	// 清理 PID 文件
	cleanupPidFile(session)

	fmt.Println("✓ 已停止")
}

// cleanupPidFile 清理 PID 文件
func cleanupPidFile(sessionName string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	pidFile := filepath.Join(homeDir, ".claude-swarm", fmt.Sprintf("%s.pid", sessionName))
	if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("⚠️  清理 PID 文件失败: %v\n", err)
	}
}

func cleanupWorktrees() {
	fmt.Println("🧹 清理 worktrees...")

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("⚠️  获取工作目录失败: %v\n", err)
		return
	}

	worktreeRoot := filepath.Join(cwd, ".worktrees")

	// 检查 .worktrees 目录是否存在
	if _, err := os.Stat(worktreeRoot); os.IsNotExist(err) {
		// 没有 worktrees，无需清理
		return
	}

	// 列出所有 worktrees
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("⚠️  列出 worktrees 失败: %v\n", err)
		return
	}

	// 解析 worktrees 并找到 agent worktrees
	lines := strings.Split(string(output), "\n")
	var worktreePaths []string
	var agentBranches []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			if strings.Contains(path, ".worktrees/agent-") {
				worktreePaths = append(worktreePaths, path)
			}
		} else if strings.HasPrefix(line, "branch ") {
			branch := strings.TrimPrefix(line, "branch refs/heads/")
			if strings.HasPrefix(branch, "agent-") && strings.HasSuffix(branch, "-branch") {
				agentBranches = append(agentBranches, branch)
			}
		}
	}

	// 删除 worktrees
	for _, path := range worktreePaths {
		cmd := exec.Command("git", "worktree", "remove", path, "--force")
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  删除 worktree %s 失败: %v\n", path, err)
		} else {
			fmt.Printf("✓ 删除 worktree: %s\n", path)
		}
	}

	// 删除分支
	for _, branch := range agentBranches {
		cmd := exec.Command("git", "branch", "-D", branch)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  删除分支 %s 失败: %v\n", branch, err)
		} else {
			fmt.Printf("✓ 删除分支: %s\n", branch)
		}
	}

	// 删除 .worktrees 目录
	if err := os.RemoveAll(worktreeRoot); err != nil {
		fmt.Printf("⚠️  删除 .worktrees 目录失败: %v\n", err)
	} else {
		fmt.Printf("✓ 清理完成\n")
	}
}
