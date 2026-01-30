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
