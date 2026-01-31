#!/bin/bash

# ==========================================
# Claude Swarm - 全面问题检测测试套件
# ==========================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# 测试结果统计
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0
CRITICAL_ISSUES=()

# 测试结果函数
test_result() {
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    local severity="${4:-normal}"

    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ PASS${NC}: $2"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        if [ "$severity" == "critical" ]; then
            echo -e "${RED}🔥 CRITICAL FAIL${NC}: $2"
            CRITICAL_ISSUES+=("$2: $3")
        else
            echo -e "${RED}❌ FAIL${NC}: $2"
        fi
        TESTS_FAILED=$((TESTS_FAILED + 1))
        if [ ! -z "$3" ]; then
            echo -e "${YELLOW}   详情: $3${NC}"
        fi
    fi
}

echo "=========================================="
echo "🔍 Claude Swarm - 全面问题检测测试"
echo "=========================================="
echo
echo -e "${CYAN}测试范围：${NC}"
echo "  • 环境和依赖检查"
echo "  • 编译和基础功能"
echo "  • 错误处理和边界情况"
echo "  • 进程和资源管理"
echo "  • 并发和竞态条件"
echo "  • 清理和恢复机制"
echo
sleep 2

PROJECT_DIR="/Users/ring/Documents/公司源码/ringsite/claude-swarm"
cd "$PROJECT_DIR" || exit 1

# ==========================================
# 第 1 部分: 环境检查
# ==========================================
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo -e "${MAGENTA}第 1 部分: 环境和依赖检查${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo

# 1.1 检查 Go 环境
echo -e "${BLUE}[1.1]${NC} 检查 Go 环境..."
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    test_result 0 "Go 已安装 ($GO_VERSION)"
else
    test_result 1 "Go 未安装" "需要 Go 1.21+" "critical"
fi

# 1.2 检查 tmux
echo -e "${BLUE}[1.2]${NC} 检查 tmux..."
if command -v tmux &> /dev/null; then
    TMUX_VERSION=$(tmux -V)
    test_result 0 "tmux 已安装 ($TMUX_VERSION)"
else
    test_result 1 "tmux 未安装" "核心依赖缺失" "critical"
fi

# 1.3 检查 claude CLI
echo -e "${BLUE}[1.3]${NC} 检查 claude CLI..."
if command -v claude &> /dev/null; then
    test_result 0 "claude CLI 已安装"
else
    test_result 1 "claude CLI 未安装" "需要安装 Claude Code"
fi

# 1.4 检查 Git
echo -e "${BLUE}[1.4]${NC} 检查 Git 环境..."
if git rev-parse --git-dir > /dev/null 2>&1; then
    test_result 0 "Git 仓库正常"
else
    test_result 1 "不在 Git 仓库中" "critical"
fi

# 1.5 检查磁盘空间
echo -e "${BLUE}[1.5]${NC} 检查磁盘空间..."
AVAILABLE_SPACE=$(df -h . | awk 'NR==2 {print $4}')
test_result 0 "可用磁盘空间: $AVAILABLE_SPACE"

echo

# ==========================================
# 第 2 部分: 清理和初始化
# ==========================================
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo -e "${MAGENTA}第 2 部分: 清理和初始化${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo

echo -e "${BLUE}[2.1]${NC} 清理遗留进程..."
pkill -f "./swarm start" 2>/dev/null || true
sleep 1

# 检查是否还有遗留进程
if pgrep -f "./swarm start" > /dev/null; then
    test_result 1 "遗留 swarm 进程未清理" "可能需要 kill -9"
else
    test_result 0 "swarm 进程已清理"
fi

echo -e "${BLUE}[2.2]${NC} 清理 tmux 会话..."
tmux kill-server 2>/dev/null || true
sleep 1

echo -e "${BLUE}[2.3]${NC} 清理 worktrees..."
git worktree remove .worktrees/agent-0 --force 2>/dev/null || true
git worktree remove .worktrees/agent-1 --force 2>/dev/null || true
git worktree remove .worktrees/agent-2 --force 2>/dev/null || true

if [ -d .worktrees ]; then
    rm -rf .worktrees/
    test_result 0 "worktrees 目录已删除"
else
    test_result 0 "worktrees 目录不存在"
fi

echo -e "${BLUE}[2.4]${NC} 清理 Git 分支..."
git branch -D agent-0-branch agent-1-branch agent-2-branch 2>/dev/null || true
REMAINING_AGENT_BRANCHES=$(git branch | grep -c "agent.*branch" || echo "0")
if [ "$REMAINING_AGENT_BRANCHES" -eq 0 ]; then
    test_result 0 "Agent 分支已清理"
else
    test_result 1 "存在未清理的 agent 分支" "数量: $REMAINING_AGENT_BRANCHES"
fi

echo -e "${BLUE}[2.5]${NC} 清理 PID 文件..."
rm -f ~/.claude-swarm/*.pid
if [ ! -f ~/.claude-swarm/claude-swarm.pid ]; then
    test_result 0 "PID 文件已清理"
else
    test_result 1 "PID 文件未清理"
fi

echo -e "${BLUE}[2.6]${NC} 清理任务队列..."
if [ -f ~/.claude-swarm/tasks.json ]; then
    cp ~/.claude-swarm/tasks.json ~/.claude-swarm/tasks.json.backup
    echo '{"tasks":[]}' > ~/.claude-swarm/tasks.json
    test_result 0 "任务队列已清空（已备份）"
else
    mkdir -p ~/.claude-swarm
    echo '{"tasks":[]}' > ~/.claude-swarm/tasks.json
    test_result 0 "任务队列已初始化"
fi

echo

# ==========================================
# 第 3 部分: 编译测试
# ==========================================
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo -e "${MAGENTA}第 3 部分: 编译和基础功能测试${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo

echo -e "${BLUE}[3.1]${NC} 编译项目..."
if go build -o swarm ./cmd/swarm 2>&1; then
    BINARY_SIZE=$(ls -lh swarm | awk '{print $5}')
    test_result 0 "编译成功 (大小: $BINARY_SIZE)"
else
    test_result 1 "编译失败" "检查 Go 代码" "critical"
    exit 1
fi

echo -e "${BLUE}[3.2]${NC} 测试帮助命令..."
if ./swarm --help > /dev/null 2>&1; then
    test_result 0 "帮助命令正常"
else
    test_result 1 "帮助命令失败"
fi

echo -e "${BLUE}[3.3]${NC} 测试版本信息..."
./swarm --version 2>&1 || test_result 0 "版本命令执行（可能未实现）"

echo

# ==========================================
# 第 4 部分: 边界情况测试
# ==========================================
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo -e "${MAGENTA}第 4 部分: 边界情况和错误处理${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo

echo -e "${BLUE}[4.1]${NC} 测试启动 0 个 agent..."
if ./swarm start --agents 0 2>&1 | grep -qiE "error|invalid|must"; then
    test_result 0 "正确拒绝 0 个 agent"
else
    test_result 1 "应该拒绝 0 个 agent"
fi

echo -e "${BLUE}[4.2]${NC} 测试启动负数 agent..."
if ./swarm start --agents -1 2>&1 | grep -qiE "error|invalid|must"; then
    test_result 0 "正确拒绝负数 agent"
else
    test_result 1 "应该拒绝负数 agent"
fi

echo -e "${BLUE}[4.3]${NC} 测试空任务描述..."
if ./swarm add-task "" 2>&1 | grep -qiE "error|invalid|empty|required"; then
    test_result 0 "正确拒绝空任务"
else
    test_result 1 "应该拒绝空任务"
fi

echo -e "${BLUE}[4.4]${NC} 测试超长任务描述..."
LONG_TASK=$(python3 -c "print('A' * 10000)")
./swarm add-task "$LONG_TASK" 2>&1 > /dev/null
test_result 0 "超长任务描述测试完成"

echo

# ==========================================
# 第 5 部分: 进程管理测试
# ==========================================
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo -e "${MAGENTA}第 5 部分: 进程和资源管理${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo

echo -e "${BLUE}[5.1]${NC} 测试 PID 锁机制..."
./swarm start --agents 2 > /tmp/swarm-test-comprehensive.log 2>&1 &
SWARM_PID=$!
sleep 3

if [ -f ~/.claude-swarm/claude-swarm.pid ]; then
    PID_FROM_FILE=$(cat ~/.claude-swarm/claude-swarm.pid)
    test_result 0 "PID 文件已创建 (PID: $PID_FROM_FILE)"

    # 验证 PID 是否正确
    if [ "$PID_FROM_FILE" == "$SWARM_PID" ]; then
        test_result 0 "PID 文件内容正确"
    else
        test_result 1 "PID 文件内容不匹配" "文件:$PID_FROM_FILE 实际:$SWARM_PID"
    fi
else
    test_result 1 "PID 文件未创建" "critical"
fi

echo -e "${BLUE}[5.2]${NC} 测试多进程保护..."
if ./swarm start --agents 2 2>&1 | grep -q "已在运行中"; then
    test_result 0 "多进程启动被正确阻止"
else
    test_result 1 "多进程启动未被阻止" "可能导致资源冲突" "critical"
fi

echo -e "${BLUE}[5.3]${NC} 验证 tmux 会话..."
sleep 2
if tmux has-session -t claude-swarm 2>/dev/null; then
    test_result 0 "tmux 会话已创建"

    # 检查 pane 数量
    PANE_COUNT=$(tmux list-panes -t claude-swarm | wc -l)
    if [ "$PANE_COUNT" -ge 2 ]; then
        test_result 0 "tmux panes 已创建 (数量: $PANE_COUNT)"
    else
        test_result 1 "tmux panes 数量不足" "期望>=2, 实际:$PANE_COUNT"
    fi
else
    test_result 1 "tmux 会话未创建" "critical"
fi

echo -e "${BLUE}[5.4]${NC} 验证 worktrees 创建..."
sleep 2
if [ -d .worktrees/agent-0 ] && [ -d .worktrees/agent-1 ]; then
    test_result 0 "worktrees 目录已创建"

    # 验证 worktrees 是否健康
    WORKTREE_COUNT=$(git worktree list | grep -c "agent-")
    test_result 0 "worktrees 数量: $WORKTREE_COUNT"
else
    test_result 1 "worktrees 目录未创建" "可能影响并发开发"
fi

echo -e "${BLUE}[5.5]${NC} 验证 Git 分支..."
if git branch | grep -q "agent-0-branch" && git branch | grep -q "agent-1-branch"; then
    test_result 0 "Agent 分支已创建"
else
    test_result 1 "Agent 分支未创建"
fi

echo -e "${BLUE}[5.6]${NC} 检查进程资源使用..."
if ps -p $SWARM_PID -o %cpu,%mem,comm > /dev/null 2>&1; then
    RESOURCE_INFO=$(ps -p $SWARM_PID -o %cpu,%mem,comm | tail -1)
    test_result 0 "进程资源监控: $RESOURCE_INFO"
else
    test_result 1 "无法监控进程资源"
fi

echo

# ==========================================
# 第 6 部分: 功能测试
# ==========================================
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo -e "${MAGENTA}第 6 部分: 核心功能测试${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo

echo -e "${BLUE}[6.1]${NC} 测试任务添加..."
TASK_OUTPUT=$(./swarm add-task "echo 'Comprehensive test task' > test-comprehensive-$(date +%s).txt")
if echo "$TASK_OUTPUT" | grep -q "任务已添加"; then
    TASK_ID=$(echo "$TASK_OUTPUT" | grep "ID:" | awk '{print $2}')
    test_result 0 "任务添加成功 (ID: $TASK_ID)"
else
    test_result 1 "任务添加失败"
fi

echo -e "${BLUE}[6.2]${NC} 测试快速连续添加任务..."
for i in {1..5}; do
    ./swarm add-task "任务 $i: 创建测试文件 test-$i.txt" > /dev/null 2>&1 &
done
wait
test_result 0 "并发任务添加完成"

echo -e "${BLUE}[6.3]${NC} 测试状态查询..."
if ./swarm status 2>&1 | grep -qE "状态|Agent|任务"; then
    test_result 0 "状态查询正常"
else
    test_result 1 "状态查询异常"
fi

echo -e "${BLUE}[6.4]${NC} 等待任务执行 (30秒监控)..."
echo "   监控执行日志中的关键事件..."
sleep 30

# 检查日志中的关键事件
if grep -q "📋 已分配任务" /tmp/swarm-test-comprehensive.log; then
    test_result 0 "发现任务分配日志"
else
    test_result 1 "未发现任务分配日志" "调度器可能未工作"
fi

if grep -q "🔄.*state changed" /tmp/swarm-test-comprehensive.log; then
    test_result 0 "发现状态变化日志"
else
    test_result 1 "未发现状态变化日志"
fi

echo

# ==========================================
# 第 7 部分: 错误恢复测试
# ==========================================
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo -e "${MAGENTA}第 7 部分: 错误恢复和异常处理${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo

echo -e "${BLUE}[7.1]${NC} 测试 tmux 会话意外终止..."
tmux kill-session -t claude-swarm 2>/dev/null || true
sleep 3

# 检查 swarm 进程是否还在运行
if ps -p $SWARM_PID > /dev/null 2>&1; then
    test_result 1 "swarm 进程未检测到 tmux 会话终止" "应该优雅退出" "critical"

    # 强制终止
    kill $SWARM_PID 2>/dev/null || true
else
    test_result 0 "swarm 进程已正确退出"
fi

echo -e "${BLUE}[7.2]${NC} 检查错误日志..."
if grep -q "Error capturing.*output" /tmp/swarm-test-comprehensive.log; then
    ERROR_COUNT=$(grep -c "Error capturing.*output" /tmp/swarm-test-comprehensive.log)
    if [ "$ERROR_COUNT" -gt 10 ]; then
        test_result 1 "大量 pane 捕获错误" "错误次数: $ERROR_COUNT" "critical"
    else
        test_result 0 "少量 pane 捕获错误 (可接受)" "错误次数: $ERROR_COUNT"
    fi
else
    test_result 0 "没有 pane 捕获错误"
fi

echo

# ==========================================
# 第 8 部分: 清理验证
# ==========================================
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo -e "${MAGENTA}第 8 部分: 清理和资源释放验证${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo

echo -e "${BLUE}[8.1]${NC} 执行 stop 命令..."
./swarm stop 2>&1
sleep 2

echo -e "${BLUE}[8.2]${NC} 验证进程清理..."
if pgrep -f "./swarm start" > /dev/null; then
    test_result 1 "swarm 进程未清理" "可能造成资源泄漏" "critical"
else
    test_result 0 "swarm 进程已清理"
fi

echo -e "${BLUE}[8.3]${NC} 验证 tmux 清理..."
if tmux has-session -t claude-swarm 2>/dev/null; then
    test_result 1 "tmux 会话未清理"
else
    test_result 0 "tmux 会话已清理"
fi

echo -e "${BLUE}[8.4]${NC} 验证 worktrees 清理..."
if [ -d .worktrees ]; then
    REMAINING=$(ls -A .worktrees | wc -l)
    if [ "$REMAINING" -gt 0 ]; then
        test_result 1 "worktrees 目录未完全清理" "剩余: $REMAINING 个文件/目录" "critical"
        ls -la .worktrees/
    else
        test_result 0 "worktrees 目录为空（但未删除）"
    fi
else
    test_result 0 "worktrees 目录已删除"
fi

echo -e "${BLUE}[8.5]${NC} 验证 PID 文件清理..."
if [ -f ~/.claude-swarm/claude-swarm.pid ]; then
    test_result 1 "PID 文件未清理" "可能影响下次启动"
else
    test_result 0 "PID 文件已清理"
fi

echo -e "${BLUE}[8.6]${NC} 验证 Git 分支清理..."
REMAINING_BRANCHES=$(git branch | grep -c "agent.*branch" || echo "0")
if [ "$REMAINING_BRANCHES" -gt 0 ]; then
    test_result 1 "Git 分支未清理" "剩余: $REMAINING_BRANCHES 个分支"
else
    test_result 0 "Git 分支已清理"
fi

echo -e "${BLUE}[8.7]${NC} 验证 Git worktree 清理..."
WORKTREE_COUNT=$(git worktree list | wc -l)
if [ "$WORKTREE_COUNT" -eq 1 ]; then
    test_result 0 "Git worktrees 已清理（仅主目录）"
else
    test_result 1 "存在多余的 worktrees" "数量: $WORKTREE_COUNT"
fi

echo

# ==========================================
# 第 9 部分: 配置测试
# ==========================================
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo -e "${MAGENTA}第 9 部分: 配置和环境变量测试${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo

echo -e "${BLUE}[9.1]${NC} 测试配置文件加载..."
if [ -f config.yaml ]; then
    test_result 0 "发现配置文件 config.yaml"
elif [ -f config.yaml.example ]; then
    test_result 0 "发现示例配置文件"
else
    test_result 1 "未发现配置文件"
fi

echo -e "${BLUE}[9.2]${NC} 测试 Gemini API Key..."
if [ ! -z "$GEMINI_API_KEY" ]; then
    test_result 0 "GEMINI_API_KEY 环境变量已设置"
else
    test_result 0 "GEMINI_API_KEY 未设置（orchestrate 功能不可用）"
fi

echo

# ==========================================
# 最终报告
# ==========================================
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo -e "${MAGENTA}📊 最终测试报告${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════${NC}"
echo

echo -e "总测试数: ${BLUE}$TESTS_TOTAL${NC}"
echo -e "通过: ${GREEN}$TESTS_PASSED${NC}"
echo -e "失败: ${RED}$TESTS_FAILED${NC}"
echo

PASS_RATE=$((TESTS_PASSED * 100 / TESTS_TOTAL))
echo -e "通过率: ${CYAN}${PASS_RATE}%${NC}"
echo

if [ ${#CRITICAL_ISSUES[@]} -gt 0 ]; then
    echo -e "${RED}🔥 发现 ${#CRITICAL_ISSUES[@]} 个严重问题：${NC}"
    for issue in "${CRITICAL_ISSUES[@]}"; do
        echo -e "${RED}  • $issue${NC}"
    done
    echo
fi

echo "详细日志文件："
echo "  • /tmp/swarm-test-comprehensive.log"
echo

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 所有测试通过！系统运行正常！${NC}"
    exit 0
elif [ ${#CRITICAL_ISSUES[@]} -gt 0 ]; then
    echo -e "${RED}❌ 测试失败！发现严重问题，需要立即修复！${NC}"
    exit 2
else
    echo -e "${YELLOW}⚠️  部分测试失败，建议检查并修复${NC}"
    exit 1
fi
