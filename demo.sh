#!/bin/bash

# Claude Agent Swarm - 演示脚本
# 展示如何实时观看 AI 指挥蜂群

echo "🎬 Claude Agent Swarm 演示"
echo "================================"
echo ""

# 清理旧会话
tmux kill-session -t claude-swarm 2>/dev/null

echo "1️⃣  启动蜂群（3个 Agent）..."
./swarm start -n 3 > /tmp/demo-swarm.log 2>&1 &
SWARM_PID=$!
echo "   Swarm PID: $SWARM_PID"
sleep 5

echo ""
echo "2️⃣  添加测试任务..."
./swarm add-task "显示 Go 版本"
./swarm add-task "列出当前目录"
./swarm add-task "显示系统信息"
sleep 2

echo ""
echo "3️⃣  查看当前状态..."
./swarm status

echo ""
echo "================================"
echo "🎥 现在可以实时观看了！"
echo ""
echo "选择一个方式："
echo ""
echo "  方式 1: 附加到 tmux 查看所有窗格"
echo "    tmux attach -t claude-swarm"
echo ""
echo "  方式 2: 查看单个 Agent"
echo "    watch -n 1 'tmux capture-pane -t claude-swarm:0.0 -p -S -30'"
echo ""
echo "  方式 3: 查看协调器日志"
echo "    tail -f /tmp/demo-swarm.log"
echo ""
echo "  方式 4: 持续监控状态"
echo "    watch -n 2 ./swarm status"
echo ""
echo "================================"
echo ""
echo "💡 提示："
echo "  - 你会看到任务自动分配给 Agent"
echo "  - 你会看到 Agent 自动执行命令"
echo "  - 你会看到自动确认（如果需要）"
echo "  - 按 Ctrl+C 停止观看（不会停止蜂群）"
echo ""
echo "停止蜂群: ./swarm stop"
echo ""
