# 🐝 Claude Agent Swarm 蜂群系统使用指南

## 当前演示场景

已为 TUI Monitor 准备了 **4 个并行开发任务**，演示如何使用 3 个 Agent 并行工作。

### 📋 任务清单

1. **task-filter-001** - 任务过滤功能（1-1.5h）
   - 按状态过滤任务（1=全部，2=待处理，3=进行中，4=已完成）
   - 修改 `tasklist.go` 和 `dashboard.go`
   - 包含完整测试和优化

2. **task-stats-002** - Agent 性能统计（1.5-2h）
   - 显示每个 Agent 的任务完成数和平均耗时
   - 创建 `agentstats.go` 新组件
   - 集成到主界面

3. **task-help-003** - 帮助面板（1-1.5h）
   - 按 ? 键显示/隐藏快捷键帮助
   - 创建 `helppanel.go` 模态框组件
   - 实现交互式帮助系统

4. **task-theme-004** - 主题切换（1.5-2h）
   - 按 t 键切换亮色/暗色主题
   - 实现主题系统架构
   - 应用到所有组件

## 🚀 启动蜂群系统

### 步骤 1：查看任务队列

```bash
./swarm status
```

输出示例：
```
📊 Claude Agent Swarm 状态
============================================================

📋 任务队列: 4 个任务

  状态统计:
    待处理: 4

  最近任务:
    ⏳ task-filter-001 | 任务过滤功能
    ⏳ task-stats-002  | Agent性能统计
    ⏳ task-help-003   | 帮助面板
    ⏳ task-theme-004  | 主题切换
```

### 步骤 2：启动 3 个 Agent（终端 1）

```bash
./swarm start --agents 3
```

这将：
- ✓ 创建 tmux 会话 `claude-swarm`
- ✓ 启动 3 个 Claude Agent（独立窗格）
- ✓ 启动任务调度器
- ✓ 启动监控系统
- ✓ 自动分配任务给空闲 Agent

输出示例：
```
🚀 启动 Claude Agent Swarm...

✓ Created tmux session: claude-swarm
✓ Started agent-0 in pane 0
✓ Started agent-1 in pane 1
✓ Started agent-2 in pane 2
✓ Coordinator running...
  Monitor interval: 5s
  Agents: 3

Attach to session: tmux attach -t claude-swarm

按 Ctrl+C 停止...
```

### 步骤 3：启动 TUI 监控面板（终端 2）

在新终端窗口运行：

```bash
./swarm monitor
```

您将看到交互式可视化界面：

```
┌─────────────────────────────────────────────────────────────┐
│               Claude Agent Swarm Monitor                    │
├─────────────┬──────────────────┬──────────────────────────┤
│  Tasks ◄    │     Agents       │         Logs             │
│             │                  │                          │
│ ● task-001  │  ┌──────────┐   │  agent-0                 │
│ ○ task-002  │  │agent-0 ● │   │  State: working          │
│ ○ task-003  │  └──────────┘   │  Task: task-filter-001   │
│ ○ task-004  │  ┌──────────┐   │                          │
│             │  │agent-1 ● │   │  正在修改 tasklist.go... │
│             │  └──────────┘   │  添加 FilterMode 枚举    │
│             │  ┌──────────┐   │  实现 SetFilter 方法     │
│             │  │agent-2 ● │   │                          │
│             │  └──────────┘   │                          │
└─────────────┴──────────────────┴──────────────────────────┘
Tab: switch pane | ↑↓/jk: navigate | q: quit
```

### 步骤 4：观察实时工作

**在 TUI Monitor 中您会看到：**

1. **任务状态变化：**
   ```
   ⏳ pending → 🔄 in_progress → ✅ completed
   ```

2. **Agent 颜色编码：**
   - 🔵 蓝色边框 = 工作中 (working)
   - ⚪ 灰色边框 = 空闲 (idle)
   - 🟡 黄色边框 = 等待确认 (waiting_confirm)
   - 🔴 红色边框 = 错误 (error)

3. **实时日志流：**
   - Agent 的代码修改
   - 测试运行结果
   - 调试信息
   - 优化步骤

## 🎯 预期工作流程

### Agent-0：任务过滤功能

```
[09:00] 领取 task-filter-001
[09:02] 修改 pkg/tui/tasklist.go
        + 添加 FilterMode 枚举
        + 实现 SetFilter() 方法
[09:15] 修改 pkg/tui/dashboard.go
        + 处理数字键 1-4
[09:25] 运行 ./test-tui.sh 验证
[09:35] 测试所有过滤模式
[09:45] 性能优化
[09:50] 提交完成 ✓
[09:50] 领取 task-theme-004（如果还未完成）
```

### Agent-1：Agent 性能统计

```
[09:00] 领取 task-stats-002
[09:05] 创建 pkg/tui/agentstats.go
        + AgentStats 结构
        + calculateStats() 函数
        + AgentStatsView 组件
[09:40] 集成到 dashboard.go
[09:55] 添加样式 (styles.go)
[10:10] 测试统计计算
[10:25] 性能优化（缓存）
[10:30] 提交完成 ✓
```

### Agent-2：帮助面板

```
[09:00] 领取 task-help-003
[09:05] 创建 pkg/tui/helppanel.go
        + HelpPanelView 组件
        + Render() 方法
        + 快捷键列表
[09:30] 修改 dashboard.go
        + showHelp 字段
        + 处理 ? 键
[09:45] 添加模态框样式
[09:55] 测试显示/隐藏
[10:05] 验证不干扰正常操作
[10:10] 提交完成 ✓
```

## ⌨️ TUI Monitor 操作

### 键盘快捷键

- `Tab` - 切换面板（Tasks → Agents → Logs）
- `Shift+Tab` - 反向切换
- `↑` / `k` - 向上导航
- `↓` / `j` - 向下导航
- `←` / `h` - 向左导航（Agent 网格）
- `→` / `l` - 向右导航（Agent 网格）
- `Enter` - 选择当前项
- `q` / `Ctrl+C` - 退出 Monitor

### 使用技巧

1. **监控特定 Agent**：
   - 按 Tab 切换到 Agents 面板
   - 使用方向键选择 Agent
   - 右侧日志查看器显示该 Agent 输出

2. **跟踪特定任务**：
   - 在 Tasks 面板选择任务
   - 日志查看器显示执行该任务的 Agent

3. **快速浏览**：
   - Monitor 每 2 秒自动刷新
   - 无需手动刷新

## 📊 监控要点

### 正常工作流程

```
Agent 空闲 → 调度器分配任务 → Agent 开始工作 →
等待用户确认 → 自动确认 → 继续工作 →
任务完成 → 自动合并代码 → 回到空闲状态 →
领取下一个任务
```

### 异常处理

系统会自动检测和处理：
- ✓ **等待确认**：自动响应 Y/N 提示
- ✓ **错误状态**：标记为红色，记录日志
- ✓ **卡住状态**：检测无输出超时
- ✓ **任务失败**：标记任务状态，释放 Agent

## 🔍 查看原始输出（可选）

如果需要查看 Agent 的原始 tmux 输出：

```bash
# 附加到 tmux 会话
tmux attach -t claude-swarm

# 在 tmux 中操作：
# - Ctrl+B 然后 方向键：切换窗格
# - Ctrl+B 然后 D：分离（退出但保持运行）
# - Ctrl+B 然后 [：滚动模式（q 退出）
```

## 🛑 停止蜂群

```bash
./swarm stop
```

这将：
- ✓ 停止所有 Agent
- ✓ 清理 worktrees
- ✓ 杀死 tmux 会话
- ✓ 保存最终状态

## 📈 性能预期

### 串行 vs 并行对比

| 指标 | 串行执行 | 3 个 Agent 并行 | 提升 |
|------|----------|----------------|------|
| 总时间 | 5-7 小时 | ~2 小时 | 60-70% |
| Agent 利用率 | 33% | ~90% | +57% |
| 并行任务数 | 1 | 3 | 3x |
| 等待时间 | 高 | 低 | -80% |

### 资源使用

- **CPU**：3-5% per Agent（总计 ~15%）
- **内存**：~200MB per Agent（总计 ~600MB）
- **网络**：仅 API 调用（Claude、Gemini）

## 🎓 学习要点

### 蜂群系统的优势

1. **极速并行**：多个 Agent 同时工作
2. **智能调度**：自动任务分配和负载均衡
3. **自动化协助**：无需人工干预确认和错误处理
4. **可视化监控**：实时了解工作进度
5. **任务隔离**：独立 worktree 避免冲突

### 最佳实践

1. **任务拆分**：
   - 每个任务 30 分钟到 2 小时
   - 确保任务独立，可并行执行
   - 包含明确的验证步骤

2. **Agent 数量**：
   - 简单任务：3-5 个 Agent
   - 复杂项目：5-10 个 Agent
   - 根据任务数量调整

3. **监控频率**：
   - 定期查看 TUI Monitor
   - 关注错误和卡住状态
   - 及时调整任务描述

## 💡 故障排除

### 问题：Agent 未领取任务

**原因**：可能处于 waiting_confirm 或 error 状态

**解决**：
1. 在 TUI Monitor 中查看 Agent 状态
2. 查看日志了解原因
3. 如需要，手动附加 tmux 处理

### 问题：任务执行失败

**原因**：任务描述不够清晰或有依赖问题

**解决**：
1. 查看失败任务的错误日志
2. 修改任务描述使其更明确
3. 重新添加任务

### 问题：TUI Monitor 数据不更新

**原因**：状态文件读取问题

**解决**：
1. 检查 `~/.claude-swarm/agents.json` 存在
2. 确保 swarm 正在运行
3. 重启 monitor

## 📚 相关文档

- [TUI Monitor 文档](docs/TUI_MONITOR.md)
- [主 README](README.md)
- [架构设计](docs/ARCHITECTURE.md)

## 🎉 开始使用

准备好了吗？运行以下命令开始：

```bash
# 终端 1：启动蜂群
./swarm start --agents 3

# 终端 2：监控进度
./swarm monitor

# 终端 3：查看原始输出（可选）
tmux attach -t claude-swarm
```

享受蜂群并行开发的极速体验！🚀
