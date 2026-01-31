package main

import (
	"fmt"
	"regexp"
)

// 复制当前的正则表达式
var PatternWaitingConfirm = regexp.MustCompile(`(?i)(waiting for confirmation|proceed with this plan\?|Do you want to proceed|confirm|^\s*yes/no|\(yes/no\)|\[yes/no\]|[❯►>]\s*\d+\.\s*(Yes|No)|Select an option)`)

// 测试用例
type TestCase struct {
	Name        string
	Input       string
	ShouldMatch bool // 期望是否匹配
	Issue       string
}

func main() {
	testCases := []TestCase{
		// 正常的确认提示 - 应该匹配
		{
			Name:        "正常确认提示 1",
			Input:       "Do you want to proceed? (yes/no)",
			ShouldMatch: true,
			Issue:       "",
		},
		{
			Name:        "正常确认提示 2",
			Input:       "Proceed with this plan?",
			ShouldMatch: true,
			Issue:       "",
		},
		{
			Name:        "正常确认提示 3",
			Input:       "❯ 1. Yes\n  2. No",
			ShouldMatch: true,
			Issue:       "",
		},

		// 问题场景 - 不应该匹配但可能误判
		{
			Name:        "误判场景 1: 普通句子包含 confirm",
			Input:       "This test will confirm the functionality works correctly.",
			ShouldMatch: false,
			Issue:       "问题 1: 单词 confirm 导致误判",
		},
		{
			Name:        "误判场景 2: 文档中的 confirm",
			Input:       "The system will automatically confirm receipt of your message.",
			ShouldMatch: false,
			Issue:       "问题 1: confirm 在普通句子中误判",
		},
		{
			Name:        "误判场景 3: Select 误判",
			Input:       "Please select an option from the dropdown menu.",
			ShouldMatch: false,
			Issue:       "问题 1: Select an option 过于通用",
		},
		{
			Name:        "误判场景 4: confirm 作为动词",
			Input:       "We need to confirm your identity before proceeding.",
			ShouldMatch: false,
			Issue:       "问题 1: confirm 作为普通动词误判",
		},

		// 边界情况
		{
			Name:        "边界情况 1: 空字符串",
			Input:       "",
			ShouldMatch: false,
			Issue:       "",
		},
		{
			Name:        "边界情况 2: 只有空格",
			Input:       "   ",
			ShouldMatch: false,
			Issue:       "",
		},
		{
			Name:        "边界情况 3: 特殊字符",
			Input:       "!@#$%^&*()",
			ShouldMatch: false,
			Issue:       "",
		},

		// 应该匹配但可能漏检的场景
		{
			Name:        "漏检场景 1: (Y/N) 格式",
			Input:       "Continue? (Y/N)",
			ShouldMatch: true,
			Issue:       "问题 6: 不支持 (Y/N) 格式",
		},
		{
			Name:        "漏检场景 2: Press Enter",
			Input:       "Press Enter to continue",
			ShouldMatch: true,
			Issue:       "问题 6: 不支持 Enter 确认",
		},
		{
			Name:        "漏检场景 3: 数字选择",
			Input:       "Enter a number (1-5):",
			ShouldMatch: true,
			Issue:       "问题 6: 不支持数字选择",
		},
	}

	fmt.Println("========================================")
	fmt.Println("确认机制正则表达式测试")
	fmt.Println("========================================")
	fmt.Println()

	passCount := 0
	failCount := 0
	issueCount := 0

	for _, tc := range testCases {
		matched := PatternWaitingConfirm.MatchString(tc.Input)

		// 判断测试是否通过
		passed := (matched == tc.ShouldMatch)

		status := "✓ PASS"
		if !passed {
			status = "✗ FAIL"
			failCount++
			if tc.Issue != "" {
				issueCount++
			}
		} else {
			passCount++
		}

		fmt.Printf("%s | %s\n", status, tc.Name)
		if !passed {
			fmt.Printf("   期望: %v, 实际: %v\n", tc.ShouldMatch, matched)
			fmt.Printf("   输入: %q\n", tc.Input)
			if tc.Issue != "" {
				fmt.Printf("   🐛 %s\n", tc.Issue)
			}
		}
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("测试总结")
	fmt.Println("========================================")
	fmt.Printf("通过: %d\n", passCount)
	fmt.Printf("失败: %d\n", failCount)
	fmt.Printf("发现问题: %d\n", issueCount)
	fmt.Println()

	if failCount > 0 {
		fmt.Println("⚠️  发现鲁棒性问题，需要优化正则表达式！")
	} else {
		fmt.Println("✅ 所有测试通过！")
	}
}
