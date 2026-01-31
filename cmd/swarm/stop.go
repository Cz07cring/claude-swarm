package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	forceStop bool
	keepWork  bool
	noClean   bool
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

	stopCmd.Flags().BoolVarP(&forceStop, "force", "f", false, "强制停止，不检查未提交的更改")
	stopCmd.Flags().BoolVarP(&keepWork, "keep-work", "k", false, "停止但保留 worktrees（不删除工作目录）")
	stopCmd.Flags().BoolVar(&noClean, "no-clean", false, "仅停止进程，不清理任何资源")
}

func runStop() {
	session := "claude-swarm"
	if sessionName != "" {
		session = sessionName
	}

	// 1. 检查未提交的更改（除非 --force）
	if !forceStop && !keepWork && !noClean {
		hasChanges, changes := checkUncommittedChanges()
		if hasChanges {
			fmt.Println("⚠️  检测到未提交的更改：")
			listUncommittedChanges(changes)
			fmt.Println()
			fmt.Print("确定要停止吗？未提交的工作将丢失 (y/N): ")

			var response string
			fmt.Scanln(&response)
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				fmt.Println("✓ 已取消")
				fmt.Println()
				fmt.Println("💡 提示：")
				fmt.Println("   - 使用 --force 强制停止")
				fmt.Println("   - 使用 --keep-work 保留工作目录")
				fmt.Println("   - 或者先提交更改，然后再停止")
				return
			}
		}
	}

	fmt.Printf("🛑 停止 tmux 会话: %s...\n", session)

	// 2. 检查会话是否存在
	checkCmd := exec.Command("tmux", "has-session", "-t", session)
	if err := checkCmd.Run(); err != nil {
		fmt.Printf("⚠️  会话 %s 不存在或已停止\n", session)
		// 即使会话不存在，也尝试清理资源（如果需要）
		if !noClean {
			if !keepWork {
				cleanupWorktrees()
			}
			cleanupPidFile(session)
			killOrphanedProcesses()
		}
		return
	}

	// 3. 终止 tmux 会话
	killCmd := exec.Command("tmux", "kill-session", "-t", session)
	if err := killCmd.Run(); err != nil {
		fmt.Printf("❌ 停止会话失败: %v\n", err)
		if !noClean {
			cleanupPidFile(session)
			killOrphanedProcesses()
		}
		return
	}

	// 4. 等待进程优雅退出
	fmt.Println("⏳ 等待进程优雅退出...")
	time.Sleep(10 * time.Second)

	// 5. 清理资源（根据选项）
	if !noClean {
		if !keepWork {
			cleanupWorktrees()
		} else {
			fmt.Println("💾 保留 worktrees（使用了 --keep-work 选项）")
		}
		cleanupPidFile(session)
		killOrphanedProcesses()
	} else {
		fmt.Println("⚠️  跳过资源清理（使用了 --no-clean 选项）")
	}

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

	// 🔧 FIX: 确保完全删除 .worktrees 目录
	if _, err := os.Stat(worktreeRoot); err == nil {
		// 目录存在，先尝试清理所有内容
		entries, err := os.ReadDir(worktreeRoot)
		if err == nil && len(entries) > 0 {
			fmt.Printf("🧹 清理 .worktrees 目录中的残留文件 (%d 项)...\n", len(entries))
			for _, entry := range entries {
				entryPath := filepath.Join(worktreeRoot, entry.Name())
				if err := os.RemoveAll(entryPath); err != nil {
					fmt.Printf("⚠️  删除 %s 失败: %v\n", entry.Name(), err)
				} else {
					fmt.Printf("✓ 删除: %s\n", entry.Name())
				}
			}
		}

		// 删除目录本身
		if err := os.RemoveAll(worktreeRoot); err != nil {
			fmt.Printf("⚠️  删除 .worktrees 目录失败: %v\n", err)
			// 即使失败也不阻止后续操作
		} else {
			fmt.Printf("✓ 已删除 .worktrees 目录\n")
		}
	}

	fmt.Println("✓ 清理完成")
}

// killOrphanedProcesses 清理遗留的 swarm 进程
func killOrphanedProcesses() {
	// 查找所有 swarm 进程
	cmd := exec.Command("pgrep", "-f", "swarm start")
	output, err := cmd.Output()
	if err != nil {
		// 没有找到进程，这是好事
		return
	}

	pids := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(pids) == 0 || (len(pids) == 1 && pids[0] == "") {
		return
	}

	currentPID := os.Getpid()
	hasOrphans := false

	for _, pidStr := range pids {
		if pidStr == "" {
			continue
		}

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		// 跳过当前进程
		if pid == currentPID {
			continue
		}

		if !hasOrphans {
			fmt.Println("🧹 发现遗留进程，正在清理...")
			hasOrphans = true
		}

		fmt.Printf("   终止进程: PID %d\n", pid)

		// Step 1: 尝试优雅终止 (SIGTERM)
		killCmd := exec.Command("kill", "-TERM", pidStr)
		if err := killCmd.Run(); err == nil {
			// Step 2: 等待进程退出（5 秒）
			time.Sleep(5 * time.Second)

			// Step 3: 检查进程是否还存在
			checkCmd := exec.Command("kill", "-0", pidStr)
			if checkCmd.Run() != nil {
				// 进程已优雅退出
				fmt.Printf("   ✓ 进程 %d 已优雅退出\n", pid)
				continue
			}

			// Step 4: 再等待 5 秒
			time.Sleep(5 * time.Second)

			// Step 5: 再次检查
			if checkCmd.Run() != nil {
				fmt.Printf("   ✓ 进程 %d 已退出\n", pid)
				continue
			}
		}

		// Step 6: 强制终止 (SIGKILL) - 总共等待了 10 秒
		fmt.Printf("   ⚠️  强制终止进程 %d (SIGKILL)\n", pid)
		killCmd = exec.Command("kill", "-9", pidStr)
		if err := killCmd.Run(); err != nil {
			fmt.Printf("   ❌ 无法终止进程 %d: %v\n", pid, err)
		} else {
			fmt.Printf("   ✓ 进程 %d 已强制终止\n", pid)
		}
	}

	if hasOrphans {
		fmt.Println("✓ 遗留进程清理完成")
	}
}

// checkUncommittedChanges 检查所有 worktrees 是否有未提交的更改
func checkUncommittedChanges() (bool, map[string][]string) {
	worktreesDir := ".worktrees"
	uncommittedFiles := make(map[string][]string)

	// 检查 .worktrees 目录是否存在
	if _, err := os.Stat(worktreesDir); os.IsNotExist(err) {
		return false, nil
	}

	// 列出所有 agent worktrees
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		return false, nil
	}

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "agent-") {
			continue
		}

		worktreePath := filepath.Join(worktreesDir, entry.Name())

		// 检查 git status
		cmd := exec.Command("git", "-C", worktreePath, "status", "--short")
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		outputStr := strings.TrimSpace(string(output))
		if len(outputStr) > 0 {
			files := strings.Split(outputStr, "\n")
			uncommittedFiles[entry.Name()] = files
		}
	}

	return len(uncommittedFiles) > 0, uncommittedFiles
}

// listUncommittedChanges 列出未提交的更改
func listUncommittedChanges(changes map[string][]string) {
	for agent, files := range changes {
		fmt.Printf("  📁 %s:\n", agent)
		for _, file := range files {
			fmt.Printf("     %s\n", file)
		}
	}
}
