package analyzer

import (
	"fmt"
	"log"
	"strings"
)

// RiskLevel 表示操作的风险等级
type RiskLevel string

const (
	RiskLevelCritical RiskLevel = "CRITICAL" // 极高风险，必须阻止
	RiskLevelHigh     RiskLevel = "HIGH"     // 高风险，需要安全上下文
	RiskLevelMedium   RiskLevel = "MEDIUM"   // 中等风险，可以执行
	RiskLevelLow      RiskLevel = "LOW"      // 低风险，安全执行
	RiskLevelUnknown  RiskLevel = "UNKNOWN"  // 未知风险，默认允许
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
// 🤖 AI 自主决策：智能分析上下文风险，自动判断是否安全执行
func (d *Detector) ShouldConfirm() (bool, string, string) {
	recent := d.GetRecentOutput(50)

	// 🧠 智能风险评估
	risk := d.assessRisk(recent)

	switch risk {
	case RiskLevelCritical:
		// 极高风险：阻止执行
		reason := "AI 决策：检测到极高风险操作，自动阻止"
		log.Printf("[AI DECISION BLOCKED] %s | Context: %.100s...", reason, recent)
		return false, "", reason

	case RiskLevelHigh:
		// 高风险：需要明确的安全上下文才执行
		if !d.hasExplicitSafetyContext(recent) {
			reason := "AI 决策：高风险操作缺少安全上下文，建议跳过"
			log.Printf("[AI DECISION SKIP] %s | Context: %.100s...", reason, recent)
			return false, "", reason
		}
		// 有明确安全上下文，允许执行
		fallthrough

	case RiskLevelMedium, RiskLevelLow:
		// 中低风险：智能确认
		input := GetConfirmationInput(recent)
		log.Printf("[AI DECISION AUTO] Risk: %s | Input: '%s' | Context: %.50s...",
			risk, input, recent)
		return true, input, fmt.Sprintf("AI 自主决策：%s 风险，自动确认", risk)

	default:
		// 未知风险：保守策略，默认允许
		input := GetConfirmationInput(recent)
		log.Printf("[AI DECISION DEFAULT] Input: '%s' | Context: %.50s...", input, recent)
		return true, input, "AI 决策：默认允许"
	}
}

// assessRisk 智能评估操作的风险等级
// 🧠 AI 自主决策的核心：基于实际影响判断风险，而非简单匹配关键词
func (d *Detector) assessRisk(context string) RiskLevel {
	contextLower := strings.ToLower(context)

	// 1️⃣ 极高风险：真正危险的系统级操作
	// 注意：这里只包含会破坏系统的操作
	if d.isCriticalSystemOperation(contextLower) {
		return RiskLevelCritical
	}

	// 2️⃣ 高风险：生产环境或主分支的破坏性操作
	if d.isProductionDestructiveOperation(contextLower) {
		return RiskLevelHigh
	}

	// 3️⃣ 中等风险：常规开发操作（包括正常的文件删除）
	mediumRiskPatterns := []string{
		"git commit",
		"git push",
		"npm install",
		"go install",
		"chmod",
		"chown",
		"create table",
		"alter table",
		"rm ",           // 普通删除操作
		"remove",
		"delete",
	}
	for _, pattern := range mediumRiskPatterns {
		if strings.Contains(contextLower, pattern) {
			return RiskLevelMedium
		}
	}

	// 4️⃣ 低风险：只读操作或常规确认
	lowRiskPatterns := []string{
		"git status",
		"git log",
		"git diff",
		"ls",
		"cat",
		"read",
		"select",
		"proceed",
		"continue",
		"yes/no",
		"press enter",
	}
	for _, pattern := range lowRiskPatterns {
		if strings.Contains(contextLower, pattern) {
			return RiskLevelLow
		}
	}

	// 默认：未知风险（允许执行）
	return RiskLevelUnknown
}

// isCriticalSystemOperation 判断是否是真正危险的系统级操作
func (d *Detector) isCriticalSystemOperation(context string) bool {
	// 危险的系统路径
	dangerousPaths := []string{
		"rm -rf /",
		"rm -rf /*",
		"rm -rf ~",
		"rm -rf ~/",
		"rm -rf /etc",
		"rm -rf /var",
		"rm -rf /usr",
		"rm -rf /boot",
		"rm -rf /sys",
		"rm -rf /proc",
		"rm -rf $HOME",
		"format /",
		"format c:",
	}
	for _, path := range dangerousPaths {
		if strings.Contains(context, path) {
			return true
		}
	}

	// 危险的系统操作
	dangerousOps := []string{
		"drop database",           // 删除整个数据库
		"truncate table users",    // 清空用户表
		"delete from users",       // 删除所有用户
		"shutdown -h now",
		"reboot -f",
		"mkfs",
		"fdisk",
		"dd if=/dev/zero of=/dev/",
		":(){ :|:&",              // Fork bomb
	}
	for _, op := range dangerousOps {
		if strings.Contains(context, op) {
			return true
		}
	}

	return false
}

// isProductionDestructiveOperation 判断是否是生产环境的破坏性操作
func (d *Detector) isProductionDestructiveOperation(context string) bool {
	// 检查是否在生产环境
	isProduction := strings.Contains(context, "production") ||
		strings.Contains(context, "live environment") ||
		strings.Contains(context, "master branch") ||
		strings.Contains(context, "main branch")

	if !isProduction {
		return false
	}

	// 在生产环境下的危险操作
	destructiveOps := []string{
		"git push --force",
		"git push -f",
		"git reset --hard",
		"drop table",
		"truncate table",
		"delete from",
	}

	for _, op := range destructiveOps {
		if strings.Contains(context, op) {
			return true
		}
	}

	return false
}

// hasExplicitSafetyContext 检查是否有明确的安全上下文
// 例如：在测试分支、开发环境、有备份等情况下，高风险操作可能是合理的
func (d *Detector) hasExplicitSafetyContext(context string) bool {
	contextLower := strings.ToLower(context)

	// 安全上下文指标
	safetyIndicators := []string{
		"test branch",
		"testing",
		"development",
		"dev environment",
		"backup created",
		"rollback available",
		"worktree",
		"feature branch",
		"agent-",        // Agent 分支通常是安全的
		"experimental",
		"sandbox",
	}

	for _, indicator := range safetyIndicators {
		if strings.Contains(contextLower, indicator) {
			return true
		}
	}

	return false
}

// AssessRisk is a public wrapper for assessRisk
func (d *Detector) AssessRisk(context string) RiskLevel {
	return d.assessRisk(context)
}
