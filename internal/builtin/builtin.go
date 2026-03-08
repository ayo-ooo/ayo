// Package builtin provides embedded built-in agents that ship with ayo.
package builtin

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

//go:embed prompts/*
var promptsFS embed.FS

//go:embed skills/*
var skillsFS embed.FS

// ConfigSchema is the embedded JSON schema for ayo configuration.
// Loaded from the root ayo-config-schema.json via a separate embed directive.
var ConfigSchema []byte

// AgentDefinition represents a built-in agent's configuration and content
type AgentDefinition struct {
	Handle      string
	Config      AgentConfig
	System      string
	Skills      []SkillDefinition
	Description string
}

// AgentConfig mirrors the user agent config structure
type AgentConfig struct {
	Model        string   `json:"model,omitempty"`
	Description  string   `json:"description,omitempty"`
	DelegateHint string   `json:"delegate_hint,omitempty"`
	AllowedTools []string `json:"allowed_tools,omitempty"`
}

// SkillDefinition represents a built-in skill
type SkillDefinition struct {
	Name        string
	Description string
	Content     string
}

// ListAgents returns all built-in agent handles
// Deprecated: Agents are now standalone projects created with `ayo fresh`
func ListAgents() []string {
	return []string{} // No built-in agents in build system
}

// HasAgent checks if a built-in agent exists with the given handle
// Deprecated: Agents are now standalone projects created with `ayo fresh`
func HasAgent(handle string) bool {
	return false // No built-in agents in build system
}

// LoadAgent loads a built-in agent definition
// Deprecated: Agents are now standalone projects created with `ayo fresh`
func LoadAgent(handle string) (AgentDefinition, error) {
	return AgentDefinition{}, fmt.Errorf("builtin agents are no longer supported in the build system. Use 'ayo fresh' to create new agents")
}

func loadSkill(skillPath string) (SkillDefinition, error) {
	var skill SkillDefinition

	// Read SKILL.md for metadata
	skillMD, err := skillsFS.ReadFile(path.Join(skillPath, "SKILL.md"))
	if err != nil {
		return skill, err
	}

	content := string(skillMD)

	// Parse YAML frontmatter if present
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			// Parse frontmatter
			frontmatter := parts[1]
			for _, line := range strings.Split(frontmatter, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "name:") {
					skill.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				} else if strings.HasPrefix(line, "description:") {
					skill.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				}
			}
			skill.Content = strings.TrimSpace(parts[2])
		}
	} else {
		skill.Content = content
	}

	// Default name from directory
	if skill.Name == "" {
		skill.Name = path.Base(skillPath)
	}

	return skill, nil
}

// FS returns the embedded filesystem for built-in skills.
// Agents are no longer embedded - use 'ayo fresh' to create new agents
func FS() fs.FS {
	sub, _ := fs.Sub(skillsFS, "skills")
	return sub
}

// AgentInfo contains summary information about a builtin agent for delegation hints
type AgentInfo struct {
	Handle       string
	Description  string
	DelegateHint string
}

// ListAgentInfo returns info about all builtin agents for use in system prompts
func ListAgentInfo() []AgentInfo {
	handles := ListAgents()
	infos := make([]AgentInfo, 0, len(handles))
	for _, handle := range handles {
		def, err := LoadAgent(handle)
		if err != nil {
			continue
		}
		infos = append(infos, AgentInfo{
			Handle:       handle,
			Description:  def.Config.Description,
			DelegateHint: def.Config.DelegateHint,
		})
	}
	return infos
}
