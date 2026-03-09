// Package capabilities provides OpenClaw skill integration for Ayo build system.
//
// This provider enables OpenClaw skills (SKILL.md format) to be used natively
// in Ayo projects, allowing seamless integration with the OpenClaw ecosystem.
package capabilities

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// OpenClawSkill represents an OpenClaw skill with metadata and content.
type OpenClawSkill struct {
	// Name is the skill name (from frontmatter or filename).
	Name string `json:"name" yaml:"name" toml:"name"`

	// Description explains what the skill does.
	Description string `json:"description" yaml:"description" toml:"description"`

	// Version of the skill (optional).
	Version string `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`

	// Author of the skill (optional).
	Author string `json:"author,omitempty" yaml:"author,omitempty" toml:"author,omitempty"`

	// Tags for categorization (optional).
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty" toml:"tags,omitempty"`

	// Content is the main skill content (Markdown after frontmatter).
	Content string `json:"content" yaml:"content" toml:"content"`

	// SourcePath is the path to the SKILL.md file.
	SourcePath string `json:"source_path" yaml:"source_path" toml:"source_path"`

	// OpenClawSpecific metadata
	OpenClawSpecific OpenClawMetadata `json:"openclaw,omitempty" yaml:"openclaw,omitempty" toml:"openclaw,omitempty"`
}

// OpenClawMetadata contains OpenClaw-specific skill metadata.
type OpenClawMetadata struct {
	// ComponentType indicates the type of OpenClaw component.
	ComponentType string `json:"component_type,omitempty" yaml:"component_type,omitempty" toml:"component_type,omitempty"`

	// Dependencies lists other OpenClaw components this skill depends on.
	Dependencies []string `json:"dependencies,omitempty" yaml:"dependencies,omitempty" toml:"dependencies,omitempty"`

	// Compatibility version requirements.
	Compatibility string `json:"compatibility,omitempty" yaml:"compatibility,omitempty" toml:"compatibility,omitempty"`

	// License for the skill.
	License string `json:"license,omitempty" yaml:"license,omitempty" toml:"license,omitempty"`
}

// OpenClawSkillProvider implements skill loading for OpenClaw SKILL.md format.
type OpenClawSkillProvider struct {
	// BasePaths are directories to search for SKILL.md files.
	BasePaths []string

	// FS allows using embedded filesystems (for testing).
	FS fs.FS
}

// NewOpenClawSkillProvider creates a new OpenClaw skill provider.
func NewOpenClawSkillProvider(basePaths ...string) *OpenClawSkillProvider {
	return &OpenClawSkillProvider{
		BasePaths: basePaths,
	}
}

// DiscoverSkills finds all SKILL.md files in the configured base paths.
func (p *OpenClawSkillProvider) DiscoverSkills() ([]OpenClawSkill, error) {
	var allSkills []OpenClawSkill

	for _, basePath := range p.BasePaths {
		skills, err := p.discoverSkillsInPath(basePath)
		if err != nil {
			return nil, fmt.Errorf("error discovering skills in %s: %w", basePath, err)
		}
		allSkills = append(allSkills, skills...)
	}

	return allSkills, nil
}

// discoverSkillsInPath finds SKILL.md files in a specific directory.
func (p *OpenClawSkillProvider) discoverSkillsInPath(basePath string) ([]OpenClawSkill, error) {
	var skills []OpenClawSkill

	err := filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Base(path) == "SKILL.md" {
			skill, err := p.LoadSkillFromPath(path)
			if err != nil {
				return fmt.Errorf("error loading skill from %s: %w", path, err)
			}
			skills = append(skills, skill)
		}

		return nil
	})

	return skills, err
}

// LoadSkillFromPath loads an OpenClaw skill from a SKILL.md file path.
func (p *OpenClawSkillProvider) LoadSkillFromPath(skillPath string) (OpenClawSkill, error) {
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return OpenClawSkill{}, fmt.Errorf("failed to read SKILL.md file: %w", err)
	}

	return p.ParseSkillContent(string(content), skillPath)
}

// ParseSkillContent parses OpenClaw skill content with YAML frontmatter.
func (p *OpenClawSkillProvider) ParseSkillContent(content, sourcePath string) (OpenClawSkill, error) {
	skill := OpenClawSkill{
		SourcePath: sourcePath,
	}

	// Split content into frontmatter and body
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		// No frontmatter, treat entire content as skill content
		skill.Content = strings.TrimSpace(content)
		skill.Name = strings.TrimSuffix(filepath.Base(sourcePath), ".md")
		return skill, nil
	}

	// Parse YAML frontmatter
	frontmatter := parts[1]
	err := yaml.Unmarshal([]byte(frontmatter), &skill)
	if err != nil {
		return OpenClawSkill{}, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Set content from the body
	skill.Content = strings.TrimSpace(parts[2])

	// Default name from filename if not specified
	if skill.Name == "" {
		skill.Name = strings.TrimSuffix(filepath.Base(sourcePath), ".md")
	}

	return skill, nil
}

// ConvertToSkillDefinition converts an OpenClawSkill to Ayo's SkillDefinition format.
func (s *OpenClawSkill) ConvertToSkillDefinition() SkillDefinition {
	return SkillDefinition{
		Name:        s.Name,
		Description: s.Description,
		Content:     s.Content,
		Metadata: map[string]interface{}{
			"openclaw": s.OpenClawSpecific,
			"source":   s.SourcePath,
			"version":  s.Version,
			"author":   s.Author,
		},
	}
}

// SkillDefinition represents a skill in Ayo's format.
// This is duplicated here to avoid circular imports - the real definition
// is in internal/builtin/builtin.go
type SkillDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LoadOpenClawSkillsFromProject loads OpenClaw skills from a project directory.
func LoadOpenClawSkillsFromProject(projectDir string) ([]SkillDefinition, error) {
	provider := NewOpenClawSkillProvider(
		filepath.Join(projectDir, "skills"),
		filepath.Join(projectDir, "openclaw"),
	)

	openClawSkills, err := provider.DiscoverSkills()
	if err != nil {
		return nil, fmt.Errorf("failed to discover OpenClaw skills: %w", err)
	}

	var skillDefs []SkillDefinition
	for _, skill := range openClawSkills {
		skillDefs = append(skillDefs, skill.ConvertToSkillDefinition())
	}

	return skillDefs, nil
}

// LoadOpenClawSkillsFromTOML loads OpenClaw skills from a config.toml file.
func LoadOpenClawSkillsFromTOML(configPath string) ([]SkillDefinition, error) {
	// Read the TOML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config.toml: %w", err)
	}

	var config struct {
		OpenClaw struct {
			Skills []struct {
				Name        string                 `toml:"name"`
				Description string                 `toml:"description"`
				Content     string                 `toml:"content"`
				Metadata    map[string]interface{} `toml:"metadata"`
			} `toml:"skills"`
		} `toml:"openclaw"`
	}

	err = toml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config.toml: %w", err)
	}

	var skillDefs []SkillDefinition
	for _, skill := range config.OpenClaw.Skills {
		skillDefs = append(skillDefs, SkillDefinition{
			Name:        skill.Name,
			Description: skill.Description,
			Content:     skill.Content,
			Metadata:    skill.Metadata,
		})
	}

	return skillDefs, nil
}