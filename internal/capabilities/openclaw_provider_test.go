package capabilities

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenClawSkillProvider(t *testing.T) {
	t.Run("ParseSkillContent with YAML frontmatter", func(t *testing.T) {
		provider := &OpenClawSkillProvider{}
		
		content := `---
name: test-skill
description: A test OpenClaw skill
author: Test Author
version: 1.0.0
tags:
  - test
  - example
openclaw:
  component_type: skill
  dependencies:
    - base-skill
  compatibility: ">=1.0.0"
  license: MIT
---
# Test Skill Content

This is the main content of the test skill.
It can contain **markdown** formatting.
`

		skill, err := provider.ParseSkillContent(content, "test-skill.md")
		require.NoError(t, err)
		
		assert.Equal(t, "test-skill", skill.Name)
		assert.Equal(t, "A test OpenClaw skill", skill.Description)
		assert.Equal(t, "Test Author", skill.Author)
		assert.Equal(t, "1.0.0", skill.Version)
		assert.Equal(t, []string{"test", "example"}, skill.Tags)
		assert.Equal(t, "skill", skill.OpenClawSpecific.ComponentType)
		assert.Equal(t, []string{"base-skill"}, skill.OpenClawSpecific.Dependencies)
		assert.Equal(t, ">=1.0.0", skill.OpenClawSpecific.Compatibility)
		assert.Equal(t, "MIT", skill.OpenClawSpecific.License)
		assert.Contains(t, skill.Content, "Test Skill Content")
		assert.Contains(t, skill.Content, "markdown")
	})

	t.Run("ParseSkillContent without frontmatter", func(t *testing.T) {
		provider := &OpenClawSkillProvider{}
		
		content := "# Simple Skill\n\nThis skill has no frontmatter."
		
		skill, err := provider.ParseSkillContent(content, "simple-skill.md")
		require.NoError(t, err)
		
		assert.Equal(t, "simple-skill", skill.Name)
		assert.Equal(t, "# Simple Skill\n\nThis skill has no frontmatter.", skill.Content)
	})

	t.Run("ConvertToSkillDefinition", func(t *testing.T) {
		skill := OpenClawSkill{
			Name:        "test-skill",
			Description: "Test description",
			Content:     "Test content",
			Version:     "1.0.0",
			Author:      "Test Author",
			SourcePath:  "/path/to/skill.md",
			OpenClawSpecific: OpenClawMetadata{
				ComponentType: "skill",
				Dependencies:  []string{"dep1"},
			},
		}
		
		skillDef := skill.ConvertToSkillDefinition()
		
		assert.Equal(t, "test-skill", skillDef.Name)
		assert.Equal(t, "Test description", skillDef.Description)
		assert.Equal(t, "Test content", skillDef.Content)
		assert.NotNil(t, skillDef.Metadata)
		assert.Equal(t, "skill", skillDef.Metadata["openclaw"].(OpenClawMetadata).ComponentType)
	})

	t.Run("LoadSkillFromPath with real file", func(t *testing.T) {
		// Create a temporary test file
		tempDir := t.TempDir()
		skillPath := filepath.Join(tempDir, "test-skill.md")
		
		content := `---
name: test-skill
description: Test skill for file loading
author: Test
---
# Test Skill

File-based skill content.
`
		
		err := os.WriteFile(skillPath, []byte(content), 0644)
		require.NoError(t, err)
		
		provider := &OpenClawSkillProvider{}
		skill, err := provider.LoadSkillFromPath(skillPath)
		require.NoError(t, err)
		
		assert.Equal(t, "test-skill", skill.Name)
		assert.Equal(t, "Test skill for file loading", skill.Description)
		assert.Contains(t, skill.Content, "File-based skill content")
	})

	t.Run("DiscoverSkills in directory", func(t *testing.T) {
		// Create temporary directory with multiple skills
		tempDir := t.TempDir()
		
		// Create subdirectory with skill
		subDir := filepath.Join(tempDir, "subdir")
		err := os.MkdirAll(subDir, 0755)
		require.NoError(t, err)
		
		// Create first skill
		skill1Content := `---
name: skill-one
description: First test skill
---
Content of skill one
`
		skill1Path := filepath.Join(tempDir, "SKILL.md")
		err = os.WriteFile(skill1Path, []byte(skill1Content), 0644)
		require.NoError(t, err)
		
		// Create second skill in subdirectory
		skill2Content := `---
name: skill-two
description: Second test skill
---
Content of skill two
`
		skill2Path := filepath.Join(subDir, "SKILL.md")
		err = os.WriteFile(skill2Path, []byte(skill2Content), 0644)
		require.NoError(t, err)
		
		// Create non-skill file (should be ignored)
		otherFile := filepath.Join(tempDir, "readme.txt")
		err = os.WriteFile(otherFile, []byte("Not a skill"), 0644)
		require.NoError(t, err)
		
		provider := NewOpenClawSkillProvider(tempDir)
		skills, err := provider.DiscoverSkills()
		require.NoError(t, err)
		
		// Should find both skill files
		assert.Len(t, skills, 2)
		
		// Check that we found the right skills
		foundNames := make(map[string]bool)
		for _, skill := range skills {
			foundNames[skill.Name] = true
		}
		
		assert.True(t, foundNames["skill-one"])
		assert.True(t, foundNames["skill-two"])
	})

	t.Run("LoadOpenClawSkillsFromProject", func(t *testing.T) {
		// Create temporary project structure
		tempDir := t.TempDir()
		
		// Create skills directory
		skillsDir := filepath.Join(tempDir, "skills")
		err := os.MkdirAll(skillsDir, 0755)
		require.NoError(t, err)
		
		// Create openclaw directory
		openclawDir := filepath.Join(tempDir, "openclaw")
		err = os.MkdirAll(openclawDir, 0755)
		require.NoError(t, err)
		
		// Create skill in skills directory
		skill1Content := `---
name: project-skill
description: Project-specific skill
---
Project skill content
`
		skill1Path := filepath.Join(skillsDir, "SKILL.md")
		err = os.WriteFile(skill1Path, []byte(skill1Content), 0644)
		require.NoError(t, err)
		
		// Create skill in openclaw directory
		skill2Content := `---
name: openclaw-skill
description: OpenClaw-specific skill
openclaw:
  component_type: skill
---
OpenClaw skill content
`
		skill2Path := filepath.Join(openclawDir, "SKILL.md")
		err = os.WriteFile(skill2Path, []byte(skill2Content), 0644)
		require.NoError(t, err)
		
		// Load skills from project
		skillDefs, err := LoadOpenClawSkillsFromProject(tempDir)
		require.NoError(t, err)
		
		// Should find both skills
		assert.Len(t, skillDefs, 2)
		
		// Check skill names
		foundNames := make(map[string]bool)
		for _, skillDef := range skillDefs {
			foundNames[skillDef.Name] = true
		}
		
		assert.True(t, foundNames["project-skill"])
		assert.True(t, foundNames["openclaw-skill"])
	})

	t.Run("LoadOpenClawSkillsFromTOML", func(t *testing.T) {
		// Create temporary TOML file
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.toml")
		
		configContent := `[openclaw]
  [[openclaw.skills]]
    name = "toml-skill"
    description = "Skill defined in TOML"
    content = "TOML-based skill content"
    
    [openclaw.skills.metadata]
    version = "1.0.0"
    author = "TOML Author"
`
		
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)
		
		// Load skills from TOML
		skillDefs, err := LoadOpenClawSkillsFromTOML(configPath)
		require.NoError(t, err)
		
		// Should find one skill
		assert.Len(t, skillDefs, 1)
		assert.Equal(t, "toml-skill", skillDefs[0].Name)
		assert.Equal(t, "Skill defined in TOML", skillDefs[0].Description)
		assert.Equal(t, "TOML-based skill content", skillDefs[0].Content)
		assert.Equal(t, "1.0.0", skillDefs[0].Metadata["version"])
		assert.Equal(t, "TOML Author", skillDefs[0].Metadata["author"])
	})
}