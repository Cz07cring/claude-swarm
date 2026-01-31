#!/bin/bash
# 重置所有 in_progress 任务为 pending 状态

TASK_FILE="$HOME/.claude-swarm/tasks.json"

if [ ! -f "$TASK_FILE" ]; then
    echo "❌ 任务队列文件不存在: $TASK_FILE"
    exit 1
fi

echo "🔄 重置僵尸任务..."

# 使用 jq 重置所有 in_progress 任务
jq '.tasks |= map(
    if .status == "in_progress" then
        .status = "pending" |
        .assignee_id = "" |
        .updated_at = (now | todate)
    else
        .
    end
)' "$TASK_FILE" > "$TASK_FILE.tmp"

mv "$TASK_FILE.tmp" "$TASK_FILE"

echo "✅ 重置完成！"
echo ""
./swarm status
