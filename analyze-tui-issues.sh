#!/bin/bash
# 分析 TUI 代码中的潜在问题

echo "🔍 分析 TUI 代码质量..."
echo ""

check_issue() {
    local issue_name="$1"
    local file_path="$2"
    local search_pattern="$3"
    local severity="$4"  # HIGH, MEDIUM, LOW

    local icon=""
    case $severity in
        HIGH)    icon="🔴" ;;
        MEDIUM)  icon="🟡" ;;
        LOW)     icon="🟢" ;;
    esac

    if grep -n "$search_pattern" "$file_path" > /dev/null 2>&1; then
        echo "$icon [$severity] $issue_name"
        grep -n "$search_pattern" "$file_path" | head -3 | sed 's/^/     /'
        echo ""
        return 1
    else
        return 0
    fi
}

echo "📊 检查1: 数组越界保护"
echo "----------------------------------------"

# 检查是否有适当的边界检查
if grep -n "if.*< 0" pkg/tui/*.go | grep -v "//"; then
    echo "✅ 发现边界检查"
else
    echo "⚠️  可能缺少边界检查"
fi
echo ""

echo "📊 检查2: 空指针保护"
echo "----------------------------------------"

# 检查 nil 检查
if grep -n "if.*== nil" pkg/tui/*.go | head -5; then
    echo "✅ 发现 nil 检查"
else
    echo "⚠️  可能缺少 nil 检查"
fi
echo ""

echo "📊 检查3: 除零保护"
echo "----------------------------------------"

# 查找除法运算
if grep -n " / " pkg/tui/*.go | grep -v "//"; then
    echo "⚠️  发现除法运算，需要检查除零保护:"
    grep -n " / " pkg/tui/*.go | grep -v "//" | head -5
else
    echo "✅ 未发现明显的除法运算"
fi
echo ""

echo "📊 检查4: 字符串截断安全性"
echo "----------------------------------------"

# 检查字符串截断
if grep -n "\[:.*\]" pkg/tui/*.go | head -5; then
    echo "⚠️  发现字符串截断，需要检查边界:"
    grep -n "\[:.*\]" pkg/tui/*.go | head -5
    echo ""
fi

echo "📊 检查5: 性能问题"
echo "----------------------------------------"

# 检查可能的性能问题
echo "检查循环嵌套..."
nested_loops=$(grep -n "for.*{" pkg/tui/*.go | wc -l)
echo "  发现 $nested_loops 个循环"

# 检查字符串拼接
string_concat=$(grep -n "+=" pkg/tui/*.go | grep "string\|String" | wc -l)
echo "  发现 $string_concat 个字符串拼接操作"

if [ "$string_concat" -gt 20 ]; then
    echo "  ⚠️  大量字符串拼接可能影响性能，建议使用 strings.Builder"
fi
echo ""

echo "✅ 分析完成"
