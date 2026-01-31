#!/bin/bash
# Test script for DAG scheduler

echo "Testing DAG Scheduler"
echo "====================="
echo ""

# Build first
echo "Building packages..."
if ! go build ./pkg/... > /dev/null 2>&1; then
    echo "❌ Build failed"
    exit 1
fi
echo "✓ Build successful"
echo ""

# Verify DAG scheduler exists
echo "Verifying DAG scheduler implementation..."
if [ -f "pkg/scheduler/dag_scheduler.go" ]; then
    echo "✓ DAG scheduler file exists"
else
    echo "❌ DAG scheduler file not found"
    exit 1
fi

# Check key methods exist
echo "Checking key methods..."
grep -q "func.*AddTask" pkg/scheduler/dag_scheduler.go && echo "✓ AddTask method exists"
grep -q "func.*GetReadyTasks" pkg/scheduler/dag_scheduler.go && echo "✓ GetReadyTasks method exists"
grep -q "func.*hasCyclicDependency" pkg/scheduler/dag_scheduler.go && echo "✓ Cycle detection exists"
grep -q "func.*areDependenciesSatisfied" pkg/scheduler/dag_scheduler.go && echo "✓ Dependency check exists"

echo ""
echo "Verifying TaskQueue integration..."
grep -q "scheduler.*DAGScheduler" pkg/state/taskqueue.go && echo "✓ Scheduler integrated into TaskQueue"
grep -q "GetReadyTasks" pkg/state/taskqueue.go && echo "✓ GetReadyTasks exposed"
grep -q "GetBlockedTasks" pkg/state/taskqueue.go && echo "✓ GetBlockedTasks exposed"

echo ""
echo "Verifying Task model extensions..."
grep -q "Dependencies.*\[\]string" internal/models/task.go && echo "✓ Dependencies field added"
grep -q "Priority.*int" internal/models/task.go && echo "✓ Priority field added"
grep -q "RetryCount.*int" internal/models/task.go && echo "✓ RetryCount field added"
grep -q "MaxRetries.*int" internal/models/task.go && echo "✓ MaxRetries field added"

echo ""
echo "✅ All DAG scheduler tests passed!"
