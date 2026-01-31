# Claude Swarm - 问题检测报告

**生成时间**: 2026-01-30
**测试范围**: 全局功能、边界情况、错误处理、资源管理

---

## 🔥 严重问题 (Critical)

### 1. tmux 会话异常终止后，coordinator 进程未优雅退出

**问题描述**:
当 tmux 会话被意外终止（如 `tmux kill-session`）后，swarm coordinator 进程仍然继续运行，尝试访问不存在的 tmux panes，导致：
- 频繁的 "Error capturing agent output: failed to capture pane: exit status 1"
- 频繁的 "Error sending confirmation: failed to send keys: exit status 1"
- CPU 和日志资源浪费

**复现步骤**:
```bash
./swarm start --agents 2 &
sleep 3
tmux kill-session -t claude-swarm
# swarm 进程仍在运行并持续报错
```

**影响**:
- 资源泄漏
- 日志污染
- 无法正常使用 swarm（需要手动 kill -9）

**建议修复**:
在 `pkg/controller/coordinator.go` 的 `monitorAgent()` 中添加 tmux 会话健康检查：

```go
func (c *Coordinator) monitorAgent(agent *Agent) {
    for {
        select {
        case <-time.After(c.monitorInterval):
            // 检查 tmux 会话是否存活
            if !c.isTmuxSessionAlive() {
                log.Printf("⚠️  tmux 会话已终止，停止监控")
                c.Stop()
                return
            }

            // 现有的监控逻辑...
        }
    }
}

func (c *Coordinator) isTmuxSessionAlive() bool {
    cmd := exec.Command("tmux", "has-session", "-t", c.sessionName)
    return cmd.Run() == nil
}
```

---

### 2. worktrees 目录清理不完整

**问题描述**:
执行 `./swarm stop` 后，`.worktrees/` 目录虽然删除了 Git worktrees，但目录本身仍然存在且包含残留文件（如 `.worktrees/agent-0`）。

**复现步骤**:
```bash
./swarm start --agents 2
sleep 5
./swarm stop
ls -la .worktrees/  # 目录存在且可能包含残留
```

**影响**:
- 磁盘空间占用
- 可能影响下次启动
- Git worktree 状态不一致

**建议修复**:
在 `cmd/swarm/stop.go` 的 `cleanupWorktrees()` 中，确保删除整个目录：

```go
func cleanupWorktrees() {
    // ... 现有的 worktree remove 逻辑 ...

    // 确保删除整个 .worktrees 目录
    worktreesDir := filepath.Join(repoPath, ".worktrees")
    if err := os.RemoveAll(worktreesDir); err != nil {
        log.Printf("⚠️  删除 .worktrees 目录失败: %v", err)
    } else {
        log.Printf("✓ 已删除 .worktrees 目录")
    }
}
```

---

### 3. 进程清理不完整（PID 跟踪问题）

**问题描述**:
使用 `pkill -f "./swarm start"` 无法完全清理 swarm 进程，导致测试中出现 "遗留 swarm 进程未清理"。

**原因分析**:
- swarm 可能以不同的路径运行（如 `./swarm` vs `/full/path/swarm`）
- 子进程可能没有正确传播信号

**建议修复**:
1. 在 start 命令中保存进程组 ID，stop 时使用 `kill -TERM -$PGID` 杀死整个进程组
2. 添加进程清理验证：

```go
func (c *Coordinator) Stop() {
    // 现有的停止逻辑...

    // 等待子进程退出
    time.Sleep(2 * time.Second)

    // 验证所有子进程是否已退出
    if !c.verifyCleanShutdown() {
        log.Printf("⚠️  部分子进程未退出，强制终止")
        c.forceKillAll()
    }
}
```

---

## ⚠️  中等问题 (Medium)

### 4. 任务完成检测可能失败

**问题描述**:
在某些情况下，任务完成后没有触发合并，日志中没有 "检测到任务完成"、"开始合并" 等关键信息。

**可能原因**:
- Agent 状态检测不准确（Detector 误判）
- 状态转换时机不对
- Claude CLI 输出格式变化

**建议修复**:
1. 增强状态检测的日志记录
2. 添加状态转换的调试模式
3. 实现任务超时机制（如果 30 分钟未完成，标记为超时）

---

### 5. 缺少边界情况验证

**问题描述**:
以下边界情况没有正确处理：
- 启动 0 个 agent（应该拒绝）
- 启动负数 agent（应该拒绝）
- 空任务描述（应该拒绝）

**建议修复**:
在 `cmd/swarm/start.go` 中添加参数验证：

```go
func runStart() {
    // 验证 agent 数量
    if numAgents <= 0 {
        fmt.Println("❌ Agent 数量必须大于 0")
        os.Exit(1)
    }

    if numAgents > 100 {
        fmt.Println("⚠️  Agent 数量过多，建议不超过 100 个")
        // 可选：询问用户是否继续
    }

    // 继续启动...
}
```

在 `cmd/swarm/add.go` 中：

```go
func runAddTask(cmd *cobra.Command, args []string) {
    description := strings.TrimSpace(args[0])

    if description == "" {
        fmt.Println("❌ 任务描述不能为空")
        os.Exit(1)
    }

    if len(description) > 10000 {
        fmt.Println("⚠️  任务描述过长，建议不超过 10000 字符")
    }

    // 继续添加任务...
}
```

---

## ℹ️  低优先级问题 (Low)

### 6. 缺少版本信息命令

**问题描述**:
`./swarm --version` 命令不存在。

**建议修复**:
在 `cmd/swarm/main.go` 中添加版本信息：

```go
var (
    Version   = "v2.0.0"
    BuildTime = "unknown"
    GitCommit = "unknown"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "显示版本信息",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("Claude Swarm %s\n", Version)
        fmt.Printf("Build Time: %s\n", BuildTime)
        fmt.Printf("Git Commit: %s\n", GitCommit)
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
    rootCmd.Flags().BoolP("version", "v", false, "显示版本信息")
}
```

---

### 7. 日志过于冗余

**问题描述**:
coordinator 日志包含大量重复信息，如：
- 每 2 秒打印 "Scheduler check"
- 频繁的 "No pending tasks available"

**建议修复**:
添加日志级别控制，默认只显示重要信息：

```go
type LogLevel int

const (
    LogLevelError LogLevel = iota
    LogLevelWarn
    LogLevelInfo
    LogLevelDebug
)

var currentLogLevel = LogLevelInfo

func logDebug(format string, args ...interface{}) {
    if currentLogLevel >= LogLevelDebug {
        log.Printf(format, args...)
    }
}

// 使用：
logDebug("📅 Scheduler check: agent-0 state=%s", state)
```

---

### 8. 缺少配置文件验证

**问题描述**:
如果 `config.yaml` 存在但格式错误，程序可能崩溃或行为异常。

**建议修复**:
在 `pkg/config/config.go` 中添加配置验证：

```go
func LoadConfig(path string) (*Config, error) {
    cfg := &Config{}

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, fmt.Errorf("配置文件格式错误: %w", err)
    }

    // 验证配置
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("配置验证失败: %w", err)
    }

    return cfg, nil
}

func (c *Config) Validate() error {
    if c.MonitorInterval < 1 {
        return fmt.Errorf("monitor_interval 必须 >= 1 秒")
    }

    if c.DefaultAgents < 1 || c.DefaultAgents > 100 {
        return fmt.Errorf("default_agents 必须在 1-100 之间")
    }

    return nil
}
```

---

## 📊 测试覆盖情况

| 测试类别 | 已测试 | 未测试 | 覆盖率 |
|---------|--------|--------|--------|
| 环境检查 | 5/5 | 0 | 100% |
| 编译和基础功能 | 3/3 | 0 | 100% |
| 边界情况 | 0/4 | 4 | 0% |
| 进程管理 | 6/6 | 0 | 100% |
| 功能测试 | 4/6 | 2 | 67% |
| 错误恢复 | 2/3 | 1 | 67% |
| 清理验证 | 7/7 | 0 | 100% |
| 配置测试 | 2/3 | 1 | 67% |
| **总计** | **29/37** | **8** | **78%** |

---

## 🔧 推荐修复优先级

### P0 (立即修复)
1. ✅ tmux 会话异常终止检测
2. ✅ worktrees 目录清理不完整
3. ✅ 进程清理不完整

### P1 (本周修复)
4. ⚠️  任务完成检测失败
5. ⚠️  缺少边界情况验证

### P2 (下周修复)
6. 📝 缺少版本信息命令
7. 📝 日志过于冗余
8. 📝 缺少配置文件验证

---

## 📝 测试建议

### 添加单元测试

建议为以下模块添加单元测试：

1. **pkg/analyzer/detector.go**
   - 测试各种 Claude 输出模式的识别
   - 边界情况（空输出、超长输出、特殊字符）

2. **pkg/state/task_queue.go**
   - 并发访问测试
   - 任务状态转换测试
   - JSON 序列化/反序列化测试

3. **pkg/controller/coordinator.go**
   - 任务调度算法测试
   - Agent 状态转换测试
   - 错误恢复测试

### 添加集成测试

```bash
# tests/integration/test_basic_workflow.sh
# tests/integration/test_concurrent_tasks.sh
# tests/integration/test_error_recovery.sh
# tests/integration/test_cleanup.sh
```

### 添加性能测试

测试场景：
- 启动 50 个 agents 的性能
- 同时添加 100 个任务的性能
- 长时间运行（24 小时）的稳定性

---

## 📚 参考

- 测试日志: `/tmp/swarm-test-comprehensive.log`
- 原始测试脚本: `run-full-test.sh`
- 全面测试脚本: `comprehensive-test.sh`

---

## 🎯 下一步行动

1. **立即**: 修复 P0 严重问题
2. **本周**: 实现边界情况验证，修复任务完成检测
3. **下周**: 添加单元测试框架，实现基础测试用例
4. **本月**: 完善集成测试，达到 90% 测试覆盖率

---

*报告生成时间: 2026-01-30 21:30*
*测试工具版本: comprehensive-test v1.0*
