package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/state"
)

// OrchestratorBrain AI主脑 - 使用Gemini进行智能决策
type OrchestratorBrain struct {
	client      *genai.Client
	taskQueue   *state.TaskQueue
	context     *ConversationContext
	modelName   string
}

// NewOrchestratorBrain 创建AI主脑
// apiKey如果为空，会从环境变量GEMINI_API_KEY读取
func NewOrchestratorBrain(apiKey string, taskQueue *state.TaskQueue) (*OrchestratorBrain, error) {
	ctx := context.Background()

	// 初始化Gemini客户端（从环境变量读取API Key）
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// 使用Gemini 3 Flash Preview模型（最新）
	modelName := "gemini-3-flash-preview"

	brain := &OrchestratorBrain{
		client:    client,
		taskQueue: taskQueue,
		modelName: modelName,
		context: &ConversationContext{
			Conversations: make([]Message, 0),
			TaskHistory:   make([]*models.Task, 0),
			Decisions:     make([]Decision, 0),
		},
	}

	log.Printf("✓ AI主脑初始化成功 (模型: %s)", modelName)
	return brain, nil
}

// Close 关闭客户端（新版SDK不需要显式关闭）
func (b *OrchestratorBrain) Close() error {
	// 新版Gemini SDK不需要显式关闭
	return nil
}

// AnalyzeRequirement AI分析用户需求
func (b *OrchestratorBrain) AnalyzeRequirement(ctx context.Context, requirement string) (*AnalysisResult, error) {
	log.Printf("🧠 AI主脑开始分析需求...")

	prompt := b.buildAnalysisPrompt(requirement)

	// 调用Gemini API
	result, err := b.client.Models.GenerateContent(
		ctx,
		b.modelName,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Gemini API调用失败: %w", err)
	}

	// 获取响应文本
	responseText := result.Text()

	// 解析JSON响应
	analysisResult, err := b.parseAnalysisResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("解析AI响应失败: %w\n原始响应: %s", err, responseText)
	}

	// 保存到上下文
	b.context.Requirement = requirement
	b.context.AnalysisResult = analysisResult
	b.context.Conversations = append(b.context.Conversations, Message{
		Role:      "user",
		Content:   requirement,
		Timestamp: time.Now(),
	})
	b.context.Conversations = append(b.context.Conversations, Message{
		Role:      "assistant",
		Content:   responseText,
		Timestamp: time.Now(),
	})

	log.Printf("✓ AI分析完成: %d个模块, %d个任务", len(analysisResult.Modules), len(analysisResult.Tasks))
	return analysisResult, nil
}

// buildAnalysisPrompt 构建分析提示词
func (b *OrchestratorBrain) buildAnalysisPrompt(requirement string) string {
	return fmt.Sprintf(`你是一个资深软件架构师和项目经理，负责分析用户需求并拆分成可并行开发的任务。

用户需求：
%s

请按以下步骤分析：

1. **理解需求**：总结需求的核心功能和目标

2. **模块拆分**：将需求拆分成独立的功能模块（3-8个模块）
   - 每个模块应该是独立的功能单元
   - 模块之间的耦合度要低
   - 考虑可并行开发

3. **任务生成**：为每个模块生成具体的开发任务
   - 任务要具体、可执行
   - 每个任务预计30分钟到2小时完成
   - 明确任务描述，让AI agent能理解
   - 任务描述要包含具体要实现的功能，而不仅仅是"设计"或"规划"

4. **依赖分析**：识别任务之间的依赖关系
   - 哪些任务必须先完成
   - 哪些任务可以并行

5. **文件预测**：预测每个任务可能涉及的文件路径

请以JSON格式返回，格式如下：
{
  "summary": "需求概要（一句话）",
  "complexity": "low|medium|high",
  "estimated_time": "预计总时间",
  "modules": [
    {
      "name": "模块名",
      "description": "模块描述",
      "files": ["涉及的文件路径"],
      "priority": 1-10
    }
  ],
  "tasks": [
    {
      "id": "task-001",
      "description": "具体任务描述（给AI agent执行），例如：'创建一个todo.go文件，实现AddTask函数用于添加新任务到数组'",
      "module": "所属模块名",
      "files": ["涉及的文件"],
      "dependencies": ["依赖的任务ID"],
      "priority": 1-10,
      "estimated": "30m|1h|2h"
    }
  ],
  "dependencies": {
    "task-002": ["task-001"],
    "task-003": ["task-001"]
  }
}

重要要求：
- 只返回JSON，不要额外的解释文字
- 不要用markdown代码块包裹JSON
- task ID格式：task-001, task-002...
- 任务描述要清晰具体，让Claude Code agent能直接执行
- 任务描述要包含要创建的文件名和具体要实现的功能
- 考虑Git分支隔离，每个task在独立分支开发`, requirement)
}

// parseAnalysisResponse 解析AI响应
func (b *OrchestratorBrain) parseAnalysisResponse(response string) (*AnalysisResult, error) {
	// 去除可能的markdown代码块标记
	response = cleanJSONResponse(response)

	var result AnalysisResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w\n原始响应: %s", err, response)
	}

	// 生成任务ID（如果没有）
	for i, task := range result.Tasks {
		if task.ID == "" {
			task.ID = fmt.Sprintf("task-%03d", i+1)
		}
	}

	return &result, nil
}

// CreateTasksFromAnalysis 将AI分析结果转换为任务队列
func (b *OrchestratorBrain) CreateTasksFromAnalysis(ctx context.Context, result *AnalysisResult) error {
	log.Printf("📋 创建任务队列: %d个任务", len(result.Tasks))

	for _, taskSpec := range result.Tasks {
		task := &models.Task{
			ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
			Description: taskSpec.Description,
			Status:      models.TaskStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// 添加到任务队列
		if err := b.taskQueue.AddTask(task); err != nil {
			return fmt.Errorf("添加任务失败: %w", err)
		}

		log.Printf("  ✓ %s: %s", task.ID, task.Description)
	}

	return nil
}

// MonitorProgress AI监控所有Agent的进展
func (b *OrchestratorBrain) MonitorProgress(ctx context.Context, agents []*models.AgentStatus) (*ProgressReport, error) {
	// 收集所有任务状态
	tasks := b.taskQueue.ListTasks()

	report := &ProgressReport{
		Timestamp:   time.Now(),
		TotalTasks:  len(tasks),
		AgentStatus: make(map[string]*AgentProgress),
	}

	// 统计任务状态
	for _, task := range tasks {
		switch task.Status {
		case models.TaskStatusCompleted:
			report.CompletedTasks++
		case models.TaskStatusInProgress:
			report.InProgressTasks++
		case models.TaskStatusFailed:
			report.FailedTasks++
		}
	}

	// 计算进度
	if report.TotalTasks > 0 {
		report.OverallProgress = float64(report.CompletedTasks) / float64(report.TotalTasks) * 100
	}

	// 收集Agent状态
	for _, agent := range agents {
		progress := &AgentProgress{
			AgentID:    agent.AgentID,
			State:      agent.State,
			LastUpdate: agent.LastUpdate,
		}

		// 检测卡住
		if time.Since(agent.LastUpdate) > 3*time.Minute {
			progress.IsStuck = true
			progress.StuckReason = "长时间无响应"
		}

		report.AgentStatus[agent.AgentID] = progress
	}

	return report, nil
}

// DecideNextAction AI决策下一步行动
func (b *OrchestratorBrain) DecideNextAction(ctx context.Context, progress *ProgressReport) (*Action, error) {
	// 简单规则引擎（后续可以用Gemini增强）

	// 1. 检查是否有Agent卡住
	for agentID, agentProgress := range progress.AgentStatus {
		if agentProgress.IsStuck {
			return &Action{
				Type:        ActionHelpAgent,
				TargetAgent: agentID,
				Reason:      fmt.Sprintf("Agent %s 卡住: %s", agentID, agentProgress.StuckReason),
			}, nil
		}
	}

	// 2. 检查是否有空闲Agent可以分配任务
	for agentID, agentProgress := range progress.AgentStatus {
		if agentProgress.State == models.AgentStateIdle {
			// TODO: 考虑任务依赖和优先级
			return &Action{
				Type:        ActionAssignTask,
				TargetAgent: agentID,
				Reason:      fmt.Sprintf("Agent %s 空闲，可分配新任务", agentID),
			}, nil
		}
	}

	// 3. 所有Agent都在工作，等待
	return &Action{
		Type:   ActionWait,
		Reason: "所有Agent都在工作中",
	}, nil
}

// cleanJSONResponse 清理响应中的markdown标记和多余空白
func cleanJSONResponse(response string) string {
	// 去除 ```json 和 ``` 标记
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)
	return response
}
