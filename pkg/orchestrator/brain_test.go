package orchestrator

import (
	"encoding/json"
	"testing"
)

func TestValidateDependencies(t *testing.T) {
	tests := []struct {
		name    string
		tasks   []*TaskSpec
		wantErr bool
	}{
		{
			name: "valid dependencies",
			tasks: []*TaskSpec{
				{ID: "task-1", Description: "First", Dependencies: []string{}},
				{ID: "task-2", Description: "Second", Dependencies: []string{"task-1"}},
				{ID: "task-3", Description: "Third", Dependencies: []string{"task-1", "task-2"}},
			},
			wantErr: false,
		},
		{
			name: "missing dependency",
			tasks: []*TaskSpec{
				{ID: "task-1", Description: "First", Dependencies: []string{}},
				{ID: "task-2", Description: "Second", Dependencies: []string{"task-999"}},
			},
			wantErr: true,
		},
		{
			name: "circular dependency - direct",
			tasks: []*TaskSpec{
				{ID: "task-1", Description: "First", Dependencies: []string{"task-2"}},
				{ID: "task-2", Description: "Second", Dependencies: []string{"task-1"}},
			},
			wantErr: true,
		},
		{
			name: "circular dependency - indirect",
			tasks: []*TaskSpec{
				{ID: "task-1", Description: "First", Dependencies: []string{"task-3"}},
				{ID: "task-2", Description: "Second", Dependencies: []string{"task-1"}},
				{ID: "task-3", Description: "Third", Dependencies: []string{"task-2"}},
			},
			wantErr: true,
		},
		{
			name: "self dependency",
			tasks: []*TaskSpec{
				{ID: "task-1", Description: "First", Dependencies: []string{"task-1"}},
			},
			wantErr: true,
		},
		{
			name:    "empty tasks",
			tasks:   []*TaskSpec{},
			wantErr: false,
		},
		{
			name: "complex valid DAG",
			tasks: []*TaskSpec{
				{ID: "task-1", Description: "Root", Dependencies: []string{}},
				{ID: "task-2", Description: "Child1", Dependencies: []string{"task-1"}},
				{ID: "task-3", Description: "Child2", Dependencies: []string{"task-1"}},
				{ID: "task-4", Description: "Merge", Dependencies: []string{"task-2", "task-3"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个不需要实际 API key 的 brain（仅用于测试验证功能）
			brain := &OrchestratorBrain{}

			result := &AnalysisResult{
				Tasks: tt.tasks,
			}

			err := brain.ValidateDependencies(result)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDependencies() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectCyclicDependencies(t *testing.T) {
	tests := []struct {
		name    string
		tasks   []*TaskSpec
		wantErr bool
	}{
		{
			name: "no cycle",
			tasks: []*TaskSpec{
				{ID: "A", Dependencies: []string{}},
				{ID: "B", Dependencies: []string{"A"}},
				{ID: "C", Dependencies: []string{"B"}},
			},
			wantErr: false,
		},
		{
			name: "simple cycle",
			tasks: []*TaskSpec{
				{ID: "A", Dependencies: []string{"B"}},
				{ID: "B", Dependencies: []string{"A"}},
			},
			wantErr: true,
		},
		{
			name: "three-node cycle",
			tasks: []*TaskSpec{
				{ID: "A", Dependencies: []string{"B"}},
				{ID: "B", Dependencies: []string{"C"}},
				{ID: "C", Dependencies: []string{"A"}},
			},
			wantErr: true,
		},
		{
			name: "diamond pattern (no cycle)",
			tasks: []*TaskSpec{
				{ID: "A", Dependencies: []string{}},
				{ID: "B", Dependencies: []string{"A"}},
				{ID: "C", Dependencies: []string{"A"}},
				{ID: "D", Dependencies: []string{"B", "C"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brain := &OrchestratorBrain{}

			err := brain.detectCyclicDependencies(tt.tasks)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectCyclicDependencies() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCleanJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain JSON",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with markdown code block",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with only backticks",
			input:    "```\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with extra whitespace",
			input:    "   {\"key\": \"value\"}   ",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with markdown and whitespace",
			input:    "  ```json\n  {\"key\": \"value\"}  \n```  ",
			expected: `{"key": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanJSONResponse(tt.input)
			if result != tt.expected {
				t.Errorf("cleanJSONResponse() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseAnalysisResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid JSON",
			input: `{
				"summary": "Test summary",
				"complexity": "medium",
				"estimated_time": "2h",
				"modules": [],
				"tasks": [
					{
						"id": "task-1",
						"description": "Test task",
						"module": "test",
						"priority": 5,
						"estimated": "1h"
					}
				],
				"dependencies": {}
			}`,
			wantErr: false,
		},
		{
			name: "valid JSON with markdown",
			input: "```json\n" + `{
				"summary": "Test",
				"tasks": []
			}` + "\n```",
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{invalid json}`,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brain := &OrchestratorBrain{}

			result, err := brain.parseAnalysisResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAnalysisResponse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && result == nil {
				t.Error("parseAnalysisResponse() returned nil result without error")
			}
		})
	}
}

func TestTaskSpecIDGeneration(t *testing.T) {
	brain := &OrchestratorBrain{}

	input := `{
		"summary": "Test",
		"tasks": [
			{"description": "Task without ID", "module": "test"},
			{"id": "custom-id", "description": "Task with ID", "module": "test"}
		]
	}`

	result, err := brain.parseAnalysisResponse(input)
	if err != nil {
		t.Fatalf("parseAnalysisResponse() failed: %v", err)
	}

	if len(result.Tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(result.Tasks))
	}

	// 第一个任务应该有自动生成的ID
	if result.Tasks[0].ID == "" {
		t.Error("Expected auto-generated ID for first task, got empty string")
	}

	// 第二个任务应该保留自定义ID
	if result.Tasks[1].ID != "custom-id" {
		t.Errorf("Expected custom-id for second task, got %s", result.Tasks[1].ID)
	}
}

// ============== Smart Merge Functionality Tests ==============

func TestMergeDecisionTypes(t *testing.T) {
	// Test MergeDecision struct initialization and JSON serialization
	decision := MergeDecision{
		ShouldMerge:     true,
		MergeOrder:      []string{"agent-0-branch", "agent-1-branch"},
		Reason:          "All branches are ready to merge",
		PotentialIssues: []string{"Possible conflict in file.go"},
	}

	if !decision.ShouldMerge {
		t.Error("Expected ShouldMerge to be true")
	}
	if len(decision.MergeOrder) != 2 {
		t.Errorf("Expected 2 branches in MergeOrder, got %d", len(decision.MergeOrder))
	}
	if decision.MergeOrder[0] != "agent-0-branch" {
		t.Errorf("Expected first branch to be 'agent-0-branch', got %s", decision.MergeOrder[0])
	}
	if len(decision.PotentialIssues) != 1 {
		t.Errorf("Expected 1 potential issue, got %d", len(decision.PotentialIssues))
	}
}

func TestConflictResolutionTypes(t *testing.T) {
	// Test ConflictResolution struct
	resolution := ConflictResolution{
		CanAutoResolve: false,
		Resolution:     "Manual merge required",
		FileResolutions: map[string]string{
			"main.go": "Keep both changes",
			"util.go": "Use current branch",
		},
		NeedsHumanReview: true,
		Reason:           "Complex logic conflict",
	}

	if resolution.CanAutoResolve {
		t.Error("Expected CanAutoResolve to be false")
	}
	if len(resolution.FileResolutions) != 2 {
		t.Errorf("Expected 2 file resolutions, got %d", len(resolution.FileResolutions))
	}
	if resolution.FileResolutions["main.go"] != "Keep both changes" {
		t.Errorf("Unexpected file resolution for main.go")
	}
	if !resolution.NeedsHumanReview {
		t.Error("Expected NeedsHumanReview to be true")
	}
}

func TestMergeStatusTypes(t *testing.T) {
	// Test MergeStatus struct
	status := MergeStatus{
		Branch:       "agent-0-feature-branch",
		AgentID:      "agent-0",
		HasChanges:   true,
		CommitCount:  5,
		Files:        []string{"main.go", "util.go", "test.go"},
		ReadyToMerge: true,
	}

	if status.Branch != "agent-0-feature-branch" {
		t.Errorf("Expected branch 'agent-0-feature-branch', got %s", status.Branch)
	}
	if status.AgentID != "agent-0" {
		t.Errorf("Expected agent ID 'agent-0', got %s", status.AgentID)
	}
	if !status.HasChanges {
		t.Error("Expected HasChanges to be true")
	}
	if status.CommitCount != 5 {
		t.Errorf("Expected 5 commits, got %d", status.CommitCount)
	}
	if len(status.Files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(status.Files))
	}
	if !status.ReadyToMerge {
		t.Error("Expected ReadyToMerge to be true")
	}
}

func TestDecideMergeStrategyNoStatus(t *testing.T) {
	// Test DecideMergeStrategy with empty status list (no API call needed)
	brain := &OrchestratorBrain{}

	decision, err := brain.DecideMergeStrategy(nil, []*MergeStatus{})
	if err != nil {
		t.Fatalf("DecideMergeStrategy failed: %v", err)
	}

	if decision.ShouldMerge {
		t.Error("Expected ShouldMerge to be false for empty status list")
	}
	if decision.Reason != "没有待合并的分支" {
		t.Errorf("Unexpected reason: %s", decision.Reason)
	}
}

func TestMergeStatusReadyToMergeLogic(t *testing.T) {
	tests := []struct {
		name         string
		hasChanges   bool
		commitCount  int
		expectReady  bool
	}{
		{
			name:         "has commits, clean worktree",
			hasChanges:   false, // clean
			commitCount:  3,
			expectReady:  true, // CommitCount > 0 && !HasChanges
		},
		{
			name:         "has commits, dirty worktree",
			hasChanges:   true, // dirty
			commitCount:  3,
			expectReady:  false, // has uncommitted changes
		},
		{
			name:         "no commits, clean worktree",
			hasChanges:   false,
			commitCount:  0,
			expectReady:  false, // no commits to merge
		},
		{
			name:         "no commits, dirty worktree",
			hasChanges:   true,
			commitCount:  0,
			expectReady:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := MergeStatus{
				HasChanges:  tt.hasChanges,
				CommitCount: tt.commitCount,
			}
			// Calculate ReadyToMerge using the same logic as coordinator
			status.ReadyToMerge = status.CommitCount > 0 && !status.HasChanges

			if status.ReadyToMerge != tt.expectReady {
				t.Errorf("ReadyToMerge = %v, want %v", status.ReadyToMerge, tt.expectReady)
			}
		})
	}
}

func TestParseMergeDecisionJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*MergeDecision) bool
	}{
		{
			name: "valid merge decision",
			input: `{
				"should_merge": true,
				"merge_order": ["branch-a", "branch-b"],
				"reason": "All ready",
				"potential_issues": ["conflict in x.go"]
			}`,
			wantErr: false,
			check: func(d *MergeDecision) bool {
				return d.ShouldMerge && len(d.MergeOrder) == 2 && len(d.PotentialIssues) == 1
			},
		},
		{
			name: "valid merge decision with markdown",
			input: "```json\n" + `{
				"should_merge": false,
				"merge_order": [],
				"reason": "Not ready"
			}` + "\n```",
			wantErr: false,
			check: func(d *MergeDecision) bool {
				return !d.ShouldMerge && len(d.MergeOrder) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaned := cleanJSONResponse(tt.input)
			var decision MergeDecision
			err := json.Unmarshal([]byte(cleaned), &decision)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.check != nil && !tt.check(&decision) {
				t.Error("Check failed for parsed decision")
			}
		})
	}
}

func TestParseConflictResolutionJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid conflict resolution",
			input: `{
				"can_auto_resolve": false,
				"resolution": "Manual merge",
				"file_resolutions": {"a.go": "keep both"},
				"needs_human_review": true,
				"reason": "Complex conflict"
			}`,
			wantErr: false,
		},
		{
			name: "auto resolvable conflict",
			input: `{
				"can_auto_resolve": true,
				"resolution": "Use theirs",
				"file_resolutions": {},
				"needs_human_review": false,
				"reason": "Simple change"
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaned := cleanJSONResponse(tt.input)
			var resolution ConflictResolution
			err := json.Unmarshal([]byte(cleaned), &resolution)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
