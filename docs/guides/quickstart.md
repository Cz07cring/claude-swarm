# 快速开始指南

本指南将帮助你在 5 分钟内开始使用 Claude Agent Swarm。

## 前置要求检查

在开始之前，确保你已经安装了以下工具：

### 1. 检查 Go

```bash
go version
# 应该显示: go version go1.21 或更高版本
```

如果未安装：
```bash
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.21.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### 2. 检查 tmux

```bash
tmux -V
# 应该显示: tmux 3.x 或更高版本
```

如果未安装：
```bash
# macOS
brew install tmux

# Ubuntu/Debian
sudo apt install tmux

# CentOS/RHEL
sudo yum install tmux
```

### 3. 检查 Claude Code

```bash
claude --version
# 应该显示 Claude Code 版本
```

如果未安装，访问 [claude.ai/claude-code](https://claude.ai/claude-code)

## 安装 Claude Swarm

### 方式 1: 从源码构建

```bash
# 1. 克隆仓库
git clone https://github.com/yourusername/claude-swarm.git
cd claude-swarm

# 2. 安装依赖
go mod download

# 3. 构建
go build -o swarm ./cmd/swarm

# 4. （可选）移动到 PATH
sudo mv swarm /usr/local/bin/
```

### 方式 2: 直接运行

```bash
# 克隆后直接运行
cd claude-swarm
go run ./cmd/swarm start
```

## 第一次使用

### Step 1: 启动集群

```bash
# 启动 3 个 Agent
./swarm start
```

你应该看到：
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

**保持这个终端打开！** 协调器正在运行。

### Step 2: 添加第一个任务

打开**新的终端窗口**，运行：

```bash
./swarm add-task "列出当前目录的文件"
```

输出：
```
✓ 任务已添加
  ID: task-1738239876
  描述: 列出当前目录的文件
  状态: pending
```

### Step 3: 查看状态

```bash
./swarm status
```

你应该看到任务被分配给某个 Agent：

```
📊 Claude Agent Swarm 状态
============================================================

✓ 会话: claude-swarm (运行中)

  窗格数量: 3

📋 任务队列: 1 个任务

  状态统计:
    进行中: 1

  最近任务:
    🔄 task-1738239876 | 列出当前目录的文件
      状态: in_progress | Agent: agent-0 | 刚刚
```

### Step 4: 查看实时输出

```bash
tmux attach -t claude-swarm
```

你会看到 3 个窗格，每个运行一个 Claude 实例。

**tmux 快捷键：**
- `Ctrl+B` 然后按 `←` 或 `→` - 切换窗格
- `Ctrl+B` 然后按 `D` - 退出（不停止会话）
- `Ctrl+B` 然后按 `Z` - 放大/缩小当前窗格

### Step 5: 添加更多任务

回到另一个终端：

```bash
./swarm add-task "显示系统信息"
./swarm add-task "查看 Git 状态"
./swarm add-task "列出 Go 版本"
```

观察任务被自动分配给空闲的 Agent！

### Step 6: 停止集群

```bash
# 在协调器终端按 Ctrl+C
# 或在另一个终端运行：
./swarm stop
```

## 实际使用场景

### 场景 1: 并行开发功能

```bash
# 启动集群
./swarm start -n 3

# 添加多个功能任务
./swarm add-task "创建用户注册 API 端点"
./swarm add-task "创建用户登录 API 端点"
./swarm add-task "创建用户注销 API 端点"

# 监控进度
watch -n 2 ./swarm status
```

### 场景 2: 批量测试

```bash
# 启动更多 Agent
./swarm start -n 5

# 批量添加测试任务
for module in auth user payment order notification
do
  ./swarm add-task "为 $module 模块编写单元测试"
done
```

### 场景 3: 代码审查和重构

```bash
./swarm add-task "审查 user.go 文件并提出改进建议"
./swarm add-task "重构 database.go 使用依赖注入"
./swarm add-task "优化 query.go 的数据库查询"
```

## 高级配置

### 自定义监控间隔

```bash
# 每 3 秒检查一次状态（默认 5 秒）
./swarm start -i 3
```

### 自定义会话名称

```bash
# 使用自定义会话名称
./swarm start -s dev-swarm

# 查看状态时需要指定
./swarm status -s dev-swarm
```

### 自定义任务队列路径

```bash
# 使用项目目录的任务队列
./swarm start -q ./tasks.json
./swarm add-task -q ./tasks.json "任务描述"
```

## 自动确认功能

Claude Swarm 会自动检测 Claude 的等待确认状态并发送 "yes"。

**安全检查：** 如果检测到危险关键词（如 delete, remove, force），不会自动确认。

```bash
# 这个会自动确认（安全）
./swarm add-task "创建一个新文件"

# 这个不会自动确认（检测到 delete）
./swarm add-task "删除所有临时文件"
```

## 故障排除

### 问题: "command not found: tmux"

```bash
# 安装 tmux
brew install tmux  # macOS
sudo apt install tmux  # Linux
```

### 问题: "command not found: claude"

```bash
# 检查 Claude 是否安装
which claude

# 添加到 PATH（如果已安装）
export PATH=$PATH:~/.claude/bin
```

### 问题: 会话已存在

```bash
# 终止旧会话
tmux kill-session -t claude-swarm

# 重新启动
./swarm start
```

### 问题: Agent 没有响应

```bash
# 附加到 tmux 查看
tmux attach -t claude-swarm

# 手动在窗格中重启 claude
# 在窗格中按 Ctrl+C，然后输入: claude
```

## 下一步

- 阅读 [完整架构文档](../architecture/full-plan.md)
- 查看 [MVP 实施指南](./mvp-guide.md)
- 探索 [API 参考](../api/reference.md)

## 需要帮助？

- 查看 [GitHub Issues](https://github.com/yourusername/claude-swarm/issues)
- 阅读 [FAQ](./faq.md)
- 加入讨论

---

祝你使用愉快！🐝
