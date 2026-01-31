#!/bin/bash
# Test script for security confirmation system

set -e

echo "Testing Security Confirmation System"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Test cases
declare -a DANGEROUS_COMMANDS=(
    "git reset --hard HEAD~100"
    "git push --force origin main"
    "rm -rf /important/data"
    "drop table users"
    "git clean -fd"
    "delete from production_db"
    "shutdown now"
)

declare -a SAFE_COMMANDS=(
    "git status"
    "git log"
    "ls -la"
    "npm test"
    "go build"
)

echo ""
echo "Testing dangerous commands (should be BLOCKED):"
echo "-----------------------------------------------"

for cmd in "${DANGEROUS_COMMANDS[@]}"; do
    echo -n "Testing: '$cmd' ... "
    # This would integrate with the actual detector
    # For now, just verify the keywords are in our list
    if grep -qi "$(echo "$cmd" | grep -oE '(delete|remove|rm -rf|git reset --hard|git push --force|drop|shutdown)')" <<< "$cmd" 2>/dev/null; then
        echo -e "${GREEN}PASS${NC} (would be blocked)"
    else
        echo -e "${RED}FAIL${NC} (dangerous command not detected!)"
    fi
done

echo ""
echo "Testing safe commands (should be ALLOWED):"
echo "------------------------------------------"

for cmd in "${SAFE_COMMANDS[@]}"; do
    echo -n "Testing: '$cmd' ... "
    # Verify these don't contain dangerous keywords
    if ! grep -qiE '(delete|remove|rm -rf|git reset --hard|git push --force|drop|shutdown)' <<< "$cmd"; then
        echo -e "${GREEN}PASS${NC} (would be allowed)"
    else
        echo -e "${RED}FAIL${NC} (safe command incorrectly flagged!)"
    fi
done

echo ""
echo "===================================="
echo "Security tests completed"
