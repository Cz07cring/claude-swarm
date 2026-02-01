package analyzer

import (
	"strings"
	"testing"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
)

// TestNewDetector tests detector creation
func TestNewDetector(t *testing.T) {
	d := NewDetector()

	if d == nil {
		t.Fatal("NewDetector() returned nil")
	}

	if d.contextWindow == nil {
		t.Error("contextWindow not initialized")
	}

	if cap(d.contextWindow) != ContextWindowSize {
		t.Errorf("contextWindow capacity = %d, want %d", cap(d.contextWindow), ContextWindowSize)
	}

	if d.lastOutput.IsZero() {
		t.Error("lastOutput not initialized")
	}
}

// TestAnalyzeWaitingConfirm tests confirmation detection
func TestAnalyzeWaitingConfirm(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected models.AgentState
	}{
		{
			name:     "yes/no prompt",
			output:   "Do you want to proceed? (yes/no)",
			expected: models.AgentStateWaitingConfirm,
		},
		{
			name:     "Y/N prompt",
			output:   "Continue with this plan? (Y/N)",
			expected: models.AgentStateWaitingConfirm,
		},
		{
			name:     "press enter",
			output:   "Press Enter to continue",
			expected: models.AgentStateWaitingConfirm,
		},
		{
			name:     "option list",
			output:   "❯ 1. Yes\n  2. No",
			expected: models.AgentStateWaitingConfirm,
		},
		{
			name:     "proceed with plan",
			output:   "Proceed with this plan?",
			expected: models.AgentStateWaitingConfirm,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector()
			state := d.Analyze(tt.output)

			if state != tt.expected {
				t.Errorf("Analyze() = %v, want %v\nOutput: %q", state, tt.expected, tt.output)
			}

			// Verify waitingConfirmSince is set
			if d.waitingConfirmSince.IsZero() {
				t.Error("waitingConfirmSince not set for confirmation state")
			}
		})
	}
}

// TestAnalyzeError tests error detection
func TestAnalyzeError(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{"error colon", "error: file not found"},
		{"failed to", "failed to connect to server"},
		{"cannot", "cannot read file"},
		{"exception", "exception occurred"},
		{"fatal", "fatal: repository not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector()
			state := d.Analyze(tt.output)

			if state != models.AgentStateError {
				t.Errorf("Analyze() = %v, want %v", state, models.AgentStateError)
			}
		})
	}
}

// TestAnalyzeIdle tests idle state detection
func TestAnalyzeIdle(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{"try asking", "❯ Try asking me anything"},
		{"welcome", "❯ Welcome to Claude"},
		{"shortcuts", "Press Ctrl+C for shortcuts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector()
			state := d.Analyze(tt.output)

			if state != models.AgentStateIdle {
				t.Errorf("Analyze() = %v, want %v", state, models.AgentStateIdle)
			}
		})
	}
}

// TestAnalyzeWorking tests working state detection
func TestAnalyzeWorking(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{"normal output", "Processing your request..."},
		{"tool call", "Using tool: Read"},
		{"thinking", "Let me analyze this code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector()
			state := d.Analyze(tt.output)

			if state != models.AgentStateWorking {
				t.Errorf("Analyze() = %v, want %v", state, models.AgentStateWorking)
			}
		})
	}
}

// TestAnalyzeStuck tests stuck detection
func TestAnalyzeStuck(t *testing.T) {
	d := NewDetector()

	// Set lastOutput to past threshold
	d.lastOutput = time.Now().Add(-StuckThreshold - time.Second)

	state := d.Analyze("")

	if state != models.AgentStateStuck {
		t.Errorf("Analyze() = %v, want %v", state, models.AgentStateStuck)
	}
}

// TestAnalyzeEmptyNotStuck tests empty output within threshold
func TestAnalyzeEmptyNotStuck(t *testing.T) {
	d := NewDetector()

	// Recent lastOutput
	d.lastOutput = time.Now().Add(-10 * time.Second)

	state := d.Analyze("")

	if state != models.AgentStateIdle {
		t.Errorf("Analyze() = %v, want %v", state, models.AgentStateIdle)
	}
}

// TestStateTransition tests state transitions
func TestStateTransition(t *testing.T) {
	d := NewDetector()

	// Start with working
	state1 := d.Analyze("Processing...")
	if state1 != models.AgentStateWorking {
		t.Errorf("State 1 = %v, want Working", state1)
	}

	// Transition to waiting confirm
	state2 := d.Analyze("Continue? (yes/no)")
	if state2 != models.AgentStateWaitingConfirm {
		t.Errorf("State 2 = %v, want WaitingConfirm", state2)
	}

	// Verify waitingConfirmSince is set
	if d.waitingConfirmSince.IsZero() {
		t.Error("waitingConfirmSince should be set")
	}

	// Transition back to working
	state3 := d.Analyze("Continuing with task...")
	if state3 != models.AgentStateWorking {
		t.Errorf("State 3 = %v, want Working", state3)
	}

	// Verify waitingConfirmSince is reset
	if !d.waitingConfirmSince.IsZero() {
		t.Error("waitingConfirmSince should be reset")
	}
}

// TestContextWindow tests context window management
func TestContextWindow(t *testing.T) {
	d := NewDetector()

	// Add unique first line that should be removed
	d.Analyze("FIRST_LINE_MARKER")

	// Add lines exceeding window size (201 more lines)
	for i := 0; i < ContextWindowSize+1; i++ {
		d.Analyze("Line " + string(rune('0'+i%10)))
	}

	// Verify window size is maintained
	if len(d.contextWindow) != ContextWindowSize {
		t.Errorf("contextWindow length = %d, want %d", len(d.contextWindow), ContextWindowSize)
	}

	// Verify first line was removed (should have been pushed out)
	context := d.GetContext()
	if strings.Contains(context, "FIRST_LINE_MARKER") {
		t.Error("Old lines (FIRST_LINE_MARKER) should have been removed from context window")
	}
}

// TestGetRecentOutput tests recent output retrieval
func TestGetRecentOutput(t *testing.T) {
	d := NewDetector()

	// Add some lines
	lines := []string{
		"Line 1",
		"Line 2",
		"Line 3",
		"Line 4",
		"Line 5",
	}

	for _, line := range lines {
		d.Analyze(line)
	}

	// Get last 3 lines
	recent := d.GetRecentOutput(3)
	expected := "Line 3\nLine 4\nLine 5"

	if recent != expected {
		t.Errorf("GetRecentOutput(3) = %q, want %q", recent, expected)
	}

	// Request more than available
	all := d.GetRecentOutput(100)
	if !strings.Contains(all, "Line 1") || !strings.Contains(all, "Line 5") {
		t.Error("GetRecentOutput should return all available lines when n > len")
	}

	// Request 0
	empty := d.GetRecentOutput(0)
	if empty != "" {
		t.Errorf("GetRecentOutput(0) = %q, want empty string", empty)
	}
}

// TestIsConfirmTimeout tests confirmation timeout detection
func TestIsConfirmTimeout(t *testing.T) {
	d := NewDetector()

	// Initially no timeout
	if d.IsConfirmTimeout() {
		t.Error("Should not timeout initially")
	}

	// Set waitingConfirmSince to past threshold
	d.waitingConfirmSince = time.Now().Add(-6 * time.Minute)

	if !d.IsConfirmTimeout() {
		t.Error("Should timeout after 5 minutes")
	}

	// Set waitingConfirmSince to within threshold
	d.waitingConfirmSince = time.Now().Add(-3 * time.Minute)

	if d.IsConfirmTimeout() {
		t.Error("Should not timeout before 5 minutes")
	}

	// Reset (zero time)
	d.waitingConfirmSince = time.Time{}

	if d.IsConfirmTimeout() {
		t.Error("Should not timeout when zero time")
	}
}

// TestGetConfirmWaitDuration tests wait duration tracking
func TestGetConfirmWaitDuration(t *testing.T) {
	d := NewDetector()

	// Initially zero duration
	if d.GetConfirmWaitDuration() != 0 {
		t.Error("Initial duration should be 0")
	}

	// Set waitingConfirmSince
	testDuration := 2 * time.Minute
	d.waitingConfirmSince = time.Now().Add(-testDuration)

	duration := d.GetConfirmWaitDuration()

	// Allow small tolerance (100ms)
	tolerance := 100 * time.Millisecond
	if duration < testDuration-tolerance || duration > testDuration+tolerance {
		t.Errorf("Duration = %v, want ~%v", duration, testDuration)
	}
}

// TestConfirmStatsIntegration tests statistics tracking
func TestConfirmStatsIntegration(t *testing.T) {
	d := NewDetector()

	// Initial stats
	stats := d.GetConfirmStats()
	if stats.TotalRequests != 0 {
		t.Error("Initial TotalRequests should be 0")
	}

	// Simulate confirmations through ShouldConfirm
	// Note: The new AI-based risk assessment system:
	// - Critical risk (rm -rf, drop database, etc): BLOCKED
	// - High risk (production + destructive): BLOCKED unless explicit safety context
	// - Medium risk (normal delete, git push, etc): AUTO CONFIRM
	// - Low risk (read, select, yes/no prompts): AUTO CONFIRM
	// Use fresh detector for each case to avoid context pollution
	testCases := []struct {
		context  string
		expected bool
	}{
		{"Proceed with this plan to create new files? (yes/no)", true},  // Low risk (yes/no pattern)
		{"Delete all files? (yes/no)", true},                            // Medium risk (delete keyword)
		{"Create new feature?\n❯ 1. Yes\n  2. No", true},                // Low risk (option list)
		{"Deploy to production environment. OK?", true},                 // Medium risk (production but not critical)
		{"Drop database?\n❯ 1. Yes\n  2. No", false},                    // Critical risk (drop database)
	}

	for _, tc := range testCases {
		// Use fresh detector to avoid context pollution
		testDetector := NewDetector()
		testDetector.Analyze(tc.context)
		shouldConfirm, _, _ := testDetector.ShouldConfirm()

		if shouldConfirm != tc.expected {
			t.Errorf("ShouldConfirm for %q = %v, want %v", tc.context, shouldConfirm, tc.expected)
		}

		// Accumulate stats to main detector
		if shouldConfirm {
			d.confirmStats.AutoConfirmed++
		} else {
			d.confirmStats.Blocked++
		}
		d.confirmStats.TotalRequests++
		if shouldConfirm {
			d.confirmStats.LastConfirmTime = time.Now()
		}
	}

	// Check final stats
	stats = d.GetConfirmStats()

	if stats.TotalRequests != 5 {
		t.Errorf("TotalRequests = %d, want 5", stats.TotalRequests)
	}

	// 4 auto confirmed (all except Drop database)
	if stats.AutoConfirmed != 4 {
		t.Errorf("AutoConfirmed = %d, want 4", stats.AutoConfirmed)
	}

	// 1 blocked (Drop database - critical risk)
	if stats.Blocked != 1 {
		t.Errorf("Blocked = %d, want 1", stats.Blocked)
	}

	if stats.LastConfirmTime.IsZero() {
		t.Error("LastConfirmTime should be set")
	}
}

// TestReset tests detector reset
func TestReset(t *testing.T) {
	d := NewDetector()

	// Add some data
	d.Analyze("Line 1")
	d.Analyze("Line 2")
	d.Analyze("Continue? (yes/no)")

	// Verify state before reset
	if len(d.contextWindow) == 0 {
		t.Error("Context should have data before reset")
	}

	if d.waitingConfirmSince.IsZero() {
		t.Error("waitingConfirmSince should be set before reset")
	}

	// Reset
	d.Reset()

	// Verify state after reset
	if len(d.contextWindow) != 0 {
		t.Error("Context should be empty after reset")
	}

	if !d.waitingConfirmSince.IsZero() {
		t.Error("waitingConfirmSince should be reset")
	}

	if d.lastOutput.IsZero() {
		t.Error("lastOutput should be refreshed")
	}
}

// TestResetConfirmStats tests statistics reset
func TestResetConfirmStats(t *testing.T) {
	d := NewDetector()

	// Add some stats
	d.confirmStats.TotalRequests = 10
	d.confirmStats.AutoConfirmed = 5
	d.confirmStats.Blocked = 2
	d.confirmStats.ManualRequired = 3
	d.confirmStats.LastConfirmTime = time.Now()

	// Reset stats
	d.ResetConfirmStats()

	// Verify reset
	stats := d.GetConfirmStats()
	if stats.TotalRequests != 0 {
		t.Errorf("TotalRequests after reset = %d, want 0", stats.TotalRequests)
	}
	if stats.AutoConfirmed != 0 {
		t.Errorf("AutoConfirmed after reset = %d, want 0", stats.AutoConfirmed)
	}
	if stats.Blocked != 0 {
		t.Errorf("Blocked after reset = %d, want 0", stats.Blocked)
	}
	if stats.ManualRequired != 0 {
		t.Errorf("ManualRequired after reset = %d, want 0", stats.ManualRequired)
	}
}

// TestAnalyzeErrorTypes tests error classification
func TestAnalyzeErrorTypes(t *testing.T) {
	tests := []struct {
		name         string
		output       string
		expectedType ErrorType
	}{
		// Retryable errors
		{
			name:         "timeout",
			output:       "error: connection timeout",
			expectedType: ErrorTypeRetryable,
		},
		{
			name:         "network error",
			output:       "network unreachable",
			expectedType: ErrorTypeRetryable,
		},
		{
			name:         "rate limit",
			output:       "rate limit exceeded, try again",
			expectedType: ErrorTypeRetryable,
		},
		{
			name:         "503 error",
			output:       "503 Service Unavailable",
			expectedType: ErrorTypeRetryable,
		},

		// Non-retryable errors
		{
			name:         "syntax error",
			output:       "syntax error: unexpected token",
			expectedType: ErrorTypeNonRetryable,
		},
		{
			name:         "not found",
			output:       "404 not found",
			expectedType: ErrorTypeNonRetryable,
		},
		{
			name:         "permission denied",
			output:       "permission denied",
			expectedType: ErrorTypeNonRetryable,
		},

		// Fatal errors
		{
			name:         "panic",
			output:       "panic: runtime error",
			expectedType: ErrorTypeFatal,
		},
		{
			name:         "out of memory",
			output:       "fatal: out of memory",
			expectedType: ErrorTypeFatal,
		},
		{
			name:         "disk full",
			output:       "error: no space left on device",
			expectedType: ErrorTypeFatal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector()
			details := d.AnalyzeError(tt.output)

			if details.Type != tt.expectedType {
				t.Errorf("AnalyzeError() type = %v, want %v\nOutput: %q",
					details.Type, tt.expectedType, tt.output)
			}

			if details.Context != tt.output {
				t.Error("Context should preserve original output")
			}

			if details.Message == "" {
				t.Error("Message should not be empty")
			}
		})
	}
}

// TestMultilineAnalysis tests analysis with multiline output
func TestMultilineAnalysis(t *testing.T) {
	d := NewDetector()

	output := `Running analysis...
Processing files...
Found 5 issues

Do you want to proceed with fixes? (yes/no)
`

	state := d.Analyze(output)

	if state != models.AgentStateWaitingConfirm {
		t.Errorf("Analyze() = %v, want WaitingConfirm", state)
	}

	// Verify all lines are in context
	context := d.GetContext()
	if !strings.Contains(context, "Running analysis") {
		t.Error("Context should contain all output lines")
	}
	if !strings.Contains(context, "yes/no") {
		t.Error("Context should contain confirmation prompt")
	}
}

// TestConcurrentAccess tests basic concurrent safety
func TestConcurrentAccess(t *testing.T) {
	d := NewDetector()

	done := make(chan bool)

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			d.Analyze("Line " + string(rune('A'+i%26)))
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			d.GetRecentOutput(10)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Just verify no panic occurred
	if len(d.contextWindow) > ContextWindowSize {
		t.Error("Context window exceeded size limit under concurrent access")
	}
}

// BenchmarkAnalyze benchmarks state detection
func BenchmarkAnalyze(b *testing.B) {
	d := NewDetector()
	output := "Do you want to proceed? (yes/no)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Analyze(output)
	}
}

// BenchmarkAnalyzeError benchmarks error classification
func BenchmarkAnalyzeError(b *testing.B) {
	d := NewDetector()
	output := "error: connection timeout - please try again later"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.AnalyzeError(output)
	}
}

// BenchmarkContextWindow benchmarks context management
func BenchmarkContextWindow(b *testing.B) {
	d := NewDetector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Analyze("Line of output for testing")
	}
}
