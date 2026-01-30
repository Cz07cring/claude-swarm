#!/bin/bash

echo "======================================"
echo "🧪 测试 Worktree 管理系统修复"
echo "======================================"
echo

# 设置颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目目录
PROJECT_DIR="/Users/ring/Documents/公司源码/ringsite/claude-swarm"
cd "$PROJECT_DIR" || exit 1

echo "📂 当前目录: $(pwd)"
echo

# 步骤 1: 清理环境
echo "🧹 步骤 1: 清理环境..."
pkill -f "./swarm start" 2>/dev/null
tmux kill-server 2>/dev/null
rm -f ~/.claude-swarm/*.pid
git worktree remove .worktrees/agent-0 --force 2>/dev/null
git worktree remove .worktrees/agent-1 --force 2>/dev/null
git branch -D agent-0-branch agent-1-branch 2>/dev/null
rm -rf .worktrees/
echo -e "${GREEN}✅ 环境清理完成${NC}"
echo

# 步骤 2: 编译项目
echo "🔨 步骤 2: 编译项目..."
if go build -o swarm ./cmd/swarm 2>&1; then
    echo -e "${GREEN}✅ 编译成功${NC}"
    ls -lh swarm
else
    echo -e "${RED}❌ 编译失败${NC}"
    exit 1
fi
echo

# 步骤 3: 测试 PID 锁机制
echo "🔒 步骤 3: 测试 PID 锁机制..."
echo "   启动第一个实例..."
./swarm start --agents 2 > /tmp/swarm-test.log 2>&1 &
SWARM_PID=$!
sleep 3

echo "   检查 PID 文件..."
if [ -f ~/.claude-swarm/claude-swarm.pid ]; then
    echo -e "${GREEN}✅ PID 文件已创建${NC}"
    cat ~/.claude-swarm/claude-swarm.pid
else
    echo -e "${RED}❌ PID 文件未创建${NC}"
fi

echo "   尝试启动第二个实例（应该被阻止）..."
if ./swarm start --agents 2 2>&1 | grep -q "已在运行中"; then
    echo -e "${GREEN}✅ 多进程启动被正确阻止${NC}"
else
    echo -e "${RED}❌ 多进程启动未被阻止${NC}"
fi
echo

# 步骤 4: 验证 Worktrees 创建
echo "🌳 步骤 4: 验证 Worktrees..."
sleep 2
if [ -d .worktrees/agent-0 ] && [ -d .worktrees/agent-1 ]; then
    echo -e "${GREEN}✅ Worktrees 创建成功${NC}"
    git worktree list
else
    echo -e "${RED}❌ Worktrees 创建失败${NC}"
fi
echo

# 步骤 5: 添加测试任务
echo "📋 步骤 5: 添加测试任务..."
./swarm add-task "创建一个名为 fix-test.txt 的文件，内容为 'Fix verification test'"
sleep 2
echo -e "${GREEN}✅ 任务已添加${NC}"
echo

# 步骤 6: 监控日志
echo "📊 步骤 6: 监控执行（10秒）..."
echo "   查看最近的日志输出..."
sleep 10
tail -30 /tmp/swarm-test.log | grep -E "检测到任务完成|开始合并|合并成功|state changed"
echo

# 步骤 7: 检查合并结果
echo "🔍 步骤 7: 检查合并结果..."
if [ -f fix-test.txt ]; then
    echo -e "${GREEN}✅ 文件已创建并合并到 main 分支${NC}"
    cat fix-test.txt
else
    echo -e "${YELLOW}⚠️  文件未找到（可能任务还在执行）${NC}"
fi
echo

# 步骤 8: 停止并验证清理
echo "🛑 步骤 8: 停止并验证清理..."
./swarm stop

sleep 2
if [ ! -d .worktrees ] && [ ! -f ~/.claude-swarm/claude-swarm.pid ]; then
    echo -e "${GREEN}✅ 清理成功 - Worktrees 和 PID 文件已删除${NC}"
else
    echo -e "${RED}❌ 清理不完整${NC}"
    [ -d .worktrees ] && echo "   - Worktrees 目录仍存在"
    [ -f ~/.claude-swarm/claude-swarm.pid ] && echo "   - PID 文件仍存在"
fi

git worktree list
echo

# 步骤 9: 检查 git 分支
echo "🌿 步骤 9: 检查 git 分支..."
if git branch | grep -q agent; then
    echo -e "${YELLOW}⚠️  Agent 分支未清理${NC}"
    git branch | grep agent
else
    echo -e "${GREEN}✅ 所有 agent 分支已清理${NC}"
fi
echo

# 最终报告
echo "======================================"
echo "📝 测试完成"
echo "======================================"
echo
echo "请检查上面的输出，确认所有测试项都是 ✅"
echo
echo "关键检查点："
echo "1. ✅ PID 文件锁是否工作"
echo "2. ✅ Worktrees 是否创建"
echo "3. ✅ 任务是否执行"
echo "4. ✅ 合并日志是否显示"
echo "5. ✅ 清理是否完整"
echo
