# Claude Agent 蜂群协作开发环境 - 完整实施计划

## 项目概述

基于 tmux 多 Agent 协作系统思想，使用 Golang 构建跨平台的 Claude Code 集群开发环境。

**参考方案：** [AI蜂群协作-tmux多Agent协作系统](https://github.com/tukuaiai/vibe-coding-cn/blob/main/i18n/zh/documents/02-%E6%96%B9%E6%B3%95%E8%AE%BA/AI%E8%9C%82%E7%BE%A4%E5%8D%8F%E4%BD%9C-tmux%E5%A4%9AAgent%E5%8D%8F%E4%BD%9C%E7%B3%BB%E7%BB%9F.md)

## 核心功能

1. **终端复用管理** - 使用 tmux 管理多个 Claude Code 实例
2. **感知能力** - 读取每个窗格(pane)的 Claude 会话输出
3. **控制能力** - 向窗格发送命令（如确认、输入）
4. **协调能力** - 基于共享状态文件实现任务同步
5. **Git 分支管理** - 每个 agent 独立分支，最终合并
6. **主控端智能** - 任务分发、状态监控、自动救援

## 技术架构

```
┌─────────────────────────────────────────────────────┐
│            主控端 (Master Coordinator)               │
│  - goroutine 池管理多个 agent                        │
│  - channel 通信（任务分发、状态收集）                 │
│  - tmux 会话编排                                     │
│  - Git 分支协调                                      │
└─────────────────────────────────────────────────────┘
                      │
        ┌─────────────┴─────────────┐
        │   tmux Session: claude-swarm   │
        │                                │
        │  ┌──────────┬──────────┬──────────┐
        │  │ Pane 0   │ Pane 1   │ Pane 2   │
        │  │ Agent-1  │ Agent-2  │ Agent-3  │
        │  │ branch-1 │ branch-2 │ branch-3 │
        │  │ claude   │ claude   │ claude   │
        │  └──────────┴──────────┴──────────┘
        └────────────────────────────────────┘
                      │
        ┌─────────────┴─────────────┐
        │   共享状态 (内存 + 文件)    │
        │  ~/.claude-swarm/          │
        │  - tasks.db (SQLite)       │
        │  - status.json (状态)      │
        │  - logs/ (日志)            │
        └────────────────────────────┘
```

## 实施阶段

### Phase 1: tmux 终端复用基础

**目标：** 实现 tmux 会话创建、管理和基础通信

**功能：**
- tmux 会话管理（创建、销毁）
- 窗格控制（分割、布局）
- 感知能力（capture-pane）
- 控制能力（send-keys）

**目录结构：**
```
pkg/tmux/
├── session.go          # tmux 会话管理
├── pane.go            # 窗格控制
├── capture.go         # 输出捕获
├── sender.go          # 命令发送
├── layout.go          # 窗格布局
└── types.go           # 类型定义
```

**核心 API：**
```go
// 创建会话
session, err := tmux.NewSession("claude-swarm")

// 分割窗格
pane, err := session.SplitPane(true) // horizontal

// 捕获输出
output, err := pane.Capture()

// 发送命令
err := pane.SendKeys("yes\n")
```

**依赖：**
- Go 标准库 `os/exec`
- 系统依赖：tmux

### Phase 2: 共享状态与协作机制

**目标：** 实现任务队列、状态同步、并发安全

**功能：**
- SQLite 任务队列（支持事务）
- sync.Map 运行时状态缓存
- 文件锁（跨进程同步）
- 结构化日志系统

**目录结构：**
```
pkg/state/
├── taskqueue.go       # 任务队列（SQLite）
├── status.go          # 状态管理（sync.Map）
├── lock.go           # 文件锁
└── logger.go         # 日志系统

internal/models/
├── task.go           # 任务模型
├── agent.go          # Agent 模型
└── status.go         # 状态模型
```

**数据模型：**
```go
type Task struct {
    ID          string
    Description string
    Branch      string
    Status      string // pending, in_progress, completed, failed
    AssigneeID  string
    Dependencies []string
    CreatedAt   time.Time
}

type AgentStatus struct {
    AgentID     string
    State       string // idle, working, waiting_confirm, error, stuck
    CurrentTask *Task
    LastUpdate  time.Time
}
```

**核心 API：**
```go
// 创建任务队列
tq, err := state.NewTaskQueue("~/.claude-swarm/tasks.db")

// 添加任务
err := tq.AddTask(task)

// 领取任务（原子操作）
task, err := tq.ClaimTask(agentID)

// 更新任务状态
err := tq.UpdateTaskStatus(taskID, "completed")
```

**依赖：**
- `github.com/mattn/go-sqlite3` - SQLite 驱动
- `github.com/sirupsen/logrus` - 日志库
- `github.com/gofrs/flock` - 文件锁

### Phase 3: Git 分支管理与协调

**目标：** Git worktree 管理、自动合并

**功能：**
- Git 仓库操作
- Worktree 生命周期管理
- 分支创建和切换
- 自动合并策略
- 冲突检测

**目录结构：**
```
pkg/git/
├── repository.go      # 仓库操作
├── worktree.go       # worktree 管理
├── branch.go         # 分支操作
├── merge.go          # 合并策略
└── conflict.go       # 冲突检测
```

**工作区结构：**
```
/project-root/
├── .git/                    # 主仓库
├── worktrees/
│   ├── agent-1/            # Agent 1 工作目录
│   │   └── .git -> ../../.git/worktrees/agent-1
│   ├── agent-2/
│   └── agent-3/
```

**核心 API：**
```go
// 打开仓库
repo, err := git.OpenRepository("/path/to/repo")

// 创建 worktree
worktreePath, err := repo.CreateWorktree("agent-1", "feature/task-1")

// 合并分支
err := repo.MergeBranch("feature/task-1")

// 检测冲突
hasConflict, files := repo.DetectConflict("feature/task-1")
```

**依赖：**
- `github.com/go-git/go-git/v5` - Git 操作库

### Phase 4: Claude 会话状态分析

**目标：** 智能识别 Claude Code 的状态和问题

**功能：**
- 输出解析（正则模式匹配）
- 状态检测（状态机）
- 上下文管理（保持最近 N 行）
- 安全检查（判断是否可自动确认）

**目录结构：**
```
pkg/analyzer/
├── parser.go         # 输出解析
├── detector.go       # 状态检测
├── patterns.go       # 正则模式
└── context.go        # 上下文管理
```

**状态定义：**
```go
type State string

const (
    StateIdle             State = "idle"
    StateWorking          State = "working"
    StateWaitingConfirm   State = "waiting_confirm"
    StateError            State = "error"
    StateStuck            State = "stuck"
)
```

**检测模式：**
```go
var patterns = map[State]*regexp.Regexp{
    StateWaitingConfirm: regexp.MustCompile(`(?i)(waiting for confirmation|proceed with this plan\?)`),
    StateError:          regexp.MustCompile(`(?i)(error:|failed to|cannot)`),
}
```

**核心 API：**
```go
// 创建分析器
analyzer := analyzer.NewAnalyzer()

// 分析输出
state := analyzer.Analyze(output)

// 检查是否可安全确认
safe := analyzer.SafeToConfirm()
```

### Phase 5: 主控端智能调度

**目标：** 任务分发、监控、救援

**功能：**
- 任务拆分引擎
- 调度算法（负载均衡）
- 监控循环（goroutine 池）
- 自动救援机制
- 依赖管理（DAG）

**目录结构：**
```
pkg/controller/
├── coordinator.go    # 总协调器
├── scheduler.go      # 调度器
├── monitor.go        # 监控器（goroutine 池）
├── rescue.go         # 救援引擎
└── splitter.go       # 任务拆分
```

**协调流程：**
```
1. 启动 N 个 agent（tmux 窗格 + goroutine 监控）
2. 监控循环：每 5 秒捕获输出并分析状态
3. 调度循环：为空闲 agent 分配任务
4. 救援循环：检测卡住/错误的 agent 并处理
```

**核心 API：**
```go
// 创建协调器
coordinator, err := controller.NewCoordinator(3) // 3 个 agent

// 启动
coordinator.Start()

// 添加任务
coordinator.AddTask(task)

// 停止
coordinator.Stop()
```

**救援策略：**
```go
type RescueStrategy struct {
    Detect func(state AgentStatus) bool
    Action func(agent *Agent) error
}

// 示例：等待确认自动发送 yes
{
    Detect: func(state AgentStatus) bool {
        return state.State == StateWaitingConfirm && SafeToConfirm()
    },
    Action: func(agent *Agent) error {
        return agent.Pane.SendKeys("yes\n")
    },
}
```

### Phase 6: CLI 与 TUI 界面

**目标：** 命令行工具和终端仪表板

**功能：**
- CLI 命令（start, stop, status, add-task, logs）
- TUI 实时仪表板（显示所有 agent 状态）
- 日志查看和过滤

**目录结构：**
```
cmd/swarm/
├── main.go           # CLI 入口
├── start.go          # start 命令
├── stop.go           # stop 命令
├── status.go         # status 命令
└── add.go            # add-task 命令

pkg/tui/
├── app.go           # Bubble Tea 应用
├── models.go        # TUI 模型
└── views.go         # 视图组件
```

**CLI 命令：**
```bash
# 启动集群（3 个 agent）
swarm start -n 3

# 启动并显示 TUI
swarm start -n 3 --tui

# 查看状态
swarm status

# 添加任务
swarm add-task "实现用户登录功能"

# 查看日志
swarm logs agent-1

# 停止集群
swarm stop
```

**依赖：**
- `github.com/spf13/cobra` - CLI 框架
- `github.com/charmbracelet/bubbletea` - TUI 框架
- `github.com/charmbracelet/lipgloss` - TUI 样式

### Phase 7: 跨平台支持与打包

**目标：** 跨平台兼容和发布

**功能：**
- 处理平台差异（Windows/macOS/Linux）
- 构建脚本和发布流程
- Docker 镜像
- 安装脚本

**文件：**
```
scripts/
├── build.sh          # 构建脚本
└── install.sh        # 安装脚本

.goreleaser.yaml      # GoReleaser 配置
Dockerfile            # Docker 镜像
```

**构建命令：**
```bash
# 本地构建
go build -o swarm ./cmd/swarm

# 跨平台构建
GOOS=linux GOARCH=amd64 go build -o swarm-linux ./cmd/swarm
GOOS=darwin GOARCH=arm64 go build -o swarm-darwin ./cmd/swarm

# 使用 goreleaser 发布
goreleaser release --snapshot --clean
```

## 最终项目结构

```
claude-swarm/
├── cmd/
│   └── swarm/              # CLI 入口
│       ├── main.go
│       ├── start.go
│       ├── stop.go
│       ├── status.go
│       └── add.go
├── pkg/
│   ├── tmux/              # Phase 1: tmux 控制
│   ├── state/             # Phase 2: 状态管理
│   ├── git/               # Phase 3: Git 操作
│   ├── analyzer/          # Phase 4: 会话分析
│   ├── controller/        # Phase 5: 主控端
│   └── tui/               # Phase 6: TUI 界面
├── internal/
│   └── models/            # 数据模型
├── config/
│   └── config.yaml        # 配置文件
├── docs/                  # 文档
│   ├── architecture/
│   ├── guides/
│   └── api/
├── scripts/
│   ├── build.sh
│   └── install.sh
├── .goreleaser.yaml
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## 技术栈

**核心依赖：**
```go
require (
    github.com/spf13/cobra v1.8.0              // CLI 框架
    github.com/charmbracelet/bubbletea v0.25.0 // TUI 框架
    github.com/charmbracelet/lipgloss v0.9.1   // TUI 样式
    github.com/mattn/go-sqlite3 v1.14.18       // SQLite
    github.com/go-git/go-git/v5 v5.11.0       // Git 操作
    github.com/sirupsen/logrus v1.9.3          // 日志
    github.com/spf13/viper v1.18.2             // 配置管理
    github.com/gofrs/flock v0.8.1              // 文件锁
)
```

**系统依赖：**
- Go 1.21+
- tmux (Linux/macOS)
- Git

**开发工具：**
- `goreleaser` - 自动发布
- `golangci-lint` - 代码检查

## 开发路线图

### MVP 阶段（1-2 天）
- ✅ Phase 1: tmux 基础操作
- ✅ Phase 2: 简单任务队列
- ✅ Phase 4: 基础状态检测
- ✅ 最小 CLI（start/stop）

### 完善阶段（3-5 天）
- Phase 3: Git worktree 管理
- Phase 5: 智能调度和救援
- Phase 6: 完整 CLI 和 TUI

### 优化阶段（2-3 天）
- Phase 7: 跨平台支持
- 性能优化
- 文档完善
- 测试覆盖

## 风险评估

**高风险：**
- ❗ **tmux 依赖** - 必须在支持 tmux 的环境（Linux/macOS）
  - *缓解：* 文档说明系统要求，Windows 用户使用 WSL/Docker

- ❗ **Claude Code 多实例** - 可能有 license 或资源限制
  - *缓解：* 先验证多实例可行性

- ❗ **自动确认风险** - 误判可能导致错误操作
  - *缓解：* 实现安全检查，危险操作需人工确认

**中等风险：**
- ⚠️ **状态检测准确性** - 误判 agent 状态
  - *缓解：* 设置更长超时，记录误判优化规则

- ⚠️ **Git 合并冲突** - 多分支同时修改
  - *缓解：* 冲突避免算法，人工介入复杂冲突

**低风险：**
- ℹ️ **学习曲线** - 用户需要理解基本概念
  - *缓解：* 提供详细文档和示例

## 协作规则（基于参考方案）

### 1. 先查后做
Agent 领取任务前检查其他 agent 状态，避免重复工作

### 2. 避免冲突
分析任务的文件依赖，避免分配修改相同文件的任务

### 3. 主动救援
每个 agent 定期检查其他 agent，发现卡住主动帮助

### 4. 状态广播
完成任务后写入共享日志，其他 agent 可感知进度

## 开源发布清单

- [ ] README.md（中英文）
- [ ] LICENSE（MIT）
- [ ] CONTRIBUTING.md
- [ ] 示例配置文件
- [ ] 使用文档
- [ ] 架构文档
- [ ] API 文档
- [ ] GitHub Actions CI/CD
- [ ] Docker 镜像
- [ ] Homebrew formula (macOS)
- [ ] 演示视频/GIF

## 参考资料

- [AI蜂群协作-tmux多Agent协作系统](https://github.com/tukuaiai/vibe-coding-cn/blob/main/i18n/zh/documents/02-%E6%96%B9%E6%B3%95%E8%AE%BA/AI%E8%9C%82%E7%BE%A4%E5%8D%8F%E4%BD%9C-tmux%E5%A4%9AAgent%E5%8D%8F%E4%BD%9C%E7%B3%BB%E7%BB%9F.md)
- [tmux 文档](https://github.com/tmux/tmux/wiki)
- [go-git 文档](https://github.com/go-git/go-git)
- [Bubble Tea 文档](https://github.com/charmbracelet/bubbletea)

---

**文档版本：** v1.0
**更新日期：** 2026-01-30
**状态：** 规划阶段
