# Claude Swarm 🐝

<div align="center">

**AI 驱动的多 Agent 开发系统**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-v2.0-blue.svg)](https://github.com/Cz07cring/claude-swarm)

[English](README.md) • [简体中文](README_ZH.md)

</div>

---

## 什么是 Claude Swarm？

**AI 驱动的多 Agent 系统**，编排多个 Claude Code 实例并行开发。一条命令，多个 Agent，极速完成。

```bash
# 启动 5 个 agents
./swarm start-v2 --agents 5

# 每个任务 10-12 秒完成
# 全自动，零冲突
```

---

## ✨ 核心特性

### 🚀 直接 CLI 执行
- **可靠**：完全控制 Claude 执行
- **快速**：每任务 10-12 秒
- **免费**：无 API 成本

### 🧠 AI 风险评估
- 执行前安全检查
- 自动阻止危险操作
- 生产环境安全

### 🔄 智能重试
- 自动检测可重试错误
- 可配置重试次数
- 首次重试成功率 80%

### 🌳 Git Worktree 隔离
- 零文件冲突
- 并行开发
- 干净的合并工作流

---

## 🚀 快速开始

### 前置要求

```bash
# 必需
Go 1.21+          # 构建运行
Claude Code       # 任务执行
Git 2.25+         # Worktree 支持

# 可选
Gemini API Key    # AI 任务生成
```

### 安装

```bash
# 克隆并构建
git clone https://github.com/Cz07cring/claude-swarm.git
cd claude-swarm
go build -o swarm ./cmd/swarm
```

### 运行第一个任务

```bash
# 1. 创建任务
cat > ~/.claude-swarm/tasks.json << 'EOF'
{
  "tasks": [{
    "id": "task-1",
    "description": "创建 hello.go，包含 main 函数",
    "status": "pending",
    "priority": 5,
    "max_retries": 3
  }]
}
EOF

# 2. 启动集群
./swarm start-v2 --agents 3

# 3. 观察执行
# 任务约 11 秒完成
```

---

## 📋 命令

```bash
# 启动 agents
swarm start-v2 --agents N

# 添加任务
swarm add-task "任务描述"

# 监控（TUI）
swarm monitor

# 查看状态
swarm status

# 停止
swarm stop
```

### 配合 AI 主脑

```bash
# AI 从描述生成任务队列
swarm orchestrate "构建带用户 CRUD 的 REST API"

# 然后运行
swarm start-v2 --agents 5
```

---

## 🏗️ 架构

```
任务队列 (JSON)
    ↓
CoordinatorV2
    ├── Agent 0 (worktree-0) ⚡
    ├── Agent 1 (worktree-1) ⚡
    └── Agent N (worktree-n) ⚡
         ↓
ClaudeExecutor
  • echo | claude --dangerously-skip-permissions
  • AI 风险评估
  • 失败自动重试
```

**关键点：**
- 每个 agent 在独立的 git worktree
- 直接 CLI 执行（无 tmux）
- 执行前 AI 安全层
- 网络/临时错误自动重试

---

## 📊 性能

| 指标 | 数值 |
|------|------|
| 任务速度 | 10-12秒 |
| 可靠性 | >95% |
| 内存/Agent | ~50MB |
| 重试成功率 | 80% |

**加速示例：**
- 10 任务，1 agent：110秒
- 10 任务，5 agents：24秒（4.6倍快）
- 10 任务，10 agents：12秒（9倍快）

---

## 📖 使用示例

### 简单任务

```bash
# 并行执行
./swarm start-v2 --agents 3

# 任务同时运行：
# Agent-0: 创建 README (11s)
# Agent-1: 编写测试 (12s)
# Agent-2: 添加 CI/CD (10s)
```

### 带依赖关系

```json
{
  "tasks": [
    {
      "id": "t1",
      "description": "创建数据库结构",
      "status": "pending"
    },
    {
      "id": "t2",
      "description": "实现 API 端点",
      "dependencies": ["t1"]
    }
  ]
}
```

### 生产部署

```bash
# 带重试的任务
{
  "id": "deploy",
  "description": "部署到生产环境",
  "max_retries": 5,
  "priority": 10
}

# 启动并监控
./swarm start-v2 --agents 1 &
./swarm monitor
```

---

## 🎨 TUI 监控

实时仪表板包含：
- **Agent 网格**：可视化状态（5x5 网格）
- **任务列表**：进度跟踪
- **日志查看器**：实时输出

**快捷键：**
- `Tab`：切换面板
- `j/k`：导航
- `Enter`：查看日志
- `q`：退出

---

## 📚 文档

- [V2 架构](docs/V2_INTEGRATION_COMPLETE.md) - 技术细节
- [用户指南](docs/guides/USER_GUIDE.md) - 完整教程
- [测试报告](docs/reports/) - 验证结果

---

## 🗺️ 路线图

**当前 (v2.0)：**
- ✅ 直接 CLI 执行
- ✅ AI 风险评估
- ✅ 智能重试
- ✅ Worktree 隔离

**下一步 (v2.1)：**
- 增强 DAG 调度
- 自动 git 合并
- Web 仪表板
- Prometheus 指标

---

## 💡 常见问题

**Q: 与手动运行 Claude 有何不同？**
A: 自动化并行执行、任务管理、错误处理和冲突预防。多任务项目快 5-10 倍。

**Q: 是否免费？**
A: 是的。使用免费的 Claude CLI。无 API 成本。

**Q: Agent 失败怎么办？**
A: 网络/临时错误自动重试。永久失败会标记和记录。

**Q: Agents 会冲突吗？**
A: 不会。每个 agent 在独立的 git worktree 中工作。

---

## 🤝 贡献

```bash
# Fork、克隆、创建分支
git checkout -b feature/amazing

# 修改、测试
go test ./...

# 提交 PR
```

---

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE)

---

<div align="center">

**⚡ v2.0** - 生产级可靠性遇上极速性能

**🚀 10-12秒/任务** • **🧠 AI 驱动** • **💯 免费**

[GitHub](https://github.com/Cz07cring) • [Issues](https://github.com/Cz07cring/claude-swarm/issues)

</div>
