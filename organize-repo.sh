#!/bin/bash
# 优化仓库目录结构

set -e

echo "📁 开始优化目录结构..."
echo ""

# 创建新的目录结构
echo "1. 创建目录结构..."
mkdir -p docs/guides
mkdir -p docs/reports
mkdir -p docs/tui
mkdir -p scripts/tests
mkdir -p scripts/tools
mkdir -p .archive

echo "✅ 目录创建完成"
echo ""

# 移动用户指南类文档
echo "2. 整理用户指南..."
mv -v USER_GUIDE.md docs/guides/ 2>/dev/null || true
mv -v CONFIG_GUIDE.md docs/guides/ 2>/dev/null || true
mv -v SWARM_DEMO_GUIDE.md docs/guides/ 2>/dev/null || true

# 移动测试报告
echo ""
echo "3. 整理测试报告..."
mv -v P0_TEST_REPORT.md docs/reports/ 2>/dev/null || true
mv -v P0_COMPLETION_SUMMARY.md docs/reports/ 2>/dev/null || true
mv -v TEST_RESULTS.md docs/reports/ 2>/dev/null || true
mv -v TUI_TEST_REPORT.md docs/reports/ 2>/dev/null || true
mv -v COMPLEX_TASK_TEST_REPORT.md docs/reports/ 2>/dev/null || true
mv -v CONFIRMATION_ISSUES_REPORT.md docs/reports/ 2>/dev/null || true

# 移动改进和修复文档
echo ""
echo "4. 整理改进文档..."
mv -v CODE_QUALITY_IMPROVEMENTS.md docs/reports/ 2>/dev/null || true
mv -v IMPROVEMENTS_SUMMARY.md docs/reports/ 2>/dev/null || true
mv -v IMPLEMENTATION_SUMMARY.md docs/reports/ 2>/dev/null || true
mv -v CONFIGURATION_UPDATE.md docs/reports/ 2>/dev/null || true
mv -v HIGH_PRIORITY_FIXES.md docs/reports/ 2>/dev/null || true
mv -v MEDIUM_PRIORITY_FIXES.md docs/reports/ 2>/dev/null || true
mv -v QUICK_FIXES.md docs/reports/ 2>/dev/null || true
mv -v FIXES.md docs/reports/ 2>/dev/null || true
mv -v PROGRESS.md docs/reports/ 2>/dev/null || true
mv -v ISSUES_DETECTED.md docs/reports/ 2>/dev/null || true
mv -v FINAL_REPORT.md docs/reports/ 2>/dev/null || true

# 移动 TUI 相关文档
echo ""
echo "5. 整理 TUI 文档..."
mv -v TUI_BUG_FIXES.md docs/tui/ 2>/dev/null || true
mv -v TUI_DEMO.md docs/tui/ 2>/dev/null || true
mv -v TUI_OPTIMIZATION_SUMMARY.md docs/tui/ 2>/dev/null || true
mv -v TUI_TEST_PLAN.md docs/tui/ 2>/dev/null || true
mv -v TUI_UX_IMPROVEMENTS.md docs/tui/ 2>/dev/null || true

# 移动测试脚本
echo ""
echo "6. 整理测试脚本..."
mv -v test-*.sh scripts/tests/ 2>/dev/null || true
mv -v *-test.sh scripts/tests/ 2>/dev/null || true
mv -v comprehensive-validation.sh scripts/tests/ 2>/dev/null || true
mv -v quick-validation.sh scripts/tests/ 2>/dev/null || true
mv -v run-full-test.sh scripts/tests/ 2>/dev/null || true

# 移动工具脚本
echo ""
echo "7. 整理工具脚本..."
mv -v build.sh scripts/tools/ 2>/dev/null || true
mv -v demo.sh scripts/tools/ 2>/dev/null || true
mv -v swarm-demo.sh scripts/tools/ 2>/dev/null || true
mv -v reset-tasks.sh scripts/tools/ 2>/dev/null || true
mv -v create-stress-test-data.sh scripts/tools/ 2>/dev/null || true
mv -v analyze-tui-issues.sh scripts/tools/ 2>/dev/null || true
mv -v ux-analysis.sh scripts/tools/ 2>/dev/null || true
mv -v deep-test-tui.sh scripts/tools/ 2>/dev/null || true
mv -v automated-tui-test.sh scripts/tools/ 2>/dev/null || true

# 归档临时文件
echo ""
echo "8. 归档临时文件..."
mv -v *.log .archive/ 2>/dev/null || true
mv -v test-worktree-*.txt .archive/ 2>/dev/null || true
mv -v test-*.go .archive/ 2>/dev/null || true
mv -v swarm .archive/ 2>/dev/null || echo "swarm 二进制不存在，跳过"

echo ""
echo "✅ 目录结构优化完成！"
echo ""
echo "📊 新的目录结构："
echo ""
tree -L 2 -I 'bin|pkg|internal|cmd|.archive|.git' .
