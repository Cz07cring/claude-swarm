#!/bin/bash
# TUI 用户体验自动化测试脚本

set -e

BOLD="\033[1m"
GREEN="\033[0;32m"
RED="\033[0;31m"
YELLOW="\033[0;33m"
BLUE="\033[0;34m"
RESET="\033[0m"

echo -e "${BOLD}🧪 Claude Swarm TUI 用户体验测试${RESET}"
echo "=========================================="
echo ""

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试函数
run_test() {
    local test_name="$1"
    local test_cmd="$2"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "${BLUE}▶ 测试 ${TOTAL_TESTS}: ${test_name}${RESET}"

    if eval "$test_cmd"; then
        echo -e "${GREEN}  ✓ 通过${RESET}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}  ✗ 失败${RESET}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
    echo ""
}

# 1. 基础检查
echo -e "${BOLD}📦 阶段 1: 基础检查${RESET}"
echo "----------------------------------------"
run_test "检查编译产物存在" "test -f ./bin/swarm"
run_test "检查二进制可执行" "test -x ./bin/swarm"
run_test "检查帮助命令" "./bin/swarm --help > /dev/null 2>&1"
echo ""

# 2. 数据文件检查
echo -e "${BOLD}📁 阶段 2: 数据文件检查${RESET}"
echo "----------------------------------------"
run_test "检查数据目录存在" "test -d ~/.claude-swarm"
run_test "检查任务文件存在" "test -f ~/.claude-swarm/tasks.json"
run_test "检查 Agent 文件存在" "test -f ~/.claude-swarm/agents.json"
echo ""

# 3. 数据内容分析
echo -e "${BOLD}📊 阶段 3: 数据内容分析${RESET}"
echo "----------------------------------------"

TOTAL_TASKS=$(cat ~/.claude-swarm/tasks.json | grep -o '"id"' | wc -l | tr -d ' ')
COMPLETED_TASKS=$(cat ~/.claude-swarm/tasks.json | grep -o '"status": "completed"' | wc -l | tr -d ' ')
FAILED_TASKS=$(cat ~/.claude-swarm/tasks.json | grep -o '"status": "failed"' | wc -l | tr -d ' ')
TOTAL_AGENTS=$(cat ~/.claude-swarm/agents.json | grep -o '"agent_id"' | wc -l | tr -d ' ')

echo -e "  ${GREEN}✓${RESET} 任务总数: $TOTAL_TASKS (完成: $COMPLETED_TASKS, 失败: $FAILED_TASKS)"
echo -e "  ${GREEN}✓${RESET} Agent 总数: $TOTAL_AGENTS"
echo ""

# 4. 性能测试
echo -e "${BOLD}⚡ 阶段 4: 性能测试${RESET}"
echo "----------------------------------------"

BINARY_SIZE=$(du -h ./bin/swarm | awk '{print $1}')
echo -e "  ${BLUE}ℹ${RESET}  二进制文件大小: $BINARY_SIZE"

START_TIME=$(date +%s%N 2>/dev/null || date +%s000000000)
./bin/swarm status > /dev/null 2>&1 || true
END_TIME=$(date +%s%N 2>/dev/null || date +%s000000000)
DURATION=$(( (END_TIME - START_TIME) / 1000000 ))
echo -e "  ${BLUE}ℹ${RESET}  状态命令响应时间: ${DURATION}ms"
echo ""

# 5. 功能完整性检查
echo -e "${BOLD}⌨️  阶段 5: 功能完整性检查${RESET}"
echo "----------------------------------------"

check_feature() {
    local feature_name="$1"
    local search_pattern="$2"
    local file_path="$3"

    if grep -q "$search_pattern" "$file_path" 2>/dev/null; then
        echo -e "  ${GREEN}✓${RESET} $feature_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "  ${RED}✗${RESET} $feature_name"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

check_feature "状态栏渲染" "renderStatusBar" "pkg/tui/dashboard.go"
check_feature "滚动功能" "ScrollUp" "pkg/tui/logviewer.go"
check_feature "自适应网格" "calculateOptimalGrid" "pkg/tui/agentgrid.go"
check_feature "Home/End 支持" "MoveToFirst" "pkg/tui/agentgrid.go"
echo ""

# 最终报告
echo -e "${BOLD}📋 测试报告${RESET}"
echo "=========================================="
echo -e "  总计测试: ${BOLD}$TOTAL_TESTS${RESET}"
echo -e "  通过: ${GREEN}$PASSED_TESTS${RESET}"
echo -e "  失败: ${RED}$FAILED_TESTS${RESET}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}${BOLD}✓ 所有测试通过！${RESET}\n"
    exit 0
else
    PASS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo -e "\n${YELLOW}⚠ 通过率: ${PASS_RATE}%${RESET}\n"
    exit 1
fi
