#!/bin/bash

echo "🧹 Cleaning up remaining test files..."

# 移动测试代码文件到 test/ 目录
mkdir -p test/manual

echo "📦 Moving test files..."
[ -f test_dag_cycle.go ] && mv test_dag_cycle.go test/manual/ && echo "  ✓ test_dag_cycle.go → test/manual/"
[ -f test_retry_logic.go ] && mv test_retry_logic.go test/manual/ && echo "  ✓ test_retry_logic.go → test/manual/"
[ -f test-ai-decision.go ] && mv test-ai-decision.go test/manual/ && echo "  ✓ test-ai-decision.go → test/manual/"
[ -f test-claude-executor.go ] && mv test-claude-executor.go test/manual/ && echo "  ✓ test-claude-executor.go → test/manual/"
[ -f test-robustness.sh ] && mv test-robustness.sh scripts/test/ && echo "  ✓ test-robustness.sh → scripts/test/"

echo ""
echo "🗑️  Removing test binaries..."
[ -f test-executor ] && rm -f test-executor && echo "  ✓ Removed test-executor"

echo ""
echo "✅ Cleanup complete!"
