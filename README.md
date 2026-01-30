# Claude Agent Swarm 🐝

> 基于 tmux 的 Claude Code 多 Agent 协作开发环境

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 简介

Claude Agent Swarm 是一个创新的多 Agent 协作系统，能够同时管理多个 Claude Code 实例，实现任务自动分发、状态监控和智能协助。

### 核心特性

- 🚀 **并行开发** - 同时运行多个 Claude Agent，提升开发效率
- 🔄 **自动调度** - 智能任务分发和负载均衡
- 🤖 **智能协助** - 自动检测并处理等待确认、错误等状态
- 📊 **实时监控** - 监控所有 Agent 的运行状态
- 🎯 **简单易用** - 一条命令启动，CLI 友好
- 🌍 **跨平台** - 支持 macOS 和 Linux（Go 编译）

### 工作原理

```
用户添加任务
    ↓
任务队列（JSON 文件）
    ↓
调度器分配给空闲 Agent
    ↓
Agent 在独立 tmux 窗格中执行
    ↓
监控器检测状态（每 5 秒）
    ↓
自动处理确认/错误
    ↓
任务完成，Agent 变为空闲
```

## 快速开始

### 前置要求

- **Go 1.21+** - [安装 Go](https://go.dev/doc/install)
- **tmux** - 终端复用器
- **Claude Code** - [安装 Claude CLI](https://claude.ai/claude-code)

```bash
# macOS 安装 tmux
brew install tmux

# Linux 安装 tmux
sudo apt install tmux  # Ubuntu/Debian
sudo yum install tmux  # CentOS/RHEL
```

### 安装

```bash
# 克隆仓库
git clone https://github.com/yourusername/claude-swarm.git
cd claude-swarm

# 构建
go build -o swarm ./cmd/swarm

# 或直接运行
go run ./cmd/swarm
```

### 使用

#### 1. 启动 Agent 集群

```bash
# 启动 3 个 Agent（默认）
./swarm start

# 启动指定数量的 Agent
./swarm start -n 5

# 自定义会话名称
./swarm start -s my-swarm
```

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

#### 2. 添加任务

```bash
# 添加任务
./swarm add-task "创建一个 HTTP 服务器"
./swarm add-task "编写单元测试"
./swarm add-task "优化数据库查询"
```

输出示例：
```
✓ 任务已添加
  ID: task-1738239876
  描述: 创建一个 HTTP 服务器
  状态: pending
```

#### 3. 查看状态

```bash
./swarm status
```

输出示例：
```
📊 Claude Agent Swarm 状态
============================================================

✓ 会话: claude-swarm (运行中)

  窗格数量: 3

📋 任务队列: 5 个任务

  状态统计:
    待处理: 2
    进行中: 1
    已完成: 2

  最近任务:
    ✅ task-1738239900 | 编写单元测试
      状态: completed | Agent: agent-1 | 5 分钟前
    🔄 task-1738239876 | 创建一个 HTTP 服务器
      状态: in_progress | Agent: agent-0 | 10 分钟前

============================================================

💡 提示:
  - 查看实时输出: tmux attach -t claude-swarm
  - 添加任务: swarm add-task "任务描述"
  - 停止集群: swarm stop
```

#### 4. 查看实时输出

```bash
# 附加到 tmux 会话
tmux attach -t claude-swarm

# 退出（但不停止会话）: Ctrl+B 然后按 D
```

#### 5. 停止集群

```bash
./swarm stop
```

## 命令参考

### `swarm start`

启动 Agent 集群

```bash
swarm start [flags]
```

**选项：**
- `-n, --agents int` - Agent 数量（默认: 3）
- `-s, --session string` - tmux 会话名称（默认: claude-swarm）
- `-q, --queue string` - 任务队列文件路径（默认: ~/.claude-swarm/tasks.json）
- `-i, --interval int` - 监控间隔秒数（默认: 5）

**示例：**
```bash
# 启动 5 个 Agent，监控间隔 3 秒
swarm start -n 5 -i 3

# 自定义会话名称和队列路径
swarm start -s dev-swarm -q /tmp/tasks.json
```

### `swarm add-task`

添加任务到队列

```bash
swarm add-task [description]
```

**示例：**
```bash
swarm add-task "实现用户登录功能"
swarm add-task "修复注册页面的 bug"
```

### `swarm status`

查看集群和任务状态

```bash
swarm status
```

### `swarm stop`

停止集群

```bash
swarm stop
```

## 工作流示例

### 示例 1: 并行开发多个功能

```bash
# 1. 启动 3 个 Agent
swarm start -n 3

# 2. 添加多个任务
swarm add-task "实现用户注册 API"
swarm add-task "实现用户登录 API"
swarm add-task "实现密码重置 API"

# 3. 查看状态
swarm status

# 4. 附加到 tmux 查看实时进度
tmux attach -t claude-swarm
```

### 示例 2: 批量处理重复任务

```bash
# 启动集群
swarm start -n 5

# 批量添加任务
for feature in login register profile settings
do
  swarm add-task "为 $feature 功能编写单元测试"
done

# 监控进度
watch -n 2 swarm status
```

## 配置

任务队列默认存储在 `~/.claude-swarm/tasks.json`，格式如下：

```json
{
  "tasks": [
    {
      "id": "task-1738239876",
      "description": "创建一个 HTTP 服务器",
      "status": "in_progress",
      "assignee_id": "agent-0",
      "created_at": "2026-01-30T10:00:00Z",
      "updated_at": "2026-01-30T10:05:00Z"
    }
  ]
}
```

## 架构

### 项目结构

```
claude-swarm/
├── cmd/swarm/           # CLI 命令
│   ├── main.go
│   ├── start.go
│   ├── stop.go
│   ├── add.go
│   └── status.go
├── pkg/
│   ├── tmux/           # tmux 会话和窗格管理
│   ├── state/          # 任务队列和状态管理
│   ├── analyzer/       # Claude 输出分析和状态检测
│   └── controller/     # 协调器（调度、监控、救援）
├── internal/models/    # 数据模型
├── docs/              # 文档
└── README.md
```

### 核心组件

1. **tmux Manager** - 管理 tmux 会话和窗格
   - 创建/销毁会话
   - 分割窗格
   - 捕获输出（`capture-pane`）
   - 发送命令（`send-keys`）

2. **Task Queue** - 任务队列管理
   - JSON 文件存储
   - FIFO 调度
   - 原子操作（避免并发冲突）

3. **Analyzer** - 状态检测
   - 正则模式匹配
   - 识别等待确认、错误、卡住等状态
   - 安全检查（判断是否可自动确认）

4. **Coordinator** - 协调器
   - 任务调度（分配给空闲 Agent）
   - 状态监控（goroutine 池）
   - 自动救援（处理确认、错误、卡住）

## MVP 范围

当前 MVP 版本包含：

✅ tmux 会话管理
✅ 基础感知和控制（capture-pane, send-keys）
✅ 简单任务队列（JSON 文件）
✅ 基础状态检测（等待确认、错误）
✅ 自动确认功能
✅ CLI 命令（start, stop, add-task, status）

暂不包含：

❌ Git worktree 管理
❌ SQLite 数据库
❌ 复杂调度算法
❌ P2P 救援机制
❌ TUI 仪表板

## 故障排除

### tmux 会话创建失败

```bash
# 检查 tmux 是否安装
which tmux

# 查看现有会话
tmux ls

# 手动终止旧会话
tmux kill-session -t claude-swarm
```

### Claude 未启动

```bash
# 检查 claude 是否在 PATH 中
which claude

# 手动附加到 tmux 并启动
tmux attach -t claude-swarm
# 在窗格中输入: claude
```

### 任务队列损坏

```bash
# 删除任务队列文件
rm ~/.claude-swarm/tasks.json

# 重新启动
swarm start
```

## 开发

### 构建

```bash
# 开发模式运行
go run ./cmd/swarm start

# 构建二进制
go build -o swarm ./cmd/swarm

# 跨平台构建
GOOS=linux GOARCH=amd64 go build -o swarm-linux ./cmd/swarm
GOOS=darwin GOARCH=arm64 go build -o swarm-darwin ./cmd/swarm
```

### 测试

```bash
# 运行测试
go test ./...

# 测试覆盖率
go test -cover ./...
```

## 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md)

## 路线图

- [ ] Phase 1: MVP ✅（当前）
- [ ] Phase 2: Git worktree 管理
- [ ] Phase 3: SQLite 数据库
- [ ] Phase 4: TUI 仪表板
- [ ] Phase 5: 复杂调度算法（优先级、依赖）
- [ ] Phase 6: P2P 救援机制
- [ ] Phase 7: Windows 支持
- [ ] Phase 8: Docker 镜像

## 许可证

MIT License - 详见 [LICENSE](LICENSE)

## 参考

- [AI蜂群协作-tmux多Agent协作系统](https://github.com/tukuaiai/vibe-coding-cn/blob/main/i18n/zh/documents/02-%E6%96%B9%E6%B3%95%E8%AE%BA/AI%E8%9C%82%E7%BE%A4%E5%8D%8F%E4%BD%9C-tmux%E5%A4%9AAgent%E5%8D%8F%E4%BD%9C%E7%B3%BB%E7%BB%9F.md)
- [tmux 文档](https://github.com/tmux/tmux/wiki)

## 联系

- GitHub: [@yourusername](https://github.com/yourusername)
- Issues: [GitHub Issues](https://github.com/yourusername/claude-swarm/issues)

---

**⚠️ 注意：** 这是一个实验性项目，请在生产环境使用前充分测试。
