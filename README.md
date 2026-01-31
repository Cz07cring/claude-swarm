# Claude Swarm 🐝

<div align="center">

**AI-Powered Multi-Agent Development System**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-v2.0-blue.svg)](https://github.com/Cz07cring/claude-swarm)

[English](README.md) • [简体中文](README_ZH.md)

</div>

---

## What is Claude Swarm?

An **AI-driven multi-agent system** that orchestrates multiple Claude Code instances for parallel development. One command, multiple agents, blazing fast results.

```bash
# Start 5 agents
./swarm start-v2 --agents 5

# Each task completes in 10-12 seconds
# Fully automated, zero conflicts
```

---

## ✨ Key Features

### 🚀 Direct CLI Execution
- **Reliable**: Full control over Claude execution
- **Fast**: 10-12 seconds per task
- **Free**: No API costs

### 🧠 AI Risk Assessment
- Pre-execution safety checks
- Auto-blocks dangerous operations
- Production-safe

### 🔄 Smart Retry
- Auto-detects retryable errors
- Configurable retry limits
- 80% first-retry success rate

### 🌳 Git Worktree Isolation
- Zero file conflicts
- Parallel development
- Clean merge workflow

---

## 🚀 Quick Start

### Prerequisites

```bash
# Required
Go 1.21+          # Build and run
Claude Code       # Task execution
Git 2.25+         # Worktree support

# Optional
Gemini API Key    # For AI task generation
```

### Installation

```bash
# Clone and build
git clone https://github.com/Cz07cring/claude-swarm.git
cd claude-swarm
go build -o swarm ./cmd/swarm
```

### Run Your First Task

```bash
# 1. Create task
cat > ~/.claude-swarm/tasks.json << 'EOF'
{
  "tasks": [{
    "id": "task-1",
    "description": "Create hello.go with main function",
    "status": "pending",
    "priority": 5,
    "max_retries": 3
  }]
}
EOF

# 2. Start swarm
./swarm start-v2 --agents 3

# 3. Watch it work
# Task completes in ~11 seconds
```

---

## 📋 Commands

```bash
# Start agents
swarm start-v2 --agents N

# Add task
swarm add-task "your task description"

# Monitor (TUI)
swarm monitor

# Check status
swarm status

# Stop
swarm stop
```

### With AI Orchestrator

```bash
# AI generates task queue from description
swarm orchestrate "Build a REST API with user CRUD"

# Then run
swarm start-v2 --agents 5
```

---

## 🏗️ Architecture

```
Task Queue (JSON)
    ↓
CoordinatorV2
    ├── Agent 0 (worktree-0) ⚡
    ├── Agent 1 (worktree-1) ⚡
    └── Agent N (worktree-n) ⚡
         ↓
ClaudeExecutor
  • echo | claude --dangerously-skip-permissions
  • AI risk assessment
  • Auto retry on failure
```

**Key Points:**
- Each agent in isolated git worktree
- Direct CLI execution (no tmux)
- AI safety layer before execution
- Auto-retry on network/temp errors

---

## 📊 Performance

| Metric | Value |
|--------|-------|
| Task Speed | 10-12s |
| Reliability | >95% |
| Memory/Agent | ~50MB |
| Retry Success | 80% |

**Speedup Example:**
- 10 tasks, 1 agent: 110s
- 10 tasks, 5 agents: 24s (4.6x faster)
- 10 tasks, 10 agents: 12s (9x faster)

---

## 📖 Usage Examples

### Simple Tasks

```bash
# Parallel execution
./swarm start-v2 --agents 3

# Tasks run simultaneously:
# Agent-0: Create README (11s)
# Agent-1: Write tests (12s)
# Agent-2: Add CI/CD (10s)
```

### With Dependencies

```json
{
  "tasks": [
    {
      "id": "t1",
      "description": "Create database schema",
      "status": "pending"
    },
    {
      "id": "t2",
      "description": "Implement API endpoints",
      "dependencies": ["t1"]
    }
  ]
}
```

### Production Deploy

```bash
# Task with retry
{
  "id": "deploy",
  "description": "Deploy to production",
  "max_retries": 5,
  "priority": 10
}

# Start with monitoring
./swarm start-v2 --agents 1 &
./swarm monitor
```

---

## 🎨 TUI Monitor

Real-time dashboard with:
- **Agent Grid**: Visual status (5x5 grid)
- **Task List**: Progress tracking
- **Log Viewer**: Real-time output

**Keyboard:**
- `Tab`: Switch panels
- `j/k`: Navigate
- `Enter`: View logs
- `q`: Quit

---

## 📚 Documentation

- [V2 Architecture](docs/V2_INTEGRATION_COMPLETE.md) - Technical details
- [User Guide](docs/guides/USER_GUIDE.md) - Complete tutorial
- [Test Reports](docs/reports/) - Validation results

---

## 🗺️ Roadmap

**Current (v2.0):**
- ✅ Direct CLI execution
- ✅ AI risk assessment
- ✅ Smart retry
- ✅ Worktree isolation

**Next (v2.1):**
- Enhanced DAG scheduling
- Auto git merge
- Web dashboard
- Prometheus metrics

---

## 💡 FAQ

**Q: How is this different from running Claude manually?**
A: Automates parallel execution, task management, error handling, and conflict prevention. 5-10x faster for multi-task projects.

**Q: Is it free?**
A: Yes. Uses free Claude CLI. No API costs.

**Q: What if an agent fails?**
A: Auto-retries on network/temp errors. Permanent failures marked and logged.

**Q: Can agents conflict?**
A: No. Each agent works in isolated git worktree.

---

## 🤝 Contributing

```bash
# Fork, clone, create branch
git checkout -b feature/amazing

# Make changes, test
go test ./...

# Submit PR
```

---

## 📄 License

MIT License - see [LICENSE](LICENSE)

---

<div align="center">

**⚡ v2.0** - Production-ready reliability meets blazing speed

**🚀 10-12s/task** • **🧠 AI-powered** • **💯 Free**

[GitHub](https://github.com/Cz07cring) • [Issues](https://github.com/Cz07cring/claude-swarm/issues)

</div>
