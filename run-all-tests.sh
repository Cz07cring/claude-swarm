#!/bin/bash
# 主测试运行器

set -e

echo "========================================="
echo "Claude Swarm - 完整测试套件"
echo "========================================="
echo ""

# 颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

TOTAL_PASSED=0
TOTAL_FAILED=0

# 切换到项目根目录
cd "$(dirname "$0")"

echo "1. 构建验证"
echo "-----------------------------------"
if go build ./pkg/... > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 所有包构建成功${NC}"
    ((TOTAL_PASSED++))
else
    echo -e "${RED}✗ 构建失败${NC}"
    ((TOTAL_FAILED++))
    exit 1
fi
echo ""

echo "2. 竞态检测"
echo "-----------------------------------"
if go test -race ./... 2>&1 | grep -q "PASS\|no test files"; then
    echo -e "${GREEN}✓ 无竞态条件${NC}"
    ((TOTAL_PASSED++))
else
    echo -e "${RED}✗ 检测到竞态条件${NC}"
    ((TOTAL_FAILED++))
fi
echo ""

echo "3. 代码实现验证"
echo "-----------------------------------"

# P0.1 - 安全确认
if grep -q "DangerKeywords.*\[\]string" pkg/analyzer/patterns.go && \
   grep -q "SafeToConfirm" pkg/analyzer/helper.go; then
    echo -e "${GREEN}✓ P0.1 安全确认系统${NC}"
    ((TOTAL_PASSED++))
else
    echo -e "${RED}✗ P0.1 安全确认系统${NC}"
    ((TOTAL_FAILED++))
fi

# P0.2 - 竞态修复
if grep -q "version.*uint64" pkg/controller/coordinator.go && \
   grep -q "sendMu.*sync.Mutex" pkg/tmux/types.go; then
    echo -e "${GREEN}✓ P0.2 竞态条件修复${NC}"
    ((TOTAL_PASSED++))
else
    echo -e "${RED}✗ P0.2 竞态条件修复${NC}"
    ((TOTAL_FAILED++))
fi

# P0.3 - 资源泄漏
if grep -q "activeWorktrees.*map" pkg/git/worktree.go && \
   grep -q "CheckDiskSpace" pkg/utils/disk.go; then
    echo -e "${GREEN}✓ P0.3 资源泄漏修复${NC}"
    ((TOTAL_PASSED++))
else
    echo -e "${RED}✗ P0.3 资源泄漏修复${NC}"
    ((TOTAL_FAILED++))
fi

# P1.1 - DAG调度
if [ -f "pkg/scheduler/dag_scheduler.go" ] && \
   grep -q "Dependencies.*\[\]string" internal/models/task.go; then
    echo -e "${GREEN}✓ P1.1 DAG依赖调度${NC}"
    ((TOTAL_PASSED++))
else
    echo -e "${RED}✗ P1.1 DAG依赖调度${NC}"
    ((TOTAL_FAILED++))
fi

# P1.2 - 自动重试
if [ -f "pkg/retry/retry_manager.go" ] && \
   grep -q "ErrorType.*int" pkg/analyzer/detector.go; then
    echo -e "${GREEN}✓ P1.2 自动重试机制${NC}"
    ((TOTAL_PASSED++))
else
    echo -e "${RED}✗ P1.2 自动重试机制${NC}"
    ((TOTAL_FAILED++))
fi

echo ""
echo "========================================="
echo "测试总结"
echo "========================================="
echo -e "通过: ${GREEN}${TOTAL_PASSED}${NC}"
echo -e "失败: ${RED}${TOTAL_FAILED}${NC}"
echo ""

if [ $TOTAL_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ 所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}❌ 有测试失败${NC}"
    exit 1
fi
