package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 主配置结构
type Config struct {
	Gemini GeminiConfig `yaml:"gemini"`
	Swarm  SwarmConfig  `yaml:"swarm"`
	Git    GitConfig    `yaml:"git"`
}

// GeminiConfig Gemini API 配置
type GeminiConfig struct {
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
	Timeout int    `yaml:"timeout"`
}

// SwarmConfig Swarm 配置
type SwarmConfig struct {
	DefaultAgents   int    `yaml:"default_agents"`
	MonitorInterval int    `yaml:"monitor_interval"`
	SessionName     string `yaml:"session_name"`
	TaskQueuePath   string `yaml:"task_queue_path"`
}

// GitConfig Git 配置
type GitConfig struct {
	RepoPath     string `yaml:"repo_path"`
	WorktreesDir string `yaml:"worktrees_dir"`
	MainBranch   string `yaml:"main_branch"`
}

// Load 加载配置文件
// 优先级：1. 指定路径 2. ./config.yaml 3. ~/.claude-swarm/config.yaml 4. 环境变量
func Load(configPath string) (*Config, error) {
	config := &Config{
		// 默认值
		Gemini: GeminiConfig{
			Model:   "gemini-3-flash-preview",
			Timeout: 30,
		},
		Swarm: SwarmConfig{
			DefaultAgents:   3,
			MonitorInterval: 5,
			SessionName:     "claude-swarm",
			TaskQueuePath:   "~/.claude-swarm/tasks.json",
		},
		Git: GitConfig{
			RepoPath:     ".",
			WorktreesDir: ".worktrees",
			MainBranch:   "main",
		},
	}

	// 查找配置文件
	var configFile string
	if configPath != "" {
		// 1. 使用指定路径
		configFile = configPath
	} else {
		// 2. 尝试当前目录
		if _, err := os.Stat("config.yaml"); err == nil {
			configFile = "config.yaml"
		} else {
			// 3. 尝试 ~/.claude-swarm/config.yaml
			homeDir, err := os.UserHomeDir()
			if err == nil {
				homePath := filepath.Join(homeDir, ".claude-swarm", "config.yaml")
				if _, err := os.Stat(homePath); err == nil {
					configFile = homePath
				}
			}
		}
	}

	// 如果找到配置文件，读取它
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
		}
	}

	// 4. 环境变量覆盖（最高优先级）
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		config.Gemini.APIKey = apiKey
	}

	// 验证必填项
	if config.Gemini.APIKey == "" {
		return nil, fmt.Errorf("gemini api_key is required (set in config.yaml or GEMINI_API_KEY env var)")
	}

	return config, nil
}

// LoadOrDefault 加载配置，如果失败则使用默认值（从环境变量读取 API Key）
func LoadOrDefault() *Config {
	config, err := Load("")
	if err != nil {
		// 如果加载失败，使用默认值
		config = &Config{
			Gemini: GeminiConfig{
				APIKey:  os.Getenv("GEMINI_API_KEY"),
				Model:   "gemini-3-flash-preview",
				Timeout: 30,
			},
			Swarm: SwarmConfig{
				DefaultAgents:   3,
				MonitorInterval: 5,
				SessionName:     "claude-swarm",
				TaskQueuePath:   "~/.claude-swarm/tasks.json",
			},
			Git: GitConfig{
				RepoPath:     ".",
				WorktreesDir: ".worktrees",
				MainBranch:   "main",
			},
		}
	}
	return config
}
