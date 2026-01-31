# Claude Swarm - 快速修复方案

**目标**: 修复所有 P0 严重问题
**预计时间**: 30-60 分钟

---

## 修复 1: tmux 会话异常终止检测

### 文件: `pkg/controller/coordinator.go`

#### 添加会话健康检查方法

```go
// isTmuxSessionAlive 检查 tmux 会话是否存活
func (c *Coordinator) isTmuxSessionAlive() bool {
    cmd := exec.Command("tmux", "has-session", "-t", c.sessionName)
    err := cmd.Run()
    return err == nil
}
```

#### 修改 monitorAgent 方法

在 `monitorAgent()` 函数中添加会话健康检查：

```go
func (c *Coordinator) monitorAgent(agent *Agent) {
    ticker := time.NewTicker(c.monitorInterval)
    defer ticker.Stop()

    sessionDead := false

    for {
        select {
        case <-ticker.C:
            // 🔧 FIX: 检查 tmux 会话是否存活
            if !sessionDead && !c.isTmuxSessionAlive() {
                log.Printf("❌ tmux 会话 '%s' 已终止，停止监控 %s", c.sessionName, agent.ID)
                sessionDead = true
                go func() {
                    time.Sleep(5 * time.Second)
                    if !c.isTmuxSessionAlive() {
                        log.Printf("⚠️  tmux 会话持续不可用，停止 coordinator")
                        os.Exit(1) // 优雅退出
                    }
                }()
                continue
            }

            // 如果会话已死，跳过所有操作
            if sessionDead {
                continue
            }

            // ... 现有的监控逻辑 ...
        }
    }
}
```

---

## 修复 2: worktrees 目录完全清理

### 文件: `cmd/swarm/stop.go`

#### 修改 cleanupWorktrees 方法

```go
func cleanupWorktrees() {
    fmt.Println("🧹 清理 worktrees...")

    // 获取当前目录
    cwd, err := os.Getwd()
    if err != nil {
        fmt.Printf("❌ 获取当前目录失败: %v\n", err)
        return
    }

    worktreesPath := filepath.Join(cwd, ".worktrees")

    // 列出所有 worktrees
    cmd := exec.Command("git", "worktree", "list")
    output, err := cmd.Output()
    if err != nil {
        fmt.Printf("⚠️  获取 worktree 列表失败: %v\n", err)
    } else {
        lines := strings.Split(strings.TrimSpace(string(output)), "\n")
        for _, line := range lines {
            if strings.Contains(line, ".worktrees/agent-") {
                parts := strings.Fields(line)
                if len(parts) >= 1 {
                    worktreePath := parts[0]

                    // 删除 worktree
                    cmd := exec.Command("git", "worktree", "remove", worktreePath, "--force")
                    if err := cmd.Run(); err != nil {
                        fmt.Printf("⚠️  删除 worktree 失败: %s\n", worktreePath)
                    } else {
                        fmt.Printf("✓ 删除 worktree: %s\n", worktreePath)
                    }
                }
            }
        }
    }

    // 删除所有 agent 分支
    cmd = exec.Command("git", "branch")
    output, err = cmd.Output()
    if err == nil {
        branches := strings.Split(strings.TrimSpace(string(output)), "\n")
        for _, branch := range branches {
            branch = strings.TrimSpace(strings.TrimPrefix(branch, "*"))
            if strings.Contains(branch, "agent-") && strings.Contains(branch, "-branch") {
                cmd := exec.Command("git", "branch", "-D", branch)
                if err := cmd.Run(); err != nil {
                    fmt.Printf("⚠️  删除分支失败: %s\n", branch)
                } else {
                    fmt.Printf("✓ 删除分支: %s\n", branch)
                }
            }
        }
    }

    // 🔧 FIX: 确保删除整个 .worktrees 目录
    if _, err := os.Stat(worktreesPath); err == nil {
        // 先尝试删除目录中的所有内容
        entries, err := os.ReadDir(worktreesPath)
        if err == nil {
            for _, entry := range entries {
                entryPath := filepath.Join(worktreesPath, entry.Name())
                if err := os.RemoveAll(entryPath); err != nil {
                    fmt.Printf("⚠️  删除 %s 失败: %v\n", entryPath, err)
                }
            }
        }

        // 删除目录本身
        if err := os.RemoveAll(worktreesPath); err != nil {
            fmt.Printf("⚠️  删除 .worktrees 目录失败: %v\n", err)
        } else {
            fmt.Printf("✓ 已删除 .worktrees 目录\n")
        }
    } else {
        fmt.Printf("✓ .worktrees 目录不存在\n")
    }

    fmt.Println("✓ 清理完成")
}
```

---

## 修复 3: 进程清理改进

### 文件: `cmd/swarm/start.go`

#### 添加进程组管理

```go
func runStart() {
    // ... 现有的初始化代码 ...

    // 设置进程组，确保子进程可以一起被终止
    cmd := exec.Command("tmux", "new-session", "-d", "-s", session)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Setpgid: true,
    }

    // ... 继续现有代码 ...
}
```

### 文件: `cmd/swarm/stop.go`

#### 改进进程终止逻辑

```go
func runStop() {
    session := "claude-swarm"
    fmt.Printf("🛑 停止 tmux 会话: %s...\n", session)

    // 检查会话是否存在
    checkCmd := exec.Command("tmux", "has-session", "-t", session)
    if err := checkCmd.Run(); err != nil {
        fmt.Printf("⚠️  会话 %s 不存在或已停止\n", session)
        cleanupWorktrees()
        cleanupPidFile(session)

        // 🔧 FIX: 即使会话不存在，也尝试清理遗留进程
        killOrphanedProcesses()
        return
    }

    // 清理 worktrees（在杀死会话之前）
    cleanupWorktrees()

    // 杀死 tmux 会话
    killCmd := exec.Command("tmux", "kill-session", "-t", session)
    if err := killCmd.Run(); err != nil {
        fmt.Printf("❌ 停止会话失败: %v\n", err)
        cleanupPidFile(session)
        killOrphanedProcesses()
        return
    }

    // 清理 PID 文件
    cleanupPidFile(session)

    // 🔧 FIX: 等待并确保所有进程退出
    time.Sleep(2 * time.Second)
    killOrphanedProcesses()

    fmt.Println("✓ 已停止")
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
    for _, pidStr := range pids {
        if pidStr == "" {
            continue
        }

        pid, err := strconv.Atoi(pidStr)
        if err != nil {
            continue
        }

        // 跳过当前进程
        if pid == os.Getpid() {
            continue
        }

        fmt.Printf("🧹 清理遗留进程: PID %d\n", pid)

        // 尝试优雅终止
        killCmd := exec.Command("kill", "-TERM", pidStr)
        if err := killCmd.Run(); err == nil {
            time.Sleep(1 * time.Second)

            // 检查进程是否还存在
            if checkCmd := exec.Command("kill", "-0", pidStr); checkCmd.Run() != nil {
                // 进程已退出
                continue
            }
        }

        // 强制终止
        fmt.Printf("⚠️  强制终止进程: PID %d\n", pid)
        killCmd = exec.Command("kill", "-9", pidStr)
        _ = killCmd.Run()
    }
}
```

---

## 修复 4: 边界情况验证

### 文件: `cmd/swarm/start.go`

#### 添加参数验证

在 `runStart()` 函数开头添加：

```go
func runStart() {
    // 🔧 FIX: 验证 agent 数量
    if numAgents <= 0 {
        fmt.Println("❌ Agent 数量必须大于 0")
        os.Exit(1)
    }

    if numAgents > 100 {
        fmt.Printf("⚠️  Agent 数量过多 (%d)，建议不超过 100 个\n", numAgents)
        fmt.Print("是否继续? (y/N): ")

        var response string
        fmt.Scanln(&response)
        if strings.ToLower(response) != "y" {
            fmt.Println("已取消")
            os.Exit(0)
        }
    }

    // ... 继续现有代码 ...
}
```

### 文件: `cmd/swarm/add.go`

#### 添加任务描述验证

```go
func runAddTask(cmd *cobra.Command, args []string) {
    if len(args) == 0 {
        fmt.Println("❌ 请提供任务描述")
        os.Exit(1)
    }

    description := strings.TrimSpace(args[0])

    // 🔧 FIX: 验证任务描述
    if description == "" {
        fmt.Println("❌ 任务描述不能为空")
        os.Exit(1)
    }

    if len(description) > 10000 {
        fmt.Printf("⚠️  任务描述过长 (%d 字符)，建议不超过 10000 字符\n", len(description))
        fmt.Print("是否继续? (y/N): ")

        var response string
        fmt.Scanln(&response)
        if strings.ToLower(response) != "y" {
            fmt.Println("已取消")
            os.Exit(0)
        }
    }

    // ... 继续现有代码 ...
}
```

---

## 应用修复

### 步骤 1: 备份当前代码

```bash
git checkout -b fix/critical-issues
```

### 步骤 2: 应用修复

按照上述修复方案，依次修改文件：
1. `pkg/controller/coordinator.go`
2. `cmd/swarm/stop.go`
3. `cmd/swarm/start.go`
4. `cmd/swarm/add.go`

### 步骤 3: 编译测试

```bash
go build -o swarm ./cmd/swarm
./comprehensive-test.sh
```

### 步骤 4: 验证修复

运行以下测试场景：

```bash
# 测试 1: tmux 会话异常终止
./swarm start --agents 2 &
sleep 3
tmux kill-session -t claude-swarm
sleep 5
ps aux | grep swarm  # 应该没有遗留进程

# 测试 2: 清理完整性
./swarm start --agents 2
sleep 5
./swarm stop
ls -la .worktrees  # 应该不存在

# 测试 3: 边界情况
./swarm start --agents 0  # 应该报错
./swarm start --agents -1  # 应该报错
./swarm add-task ""  # 应该报错
```

### 步骤 5: 提交修复

```bash
git add -A
git commit -m "修复 P0 严重问题

- 添加 tmux 会话健康检查，防止访问死亡的 pane
- 完全清理 worktrees 目录和遗留文件
- 改进进程清理，确保没有遗留进程
- 添加参数边界验证，防止无效输入

修复 #1, #2, #3, #5"
```

---

## 预期结果

修复后，所有 P0 问题应该得到解决：

- ✅ tmux 会话异常终止时，coordinator 优雅退出
- ✅ stop 命令完全清理所有资源
- ✅ 没有遗留进程
- ✅ 参数验证防止无效输入

---

## 下一步

1. 运行完整测试套件，确保所有测试通过
2. 更新文档，说明新的行为和限制
3. 开始修复 P1 问题（任务完成检测）

---

*修复方案版本: 1.0*
*创建时间: 2026-01-30*
