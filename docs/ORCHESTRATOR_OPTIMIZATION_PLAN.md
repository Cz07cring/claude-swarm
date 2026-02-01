# Gemini 主脑优化计划

**日期**: 2026-02-01
**当前版本**: V2 (基础版)
**目标版本**: V2.5 (智能增强版)
**优先级**: P0 (核心功能缺陷修复)

---

## 📊 当前状况分析

### ✅ 已实现的功能

1. **基础需求分析** (`AnalyzeRequirement`)
   - ✅ Gemini API 集成
   - ✅ 智能 Prompt 设计（要求具体任务描述）
   - ✅ JSON 响应解析
   - ✅ 模块和任务拆分
   - ✅ 依赖关系识别
   - ✅ 重试机制（指数退避：1s, 3s, 10s）
   - ✅ 超时控制（2分钟）
   - ✅ 上下文管理（防止内存泄漏）

2. **人工审批流程** (`requestApproval`)
   - ✅ 分析结果展示
   - ✅ 详细信息查看
   - ✅ 批准/拒绝选项

3. **进度监控框架** (`MonitorProgress`)
   - ✅ 任务状态统计
   - ✅ Agent 卡住检测（3分钟无响应）
   - ✅ 进度百分比计算

4. **决策引擎框架** (`DecideNextAction`)
   - ✅ 基于规则的简单决策
   - ✅ Agent 卡住处理
   - ✅ 空闲 Agent 任务分配

### ❌ 存在的问题

#### 🔴 P0 - 严重问题（核心功能缺陷）

1. **任务创建丢失关键信息** (`CreateTasksFromAnalysis`:250-272)
   ```go
   // 当前代码：
   task := &models.Task{
       ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
       Description: taskSpec.Description,
       Status:      models.TaskStatusPending,
       // ❌ 缺少：Priority, Dependencies, MaxRetries
   }
   ```
   **影响**：
   - 任务优先级丢失，无法按重要性排序
   - 依赖关系丢失，无法实现 DAG 调度
   - AI 分析的预估时间信息丢失

2. **没有使用 DAG 依赖调度**
   - AI 分析出了依赖关系，但任务创建时没有应用
   - 导致应该按顺序执行的任务可能并行执行
   - 例如：创建数据库模型应该在实现 API 之前

3. **MonitorProgress 和 DecideNextAction 未集成**
   - 这两个函数已实现但从未被调用
   - 主脑无法实时监控和干预 Agent

#### 🟡 P1 - 重要问题（功能不完善）

4. **错误处理不智能**
   - Agent 失败时没有智能分析原因
   - 没有自动重试策略
   - 没有失败任务重新分配机制

5. **缺少卡住 Agent 的自动处理**
   - 检测到卡住但没有实际处理逻辑
   - 没有自动提示或帮助机制

6. **没有任务质量检查**
   - 任务完成后没有验证
   - 没有检查是否符合需求

7. **缺少进度追踪和预测**
   - 无法预测剩余时间
   - 没有可视化进度报告

#### 🟢 P2 - 改进建议（体验优化）

8. **Prompt 可以更智能**
   - 可以根据项目类型调整
   - 可以学习用户偏好

9. **缺少上下文感知**
   - 没有读取现有代码结构
   - 不知道项目的技术栈

10. **没有实现自适应调度**
    - Agent 性能不同，应动态调整任务分配
    - 快的 Agent 应该分配更多任务

---

## 🎯 优化计划

### Phase 1: P0 问题修复（1-2小时）⭐⭐⭐

#### 1.1 修复 CreateTasksFromAnalysis

**文件**: `pkg/orchestrator/brain.go:250-272`

**问题**：任务创建时丢失 Priority、Dependencies、MaxRetries

**修复方案**：
```go
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
            Priority:    taskSpec.Priority,        // ✅ 添加优先级
            MaxRetries:  3,                        // ✅ 设置重试次数
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
            task, _ := b.taskQueue.GetTask(actualID)

            // 转换 AI 的 ID 为实际 ID
            actualDeps := make([]string, 0)
            for _, depID := range taskSpec.Dependencies {
                if actualDepID, exists := taskIDMap[depID]; exists {
                    actualDeps = append(actualDeps, actualDepID)
                } else {
                    log.Printf("⚠️  警告：任务 %s 依赖的任务 %s 不存在", taskSpec.ID, depID)
                }
            }

            // 更新任务的依赖
            task.Dependencies = actualDeps
            // 需要添加 UpdateTask 方法到 TaskQueue
        }
    }

    return nil
}
```

**需要扩展**：
- `pkg/state/taskqueue.go`: 添加 `UpdateTask(task *models.Task) error` 方法

#### 1.2 验证依赖关系正确性

**新增函数**：
```go
// ValidateDependencies 验证依赖关系是否合理
func (b *OrchestratorBrain) ValidateDependencies(result *AnalysisResult) error {
    taskMap := make(map[string]bool)
    for _, task := range result.Tasks {
        taskMap[task.ID] = true
    }

    // 检查依赖是否存在
    for _, task := range result.Tasks {
        for _, depID := range task.Dependencies {
            if !taskMap[depID] {
                return fmt.Errorf("任务 %s 依赖的任务 %s 不存在", task.ID, depID)
            }
        }
    }

    // 检查循环依赖（使用 DAG 检测）
    if hasCycle := detectCycle(result.Tasks); hasCycle {
        return fmt.Errorf("检测到循环依赖")
    }

    return nil
}
```

#### 1.3 集成到 orchestrate 命令

**文件**: `cmd/swarm/orchestrate.go:124-128`

**添加验证步骤**：
```go
// 验证依赖关系
if err := brain.ValidateDependencies(result); err != nil {
    log.Fatalf("❌ 依赖关系验证失败: %v", err)
}

// 创建任务（现在会正确设置优先级和依赖）
if err := brain.CreateTasksFromAnalysis(ctx, result); err != nil {
    log.Fatalf("❌ 创建任务失败: %v", err)
}
```

---

### Phase 2: P1 问题修复（2-3小时）⭐⭐

#### 2.1 实现智能错误处理

**新增功能**: `DiagnoseFailure` - 分析失败原因并提供建议

```go
// DiagnoseFailure 使用 Gemini 分析任务失败原因
func (b *OrchestratorBrain) DiagnoseFailure(ctx context.Context, task *models.Task) (*FailureDiagnosis, error) {
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

返回JSON格式：
{
  "root_cause": "根本原因分析",
  "should_retry": true/false,
  "retry_suggestion": "如何修改任务描述以提高成功率",
  "alternative_action": "如果不重试，建议的替代方案",
  "estimated_success_rate": "预估重试成功率 (0-100)"
}`, task.ID, task.Description, task.RetryCount, task.MaxRetries, task.LastError)

    // 调用 Gemini
    result, err := b.callGemini(ctx, prompt)
    if err != nil {
        return nil, err
    }

    var diagnosis FailureDiagnosis
    if err := json.Unmarshal([]byte(result), &diagnosis); err != nil {
        return nil, err
    }

    return &diagnosis, nil
}

type FailureDiagnosis struct {
    RootCause           string  `json:"root_cause"`
    ShouldRetry         bool    `json:"should_retry"`
    RetrySuggestion     string  `json:"retry_suggestion"`
    AlternativeAction   string  `json:"alternative_action"`
    EstimatedSuccessRate int    `json:"estimated_success_rate"`
}
```

#### 2.2 实现卡住 Agent 的智能帮助

**新增功能**: `HelpStuckAgent` - 分析 Agent 卡住原因并提供帮助

```go
// HelpStuckAgent 帮助卡住的 Agent
func (b *OrchestratorBrain) HelpStuckAgent(ctx context.Context, agentID string, task *models.Task, lastOutput string) (*AgentHelp, error) {
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

返回JSON：
{
  "stuck_point": "卡住的具体位置/问题",
  "hint": "给Agent的提示（一两句话）",
  "should_reassign": true/false,
  "reassign_reason": "如果需要重新分配，说明原因"
}`, agentID, task.Description, lastOutput)

    result, err := b.callGemini(ctx, prompt)
    if err != nil {
        return nil, err
    }

    var help AgentHelp
    if err := json.Unmarshal([]byte(result), &help); err != nil {
        return nil, err
    }

    return &help, nil
}
```

#### 2.3 实现任务质量检查

**新增功能**: `ValidateTaskCompletion` - 检查任务是否真正完成

```go
// ValidateTaskCompletion 验证任务完成质量
func (b *OrchestratorBrain) ValidateTaskCompletion(ctx context.Context, task *models.Task, output string) (*QualityReport, error) {
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

返回JSON：
{
  "is_complete": true/false,
  "quality_score": 0-100,
  "issues": ["发现的问题列表"],
  "needs_rework": true/false,
  "rework_instructions": "如果需要返工，具体要改什么"
}`, task.Description, output)

    result, err := b.callGemini(ctx, prompt)
    if err != nil {
        return nil, err
    }

    var report QualityReport
    if err := json.Unmarshal([]byte(result), &report); err != nil {
        return nil, err
    }

    return &report, nil
}
```

---

### Phase 3: 集成到 start-v2 流程（2-3小时）⭐

#### 3.1 创建主脑监控循环

**文件**: `cmd/swarm/start_v2.go`

**集成方案**：
```go
// 在 start-v2 中启动主脑监控协程
go func() {
    ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // 收集 Agent 状态
            agents := collectAgentStatus()

            // AI 监控进度
            progress, err := brain.MonitorProgress(ctx, agents)
            if err != nil {
                log.Printf("⚠️  主脑监控失败: %v", err)
                continue
            }

            // AI 决策下一步行动
            action, err := brain.DecideNextAction(ctx, progress)
            if err != nil {
                log.Printf("⚠️  主脑决策失败: %v", err)
                continue
            }

            // 执行行动
            executeAction(action)

        case <-ctx.Done():
            return
        }
    }
}()
```

#### 3.2 实现行动执行器

```go
func executeAction(action *orchestrator.Action) {
    switch action.Type {
    case orchestrator.ActionHelpAgent:
        // 帮助卡住的 Agent
        log.Printf("🆘 主脑介入: %s", action.Reason)
        // 发送提示给 Agent

    case orchestrator.ActionReassignTask:
        // 重新分配任务
        log.Printf("🔄 重新分配任务: %s", action.Reason)

    case orchestrator.ActionRestartAgent:
        // 重启 Agent
        log.Printf("♻️  重启Agent: %s", action.Reason)

    case orchestrator.ActionWait:
        // 等待
        log.Printf("⏳ 主脑等待: %s", action.Reason)
    }
}
```

---

### Phase 4: P2 优化（3-4小时，可选）⭐

#### 4.1 上下文感知

**功能**：读取项目信息，提供更精准的任务拆分

```go
// AnalyzeProjectContext 分析项目上下文
func (b *OrchestratorBrain) AnalyzeProjectContext() (*ProjectContext, error) {
    // 读取 go.mod, package.json 等
    // 分析目录结构
    // 识别技术栈
    // 读取 README
}
```

#### 4.2 自适应调度

**功能**：根据 Agent 性能动态调整任务分配

```go
// 记录每个 Agent 的表现
type AgentPerformance struct {
    TasksCompleted    int
    AverageTime       time.Duration
    FailureRate       float64
    PreferredTaskType string
}

// 优先分配任务给表现好的 Agent
```

#### 4.3 学习用户偏好

**功能**：记录用户对分析结果的修改，下次改进

```go
// 保存用户反馈
type UserFeedback struct {
    OriginalAnalysis *AnalysisResult
    UserModifications []Modification
    Timestamp        time.Time
}
```

---

## 📋 实施优先级

### 立即修复（本周）

1. ✅ **修复 CreateTasksFromAnalysis**（30分钟）
   - 添加 Priority、Dependencies 支持
   - 添加依赖验证

2. ✅ **添加错误诊断**（1小时）
   - DiagnoseFailure 函数
   - 失败任务智能重试

3. ✅ **集成监控循环**（1-2小时）
   - 在 start-v2 中启动主脑
   - 实现基础行动执行

### 本月完成

4. 🔄 **任务质量检查**（1小时）
   - ValidateTaskCompletion
   - 自动检测需要返工的任务

5. 🔄 **卡住 Agent 帮助**（1小时）
   - HelpStuckAgent
   - 智能提示生成

### 未来增强

6. 📅 **上下文感知**（2-3小时）
7. 📅 **自适应调度**（2-3小时）
8. 📅 **学习用户偏好**（3-4小时）

---

## 🎯 成功标准

### Phase 1 完成标准
- ✅ 任务创建时正确设置优先级和依赖
- ✅ DAG 调度器按依赖顺序执行任务
- ✅ 可以通过 `swarm status` 看到依赖关系

### Phase 2 完成标准
- ✅ 失败任务自动分析原因
- ✅ 根据诊断决定是否重试
- ✅ 卡住的 Agent 能收到主脑的帮助

### Phase 3 完成标准
- ✅ start-v2 自动启动主脑监控
- ✅ 主脑每30秒分析一次进度
- ✅ 主脑能自动处理常见问题

---

## 📊 预期效果

### 优化前
- ❌ 任务无序执行，可能顺序错误
- ❌ 失败任务盲目重试
- ❌ Agent 卡住无人管
- ❌ 任务完成质量无法保证

### 优化后
- ✅ 任务按依赖顺序执行
- ✅ 失败任务智能诊断和重试
- ✅ Agent 卡住主脑介入帮助
- ✅ 任务完成质量自动检查
- ✅ 整体成功率提升 30-50%

---

## 🛠️ 技术细节

### 需要扩展的 API

#### pkg/state/taskqueue.go
```go
// 添加更新任务的方法
func (tq *TaskQueue) UpdateTask(task *models.Task) error

// 添加按依赖查询的方法
func (tq *TaskQueue) GetTasksWaitingFor(taskID string) []*models.Task
```

#### pkg/orchestrator/brain.go
```go
// 添加通用的 Gemini 调用方法
func (b *OrchestratorBrain) callGemini(ctx context.Context, prompt string) (string, error)

// 添加诊断、帮助、验证等方法
func (b *OrchestratorBrain) DiagnoseFailure(...)
func (b *OrchestratorBrain) HelpStuckAgent(...)
func (b *OrchestratorBrain) ValidateTaskCompletion(...)
func (b *OrchestratorBrain) ValidateDependencies(...)
```

---

## 📝 测试计划

### 单元测试
```bash
# 测试依赖关系创建
go test -v ./pkg/orchestrator -run TestCreateTasksWithDependencies

# 测试循环依赖检测
go test -v ./pkg/orchestrator -run TestDetectCyclicDependencies

# 测试失败诊断
go test -v ./pkg/orchestrator -run TestDiagnoseFailure
```

### 集成测试
```bash
# 测试完整流程
swarm orchestrate "创建一个博客系统，包括文章管理、评论功能、用户系统"
swarm status --verbose  # 检查依赖关系
swarm start-v2 --agents 3
# 观察主脑是否正确监控和干预
```

### 压力测试
```bash
# 测试大量任务的场景
swarm orchestrate "实现一个完整的电商系统"  # 可能生成20+任务
# 检查依赖关系是否正确
# 检查性能是否可接受
```

---

**创建时间**: 2026-02-01
**预计实施时间**: Phase 1-2 = 3-5 小时，Phase 3 = 2-3 小时
**总预计时间**: 5-8 小时
**优先级**: P0 > P1 > P2
