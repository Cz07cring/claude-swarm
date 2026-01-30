#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试结果统计
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# 测试结果函数
test_result() {
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ PASS${NC}: $2"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}❌ FAIL${NC}: $2"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        if [ ! -z "$3" ]; then
            echo -e "${YELLOW}   详情: $3${NC}"
        fi
    fi
}

echo "======================================"
echo "🧪 Claude Swarm - 完整测试套件"
echo "======================================"
echo

# 切换到项目目录
cd "/Users/ring/Documents/公司源码/ringsite/claude-swarm" || exit 1
echo -e "${BLUE}📂 项目目录:${NC} $(pwd)"
echo

# ==========================================
# 测试 1: 编译
# ==========================================
echo -e "${BLUE}[测试 1/8]${NC} 编译项目..."
if go build -o swarm ./cmd/swarm 2>&1; then
    test_result 0 "编译成功"
    ls -lh swarm | awk '{print "   大小: " $5}'
else
    test_result 1 "编译失败" "请检查 Go 环境和依赖"
    exit 1
fi
echo

# ==========================================
# 测试 2: 清理环境
# ==========================================
echo -e "${BLUE}[测试 2/8]${NC} 清理环境..."
pkill -f "./swarm start" 2>/dev/null
tmux kill-server 2>/dev/null
git worktree remove .worktrees/agent-0 --force 2>/dev/null
git worktree remove .worktrees/agent-1 --force 2>/dev/null
git branch -D agent-0-branch agent-1-branch 2>/dev/null
rm -rf .worktrees/
rm -f ~/.claude-swarm/*.pid
test_result 0 "环境清理完成"
echo

# ==========================================
# 测试 3: PID 锁机制
# ==========================================
echo -e "${BLUE}[测试 3/8]${NC} 测试 PID 锁机制..."

# 3.1 启动第一个实例
./swarm start --agents 2 > /tmp/swarm-test.log 2>&1 &
SWARM_PID=$!
sleep 3

# 3.2 检查 PID 文件
if [ -f ~/.claude-swarm/claude-swarm.pid ]; then
    PID_FROM_FILE=$(cat ~/.claude-swarm/claude-swarm.pid)
    test_result 0 "PID 文件已创建 (PID: $PID_FROM_FILE)"
else
    test_result 1 "PID 文件未创建"
fi

# 3.3 尝试启动第二个实例
if ./swarm start --agents 2 2>&1 | grep -q "已在运行中"; then
    test_result 0 "多进程启动被正确阻止"
else
    test_result 1 "多进程启动未被阻止"
fi
echo

# ==========================================
# 测试 4: Worktrees 创建
# ==========================================
echo -e "${BLUE}[测试 4/8]${NC} 验证 Worktrees 创建..."
sleep 2

if [ -d .worktrees/agent-0 ] && [ -d .worktrees/agent-1 ]; then
    test_result 0 "Worktrees 目录已创建"
    git worktree list | grep -E "agent-0|agent-1" | sed 's/^/   /'
else
    test_result 1 "Worktrees 目录未创建"
fi

if git branch | grep -q "agent-0-branch" && git branch | grep -q "agent-1-branch"; then
    test_result 0 "Agent 分支已创建"
else
    test_result 1 "Agent 分支未创建"
fi
echo

# ==========================================
# 测试 5: 添加任务
# ==========================================
echo -e "${BLUE}[测试 5/8]${NC} 测试任务添加..."
TASK_OUTPUT=$(./swarm add-task "创建一个名为 test-$(date +%s).txt 的文件，内容为 'Automated test'")
if echo "$TASK_OUTPUT" | grep -q "任务已添加"; then
    TASK_ID=$(echo "$TASK_OUTPUT" | grep "ID:" | awk '{print $2}')
    test_result 0 "任务添加成功 (ID: $TASK_ID)"
else
    test_result 1 "任务添加失败"
fi
echo

# ==========================================
# 测试 6: 等待并检查合并日志
# ==========================================
echo -e "${BLUE}[测试 6/8]${NC} 等待任务执行并检查合并日志..."
echo "   等待 15 秒让任务执行..."
sleep 15

# 检查日志中的关键信息
echo "   查找合并日志..."
if grep -q "检测到任务完成" /tmp/swarm-test.log; then
    test_result 0 "发现任务完成检测日志"
    grep "检测到任务完成" /tmp/swarm-test.log | tail -1 | sed 's/^/   /'
else
    test_result 1 "未发现任务完成检测日志"
fi

if grep -q "开始合并.*到 main 分支" /tmp/swarm-test.log; then
    test_result 0 "发现合并开始日志"
    grep "开始合并.*到 main 分支" /tmp/swarm-test.log | tail -1 | sed 's/^/   /'
else
    test_result 1 "未发现合并开始日志"
fi

if grep -q "合并成功" /tmp/swarm-test.log; then
    test_result 0 "发现合并成功日志"
    grep "合并成功" /tmp/swarm-test.log | tail -1 | sed 's/^/   /'
else
    test_result 1 "未发现合并成功日志"
fi
echo

# ==========================================
# 测试 7: 停止并验证清理
# ==========================================
echo -e "${BLUE}[测试 7/8]${NC} 测试停止和清理..."
./swarm stop

sleep 2

# 检查 worktrees 清理
if [ ! -d .worktrees ]; then
    test_result 0 "Worktrees 目录已删除"
else
    test_result 1 "Worktrees 目录未删除"
fi

# 检查分支清理
if ! git branch | grep -q "agent.*branch"; then
    test_result 0 "Agent 分支已删除"
else
    test_result 1 "Agent 分支未删除"
fi

# 检查 PID 文件清理
if [ ! -f ~/.claude-swarm/claude-swarm.pid ]; then
    test_result 0 "PID 文件已删除"
else
    test_result 1 "PID 文件未删除"
fi
echo

# ==========================================
# 测试 8: Git 状态验证
# ==========================================
echo -e "${BLUE}[测试 8/8]${NC} 验证 Git 状态..."
WORKTREE_COUNT=$(git worktree list | wc -l)
if [ "$WORKTREE_COUNT" -eq 1 ]; then
    test_result 0 "只剩主 worktree"
else
    test_result 1 "存在多个 worktrees" "当前数量: $WORKTREE_COUNT"
fi
echo

# ==========================================
# 最终报告
# ==========================================
echo "======================================"
echo "📊 测试结果汇总"
echo "======================================"
echo -e "总测试数: ${BLUE}$TESTS_TOTAL${NC}"
echo -e "通过: ${GREEN}$TESTS_PASSED${NC}"
echo -e "失败: ${RED}$TESTS_FAILED${NC}"
echo

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 所有测试通过！修复成功！${NC}"
    exit 0
else
    echo -e "${RED}⚠️  有 $TESTS_FAILED 个测试失败${NC}"
    echo
    echo "查看完整日志:"
    echo "  tail -100 /tmp/swarm-test.log"
    exit 1
fi
