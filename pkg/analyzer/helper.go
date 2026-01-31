package analyzer

import (
	"log"
	"strings"
	"time"
)

// GetConfirmationInput 根据提示类型返回应该发送的确认输入
// 🔧 P1 FIX: 支持更多确认格式
func GetConfirmationInput(context string) string {
	contextLower := strings.ToLower(context)

	// 1. 检查是否是选项列表格式
	if strings.Contains(context, "❯ 1. Yes") ||
	   strings.Contains(context, "1. Yes") ||
	   strings.Contains(context, "► 1. Yes") {
		return "1"
	}

	// 2. 🔧 NEW: 支持 Press Enter 格式
	if strings.Contains(contextLower, "press enter") ||
	   strings.Contains(contextLower, "hit enter") ||
	   strings.Contains(contextLower, "enter to continue") {
		return "" // 发送空行（回车）
	}

	// 3. 🔧 NEW: 支持大写 (Y/N) 格式（检查原始字符串）
	if strings.Contains(context, "(Y/N)") {
		return "Y"
	}

	// 4. 支持小写 (y/n) 格式（检查原始字符串）
	if strings.Contains(context, "(y/n)") {
		return "y"
	}

	// 5. 🔧 NEW: 支持 [Y/n] 格式（默认 Yes）
	if strings.Contains(context, "[Y/n]") {
		return "Y"
	}

	// 6. 🔧 NEW: 支持 [y/N] 格式（默认 No）
	if strings.Contains(context, "[y/N]") {
		// 这种情况比较危险，默认 No 可能不是用户想要的
		// 记录警告
		log.Printf("[CONFIRMATION] WARNING: Detected [y/N] format (default No), sending 'y'")
		return "y"
	}

	// 7. 支持 yes/no 格式
	if strings.Contains(contextLower, "yes/no") ||
	   strings.Contains(contextLower, "[yes/no]") {
		return "yes"
	}

	// 8. 🔧 NEW: 支持数字范围选择 (1-5)
	if strings.Contains(contextLower, "(1-") ||
	   strings.Contains(contextLower, "1-") && strings.Contains(contextLower, "):") {
		return "1" // 默认选择第一个选项
	}

	// 默认发送 yes
	return "yes"
}

// ShouldConfirm 综合判断是否应该自动确认
// 返回: (shouldConfirm bool, input string, reason string)
// 🔧 P1 FIX: 增强日志和统计
func (d *Detector) ShouldConfirm() (bool, string, string) {
	recent := d.GetRecentOutput(50)

	// 🔧 P1 FIX: 更新统计 - 总请求数
	d.confirmStats.TotalRequests++

	// 恢复安全检查 - 在自动确认前验证操作安全性
	if !d.SafeToConfirm() {
		reason := "检测到危险操作或无法判断安全性"
		// 🔧 P1 FIX: 增强日志格式
		log.Printf("[CONFIRMATION BLOCKED] Decision: REJECT | Reason: %s | Total: %d | Auto: %d | Manual: %d | Blocked: %d | Context: %.100s...",
			reason, d.confirmStats.TotalRequests, d.confirmStats.AutoConfirmed, d.confirmStats.ManualRequired, d.confirmStats.Blocked+1, recent)
		d.confirmStats.Blocked++
		return false, "", reason
	}

	// 额外的上下文检查 - 需要人工确认的特殊情况
	if d.requiresManualConfirmation(recent) {
		reason := "需要人工确认（特殊上下文）"
		// 🔧 P1 FIX: 增强日志格式
		log.Printf("[CONFIRMATION MANUAL] Decision: MANUAL | Reason: %s | Total: %d | Auto: %d | Manual: %d | Context: %.100s...",
			reason, d.confirmStats.TotalRequests, d.confirmStats.AutoConfirmed, d.confirmStats.ManualRequired+1, recent)
		d.confirmStats.ManualRequired++
		return false, "", reason
	}

	// 确定要发送的输入
	input := GetConfirmationInput(recent)

	// 🔧 P1 FIX: 更新统计和增强日志
	d.confirmStats.AutoConfirmed++
	d.confirmStats.LastConfirmTime = time.Now()

	// 记录自动确认决策（用于审计）
	log.Printf("[CONFIRMATION AUTO] Decision: CONFIRM | Input: '%s' | Total: %d | Auto: %d (%.1f%%) | Manual: %d | Blocked: %d | Context: %.100s...",
		input,
		d.confirmStats.TotalRequests,
		d.confirmStats.AutoConfirmed,
		float64(d.confirmStats.AutoConfirmed)/float64(d.confirmStats.TotalRequests)*100,
		d.confirmStats.ManualRequired,
		d.confirmStats.Blocked,
		recent)

	return true, input, "自动确认（已通过安全检查）"
}

// requiresManualConfirmation 检查是否需要人工确认
func (d *Detector) requiresManualConfirmation(context string) bool {
	contextLower := strings.ToLower(context)

	// 需要人工确认的关键词
	manualConfirmKeywords := []string{
		"irreversible",
		"cannot be undone",
		"permanent",
		"production",
		"live environment",
		"critical",
		"warning",
		"caution",
	}

	for _, keyword := range manualConfirmKeywords {
		if strings.Contains(contextLower, keyword) {
			return true
		}
	}

	return false
}
