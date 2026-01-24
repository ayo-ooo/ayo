// Package delegates provides functionality for resolving agent delegation.
// Delegation allows agents to hand off specific task types to other agents.
//
// Delegation resolution follows a priority chain:
//  1. Directory config (.ayo.json in cwd or parent)
//  2. Agent config (delegates field in agent's config.json)
//  3. Global config (delegates field in ~/.config/ayo/ayo.json)
package delegates

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/paths"
)

// TaskType represents a type of task that can be delegated.
type TaskType string

// Common task types
const (
	TaskTypeCoding   TaskType = "coding"
	TaskTypeResearch TaskType = "research"
	TaskTypeDebug    TaskType = "debug"
	TaskTypeTest     TaskType = "test"
	TaskTypeDocs     TaskType = "docs"
)

// DirectoryConfig represents the .ayo.json configuration file.
type DirectoryConfig struct {
	// Delegates maps task types to agent handles at the directory level.
	Delegates map[string]string `json:"delegates,omitempty"`

	// Model overrides the default model for this directory.
	Model string `json:"model,omitempty"`

	// Agent specifies the default agent for this directory.
	Agent string `json:"agent,omitempty"`
}

// Resolution contains the result of resolving a delegation.
type Resolution struct {
	// Agent is the resolved agent handle (e.g., "@crush").
	Agent string

	// Source indicates where the delegation was resolved from.
	Source ResolutionSource

	// Path is the path to the config file if applicable.
	Path string
}

// ResolutionSource indicates where a delegation was resolved from.
type ResolutionSource int

const (
	// SourceNone means no delegation was configured.
	SourceNone ResolutionSource = iota
	// SourceDirectory means the delegation came from .ayo.json.
	SourceDirectory
	// SourceAgent means the delegation came from the agent's config.
	SourceAgent
	// SourceGlobal means the delegation came from the global config.
	SourceGlobal
)

func (s ResolutionSource) String() string {
	switch s {
	case SourceDirectory:
		return "directory"
	case SourceAgent:
		return "agent"
	case SourceGlobal:
		return "global"
	default:
		return "none"
	}
}

// Resolve finds the appropriate agent for a task type using the priority chain.
// Returns empty Resolution if no delegation is configured.
func Resolve(taskType TaskType, agentDelegates map[string]string, globalCfg config.Config) Resolution {
	taskStr := string(taskType)

	// 1. Check directory config
	dirConfig, dirPath := LoadDirectoryConfig()
	if dirConfig != nil {
		if agent, ok := dirConfig.Delegates[taskStr]; ok && agent != "" {
			return Resolution{
				Agent:  normalizeHandle(agent),
				Source: SourceDirectory,
				Path:   dirPath,
			}
		}
	}

	// 2. Check agent config
	if agentDelegates != nil {
		if agent, ok := agentDelegates[taskStr]; ok && agent != "" {
			return Resolution{
				Agent:  normalizeHandle(agent),
				Source: SourceAgent,
			}
		}
	}

	// 3. Check global config
	if globalCfg.Delegates != nil {
		if agent, ok := globalCfg.Delegates[taskStr]; ok && agent != "" {
			return Resolution{
				Agent:  normalizeHandle(agent),
				Source: SourceGlobal,
				Path:   paths.ConfigFile(),
			}
		}
	}

	return Resolution{}
}

// ResolveWithFallback is like Resolve but returns a fallback agent if no delegation is found.
func ResolveWithFallback(taskType TaskType, agentDelegates map[string]string, globalCfg config.Config, fallback string) Resolution {
	res := Resolve(taskType, agentDelegates, globalCfg)
	if res.Agent == "" {
		return Resolution{
			Agent:  normalizeHandle(fallback),
			Source: SourceNone,
		}
	}
	return res
}

// LoadDirectoryConfig loads the .ayo.json from the current directory or parents.
// Returns nil if no config file is found.
func LoadDirectoryConfig() (*DirectoryConfig, string) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, ""
	}

	configPath := paths.FindDirectoryConfig(wd)
	if configPath == "" {
		return nil, ""
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, ""
	}

	var dirConfig DirectoryConfig
	if err := json.Unmarshal(data, &dirConfig); err != nil {
		return nil, ""
	}

	return &dirConfig, configPath
}

// LoadDirectoryConfigFrom loads .ayo.json from a specific directory.
func LoadDirectoryConfigFrom(dir string) (*DirectoryConfig, error) {
	configPath := filepath.Join(dir, ".ayo.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var dirConfig DirectoryConfig
	if err := json.Unmarshal(data, &dirConfig); err != nil {
		return nil, err
	}

	return &dirConfig, nil
}

// GetAllDelegates returns a merged map of all delegates from all sources.
// Higher priority sources override lower priority sources.
func GetAllDelegates(agentDelegates map[string]string, globalCfg config.Config) map[string]string {
	result := make(map[string]string)

	// Start with global (lowest priority)
	if globalCfg.Delegates != nil {
		for k, v := range globalCfg.Delegates {
			result[k] = normalizeHandle(v)
		}
	}

	// Override with agent config
	if agentDelegates != nil {
		for k, v := range agentDelegates {
			result[k] = normalizeHandle(v)
		}
	}

	// Override with directory config (highest priority)
	dirConfig, _ := LoadDirectoryConfig()
	if dirConfig != nil && dirConfig.Delegates != nil {
		for k, v := range dirConfig.Delegates {
			result[k] = normalizeHandle(v)
		}
	}

	return result
}

// normalizeHandle ensures agent handles start with @.
func normalizeHandle(handle string) string {
	if handle == "" {
		return ""
	}
	if !strings.HasPrefix(handle, "@") {
		return "@" + handle
	}
	return handle
}

// SaveDirectoryConfig writes a .ayo.json file to the specified directory.
func SaveDirectoryConfig(dir string, cfg *DirectoryConfig) error {
	configPath := filepath.Join(dir, ".ayo.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0o644)
}
