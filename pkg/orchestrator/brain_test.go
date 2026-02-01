package orchestrator

import (
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
