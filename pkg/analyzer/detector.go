package analyzer

import (
	"strings"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
)

const (
	// ContextWindowSize is the number of lines to keep in context
	// 🔧 FIX: 增加到 200 行以获取更完整的上下文
	ContextWindowSize = 200

	// StuckThreshold is the duration after which an agent is considered stuck
	StuckThreshold = 60 * time.Second
)

// ErrorType classifies errors for retry logic
type ErrorType int

const (
	ErrorTypeUnknown      ErrorType = iota // Unknown error
	ErrorTypeRetryable                     // Retryable (network, temporary failures)
	ErrorTypeNonRetryable                  // Non-retryable (syntax, logic errors)
	ErrorTypeFatal                         // Fatal (requires human intervention)
)

// ErrorDetails contains detailed information about an error
type ErrorDetails struct {
	Type    ErrorType
	Message string
	Context string
}

// ConfirmStats tracks confirmation statistics
// 🔧 P1 FIX: 添加确认统计（用于监控和审计）
type ConfirmStats struct {
	TotalRequests   int
	AutoConfirmed   int
	ManualRequired  int
	Blocked         int
	TimeoutCount    int
	LastConfirmTime time.Time
}

// Detector analyzes Claude output and detects state
type Detector struct {
	contextWindow       []string
	lastOutput          time.Time
	waitingConfirmSince time.Time     // 🔧 P1 FIX: 追踪进入确认等待状态的时间
	confirmStats        ConfirmStats  // 🔧 P1 FIX: 确认统计信息
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

	// Get recent context (last 50 lines for better analysis)
	// 🔧 FIX: 增加分析行数以获取更多上下文
	recentLines := d.contextWindow
	if len(recentLines) > 50 {
		recentLines = recentLines[len(recentLines)-50:]
	}
	recent := strings.Join(recentLines, "\n")

	// Check patterns in order of priority
	if PatternWaitingConfirm.MatchString(recent) {
		// 🔧 P1 FIX: 追踪确认等待开始时间
		if d.waitingConfirmSince.IsZero() {
			d.waitingConfirmSince = time.Now()
		}
		return models.AgentStateWaitingConfirm
	}

	// 🔧 P1 FIX: 非确认状态，重置等待时间
	d.waitingConfirmSince = time.Time{}

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
	// Get recent context (last 100 lines for comprehensive analysis)
	// 🔧 FIX: 增加到 100 行以获取更完整的危险操作上下文
	recentLines := d.contextWindow
	if len(recentLines) > 100 {
		recentLines = recentLines[len(recentLines)-100:]
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

	// 5. 🔧 FIX: 默认拒绝（安全优先原则）
	// 只有明确识别为安全操作才确认，未知场景默认拒绝
	return false
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
	d.waitingConfirmSince = time.Time{} // 🔧 P1 FIX: 重置确认等待时间
}

// IsConfirmTimeout checks if confirmation has been waiting too long
// 🔧 P1 FIX: 添加确认超时检测（5分钟）
func (d *Detector) IsConfirmTimeout() bool {
	if d.waitingConfirmSince.IsZero() {
		return false
	}
	return time.Since(d.waitingConfirmSince) > 5*time.Minute
}

// GetConfirmWaitDuration returns how long we've been waiting for confirmation
// 🔧 P1 FIX: 获取确认等待时长（用于日志和监控）
func (d *Detector) GetConfirmWaitDuration() time.Duration {
	if d.waitingConfirmSince.IsZero() {
		return 0
	}
	return time.Since(d.waitingConfirmSince)
}

// GetConfirmStats returns confirmation statistics
// 🔧 P1 FIX: 获取确认统计信息（用于监控和报告）
func (d *Detector) GetConfirmStats() ConfirmStats {
	return d.confirmStats
}

// ResetConfirmStats resets confirmation statistics
// 🔧 P1 FIX: 重置统计信息（用于测试或周期性重置）
func (d *Detector) ResetConfirmStats() {
	d.confirmStats = ConfirmStats{}
}

// AnalyzeError analyzes the output to determine error type and details
func (d *Detector) AnalyzeError(output string) *ErrorDetails {
	outputLower := strings.ToLower(output)

	details := &ErrorDetails{
		Type:    ErrorTypeUnknown,
		Context: output,
	}

	// Retryable errors (network, temporary failures)
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"network unreachable",
		"temporary failure",
		"try again",
		"rate limit",
		"429",
		"503 service unavailable",
		"504 gateway timeout",
		"econnrefused",
		"econnreset",
		"etimedout",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(outputLower, pattern) {
			details.Type = ErrorTypeRetryable
			details.Message = "Network or temporary failure detected"
			return details
		}
	}

	// Non-retryable errors (syntax, logic, validation)
	nonRetryablePatterns := []string{
		"syntax error",
		"parse error",
		"invalid syntax",
		"unexpected token",
		"undefined",
		"not defined",
		"cannot find",
		"no such file",
		"permission denied",
		"access denied",
		"401 unauthorized",
		"403 forbidden",
		"404 not found",
		"validation error",
		"invalid argument",
		"type error",
	}

	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(outputLower, pattern) {
			details.Type = ErrorTypeNonRetryable
			details.Message = "Syntax or logic error detected"
			return details
		}
	}

	// Fatal errors (requires human intervention)
	fatalPatterns := []string{
		"panic",
		"fatal error",
		"segmentation fault",
		"out of memory",
		"disk full",
		"no space left",
		"database locked",
		"corruption",
		"critical error",
	}

	for _, pattern := range fatalPatterns {
		if strings.Contains(outputLower, pattern) {
			details.Type = ErrorTypeFatal
			details.Message = "Fatal error requiring human intervention"
			return details
		}
	}

	// If we detected an error state but can't classify it, treat as retryable
	if strings.Contains(outputLower, "error") || strings.Contains(outputLower, "failed") {
		details.Type = ErrorTypeRetryable
		details.Message = "Unclassified error - treating as retryable"
	}

	return details
}
