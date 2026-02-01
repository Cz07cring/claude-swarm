package orchestrator

import (
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
)

// AnalysisResult AI分析用户需求的结果
type AnalysisResult struct {
	Summary     string           `json:"summary"`      // 需求概要
	Modules     []Module         `json:"modules"`      // 拆分的模块
	Tasks       []*TaskSpec      `json:"tasks"`        // 生成的任务列表
	Dependencies map[string][]string `json:"dependencies"` // 任务依赖关系 taskID -> [依赖的taskIDs]
	EstimatedTime string          `json:"estimated_time"` // 预计完成时间
	Complexity   string           `json:"complexity"`   // 复杂度 low/medium/high
}

// Module 需求模块
type Module struct {
	Name        string   `json:"name"`        // 模块名称
	Description string   `json:"description"` // 模块描述
	Files       []string `json:"files"`       // 涉及的文件
	Priority    int      `json:"priority"`    // 优先级 1-10
}

// TaskSpec AI生成的任务规格
type TaskSpec struct {
	ID          string   `json:"id"`          // 任务ID
	Description string   `json:"description"` // 任务描述
	Module      string   `json:"module"`      // 所属模块
	Files       []string `json:"files"`       // 涉及的文件
	Dependencies []string `json:"dependencies"` // 依赖的任务ID
	Priority    int      `json:"priority"`    // 优先级
	Estimated   string   `json:"estimated"`   // 预计耗时
}

// ProgressReport 进展报告
type ProgressReport struct {
	Timestamp      time.Time              `json:"timestamp"`
	TotalTasks     int                    `json:"total_tasks"`
	CompletedTasks int                    `json:"completed_tasks"`
	InProgressTasks int                   `json:"in_progress_tasks"`
	FailedTasks    int                    `json:"failed_tasks"`
	AgentStatus    map[string]*AgentProgress `json:"agent_status"` // agentID -> progress
	OverallProgress float64               `json:"overall_progress"` // 0-100
	EstimatedTimeLeft string              `json:"estimated_time_left"`
	Blockers       []string               `json:"blockers"` // 阻塞问题
}

// AgentProgress Agent进展
type AgentProgress struct {
	AgentID      string             `json:"agent_id"`
	State        models.AgentState  `json:"state"`
	CurrentTask  *models.Task       `json:"current_task,omitempty"`
	TaskProgress string             `json:"task_progress"` // AI分析的进展描述
	IsStuck      bool               `json:"is_stuck"`
	StuckReason  string             `json:"stuck_reason,omitempty"`
	LastUpdate   time.Time          `json:"last_update"`
}

// Action AI决策的行动
type Action struct {
	Type        ActionType `json:"type"`
	TargetAgent string     `json:"target_agent,omitempty"`
	TaskID      string     `json:"task_id,omitempty"`
	Command     string     `json:"command,omitempty"`
	Reason      string     `json:"reason"`
}

// ActionType 行动类型
type ActionType string

const (
	ActionAssignTask    ActionType = "assign_task"     // 分配任务
	ActionReassignTask  ActionType = "reassign_task"   // 重新分配任务
	ActionHelpAgent     ActionType = "help_agent"      // 帮助卡住的Agent
	ActionRestartAgent  ActionType = "restart_agent"   // 重启Agent
	ActionMergeBranch   ActionType = "merge_branch"    // 合并分支
	ActionWait          ActionType = "wait"            // 等待
	ActionAskUser       ActionType = "ask_user"        // 询问用户
)

// ConversationContext AI对话上下文
type ConversationContext struct {
	Requirement     string                 `json:"requirement"`      // 原始需求
	AnalysisResult  *AnalysisResult        `json:"analysis_result"`  // 分析结果
	TaskHistory     []*models.Task         `json:"task_history"`     // 任务历史
	Conversations   []Message              `json:"conversations"`    // 对话历史
	CurrentPhase    string                 `json:"current_phase"`    // 当前阶段
	Decisions       []Decision             `json:"decisions"`        // 决策记录
}

// Message 对话消息
type Message struct {
	Role      string    `json:"role"`    // user/assistant
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Decision AI决策记录
type Decision struct {
	Timestamp   time.Time  `json:"timestamp"`
	Action      *Action    `json:"action"`
	Reasoning   string     `json:"reasoning"`   // 决策理由
	Result      string     `json:"result"`      // 执行结果
}

// FailureDiagnosis 失败任务诊断结果
type FailureDiagnosis struct {
	RootCause            string  `json:"root_cause"`             // 根本原因分析
	ShouldRetry          bool    `json:"should_retry"`           // 是否值得重试
	RetrySuggestion      string  `json:"retry_suggestion"`       // 重试建议
	AlternativeAction    string  `json:"alternative_action"`     // 替代方案
	EstimatedSuccessRate int     `json:"estimated_success_rate"` // 预估成功率 0-100
}

// AgentHelp Agent 帮助信息
type AgentHelp struct {
	StuckPoint       string `json:"stuck_point"`        // 卡住的具体位置
	Hint             string `json:"hint"`               // 给 Agent 的提示
	ShouldReassign   bool   `json:"should_reassign"`    // 是否重新分配
	ReassignReason   string `json:"reassign_reason"`    // 重新分配原因
}

// QualityReport 任务质量报告
type QualityReport struct {
	IsComplete         bool     `json:"is_complete"`         // 是否完成
	QualityScore       int      `json:"quality_score"`       // 质量评分 0-100
	Issues             []string `json:"issues"`              // 发现的问题
	NeedsRework        bool     `json:"needs_rework"`        // 是否需要返工
	ReworkInstructions string   `json:"rework_instructions"` // 返工指示
}

// MergeDecision 合并决策
type MergeDecision struct {
	ShouldMerge     bool     `json:"should_merge"`      // 是否应该合并
	MergeOrder      []string `json:"merge_order"`       // 合并顺序（分支名）
	Reason          string   `json:"reason"`            // 决策理由
	PotentialIssues []string `json:"potential_issues"`  // 潜在问题
}

// ConflictResolution 冲突解决方案
type ConflictResolution struct {
	CanAutoResolve   bool              `json:"can_auto_resolve"`   // 是否可以自动解决
	Resolution       string            `json:"resolution"`         // 解决方案
	FileResolutions  map[string]string `json:"file_resolutions"`   // 每个文件的解决方案
	NeedsHumanReview bool              `json:"needs_human_review"` // 是否需要人工审核
	Reason           string            `json:"reason"`             // 理由
}

// MergeStatus 合并状态
type MergeStatus struct {
	Branch      string   `json:"branch"`       // 分支名
	AgentID     string   `json:"agent_id"`     // Agent ID
	HasChanges  bool     `json:"has_changes"`  // 是否有改动
	CommitCount int      `json:"commit_count"` // 提交数量
	Files       []string `json:"files"`        // 修改的文件
	ReadyToMerge bool    `json:"ready_to_merge"` // 是否可以合并
}
