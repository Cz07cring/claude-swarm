package main

import (
	"fmt"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/analyzer"
	"github.com/yourusername/claude-swarm/pkg/retry"
)

func main() {
	rm := retry.NewRetryManager(retry.DefaultRetryConfig())

	fmt.Println("测试重试逻辑...")
	fmt.Println()

	// 测试 1: 可重试错误 (网络超时)
	fmt.Println("1. 测试可重试错误 (网络超时):")
	task1 := &models.Task{
		ID:         "task1",
		RetryCount: 0,
		MaxRetries: 3,
	}
	errorDetails1 := &analyzer.ErrorDetails{
		Type:    analyzer.ErrorTypeRetryable,
		Message: "Connection timeout",
		Context: "Error: ETIMEDOUT - connection timeout after 30s",
	}

	if rm.ShouldRetry(task1, errorDetails1) {
		fmt.Println("  ✓ 应该重试 (正确)")
		rm.RecordRetry(task1, errorDetails1)
		delay := rm.CalculateDelay(task1.RetryCount - 1)
		fmt.Printf("  ✓ 延迟时间: %v\n", delay)
		fmt.Printf("  ✓ 重试次数: %d/%d\n", task1.RetryCount, task1.MaxRetries)
	} else {
		fmt.Println("  ✗ 不应该重试 (错误!)")
	}
	fmt.Println()

	// 测试 2: 不可重试错误 (语法错误)
	fmt.Println("2. 测试不可重试错误 (语法错误):")
	task2 := &models.Task{
		ID:         "task2",
		RetryCount: 0,
		MaxRetries: 3,
	}
	errorDetails2 := &analyzer.ErrorDetails{
		Type:    analyzer.ErrorTypeNonRetryable,
		Message: "Syntax error",
		Context: "SyntaxError: Unexpected token",
	}

	if !rm.ShouldRetry(task2, errorDetails2) {
		fmt.Println("  ✓ 不应该重试 (正确)")
	} else {
		fmt.Println("  ✗ 不应该重试但却要重试 (错误!)")
	}
	fmt.Println()

	// 测试 3: 致命错误
	fmt.Println("3. 测试致命错误 (OOM):")
	task3 := &models.Task{
		ID:         "task3",
		RetryCount: 0,
		MaxRetries: 3,
	}
	errorDetails3 := &analyzer.ErrorDetails{
		Type:    analyzer.ErrorTypeFatal,
		Message: "Out of memory",
		Context: "Fatal: JavaScript heap out of memory",
	}

	if !rm.ShouldRetry(task3, errorDetails3) {
		fmt.Println("  ✓ 不应该重试 (正确)")
	} else {
		fmt.Println("  ✗ 不应该重试但却要重试 (错误!)")
	}
	fmt.Println()

	// 测试 4: 超过重试次数
	fmt.Println("4. 测试超过重试次数:")
	task4 := &models.Task{
		ID:         "task4",
		RetryCount: 3,
		MaxRetries: 3,
	}
	errorDetails4 := &analyzer.ErrorDetails{
		Type:    analyzer.ErrorTypeRetryable,
		Message: "Network error",
		Context: "Connection refused",
	}

	if !rm.ShouldRetry(task4, errorDetails4) {
		fmt.Println("  ✓ 已达重试上限，不再重试 (正确)")
	} else {
		fmt.Println("  ✗ 应该停止重试但却继续 (错误!)")
	}
	fmt.Println()

	// 测试 5: 指数退避延迟
	fmt.Println("5. 测试指数退避延迟:")
	for i := 0; i < 5; i++ {
		delay := rm.CalculateDelay(i)
		fmt.Printf("  重试 %d: 延迟 %v\n", i+1, delay)
	}
	fmt.Println()

	// 测试 6: 错误分析
	fmt.Println("6. 测试错误类型检测:")
	detector := analyzer.NewDetector()

	testCases := []struct {
		output       string
		expectedType analyzer.ErrorType
	}{
		{"Error: ETIMEDOUT", analyzer.ErrorTypeRetryable},
		{"SyntaxError: unexpected token", analyzer.ErrorTypeNonRetryable},
		{"Fatal: out of memory", analyzer.ErrorTypeFatal},
		{"Error: connection refused", analyzer.ErrorTypeRetryable},
		{"404 not found", analyzer.ErrorTypeNonRetryable},
		{"panic: runtime error", analyzer.ErrorTypeFatal},
	}

	allCorrect := true
	for _, tc := range testCases {
		details := detector.AnalyzeError(tc.output)
		if details.Type == tc.expectedType {
			fmt.Printf("  ✓ '%s' -> %v (正确)\n", tc.output[:min(len(tc.output), 30)], details.Type)
		} else {
			fmt.Printf("  ✗ '%s' -> got %v, expected %v (错误)\n", tc.output[:min(len(tc.output), 30)], details.Type, tc.expectedType)
			allCorrect = false
		}
	}

	if allCorrect {
		fmt.Println("\n✅ 所有重试逻辑测试通过！")
	} else {
		fmt.Println("\n❌ 部分测试失败")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
