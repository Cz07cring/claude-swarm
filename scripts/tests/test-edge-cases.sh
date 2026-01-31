#!/bin/bash
# 测试 TUI 边界情况

echo "🔍 测试 TUI 边界情况..."
echo ""

# 测试 1: 空数据情况
echo "1. 测试空 Agent 列表显示"
EMPTY_AGENTS='{"agents":[],"updated_at":"2026-01-31T23:00:00+08:00"}'
echo "$EMPTY_AGENTS" > /tmp/test-agents-empty.json

# 测试 2: 大量数据情况
echo "2. 测试大量 Agent（50个）"
cat > /tmp/test-agents-many.json << 'AGENTS'
{
  "agents": [
AGENTS

for i in {1..50}; do
    if [ $i -lt 50 ]; then
        echo "    {\"agent_id\":\"agent-$i\",\"state\":\"idle\",\"last_update\":\"2026-01-31T23:00:00+08:00\"}," >> /tmp/test-agents-many.json
    else
        echo "    {\"agent_id\":\"agent-$i\",\"state\":\"idle\",\"last_update\":\"2026-01-31T23:00:00+08:00\"}" >> /tmp/test-agents-many.json
    fi
done

echo "  ]," >> /tmp/test-agents-many.json
echo '  "updated_at":"2026-01-31T23:00:00+08:00"' >> /tmp/test-agents-many.json
echo "}" >> /tmp/test-agents-many.json

# 测试 3: 超长描述
echo "3. 测试超长任务描述"
LONG_DESC="这是一个非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常长的任务描述，用来测试 TUI 是否能够正确处理和显示超长文本内容而不会导致布局混乱或者崩溃"

cat > /tmp/test-tasks-long.json << TASKS
{
  "tasks": [
    {
      "id": "task-1",
      "description": "$LONG_DESC",
      "status": "pending",
      "created_at": "2026-01-31T23:00:00+08:00",
      "updated_at": "2026-01-31T23:00:00+08:00"
    }
  ]
}
TASKS

# 测试 4: 超长日志输出
echo "4. 测试超长 Agent 输出"
LONG_OUTPUT=""
for i in {1..1000}; do
    LONG_OUTPUT="${LONG_OUTPUT}Log line $i: This is a very long log message to test the scrolling functionality\\n"
done

cat > /tmp/test-agents-longlog.json << AGENTLOG
{
  "agents": [
    {
      "agent_id": "agent-1",
      "state": "working",
      "last_update": "2026-01-31T23:00:00+08:00",
      "output": "$LONG_OUTPUT"
    }
  ],
  "updated_at": "2026-01-31T23:00:00+08:00"
}
AGENTLOG

echo ""
echo "✅ 测试数据文件已创建:"
echo "   - /tmp/test-agents-empty.json (空列表)"
echo "   - /tmp/test-agents-many.json (50 个 Agent)"
echo "   - /tmp/test-tasks-long.json (超长描述)"
echo "   - /tmp/test-agents-longlog.json (超长日志)"
echo ""
echo "💡 可以使用以下命令测试:"
echo "   ./bin/swarm monitor --state /tmp/test-agents-empty.json"
echo "   ./bin/swarm monitor --state /tmp/test-agents-many.json"
