#!/bin/bash

echo "======================================"
echo "🔥 测试冲突场景 - 主控脑处理"
echo "======================================"
echo

PROJECT_DIR="/Users/ring/Documents/公司源码/ringsite/claude-swarm"
cd "$PROJECT_DIR" || exit 1

# 清理环境
echo "🧹 清理环境..."
./swarm stop 2>/dev/null
rm -f conflict-test.txt

# 启动 swarm
echo "🚀 启动 2 个 agents..."
./swarm start --agents 2 > /tmp/swarm-conflict-test.log 2>&1 &
sleep 5

# 添加两个会产生冲突的任务
echo "📋 添加冲突任务..."
echo ""
echo "任务 1: 创建 conflict-test.txt，内容为 'Version from Agent 0'"
./swarm add-task "创建一个名为 conflict-test.txt 的文件，内容为 'Version from Agent 0'"
sleep 2

echo "任务 2: 创建 conflict-test.txt，内容为 'Version from Agent 1'"
./swarm add-task "创建一个名为 conflict-test.txt 的文件，内容为 'Version from Agent 1'"
sleep 2

echo ""
echo "⏳ 等待 20 秒让任务执行并观察冲突..."
sleep 20

echo ""
echo "📊 查看日志中的冲突处理..."
echo "----------------------------------------"
grep -E "冲突|conflict|merge|合并" /tmp/swarm-conflict-test.log | tail -20
echo "----------------------------------------"

echo ""
echo "📁 检查 conflict-test.txt 文件..."
if [ -f conflict-test.txt ]; then
    echo "✅ 文件存在，内容："
    cat conflict-test.txt
else
    echo "❌ 文件不存在"
fi

echo ""
echo "🛑 停止测试..."
./swarm stop

echo ""
echo "======================================"
echo "测试完成"
echo "======================================"
echo "完整日志: /tmp/swarm-conflict-test.log"
