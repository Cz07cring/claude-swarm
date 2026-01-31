package retry

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
	"github.com/yourusername/claude-swarm/pkg/analyzer"
)

// RetryManager manages task retry logic
type RetryManager struct {
	config RetryConfig
}

// RetryConfig contains retry configuration
type RetryConfig struct {
	MaxRetries    int           // Maximum number of retries per task
	InitialDelay  time.Duration // Initial delay before first retry
	MaxDelay      time.Duration // Maximum delay between retries
	BackoffFactor float64       // Exponential backoff multiplier
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  5 * time.Second,
		MaxDelay:      5 * time.Minute,
		BackoffFactor: 2.0,
	}
}

// NewRetryManager creates a new retry manager
func NewRetryManager(config RetryConfig) *RetryManager {
	// Validate config
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.InitialDelay <= 0 {
		config.InitialDelay = 5 * time.Second
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = 5 * time.Minute
	}
	if config.BackoffFactor <= 1.0 {
		config.BackoffFactor = 2.0
	}

	return &RetryManager{
		config: config,
	}
}

// ShouldRetry determines if a task should be retried based on error type and retry count
func (rm *RetryManager) ShouldRetry(task *models.Task, errorDetails *analyzer.ErrorDetails) bool {
	// Check if we've exceeded max retries
	if task.RetryCount >= task.MaxRetries {
		log.Printf("[RETRY] Task %s exceeded max retries (%d/%d)", task.ID, task.RetryCount, task.MaxRetries)
		return false
	}

	// Check error type
	switch errorDetails.Type {
	case analyzer.ErrorTypeRetryable:
		// Retryable errors should be retried
		log.Printf("[RETRY] Task %s is retryable (retry %d/%d): %s",
			task.ID, task.RetryCount+1, task.MaxRetries, errorDetails.Message)
		return true

	case analyzer.ErrorTypeNonRetryable:
		// Non-retryable errors (syntax, logic) should not be retried
		log.Printf("[RETRY] Task %s has non-retryable error: %s", task.ID, errorDetails.Message)
		return false

	case analyzer.ErrorTypeFatal:
		// Fatal errors require human intervention
		log.Printf("[RETRY] Task %s has fatal error: %s", task.ID, errorDetails.Message)
		return false

	case analyzer.ErrorTypeUnknown:
		// Unknown errors: retry conservatively (only if retry count is low)
		if task.RetryCount < 2 {
			log.Printf("[RETRY] Task %s has unknown error, retrying cautiously (retry %d/%d)",
				task.ID, task.RetryCount+1, task.MaxRetries)
			return true
		}
		log.Printf("[RETRY] Task %s has unknown error and exceeded cautious retry limit", task.ID)
		return false

	default:
		return false
	}
}

// CalculateDelay calculates the delay before next retry using exponential backoff
func (rm *RetryManager) CalculateDelay(retryCount int) time.Duration {
	// Exponential backoff: delay = initialDelay * (backoffFactor ^ retryCount)
	delay := float64(rm.config.InitialDelay) * math.Pow(rm.config.BackoffFactor, float64(retryCount))

	// Cap at max delay
	if delay > float64(rm.config.MaxDelay) {
		delay = float64(rm.config.MaxDelay)
	}

	return time.Duration(delay)
}

// RecordRetry records a retry attempt on a task
func (rm *RetryManager) RecordRetry(task *models.Task, errorDetails *analyzer.ErrorDetails) {
	task.RetryCount++
	task.LastError = fmt.Sprintf("[%s] %s (Context: %.200s)",
		rm.errorTypeString(errorDetails.Type),
		errorDetails.Message,
		errorDetails.Context)

	log.Printf("[RETRY] Recorded retry for task %s: %d/%d retries, error: %s",
		task.ID, task.RetryCount, task.MaxRetries, errorDetails.Message)
}

// GetRetryInfo returns human-readable retry information for a task
func (rm *RetryManager) GetRetryInfo(task *models.Task) string {
	if task.RetryCount == 0 {
		return "No retries"
	}

	nextDelay := rm.CalculateDelay(task.RetryCount)
	return fmt.Sprintf("%d/%d retries, next delay: %v",
		task.RetryCount, task.MaxRetries, nextDelay)
}

// errorTypeString converts error type to string
func (rm *RetryManager) errorTypeString(errType analyzer.ErrorType) string {
	switch errType {
	case analyzer.ErrorTypeRetryable:
		return "RETRYABLE"
	case analyzer.ErrorTypeNonRetryable:
		return "NON_RETRYABLE"
	case analyzer.ErrorTypeFatal:
		return "FATAL"
	case analyzer.ErrorTypeUnknown:
		return "UNKNOWN"
	default:
		return "UNKNOWN"
	}
}
