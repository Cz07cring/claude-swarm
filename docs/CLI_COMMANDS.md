# Claude Swarm CLI 命令参考

本文档介绍 Claude Swarm 的所有 CLI 命令及其用法。

## 核心命令

### add-task - 快速添加单个任务

快速添加单个任务到任务队列，支持设置优先级、依赖关系和重试次数。

**基础用法**:
```bash
swarm add-task "创建 README.md 文件"
```

**高级用法**:
```bash
# 设置优先级（1-10，10最高）
swarm add-task "编写单元测试" --priority 8

# 设置依赖关系
swarm add-task "编写单元测试" --dependencies task-1,task-2

# 设置最大重试次数
swarm add-task "部署到生产环境" --max-retries 5

# 自定义任务 ID
swarm add-task "初始化项目" --id init-project

# 组合使用
swarm add-task "编写测试" \
  --priority 8 \
  --dependencies task-1,task-2 \
  --max-retries 5 \
  --id test-task
```

**参数说明**:
- `--priority, -p`: 任务优先级（1-10），默认 5
- `--dependencies, -d`: 依赖的任务 ID（逗号分隔）
- `--max-retries`: 最大重试次数，默认 3
- `--id`: 自定义任务 ID（留空自动生成）
- `--queue`: 任务队列文件路径，默认 `~/.claude-swarm/tasks.json`

---

### status - 查看任务队列状态

查看任务队列的当前状态，包括统计信息和任务详情。

**基础用法**:
```bash
swarm status
```

**输出示例**:
```
📊 Claude Swarm 任务状态
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📈 统计:
  ✅ 已完成: 8 / 12 (67%)
  🔄 进行中: 2
  ⏳ 待执行: 1
  ❌ 失败: 1

  [███████████████████████████░░░░░░░░░] 67%

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 任务详情:

[in_progress] task-2 (优先级: 7, 分配给: agent-0)
  🔄 实现 API 接口
  依赖: task-1 ✓

[pending] task-3 (优先级: 6)
  ⏳ 编写测试
  依赖: task-2 ⚠️ 未满足
```

**高级用法**:
```bash
# 详细模式（显示创建时间、更新时间等）
swarm status --verbose

# 仅显示特定状态的任务
swarm status --filter pending
swarm status --filter in_progress
swarm status --filter completed
swarm status --filter failed
```

**参数说明**:
- `--verbose, -v`: 显示详细信息
- `--filter, -f`: 过滤任务状态（pending/in_progress/completed/failed）
- `--queue`: 任务队列文件路径，默认 `~/.claude-swarm/tasks.json`

---

### batch-add - 批量添加任务

从文件、stdin 或交互式模式批量添加任务。

**文件格式**（每行一个任务）:
```
# 注释行会被忽略
描述文本 | priority:8
实现登录接口 | priority:7 | depends:task-1
编写测试 | priority:6 | depends:task-2 | max-retries:5
部署到生产 | priority:9 | id:deploy-task
```

**从文件添加**:
```bash
swarm batch-add --file tasks.txt
```

**从 stdin 添加**:
```bash
cat tasks.txt | swarm batch-add --stdin

# 或使用 heredoc
swarm batch-add --stdin <<EOF
创建数据库模型 | priority:9
实现API接口 | priority:8
编写文档 | priority:5
EOF
```

**交互式模式**:
```bash
swarm batch-add --interactive
# 然后逐行输入任务，空行结束
```

**参数语法**:
- `priority:X` 或 `p:X`: 设置优先级（1-10）
- `depends:task-1,task-2` 或 `d:task-1,task-2`: 设置依赖
- `max-retries:X` 或 `r:X`: 设置最大重试次数
- `id:custom-id`: 设置自定义 ID

**参数说明**:
- `--file, -f`: 从文件读取任务
- `--stdin`: 从标准输入读取任务
- `--interactive, -i`: 交互式模式
- `--queue`: 任务队列文件路径，默认 `~/.claude-swarm/tasks.json`

---

### clean - 清理任务队列

清理任务队列中的已完成、失败或所有任务。

**清理已完成的任务**:
```bash
swarm clean --completed
```

**清理失败的任务**:
```bash
swarm clean --failed
```

**清理所有任务**（危险操作）:
```bash
swarm clean --all
```

**跳过确认提示**:
```bash
swarm clean --completed --force
```

**参数说明**:
- `--completed`: 清理已完成的任务
- `--failed`: 清理失败的任务
- `--all`: 清理所有任务（危险操作）
- `--force, -f`: 跳过确认提示
- `--queue`: 任务队列文件路径，默认 `~/.claude-swarm/tasks.json`

**注意**:
- 必须指定且只能指定一种清理模式
- 使用 `--all` 时会要求确认，除非使用 `--force`
- 清理操作不可撤销

---

### orchestrate - AI 主脑分析需求

使用 Gemini AI 分析需求并自动拆分任务。

**基础用法**:
```bash
swarm orchestrate "实现用户登录系统"
```

**高级用法**:
```bash
# 自动审批并启动 Agent 集群
swarm orchestrate "创建博客系统" --auto-start --agents 5

# 跳过人工审批
swarm orchestrate "重构认证模块" --auto-approve

# 指定配置文件和 API Key
swarm orchestrate "优化数据库查询" \
  --config ./config.yaml \
  --api-key YOUR_API_KEY
```

**参数说明**:
- `--api-key, -k`: Gemini API Key
- `--config, -c`: 配置文件路径
- `--auto-start`: 分析并审批通过后自动启动 Agent 集群
- `--auto-approve`: 跳过人工审批，自动创建任务
- `--agents, -n`: Agent 数量（1-10），默认 5
- `--tasks`: 任务队列文件路径，默认 `~/.claude-swarm/tasks.json`

---

### start - 启动 Agent 集群

启动 Claude Swarm Agent 集群执行任务。

**基础用法**:
```bash
swarm start
```

**高级用法**:
```bash
# 指定 Agent 数量
swarm start --agents 5

# 指定任务队列文件
swarm start --tasks ./my-tasks.json
```

**参数说明**:
- `--agents, -n`: Agent 数量（1-10），默认 3
- `--tasks, -t`: 任务队列文件路径，默认 `~/.claude-swarm/tasks.json`

---

### monitor - 启动 TUI 监控面板

启动交互式监控面板查看 Agent 状态和任务进度。

**用法**:
```bash
swarm monitor
```

---

## 完整工作流示例

### 示例 1: 手动创建和管理任务

```bash
# 1. 添加任务
swarm add-task "创建数据库模型" --priority 9 --id db-models
swarm add-task "实现 API 接口" --priority 8 --dependencies db-models
swarm add-task "编写测试" --priority 7 --dependencies db-models

# 2. 查看状态
swarm status

# 3. 启动执行
swarm start --agents 3

# 4. 在另一个终端监控进度
swarm monitor

# 5. 执行完成后查看状态
swarm status

# 6. 清理已完成的任务
swarm clean --completed --force
```

### 示例 2: 使用 AI 主脑和批量添加

```bash
# 1. 使用 AI 主脑分析需求并生成任务
swarm orchestrate "实现用户认证系统"

# 2. 手动添加额外任务
swarm add-task "添加登录日志" --priority 6

# 3. 批量添加测试任务
cat > test-tasks.txt <<EOF
单元测试 - 用户模型 | priority:8
单元测试 - 认证服务 | priority:8
集成测试 - 登录流程 | priority:7
EOF

swarm batch-add --file test-tasks.txt

# 4. 查看所有任务
swarm status --verbose

# 5. 启动执行
swarm start --agents 5

# 6. 清理
swarm clean --completed --force
```

### 示例 3: 处理失败任务

```bash
# 1. 查看失败的任务
swarm status --filter failed

# 2. 清理失败任务（需要重新创建）
swarm clean --failed --force

# 3. 重新添加修正后的任务
swarm add-task "修正后的任务" --priority 8
```

---

## 任务队列文件格式

任务队列存储在 `~/.claude-swarm/tasks.json`:

```json
{
  "tasks": [
    {
      "id": "task-1",
      "description": "创建数据库模型",
      "status": "completed",
      "assignee_id": "agent-0",
      "created_at": "2026-02-01T10:00:00Z",
      "updated_at": "2026-02-01T10:05:00Z",
      "dependencies": [],
      "priority": 9,
      "retry_count": 0,
      "max_retries": 3,
      "last_error": ""
    },
    {
      "id": "task-2",
      "description": "实现 API 接口",
      "status": "in_progress",
      "assignee_id": "agent-1",
      "created_at": "2026-02-01T10:01:00Z",
      "updated_at": "2026-02-01T10:06:00Z",
      "dependencies": ["task-1"],
      "priority": 8,
      "retry_count": 0,
      "max_retries": 3,
      "last_error": ""
    }
  ]
}
```

**字段说明**:
- `id`: 任务唯一标识符
- `description`: 任务描述
- `status`: 任务状态（pending/in_progress/completed/failed）
- `assignee_id`: 分配的 Agent ID
- `created_at`: 创建时间
- `updated_at`: 最后更新时间
- `dependencies`: 依赖的任务 ID 列表
- `priority`: 优先级（1-10）
- `retry_count`: 当前重试次数
- `max_retries`: 最大重试次数
- `last_error`: 最后一次失败的错误信息

---

## 常见问题

### Q: 如何查看任务的依赖关系？
A: 使用 `swarm status --verbose` 可以看到每个任务的依赖关系及其满足状态。

### Q: 如何修改已存在的任务？
A: 目前需要先删除任务（`clean`），然后重新添加。或者直接编辑 `~/.claude-swarm/tasks.json` 文件。

### Q: 任务被阻塞是什么意思？
A: 任务被阻塞表示它的依赖任务还未完成。在 `swarm status` 输出中会显示 "⚠️ 未满足"。

### Q: 如何重试失败的任务？
A: 目前需要清理失败任务（`swarm clean --failed`）然后重新添加。

### Q: 可以同时运行多个 swarm 实例吗？
A: 可以，但它们应该使用不同的任务队列文件（通过 `--queue` 参数指定）。

---

## 快捷参考

| 命令 | 用途 | 常用选项 |
|------|------|----------|
| `add-task` | 添加单个任务 | `-p`, `-d`, `--max-retries`, `--id` |
| `batch-add` | 批量添加任务 | `-f`, `--stdin`, `-i` |
| `status` | 查看队列状态 | `-v`, `-f` |
| `clean` | 清理任务 | `--completed`, `--failed`, `--all`, `-f` |
| `orchestrate` | AI 分析需求 | `--auto-start`, `--auto-approve`, `-n` |
| `start` | 启动 Agent | `-n`, `-t` |
| `monitor` | 监控面板 | 无 |
