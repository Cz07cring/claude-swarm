package main

import (
	"fmt"
	"strings"
)

// 当前的危险关键词（从 patterns.go 复制）
var DangerKeywords = []string{
	// File operations
	"delete",
	"remove",
	"rm -rf",
	"rm -r",
	"truncate",
	"unlink",
	"destroy",

	// Git dangerous operations
	"git reset --hard",
	"git push --force",
	"git push -f",
	"git clean -f",
	"git clean -fd",
	"git branch -D",
	"git rebase --hard",
	"--force-with-lease",

	// Database operations
	"drop table",
	"drop database",
	"truncate table",
	"delete from",

	// System operations
	"kill -9",
	"killall",
	"shutdown",
	"reboot",
	"format",
	"fdisk",
	"mkfs",

	// General danger indicators
	"force",
	"destructive",
	"purge",
	"wipe",
	"erase",
	"overwrite",
}

// 测试场景
type DangerTest struct {
	Name        string
	Command     string
	ShouldBlock bool
	Missing     string
}

func main() {
	testCases := []DangerTest{
		// 应该被检测的危险操作
		{
			Name:        "已覆盖: rm -rf",
			Command:     "rm -rf /tmp/data",
			ShouldBlock: true,
			Missing:     "",
		},
		{
			Name:        "已覆盖: git push --force",
			Command:     "git push --force origin main",
			ShouldBlock: true,
			Missing:     "",
		},
		{
			Name:        "已覆盖: DROP TABLE",
			Command:     "DROP TABLE users;",
			ShouldBlock: true,
			Missing:     "",
		},

		// 缺失的危险操作
		{
			Name:        "缺失: sudo rm",
			Command:     "sudo rm /etc/passwd",
			ShouldBlock: true,
			Missing:     "sudo rm",
		},
		{
			Name:        "缺失: chmod 777",
			Command:     "chmod 777 /var/www",
			ShouldBlock: true,
			Missing:     "chmod 777",
		},
		{
			Name:        "缺失: chown -R",
			Command:     "chown -R root:root /",
			ShouldBlock: true,
			Missing:     "chown",
		},
		{
			Name:        "缺失: DROP USER",
			Command:     "DROP USER admin@localhost;",
			ShouldBlock: true,
			Missing:     "drop user",
		},
		{
			Name:        "缺失: ALTER TABLE DROP",
			Command:     "ALTER TABLE users DROP COLUMN password;",
			ShouldBlock: true,
			Missing:     "alter table",
		},
		{
			Name:        "缺失: REVOKE",
			Command:     "REVOKE ALL PRIVILEGES ON *.* FROM 'user'@'localhost';",
			ShouldBlock: true,
			Missing:     "revoke",
		},
		{
			Name:        "缺失: dd 命令",
			Command:     "dd if=/dev/zero of=/dev/sda",
			ShouldBlock: true,
			Missing:     "dd if=",
		},
		{
			Name:        "缺失: > 覆盖重要文件",
			Command:     "echo '' > /etc/hosts",
			ShouldBlock: true,
			Missing:     "> /etc/",
		},
		{
			Name:        "缺失: :(){ :|:& };: fork bomb",
			Command:     ":(){ :|:& };:",
			ShouldBlock: true,
			Missing:     "fork bomb pattern",
		},

		// 安全操作 - 不应该阻止
		{
			Name:        "安全: 读取文件",
			Command:     "cat /etc/hosts",
			ShouldBlock: false,
			Missing:     "",
		},
		{
			Name:        "安全: 列出文件",
			Command:     "ls -la /tmp",
			ShouldBlock: false,
			Missing:     "",
		},
	}

	fmt.Println("========================================")
	fmt.Println("危险关键词覆盖率测试")
	fmt.Println("========================================")
	fmt.Println()

	blocked := 0
	missed := 0
	falsePositive := 0

	for _, tc := range testCases {
		commandLower := strings.ToLower(tc.Command)
		isBlocked := false

		// 检查是否被任何关键词阻止
		for _, keyword := range DangerKeywords {
			if strings.Contains(commandLower, keyword) {
				isBlocked = true
				break
			}
		}

		if tc.ShouldBlock {
			if isBlocked {
				fmt.Printf("✓ BLOCKED | %s\n", tc.Name)
				blocked++
			} else {
				fmt.Printf("✗ MISSED  | %s\n", tc.Name)
				fmt.Printf("   命令: %s\n", tc.Command)
				fmt.Printf("   🐛 缺失关键词: %s\n", tc.Missing)
				missed++
			}
		} else {
			if !isBlocked {
				fmt.Printf("✓ ALLOWED | %s\n", tc.Name)
			} else {
				fmt.Printf("✗ FALSE+  | %s\n", tc.Name)
				fmt.Printf("   命令: %s\n", tc.Command)
				falsePositive++
			}
		}
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("测试总结")
	fmt.Println("========================================")
	fmt.Printf("正确阻止: %d\n", blocked)
	fmt.Printf("漏检危险: %d\n", missed)
	fmt.Printf("误报安全: %d\n", falsePositive)
	fmt.Printf("总测试数: %d\n", len(testCases))
	fmt.Println()

	coverage := float64(blocked) / float64(blocked+missed) * 100
	fmt.Printf("危险操作覆盖率: %.1f%%\n", coverage)
	fmt.Println()

	if missed > 0 {
		fmt.Printf("⚠️  发现 %d 个缺失的危险关键词！\n", missed)
		fmt.Println("\n建议添加的关键词:")
		fmt.Println("  - sudo rm")
		fmt.Println("  - chmod 777")
		fmt.Println("  - chown")
		fmt.Println("  - drop user")
		fmt.Println("  - alter table")
		fmt.Println("  - revoke")
		fmt.Println("  - dd if=")
		fmt.Println("  - > /etc/")
	} else {
		fmt.Println("✅ 危险关键词覆盖率良好！")
	}
}
