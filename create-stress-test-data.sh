#!/bin/bash

# 压力测试数据生成脚本
# 创建大量任务和 Agent 用于测试 TUI 性能

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

NUM_TASKS=${1:-100}
NUM_AGENTS=${2:-20}

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  TUI 压力测试数据生成${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "生成参数:"
echo "  任务数量: $NUM_TASKS"
echo "  Agent 数量: $NUM_AGENTS"
echo ""

mkdir -p ~/.claude-swarm

# 生成任务数据
echo -n "生成任务数据..."
cat > ~/.claude-swarm/tasks.json << 'HEADER'
{
  "tasks": [
HEADER

STATUSES=("pending" "in_progress" "completed" "failed")
TASK_TEMPLATES=(
  "实现用户认证系统"
  "创建数据库架构"
  "编写单元测试"
  "实现 REST API"
  "添加日志功能"
  "优化数据库查询"
  "修复内存泄漏"
  "添加缓存层"
  "实现搜索功能"
  "升级依赖包"
  "重构代码结构"
  "添加监控告警"
  "实现消息队列"
  "添加文档注释"
  "性能优化"
  "安全加固"
  "实现导出功能"
  "添加国际化"
  "修复并发问题"
  "实现定时任务"
)

for i in $(seq 0 $((NUM_TASKS - 1))); do
    TASK_ID=$(printf "task-%03d" $i)
    STATUS_INDEX=$((i % 4))
    STATUS=${STATUSES[$STATUS_INDEX]}
    TEMPLATE_INDEX=$((i % ${#TASK_TEMPLATES[@]}))
    DESC="${TASK_TEMPLATES[$TEMPLATE_INDEX]} #$i"

    # 随机分配 Agent (40% 概率)
    ASSIGNEE=""
    if [ $((RANDOM % 100)) -lt 40 ] && [ $NUM_AGENTS -gt 0 ]; then
        AGENT_NUM=$((RANDOM % NUM_AGENTS))
        ASSIGNEE="agent-$AGENT_NUM"
    fi

    # 生成时间戳 (过去 24 小时内随机)
    HOURS_AGO=$((RANDOM % 24))
    MINS_AGO=$((RANDOM % 60))
    CREATED=$(date -u -v-${HOURS_AGO}H -v-${MINS_AGO}M +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || \
              date -u -d "$HOURS_AGO hours ago $MINS_AGO minutes ago" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || \
              echo "2026-01-30T10:00:00Z")

    UPDATED=$CREATED
    if [ -n "$ASSIGNEE" ]; then
        UPDATE_MINS_AGO=$((RANDOM % MINS_AGO))
        UPDATED=$(date -u -v-${HOURS_AGO}H -v-${UPDATE_MINS_AGO}M +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || \
                  date -u -d "$HOURS_AGO hours ago $UPDATE_MINS_AGO minutes ago" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || \
                  echo "2026-01-30T10:05:00Z")
    fi

    # 输出 JSON
    cat >> ~/.claude-swarm/tasks.json << EOF
    {
      "id": "$TASK_ID",
      "description": "$DESC",
      "status": "$STATUS",
EOF

    if [ -n "$ASSIGNEE" ]; then
        echo "      \"assignee_id\": \"$ASSIGNEE\"," >> ~/.claude-swarm/tasks.json
    fi

    cat >> ~/.claude-swarm/tasks.json << EOF
      "created_at": "$CREATED",
      "updated_at": "$UPDATED"
    }
EOF

    if [ $i -lt $((NUM_TASKS - 1)) ]; then
        echo "," >> ~/.claude-swarm/tasks.json
    fi
done

cat >> ~/.claude-swarm/tasks.json << 'FOOTER'
  ]
}
FOOTER

echo -e " ${GREEN}✓${NC}"

# 生成 Agent 数据
echo -n "生成 Agent 数据..."
cat > ~/.claude-swarm/agents.json << 'HEADER'
{
  "agents": [
HEADER

AGENT_STATES=("idle" "working" "error" "waiting_confirm" "idle" "working" "idle")
OUTPUT_TEMPLATES=(
  "等待新任务..."
  "正在执行任务...\n进度: 50%"
  "错误: 依赖包冲突\nError: Package conflict detected"
  "准备提交更改\n是否继续? (y/n)"
  "任务完成，空闲中"
  "正在编译代码...\nBuild in progress..."
  "测试通过: 100%\nAll tests passed"
)

for i in $(seq 0 $((NUM_AGENTS - 1))); do
    AGENT_ID="agent-$i"
    STATE_INDEX=$((i % ${#AGENT_STATES[@]}))
    STATE=${AGENT_STATES[$STATE_INDEX]}

    OUTPUT_INDEX=$((i % ${#OUTPUT_TEMPLATES[@]}))
    OUTPUT="${OUTPUT_TEMPLATES[$OUTPUT_INDEX]}"

    # 为工作中的 Agent 分配任务
    TASK_JSON=""
    if [ "$STATE" == "working" ] || [ "$STATE" == "error" ]; then
        TASK_NUM=$((RANDOM % NUM_TASKS))
        TASK_ID=$(printf "task-%03d" $TASK_NUM)
        TASK_STATUS="in_progress"
        if [ "$STATE" == "error" ]; then
            TASK_STATUS="failed"
        fi

        TASK_JSON=$(cat << TASKEOF
      "current_task": {
        "id": "$TASK_ID",
        "description": "关联任务 #$TASK_NUM",
        "status": "$TASK_STATUS"
      },
TASKEOF
)
    fi

    # 生成更新时间
    MINS_AGO=$((RANDOM % 30))
    LAST_UPDATE=$(date -u -v-${MINS_AGO}M +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || \
                  date -u -d "$MINS_AGO minutes ago" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || \
                  echo "2026-01-30T10:00:00Z")

    # 输出 JSON
    cat >> ~/.claude-swarm/agents.json << EOF
    {
      "agent_id": "$AGENT_ID",
      "state": "$STATE",
$TASK_JSON
      "last_update": "$LAST_UPDATE",
      "output": "$OUTPUT"
    }
EOF

    if [ $i -lt $((NUM_AGENTS - 1)) ]; then
        echo "," >> ~/.claude-swarm/agents.json
    fi
done

cat >> ~/.claude-swarm/agents.json << FOOTER
  ],
  "updated_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
FOOTER

echo -e " ${GREEN}✓${NC}"

echo ""
echo -e "${GREEN}压力测试数据生成完成！${NC}"
echo ""
echo "统计信息:"
echo "  ✓ 创建 $NUM_TASKS 个任务"
echo "  ✓ 创建 $NUM_AGENTS 个 Agent"
echo "  ✓ 数据文件大小:"
du -h ~/.claude-swarm/tasks.json ~/.claude-swarm/agents.json

echo ""
echo "现在可以运行压力测试:"
echo "  ./swarm monitor"
echo ""
echo "测试要点:"
echo "  1. 观察导航流畅度"
echo "  2. 测试滚动性能"
echo "  3. 监控内存使用"
echo "  4. 检查响应速度"
echo ""
