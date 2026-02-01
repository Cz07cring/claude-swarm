# Claude Swarm 🐝

<div align="center">

**AI-Powered Multi-Agent Development System**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-2.0-blue.svg)](https://github.com/Cz07cring/claude-swarm)

[English](README.md) • [简体中文](README_ZH.md)

</div>

---

## What is Claude Swarm?

An **AI-driven multi-agent system** that orchestrates multiple Claude Code instances for parallel development.

```bash
# Simplest way - just run!
swarm run "Create a REST API with user CRUD"

# Or use multiple agents
swarm start --agents 5
```

**Why Claude Swarm?**
- 🚀 **5-10x faster** - Parallel task execution
- 🆓 **100% free** - Uses Claude CLI, no API costs
- 🔀 **Zero conflicts** - Git worktree isolation
- 🧠 **AI-powered** - Smart task orchestration

---

## 🚀 Quick Start

### Installation

```bash
# Option 1: Go install (recommended)
go install github.com/Cz07cring/claude-swarm/cmd/swarm@latest

# Option 2: Build from source
git clone https://github.com/Cz07cring/claude-swarm.git
cd claude-swarm && make install

# Verify installation
swarm doctor
```

### Three Ways to Use

#### 1. Quick Run (Simplest)
```bash
# No setup required - just run!
swarm run "Add error handling to main.go"
swarm run "Write unit tests for user.go"

# Pipe input
echo "Fix the authentication bug" | swarm run
```

#### 2. Multi-Agent Mode
```bash
# Initialize project
cd your-project
swarm init

# Add tasks
swarm add-task "Create user model"
swarm add-task "Add authentication" --priority 8
swarm add-task "Write tests"

# Run with 5 parallel agents
swarm start --agents 5

# Monitor progress
swarm status
```

#### 3. AI-Powered Mode
```bash
# Let AI break down your requirements
export GEMINI_API_KEY=your-key
swarm orchestrate "Build a blog system with posts and comments"

# Run with AI monitoring
swarm start --agents 5 --with-brain
```

---

## 📋 Commands

| Command | Description |
|---------|-------------|
| `swarm run "task"` | **Quick run** - Execute single task instantly |
| `swarm init` | Initialize project configuration |
| `swarm add-task "desc"` | Add task to queue |
| `swarm start --agents N` | Start N parallel agents |
| `swarm status` | View task queue status |
| `swarm monitor` | Real-time TUI dashboard |
| `swarm orchestrate "req"` | AI generates tasks from requirement |
| `swarm doctor` | Check system environment |
| `swarm clean` | Clean task queue |

---

## 🏗️ Architecture

```
                    ┌─────────────────┐
                    │   swarm start   │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   Coordinator   │
                    │  + AI Brain     │
                    └────────┬────────┘
           ┌─────────────────┼─────────────────┐
           │                 │                 │
    ┌──────▼──────┐   ┌──────▼──────┐   ┌──────▼──────┐
    │   Agent 0   │   │   Agent 1   │   │   Agent N   │
    │ (worktree)  │   │ (worktree)  │   │ (worktree)  │
    └──────┬──────┘   └──────┬──────┘   └──────┬──────┘
           │                 │                 │
           └─────────────────┼─────────────────┘
                             │
                    ┌────────▼────────┐
                    │  Auto Merge     │
                    │  to main        │
                    └─────────────────┘
```

**Key Features:**
- Each agent in isolated git worktree
- Automatic merge to main (Fast-forward / Three-way)
- Conflict detection with auto-abort
- AI-powered task monitoring (optional)

---

## 📊 Performance

| Metric | Result |
|--------|--------|
| Task Speed | **~10s** per task |
| Reliability | **100%** (60/60 tests) |
| Speedup | **4-5x** with 5 agents |
| Git Merge | Fast-forward + Three-way |

**Benchmark:**
- 5 tasks, 3 agents: 22s (vs 55s single = **2.5x faster**)
- 20 tasks, 5 agents: 53s (vs 220s single = **4.1x faster**)

---

## ✨ Key Features

| Feature | Description |
|---------|-------------|
| 🚀 **Quick Run** | `swarm run "task"` - No setup needed |
| 🔀 **Auto Merge** | Automatic git merge to main branch |
| 🧠 **AI Brain** | Gemini-powered task orchestration |
| 🔄 **Smart Retry** | Auto-retry on transient failures |
| 📊 **TUI Monitor** | Real-time progress dashboard |
| 🌳 **Worktree Isolation** | Zero file conflicts |

---

## 📚 Documentation

- [Architecture](docs/ARCHITECTURE.md) - System design
- [User Guide](docs/USAGE_GUIDE.md) - Complete tutorial
- [CLI Commands](docs/CLI_COMMANDS.md) - Command reference

---

## 🗺️ Roadmap

**V2.0 (Current):**
- ✅ Quick run command (`swarm run`)
- ✅ Project initialization (`swarm init`)
- ✅ Multi-agent parallel execution
- ✅ Auto git merge (Fast-forward + Three-way)
- ✅ AI task orchestration
- ✅ TUI monitoring

**V2.1 (Coming):**
- Web dashboard
- Enhanced DAG scheduling
- Prometheus metrics
- Conflict resolution tools

---

## 💡 FAQ

**Q: Is it free?**
A: Yes! Uses free Claude CLI. No API costs.

**Q: What about file conflicts?**
A: Each agent works in isolated git worktree. Auto-merge handles the rest.

**Q: What if merge conflicts occur?**
A: System detects conflicts, auto-aborts, and logs clearly. First agent wins, others preserved for review.

**Q: Do I need Gemini API?**
A: Optional. Only needed for `orchestrate` and `--with-brain` features.

---

## 🤝 Contributing

```bash
git clone https://github.com/Cz07cring/claude-swarm.git
cd claude-swarm
make test
# Make changes, submit PR
```

---

## 📄 License

MIT License - see [LICENSE](LICENSE)

---

<div align="center">

**⚡ Claude Swarm v2.0**

**🚀 10s/task** • **🔀 Auto-merge** • **🧠 AI-powered** • **💯 Free**

[GitHub](https://github.com/Cz07cring/claude-swarm) • [Issues](https://github.com/Cz07cring/claude-swarm/issues)

</div>
