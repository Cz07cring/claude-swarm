package analyzer

import (
	"testing"
	"time"
)

// TestGetConfirmationInput tests input format detection
func TestGetConfirmationInput(t *testing.T) {
	tests := []struct {
		name     string
		context  string
		expected string
	}{
		// 选项列表格式
		{
			name:     "option list with arrow 1",
			context:  "❯ 1. Yes\n  2. No",
			expected: "1",
		},
		{
			name:     "option list with arrow 2",
			context:  "► 1. Yes\n  2. No",
			expected: "1",
		},
		{
			name:     "option list plain",
			context:  "1. Yes\n2. No",
			expected: "1",
		},

		// Press Enter 格式
		{
			name:     "press enter lowercase",
			context:  "Press Enter to continue",
			expected: "",
		},
		{
			name:     "hit enter",
			context:  "Hit Enter when ready",
			expected: "",
		},
		{
			name:     "enter to continue",
			context:  "Enter to continue",
			expected: "",
		},

		// (Y/N) 格式
		{
			name:     "uppercase Y/N",
			context:  "Continue? (Y/N)",
			expected: "Y",
		},

		// (y/n) 格式
		{
			name:     "lowercase y/n",
			context:  "Continue? (y/n)",
			expected: "y",
		},

		// [Y/n] 格式（默认 Yes）
		{
			name:     "Y default",
			context:  "Proceed? [Y/n]",
			expected: "Y",
		},

		// [y/N] 格式（默认 No）
		{
			name:     "N default",
			context:  "Proceed? [y/N]",
			expected: "y",
		},

		// yes/no 格式
		{
			name:     "yes/no",
			context:  "Proceed? (yes/no)",
			expected: "yes",
		},
		{
			name:     "[yes/no]",
			context:  "Continue? [yes/no]",
			expected: "yes",
		},

		// 数字范围格式
		{
			name:     "number range (1-5)",
			context:  "Select an option (1-5):",
			expected: "1",
		},
		{
			name:     "number range 1-3",
			context:  "Choose: 1-3):",
			expected: "1",
		},

		// 默认格式
		{
			name:     "default yes",
			context:  "Some other format",
			expected: "yes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetConfirmationInput(tt.context)
			if result != tt.expected {
				t.Errorf("GetConfirmationInput() = %q, want %q\nContext: %q",
					result, tt.expected, tt.context)
			}
		})
	}
}

// TestDetectorSafeToConfirm tests the safety check logic
func TestDetectorSafeToConfirm(t *testing.T) {
	tests := []struct {
		name       string
		context    string
		shouldAllow bool
	}{
		// 应该允许的安全操作
		{
			name:        "plan confirmation safe",
			context:     "Proceed with this plan? This will create new files.",
			shouldAllow: true,
		},
		{
			name:        "option list with read action",
			context:     "Read the file?\n❯ 1. Yes\n  2. No",
			shouldAllow: true,
		},
		{
			name:        "option list with analyze action",
			context:     "Analyze the code?\n❯ 1. Yes\n  2. No",
			shouldAllow: true,
		},

		// 应该阻止的危险操作
		{
			name:        "delete operation",
			context:     "This will delete all files. Proceed? (yes/no)",
			shouldAllow: false,
		},
		{
			name:        "rm -rf command",
			context:     "Execute: rm -rf /tmp/data. Continue? [y/N]",
			shouldAllow: false,
		},
		{
			name:        "force push",
			context:     "git push --force to main. Are you sure?",
			shouldAllow: false,
		},
		{
			name:        "DROP TABLE",
			context:     "DROP TABLE users will be executed. Confirm?",
			shouldAllow: false,
		},
		{
			name:        "sudo rm",
			context:     "Run: sudo rm /etc/passwd. Proceed?",
			shouldAllow: false,
		},
		{
			name:        "chmod 777",
			context:     "Set permissions to 777. Continue? (Y/N)",
			shouldAllow: false,
		},

		// 需要人工确认的特殊上下文
		{
			name:        "irreversible operation",
			context:     "This action is irreversible. Proceed? (yes/no)",
			shouldAllow: false,
		},
		{
			name:        "cannot be undone",
			context:     "This cannot be undone. Continue? [y/N]",
			shouldAllow: false,
		},
		{
			name:        "production environment",
			context:     "Deploy to production environment. Confirm?",
			shouldAllow: false,
		},
		{
			name:        "permanent change",
			context:     "This is a permanent change. Proceed? (Y/N)",
			shouldAllow: false,
		},

		// 覆盖文件操作
		{
			name:        "overwrite file",
			context:     "File exists. Overwrite? (yes/no)",
			shouldAllow: false,
		},
		{
			name:        "replace file",
			context:     "Replace existing file? [Y/n]",
			shouldAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector()

			// 模拟输出
			d.Analyze(tt.context)

			result := d.SafeToConfirm()
			if result != tt.shouldAllow {
				t.Errorf("SafeToConfirm() = %v, want %v\nContext: %q",
					result, tt.shouldAllow, tt.context)
			}
		})
	}
}

// TestDetectorShouldConfirm tests the comprehensive confirmation logic
func TestDetectorShouldConfirm(t *testing.T) {
	tests := []struct {
		name           string
		context        string
		shouldConfirm  bool
		expectedInput  string
	}{
		{
			name:           "safe plan confirmation",
			context:        "Proceed with this plan? This will create new files. (yes/no)",
			shouldConfirm:  true,
			expectedInput:  "yes",
		},
		{
			name:           "medium risk delete - auto confirmed",
			context:        "Delete all files? (yes/no)",
			shouldConfirm:  true,
			expectedInput:  "yes",
		},
		{
			name:           "production deployment - auto confirmed",
			context:        "Deploy to production environment. Continue? [Y/n]",
			shouldConfirm:  true,
			expectedInput:  "Y", // [Y/n] format returns Y
		},
		{
			name:           "critical drop database - blocked",
			context:        "Drop database? (yes/no)",
			shouldConfirm:  false,
			expectedInput:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector()
			d.Analyze(tt.context)

			shouldConfirm, input, _ := d.ShouldConfirm()

			if shouldConfirm != tt.shouldConfirm {
				t.Errorf("ShouldConfirm() decision = %v, want %v", shouldConfirm, tt.shouldConfirm)
			}

			if shouldConfirm && input != tt.expectedInput {
				t.Errorf("ShouldConfirm() input = %q, want %q", input, tt.expectedInput)
			}
		})
	}
}


// TestConfirmationStatistics tests the statistics tracking
func TestConfirmationStatistics(t *testing.T) {
	d := NewDetector()

	// 初始状态
	stats := d.GetConfirmStats()
	if stats.TotalRequests != 0 {
		t.Errorf("Initial TotalRequests = %d, want 0", stats.TotalRequests)
	}

	// 模拟几次确认请求
	// Note: The new AI-based risk assessment auto-confirms most operations except critical ones
	// - Low/Medium risk: auto-confirmed
	// - Critical risk (rm -rf, drop database): blocked
	contexts := []struct {
		context      string
		shouldAuto   bool
	}{
		{"Proceed with this plan to create files? (yes/no)", true},  // Low risk, auto
		{"Delete all? (yes/no)", true},                              // Medium risk, auto (normal delete)
		{"Read the file?\n❯ 1. Yes\n  2. No", true},                 // Default, auto
		{"Drop database production?", false},                         // Critical risk, blocked
	}

	for _, tc := range contexts {
		// Use fresh detector to avoid context pollution
		testDetector := NewDetector()
		testDetector.Analyze(tc.context)
		shouldConfirm, _, _ := testDetector.ShouldConfirm()

		// Accumulate stats to main detector
		d.confirmStats.TotalRequests++
		if shouldConfirm {
			d.confirmStats.AutoConfirmed++
			d.confirmStats.LastConfirmTime = time.Now()
		}
	}

	// 检查统计
	stats = d.GetConfirmStats()
	if stats.TotalRequests != 4 {
		t.Errorf("TotalRequests = %d, want 4", stats.TotalRequests)
	}
	// 3 auto confirmed (all except Drop database which is critical)
	if stats.AutoConfirmed != 3 {
		t.Errorf("AutoConfirmed = %d, want 3", stats.AutoConfirmed)
	}

	// 测试重置
	d.ResetConfirmStats()
	stats = d.GetConfirmStats()
	if stats.TotalRequests != 0 {
		t.Errorf("After reset TotalRequests = %d, want 0", stats.TotalRequests)
	}
}

// BenchmarkGetConfirmationInput benchmarks input detection
func BenchmarkGetConfirmationInput(b *testing.B) {
	context := "Do you want to proceed? (yes/no)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetConfirmationInput(context)
	}
}

// BenchmarkShouldConfirm benchmarks the full confirmation logic
func BenchmarkShouldConfirm(b *testing.B) {
	d := NewDetector()
	context := "Create new file? This will add a file to the project. (yes/no)"
	d.Analyze(context)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.ShouldConfirm()
	}
}
