package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/executor"
)

func main() {
	log.SetFlags(log.Ltime)

	// Get current directory
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	fmt.Println("🧪 Claude Executor 测试 (免费方案)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Create executor
	exec := executor.NewClaudeExecutor(workDir)

	// Test tasks
	tasks := []*models.Task{
		{
			ID:          "test-1",
			Description: "创建一个名为 hello.txt 的文件，内容是 'Hello from Swarm!'",
			Status:      models.TaskStatusPending,
		},
		{
			ID:          "test-2",
			Description: "创建一个 Go 程序 simple.go，包含 main 函数输出 'Claude Swarm works!'",
			Status:      models.TaskStatusPending,
		},
	}

	for i, task := range tasks {
		fmt.Printf("\n📝 测试 %d: %s\n", i+1, task.Description)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Execute with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		err = exec.ExecuteTask(ctx, task)
		cancel()

		if err != nil {
			fmt.Printf("❌ 失败: %v\n", err)
		} else {
			fmt.Printf("✅ 成功!\n")
		}
	}

	// Check created files
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 检查创建的文件:")
	fmt.Println()

	files := []string{"hello.txt", "simple.go"}
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			content, _ := os.ReadFile(file)
			fmt.Printf("✅ %s (%.2f KB)\n", file, float64(info.Size())/1024)
			fmt.Printf("   内容预览: %s\n\n", string(content)[:min(100, len(content))])
		} else {
			fmt.Printf("❌ %s (未创建)\n\n", file)
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🎉 测试完成!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
