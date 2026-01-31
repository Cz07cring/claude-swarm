#!/bin/bash

# TUI Monitor 自动化测试脚本
# 此脚本执行非交互式的 TUI 功能测试

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "=========================================="
echo "  TUI Monitor 自动化测试"
echo "=========================================="
echo ""

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
ISSUES_FOUND=0

# 记录问题的数组
declare -a ISSUES

# 测试函数
run_test() {
    local test_name="$1"
    local test_cmd="$2"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -n "测试 $TOTAL_TESTS: $test_name ... "

    if eval "$test_cmd" &>/dev/null; then
        echo -e "${GREEN}✓ 通过${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}✗ 失败${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
        ISSUES+=("$test_name")
        return 1
    fi
}

# 记录问题
record_issue() {
    local severity="$1"
    local description="$2"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
    echo -e "${YELLOW}  ⚠ [$severity] $description${NC}"
    ISSUES+=("[$severity] $description")
}

# ==================== 测试开始 ====================

echo -e "${BLUE}=== 第一部分: 环境检查 ===${NC}"
echo ""

# 测试 1: 检查二进制文件
run_test "检查 swarm 二进制文件存在" "[ -f ./swarm ]"

# 测试 2: 检查二进制文件可执行
run_test "检查 swarm 可执行" "[ -x ./swarm ]"

# 测试 3: 检查 Go 版本
echo -n "测试 $((TOTAL_TESTS + 1)): 检查 Go 版本 ... "
GO_VERSION=$(go version 2>/dev/null || echo "")
if [ -n "$GO_VERSION" ]; then
    echo -e "${GREEN}✓ 通过${NC} ($GO_VERSION)"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}✗ 失败${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

echo ""
echo -e "${BLUE}=== 第二部分: 测试数据准备 ===${NC}"
echo ""

# 测试 4: 创建测试目录
run_test "创建测试目录" "mkdir -p ~/.claude-swarm"

# 清理旧数据
echo "清理旧测试数据..."
rm -f ~/.claude-swarm/tasks.json ~/.claude-swarm/agents.json
rm -f ~/.claude-swarm/tasks.json.lock ~/.claude-swarm/agents.json.lock

# 测试 5: 创建测试任务数据
echo -n "测试 $((TOTAL_TESTS + 1)): 创建测试任务数据 ... "
cat > ~/.claude-swarm/tasks.json << 'EOF'
{
  "tasks": [
    {
      "id": "task-001",
      "description": "实现用户认证功能",
      "status": "pending",
      "created_at": "2026-01-30T10:00:00Z",
      "updated_at": "2026-01-30T10:00:00Z"
    },
    {
      "id": "task-002",
      "description": "创建数据库架构",
      "status": "in_progress",
      "assignee_id": "agent-0",
      "created_at": "2026-01-30T10:01:00Z",
      "updated_at": "2026-01-30T10:05:00Z"
    },
    {
      "id": "task-003",
      "description": "编写单元测试",
      "status": "completed",
      "assignee_id": "agent-1",
      "created_at": "2026-01-30T10:02:00Z",
      "updated_at": "2026-01-30T10:10:00Z"
    },
    {
      "id": "task-004",
      "description": "实现 API 端点",
      "status": "pending",
      "created_at": "2026-01-30T10:03:00Z",
      "updated_at": "2026-01-30T10:03:00Z"
    },
    {
      "id": "task-005",
      "description": "添加日志记录",
      "status": "failed",
      "assignee_id": "agent-2",
      "created_at": "2026-01-30T10:04:00Z",
      "updated_at": "2026-01-30T10:12:00Z"
    }
  ]
}
EOF

if [ -f ~/.claude-swarm/tasks.json ]; then
    echo -e "${GREEN}✓ 通过${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}✗ 失败${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# 测试 6: 创建测试 Agent 数据
echo -n "测试 $((TOTAL_TESTS + 1)): 创建测试 Agent 数据 ... "
cat > ~/.claude-swarm/agents.json << 'EOF'
{
  "agents": [
    {
      "agent_id": "agent-0",
      "state": "working",
      "current_task": {
        "id": "task-002",
        "description": "创建数据库架构",
        "status": "in_progress"
      },
      "last_update": "2026-01-30T10:05:30Z",
      "output": "正在创建数据库表...\nCREATE TABLE users (\n  id SERIAL PRIMARY KEY,\n  username VARCHAR(50) UNIQUE NOT NULL\n);"
    },
    {
      "agent_id": "agent-1",
      "state": "idle",
      "last_update": "2026-01-30T10:10:30Z",
      "output": "任务已完成，等待新任务..."
    },
    {
      "agent_id": "agent-2",
      "state": "error",
      "current_task": {
        "id": "task-005",
        "description": "添加日志记录",
        "status": "failed"
      },
      "last_update": "2026-01-30T10:12:30Z",
      "output": "错误: 无法导入日志库\nImportError: No module named 'logging'"
    },
    {
      "agent_id": "agent-3",
      "state": "idle",
      "last_update": "2026-01-30T10:00:00Z",
      "output": ""
    },
    {
      "agent_id": "agent-4",
      "state": "waiting_confirm",
      "last_update": "2026-01-30T10:13:00Z",
      "output": "准备提交更改到 git\n是否继续? (y/n)"
    }
  ],
  "updated_at": "2026-01-30T10:13:00Z"
}
EOF

if [ -f ~/.claude-swarm/agents.json ]; then
    echo -e "${GREEN}✓ 通过${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}✗ 失败${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

echo ""
echo -e "${BLUE}=== 第三部分: 数据验证 ===${NC}"
echo ""

# 测试 7: 验证任务数据格式
echo -n "测试 $((TOTAL_TESTS + 1)): 验证任务 JSON 格式 ... "
if jq empty ~/.claude-swarm/tasks.json 2>/dev/null; then
    echo -e "${GREEN}✓ 通过${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}✗ 失败${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
    record_issue "高" "任务 JSON 格式无效"
fi

# 测试 8: 验证 Agent 数据格式
echo -n "测试 $((TOTAL_TESTS + 1)): 验证 Agent JSON 格式 ... "
if jq empty ~/.claude-swarm/agents.json 2>/dev/null; then
    echo -e "${GREEN}✓ 通过${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}✗ 失败${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
    record_issue "高" "Agent JSON 格式无效"
fi

# 测试 9: 统计任务状态
echo -n "测试 $((TOTAL_TESTS + 1)): 统计任务状态分布 ... "
PENDING_COUNT=$(jq '[.tasks[] | select(.status == "pending")] | length' ~/.claude-swarm/tasks.json)
IN_PROGRESS_COUNT=$(jq '[.tasks[] | select(.status == "in_progress")] | length' ~/.claude-swarm/tasks.json)
COMPLETED_COUNT=$(jq '[.tasks[] | select(.status == "completed")] | length' ~/.claude-swarm/tasks.json)
FAILED_COUNT=$(jq '[.tasks[] | select(.status == "failed")] | length' ~/.claude-swarm/tasks.json)

if [ "$PENDING_COUNT" -eq 2 ] && [ "$IN_PROGRESS_COUNT" -eq 1 ] && \
   [ "$COMPLETED_COUNT" -eq 1 ] && [ "$FAILED_COUNT" -eq 1 ]; then
    echo -e "${GREEN}✓ 通过${NC}"
    echo "  待处理: $PENDING_COUNT, 进行中: $IN_PROGRESS_COUNT, 已完成: $COMPLETED_COUNT, 失败: $FAILED_COUNT"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}✗ 失败${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# 测试 10: 统计 Agent 状态
echo -n "测试 $((TOTAL_TESTS + 1)): 统计 Agent 状态分布 ... "
IDLE_COUNT=$(jq '[.agents[] | select(.state == "idle")] | length' ~/.claude-swarm/agents.json)
WORKING_COUNT=$(jq '[.agents[] | select(.state == "working")] | length' ~/.claude-swarm/agents.json)
ERROR_COUNT=$(jq '[.agents[] | select(.state == "error")] | length' ~/.claude-swarm/agents.json)
WAITING_COUNT=$(jq '[.agents[] | select(.state == "waiting_confirm")] | length' ~/.claude-swarm/agents.json)

if [ "$IDLE_COUNT" -eq 2 ] && [ "$WORKING_COUNT" -eq 1 ] && \
   [ "$ERROR_COUNT" -eq 1 ] && [ "$WAITING_COUNT" -eq 1 ]; then
    echo -e "${GREEN}✓ 通过${NC}"
    echo "  空闲: $IDLE_COUNT, 工作中: $WORKING_COUNT, 错误: $ERROR_COUNT, 等待确认: $WAITING_COUNT"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}✗ 失败${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

echo ""
echo -e "${BLUE}=== 第四部分: 功能检查 ===${NC}"
echo ""

# 检查代码中的功能实现
echo "检查 TUI 功能实现..."

# 测试 11: 检查帮助面板功能
echo -n "测试 $((TOTAL_TESTS + 1)): 检查帮助面板实现 (按键 ?) ... "
if grep -q "\"?\"" pkg/tui/dashboard.go 2>/dev/null; then
    echo -e "${GREEN}✓ 已实现${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${YELLOW}○ 未实现${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
    record_issue "中" "帮助面板功能未实现 (按键 ?)"
fi

# 测试 12: 检查任务过滤功能
echo -n "测试 $((TOTAL_TESTS + 1)): 检查任务过滤实现 (按键 1-4) ... "
if grep -q "filter" pkg/tui/tasklist.go 2>/dev/null && \
   grep -q "\"1\"\|\"2\"\|\"3\"\|\"4\"" pkg/tui/dashboard.go 2>/dev/null; then
    echo -e "${GREEN}✓ 已实现${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${YELLOW}○ 未实现${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
    record_issue "中" "任务过滤功能未实现 (按键 1-4)"
fi

# 测试 13: 检查主题切换功能
echo -n "测试 $((TOTAL_TESTS + 1)): 检查主题切换实现 (按键 t) ... "
if grep -q "theme" pkg/tui/dashboard.go 2>/dev/null || \
   grep -q "\"t\"" pkg/tui/dashboard.go 2>/dev/null; then
    echo -e "${GREEN}✓ 已实现${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${YELLOW}○ 未实现${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
    record_issue "低" "主题切换功能未实现 (按键 t)"
fi

# 测试 14: 检查性能统计功能
echo -n "测试 $((TOTAL_TESTS + 1)): 检查 Agent 性能统计实现 ... "
if grep -q "statistics\|stats\|performance" pkg/tui/agentgrid.go 2>/dev/null; then
    echo -e "${GREEN}✓ 已实现${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${YELLOW}○ 未实现${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
    record_issue "中" "Agent 性能统计功能未实现"
fi

# 测试 15: 检查基础导航功能
echo -n "测试 $((TOTAL_TESTS + 1)): 检查基础导航功能 (Tab, hjkl) ... "
if grep -q "tab" pkg/tui/dashboard.go 2>/dev/null && \
   grep -q "\"h\"\|\"j\"\|\"k\"\|\"l\"" pkg/tui/dashboard.go 2>/dev/null; then
    echo -e "${GREEN}✓ 已实现${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}✗ 未实现${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
    record_issue "高" "基础导航功能缺失"
fi

echo ""
echo -e "${BLUE}=== 第五部分: 代码质量检查 ===${NC}"
echo ""

# 测试 16: 检查错误处理
echo -n "测试 $((TOTAL_TESTS + 1)): 检查错误处理 ... "
ERROR_CHECKS=$(grep -c "if err != nil" pkg/tui/*.go 2>/dev/null || echo "0")
if [ "$ERROR_CHECKS" -gt 5 ]; then
    echo -e "${GREEN}✓ 通过${NC} (找到 $ERROR_CHECKS 处错误检查)"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${YELLOW}⚠ 警告${NC} (仅找到 $ERROR_CHECKS 处错误检查)"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
    record_issue "低" "错误处理可能不够完善"
fi

# 测试 17: 检查空值检查
echo -n "测试 $((TOTAL_TESTS + 1)): 检查空值保护 ... "
NIL_CHECKS=$(grep -c "!= nil\|== nil" pkg/tui/*.go 2>/dev/null || echo "0")
if [ "$NIL_CHECKS" -gt 10 ]; then
    echo -e "${GREEN}✓ 通过${NC} (找到 $NIL_CHECKS 处空值检查)"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${YELLOW}⚠ 警告${NC} (仅找到 $NIL_CHECKS 处空值检查)"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
    record_issue "中" "空值保护可能不够完善"
fi

# ==================== 测试总结 ====================

echo ""
echo "=========================================="
echo "  测试结果总结"
echo "=========================================="
echo ""

echo "执行统计:"
echo "  总测试数: $TOTAL_TESTS"
echo -e "  ${GREEN}通过: $PASSED_TESTS${NC}"
echo -e "  ${RED}失败: $FAILED_TESTS${NC}"
echo -e "  成功率: $(awk "BEGIN {printf \"%.1f\", ($PASSED_TESTS/$TOTAL_TESTS)*100}")%"
echo ""

echo "发现的问题: $ISSUES_FOUND 个"
if [ $ISSUES_FOUND -gt 0 ]; then
    echo ""
    echo "问题清单:"
    for issue in "${ISSUES[@]}"; do
        echo "  • $issue"
    done
fi

echo ""
echo "测试数据位置:"
echo "  任务: ~/.claude-swarm/tasks.json"
echo "  Agent: ~/.claude-swarm/agents.json"
echo ""

echo "下一步:"
echo "  1. 查看测试计划: cat TUI_TEST_PLAN.md"
echo "  2. 运行交互式测试: ./swarm monitor"
echo "  3. 查看详细测试结果: cat TUI_TEST_REPORT.md"
echo ""

# 生成简要报告
cat > /tmp/tui_test_summary.txt << EOF
TUI Monitor 自动化测试报告
测试日期: $(date +%Y-%m-%d\ %H:%M:%S)
=======================================

执行统计:
- 总测试数: $TOTAL_TESTS
- 通过: $PASSED_TESTS
- 失败: $FAILED_TESTS
- 成功率: $(awk "BEGIN {printf \"%.1f\", ($PASSED_TESTS/$TOTAL_TESTS)*100}")%

发现的问题: $ISSUES_FOUND 个

EOF

for issue in "${ISSUES[@]}"; do
    echo "• $issue" >> /tmp/tui_test_summary.txt
done

echo "测试报告已保存到: /tmp/tui_test_summary.txt"
echo ""

# 返回退出码
if [ $FAILED_TESTS -eq 0 ]; then
    exit 0
else
    exit 1
fi
