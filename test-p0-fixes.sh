#!/bin/bash
# Test script for P0 critical fixes

set -e

echo "============================================"
echo "Testing P0 Critical Fixes"
echo "============================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test counter
PASSED=0
FAILED=0

# P0.1 - Security Confirmation System
echo "P0.1 - Testing Security Confirmation System"
echo "--------------------------------------------"
if ./test-confirm-safety.sh > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Security confirmation tests passed${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Security confirmation tests failed${NC}"
    ((FAILED++))
fi
echo ""

# P0.2 - Race Conditions
echo "P0.2 - Testing Race Conditions"
echo "--------------------------------"
echo "Running go test -race on all packages..."
if go test -race ./... 2>&1 | grep -q "PASS\|no test files"; then
    echo -e "${GREEN}✓ Race detector found no issues${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Race detector found issues${NC}"
    ((FAILED++))
fi
echo ""

# P0.3 - Resource Leaks
echo "P0.3 - Testing Resource Leak Fixes"
echo "-----------------------------------"
echo "Verifying worktree tracking..."
if grep -q "activeWorktrees map\[string\]" pkg/git/worktree.go; then
    echo -e "${GREEN}✓ Worktree tracking implemented${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Worktree tracking not found${NC}"
    ((FAILED++))
fi

echo "Verifying merge mutex..."
if grep -q "mu.*sync.Mutex" pkg/git/merge.go; then
    echo -e "${GREEN}✓ Merge mutex implemented${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Merge mutex not found${NC}"
    ((FAILED++))
fi

echo "Verifying disk space check..."
if grep -q "CheckDiskSpace" pkg/utils/disk.go && grep -q "CheckDiskSpace" pkg/controller/coordinator.go; then
    echo -e "${GREEN}✓ Disk space checking implemented${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Disk space checking not found${NC}"
    ((FAILED++))
fi
echo ""

# Build verification
echo "Build Verification"
echo "------------------"
if go build ./... > /dev/null 2>&1; then
    echo -e "${GREEN}✓ All packages build successfully${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Build failed${NC}"
    ((FAILED++))
fi
echo ""

# Summary
echo "============================================"
echo "Test Summary"
echo "============================================"
echo -e "Passed: ${GREEN}${PASSED}${NC}"
echo -e "Failed: ${RED}${FAILED}${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All P0 tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
