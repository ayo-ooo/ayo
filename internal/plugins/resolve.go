package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexcabrera/ayo/internal/paths"
)

// ConflictType indicates the type of naming conflict.
type ConflictType int

const (
	ConflictAgent ConflictType = iota
	ConflictSkill
	ConflictTool
)

func (ct ConflictType) String() string {
	switch ct {
	case ConflictAgent:
		return "agent"
	case ConflictSkill:
		return "skill"
	case ConflictTool:
		return "tool"
	default:
		return "unknown"
	}
}

// Conflict represents a naming conflict during installation.
type Conflict struct {
	Type         ConflictType
	Name         string // The conflicting name (e.g., "@crush")
	ExistingPath string // Path to existing entity
	NewPath      string // Path to new entity from plugin
	Source       string // Where the existing entity comes from (e.g., "builtin", "user", "plugin:foo")
}

// ConflictResolution describes how to resolve a conflict.
type ConflictResolution struct {
	Conflict  Conflict
	Action    ResolutionAction
	RenameTo  string // New name if action is Rename
}

// ResolutionAction is the action to take for a conflict.
type ResolutionAction int

const (
	// Skip leaves the existing entity and doesn't install the new one.
	ResolutionSkip ResolutionAction = iota
	// Replace overwrites the existing entity with the new one.
	ResolutionReplace
	// Rename installs the new entity under a different name.
	ResolutionRename
)

func (ra ResolutionAction) String() string {
	switch ra {
	case ResolutionSkip:
		return "skip"
	case ResolutionReplace:
		return "replace"
	case ResolutionRename:
		return "rename"
	default:
		return "unknown"
	}
}

// DetectConflicts checks for naming conflicts between a plugin and existing entities.
func DetectConflicts(manifest *Manifest, pluginDir string) ([]Conflict, error) {
	var conflicts []Conflict

	// Check agent conflicts
	for _, agent := range manifest.Agents {
		if conflict := detectAgentConflict(agent, pluginDir); conflict != nil {
			conflicts = append(conflicts, *conflict)
		}
	}

	// Check skill conflicts
	for _, skill := range manifest.Skills {
		if conflict := detectSkillConflict(skill, pluginDir); conflict != nil {
			conflicts = append(conflicts, *conflict)
		}
	}

	// Check tool conflicts
	for _, tool := range manifest.Tools {
		if conflict := detectToolConflict(tool, pluginDir); conflict != nil {
			conflicts = append(conflicts, *conflict)
		}
	}

	return conflicts, nil
}

// detectAgentConflict checks if an agent handle already exists.
func detectAgentConflict(handle string, pluginDir string) *Conflict {
	// Normalize handle
	if !strings.HasPrefix(handle, "@") {
		handle = "@" + handle
	}

	newPath := filepath.Join(pluginDir, "agents", handle)

	// Check built-in agents
	builtinDir := paths.BuiltinAgentsDir()
	builtinPath := filepath.Join(builtinDir, handle)
	if _, err := os.Stat(builtinPath); err == nil {
		return &Conflict{
			Type:         ConflictAgent,
			Name:         handle,
			ExistingPath: builtinPath,
			NewPath:      newPath,
			Source:       "builtin",
		}
	}

	// Check user agents
	userDir := paths.AgentsDir()
	userPath := filepath.Join(userDir, handle)
	if _, err := os.Stat(userPath); err == nil {
		return &Conflict{
			Type:         ConflictAgent,
			Name:         handle,
			ExistingPath: userPath,
			NewPath:      newPath,
			Source:       "user",
		}
	}

	// Check other plugins
	registry, err := LoadRegistry()
	if err != nil {
		return nil
	}

	for _, plugin := range registry.ListEnabled() {
		for _, existingAgent := range plugin.Agents {
			resolvedHandle := plugin.GetResolvedAgentHandle(existingAgent)
			if resolvedHandle == handle {
				return &Conflict{
					Type:         ConflictAgent,
					Name:         handle,
					ExistingPath: filepath.Join(plugin.Path, "agents", existingAgent),
					NewPath:      newPath,
					Source:       fmt.Sprintf("plugin:%s", plugin.Name),
				}
			}
		}
	}

	return nil
}

// detectSkillConflict checks if a skill name already exists.
func detectSkillConflict(name string, pluginDir string) *Conflict {
	newPath := filepath.Join(pluginDir, "skills", name)

	// Check built-in skills
	builtinDir := paths.BuiltinSkillsDir()
	builtinPath := filepath.Join(builtinDir, name)
	if _, err := os.Stat(builtinPath); err == nil {
		return &Conflict{
			Type:         ConflictSkill,
			Name:         name,
			ExistingPath: builtinPath,
			NewPath:      newPath,
			Source:       "builtin",
		}
	}

	// Check user skills
	userDir := paths.SkillsDir()
	userPath := filepath.Join(userDir, name)
	if _, err := os.Stat(userPath); err == nil {
		return &Conflict{
			Type:         ConflictSkill,
			Name:         name,
			ExistingPath: userPath,
			NewPath:      newPath,
			Source:       "user",
		}
	}

	// Check other plugins
	registry, err := LoadRegistry()
	if err != nil {
		return nil
	}

	for _, plugin := range registry.ListEnabled() {
		for _, existingSkill := range plugin.Skills {
			if existingSkill == name {
				return &Conflict{
					Type:         ConflictSkill,
					Name:         name,
					ExistingPath: filepath.Join(plugin.Path, "skills", existingSkill),
					NewPath:      newPath,
					Source:       fmt.Sprintf("plugin:%s", plugin.Name),
				}
			}
		}
	}

	return nil
}

// detectToolConflict checks if a tool name already exists.
func detectToolConflict(name string, pluginDir string) *Conflict {
	newPath := filepath.Join(pluginDir, "tools", name)

	// Built-in tools are hardcoded, check against known list
	builtinTools := []string{"bash", "plan", "agent_call"}
	for _, bt := range builtinTools {
		if bt == name {
			return &Conflict{
				Type:         ConflictTool,
				Name:         name,
				ExistingPath: "(builtin)",
				NewPath:      newPath,
				Source:       "builtin",
			}
		}
	}

	// Check other plugins
	registry, err := LoadRegistry()
	if err != nil {
		return nil
	}

	for _, plugin := range registry.ListEnabled() {
		for _, existingTool := range plugin.Tools {
			if existingTool == name {
				return &Conflict{
					Type:         ConflictTool,
					Name:         name,
					ExistingPath: filepath.Join(plugin.Path, "tools", existingTool),
					NewPath:      newPath,
					Source:       fmt.Sprintf("plugin:%s", plugin.Name),
				}
			}
		}
	}

	return nil
}

// ValidateAgentRename checks if a rename is valid.
func ValidateAgentRename(oldHandle, newHandle string) error {
	if newHandle == "" {
		return fmt.Errorf("new handle cannot be empty")
	}

	// Normalize handles
	if !strings.HasPrefix(newHandle, "@") {
		newHandle = "@" + newHandle
	}

	// Check reserved namespace
	if strings.HasPrefix(newHandle, "@ayo") {
		return fmt.Errorf("cannot use reserved 'ayo' namespace")
	}

	// Check if new handle already exists
	if conflict := detectAgentConflict(newHandle, ""); conflict != nil {
		return fmt.Errorf("handle %s already exists (%s)", newHandle, conflict.Source)
	}

	return nil
}

// SuggestRename generates a suggested rename for a conflicting entity.
func SuggestRename(name string, conflictType ConflictType) string {
	switch conflictType {
	case ConflictAgent:
		// For agents, suggest adding a suffix
		if strings.HasPrefix(name, "@") {
			return name + "-plugin"
		}
		return "@" + name + "-plugin"
	case ConflictSkill:
		return name + "-plugin"
	case ConflictTool:
		return name + "-ext"
	default:
		return name + "-alt"
	}
}
