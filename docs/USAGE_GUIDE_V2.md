# Claude Swarm V2 使用指南

## 📚 目录

1. [快速开始](#快速开始)
2. [基本命令](#基本命令)
3. [任务管理](#任务管理)
4. [最佳实践](#最佳实践)
5. [故障排查](#故障排查)

---

## 快速开始

### 1. 准备任务文件

创建任务文件 `~/.claude-swarm/tasks.json`:

```bash
mkdir -p ~/.claude-swarm

cat > ~/.claude-swarm/tasks.json << 'EOF'
{
  "tasks": [
    {
      "id": "task-1",
      "description": "创建一个简单的 hello.go 文件，包含 main 函数打印 Hello World",
      "status": "pending",
      "priority": 5,
      "retry_count": 0,
      "max_retries": 3
    }
  ]
}
EOF
```

### 2. 启动 Swarm

```bash
# 使用 3 个 agents
./swarm start-v2 --agents 3

# 指定自定义任务文件
./swarm start-v2 --agents 3 --tasks /path/to/custom-tasks.json
```

### 3. 监控进度

在另一个终端窗口：

```bash
# 实时监控任务状态
watch -n 1 'cat ~/.claude-swarm/tasks.json | jq ".tasks[] | {id, status}"'

# 查看完整任务信息
cat ~/.claude-swarm/tasks.json | jq .
```

---

## 基本命令

### start-v2

启动 Claude Swarm V2 系统。

```bash
swarm start-v2 [flags]
```

**参数:**
- `--agents <num>`: Agent 数量（默认: 3）
- `--tasks <path>`: 任务文件路径（默认: ~/.claude-swarm/tasks.json）

**示例:**

```bash
# 启动 5 个 agents
swarm start-v2 --agents 5

# 使用自定义任务文件
swarm start-v2 --agents 3 --tasks ./my-tasks.json

# 后台运行并保存日志
swarm start-v2 --agents 3 > /tmp/swarm.log 2>&1 &
```

**停止:**

按 `Ctrl+C` 优雅停止 swarm。系统会：
1. 停止调度器
2. 等待正在执行的任务完成
3. 重置 in_progress 任务为 pending
4. 清理 worktrees

---

## 任务管理

### 任务文件格式

```json
{
  "tasks": [
    {
      "id": "unique-task-id",
      "description": "任务描述（会发送给 Claude）",
      "status": "pending|in_progress|completed|failed",
      "assignee_id": "agent-0",  // 可选，正在执行的 agent
      "priority": 5,             // 1-10，数字越大优先级越高
      "retry_count": 0,          // 当前重试次数
      "max_retries": 3,          // 最大重试次数
      "dependencies": ["task-1", "task-2"],  // 依赖的任务 IDs
      "last_error": ""           // 最后的错误信息
    }
  ]
}
```

### 任务状态

| 状态 | 说明 |
|------|------|
| `pending` | 等待执行 |
| `in_progress` | 正在执行中 |
| `completed` | 已完成 |
| `failed` | 失败（超过最大重试次数） |

### DAG 依赖调度

V2 支持任务依赖关系（DAG）：

```json
{
  "tasks": [
    {
      "id": "build-api",
      "description": "创建 RESTful API server.go",
      "status": "pending",
      "priority": 5
    },
    {
      "id": "write-tests",
      "description": "为 server.go 编写测试",
      "status": "pending",
      "dependencies": ["build-api"],  // 依赖 build-api
      "priority": 5
    },
    {
      "id": "write-docs",
      "description": "编写 API 文档",
      "status": "pending",
      "dependencies": ["build-api"],  // 依赖 build-api
      "priority": 3
    }
  ]
}
```

**执行顺序:**
1. `build-api` 先执行
2. `build-api` 完成后，`write-tests` 和 `write-docs` 并行执行
3. `write-tests` 优先级更高，会先被分配

---

## 最佳实践

### 1. 任务描述要清晰具体

✅ **好的描述:**
```json
{
  "description": "创建一个 HTTP server (Go)，监听 8080 端口，包含 /health 健康检查接口和 /api/users GET 接口"
}
```

❌ **不好的描述:**
```json
{
  "description": "做一个服务器"
}
```

### 2. 合理设置优先级

- **高优先级 (8-10)**: 紧急的、快速的任务
- **中优先级 (5-7)**: 正常任务
- **低优先级 (1-4)**: 不紧急的任务

### 3. 使用 DAG 依赖管理复杂项目

对于大型项目，将任务分解为多个小任务，使用依赖关系串联：

```json
{
  "tasks": [
    {"id": "1-setup", "description": "初始化项目结构"},
    {"id": "2-models", "description": "创建数据模型", "dependencies": ["1-setup"]},
    {"id": "3-api", "description": "实现 API 接口", "dependencies": ["2-models"]},
    {"id": "4-tests", "description": "编写测试", "dependencies": ["3-api"]}
  ]
}
```

### 4. 监控和调试

```bash
# 实时监控（每秒刷新）
watch -n 1 'cat ~/.claude-swarm/tasks.json | jq ".tasks[] | {id, status}"'

# 查看失败的任务
cat ~/.claude-swarm/tasks.json | jq '.tasks[] | select(.status=="failed")'

# 查看正在执行的任务
cat ~/.claude-swarm/tasks.json | jq '.tasks[] | select(.status=="in_progress")'

# 查看任务统计
cat ~/.claude-swarm/tasks.json | jq '.tasks | group_by(.status) | map({status: .[0].status, count: length})'
```

### 5. 处理失败的任务

V2 有自动重试机制，但如果任务仍然失败：

```bash
# 1. 查看失败原因
cat ~/.claude-swarm/tasks.json | jq '.tasks[] | select(.id=="task-1") | .last_error'

# 2. 修改任务描述或重置状态
# 编辑 tasks.json，将 status 改为 "pending"，retry_count 改为 0

# 3. 重新启动 swarm
./swarm start-v2 --agents 3
```

---

## 故障排查

### 问题 1: Agent 卡住不动

**原因**: 任务可能需要人工确认

**解决**:
- 检查日志：`tail -f /tmp/swarm-v2-run.log`
- V2 使用 `--dangerously-skip-permissions`，应该自动确认
- 如果仍然卡住，可能是任务描述有问题

### 问题 2: 任务一直失败

**原因**: 任务描述可能不清楚或不可执行

**解决**:
1. 查看错误：`cat ~/.claude-swarm/tasks.json | jq '.tasks[] | select(.id=="task-1") | .last_error'`
2. 简化任务描述，使其更具体
3. 降低任务复杂度，拆分为多个子任务

### 问题 3: Git worktree 冲突

**原因**: 之前的运行没有正确清理

**解决**:
```bash
# 清理所有 worktrees
rm -rf .worktrees
git worktree prune

# 删除 agent 分支
git branch -D agent-0-branch agent-1-branch agent-2-branch

# 重新启动
./swarm start-v2 --agents 3
```

### 问题 4: 磁盘空间不足

**原因**: Worktrees 占用磁盘空间

**解决**:
```bash
# 检查 worktrees 大小
du -sh .worktrees

# 停止 swarm，清理后重启
# Ctrl+C 停止，然后：
rm -rf .worktrees
git worktree prune
./swarm start-v2 --agents 3
```

---

## 性能调优

### Agent 数量选择

| 场景 | 建议 Agent 数 |
|------|--------------|
| 小项目（< 10 任务） | 2-3 |
| 中型项目（10-50 任务） | 3-5 |
| 大型项目（> 50 任务） | 5-10 |

**注意**:
- 每个 agent 需要独立的 worktree（约 100MB+）
- Agent 数量过多不会提升性能（受限于 Claude CLI 响应时间）
- 建议从 3 个开始，根据实际情况调整

### 任务分配策略

- **短任务优先**: 设置更高的优先级
- **长任务并行**: 确保没有不必要的依赖关系
- **合理重试**: 设置 `max_retries = 2-3`

---

## 高级用法

### 1. 鲁棒性测试

运行长时间测试验证系统稳定性：

```bash
# 5 分钟测试
./test-robustness.sh 3 300

# 30 分钟测试
./test-robustness.sh 5 1800
```

### 2. 自定义日志位置

```bash
# 保存日志到指定位置
./swarm start-v2 --agents 3 > ~/my-swarm-$(date +%Y%m%d-%H%M%S).log 2>&1
```

### 3. 批量任务生成

```bash
# 使用脚本生成任务
cat > generate-tasks.sh << 'EOF'
#!/bin/bash
cat > ~/.claude-swarm/tasks.json << 'JSON'
{
  "tasks": [
EOF

for i in {1..10}; do
    cat >> generate-tasks.sh << EOF
    {"id": "task-$i", "description": "创建 file$i.txt", "status": "pending", "priority": 5},
EOF
done

cat >> generate-tasks.sh << 'EOF'
  ]
}
JSON
EOF

chmod +x generate-tasks.sh
./generate-tasks.sh
```

---

## 参考资料

- [V2 架构完整报告](./V2_INTEGRATION_COMPLETE.md)
- [TUI 监控指南](./TUI_MONITOR.md)
- [主项目 README](../README.md)

---

**Last Updated**: 2026-02-01
**Version**: V2.0
