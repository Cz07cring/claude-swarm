package analyzer

import (
	"log"
	"strings"
)

// GetConfirmationInput 根据提示类型返回应该发送的确认输入
func GetConfirmationInput(context string) string {
	contextLower := strings.ToLower(context)

	// 检查是否是选项列表格式
	if strings.Contains(context, "❯ 1. Yes") ||
	   strings.Contains(context, "1. Yes") {
		// 发送选项编号
		return "1"
	}

	// 检查是否是 (y/n) 格式
	if strings.Contains(contextLower, "(y/n)") {
		return "y"
	}

	// 检查是否是 yes/no 格式
	if strings.Contains(contextLower, "yes/no") ||
	   strings.Contains(contextLower, "[yes/no]") {
		return "yes"
	}

	// 默认发送 yes
	return "yes"
}

// ShouldConfirm 综合判断是否应该自动确认
// 返回: (shouldConfirm bool, input string, reason string)
func (d *Detector) ShouldConfirm() (bool, string, string) {
	recent := d.GetRecentOutput(50)

	// 恢复安全检查 - 在自动确认前验证操作安全性
	if !d.SafeToConfirm() {
		reason := "检测到危险操作或无法判断安全性"
		log.Printf("[SECURITY] Auto-confirmation blocked: %s | Context: %.100s...", reason, recent)
		return false, "", reason
	}

	// 额外的上下文检查 - 需要人工确认的特殊情况
	if d.requiresManualConfirmation(recent) {
		reason := "需要人工确认（特殊上下文）"
		log.Printf("[SECURITY] Auto-confirmation requires manual review: %s | Context: %.100s...", reason, recent)
		return false, "", reason
	}

	// 确定要发送的输入
	input := GetConfirmationInput(recent)

	// 记录自动确认决策（用于审计）
	log.Printf("[CONFIRMATION] Auto-confirming with input '%s' | Context: %.100s...", input, recent)

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
