#!/bin/bash
# 快速添加测试任务

TASKS_FILE="$HOME/.claude-swarm/tasks.json"

# 确保目录存在
mkdir -p ~/.claude-swarm

# 场景 1: 基础测试（3个简单任务）
cat > "$TASKS_FILE" << 'TASKS'
{
  "tasks": [
    {
      "id": "test-1",
      "description": "在 test/swarm-robustness/results/ 目录下创建一个 hello.go 文件，包含 main 函数输出 'Hello from Swarm!'",
      "status": "pending",
      "priority": 5,
      "max_retries": 3,
      "retry_count": 0,
      "dependencies": [],
      "created_at": "2026-01-31T23:55:00Z",
      "updated_at": "2026-01-31T23:55:00Z"
    },
    {
      "id": "test-2",
      "description": "创建一个 calculator.go 文件，实现加减乘除四个函数",
      "status": "pending",
      "priority": 5,
      "max_retries": 3,
      "retry_count": 0,
      "dependencies": [],
      "created_at": "2026-01-31T23:55:01Z",
      "updated_at": "2026-01-31T23:55:01Z"
    },
    {
      "id": "test-3",
      "description": "创建 test-summary.md 文件，总结前面创建的两个文件的功能",
      "status": "pending",
      "priority": 3,
      "max_retries": 3,
      "retry_count": 0,
      "dependencies": ["test-1", "test-2"],
      "created_at": "2026-01-31T23:55:02Z",
      "updated_at": "2026-01-31T23:55:02Z"
    }
  ]
}
TASKS

echo "✅ 测试任务已添加到: $TASKS_FILE"
echo ""
cat "$TASKS_FILE" | python3 -m json.tool 2>/dev/null || cat "$TASKS_FILE"
echo ""
echo "任务说明:"
echo "  test-1: 创建 hello.go (独立任务)"
echo "  test-2: 创建 calculator.go (独立任务)"
echo "  test-3: 创建总结文档 (依赖 test-1 和 test-2)"
echo ""
echo "启动蜂群: go run cmd/swarm/main.go --agents 3"
