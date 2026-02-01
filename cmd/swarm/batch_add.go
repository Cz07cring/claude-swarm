package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/state"
)

var batchAddCmd = &cobra.Command{
	Use:   "batch-add",
	Short: "批量添加任务到队列",
	Long: `从文件、stdin 或交互式模式批量添加任务。

文件格式（每行一个任务）:
  描述文本 | priority:8 | depends:task-1,task-2 | max-retries:5

示例:
  # 从文件批量添加
  swarm batch-add --file tasks.txt

  # 从 stdin
  cat tasks.txt | swarm batch-add --stdin

  # 交互式模式（连续输入，空行结束）
  swarm batch-add --interactive`,
	Run: runBatchAdd,
}

var (
	batchFile        string
	batchStdin       bool
	batchInteractive bool
)

func init() {
	rootCmd.AddCommand(batchAddCmd)

	batchAddCmd.Flags().StringVarP(&batchFile, "file", "f", "", "从文件读取任务")
	batchAddCmd.Flags().BoolVar(&batchStdin, "stdin", false, "从标准输入读取任务")
	batchAddCmd.Flags().BoolVarP(&batchInteractive, "interactive", "i", false, "交互式模式（连续输入，空行结束）")
	batchAddCmd.Flags().StringVar(&taskQueuePath, "queue", "~/.claude-swarm/tasks.json", "任务队列文件路径")
}

func runBatchAdd(cmd *cobra.Command, args []string) {
	// 验证参数
	modeCount := 0
	if batchFile != "" {
		modeCount++
	}
	if batchStdin {
		modeCount++
	}
	if batchInteractive {
		modeCount++
	}

	if modeCount == 0 {
		log.Fatal("❌ 请指定输入模式: --file, --stdin, 或 --interactive")
	}
	if modeCount > 1 {
		log.Fatal("❌ 只能指定一种输入模式")
	}

	// 初始化任务队列
	taskQueue, err := state.NewTaskQueue(expandPath(taskQueuePath))
	if err != nil {
		log.Fatalf("❌ 无法打开任务队列: %v", err)
	}
	defer taskQueue.Close()

	// 读取任务行
	var lines []string
	if batchFile != "" {
		lines, err = readLinesFromFile(batchFile)
		if err != nil {
			log.Fatalf("❌ 读取文件失败: %v", err)
		}
	} else if batchStdin {
		lines, err = readLinesFromStdin()
		if err != nil {
			log.Fatalf("❌ 从 stdin 读取失败: %v", err)
		}
	} else if batchInteractive {
		lines, err = readLinesInteractive()
		if err != nil {
			log.Fatalf("❌ 交互式输入失败: %v", err)
		}
	}

	// 解析并添加任务
	added := 0
	failed := 0

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析任务
		task, err := parseTaskLine(line)
		if err != nil {
			fmt.Printf("❌ 第 %d 行解析失败: %v\n", i+1, err)
			fmt.Printf("   内容: %s\n", line)
			failed++
			continue
		}

		// 验证依赖
		if len(task.Dependencies) > 0 {
			if err := validateDependencies(taskQueue, task.Dependencies); err != nil {
				fmt.Printf("⚠️  第 %d 行依赖验证失败: %v\n", i+1, err)
				fmt.Printf("   将继续添加，但任务可能被阻塞\n")
			}
		}

		// 添加到队列
		if err := taskQueue.AddTask(task); err != nil {
			fmt.Printf("❌ 第 %d 行添加失败: %v\n", i+1, err)
			failed++
			continue
		}

		fmt.Printf("✅ 已添加: %s - %s (优先级: %d)\n", task.ID, task.Description, task.Priority)
		added++
	}

	// 总结
	fmt.Println()
	fmt.Println(strings.Repeat("━", 60))
	fmt.Printf("📊 批量添加完成: 成功 %d 个，失败 %d 个\n", added, failed)
}

// parseTaskLine parses a task line in format: "description | key:value | key:value"
func parseTaskLine(line string) (*models.Task, error) {
	parts := strings.Split(line, "|")

	if len(parts) == 0 {
		return nil, fmt.Errorf("空行")
	}

	// 第一部分是描述
	description := strings.TrimSpace(parts[0])
	if description == "" {
		return nil, fmt.Errorf("任务描述不能为空")
	}

	// 默认值
	task := &models.Task{
		Description: description,
		Status:      models.TaskStatusPending,
		Priority:    5,
		MaxRetries:  3,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 解析其他参数
	for i := 1; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])
		if part == "" {
			continue
		}

		// 分割 key:value
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("无效的参数格式: %s", part)
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "priority", "p":
			priority, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("无效的优先级: %s", value)
			}
			if priority < 1 || priority > 10 {
				return nil, fmt.Errorf("优先级必须在 1-10 之间: %d", priority)
			}
			task.Priority = priority

		case "depends", "dependencies", "d":
			deps := strings.Split(value, ",")
			for i, dep := range deps {
				deps[i] = strings.TrimSpace(dep)
			}
			task.Dependencies = deps

		case "max-retries", "retries", "r":
			retries, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("无效的重试次数: %s", value)
			}
			if retries < 0 {
				return nil, fmt.Errorf("重试次数不能为负数: %d", retries)
			}
			task.MaxRetries = retries

		case "id":
			task.ID = value

		default:
			return nil, fmt.Errorf("未知参数: %s", key)
		}
	}

	return task, nil
}

// readLinesFromFile reads lines from a file
func readLinesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// readLinesFromStdin reads lines from stdin
func readLinesFromStdin() ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// readLinesInteractive reads lines interactively until empty line
func readLinesInteractive() ([]string, error) {
	fmt.Println("📝 交互式批量添加任务")
	fmt.Println("每行输入一个任务，格式: 描述 | priority:X | depends:task-id")
	fmt.Println("输入空行结束")
	fmt.Println(strings.Repeat("━", 60))

	var lines []string
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			break
		}

		lines = append(lines, line)
	}

	return lines, scanner.Err()
}
