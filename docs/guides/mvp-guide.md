# MVP-Go 实施指南

## MVP 目标

快速验证 Claude Agent 蜂群协作的核心概念，实现最小可行产品。

**时间：** 1-2 天
**范围：** 核心功能验证，不包括完整的 Git 管理和复杂调度

## MVP 功能范围

### 包含功能 ✅

1. **tmux 会话管理**
   - 创建 claude-swarm 会话
   - 分割窗格（3 个 agent）
   - 在每个窗格启动 claude

2. **基础感知和控制**
   - 使用 `capture-pane` 读取输出
   - 使用 `send-keys` 发送命令
   - 检测"等待确认"状态

3. **简单任务队列**
   - JSON 文件存储任务
   - 手动添加任务
   - Agent 按顺序领取任务

4. **基础状态检测**
   - 识别等待确认
   - 识别错误信息
   - 自动发送 "yes" 确认

5. **最小 CLI**
   - `swarm start` - 启动集群
   - `swarm stop` - 停止集群
   - `swarm add-task <description>` - 添加任务
   - `swarm status` - 查看状态

### 不包含功能 ❌

- ❌ Git worktree 管理（手动管理分支）
- ❌ SQLite 数据库（使用 JSON 文件）
- ❌ 复杂调度算法（FIFO 队列即可）
- ❌ P2P 救援机制（仅主控端救援）
- ❌ TUI 仪表板（命令行输出即可）
- ❌ 任务依赖管理
- ❌ 自动合并策略

## MVP 项目结构

```
claude-swarm/
├── cmd/
│   └── swarm/
│       ├── main.go          # CLI 入口
│       ├── start.go         # start 命令
│       ├── stop.go          # stop 命令
│       ├── add.go           # add-task 命令
│       └── status.go        # status 命令
├── pkg/
│   ├── tmux/
│   │   ├── session.go       # tmux 会话管理
│   │   ├── pane.go          # 窗格操作
│   │   └── types.go         # 类型定义
│   ├── state/
│   │   ├── taskqueue.go     # 任务队列（JSON）
│   │   └── status.go        # 状态管理
│   └── analyzer/
│       ├── detector.go      # 状态检测
│       └── patterns.go      # 正则模式
├── internal/
│   └── models/
│       └── task.go          # 任务模型
├── docs/                    # 文档
├── go.mod
├── go.sum
└── README.md
```

## 实施步骤

### Step 1: 项目初始化（15 分钟）

```bash
# 创建项目目录
mkdir -p claude-swarm/{cmd/swarm,pkg/{tmux,state,analyzer},internal/models,docs}
cd claude-swarm

# 初始化 Go 模块
go mod init github.com/yourusername/claude-swarm

# 安装依赖
go get github.com/spf13/cobra@latest
go get github.com/sirupsen/logrus@latest
```

### Step 2: 实现 tmux 模块（1-2 小时）

**文件：** `pkg/tmux/types.go`
```go
package tmux

type Session struct {
    Name  string
    Panes []*Pane
}

type Pane struct {
    ID      string
    Index   int
    AgentID string
}
```

**文件：** `pkg/tmux/session.go`
- 实现 `NewSession(name string) (*Session, error)`
- 实现 `SplitPane(horizontal bool) (*Pane, error)`
- 实现 `Kill() error`

**文件：** `pkg/tmux/pane.go`
- 实现 `Capture() (string, error)`
- 实现 `SendKeys(keys string) error`

**测试：**
```bash
# 手动测试
go run cmd/swarm/main.go test-tmux
# 应该创建 tmux 会话并显示输出
```

### Step 3: 实现任务队列（30 分钟）

**文件：** `internal/models/task.go`
```go
package models

import "time"

type Task struct {
    ID          string    `json:"id"`
    Description string    `json:"description"`
    Status      string    `json:"status"` // pending, in_progress, completed
    AssigneeID  string    `json:"assignee_id,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
}
```

**文件：** `pkg/state/taskqueue.go`
- JSON 文件读写
- 添加任务
- 领取任务（FIFO）
- 更新任务状态

### Step 4: 实现状态检测（1 小时）

**文件：** `pkg/analyzer/patterns.go`
```go
package analyzer

import "regexp"

var (
    PatternWaitingConfirm = regexp.MustCompile(`(?i)(waiting for confirmation|proceed with this plan\?)`)
    PatternError          = regexp.MustCompile(`(?i)(error:|failed to|cannot)`)
)
```

**文件：** `pkg/analyzer/detector.go`
- 分析输出识别状态
- 判断是否可安全确认

### Step 5: 实现主控端逻辑（2-3 小时）

创建 coordinator 包，实现：
- 启动 N 个 agent（tmux 窗格）
- 监控循环（goroutine）
- 任务分配
- 自动确认

### Step 6: 实现 CLI（1 小时）

**文件：** `cmd/swarm/main.go`
- 使用 cobra 创建 CLI 框架

**命令实现：**
- `start` - 启动集群
- `stop` - 停止集群
- `add-task` - 添加任务
- `status` - 显示状态

### Step 7: 测试验证（1-2 小时）

手动测试场景：
1. 启动 3 个 agent
2. 添加简单任务（如 "列出当前目录"）
3. 观察 agent 是否自动执行
4. 添加需要确认的任务
5. 验证自动确认功能

## MVP 数据流

```
1. 用户添加任务
   → swarm add-task "实现登录功能"
   → 写入 ~/.claude-swarm/tasks.json

2. 调度器检测到任务
   → 查找空闲 agent
   → 分配任务给 agent-1

3. 控制器发送任务
   → tmux send-keys -t pane-1 "实现登录功能\n"

4. 监控器检测输出
   → 每 5 秒 capture-pane
   → 发现 "WAITING FOR CONFIRMATION"

5. 自动确认
   → 检查是否安全
   → tmux send-keys -t pane-1 "yes\n"

6. 任务完成
   → 更新 tasks.json status = "completed"
   → agent-1 变为空闲
```

## 配置文件

**文件：** `~/.claude-swarm/config.json`
```json
{
  "session_name": "claude-swarm",
  "num_agents": 3,
  "monitor_interval": 5,
  "auto_confirm": true,
  "safe_keywords": ["yes", "proceed", "confirm"],
  "danger_keywords": ["delete", "remove", "force", "rm -rf"]
}
```

## 使用示例

### 启动集群

```bash
# 启动 3 个 agent
swarm start

# 输出：
# ✓ Created tmux session: claude-swarm
# ✓ Started agent-0 in pane 0
# ✓ Started agent-1 in pane 1
# ✓ Started agent-2 in pane 2
# ✓ Coordinator running...
#
# Attach to session: tmux attach -t claude-swarm
```

### 添加任务

```bash
# 添加任务
swarm add-task "创建一个简单的 HTTP 服务器"

# 输出：
# ✓ Task added: task-001
# Description: 创建一个简单的 HTTP 服务器
```

### 查看状态

```bash
swarm status

# 输出：
# Claude Agent Swarm Status
#
# Agents:
#   agent-0: WORKING   | Task: task-001 | 创建一个简单的 HTTP 服务器
#   agent-1: IDLE      | Task: -
#   agent-2: IDLE      | Task: -
#
# Tasks:
#   task-001: IN_PROGRESS | agent-0 | 创建一个简单的 HTTP 服务器
#
# Session: claude-swarm (active)
```

### 停止集群

```bash
swarm stop

# 输出：
# ✓ Stopped coordinator
# ✓ Killed tmux session: claude-swarm
```

## 验证清单

在完成 MVP 后，验证以下功能：

- [ ] 能够启动 tmux 会话和多个窗格
- [ ] 能够在每个窗格启动 claude
- [ ] 能够捕获窗格输出
- [ ] 能够向窗格发送命令
- [ ] 能够添加任务到队列
- [ ] 空闲 agent 能自动领取任务
- [ ] 能够检测"等待确认"状态
- [ ] 能够自动发送 "yes" 确认
- [ ] 能够查看当前状态
- [ ] 能够正确停止集群

## 已知限制（MVP 阶段）

1. **无 Git 管理** - 需要手动管理分支
2. **简单队列** - FIFO，无优先级和依赖
3. **无冲突检测** - 可能多个 agent 修改同一文件
4. **无持久化** - 重启后状态丢失（仅 JSON 文件）
5. **无 TUI** - 需要手动 attach tmux 查看
6. **有限的救援** - 仅支持自动确认，无错误重试

## 后续改进方向

完成 MVP 验证后，可以按以下顺序改进：

1. **Phase 3: Git 管理** - 增加 worktree 支持
2. **Phase 5: 智能调度** - 优先级、依赖、负载均衡
3. **Phase 6: TUI** - 实时仪表板
4. **Phase 7: 跨平台** - Windows 支持、Docker 镜像

## 故障排除

### 问题：tmux 会话创建失败
```bash
# 检查 tmux 是否安装
which tmux

# 检查是否有同名会话
tmux ls
tmux kill-session -t claude-swarm
```

### 问题：无法捕获窗格输出
```bash
# 手动测试
tmux capture-pane -p -t claude-swarm:0.0

# 检查窗格 ID
tmux list-panes -t claude-swarm -F "#{pane_id} #{pane_index}"
```

### 问题：claude 未启动
```bash
# 检查 claude 是否在 PATH 中
which claude

# 手动在 tmux 窗格中启动
tmux send-keys -t claude-swarm:0.0 "claude" Enter
```

## 资源

- [tmux 手册](https://man.openbsd.org/tmux.1)
- [Cobra CLI 文档](https://github.com/spf13/cobra)
- [Go 正则表达式](https://pkg.go.dev/regexp)

---

**文档版本：** v1.0
**更新日期：** 2026-01-30
**状态：** 实施阶段
