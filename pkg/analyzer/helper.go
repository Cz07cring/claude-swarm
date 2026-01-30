package analyzer

import "strings"

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
	if !d.SafeToConfirm() {
		return false, "", "检测到危险操作或无法判断安全性"
	}

	recent := d.GetRecentOutput(50)

	// 确定要发送的输入
	input := GetConfirmationInput(recent)

	return true, input, "安全检查通过"
}
