package builtin

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// SkillInfo contains summary information about a builtin skill.
type SkillInfo struct {
	Name        string
	Description string
	Path        string // Relative path in embedded FS
}

// ListBuiltinSkills returns all shared built-in skill names.
func ListBuiltinSkills() []string {
	entries, err := skillsFS.ReadDir("skills")
	if err != nil {
		return nil
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names
}

// HasBuiltinSkill checks if a shared built-in skill exists.
func HasBuiltinSkill(name string) bool {
	skillsPath := path.Join("skills", name, "SKILL.md")
	_, err := skillsFS.ReadFile(skillsPath)
	return err == nil
}

// LoadBuiltinSkill loads a shared built-in skill definition.
func LoadBuiltinSkill(name string) (SkillDefinition, error) {
	skillsPath := path.Join("skills", name)
	return loadSkillFromFS(skillsFS, skillsPath, name)
}

// ListAgentBuiltinSkills returns skill names for a specific built-in agent.
// Deprecated: Agents are now standalone projects created with `ayo fresh`
func ListAgentBuiltinSkills(handle string) []string {
	return nil // No built-in agents in build system
}

// LoadAgentBuiltinSkill loads an agent-specific built-in skill.
// Deprecated: Agents are now standalone projects created with `ayo fresh`
func LoadAgentBuiltinSkill(handle, skillName string) (SkillDefinition, error) {
	return SkillDefinition{}, fmt.Errorf("agent-specific built-in skills are no longer supported. Use 'ayo fresh' to create agents with custom skills")
}

// loadSkillFromFS loads a skill definition from an embedded filesystem.
func loadSkillFromFS(fsys embed.FS, skillPath, skillName string) (SkillDefinition, error) {
	var skill SkillDefinition

	skillsMDPath := path.Join(skillPath, "SKILL.md")
	data, err := fsys.ReadFile(skillsMDPath)
	if err != nil {
		return skill, fmt.Errorf("skill %s not found: %w", skillName, err)
	}

	content := string(data)

	// Parse YAML frontmatter
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
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

	if skill.Name == "" {
		skill.Name = skillName
	}

	return skill, nil
}

// SkillsFS returns the embedded filesystem for shared built-in skill.
func SkillsFS() fs.FS {
	sub, _ := fs.Sub(skillsFS, "skills")
	return sub
}

// GetAllBuiltinSkillInfos returns info about all built-in skills (shared + agent-specific).
func GetAllBuiltinSkillInfos() []SkillInfo {
	var infos []SkillInfo

	// Shared built-in skills
	for _, name := range ListBuiltinSkills() {
		skills, err := LoadBuiltinSkill(name)
		if err != nil {
			continue
		}
		infos = append(infos, SkillInfo{
			Name:        skills.Name,
			Description: skills.Description,
			Path:        path.Join("skills", name),
		})
	}

	return infos
}