package main

import (
	"fmt"
	"log"

	"github.com/yourusername/claude-swarm/pkg/analyzer"
)

func main() {
	log.SetFlags(log.Ltime)

	detector := analyzer.NewDetector()

	// 测试场景
	testCases := []struct {
		name    string
		context string
	}{
		{
			name:    "低风险 - 普通确认",
			context: "Do you want to proceed? (y/n)",
		},
		{
			name:    "中等风险 - 删除构建目录",
			context: "Remove build directory? Execute: rm -rf build/ (y/n)",
		},
		{
			name:    "中等风险 - Git commit",
			context: "Proceed with git commit? (yes/no)",
		},
		{
			name:    "高风险 - Force push (无安全上下文)",
			context: "Execute: git push --force to main branch? (y/n)",
		},
		{
			name:    "高风险 - Force push (有安全上下文)",
			context: "Execute: git push --force to agent-0 branch in worktree? (y/n)",
		},
		{
			name:    "极高风险 - 删除系统目录",
			context: "WARNING: This will execute: rm -rf /etc. Continue? (y/n)",
		},
		{
			name:    "极高风险 - 删除数据库",
			context: "Execute: DROP DATABASE production. This is irreversible! (y/n)",
		},
		{
			name:    "低风险 - 继续操作",
			context: "Press Enter to continue with the installation...",
		},
	}

	fmt.Println("🤖 AI 自主决策系统测试")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	for i, tc := range testCases {
		fmt.Printf("测试 %d: %s\n", i+1, tc.name)
		fmt.Printf("上下文: %s\n", tc.context)

		// 模拟接收到这段输出
		detector.Analyze(tc.context)

		// AI 自主决策
		shouldConfirm, input, reason := detector.ShouldConfirm()

		if shouldConfirm {
			fmt.Printf("✅ 决策: 自动确认\n")
			fmt.Printf("   输入: '%s'\n", input)
			fmt.Printf("   原因: %s\n", reason)
		} else {
			fmt.Printf("🚫 决策: 阻止执行\n")
			fmt.Printf("   原因: %s\n", reason)
		}

		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✓ 测试完成")
}
