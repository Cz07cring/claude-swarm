package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
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

	// 如果没有传入 apiKey，尝试从环境变量读取
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("gemini API key is required (pass as parameter or set GEMINI_API_KEY env var)")
		}
	}

	// 初始化Gemini客户端（使用 apiKey）
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
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

	// 添加超时控制（2分钟）
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// 调用Gemini API with重试机制
	var result *genai.GenerateContentResponse
	var err error

	// 指数退避重试策略：1s, 3s, 10s
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 10 * time.Second}
	maxAttempts := len(retryDelays) + 1 // 1次初始 + 3次重试

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			log.Printf("⚠️  API调用失败，第 %d/%d 次重试...", attempt, maxAttempts-1)

			// 等待重试延迟
			select {
			case <-time.After(retryDelays[attempt-1]):
				// 继续重试
			case <-ctx.Done():
				return nil, fmt.Errorf("API调用取消: %w", ctx.Err())
			}
		}

		result, err = b.client.Models.GenerateContent(
			ctx,
			b.modelName,
			genai.Text(prompt),
			nil,
		)

		if err == nil {
			// 成功，退出重试循环
			break
		}

		// 检查是否是不可重试的错误
		if ctx.Err() != nil {
			// Context 取消或超时，不重试
			return nil, fmt.Errorf("API调用超时或取消: %w", ctx.Err())
		}

		// 记录错误，准备重试
		log.Printf("⚠️  API调用失败 (尝试 %d/%d): %v", attempt+1, maxAttempts, err)
	}

	if err != nil {
		return nil, fmt.Errorf("Gemini API调用失败（已重试%d次）: %w", maxAttempts-1, err)
	}

	// 获取响应文本
	responseText := result.Text()

	// 解析JSON响应
	analysisResult, err := b.parseAnalysisResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("解析AI响应失败: %w\n原始响应: %s", err, responseText)
	}

	// 保存到上下文（限制大小）
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

	// 限制上下文大小，防止内存泄漏（保留最近50条对话）
	const maxConversations = 50
	if len(b.context.Conversations) > maxConversations {
		// 保留最近的对话
		b.context.Conversations = b.context.Conversations[len(b.context.Conversations)-maxConversations:]
		log.Printf("⚠️  对话历史已满，清理旧对话（保留最近%d条）", maxConversations)
	}

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

	// 第一遍：创建所有任务（不设置依赖）
	taskIDMap := make(map[string]string) // AI生成的ID -> 实际存储的ID

	for _, taskSpec := range result.Tasks {
		// 生成唯一ID
		actualID := fmt.Sprintf("task-%d", time.Now().UnixNano())
		time.Sleep(1 * time.Millisecond) // 确保ID唯一

		task := &models.Task{
			ID:          actualID,
			Description: taskSpec.Description,
			Status:      models.TaskStatusPending,
			Priority:    taskSpec.Priority,    // ✅ 添加优先级
			MaxRetries:  3,                    // ✅ 设置重试次数
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		taskIDMap[taskSpec.ID] = actualID

		// 添加到任务队列（暂时不设置依赖）
		if err := b.taskQueue.AddTask(task); err != nil {
			return fmt.Errorf("添加任务失败: %w", err)
		}

		log.Printf("  ✓ %s: %s (优先级: %d)", actualID, task.Description, task.Priority)
	}

	// 第二遍：更新依赖关系
	for _, taskSpec := range result.Tasks {
		if len(taskSpec.Dependencies) > 0 {
			actualID := taskIDMap[taskSpec.ID]
			task, err := b.taskQueue.GetTask(actualID)
			if err != nil {
				log.Printf("⚠️  警告：获取任务 %s 失败: %v", actualID, err)
				continue
			}

			// 转换 AI 的 ID 为实际 ID
			actualDeps := make([]string, 0)
			for _, depID := range taskSpec.Dependencies {
				if actualDepID, exists := taskIDMap[depID]; exists {
					actualDeps = append(actualDeps, actualDepID)
				} else {
					log.Printf("⚠️  警告：任务 %s 依赖的任务 %s 不存在，已跳过", taskSpec.ID, depID)
				}
			}

			if len(actualDeps) > 0 {
				// 更新任务的依赖
				task.Dependencies = actualDeps

				// 保存更新
				if err := b.taskQueue.UpdateTask(task); err != nil {
					log.Printf("⚠️  警告：更新任务 %s 的依赖失败: %v", actualID, err)
				} else {
					log.Printf("  🔗 %s 依赖于 %v", actualID, actualDeps)
				}
			}
		}
	}

	log.Printf("✅ 任务队列创建完成: %d个任务，依赖关系已设置", len(result.Tasks))
	return nil
}

// ValidateDependencies 验证依赖关系是否合理
func (b *OrchestratorBrain) ValidateDependencies(result *AnalysisResult) error {
	if len(result.Tasks) == 0 {
		return nil
	}

	// 构建任务映射
	taskMap := make(map[string]*TaskSpec)
	for _, task := range result.Tasks {
		taskMap[task.ID] = task
	}

	// 检查依赖是否存在
	for _, task := range result.Tasks {
		for _, depID := range task.Dependencies {
			if _, exists := taskMap[depID]; !exists {
				return fmt.Errorf("任务 %s 依赖的任务 %s 不存在", task.ID, depID)
			}
		}
	}

	// 检查循环依赖
	if err := b.detectCyclicDependencies(result.Tasks); err != nil {
		return err
	}

	log.Printf("✅ 依赖关系验证通过")
	return nil
}

// detectCyclicDependencies 检测循环依赖
func (b *OrchestratorBrain) detectCyclicDependencies(tasks []*TaskSpec) error {
	// 构建邻接表
	graph := make(map[string][]string)
	for _, task := range tasks {
		graph[task.ID] = task.Dependencies
	}

	// 使用 DFS 检测环
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(taskID string, path []string) error
	hasCycle = func(taskID string, path []string) error {
		visited[taskID] = true
		recStack[taskID] = true
		path = append(path, taskID)

		// 检查所有依赖
		for _, depID := range graph[taskID] {
			if !visited[depID] {
				if err := hasCycle(depID, path); err != nil {
					return err
				}
			} else if recStack[depID] {
				// 发现环
				cyclePath := append(path, depID)
				return fmt.Errorf("检测到循环依赖: %v", cyclePath)
			}
		}

		recStack[taskID] = false
		return nil
	}

	// 检查所有任务
	for _, task := range tasks {
		if !visited[task.ID] {
			if err := hasCycle(task.ID, []string{}); err != nil {
				return err
			}
		}
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

// DecideNextAction AI决策下一步行动（增强版）
func (b *OrchestratorBrain) DecideNextAction(ctx context.Context, progress *ProgressReport) (*Action, error) {
	// 优先级：失败任务 > 卡住Agent > 空闲Agent > 等待

	// 1. 检查失败任务，决定是否重试
	if progress.FailedTasks > 0 {
		tasks := b.taskQueue.ListTasks()
		for _, task := range tasks {
			if task.Status == models.TaskStatusFailed {
				// 使用AI诊断失败原因
				diagnosis, err := b.DiagnoseFailure(ctx, task)
				if err != nil {
					log.Printf("⚠️  诊断失败任务出错: %v", err)
					continue
				}

				if diagnosis.ShouldRetry && task.RetryCount < task.MaxRetries {
					return &Action{
						Type:   ActionReassignTask,
						TaskID: task.ID,
						Reason: fmt.Sprintf("失败任务 %s 值得重试 (成功率: %d%%): %s",
							task.ID, diagnosis.EstimatedSuccessRate, diagnosis.RetrySuggestion),
						Command: diagnosis.RetrySuggestion,
					}, nil
				} else {
					log.Printf("⚠️  任务 %s 不建议重试: %s", task.ID, diagnosis.AlternativeAction)
					// 可以记录到决策历史，但不采取行动
				}
			}
		}
	}

	// 2. 检查是否有Agent卡住
	for agentID, agentProgress := range progress.AgentStatus {
		if agentProgress.IsStuck && agentProgress.CurrentTask != nil {
			// 使用AI帮助卡住的Agent
			help, err := b.HelpStuckAgent(ctx, agentID, agentProgress.CurrentTask, "")
			if err != nil {
				log.Printf("⚠️  生成帮助信息出错: %v", err)
				// 降级为基础帮助
				return &Action{
					Type:        ActionHelpAgent,
					TargetAgent: agentID,
					TaskID:      agentProgress.CurrentTask.ID,
					Reason:      fmt.Sprintf("Agent %s 卡住: %s", agentID, agentProgress.StuckReason),
					Command:     "请检查任务描述，确认是否需要更多信息",
				}, nil
			}

			if help.ShouldReassign {
				return &Action{
					Type:        ActionReassignTask,
					TargetAgent: agentID,
					TaskID:      agentProgress.CurrentTask.ID,
					Reason:      fmt.Sprintf("重新分配任务: %s", help.ReassignReason),
				}, nil
			} else {
				return &Action{
					Type:        ActionHelpAgent,
					TargetAgent: agentID,
					TaskID:      agentProgress.CurrentTask.ID,
					Reason:      fmt.Sprintf("Agent卡在: %s", help.StuckPoint),
					Command:     help.Hint,
				}, nil
			}
		}
	}

	// 3. 检查是否有空闲Agent可以分配任务
	readyTasks := b.taskQueue.GetReadyTasks()
	if len(readyTasks) > 0 {
		for agentID, agentProgress := range progress.AgentStatus {
			if agentProgress.State == models.AgentStateIdle {
				// 分配优先级最高的就绪任务
				task := readyTasks[0]
				return &Action{
					Type:        ActionAssignTask,
					TargetAgent: agentID,
					TaskID:      task.ID,
					Reason:      fmt.Sprintf("分配任务 %s 给空闲Agent %s (优先级: %d)", task.ID, agentID, task.Priority),
				}, nil
			}
		}
	}

	// 4. 所有Agent都在工作，等待
	return &Action{
		Type:   ActionWait,
		Reason: "所有Agent都在工作中",
	}, nil
}

// callGemini 通用的 Gemini API 调用方法
func (b *OrchestratorBrain) callGemini(ctx context.Context, prompt string) (string, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// 重试策略
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 10 * time.Second}
	maxAttempts := len(retryDelays) + 1

	var result *genai.GenerateContentResponse
	var err error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			log.Printf("⚠️  API调用失败，第 %d/%d 次重试...", attempt, maxAttempts-1)
			select {
			case <-time.After(retryDelays[attempt-1]):
			case <-ctx.Done():
				return "", fmt.Errorf("API调用取消: %w", ctx.Err())
			}
		}

		result, err = b.client.Models.GenerateContent(
			ctx,
			b.modelName,
			genai.Text(prompt),
			nil,
		)

		if err == nil {
			break
		}

		if ctx.Err() != nil {
			return "", fmt.Errorf("API调用超时或取消: %w", ctx.Err())
		}

		log.Printf("⚠️  API调用失败 (尝试 %d/%d): %v", attempt+1, maxAttempts, err)
	}

	if err != nil {
		return "", fmt.Errorf("Gemini API调用失败（已重试%d次）: %w", maxAttempts-1, err)
	}

	return result.Text(), nil
}

// DiagnoseFailure 使用 Gemini 分析任务失败原因
func (b *OrchestratorBrain) DiagnoseFailure(ctx context.Context, task *models.Task) (*FailureDiagnosis, error) {
	log.Printf("🔍 AI诊断失败任务: %s", task.ID)

	prompt := fmt.Sprintf(`你是一个专业的调试专家。某个开发任务失败了，请分析原因并给出解决建议。

任务信息：
- 任务ID: %s
- 任务描述: %s
- 失败次数: %d/%d
- 错误信息: %s

请分析：
1. 失败的可能原因（技术原因、描述不清、依赖问题等）
2. 是否值得重试（true/false）
3. 如果重试，需要修改什么
4. 如果不值得重试，建议怎么处理

返回JSON格式（不要用markdown代码块包裹）：
{
  "root_cause": "根本原因分析",
  "should_retry": true,
  "retry_suggestion": "如何修改任务描述以提高成功率",
  "alternative_action": "如果不重试，建议的替代方案",
  "estimated_success_rate": 75
}`, task.ID, task.Description, task.RetryCount, task.MaxRetries, task.LastError)

	responseText, err := b.callGemini(ctx, prompt)
	if err != nil {
		return nil, err
	}

	responseText = cleanJSONResponse(responseText)

	var diagnosis FailureDiagnosis
	if err := json.Unmarshal([]byte(responseText), &diagnosis); err != nil {
		return nil, fmt.Errorf("解析诊断结果失败: %w\n原始响应: %s", err, responseText)
	}

	log.Printf("✅ 诊断完成: 成功率预估 %d%%, 建议%s",
		diagnosis.EstimatedSuccessRate,
		map[bool]string{true: "重试", false: "不重试"}[diagnosis.ShouldRetry])

	return &diagnosis, nil
}

// HelpStuckAgent 帮助卡住的 Agent
func (b *OrchestratorBrain) HelpStuckAgent(ctx context.Context, agentID string, task *models.Task, lastOutput string) (*AgentHelp, error) {
	log.Printf("🆘 AI帮助卡住的Agent: %s", agentID)

	// 限制输出长度，避免 prompt 过长
	if len(lastOutput) > 1000 {
		lastOutput = lastOutput[:1000] + "...(已截断)"
	}

	prompt := fmt.Sprintf(`你是一个资深导师，帮助卡住的AI开发Agent。

Agent信息：
- Agent ID: %s
- 当前任务: %s
- 最后输出: %s
- 卡住时长: 超过3分钟

请分析：
1. Agent可能在哪里卡住了
2. 给出具体的提示或建议
3. 是否需要重新分配任务

返回JSON格式（不要用markdown代码块包裹）：
{
  "stuck_point": "卡住的具体位置/问题",
  "hint": "给Agent的提示（一两句话，简洁明确）",
  "should_reassign": false,
  "reassign_reason": "如果需要重新分配，说明原因"
}`, agentID, task.Description, lastOutput)

	responseText, err := b.callGemini(ctx, prompt)
	if err != nil {
		return nil, err
	}

	responseText = cleanJSONResponse(responseText)

	var help AgentHelp
	if err := json.Unmarshal([]byte(responseText), &help); err != nil {
		return nil, fmt.Errorf("解析帮助信息失败: %w\n原始响应: %s", err, responseText)
	}

	log.Printf("✅ 帮助生成: %s", help.Hint)

	return &help, nil
}

// ValidateTaskCompletion 验证任务完成质量
func (b *OrchestratorBrain) ValidateTaskCompletion(ctx context.Context, task *models.Task, output string) (*QualityReport, error) {
	log.Printf("🔍 AI检查任务质量: %s", task.ID)

	// 限制输出长度
	if len(output) > 2000 {
		output = output[:2000] + "...(已截断)"
	}

	prompt := fmt.Sprintf(`你是一个代码审查专家。检查这个任务是否真正完成。

任务要求：
%s

Agent的输出：
%s

请检查：
1. 是否完成了任务描述中的所有要求
2. 代码质量如何
3. 是否有明显的bug或问题
4. 是否需要返工

返回JSON格式（不要用markdown代码块包裹）：
{
  "is_complete": true,
  "quality_score": 85,
  "issues": ["发现的问题1", "发现的问题2"],
  "needs_rework": false,
  "rework_instructions": "如果需要返工，具体要改什么"
}`, task.Description, output)

	responseText, err := b.callGemini(ctx, prompt)
	if err != nil {
		return nil, err
	}

	responseText = cleanJSONResponse(responseText)

	var report QualityReport
	if err := json.Unmarshal([]byte(responseText), &report); err != nil {
		return nil, fmt.Errorf("解析质量报告失败: %w\n原始响应: %s", err, responseText)
	}

	log.Printf("✅ 质量检查完成: 评分 %d/100, 完成度: %v",
		report.QualityScore,
		report.IsComplete)

	return &report, nil
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

// DecideMergeStrategy 决定合并策略
func (b *OrchestratorBrain) DecideMergeStrategy(ctx context.Context, mergeStatuses []*MergeStatus) (*MergeDecision, error) {
	if len(mergeStatuses) == 0 {
		return &MergeDecision{
			ShouldMerge: false,
			Reason:      "没有待合并的分支",
		}, nil
	}

	log.Printf("🧠 AI分析合并策略: %d个分支待处理", len(mergeStatuses))

	// 构建分支信息
	var branchInfo strings.Builder
	for _, status := range mergeStatuses {
		branchInfo.WriteString(fmt.Sprintf("- 分支: %s (Agent: %s)\n", status.Branch, status.AgentID))
		branchInfo.WriteString(fmt.Sprintf("  提交数: %d, 文件: %v\n", status.CommitCount, status.Files))
		branchInfo.WriteString(fmt.Sprintf("  可合并: %v\n", status.ReadyToMerge))
	}

	prompt := fmt.Sprintf(`你是一个Git合并策略专家。分析以下待合并的分支，决定最佳合并顺序。

待合并分支：
%s

请分析：
1. 这些分支是否有潜在冲突（基于修改的文件）
2. 最佳合并顺序（考虑依赖关系和冲突风险）
3. 是否应该现在合并，还是等待更多任务完成

返回JSON格式（不要用markdown代码块包裹）：
{
  "should_merge": true,
  "merge_order": ["agent-0-branch", "agent-1-branch"],
  "reason": "决策理由",
  "potential_issues": ["可能的问题1", "可能的问题2"]
}`, branchInfo.String())

	responseText, err := b.callGemini(ctx, prompt)
	if err != nil {
		return nil, err
	}

	responseText = cleanJSONResponse(responseText)

	var decision MergeDecision
	if err := json.Unmarshal([]byte(responseText), &decision); err != nil {
		return nil, fmt.Errorf("解析合并决策失败: %w\n原始响应: %s", err, responseText)
	}

	log.Printf("✅ 合并决策: 合并=%v, 顺序=%v", decision.ShouldMerge, decision.MergeOrder)
	return &decision, nil
}

// ResolveConflict 使用AI分析并解决合并冲突
func (b *OrchestratorBrain) ResolveConflict(ctx context.Context, branch string, conflictFiles []string, conflictContent string) (*ConflictResolution, error) {
	log.Printf("🧠 AI分析合并冲突: %s, 冲突文件: %v", branch, conflictFiles)

	// 限制内容长度
	if len(conflictContent) > 3000 {
		conflictContent = conflictContent[:3000] + "...(已截断)"
	}

	prompt := fmt.Sprintf(`你是一个代码合并专家。分析以下合并冲突并提供解决方案。

分支: %s
冲突文件: %v

冲突内容:
%s

请分析：
1. 冲突的原因
2. 是否可以自动解决（保留两边改动/选择一边）
3. 具体的解决建议

返回JSON格式（不要用markdown代码块包裹）：
{
  "can_auto_resolve": false,
  "resolution": "解决方案描述",
  "file_resolutions": {
    "file1.go": "保留双方改动，手动合并",
    "file2.go": "使用当前分支版本"
  },
  "needs_human_review": true,
  "reason": "为什么需要/不需要人工审核"
}`, branch, conflictFiles, conflictContent)

	responseText, err := b.callGemini(ctx, prompt)
	if err != nil {
		return nil, err
	}

	responseText = cleanJSONResponse(responseText)

	var resolution ConflictResolution
	if err := json.Unmarshal([]byte(responseText), &resolution); err != nil {
		return nil, fmt.Errorf("解析冲突解决方案失败: %w\n原始响应: %s", err, responseText)
	}

	log.Printf("✅ 冲突分析完成: 可自动解决=%v, 需人工=%v", resolution.CanAutoResolve, resolution.NeedsHumanReview)
	return &resolution, nil
}

// ValidateMergeResult 验证合并结果
func (b *OrchestratorBrain) ValidateMergeResult(ctx context.Context, branch string, mergedFiles []string) (*QualityReport, error) {
	log.Printf("🧠 AI验证合并结果: %s", branch)

	prompt := fmt.Sprintf(`你是一个代码审查专家。验证以下分支合并后的代码质量。

合并的分支: %s
涉及的文件: %v

请检查：
1. 合并是否完整
2. 是否有潜在的集成问题
3. 是否需要额外的测试

返回JSON格式（不要用markdown代码块包裹）：
{
  "is_complete": true,
  "quality_score": 85,
  "issues": ["可能的问题"],
  "needs_rework": false,
  "rework_instructions": ""
}`, branch, mergedFiles)

	responseText, err := b.callGemini(ctx, prompt)
	if err != nil {
		return nil, err
	}

	responseText = cleanJSONResponse(responseText)

	var report QualityReport
	if err := json.Unmarshal([]byte(responseText), &report); err != nil {
		return nil, fmt.Errorf("解析验证结果失败: %w\n原始响应: %s", err, responseText)
	}

	log.Printf("✅ 合并验证完成: 评分=%d, 完整=%v", report.QualityScore, report.IsComplete)
	return &report, nil
}
