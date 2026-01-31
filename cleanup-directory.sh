#!/bin/bash

# Directory Cleanup Script
# 按照 CLEANUP_PLAN.md 整理项目目录结构

set -e  # 遇到错误立即退出

echo "🧹 Starting directory cleanup..."
echo ""

# 1. 创建缺失的目录
echo "📁 Creating missing directories..."
mkdir -p logs
mkdir -p test/coverage
mkdir -p test/fixtures
mkdir -p test/integration
mkdir -p scripts/test
mkdir -p scripts/build
mkdir -p scripts/utils
mkdir -p docs/reports/test
mkdir -p docs/reports/bugfix
mkdir -p docs/reports/improvements
echo "✓ Directories created"
echo ""

# 2. 移动脚本文件
echo "📦 Moving scripts..."
[ -f add_test_tasks.sh ] && mv add_test_tasks.sh scripts/test/ && echo "  ✓ add_test_tasks.sh → scripts/test/"
[ -f start_test_swarm.sh ] && mv start_test_swarm.sh scripts/test/ && echo "  ✓ start_test_swarm.sh → scripts/test/"
[ -f run-all-tests.sh ] && mv run-all-tests.sh scripts/test/ && echo "  ✓ run-all-tests.sh → scripts/test/"
[ -f organize-repo.sh ] && mv organize-repo.sh scripts/utils/ && echo "  ✓ organize-repo.sh → scripts/utils/"
echo ""

# 3. 移动测试和覆盖率文件
echo "🧪 Moving test files..."
[ -f coverage.out ] && mv coverage.out test/coverage/ && echo "  ✓ coverage.out → test/coverage/"
echo ""

# 4. 移动日志文件
echo "📝 Moving log files..."
[ -f swarm.log ] && mv swarm.log logs/ && echo "  ✓ swarm.log → logs/"
find . -maxdepth 1 -name "*.log" -type f -exec mv {} logs/ \; 2>/dev/null && echo "  ✓ Other log files → logs/" || true
echo ""

# 5. 移动二进制文件
echo "⚙️  Moving binaries..."
[ -f swarm ] && mv swarm bin/ && echo "  ✓ swarm → bin/" || echo "  ℹ swarm not found (may be already built)"
echo ""

# 6. 移动根目录的报告文件到docs/reports
echo "📊 Moving reports from root..."
[ -f TEST_RESULTS.md ] && mv TEST_RESULTS.md docs/reports/test/ && echo "  ✓ TEST_RESULTS.md → docs/reports/test/"
[ -f UNIT_TEST_SUMMARY.md ] && mv UNIT_TEST_SUMMARY.md docs/reports/test/ && echo "  ✓ UNIT_TEST_SUMMARY.md → docs/reports/test/"
[ -f DIRECTORY_STRUCTURE.md ] && mv DIRECTORY_STRUCTURE.md docs/ && echo "  ✓ DIRECTORY_STRUCTURE.md → docs/"
echo ""

# 7. 整理 docs/reports/ 中的现有报告
echo "📋 Organizing existing reports..."
cd docs/reports/

# 测试报告
for file in P0_TEST_REPORT.md P0_COMPLETION_SUMMARY.md TUI_TEST_REPORT.md COMPLEX_TASK_TEST_REPORT.md; do
    [ -f "$file" ] && mv "$file" test/ && echo "  ✓ $file → test/"
done

# Bug修复报告
for file in CONFIRMATION_ISSUES_REPORT.md HIGH_PRIORITY_FIXES.md MEDIUM_PRIORITY_FIXES.md QUICK_FIXES.md FIXES.md; do
    [ -f "$file" ] && mv "$file" bugfix/ && echo "  ✓ $file → bugfix/"
done

# 改进报告
for file in CODE_QUALITY_IMPROVEMENTS.md IMPROVEMENTS_SUMMARY.md CONFIGURATION_UPDATE.md; do
    [ -f "$file" ] && mv "$file" improvements/ && echo "  ✓ $file → improvements/"
done

cd ../..
echo ""

# 8. 删除临时文件
echo "🗑️  Removing temporary files..."
[ -f hello.txt ] && rm -f hello.txt && echo "  ✓ Removed hello.txt"
[ -f test2.txt ] && rm -f test2.txt && echo "  ✓ Removed test2.txt"
[ -f simple.go ] && rm -f simple.go && echo "  ✓ Removed simple.go"
echo ""

# 9. 更新 .gitignore
echo "📝 Updating .gitignore..."
cat > .gitignore << 'EOF'
# Binaries
/bin/swarm
/swarm
*.exe

# Test coverage
/test/coverage/
*.out
coverage.html
coverage.xml

# Logs
/logs/
*.log

# Temporary files
*.tmp
*.temp
hello.txt
test*.txt
simple.go

# OS files
.DS_Store
Thumbs.db
*~

# IDE
.vscode/
.idea/
*.swp
*.swo
*.swn

# Build artifacts
dist/
build/

# Go
vendor/

# Config (keep example)
config.yaml
!config.yaml.example

# Archive and worktrees
.archive/
.worktrees/
EOF
echo "  ✓ .gitignore updated"
echo ""

# 10. 创建文档索引
echo "📚 Creating documentation index..."
cat > docs/README.md << 'EOF'
# Claude Swarm Documentation

## 📖 User Guides

### Getting Started
- [Quick Start Guide](guides/quickstart.md)
- [Getting Started](guides/GETTING_STARTED.md)
- [User Guide](guides/USER_GUIDE.md)
- [Configuration Guide](guides/CONFIG_GUIDE.md)

### Advanced Topics
- [Gemini Setup](GEMINI_SETUP.md)
- [Swarm Demo Guide](guides/SWARM_DEMO_GUIDE.md)
- [MVP Guide](guides/mvp-guide.md)
- [Confirmation System](guides/confirmation-system.md)

## 🏗️ Architecture

- [Full Plan](architecture/full-plan.md)
- [Directory Structure](DIRECTORY_STRUCTURE.md)
- [Project Summary](PROJECT_SUMMARY.md)

## 📊 Reports

### Test Reports
- [Unit Test Summary](reports/test/UNIT_TEST_SUMMARY.md)
- [P0 Test Report](reports/test/P0_TEST_REPORT.md)
- [P0 Completion Summary](reports/test/P0_COMPLETION_SUMMARY.md)
- [TUI Test Report](reports/test/TUI_TEST_REPORT.md)
- [Complex Task Test Report](reports/test/COMPLEX_TASK_TEST_REPORT.md)

### Bug Fixes
- [Confirmation Issues Report](reports/bugfix/CONFIRMATION_ISSUES_REPORT.md)
- [High Priority Fixes](reports/bugfix/HIGH_PRIORITY_FIXES.md)
- [Medium Priority Fixes](reports/bugfix/MEDIUM_PRIORITY_FIXES.md)
- [Quick Fixes](reports/bugfix/QUICK_FIXES.md)

### Improvements
- [Code Quality Improvements](reports/improvements/CODE_QUALITY_IMPROVEMENTS.md)
- [Configuration Update](reports/improvements/CONFIGURATION_UPDATE.md)
- [Improvements Summary](reports/improvements/IMPROVEMENTS_SUMMARY.md)

## 🔧 Development

### Testing
- [Bug Fix Documentation](BUGFIX.md)
- [Test Report](TEST_REPORT.md)

### TUI Development
- [TUI Development](TUI_DEVELOPMENT.md)
- [TUI Monitor](TUI_MONITOR.md)
- [TUI Bug Fixes](tui/TUI_BUG_FIXES.md)
- [TUI Demo](tui/TUI_DEMO.md)
- [TUI Optimization](tui/TUI_OPTIMIZATION_SUMMARY.md)
- [TUI UX Improvements](tui/TUI_UX_IMPROVEMENTS.md)

## 📝 Other Documents

- [Stop Behavior Improvement](STOP_BEHAVIOR_IMPROVEMENT.md)
- [Final Report](reports/FINAL_REPORT.md)
- [Implementation Summary](reports/IMPLEMENTATION_SUMMARY.md)
- [Issues Detected](reports/ISSUES_DETECTED.md)
EOF
echo "  ✓ docs/README.md created"
echo ""

# 11. 显示清理后的目录结构
echo "✨ Cleanup complete!"
echo ""
echo "📂 New directory structure:"
tree -L 2 -I '.git|.archive|.worktrees' || ls -la

echo ""
echo "✅ All done! Please review the changes with:"
echo "   git status"
echo ""
echo "📌 Next steps:"
echo "   1. Review changes: git status"
echo "   2. Test the project: go test ./..."
echo "   3. Commit changes: git add . && git commit -m 'Reorganize project structure'"
