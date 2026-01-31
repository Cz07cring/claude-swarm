# Stop 行为改进方案

## 当前问题分析

### ⚠️ 存在的问题

1. **数据安全问题**
   - ❌ 直接删除 worktrees，可能丢失未提交的工作
   - ❌ 没有检查是否有未保存的更改
   - ❌ 没有给用户保存工作的机会

2. **进程终止问题**
   - ⚠️ SIGTERM 后只等待 1 秒就 SIGKILL
   - ⚠️ 没有给 coordinator 足够时间保存状态

3. **用户体验问题**
   - ❌ 没有确认步骤，容易误操作
   - ❌ 没有选项让用户控制清理行为

---

## 改进方案

### 方案 1：安全停止（推荐） ✅

添加检查和确认步骤：

```go
func runStop() {
    // 1. 检查是否有未提交的更改
    hasUncommittedChanges := checkUncommittedChanges()

    if hasUncommittedChanges {
        fmt.Println("⚠️  检测到未提交的更改：")
        listUncommittedChanges()
        fmt.Println()
        fmt.Print("确定要停止吗？这将丢失未提交的工作 (y/N): ")

        var response string
        fmt.Scanln(&response)

        if !strings.EqualFold(response, "y") && !strings.EqualFold(response, "yes") {
            fmt.Println("已取消")
            return
        }
    }

    // 2. 优雅关闭 coordinator（给 10 秒而不是 2 秒）
    fmt.Println("🛑 正在停止 coordinator...")
    sendGracefulShutdown()
    time.Sleep(10 * time.Second)

    // 3. 然后才清理资源
    cleanupWorktrees()
    // ...
}
```

### 方案 2：添加命令选项

```bash
# 安全停止（默认）- 检查未提交的更改
./swarm stop

# 强制停止 - 不检查，直接清理
./swarm stop --force

# 保留工作 - 停止但保留 worktrees
./swarm stop --keep-work

# 仅停止 - 不清理任何东西
./swarm stop --no-clean
```

### 方案 3：分离停止和清理

```bash
# 停止运行但保留所有工作
./swarm stop

# 单独的清理命令
./swarm clean          # 清理 worktrees
./swarm clean --all    # 清理所有（worktrees + 进程）
```

---

## 推荐的最佳实践

### 停止流程应该是：

1. **发送停止信号** → coordinator 收到信号
2. **等待 coordinator 保存状态** → 10-30 秒
3. **检查未提交的更改** → 如果有，询问用户
4. **优雅关闭 tmux 会话** → 发送 SIGTERM
5. **等待进程退出** → 最多 5 秒
6. **清理临时文件** → PID 文件等
7. **可选：清理 worktrees** → 根据用户选择

### 进程终止应该是：

```go
// 1. SIGTERM（优雅终止）
sendSignal(pid, SIGTERM)
time.Sleep(5 * time.Second)  // 等待 5 秒而不是 1 秒

// 2. 再次检查
if processExists(pid) {
    time.Sleep(5 * time.Second)  // 再等待 5 秒
}

// 3. SIGKILL（强制终止）- 总共等待 10 秒
if processExists(pid) {
    sendSignal(pid, SIGKILL)
}
```

---

## 具体改进代码

### 1. 添加未提交更改检查

```go
// checkUncommittedChanges 检查所有 worktrees 是否有未提交的更改
func checkUncommittedChanges() (bool, map[string][]string) {
    worktreesDir := ".worktrees"
    uncommittedFiles := make(map[string][]string)

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

        if len(strings.TrimSpace(string(output))) > 0 {
            files := strings.Split(strings.TrimSpace(string(output)), "\n")
            uncommittedFiles[entry.Name()] = files
        }
    }

    return len(uncommittedFiles) > 0, uncommittedFiles
}

// listUncommittedChanges 列出未提交的更改
func listUncommittedChanges(changes map[string][]string) {
    for agent, files := range changes {
        fmt.Printf("  %s:\n", agent)
        for _, file := range files {
            fmt.Printf("    %s\n", file)
        }
    }
}
```

### 2. 添加选项支持

```go
var (
    forceStop  bool
    keepWork   bool
    noClean    bool
)

func init() {
    rootCmd.AddCommand(stopCmd)

    stopCmd.Flags().BoolVar(&forceStop, "force", false, "强制停止，不检查未提交的更改")
    stopCmd.Flags().BoolVar(&keepWork, "keep-work", false, "停止但保留 worktrees")
    stopCmd.Flags().BoolVar(&noClean, "no-clean", false, "仅停止，不清理任何资源")
}

func runStop() {
    // 1. 检查未提交的更改（除非 --force）
    if !forceStop {
        hasChanges, changes := checkUncommittedChanges()
        if hasChanges {
            fmt.Println("⚠️  检测到未提交的更改：")
            listUncommittedChanges(changes)
            fmt.Println()
            fmt.Print("确定要停止吗？未提交的工作将丢失 (y/N): ")

            var response string
            fmt.Scanln(&response)

            if !strings.EqualFold(response, "y") && !strings.EqualFold(response, "yes") {
                fmt.Println("✓ 已取消")
                fmt.Println()
                fmt.Println("提示：")
                fmt.Println("  - 使用 --force 强制停止")
                fmt.Println("  - 使用 --keep-work 保留工作目录")
                fmt.Println("  - 先提交更改，然后再停止")
                return
            }
        }
    }

    // 2. 停止 tmux 会话
    fmt.Printf("🛑 停止 tmux 会话: %s...\n", session)
    killSession(session)

    // 3. 等待 coordinator 优雅关闭
    fmt.Println("⏳ 等待进程优雅退出...")
    time.Sleep(10 * time.Second)

    // 4. 清理资源（根据选项）
    if !noClean {
        if !keepWork {
            cleanupWorktrees()
        }
        cleanupPidFile(session)
        killOrphanedProcesses()
    }

    fmt.Println("✓ 已停止")
}
```

### 3. 改进进程终止逻辑

```go
func killOrphanedProcesses() {
    // ... 找到进程 ...

    for _, pidStr := range pids {
        // ...

        // Step 1: SIGTERM（优雅终止）
        fmt.Printf("   发送 SIGTERM 到进程 %d...\n", pid)
        killCmd := exec.Command("kill", "-TERM", pidStr)
        killCmd.Run()

        // Step 2: 等待 5 秒
        fmt.Printf("   等待进程退出 (5s)...\n")
        time.Sleep(5 * time.Second)

        // Step 3: 检查进程是否退出
        checkCmd := exec.Command("kill", "-0", pidStr)
        if checkCmd.Run() != nil {
            fmt.Printf("   ✓ 进程 %d 已优雅退出\n", pid)
            continue
        }

        // Step 4: 再等待 5 秒
        fmt.Printf("   进程仍在运行，再等待 5s...\n")
        time.Sleep(5 * time.Second)

        // Step 5: 最后检查
        if checkCmd.Run() != nil {
            fmt.Printf("   ✓ 进程 %d 已退出\n", pid)
            continue
        }

        // Step 6: 强制终止 (SIGKILL)
        fmt.Printf("   ⚠️  强制终止进程 %d (SIGKILL)\n", pid)
        killCmd = exec.Command("kill", "-9", pidStr)
        if err := killCmd.Run(); err != nil {
            fmt.Printf("   ❌ 无法终止进程 %d: %v\n", pid, err)
        } else {
            fmt.Printf("   ✓ 进程 %d 已强制终止\n", pid)
        }
    }
}
```

---

## 实施建议

### 短期改进（立即实施）

1. ✅ 增加等待时间：SIGTERM 后等待 10 秒而不是 1 秒
2. ✅ 添加 `--force` 选项快速停止
3. ✅ 添加 `--keep-work` 保留工作目录

### 中期改进（建议实施）

1. ✅ 检查未提交的更改
2. ✅ 用户确认提示
3. ✅ 分离 stop 和 clean 命令

### 长期改进（可选）

1. 自动备份未提交的工作
2. 提供恢复机制
3. 更详细的日志和错误处理

---

## 使用示例

### 改进后的使用方式

```bash
# 安全停止（检查未提交的更改）
./swarm stop
# ⚠️  检测到未提交的更改：
#   agent-0:
#     M pkg/tui/tasklist.go
# 确定要停止吗？未提交的工作将丢失 (y/N): n
# ✓ 已取消

# 强制停止
./swarm stop --force
# ✓ 已停止

# 保留工作目录
./swarm stop --keep-work
# 🛑 停止 tmux 会话: claude-swarm...
# ✓ 已停止（worktrees 已保留）

# 仅停止，不清理
./swarm stop --no-clean
# ✓ 已停止（未清理资源）
```

---

## 总结

### 当前行为的问题

- ❌ 不安全：直接删除可能有未保存工作的目录
- ❌ 不够优雅：进程终止太快
- ❌ 不够灵活：没有选项控制清理行为

### 改进后的优势

- ✅ 更安全：检查未提交的更改，给用户选择
- ✅ 更优雅：充分等待进程退出（10 秒）
- ✅ 更灵活：提供多个选项控制行为
- ✅ 更友好：清晰的提示和确认

### 推荐实施优先级

1. **高优先级**：增加等待时间，添加 --force 选项
2. **中优先级**：未提交更改检查，用户确认
3. **低优先级**：分离命令，自动备份

这样的改进将使 `swarm stop` 更加健壮和用户友好！
