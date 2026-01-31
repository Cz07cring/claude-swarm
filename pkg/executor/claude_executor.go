package executor

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/analyzer"
)

// ClaudeExecutor executes tasks using Claude Code CLI with echo pipe
type ClaudeExecutor struct {
	workDir  string
	detector *analyzer.Detector
	mu       sync.Mutex
}

// NewClaudeExecutor creates a new Claude executor
func NewClaudeExecutor(workDir string) *ClaudeExecutor {
	return &ClaudeExecutor{
		workDir:  workDir,
		detector: analyzer.NewDetector(),
	}
}

// ExecuteTask executes a task using Claude Code CLI
// Uses: echo "task" | claude --dangerously-skip-permissions
func (ce *ClaudeExecutor) ExecuteTask(ctx context.Context, task *models.Task) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	log.Printf("🤖 [%s] Executing task: %s", ce.workDir, task.ID)

	// 1. AI pre-assessment: check task risk before execution
	risk := ce.assessTaskRisk(task.Description)
	if risk == analyzer.RiskLevelCritical {
		log.Printf("🚫 [%s] AI blocked task: CRITICAL risk detected", ce.workDir)
		return fmt.Errorf("AI blocked: critical risk operation detected")
	}

	log.Printf("🧠 [%s] AI risk assessment: %s - proceeding", ce.workDir, risk)

	// 2. Prepare command
	// Escape single quotes in task description
	escapedTask := strings.ReplaceAll(task.Description, "'", "'\\''")

	// Build command: echo 'task' | claude --dangerously-skip-permissions
	cmdStr := fmt.Sprintf("echo '%s' | claude --dangerously-skip-permissions", escapedTask)

	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	cmd.Dir = ce.workDir

	// 3. Execute and capture output
	startTime := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	outputStr := string(output)

	// 4. Log execution details
	log.Printf("⏱️  [%s] Task completed in %s", ce.workDir, duration)
	log.Printf("📤 [%s] Claude output:\n%s", ce.workDir, outputStr)

	// 5. Analyze output for errors
	ce.detector.Analyze(outputStr)

	if err != nil {
		errorDetails := ce.detector.AnalyzeError(outputStr)
		log.Printf("❌ [%s] Task failed: %v (Error type: %v)", ce.workDir, err, errorDetails.Type)

		// Return error with type information for retry logic
		if errorDetails.Type == analyzer.ErrorTypeRetryable {
			return &RetryableError{
				Original: err,
				Details:  errorDetails,
			}
		}

		return fmt.Errorf("task execution failed: %w", err)
	}

	log.Printf("✅ [%s] Task %s completed successfully", ce.workDir, task.ID)
	return nil
}

// assessTaskRisk performs AI risk assessment on task description
func (ce *ClaudeExecutor) assessTaskRisk(description string) analyzer.RiskLevel {
	// Simulate analyzing the task description
	ce.detector.Analyze(description)

	// Use the same risk assessment logic
	risk := ce.detector.AssessRisk(description)

	return risk
}

// GetRecentOutput returns recent output for debugging
func (ce *ClaudeExecutor) GetRecentOutput(lines int) string {
	return ce.detector.GetRecentOutput(lines)
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Original error
	Details  *analyzer.ErrorDetails
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable error: %v", e.Original)
}

func (e *RetryableError) Unwrap() error {
	return e.Original
}
