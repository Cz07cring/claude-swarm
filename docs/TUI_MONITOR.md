# TUI Monitor - 终端用户界面监控面板

## 概述

TUI Monitor 是 Claude Agent Swarm 的交互式终端监控面板，提供实时的可视化界面来监控 Agent 集群和任务队列状态。

## 功能特性

### 三窗格布局

1. **任务列表 (Tasks)** - 左侧面板
   - 显示所有任务及其状态
   - 状态指示器：
     - `○` 待处理 (Pending)
     - `●` 进行中 (In Progress)
     - `✓` 已完成 (Completed)
     - `✗` 失败 (Failed)
   - 显示任务分配的 Agent
   - 显示任务更新时间

2. **Agent 网格 (Agents)** - 中间面板
   - 3x3 网格布局，最多显示 9 个 Agent
   - 颜色编码状态：
     - 灰色边框：空闲 (Idle)
     - 蓝色边框：工作中 (Working)
     - 黄色边框：等待确认 (Waiting Confirm)
     - 红色边框：错误/卡住 (Error/Stuck)
   - 显示当前任务 ID

3. **日志查看器 (Logs)** - 右侧面板
   - 显示选中 Agent 的实时输出
   - 自动滚动显示最新日志
   - 显示 Agent 状态和当前任务描述

### 实时更新

- 每 2 秒自动刷新数据
- 无需手动刷新即可查看最新状态
- 与运行中的 swarm 实例同步

### 键盘导航

- `Tab` / `Shift+Tab` - 在面板间切换
- `↑` / `k` - 向上导航
- `↓` / `j` - 向下导航
- `←` / `h` - 向左导航 (仅 Agent 网格)
- `→` / `l` - 向右导航 (仅 Agent 网格)
- `Enter` - 选择当前项
- `q` / `Ctrl+C` - 退出

## 使用方法

### 启动 TUI Monitor

```bash
# 使用默认配置
./swarm monitor

# 指定任务队列和状态文件路径
./swarm monitor --queue ~/.claude-swarm/tasks.json --state ~/.claude-swarm/agents.json
```

### 前置条件

1. 确保 swarm 已启动：
   ```bash
   ./swarm start --agents 3
   ```

2. 在新终端窗口启动 monitor：
   ```bash
   ./swarm monitor
   ```

### 典型工作流

1. **启动 Swarm**
   ```bash
   # 终端 1: 启动 swarm
   ./swarm start --agents 5
   ```

2. **启动监控面板**
   ```bash
   # 终端 2: 启动 TUI monitor
   ./swarm monitor
   ```

3. **添加任务**
   ```bash
   # 终端 3: 添加任务
   ./swarm add-task "实现用户认证功能"
   ./swarm add-task "创建数据库架构"
   ```

4. **监控执行**
   - 在 TUI 中观察任务状态变化
   - 使用 `Tab` 切换到 Agent 面板查看各 Agent 状态
   - 选择 Agent 查看其日志输出

## 架构设计

### 组件结构

```
pkg/tui/
├── dashboard.go      # 主 Bubble Tea 模型和程序入口
├── tasklist.go       # 任务列表视图组件
├── agentgrid.go      # Agent 网格视图组件
├── logviewer.go      # 日志查看器组件
└── styles.go         # Lipgloss 样式定义
```

### 数据流

```
[Coordinator]
    ↓ (写入)
[~/.claude-swarm/agents.json] (Agent 状态文件)
[~/.claude-swarm/tasks.json]  (任务队列文件)
    ↓ (读取)
[TUI Monitor]
    ↓ (渲染)
[终端界面]
```

### 状态持久化

- **Agent 状态**: `pkg/state/agentstate.go`
  - Coordinator 每 2 秒写入 Agent 状态到 JSON 文件
  - Monitor 读取文件获取实时状态
  - 使用文件锁确保并发安全

- **任务队列**: `pkg/state/taskqueue.go`
  - 已有的任务队列机制
  - Monitor 直接读取任务状态

## 技术栈

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**: TUI 框架
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)**: 终端样式库
- **Go 1.25+**: 编程语言

## 性能特性

- **低资源占用**: TUI 仅消耗 ~10MB 内存
- **非阻塞**: 不影响 Coordinator 性能
- **多实例**: 支持多个 monitor 同时运行
- **无需网络**: 纯本地文件通信

## 故障排除

### 问题: TUI 显示空数据

**原因**: Swarm 未启动或状态文件不存在

**解决方案**:
```bash
# 检查 swarm 是否运行
./swarm status

# 如果未运行，先启动
./swarm start --agents 3
```

### 问题: Agent 状态不更新

**原因**: 状态文件路径不匹配

**解决方案**:
```bash
# 确保 monitor 使用相同的状态文件路径
./swarm monitor --state ~/.claude-swarm/agents.json
```

### 问题: 日志显示延迟

**原因**: 2 秒刷新间隔

**说明**: 这是设计决定，避免过度 I/O 操作。日志会在下次刷新时更新。

## 未来增强

### 第二阶段: Web 仪表板 (计划中)

- 基于浏览器的监控界面
- 远程访问支持
- 更丰富的可视化 (图表、DAG 图)
- 团队协作功能

### 第三阶段: Web 控制 (计划中)

- 从 Web 界面编排任务
- 交互式批准工作流
- Agent 控制 (启动/停止/暂停)

## 相关文档

- [主 README](../README.md)
- [架构设计](../docs/ARCHITECTURE.md)
- [可行性分析报告](../docs/UI_FEASIBILITY_REPORT.md)

## 反馈

如有问题或建议，请提交 GitHub Issue 或 Pull Request。
