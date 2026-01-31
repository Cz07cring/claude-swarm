#!/bin/bash
# 蜂群鲁棒性测试运行器

set -e

SCENARIO=$1

if [ -z "$SCENARIO" ]; then
    echo "用法: ./run_test.sh <scenario>"
    echo ""
    echo "可用场景:"
    echo "  1 - 基础功能测试"
    echo "  2 - DAG依赖测试"
    echo "  3 - 错误恢复测试"
    echo "  4 - 并发压力测试"
    echo "  5 - 边界条件测试"
    exit 1
fi

SCENARIO_FILE="test${SCENARIO}_*.json"

echo "========================================="
echo "🐝 Claude Swarm 鲁棒性测试"
echo "========================================="
echo ""
echo "测试场景: $(ls test${SCENARIO}_*.json 2>/dev/null | head -1)"
echo ""

# 检查蜂群是否运行
if ! tmux has-session -t claude-swarm 2>/dev/null; then
    echo "❌ 蜂群未运行"
    echo "请先启动蜂群: cd ../../ && go run cmd/swarm/main.go"
    exit 1
fi

echo "✓ 检测到蜂群运行中"
echo ""

# 导入测试任务
echo "导入测试任务..."
cd ../../
go run test/swarm-robustness/import_tasks.go test/swarm-robustness/test${SCENARIO}_*.json

echo ""
echo "========================================="
echo "✅ 任务已导入到蜂群队列"
echo "========================================="
echo ""
echo "监控命令:"
echo "  tmux attach -t claude-swarm  # 查看蜂群界面"
echo "  tail -f ~/.claude-swarm/tasks.json  # 监控任务队列"
echo ""
