package analyzer

import (
	"strings"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
)

const (
	// ContextWindowSize is the number of lines to keep in context
	ContextWindowSize = 100

	// StuckThreshold is the duration after which an agent is considered stuck
	StuckThreshold = 60 * time.Second
)

// Detector analyzes Claude output and detects state
type Detector struct {
	contextWindow []string
	lastOutput    time.Time
}

// NewDetector creates a new detector
func NewDetector() *Detector {
	return &Detector{
		contextWindow: make([]string, 0, ContextWindowSize),
		lastOutput:    time.Now(),
	}
}

// Analyze analyzes the output and returns the detected state
func (d *Detector) Analyze(output string) models.AgentState {
	if output == "" {
		// Check if stuck (no output for StuckThreshold)
		if time.Since(d.lastOutput) > StuckThreshold {
			return models.AgentStateStuck
		}
		return models.AgentStateIdle
	}

	// Update last output time
	d.lastOutput = time.Now()

	// Split into lines and update context window
	lines := strings.Split(output, "\n")
	d.contextWindow = append(d.contextWindow, lines...)

	// Keep only recent lines
	if len(d.contextWindow) > ContextWindowSize {
		d.contextWindow = d.contextWindow[len(d.contextWindow)-ContextWindowSize:]
	}

	// Get recent context (last 20 lines)
	recentLines := d.contextWindow
	if len(recentLines) > 20 {
		recentLines = recentLines[len(recentLines)-20:]
	}
	recent := strings.Join(recentLines, "\n")

	// Check patterns in order of priority
	if PatternWaitingConfirm.MatchString(recent) {
		return models.AgentStateWaitingConfirm
	}

	if PatternError.MatchString(recent) {
		return models.AgentStateError
	}

	if PatternToolCall.MatchString(recent) {
		return models.AgentStateWorking
	}

	// Check if showing idle prompt
	if PatternIdle.MatchString(recent) {
		return models.AgentStateIdle
	}

	// Default to working if there's recent output
	return models.AgentStateWorking
}

// SafeToConfirm checks if it's safe to auto-confirm
func (d *Detector) SafeToConfirm() bool {
	// Get recent context (last 50 lines for better analysis)
	recentLines := d.contextWindow
	if len(recentLines) > 50 {
		recentLines = recentLines[len(recentLines)-50:]
	}
	recent := strings.Join(recentLines, "\n")
	recentLower := strings.ToLower(recent)

	// 1. 检查危险关键词
	for _, keyword := range DangerKeywords {
		if strings.Contains(recentLower, keyword) {
			return false
		}
	}

	// 2. 检查是否是计划确认（通常安全）
	if strings.Contains(recentLower, "proceed with this plan") {
		// 但如果计划包含危险操作，还是不确认
		if strings.Contains(recentLower, "delete") ||
		   strings.Contains(recentLower, "remove") ||
		   strings.Contains(recentLower, "force") {
			return false
		}
		return true
	}

	// 3. 检查是否是文件操作确认
	if strings.Contains(recentLower, "overwrite") ||
	   strings.Contains(recentLower, "replace") {
		// 覆盖现有文件 - 需要人工确认
		return false
	}

	// 4. 检查是否是选项列表（1. Yes / 2. No）
	if strings.Contains(recent, "1. Yes") ||
	   strings.Contains(recent, "❯ 1. Yes") {
		// 分析上下文，判断是否安全
		// 如果提到创建、读取、分析等安全操作 - 可以确认
		safeActions := []string{
			"create", "read", "analyze", "show", "display",
			"list", "get", "fetch", "view", "check",
		}
		for _, action := range safeActions {
			if strings.Contains(recentLower, action) {
				return true
			}
		}

		// 如果无法判断，谨慎起见不确认
		return false
	}

	// 5. 默认：简单的 yes/no 确认，如果没有危险关键词就确认
	return true
}

// GetContext returns the current context window
func (d *Detector) GetContext() string {
	return strings.Join(d.contextWindow, "\n")
}

// GetRecentOutput returns the last N lines
func (d *Detector) GetRecentOutput(n int) string {
	if n > len(d.contextWindow) {
		n = len(d.contextWindow)
	}

	if n == 0 {
		return ""
	}

	recentLines := d.contextWindow[len(d.contextWindow)-n:]
	return strings.Join(recentLines, "\n")
}

// Reset resets the detector state
func (d *Detector) Reset() {
	d.contextWindow = make([]string, 0, ContextWindowSize)
	d.lastOutput = time.Now()
}
