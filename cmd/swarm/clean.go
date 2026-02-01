package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/claude-swarm/pkg/state"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "清理任务队列和相关资源",
	Long: `清理任务队列中的已完成、失败或所有任务。

示例:
  # 清理完成的任务
  swarm clean --completed

  # 清理失败的任务
  swarm clean --failed

  # 清理所有任务（危险！需要确认）
  swarm clean --all

  # 跳过确认提示
  swarm clean --completed --force`,
	Run: runClean,
}

var (
	cleanCompleted bool
	cleanFailed    bool
	cleanAll       bool
	cleanForce     bool
)

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().BoolVar(&cleanCompleted, "completed", false, "清理已完成的任务")
	cleanCmd.Flags().BoolVar(&cleanFailed, "failed", false, "清理失败的任务")
	cleanCmd.Flags().BoolVar(&cleanAll, "all", false, "清理所有任务（危险操作）")
	cleanCmd.Flags().BoolVarP(&cleanForce, "force", "f", false, "跳过确认提示")
	cleanCmd.Flags().StringVar(&taskQueuePath, "queue", "~/.claude-swarm/tasks.json", "任务队列文件路径")
}

func runClean(cmd *cobra.Command, args []string) {
	// 验证参数
	modeCount := 0
	if cleanCompleted {
		modeCount++
	}
	if cleanFailed {
		modeCount++
	}
	if cleanAll {
		modeCount++
	}

	if modeCount == 0 {
		log.Fatal("❌ 请指定清理模式: --completed, --failed, 或 --all")
	}
	if modeCount > 1 {
		log.Fatal("❌ 只能指定一种清理模式")
	}

	// 初始化任务队列
	taskQueue, err := state.NewTaskQueue(expandPath(taskQueuePath))
	if err != nil {
		log.Fatalf("❌ 无法打开任务队列: %v", err)
	}
	defer taskQueue.Close()

	// 确定操作类型
	var operation string
	var count int

	if cleanCompleted {
		operation = "已完成"
		tasks := taskQueue.ListTasks()
		for _, task := range tasks {
			if task.Status == "completed" {
				count++
			}
		}
	} else if cleanFailed {
		operation = "失败"
		tasks := taskQueue.ListTasks()
		for _, task := range tasks {
			if task.Status == "failed" {
				count++
			}
		}
	} else if cleanAll {
		operation = "所有"
		count = len(taskQueue.ListTasks())
	}

	if count == 0 {
		fmt.Printf("✅ 没有需要清理的%s任务\n", operation)
		return
	}

	// 确认提示
	if !cleanForce {
		fmt.Printf("⚠️  将要删除 %d 个%s任务，此操作不可撤销！\n", count, operation)
		fmt.Print("是否继续？ (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("❌ 读取输入失败: %v", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			fmt.Println("❌ 操作已取消")
			return
		}
	}

	// 执行清理
	var removed int
	if cleanCompleted {
		removed, err = taskQueue.ClearCompleted()
	} else if cleanFailed {
		removed, err = taskQueue.ClearFailed()
	} else if cleanAll {
		removed, err = taskQueue.ClearAll()
	}

	if err != nil {
		log.Fatalf("❌ 清理失败: %v", err)
	}

	fmt.Printf("✅ 成功清理 %d 个%s任务\n", removed, operation)
}
