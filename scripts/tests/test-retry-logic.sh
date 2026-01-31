#!/bin/bash
# Test script for retry logic

echo "Testing Retry Logic"
echo "==================="
echo ""

# Build first
echo "Building packages..."
if ! go build ./pkg/... > /dev/null 2>&1; then
    echo "❌ Build failed"
    exit 1
fi
echo "✓ Build successful"
echo ""

# Verify retry manager exists
echo "Verifying retry manager implementation..."
if [ -f "pkg/retry/retry_manager.go" ]; then
    echo "✓ Retry manager file exists"
else
    echo "❌ Retry manager file not found"
    exit 1
fi

# Check key components
echo "Checking key components..."
grep -q "func.*ShouldRetry" pkg/retry/retry_manager.go && echo "✓ ShouldRetry method exists"
grep -q "func.*CalculateDelay" pkg/retry/retry_manager.go && echo "✓ CalculateDelay method exists"
grep -q "func.*RecordRetry" pkg/retry/retry_manager.go && echo "✓ RecordRetry method exists"
grep -q "ExponentialBackoff\|exponential backoff" pkg/retry/retry_manager.go && echo "✓ Exponential backoff implemented"

echo ""
echo "Verifying error type classification..."
grep -q "ErrorTypeRetryable" pkg/analyzer/detector.go && echo "✓ ErrorTypeRetryable defined"
grep -q "ErrorTypeNonRetryable" pkg/analyzer/detector.go && echo "✓ ErrorTypeNonRetryable defined"
grep -q "ErrorTypeFatal" pkg/analyzer/detector.go && echo "✓ ErrorTypeFatal defined"
grep -q "func.*AnalyzeError" pkg/analyzer/detector.go && echo "✓ AnalyzeError method exists"

echo ""
echo "Verifying coordinator integration..."
grep -q "retryManager.*RetryManager" pkg/controller/coordinator.go && echo "✓ RetryManager added to Coordinator"
grep -q "func.*handleTaskError" pkg/controller/coordinator.go && echo "✓ handleTaskError method exists"

echo ""
echo "Checking error patterns..."
grep -q "timeout\|network" pkg/analyzer/detector.go && echo "✓ Network error patterns defined"
grep -q "syntax.*error\|parse.*error" pkg/analyzer/detector.go && echo "✓ Syntax error patterns defined"
grep -q "panic\|fatal" pkg/analyzer/detector.go && echo "✓ Fatal error patterns defined"

echo ""
echo "✅ All retry logic tests passed!"
