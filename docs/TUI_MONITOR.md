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

#### 面板切换
- `Tab` - 切换到下一个面板（Tasks → Agents → Logs → Tasks）
- `Shift+Tab` - 切换到上一个面板（逆序）

#### 导航控制
- `↑` / `k` - 向上移动光标
- `↓` / `j` - 向下移动光标
- `←` / `h` - 向左移动（仅在 Agent 网格面板有效）
- `→` / `l` - 向右移动（仅在 Agent 网格面板有效）

#### 操作
- `Enter` - 选择当前项并在日志面板显示详情
- `q` - 退出 TUI Monitor
- `Ctrl+C` - 强制退出

#### 面板特定行为

**Tasks 面板：**
- 上下键浏览任务列表
- 选中任务后，日志面板显示执行该任务的 Agent 输出

**Agents 面板：**
- 上下左右键在 3x3 网格中导航
- 选中 Agent 后，日志面板显示该 Agent 的实时输出

**Logs 面板：**
- 仅显示，不支持交互
- 自动滚动显示最新日志
- 内容来源于当前选中的 Agent 或任务

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

#### 工作流 1: 基础监控

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

#### 工作流 2: AI 主脑驱动开发监控

1. **使用 AI 主脑生成任务**
   ```bash
   # 终端 1: AI 分析并生成任务
   ./swarm orchestrate "实现实时聊天功能"
   ```

2. **启动集群和监控**
   ```bash
   # 终端 2: 启动监控（推荐在启动前开启）
   ./swarm monitor

   # 终端 3: 启动 Agent 集群
   ./swarm start --agents 8
   ```

3. **实时监控进度**
   - Tasks 面板：查看任务依赖和完成进度
   - Agents 面板：监控各 Agent 工作状态
   - Logs 面板：查看具体执行细节

#### 工作流 3: 调试和问题排查

1. **检测卡住的 Agent**
   - 在 Agents 面板查找红色边框的 Agent（错误/卡住）
   - 选中该 Agent 查看日志面板的错误信息

2. **监控等待确认的 Agent**
   - 查找黄色边框的 Agent（等待确认）
   - 切换到 tmux 会话手动处理确认
   ```bash
   tmux attach -t claude-swarm
   ```

3. **查看任务执行详情**
   - 在 Tasks 面板选中特定任务
   - 日志面板自动显示执行该任务的 Agent 输出

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

## 使用场景示例

### 场景 1: 大规模并行开发监控

**需求**: 同时监控 10 个 Agent 的开发进度

```bash
# 启动 10 个 Agent
./swarm start --agents 10

# 启动 TUI monitor
./swarm monitor
```

**监控重点**:
- 快速扫描 Agent 网格，识别工作状态
- 定位空闲 Agent（灰色边框）
- 检查是否有 Agent 需要人工介入（黄色/红色边框）

### 场景 2: 任务进度跟踪

**需求**: 跟踪复杂项目的任务完成情况

```bash
# AI 主脑生成任务
./swarm orchestrate "实现完整的用户认证系统"

# 监控执行
./swarm monitor
```

**监控重点**:
- Tasks 面板：查看待处理、进行中、已完成任务数量
- 识别瓶颈：哪些任务耗时较长
- 评估剩余工作量

### 场景 3: 实时调试

**需求**: 调试特定 Agent 的执行问题

**操作流程**:
1. 在 Agents 面板找到目标 Agent
2. 使用方向键选中该 Agent
3. 在 Logs 面板查看实时输出
4. 如需手动介入，切换到 tmux 会话

### 场景 4: 性能优化

**需求**: 评估 Agent 集群的负载均衡情况

**监控指标**:
- Agent 空闲率：灰色边框 Agent 占比
- 任务分配均衡度：各 Agent 完成的任务数
- 系统瓶颈：是否有 Agent 长时间处于等待状态

## 最佳实践

### 1. 监控布局建议

**推荐终端布局**:
```
+-------------------+-------------------+
|                   |                   |
|   TUI Monitor     |   tmux session    |
|   (全屏)          |   (需要时查看)    |
|                   |                   |
+-------------------+-------------------+
```

**最小终端尺寸**:
- 宽度：至少 120 列
- 高度：至少 30 行

### 2. 性能优化

- **刷新间隔**: 默认 2 秒，平衡实时性和 I/O 开销
- **多实例**: 可同时运行多个 monitor，互不干扰
- **资源占用**: TUI monitor 仅消耗约 10MB 内存

### 3. 常见模式

**持续集成场景**:
```bash
# 启动长时间运行的 swarm
./swarm start --agents 5

# 在另一个终端持续监控
./swarm monitor

# 可随时退出 monitor，swarm 继续运行
```

**快速验证场景**:
```bash
# 添加几个快速任务
./swarm add-task "运行单元测试"
./swarm add-task "检查代码风格"

# 临时查看状态
./swarm monitor
# 任务完成后 'q' 退出
```

### 4. 与其他工具配合

**与 watch 命令配合**:
```bash
# 在一个终端使用 TUI
./swarm monitor

# 在另一个终端定时查看统计
watch -n 5 './swarm status'
```

**与日志工具配合**:
```bash
# TUI 用于概览
./swarm monitor

# 同时记录详细日志
./swarm start --agents 5 2>&1 | tee swarm.log
```

## 故障排除

### 问题 1: TUI 显示空数据

**症状**: Monitor 启动但不显示任何任务或 Agent

**原因**:
- Swarm 未启动
- 状态文件不存在或路径不匹配

**解决方案**:
```bash
# 检查 swarm 是否运行
./swarm status

# 如果未运行，先启动
./swarm start --agents 3

# 检查状态文件是否存在
ls -l ~/.claude-swarm/agents.json
ls -l ~/.claude-swarm/tasks.json

# 确保 monitor 使用正确路径
./swarm monitor --queue ~/.claude-swarm/tasks.json --state ~/.claude-swarm/agents.json
```

### 问题 2: Agent 状态不更新

**症状**: Agent 状态在 TUI 中保持不变

**原因**:
- 状态文件路径不匹配
- Coordinator 未正常写入状态

**解决方案**:
```bash
# 检查 Coordinator 是否在运行
ps aux | grep swarm

# 重启 swarm
./swarm stop
./swarm start --agents 3

# 确认状态文件被更新
watch -n 1 'stat ~/.claude-swarm/agents.json'
```

### 问题 3: 日志显示延迟

**症状**: 日志更新不够实时

**原因**:
- 2 秒刷新间隔（设计决定）
- tmux capture-pane 的缓冲限制

**说明**:
这是预期行为，避免过度 I/O 操作。如需实时日志，使用 tmux attach：
```bash
tmux attach -t claude-swarm
```

### 问题 4: 终端显示异常

**症状**:
- 布局错乱
- 字符显示不正常
- 颜色渲染问题

**解决方案**:
```bash
# 确保终端支持 256 色
echo $TERM  # 应显示 xterm-256color 或类似

# 设置正确的 TERM 变量
export TERM=xterm-256color

# 调整终端窗口大小至推荐尺寸（120x30 以上）

# 重启 TUI
./swarm monitor
```

### 问题 5: 无法退出 TUI

**症状**: 按 'q' 或 Ctrl+C 无响应

**解决方案**:
```bash
# 强制终止（在新终端）
pkill -f "swarm monitor"

# 或使用 Ctrl+Z 挂起，然后 kill
# 在 TUI 终端按 Ctrl+Z
bg
kill %1
```

### 问题 6: 性能问题

**症状**: TUI 响应缓慢或卡顿

**原因**:
- Agent 数量过多（>9 个时网格显示不全）
- 日志输出过大
- 系统 I/O 压力

**解决方案**:
```bash
# 减少 Agent 数量（3x3 网格最多显示 9 个）
./swarm stop
./swarm start --agents 6

# 清理旧日志
rm -f ~/.claude-swarm/*.log

# 监控系统资源
htop
```

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
