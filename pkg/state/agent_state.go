package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/yourusername/claude-swarm/internal/models"
)

// AgentStateManager manages persistent agent state
type AgentStateManager struct {
	filePath string
	mu       sync.Mutex
}

type agentStateFile struct {
	Agents []*models.AgentStatus `json:"agents"`
}

// NewAgentStateManager creates a new agent state manager
func NewAgentStateManager(filePath string) (*AgentStateManager, error) {
	// Validate and expand file path
	if filePath == "" {
		return nil, fmt.Errorf("filePath cannot be empty")
	}

	// Expand ~ to home directory
	if len(filePath) >= 2 && filePath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		filePath = filepath.Join(home, filePath[2:])
	} else if filePath == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		filePath = home
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	asm := &AgentStateManager{
		filePath: filePath,
	}

	// Initialize file if it doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := asm.save(&agentStateFile{Agents: []*models.AgentStatus{}}); err != nil {
			return nil, fmt.Errorf("failed to initialize agent state file: %w", err)
		}
	}

	return asm, nil
}

// UpdateAgents updates the agent states in the file
func (asm *AgentStateManager) UpdateAgents(agents []*models.AgentStatus) error {
	asm.mu.Lock()
	defer asm.mu.Unlock()

	asf := &agentStateFile{
		Agents: agents,
	}

	return asm.save(asf)
}

// LoadAgents loads agent states from the file
func (asm *AgentStateManager) LoadAgents() ([]*models.AgentStatus, error) {
	asm.mu.Lock()
	defer asm.mu.Unlock()

	data, err := os.ReadFile(asm.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent state file: %w", err)
	}

	var asf agentStateFile
	if err := json.Unmarshal(data, &asf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent states: %w", err)
	}

	return asf.Agents, nil
}

// save saves agent states to the file using atomic write
func (asm *AgentStateManager) save(asf *agentStateFile) error {
	data, err := json.MarshalIndent(asf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal agent states: %w", err)
	}

	// Atomic write: write to temp file then rename
	tmpFile := asm.filePath + ".tmp"
	
	// Ensure cleanup of temp file in all cases
	defer func() {
		if _, err := os.Stat(tmpFile); err == nil {
			os.Remove(tmpFile)
		}
	}()

	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename (overwrites target file atomically)
	if err := os.Rename(tmpFile, asm.filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// Close closes the agent state manager (no-op for file-based implementation)
func (asm *AgentStateManager) Close() error {
	// No resources to clean up for file-based implementation
	return nil
}
