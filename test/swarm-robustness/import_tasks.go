package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/state"
)

type TestScenario struct {
	Name  string             `json:"name"`
	Tasks []models.Task      `json:"tasks"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run import_tasks.go <scenario_file.json>")
		os.Exit(1)
	}

	scenarioFile := os.Args[1]
	
	// 读取场景文件
	data, err := ioutil.ReadFile(scenarioFile)
	if err != nil {
		log.Fatalf("读取场景文件失败: %v", err)
	}

	var scenario TestScenario
	if err := json.Unmarshal(data, &scenario); err != nil {
		log.Fatalf("解析场景文件失败: %v", err)
	}

	fmt.Printf("加载测试场景: %s\n", scenario.Name)
	fmt.Printf("任务数量: %d\n\n", len(scenario.Tasks))

	// 连接任务队列
	taskQueue, err := state.NewTaskQueue("~/.claude-swarm/tasks.json")
	if err != nil {
		log.Fatalf("创建任务队列失败: %v", err)
	}
	defer taskQueue.Close()

	// 添加任务
	for i, task := range scenario.Tasks {
		task.CreatedAt = time.Now()
		task.UpdatedAt = time.Now()
		task.Status = models.TaskStatusPending
		
		// 设置默认重试次数
		if task.MaxRetries == 0 {
			task.MaxRetries = 3
		}

		if err := taskQueue.AddTask(&task); err != nil {
			log.Printf("❌ 添加任务 %s 失败: %v", task.ID, err)
			continue
		}
		
		fmt.Printf("✓ [%d/%d] 添加任务: %s (优先级: %d)\n", 
			i+1, len(scenario.Tasks), task.ID, task.Priority)
		
		if len(task.Dependencies) > 0 {
			fmt.Printf("  依赖: %v\n", task.Dependencies)
		}
	}

	fmt.Printf("\n✅ 测试场景 '%s' 导入完成！\n", scenario.Name)
}
