#!/bin/bash

# P0 修复验证测试
# 测试 P0-1, P0-2, P0-3 的修复

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "========================================"
echo "🧪 P0 修复验证测试"
echo "========================================"
echo ""

# 测试计数器
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# 辅助函数
pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((TESTS_PASSED++))
    ((TESTS_RUN++))
}

fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    echo -e "   ${YELLOW}原因: $2${NC}"
    ((TESTS_FAILED++))
    ((TESTS_RUN++))
}

info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

section() {
    echo ""
    echo "========================================"
    echo "📋 $1"
    echo "========================================"
}

cleanup() {
    info "清理测试环境..."
    # 停止任何运行中的 swarm
    ./swarm stop 2>/dev/null || true
    # 等待清理完成
    sleep 2
    # 清理测试文件
    rm -rf .worktrees 2>/dev/null || true
    rm -f ~/.claude-swarm/*.pid 2>/dev/null || true
}

# 预清理
cleanup

# ========================================
# 测试 1: 编译和基础检查
# ========================================
section "测试 1: 编译和基础检查"

info "1.1 检查编译..."
if go build -o swarm ./cmd/swarm 2>/dev/null; then
    pass "编译成功"
else
    fail "编译失败" "无法编译 swarm"
    exit 1
fi

info "1.2 检查二进制文件..."
if [ -f "./swarm" ]; then
    SIZE=$(ls -lh swarm | awk '{print $5}')
    pass "二进制文件存在 (大小: $SIZE)"
else
    fail "二进制文件不存在" "编译后没有生成 swarm"
    exit 1
fi

info "1.3 检查版本信息..."
if ./swarm --help >/dev/null 2>&1; then
    pass "Help 命令正常"
else
    fail "Help 命令失败" "swarm --help 返回错误"
fi

# ========================================
# 测试 2: P0-2 - Worktrees 清理测试
# ========================================
section "测试 2: P0-2 - Worktrees 清理"

info "2.1 清理遗留 worktrees..."
cleanup

info "2.2 启动 swarm (2 agents)..."
timeout 10 ./swarm start --agents 2 >/dev/null 2>&1 &
SWARM_PID=$!
sleep 5

info "2.3 检查 worktrees 是否创建..."
if [ -d ".worktrees" ]; then
    WORKTREE_COUNT=$(ls -1 .worktrees | wc -l)
    pass "Worktrees 已创建 ($WORKTREE_COUNT 个)"
else
    fail "Worktrees 未创建" ".worktrees 目录不存在"
fi

info "2.4 停止 swarm..."
./swarm stop 2>&1 | grep -q "已停止" || true
sleep 2

info "2.5 检查 worktrees 是否完全清理..."
if [ -d ".worktrees" ]; then
    # 检查是否有残留文件
    if [ -z "$(ls -A .worktrees 2>/dev/null)" ]; then
        # 目录存在但是空的
        fail "Worktrees 目录残留" ".worktrees 目录存在但为空（应该被删除）"
    else
        # 目录存在且有文件
        REMAINING=$(ls -A .worktrees | wc -l)
        fail "Worktrees 清理不完整" "残留 $REMAINING 个文件/目录"
        ls -la .worktrees
    fi
else
    pass "Worktrees 完全清理（目录已删除）"
fi

# ========================================
# 测试 3: P0-3 - 进程清理测试
# ========================================
section "测试 3: P0-3 - 进程清理"

cleanup

info "3.1 启动 swarm (2 agents)..."
timeout 10 ./swarm start --agents 2 >/dev/null 2>&1 &
SWARM_PID=$!
sleep 5

info "3.2 检查 PID 文件是否创建..."
PID_FILE="$HOME/.claude-swarm/claude-swarm.pid"
if [ -f "$PID_FILE" ]; then
    SAVED_PID=$(cat "$PID_FILE")
    pass "PID 文件已创建 (PID: $SAVED_PID)"
else
    fail "PID 文件未创建" "$PID_FILE 不存在"
fi

info "3.3 检查 swarm 进程是否运行..."
RUNNING_PROCESSES=$(pgrep -f "swarm start" | wc -l)
if [ "$RUNNING_PROCESSES" -gt 0 ]; then
    pass "Swarm 进程正在运行 ($RUNNING_PROCESSES 个)"
else
    fail "Swarm 进程未运行" "找不到 swarm start 进程"
fi

info "3.4 停止 swarm..."
./swarm stop 2>&1 | grep -q "已停止" || true
sleep 3

info "3.5 检查进程是否完全清理..."
REMAINING_PROCESSES=$(pgrep -f "swarm start" | wc -l)
if [ "$REMAINING_PROCESSES" -eq 0 ]; then
    pass "所有 swarm 进程已清理"
else
    fail "进程清理不完整" "仍有 $REMAINING_PROCESSES 个进程运行"
    pgrep -f "swarm start"
fi

info "3.6 检查 PID 文件是否删除..."
if [ -f "$PID_FILE" ]; then
    fail "PID 文件未删除" "$PID_FILE 仍然存在"
else
    pass "PID 文件已删除"
fi

# ========================================
# 测试 4: P0-1 - tmux 会话异常终止检测
# ========================================
section "测试 4: P0-1 - tmux 会话异常终止检测"

cleanup

info "4.1 启动 swarm (2 agents)..."
timeout 15 ./swarm start --agents 2 >/dev/null 2>&1 &
SWARM_PID=$!
sleep 5

info "4.2 检查 tmux 会话是否创建..."
if tmux has-session -t claude-swarm 2>/dev/null; then
    pass "tmux 会话已创建"
else
    fail "tmux 会话未创建" "找不到 claude-swarm 会话"
    cleanup
fi

info "4.3 手动终止 tmux 会话（模拟异常终止）..."
tmux kill-session -t claude-swarm 2>/dev/null || true
sleep 2

info "4.4 等待 coordinator 检测并退出（最多 20 秒）..."
WAIT_COUNT=0
MAX_WAIT=20
COORDINATOR_EXITED=false

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if ! kill -0 $SWARM_PID 2>/dev/null; then
        COORDINATOR_EXITED=true
        break
    fi
    sleep 1
    ((WAIT_COUNT++))
done

if [ "$COORDINATOR_EXITED" = true ]; then
    pass "Coordinator 在 ${WAIT_COUNT}s 内检测到会话终止并退出"
else
    fail "Coordinator 未能检测到会话终止" "等待 ${MAX_WAIT}s 后进程仍在运行"
    # 强制终止
    kill -9 $SWARM_PID 2>/dev/null || true
fi

# ========================================
# 测试 5: 压力测试 - 快速启停
# ========================================
section "测试 5: 压力测试 - 快速启停"

cleanup

info "5.1 快速启停测试（3 次）..."
for i in {1..3}; do
    info "   第 $i 次启停..."
    timeout 10 ./swarm start --agents 2 >/dev/null 2>&1 &
    sleep 3
    ./swarm stop >/dev/null 2>&1
    sleep 2

    # 检查是否有残留
    REMAINING_PROCESSES=$(pgrep -f "swarm start" | wc -l)
    if [ "$REMAINING_PROCESSES" -eq 0 ] && [ ! -d ".worktrees" ]; then
        info "   ✓ 第 $i 次清理完成"
    else
        fail "快速启停测试" "第 $i 次清理不完整（进程: $REMAINING_PROCESSES, worktrees: $([ -d .worktrees ] && echo '存在' || echo '无')）"
        cleanup
        break
    fi
done

if [ "$TESTS_FAILED" -eq 0 ] || [ $i -eq 3 ]; then
    pass "快速启停测试 (3 次全部通过)"
fi

# ========================================
# 最终清理
# ========================================
cleanup

# ========================================
# 测试总结
# ========================================
echo ""
echo "========================================"
echo "📊 测试总结"
echo "========================================"
echo "总测试数: $TESTS_RUN"
echo -e "${GREEN}通过: $TESTS_PASSED${NC}"
echo -e "${RED}失败: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ 所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}❌ 有 $TESTS_FAILED 个测试失败${NC}"
    exit 1
fi
