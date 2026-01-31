#!/bin/bash
# 启动测试蜂群

echo "========================================="
echo "🐝 启动 Claude Swarm 测试环境"
echo "========================================="
echo ""

# 清理旧的任务队列（可选）
read -p "是否清空现有任务队列? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -f ~/.claude-swarm/tasks.json
    echo "✓ 任务队列已清空"
fi

# 设置测试配置
export SWARM_NUM_AGENTS=3
export SWARM_SESSION_NAME="claude-swarm-test"

echo ""
echo "配置:"
echo "  代理数量: $SWARM_NUM_AGENTS"
echo "  会话名称: $SWARM_SESSION_NAME"
echo ""

# 启动蜂群
echo "启动蜂群..."
go run cmd/swarm/main.go --agents 3 --session claude-swarm-test

