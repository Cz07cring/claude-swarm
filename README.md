# Claude Swarm 🐝

<div align="center">

**AI-Powered Multi-Agent Collaborative Development System**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Gemini](https://img.shields.io/badge/Powered_by-Gemini-4285F4?style=flat&logo=google)](https://ai.google.dev/)

[English](README.md) • [简体中文](README_ZH.md)

[Quick Start](#-quick-start) • [Features](#-features) • [Usage](#-usage) • [Documentation](#-documentation)

</div>

---

## Introduction

Claude Swarm is an innovative **AI-driven multi-agent collaboration system** that automatically splits tasks and orchestrates multiple Claude Code instances for parallel development with just one sentence describing your requirements.

```bash
# Launch complete development workflow with one command
swarm orchestrate "Build a Todo app with add, delete, and complete features"

# AI automatically splits into 8-15 tasks, then executes in parallel
swarm start --agents 8
```

**Core Philosophy**: Automate the traditional "manual task splitting → developer assignment" workflow through AI intelligent analysis and multi-agent parallel execution, dramatically boosting development efficiency.

---

## ✨ Features

### 🧠 AI Orchestrator (v2.0)

**Intelligent Requirement Analysis and Task Decomposition**

- **One-Sentence Task Queue Generation** - Describe requirements, AI auto-splits into 8-15 executable tasks
- **Modular Decomposition** - Intelligently identifies independent functional modules (3-8 modules)
- **Dependency Management** - Automatically builds task dependency graph (DAG)
- **Precise Task Descriptions** - Each task includes specific implementation steps and acceptance criteria

<details>
<summary>View AI Analysis Example</summary>

```bash
$ swarm orchestrate "Implement user authentication system"

🧠 AI Orchestrator analyzing...

════════════════════════════════════════════════════════════
📊 AI Analysis Results
════════════════════════════════════════════════════════════

📌 Summary: User authentication system (registration, login, JWT)
🎯 Complexity: medium
⏱️  Estimated Time: 8-12h

🔧 Module Breakdown (4 modules):
  1. DatabaseSchema - User table design
  2. AuthAPI - Registration and login API
  3. JWTService - Token generation and validation
  4. Testing - Unit and integration tests

📋 Task List (10 tasks):
  🟢 Task-1: Create users table schema...
  🔵 Task-2: Implement POST /api/register...
  🔵 Task-3: Implement POST /api/login...
  ...

✅ Task queue created! Total: 10 tasks
```

</details>

### 🐝 Swarm Collaboration (v1.0)

**Multi-Agent Parallel Development**

- **Parallel Execution** - Run 1-100 Claude Code instances simultaneously
- **Intelligent Scheduling** - Auto-assign tasks to idle agents
- **Status Monitoring** - Real-time detection of agent states (working/idle/waiting/error)
- **Auto Rescue** - Intelligent handling of confirmations, error recovery, stuck detection

### 🎨 TUI Visualization

**Real-time Monitoring Dashboard**

- **Agent Grid** - Visual display of all agent states (up to 5x5 grid)
- **Task List** - Real-time view of task progress and status
- **Log Viewer** - View selected agent's real-time output
- **Keyboard Navigation** - Tab to switch panels, j/k to navigate, q to quit

```
┌─────────────────────────────────────────────────────────────┐
│ Claude Swarm Monitor                       Working:3 Idle:2 │
├─────────────────────────────────────────────────────────────┤
│  Agent Grid (3x3)         │  Agent-0 Logs                   │
│  ┌─────┬─────┬─────┐     │  • Analyzing requirements...    │
│  │ 0 ⚡│ 1 ⚡│ 2 💤│     │  • Creating file auth.go        │
│  │Work │Work │Idle │     │  • Running tests...             │
│  ├─────┼─────┼─────┤     │                                 │
│  │ 3 ⚡│ 4 💤│     │     │                                 │
│  └─────┴─────┴─────┘     │                                 │
├─────────────────────────────────────────────────────────────┤
│  Task List                                                  │
│  ✅ Task-1: Create database tables                          │
│  🔄 Task-2: Implement registration API     [Agent-0]       │
│  🔄 Task-3: Implement login API            [Agent-1]       │
│  ⏳ Task-4: JWT Token validation                           │
└─────────────────────────────────────────────────────────────┘
```

---

## 🚀 Quick Start

### Prerequisites

| Dependency | Version | Installation |
|------------|---------|--------------|
| **Go** | 1.21+ | [go.dev/doc/install](https://go.dev/doc/install) |
| **tmux** | Latest | `brew install tmux` (macOS)<br>`apt install tmux` (Ubuntu) |
| **Claude Code** | Latest | [claude.ai/claude-code](https://claude.ai/claude-code) |
| **Gemini API Key** | - | [ai.google.dev](https://ai.google.dev/) (Optional, for AI Orchestrator) |

### Installation

```bash
# 1. Clone repository
git clone https://github.com/Cz07cring/claude-swarm.git
cd claude-swarm

# 2. Build
go build -o swarm ./cmd/swarm

# 3. (Optional) Configure Gemini API Key for AI Orchestrator
export GEMINI_API_KEY="your-api-key-here"
echo 'export GEMINI_API_KEY="your-key"' >> ~/.bashrc
```

### 3 Steps to Get Started

```bash
# 1. Start agent cluster (5 agents)
./swarm start --agents 5

# 2. Add tasks
./swarm add-task "Create an HTTP server"
./swarm add-task "Write unit tests"

# 3. Monitor progress
./swarm monitor  # TUI visual monitoring (recommended)
# or
./swarm status   # CLI status query
```

---

## 📖 Usage

### Scenario 1: AI Orchestrator Auto-Split (Recommended 🧠)

**Best For**: New feature development, modular refactoring, complex requirements

```bash
# Describe requirements in one sentence
./swarm orchestrate "Implement real-time chat with text, images, and online status"

# AI auto-generates 15 tasks including:
# - WebSocket module
# - Message storage module
# - File upload module
# - Online status module
# - Frontend components

# Start 10 agents for parallel development
./swarm start --agents 10

# Real-time monitoring with TUI
./swarm monitor
```

**Time Saved**: **60-80%** compared to serial development

### Scenario 2: Manual Task Addition

**Best For**: Known task list, precise control

```bash
# Start cluster
./swarm start --agents 3

# Batch add tasks
./swarm add-task "Implement user registration API"
./swarm add-task "Implement user login API"
./swarm add-task "Implement password reset API"
./swarm add-task "Write API documentation"

# Check status
./swarm status
```

### Scenario 3: Batch Repetitive Tasks

```bash
# Start cluster
./swarm start --agents 5

# Batch add tasks (shell loop)
for feature in login register profile settings dashboard
do
  ./swarm add-task "Write unit tests for $feature feature"
done

# Real-time monitoring
watch -n 2 './swarm status'
```

---

## 📋 Command Reference

### Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `orchestrate` | 🧠 AI requirement analysis | `swarm orchestrate "requirement description"` |
| `start` | Start agent cluster | `swarm start --agents 5` |
| `add-task` | Add task to queue | `swarm add-task "task description"` |
| `monitor` | 🎨 TUI visual monitoring | `swarm monitor` |
| `status` | View cluster status | `swarm status` |
| `stop` | Stop cluster | `swarm stop` |

### `orchestrate` - AI Orchestrator

```bash
swarm orchestrate [requirement description] [flags]

Flags:
  -k, --api-key string   Gemini API Key (or use env var GEMINI_API_KEY)
      --auto-start       Auto-start agent cluster after analysis
  -n, --agents int       Number of agents (default: 5)

Examples:
  # Basic usage
  swarm orchestrate "Build a blog system"

  # Auto-start after analysis
  swarm orchestrate --auto-start "Optimize database performance"

  # Specify API Key and agent count
  swarm orchestrate -k "your-key" -n 10 "Refactor auth system"
```

### `start` - Start Cluster

```bash
swarm start [flags]

Flags:
  -n, --agents int      Number of agents (default: 3)
  -i, --interval int    Monitoring interval in seconds (default: 5)
  -s, --session string  tmux session name (default: claude-swarm)

Examples:
  # Start 5 agents with 3-second monitoring interval
  swarm start -n 5 -i 3

  # Custom session name
  swarm start -s dev-swarm
```

### `monitor` - TUI Monitor

```bash
swarm monitor

Keyboard Shortcuts:
  Tab       Switch panels (Agent Grid ⇄ Task List)
  j/k       Navigate up/down
  ↑/↓       Navigate up/down
  h/l       Navigate left/right (Agent Grid)
  ←/→       Navigate left/right (Agent Grid)
  Home      Jump to first
  End       Jump to last
  Enter     Select agent to view logs
  q/Esc     Quit
```

---

## 🎨 TUI Monitor Panel

### Features

| Panel | Functionality | Shortcuts |
|-------|--------------|-----------|
| **Agent Grid** | Display all agent states (working/idle/error)<br>Dynamic grid size (2x2 to 5x5) | h/j/k/l navigation<br>Enter to view logs |
| **Task List** | Real-time task status and progress<br>Color-coded (green=done, blue=active) | j/k to scroll |
| **Log Viewer** | Selected agent's real-time output<br>Auto-scroll to bottom | PageUp/Down to scroll |
| **Status Bar** | Cluster stats (working/idle agent count, task completion) | - |

### Agent Status Icons

| Icon | State | Description |
|------|-------|-------------|
| ⚡ | Working | Agent executing task |
| 💤 | Idle | Agent waiting for task |
| ⏸️ | Waiting | Agent waiting for user input |
| ❌ | Error | Agent encountered error |
| ⏱️ | Stuck | Agent unresponsive |

📖 **Detailed Docs**: [TUI Monitor Guide](docs/tui/TUI_DEMO.md)

---

## 📁 Project Structure

```
claude-swarm/
├── cmd/swarm/              # CLI entry points
│   ├── main.go
│   ├── orchestrate.go      # AI Orchestrator command
│   ├── start.go            # Start cluster
│   ├── monitor.go          # TUI monitor
│   └── ...
├── pkg/
│   ├── orchestrator/       # AI Orchestrator (Gemini)
│   ├── controller/         # Coordinator (scheduling, monitoring)
│   ├── tui/                # TUI components
│   ├── state/              # Task queue management
│   └── tmux/               # tmux session management
├── docs/                   # 📚 Documentation
│   ├── guides/             #   User guides
│   ├── reports/            #   Test reports
│   └── tui/                #   TUI docs
├── scripts/                # 🔧 Scripts
│   ├── tests/              #   Automated tests
│   └── tools/              #   Dev tools
└── config.yaml.example     # Config template
```

📖 **Detailed Structure**: [DIRECTORY_STRUCTURE.md](DIRECTORY_STRUCTURE.md)

---

## 🏗️ Architecture

### Workflow

```
┌─────────────────┐
│  User Input     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌──────────────────┐
│ AI Orchestrator │────▶│  Task Queue      │
│  - Analysis     │     │  - pending       │
│  - Modularize   │     │  - in_progress   │
│  - Generate     │     │  - completed     │
│  - Dependencies │     └────────┬─────────┘
└─────────────────┘              │
                                 ▼
                        ┌─────────────────┐
                        │   Scheduler     │
                        │  - Assign tasks │
                        │  - Load balance │
                        └────────┬─────────┘
                                 │
                ┌────────────────┼────────────────┐
                │                │                │
                ▼                ▼                ▼
         ┌──────────┐     ┌──────────┐    ┌──────────┐
         │ Agent 0  │     │ Agent 1  │... │ Agent N  │
         │(tmux pane)│     │(tmux pane)│    │(tmux pane)│
         └────┬─────┘     └────┬─────┘    └────┬─────┘
              │                │               │
              └────────────────┼───────────────┘
                               │
                               ▼
                        ┌─────────────────┐
                        │    Monitor      │
                        │  - Detect state │
                        │  - Auto-confirm │
                        │  - Error recover│
                        │  - Stuck detect │
                        └────────┬─────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │   TUI Panel     │
                        │  - Realtime viz │
                        └─────────────────┘
```

### Core Components

| Component | Functionality | Tech Stack |
|-----------|---------------|------------|
| **AI Orchestrator** | Requirement analysis, task splitting, dependency mgmt | Gemini 3 Flash Preview |
| **Task Queue** | Task storage and state management | JSON files + file locks |
| **Coordinator** | Task scheduling, agent monitoring, auto-rescue | Go Goroutines |
| **tmux Manager** | Session management, pane control, output capture | tmux API |
| **TUI Dashboard** | Real-time visual monitoring | Bubble Tea + Lipgloss |

---

## 🛠️ Development

### Build

```bash
# Development mode
go run ./cmd/swarm start

# Build binary
go build -o swarm ./cmd/swarm

# Cross-platform build
GOOS=linux GOARCH=amd64 go build -o swarm-linux ./cmd/swarm
GOOS=darwin GOARCH=arm64 go build -o swarm-darwin ./cmd/swarm
```

### Testing

```bash
# Run all tests
go test ./...

# Test coverage
go test -cover ./...

# Integration tests
./scripts/tests/run-full-test.sh

# TUI tests
./scripts/tests/test-tui.sh
```

---

## 📚 Documentation

### User Guides

- [User Guide](docs/guides/USER_GUIDE.md) - Complete tutorial
- [Configuration Guide](docs/guides/CONFIG_GUIDE.md) - Config details
- [Getting Started](docs/guides/GETTING_STARTED.md) - Beginner's guide

### TUI Related

- [TUI Demo](docs/tui/TUI_DEMO.md) - Monitor panel usage
- [TUI Optimization](docs/tui/TUI_OPTIMIZATION_SUMMARY.md) - Features
- [TUI UX Improvements](docs/tui/TUI_UX_IMPROVEMENTS.md) - UX enhancements

### Development Docs

- [Architecture Design](docs/architecture/full-plan.md) - Complete implementation plan
- [Gemini Setup](docs/GEMINI_SETUP.md) - API configuration guide
- [Test Reports](docs/reports/) - Various test reports

---

## 🗺️ Roadmap

### ✅ Completed

- **v1.0 MVP** - Basic swarm system
  - tmux session management
  - Task queue and scheduling
  - Status monitoring and auto-rescue
  - CLI commands

- **v2.0 AI Orchestrator**
  - Gemini intelligent requirement analysis
  - Auto task splitting
  - Dependency identification
  - TUI visual monitoring

### 🚧 In Progress

- **v2.1 Enhanced Scheduling**
  - DAG dependency scheduling
  - File conflict avoidance
  - Task timeout and retry

- **v2.2 Git Worktree**
  - Agent independent branch development
  - Auto merge and conflict resolution

### ⏳ Planned

- **v3.0 Persistence**
  - SQLite database (replace JSON)
  - Task history and statistics

- **v3.1 Cross-platform**
  - Windows support
  - Docker images

- **v4.0 Web Interface**
  - Web dashboard
  - Remote control and collaboration

---

## 💡 FAQ

<details>
<summary><b>Q: tmux session creation failed?</b></summary>

```bash
# Check if tmux is installed
which tmux

# View existing sessions
tmux ls

# Manually kill old session
tmux kill-session -t claude-swarm
```
</details>

<details>
<summary><b>Q: Task queue corrupted?</b></summary>

```bash
# Backup task queue
cp ~/.claude-swarm/tasks.json ~/.claude-swarm/tasks.json.bak

# Remove corrupted queue
rm ~/.claude-swarm/tasks.json

# Restart
./swarm start
```
</details>

<details>
<summary><b>Q: Agent not responding?</b></summary>

```bash
# Attach to tmux to view real-time output
tmux attach -t claude-swarm

# View agent logs in TUI monitor
./swarm monitor

# Restart cluster
./swarm stop
./swarm start
```
</details>

<details>
<summary><b>Q: Gemini API quota exceeded?</b></summary>

Gemini 3 Flash Preview free quota:
- 60 requests/minute
- 1500 requests/day

If quota exceeded:
1. Upgrade to paid API
2. Use manual mode (`add-task`)
3. Reduce usage frequency
</details>

---

## 📊 Performance Comparison

| Scenario | Traditional Dev | Claude Swarm | Time Saved |
|----------|----------------|--------------|------------|
| **10 independent modules** | Serial 20h | 5 agents parallel 6h | **70%** ⬇️ |
| **Task splitting** | Manual 2-3h | AI 15s | **99%** ⬇️ |
| **100 unit tests** | Serial 10h | 10 agents parallel 2h | **80%** ⬇️ |

---

## 🤝 Contributing

Contributions welcome! Please follow these steps:

1. Fork this repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Create Pull Request

---

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file for details

---

## 📧 Contact

- **GitHub**: [@Cz07cring](https://github.com/Cz07cring)
- **Issues**: [Submit Issue](https://github.com/Cz07cring/claude-swarm/issues)
- **Discussions**: [Join Discussion](https://github.com/Cz07cring/claude-swarm/discussions)

---

## 🙏 Acknowledgments

- [tmux](https://github.com/tmux/tmux) - Terminal multiplexer
- [Claude Code](https://claude.ai/claude-code) - AI coding assistant
- [Google Gemini](https://ai.google.dev/) - AI Orchestrator
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [AI Swarm Collaboration](https://github.com/tukuaiai/vibe-coding-cn) - Inspiration

---

<div align="center">

**⚠️ Notice**: This is an experimental project. Please test thoroughly before production use.

Made with ❤️ by Claude Sonnet 4.5

</div>
