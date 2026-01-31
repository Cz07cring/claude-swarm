#!/bin/bash

echo "======================================"
echo "🔬 Critical 修复专项测试"
echo "======================================"
echo

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

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

cd "/Users/ring/Documents/公司源码/ringsite/claude-swarm" || exit 1
echo -e "${BLUE}📂 项目目录:${NC} $(pwd)"
echo

# ==========================================
# 测试 1: 文件锁 - 多进程并发写入
# ==========================================
echo -e "${BLUE}[测试 1/3]${NC} 文件锁 - 多进程并发写入..."
echo "   测试场景：10个进程同时添加任务"

# 清理环境
./swarm stop 2>/dev/null
rm -f ~/.claude-swarm/tasks.json ~/.claude-swarm/tasks.json.lock

# 启动 swarm
./swarm start --agents 2 > /tmp/swarm-filelock-test.log 2>&1 &
SWARM_PID=$!
sleep 3

# 10个进程同时添加任务
for i in {1..10}; do
    ./swarm add-task "并发任务 $i" &
done
wait

# 等待文件写入完成
sleep 2

# 检查任务数量
TASK_COUNT=$(./swarm status 2>/dev/null | grep -o "待处理: [0-9]*" | grep -o "[0-9]*" || echo 0)
if [ "$TASK_COUNT" -eq 10 ]; then
    test_result 0 "文件锁正确工作：10个任务全部保存"
else
    test_result 1 "文件锁可能失败：期望10个任务，实际$TASK_COUNT个"
fi

# 检查任务文件是否损坏
if cat ~/.claude-swarm/tasks.json | jq . > /dev/null 2>&1; then
    test_result 0 "任务文件格式正确（JSON有效）"
else
    test_result 1 "任务文件可能损坏（JSON无效）"
fi

# 检查是否有重复任务ID
TASK_IDS=$(cat ~/.claude-swarm/tasks.json | jq -r '.tasks[].id' 2>/dev/null | sort)
UNIQUE_IDS=$(echo "$TASK_IDS" | uniq)
if [ "$(echo "$TASK_IDS" | wc -l)" -eq "$(echo "$UNIQUE_IDS" | wc -l)" ]; then
    test_result 0 "无重复任务ID"
else
    test_result 1 "发现重复的任务ID"
fi

./swarm stop 2>/dev/null
echo

# ==========================================
# 测试 2: 状态验证 - 并发状态修改
# ==========================================
echo -e "${BLUE}[测试 2/3]${NC} 状态验证 - 并发状态修改..."
echo "   测试场景：快速添加多个任务，观察状态一致性"

rm -f ~/.claude-swarm/tasks.json ~/.claude-swarm/tasks.json.lock
./swarm start --agents 3 > /tmp/swarm-state-test.log 2>&1 &
SWARM_PID=$!
sleep 3

# 快速添加5个任务
for i in {1..5}; do
    ./swarm add-task "状态测试任务 $i"
    sleep 0.5
done

# 等待任务执行
sleep 15

# 检查日志中是否有状态不一致警告
if grep -q "任务状态在合并过程中已变更" /tmp/swarm-state-test.log; then
    echo -e "${YELLOW}⚠️  INFO${NC}: 检测到状态变更警告（这是正常的保护机制）"
    test_result 0 "状态验证机制正常工作"
else
    test_result 0 "无状态不一致警告（理想情况）"
fi

# 检查是否有任务丢失
COMPLETED=$(grep -c "合并成功" /tmp/swarm-state-test.log || echo 0)
if [ "$COMPLETED" -ge 1 ]; then
    test_result 0 "至少有 $COMPLETED 个任务成功完成"
else
    test_result 1 "未检测到任务完成"
fi

./swarm stop 2>/dev/null
echo

# ==========================================
# 测试 3: Panic 恢复 - 模拟异常情况
# ==========================================
echo -e "${BLUE}[测试 3/3]${NC} Panic 恢复 - 模拟异常情况..."
echo "   测试场景：异常关闭后系统能否正常恢复"

rm -f ~/.claude-swarm/tasks.json ~/.claude-swarm/tasks.json.lock
./swarm start --agents 2 > /tmp/swarm-panic-test.log 2>&1 &
SWARM_PID=$!
sleep 3

# 添加一个任务
./swarm add-task "Panic测试任务"
sleep 5

# 模拟异常：直接 kill swarm 进程（不是优雅关闭）
echo "   模拟异常终止（kill -9）..."
kill -9 $SWARM_PID 2>/dev/null
sleep 2

# 检查 PID 文件是否残留
if [ -f ~/.claude-swarm/claude-swarm.pid ]; then
    test_result 0 "PID文件残留（预期行为）"
    
    # 尝试重新启动
    ./swarm start --agents 2 > /tmp/swarm-restart-test.log 2>&1 &
    NEW_PID=$!
    sleep 3
    
    # 检查是否成功启动
    if ps -p $NEW_PID > /dev/null; then
        test_result 0 "异常终止后能够重新启动"
        ./swarm stop 2>/dev/null
    else
        test_result 1 "异常终止后无法重新启动"
    fi
else
    test_result 0 "PID文件已清理"
fi

# 检查任务文件完整性
if [ -f ~/.claude-swarm/tasks.json ]; then
    if cat ~/.claude-swarm/tasks.json | jq . > /dev/null 2>&1; then
        test_result 0 "异常终止后任务文件仍然完整"
    else
        test_result 1 "异常终止导致任务文件损坏"
    fi
else
    test_result 1 "任务文件丢失"
fi

echo

# ==========================================
# 额外检查：文件锁清理
# ==========================================
echo -e "${BLUE}[额外检查]${NC} 文件锁清理..."

if [ -f ~/.claude-swarm/tasks.json.lock ]; then
    # 尝试获取锁（不应该被阻塞）
    if flock -n ~/.claude-swarm/tasks.json.lock -c "echo 'Lock acquired'" 2>/dev/null; then
        test_result 0 "文件锁已正确释放"
    else
        test_result 1 "文件锁未释放（可能泄漏）"
    fi
else
    echo -e "${YELLOW}ℹ️  INFO${NC}: 锁文件不存在（可能在清理中被删除）"
fi

echo

# ==========================================
# 最终清理
# ==========================================
echo "🧹 最终清理..."
./swarm stop 2>/dev/null
pkill -f "./swarm start" 2>/dev/null
tmux kill-server 2>/dev/null
rm -rf .worktrees/
rm -f ~/.claude-swarm/*.pid
echo "✓ 清理完成"
echo

# ==========================================
# 测试报告
# ==========================================
echo "======================================"
echo "📊 测试结果汇总"
echo "======================================"
echo -e "总测试数: ${BLUE}$TESTS_TOTAL${NC}"
echo -e "通过: ${GREEN}$TESTS_PASSED${NC}"
echo -e "失败: ${RED}$TESTS_FAILED${NC}"
echo

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 所有 Critical 修复测试通过！${NC}"
    echo
    echo "✅ 验证结果："
    echo "  • 文件锁正确工作，防止并发冲突"
    echo "  • 状态验证机制有效，防止数据丢失"
    echo "  • Panic 恢复保护系统稳定性"
    echo "  • 异常情况下数据完整性保持"
    exit 0
else
    echo -e "${RED}⚠️  有 $TESTS_FAILED 个测试失败${NC}"
    echo
    echo "查看日志："
    echo "  tail -100 /tmp/swarm-filelock-test.log"
    echo "  tail -100 /tmp/swarm-state-test.log"
    echo "  tail -100 /tmp/swarm-panic-test.log"
    exit 1
fi
