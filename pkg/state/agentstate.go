package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/yourusername/claude-swarm/internal/models"
)

// AgentStateManager manages agent state persistence
type AgentStateManager struct {
	filePath string
	mu       sync.Mutex
	lockFile *os.File
	agents   map[string]*models.AgentStatus
}

type agentStateFile struct {
	Agents     []*models.AgentStatus `json:"agents"`
	UpdatedAt  time.Time             `json:"updated_at"`
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

	// Create or open lock file
	lockFilePath := filePath + ".lock"
	lockFile, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create lock file: %w", err)
	}

	asm := &AgentStateManager{
		filePath: filePath,
		lockFile: lockFile,
		agents:   make(map[string]*models.AgentStatus),
	}

	// Load existing state if available
	if err := asm.load(); err != nil {
		// If file doesn't exist, create an empty one
		if os.IsNotExist(err) {
			if err := asm.save(); err != nil {
				lockFile.Close()
				return nil, err
			}
		} else {
			lockFile.Close()
			return nil, err
		}
	}

	return asm, nil
}

// Close closes the agent state manager and releases the file lock
func (asm *AgentStateManager) Close() error {
	if asm.lockFile != nil {
		return asm.lockFile.Close()
	}
	return nil
}

// UpdateAgents updates all agent states
func (asm *AgentStateManager) UpdateAgents(agents []*models.AgentStatus) error {
	asm.mu.Lock()
	defer asm.mu.Unlock()

	// Update internal map
	asm.agents = make(map[string]*models.AgentStatus)
	for _, agent := range agents {
		// Create a copy to avoid race conditions
		agentCopy := &models.AgentStatus{
			AgentID:     agent.AgentID,
			State:       agent.State,
			CurrentTask: agent.CurrentTask,
			LastUpdate:  agent.LastUpdate,
			Output:      agent.Output,
		}
		asm.agents[agent.AgentID] = agentCopy
	}

	return asm.save()
}

// GetAgents returns all agent states
func (asm *AgentStateManager) GetAgents() ([]*models.AgentStatus, error) {
	asm.mu.Lock()
	defer asm.mu.Unlock()

	// Reload from file to get latest state
	if err := asm.load(); err != nil {
		return nil, err
	}

	agents := make([]*models.AgentStatus, 0, len(asm.agents))
	for _, agent := range asm.agents {
		agents = append(agents, agent)
	}

	return agents, nil
}

// load loads agent state from the JSON file
func (asm *AgentStateManager) load() error {
	// Acquire shared lock for reading
	if err := syscall.Flock(int(asm.lockFile.Fd()), syscall.LOCK_SH); err != nil {
		return fmt.Errorf("failed to acquire read lock: %w", err)
	}
	defer syscall.Flock(int(asm.lockFile.Fd()), syscall.LOCK_UN)

	data, err := os.ReadFile(asm.filePath)
	if err != nil {
		return err
	}

	var asf agentStateFile
	if err := json.Unmarshal(data, &asf); err != nil {
		return fmt.Errorf("failed to unmarshal agent state: %w", err)
	}

	asm.agents = make(map[string]*models.AgentStatus)
	for _, agent := range asf.Agents {
		asm.agents[agent.AgentID] = agent
	}

	return nil
}

// save saves agent state to the JSON file using atomic write
func (asm *AgentStateManager) save() error {
	// Acquire exclusive lock for writing
	if err := syscall.Flock(int(asm.lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed to acquire write lock: %w", err)
	}
	defer syscall.Flock(int(asm.lockFile.Fd()), syscall.LOCK_UN)

	agents := make([]*models.AgentStatus, 0, len(asm.agents))
	for _, agent := range asm.agents {
		agents = append(agents, agent)
	}

	asf := agentStateFile{
		Agents:    agents,
		UpdatedAt: time.Now(),
	}

	data, err := json.MarshalIndent(asf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal agent state: %w", err)
	}

	// Atomic write: write to temp file then rename
	tmpFile := asm.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename (overwrites target file atomically)
	if err := os.Rename(tmpFile, asm.filePath); err != nil {
		os.Remove(tmpFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
