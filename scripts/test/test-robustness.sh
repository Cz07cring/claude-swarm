#!/bin/bash

# Claude Swarm V2 鲁棒性测试
# 测试长时间运行、错误恢复、并发等场景

set -e

echo "🧪 Claude Swarm V2 鲁棒性测试"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 配置
NUM_AGENTS="${1:-3}"
TEST_DURATION="${2:-300}"  # 默认5分钟
TASKS_FILE="$HOME/.claude-swarm/tasks-robustness.json"
LOG_FILE="/tmp/swarm-robustness-$(date +%Y%m%d-%H%M%S).log"

echo "📋 测试配置:"
echo "  - Agents: $NUM_AGENTS"
echo "  - Duration: ${TEST_DURATION}s"
echo "  - Log: $LOG_FILE"
echo ""

# 创建测试任务
echo "📝 创建测试任务..."
cat > "$TASKS_FILE" << 'EOF'
{
  "tasks": [
    {
      "id": "long-task-1",
      "description": "创建一个包含10个函数的 utils.go 文件，每个函数都有注释和测试用例",
      "status": "pending",
      "priority": 5,
      "retry_count": 0,
      "max_retries": 3
    },
    {
      "id": "quick-task-1",
      "description": "创建test1.txt，内容是当前时间",
      "status": "pending",
      "priority": 10,
      "retry_count": 0,
      "max_retries": 3
    },
    {
      "id": "quick-task-2",
      "description": "创建test2.txt，内容是hello",
      "status": "pending",
      "priority": 10,
      "retry_count": 0,
      "max_retries": 3
    },
    {
      "id": "medium-task-1",
      "description": "创建一个HTTP server的Go代码 server.go，监听8080端口",
      "status": "pending",
      "priority": 7,
      "retry_count": 0,
      "max_retries": 3
    },
    {
      "id": "dep-task-1",
      "description": "读取test1.txt和test2.txt的内容，创建summary.md总结",
      "status": "pending",
      "dependencies": ["quick-task-1", "quick-task-2"],
      "priority": 5,
      "retry_count": 0,
      "max_retries": 3
    }
  ]
}
EOF

echo "✓ 创建了5个测试任务（包含依赖关系）"
echo ""

# 清理旧环境
echo "🧹 清理旧环境..."
rm -rf .worktrees 2>/dev/null || true
git worktree prune 2>/dev/null || true
git branch -D agent-*-branch 2>/dev/null || true
echo "✓ 清理完成"
echo ""

# 启动 swarm
echo "🚀 启动 Swarm..."
./swarm start-v2 --agents "$NUM_AGENTS" --tasks "$TASKS_FILE" > "$LOG_FILE" 2>&1 &
SWARM_PID=$!
echo "✓ Swarm started (PID: $SWARM_PID)"
echo ""

# 监控函数
monitor_swarm() {
    local duration=$1
    local start_time=$(date +%s)
    local last_check=0

    echo "📊 开始监控 (${duration}s)..."
    echo ""

    while true; do
        current_time=$(date +%s)
        elapsed=$((current_time - start_time))

        if [ $elapsed -ge $duration ]; then
            echo ""
            echo "⏰ 测试时间到 (${duration}s)"
            break
        fi

        # 每10秒检查一次
        if [ $((elapsed - last_check)) -ge 10 ]; then
            last_check=$elapsed

            # 检查进程
            if ! ps -p $SWARM_PID > /dev/null; then
                echo "❌ Swarm进程意外退出！"
                break
            fi

            # 显示状态
            echo -ne "\r⏱️  运行时间: ${elapsed}s / ${duration}s  "

            # 统计任务状态
            if [ -f "$TASKS_FILE" ]; then
                completed=$(jq '[.tasks[] | select(.status=="completed")] | length' "$TASKS_FILE" 2>/dev/null || echo 0)
                in_progress=$(jq '[.tasks[] | select(.status=="in_progress")] | length' "$TASKS_FILE" 2>/dev/null || echo 0)
                pending=$(jq '[.tasks[] | select(.status=="pending")] | length' "$TASKS_FILE" 2>/dev/null || echo 0)
                failed=$(jq '[.tasks[] | select(.status=="failed")] | length' "$TASKS_FILE" 2>/dev/null || echo 0)

                echo -ne "| ✅ $completed | 🔄 $in_progress | ⏳ $pending | ❌ $failed"
            fi
        fi

        sleep 1
    done

    echo ""
}

# 开始监控
monitor_swarm $TEST_DURATION

# 停止 swarm
echo ""
echo "🛑 停止 Swarm..."
kill $SWARM_PID 2>/dev/null || true
sleep 2
kill -9 $SWARM_PID 2>/dev/null || true
echo "✓ Swarm已停止"
echo ""

# 生成报告
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 测试报告"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 任务完成情况
echo "📋 任务状态:"
jq -r '.tasks[] | "  \(.id): \(.status)"' "$TASKS_FILE" 2>/dev/null || echo "  无法读取任务状态"
echo ""

# 统计
completed=$(jq '[.tasks[] | select(.status=="completed")] | length' "$TASKS_FILE" 2>/dev/null || echo 0)
total=$(jq '.tasks | length' "$TASKS_FILE" 2>/dev/null || echo 0)
echo "📈 统计:"
echo "  - 完成: $completed / $total"
echo "  - 成功率: $((completed * 100 / total))%"
echo ""

# 日志分析
echo "📝 关键日志:"
grep -E "(✅|❌|⚠️)" "$LOG_FILE" | tail -20 || echo "  无日志"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📄 完整日志: $LOG_FILE"
echo "✓ 测试完成!"
