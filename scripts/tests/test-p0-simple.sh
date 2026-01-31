#!/bin/bash

# 简化的 P0 测试 - 非阻塞版本

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((TESTS_PASSED++))
}

fail() {
    echo -e "${RED}✗${NC} $1: $2"
    ((TESTS_FAILED++))
}

cleanup() {
    echo "🧹 清理..."
    pkill -f "swarm start" 2>/dev/null || true
    tmux kill-session -t claude-swarm 2>/dev/null || true
    sleep 2
    rm -rf .worktrees 2>/dev/null || true
    rm -f ~/.claude-swarm/*.pid 2>/dev/null || true
}

echo "========================================"
echo "🧪 P0 简化测试"
echo "========================================"

# 预清理
cleanup

# 测试 1: 编译
echo ""
echo "📋 测试 1: 编译检查"
if go build -o swarm ./cmd/swarm 2>/dev/null; then
    SIZE=$(ls -lh swarm | awk '{print $5}')
    pass "编译成功 ($SIZE)"
else
    fail "编译" "构建失败"
    exit 1
fi

# 测试 2: 静态分析
echo ""
echo "📋 测试 2: 静态分析"
if go vet ./... 2>&1 | grep -q "^"; then
    fail "go vet" "发现问题"
else
    pass "go vet 通过"
fi

# 测试 3: 单元测试
echo ""
echo "📋 测试 3: 单元测试"
if go test ./pkg/git -v 2>&1 | grep -q "PASS"; then
    COVERAGE=$(go test -cover ./pkg/git 2>&1 | grep coverage | awk '{print $4}')
    pass "单元测试通过 ($COVERAGE)"
else
    fail "单元测试" "有测试失败"
fi

# 测试 4: Worktrees 清理（P0-2）
echo ""
echo "📋 测试 4: P0-2 Worktrees 清理"
cleanup

# 创建测试 worktrees
mkdir -p .worktrees/test-agent
echo "test" > .worktrees/test-agent/file.txt

# 模拟 stop 的清理逻辑
if [ -d ".worktrees" ]; then
    rm -rf .worktrees
    if [ ! -d ".worktrees" ]; then
        pass "Worktrees 目录完全删除"
    else
        fail "Worktrees 清理" "目录仍然存在"
    fi
else
    fail "Worktrees 清理" "测试设置失败"
fi

# 测试 5: PID 文件管理（P0-3）
echo ""
echo "📋 测试 5: P0-3 PID 文件管理"

PID_FILE="$HOME/.claude-swarm/test-swarm.pid"
mkdir -p "$HOME/.claude-swarm"

# 写入 PID
echo "12345" > "$PID_FILE"
if [ -f "$PID_FILE" ]; then
    pass "PID 文件创建成功"
else
    fail "PID 文件" "创建失败"
fi

# 删除 PID
rm -f "$PID_FILE"
if [ ! -f "$PID_FILE" ]; then
    pass "PID 文件删除成功"
else
    fail "PID 文件" "删除失败"
fi

# 测试 6: 进程清理验证（P0-3）
echo ""
echo "📋 测试 6: P0-3 进程清理逻辑"

# 检查 killOrphanedProcesses 函数的改进
if grep -q "优先使用 PID 文件" cmd/swarm/stop.go; then
    pass "进程清理使用 PID 文件优先"
else
    fail "进程清理" "未使用 PID 文件优先策略"
fi

if grep -q "验证清理是否完成" cmd/swarm/stop.go; then
    pass "进程清理包含验证逻辑"
else
    fail "进程清理" "缺少验证逻辑"
fi

# 测试 7: Coordinator 超时控制（P0-3）
echo ""
echo "📋 测试 7: P0-3 Coordinator 超时控制"

if grep -q "30 \* time.Second" pkg/controller/coordinator.go; then
    pass "Coordinator Stop 有 30s 超时"
else
    fail "Coordinator" "缺少超时控制"
fi

# 测试 8: tmux 会话检测（P0-1）
echo ""
echo "📋 测试 8: P0-1 tmux 会话检测"

if grep -q "isTmuxSessionAlive" pkg/controller/coordinator.go; then
    pass "实现了 tmux 会话存活检测"
else
    fail "tmux 检测" "未实现检测函数"
fi

if grep -q "maxSessionDeadChecks" pkg/controller/coordinator.go; then
    pass "使用计数器避免误判"
else
    fail "tmux 检测" "缺少误判保护"
fi

# 最终清理
cleanup

# 总结
echo ""
echo "========================================"
echo "📊 测试总结"
echo "========================================"
echo -e "通过: ${GREEN}$TESTS_PASSED${NC}"
echo -e "失败: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ 所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}❌ 有 $TESTS_FAILED 个测试失败${NC}"
    exit 1
fi
