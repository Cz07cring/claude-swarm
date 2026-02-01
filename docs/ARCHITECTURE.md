# Claude Swarm 系统架构

**日期**: 2026-02-01
**架构**: 直接 Claude CLI 执行

---

## 架构概览

### 核心组件

- **ClaudeExecutor** - Claude CLI 执行器
  - 使用 `echo | claude --dangerously-skip-permissions`
  - AI 风险评估
  - 自动重试错误检测
  - 10-12秒/任务性能

- **Agent** - 简化的 Agent 结构
  - 任务通道通信
  - 独立 worktree 执行

- **Coordinator** - 任务调度器
  - 基于 Go 协程的 worker 模式
  - DAG 任务调度
  - 自动重试机制

- **启动命令** - `swarm start`
  - 简化的命令行接口
  - 优雅的信号处理
  - 自动清理

### 核心功能

- **基础执行** - 任务成功执行并完成
- **文件创建** - Claude 真实创建文件
- **任务队列** - JSON 文件持久化
- **状态更新** - 自动更新任务状态
- **Worktree 隔离** - 每个 agent 独立分支
- **AI 风险评估** - 执行前自动评估

---

## 性能指标

| 指标 | 数值 |
|------|------|
| 启动时间 | ~1秒 |
| Agent 初始化 | 即时 |
| 任务执行 | 10-12秒/任务 |
| 资源使用 | 低 |

---

## 技术架构

```
┌─────────────────┐
│   swarm start   │
└────────┬────────┘
         │
┌────────▼─────────────┐
│    Coordinator       │
│  - Scheduler Loop    │
│  - Worker Pools      │
└────────┬─────────────┘
         │
    ┌────┴────┐
    │         │
┌───▼──┐  ┌──▼───┐
│Agent0│  │Agent1│
└───┬──┘  └──┬───┘
    │        │
┌───▼────────▼───┐
│ ClaudeExecutor │
│ echo | claude   │
└────────────────┘
```

---

## 使用方法

### 快速开始

```bash
# 1. 准备任务
cat > ~/.claude-swarm/tasks.json << 'EOF'
{
  "tasks": [
    {
      "id": "task-1",
      "description": "创建 hello.txt，内容是 world",
      "status": "pending",
      "priority": 5,
      "retry_count": 0,
      "max_retries": 3
    }
  ]
}
EOF

# 2. 启动 swarm
./swarm start --agents 2

# 3. 监控（另一个终端）
watch -n 1 'cat ~/.claude-swarm/tasks.json | jq ".tasks[] | {id, status}"'
```

### 鲁棒性测试

```bash
# 运行 5 分钟测试，3 个 agents
./test-robustness.sh 3 300

# 运行 30 分钟长时间测试
./test-robustness.sh 5 1800
```

---

## 设计原则

### 成功因素

1. **简化架构** - 移除复杂性
2. **直接执行** - 使用 Claude CLI 原生能力
3. **免费方案** - 避免 API 费用
4. **AI 评估** - 智能风险控制

### 核心优势

- 核心功能完整
- 性能符合预期
- 完全免费方案
- 可扩展架构
- 生产就绪

---

## 后续优化方向

- 改进状态检测
- 实现自动 git merge
- TUI 实时监控
- Prometheus metrics
- 结构化日志

---

*Claude Swarm Architecture Documentation*
