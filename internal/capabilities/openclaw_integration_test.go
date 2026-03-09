package capabilities

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexcabrera/ayo/internal/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenClawIntegration(t *testing.T) {
	t.Run("End-to-end OpenClaw skill discovery and loading", func(t *testing.T) {
		// Create a temporary project with OpenClaw skills
		tempDir := t.TempDir()
		
		// Create skills directory
		skillsDir := filepath.Join(tempDir, "skills")
		err := os.MkdirAll(skillsDir, 0755)
		require.NoError(t, err)
		
		// Create first OpenClaw skill
		skill1Content := `---
name: integration-skill-1
description: First integration test skill
author: Integration Test
tags:
  - test
  - integration
openclaw:
  component_type: skill
  dependencies:
    - base-skill
  compatibility: ">=1.0.0"
---
# Integration Skill 1

This is the first skill for integration testing.
It demonstrates OpenClaw skill format compatibility.
`
		skill1Path := filepath.Join(skillsDir, "integration-skill-1.md")
		err = os.WriteFile(skill1Path, []byte(skill1Content), 0644)
		require.NoError(t, err)
		
		// Create second OpenClaw skill
		skill2Content := `---
name: integration-skill-2
description: Second integration test skill
version: 2.0.0
---
# Integration Skill 2

This is the second skill for integration testing.
`
		skill2Path := filepath.Join(skillsDir, "integration-skill-2.md")
		err = os.WriteFile(skill2Path, []byte(skill2Content), 0644)
		require.NoError(t, err)
		
		// Test discovery service
		discoveryService := NewOpenClawDiscoveryService()
		discoveryResults, err := discoveryService.DiscoverFromLocalDirectory(skillsDir)
		require.NoError(t, err)
		
		assert.Len(t, discoveryResults, 1)
		assert.Equal(t, "local", discoveryResults[0].Source)
		assert.Equal(t, "skills", discoveryResults[0].Name)
		assert.Equal(t, 2, discoveryResults[0].ComponentCount)
		
		// Test skill provider
		provider := NewOpenClawSkillProvider(skillsDir)
		openClawSkills, err := provider.DiscoverSkills()
		require.NoError(t, err)
		
		assert.Len(t, openClawSkills, 2)
		
		// Verify first skill
		foundSkill1 := false
		foundSkill2 := false
		for _, skill := range openClawSkills {
			if skill.Name == "integration-skill-1" {
				foundSkill1 = true
				assert.Equal(t, "First integration test skill", skill.Description)
				assert.Equal(t, "Integration Test", skill.Author)
				assert.Equal(t, []string{"test", "integration"}, skill.Tags)
				assert.Equal(t, "skill", skill.OpenClawSpecific.ComponentType)
				assert.Equal(t, []string{"base-skill"}, skill.OpenClawSpecific.Dependencies)
				assert.Contains(t, skill.Content, "Integration Skill 1")
			}
			if skill.Name == "integration-skill-2" {
				foundSkill2 = true
				assert.Equal(t, "Second integration test skill", skill.Description)
				assert.Equal(t, "2.0.0", skill.Version)
				assert.Contains(t, skill.Content, "Integration Skill 2")
			}
		}
		
		assert.True(t, foundSkill1, "integration-skill-1 should be found")
		assert.True(t, foundSkill2, "integration-skill-2 should be found")
		
		// Test conversion to Ayo skill definitions
		skillDefs, err := LoadOpenClawSkillsFromProject(tempDir)
		require.NoError(t, err)
		
		assert.Len(t, skillDefs, 2)
		
		// Verify skill definitions have proper metadata
		for _, skillDef := range skillDefs {
			assert.NotEmpty(t, skillDef.Name)
			assert.NotEmpty(t, skillDef.Description)
			assert.NotEmpty(t, skillDef.Content)
			assert.NotNil(t, skillDef.Metadata)
			assert.Contains(t, skillDef.Metadata, "openclaw")
		}
	})

	t.Run("OpenClaw plugin adaptation integration", func(t *testing.T) {
		// Create a temporary OpenClaw plugin directory
		tempDir := t.TempDir()
		
		// Create OpenClaw manifest
		manifestContent := `{
  "name": "integration-plugin",
  "version": "1.0.0",
  "description": "Integration test plugin",
  "author": "Integration Test",
  "repository": "https://github.com/test/integration-plugin",
  "license": "MIT",
  "components": [
    {
      "name": "plugin-skill",
      "type": "skill",
      "description": "A skill from the plugin"
    },
    {
      "name": "plugin-tool",
      "type": "tool",
      "description": "A tool from the plugin"
    },
    {
      "name": "memory-provider",
      "type": "memory",
      "description": "A memory provider",
      "entry_point": "memory.so"
    }
  ],
  "dependencies": [
    {
      "name": "base-plugin",
      "version": ">=1.0.0",
      "type": "required"
    }
  ],
  "compatibility": ">=1.0.0",
  "ayo_version": ">=0.1.0"
}`
		
		manifestPath := filepath.Join(tempDir, "manifest.json")
		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)
		
		// Test plugin adaptation
		adapter := plugins.NewOpenClawPluginAdapter(tempDir, manifestPath)
		ayoManifest, err := adapter.Adapt()
		require.NoError(t, err)
		
		// Verify adapted manifest
		assert.Equal(t, "openclaw-integration-plugin", ayoManifest.Name)
		assert.Equal(t, "1.0.0", ayoManifest.Version)
		assert.Equal(t, "OpenClaw plugin: Integration test plugin", ayoManifest.Description)
		assert.Equal(t, "Integration Test", ayoManifest.Author)
		assert.Equal(t, "https://github.com/test/integration-plugin", ayoManifest.Repository)
		assert.Equal(t, "MIT", ayoManifest.License)
		assert.Equal(t, ">=0.1.0", ayoManifest.AyoVersion)
		
		// Verify components were converted
		assert.Contains(t, ayoManifest.Skills, "plugin-skill")
		assert.Contains(t, ayoManifest.Tools, "plugin-tool")
		assert.Len(t, ayoManifest.Providers, 1)
		assert.Equal(t, "memory-provider", ayoManifest.Providers[0].Name)
		assert.Equal(t, plugins.PluginTypeMemory, ayoManifest.Providers[0].Type)
		
		// Verify dependencies were converted
		require.NotNil(t, ayoManifest.Dependencies)
		assert.Len(t, ayoManifest.Dependencies.Plugins, 1)
		assert.Equal(t, "openclaw-base-plugin@>=1.0.0", ayoManifest.Dependencies.Plugins[0])
	})

	t.Run("OpenClaw project structure integration", func(t *testing.T) {
		// Create a complex OpenClaw project structure
		tempDir := t.TempDir()
		
		// Create main project directories
		skillsDir := filepath.Join(tempDir, "skills")
		openclawDir := filepath.Join(tempDir, "openclaw")
		agentsDir := filepath.Join(tempDir, "agents")
		
		err = os.MkdirAll(skillsDir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(openclawDir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(agentsDir, 0755)
		require.NoError(t, err)
		
		// Create skills in different directories
		skill1Content := `---
name: main-skill
description: Main project skill
---
# Main Skill

This is the main skill.
`
		skill1Path := filepath.Join(skillsDir, "main-skill.md")
		err = os.WriteFile(skill1Path, []byte(skill1Content), 0644)
		require.NoError(t, err)
		
		skill2Content := `---
name: openclaw-skill
description: OpenClaw-specific skill
openclaw:
  component_type: skill
  compatibility: ">=1.0.0"
---
# OpenClaw Skill

This is an OpenClaw-specific skill.
`
		skill2Path := filepath.Join(openclawDir, "openclaw-skill.md")
		err = os.WriteFile(skill2Path, []byte(skill2Content), 0644)
		require.NoError(t, err)
		
		// Test discovery from multiple directories
		provider := NewOpenClawSkillProvider(
			filepath.Join(tempDir, "skills"),
			filepath.Join(tempDir, "openclaw"),
		)
		
		skills, err := provider.DiscoverSkills()
		require.NoError(t, err)
		
		assert.Len(t, skills, 2)
		
		// Verify both skills were found
		foundNames := make(map[string]bool)
		for _, skill := range skills {
			foundNames[skill.Name] = true
		}
		
		assert.True(t, foundNames["main-skill"])
		assert.True(t, foundNames["openclaw-skill"])
		
		// Test loading from project
		skillDefs, err := LoadOpenClawSkillsFromProject(tempDir)
		require.NoError(t, err)
		
		assert.Len(t, skillDefs, 2)
		
		// Verify skill definitions
		for _, skillDef := range skillDefs {
			assert.NotEmpty(t, skillDef.Name)
			assert.NotEmpty(t, skillDef.Description)
			assert.NotEmpty(t, skillDef.Content)
			
			// Check that metadata contains source information
			if skillDef.Metadata != nil {
				if source, ok := skillDef.Metadata["source"].(string); ok {
					assert.Contains(t, source, "openclaw")
				}
			}
		}
	})

	t.Run("OpenClaw and Ayo skill compatibility", func(t *testing.T) {
		// Create a project with both OpenClaw and Ayo-style skills
		tempDir := t.TempDir()
		
		// Create OpenClaw skill
		openclawSkillContent := `---
name: openclaw-skill
description: OpenClaw format skill
author: Test
tags:
  - openclaw
  - test
openclaw:
  component_type: skill
  compatibility: ">=1.0.0"
---
# OpenClaw Skill

This skill uses OpenClaw format.
`
		openclawSkillPath := filepath.Join(tempDir, "openclaw-skill.md")
		err = os.WriteFile(openclawSkillPath, []byte(openclawSkillContent), 0644)
		require.NoError(t, err)
		
		// Create Ayo-style skill (simple markdown without frontmatter)
		ayoSkillContent := "# Ayo Skill\n\nThis skill uses simple markdown format without OpenClaw frontmatter."
		ayoSkillPath := filepath.Join(tempDir, "ayo-skill.md")
		err = os.WriteFile(ayoSkillPath, []byte(ayoSkillContent), 0644)
		require.NoError(t, err)
		
		// Test that OpenClaw provider can handle both formats
		provider := NewOpenClawSkillProvider(tempDir)
		skills, err := provider.DiscoverSkills()
		require.NoError(t, err)
		
		// Should find both skills
		assert.Len(t, skills, 2)
		
		foundOpenClaw := false
		foundAyo := false
		
		for _, skill := range skills {
			if skill.Name == "openclaw-skill" {
				foundOpenClaw = true
				assert.Equal(t, "OpenClaw format skill", skill.Description)
				assert.Equal(t, "Test", skill.Author)
				assert.Equal(t, "skill", skill.OpenClawSpecific.ComponentType)
				assert.Contains(t, skill.Content, "OpenClaw Skill")
			}
			if skill.Name == "ayo-skill" {
				foundAyo = true
				// Ayo-style skill should have filename as name and full content preserved
				assert.Contains(t, skill.Content, "Ayo Skill")
				assert.Contains(t, skill.Content, "simple markdown format")
			}
		}
		
		assert.True(t, foundOpenClaw, "OpenClaw skill should be found")
		assert.True(t, foundAyo, "Ayo-style skill should be found")
		
		// Test conversion to skill definitions
		skillDefs, err := LoadOpenClawSkillsFromProject(tempDir)
		require.NoError(t, err)
		
		assert.Len(t, skillDefs, 2)
		
		// Both should convert successfully
		for _, skillDef := range skillDefs {
			assert.NotEmpty(t, skillDef.Name)
			assert.NotEmpty(t, skillDef.Content)
		}
	})

	t.Run("OpenClaw manifest formats integration", func(t *testing.T) {
		testCases := []struct {
			name     string
			filename string
			content  string
		}{
			{
				name:     "JSON",
				filename: "manifest.json",
				content:  `{"name":"json-plugin","version":"1.0.0","description":"JSON plugin","components":[{"name":"json-skill","type":"skill","description":"JSON skill"}]}`,
			},
			{
				name:     "YAML",
				filename: "manifest.yaml",
				content:  "name: yaml-plugin\nversion: 1.0.0\ndescription: YAML plugin\ncomponents:\n  - name: yaml-skill\n    type: skill\n    description: YAML skill",
			},
			{
				name:     "TOML",
				filename: "manifest.toml",
				content:  "name = 'toml-plugin'\nversion = '1.0.0'\ndescription = 'TOML plugin'\n[[components]]\nname = 'toml-skill'\ntype = 'skill'\ndescription = 'TOML skill'",
			},
		}
		
		for _, tc := range testCases {
			t.Run("Manifest format:"+tc.name, func(t *testing.T) {
				// Create temporary directory for this test case
				testDir := t.TempDir()
				
				// Create manifest file
				manifestPath := filepath.Join(testDir, tc.filename)
				err := os.WriteFile(manifestPath, []byte(tc.content), 0644)
				require.NoError(t, err)
				
				// Test plugin adaptation
				adapter := plugins.NewOpenClawPluginAdapter(testDir, manifestPath)
				ayoManifest, err := adapter.Adapt()
				require.NoError(t, err, "Failed to adapt %s manifest", tc.name)
				
				// Verify adaptation
				assert.NotEmpty(t, ayoManifest.Name)
				assert.NotEmpty(t, ayoManifest.Version)
				assert.NotEmpty(t, ayoManifest.Description)
				
				// Check that components were converted
				hasSkills := len(ayoManifest.Skills) > 0
				hasTools := len(ayoManifest.Tools) > 0
				hasProviders := len(ayoManifest.Providers) > 0
				
				assert.True(t, hasSkills || hasTools || hasProviders, "Should have at least one component type")
				
				// Verify name format indicates OpenClaw adaptation
				assert.True(t, strings.HasPrefix(ayoManifest.Name, "openclaw-"), "Name should start with 'openclaw-'")
			})
		}
	})

	t.Run("OpenClaw discovery service integration", func(t *testing.T) {
		// Create a temporary project structure
		tempDir := t.TempDir()
		
		// Create multiple projects in subdirectories
		for i := 1; i <= 3; i++ {
			projectDir := filepath.Join(tempDir, fmt.Sprintf("project-%d", i))
			err := os.MkdirAll(projectDir, 0755)
			require.NoError(t, err)
			
			// Create manifest
			manifestContent := fmt.Sprintf(`{"name":"project-%d","version":"1.0.0","components":[{"name":"skill-%d","type":"skill","description":"Skill %d"}]}`,
				i, i, i)
			manifestPath := filepath.Join(projectDir, "manifest.json")
			err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
			require.NoError(t, err)
			
			// Create skill file
			skillContent := fmt.Sprintf("---\nname: skill-%d\ndescription: Skill %d\n---\n# Skill %d\n\nThis is skill %d.", i, i, i, i)
			skillPath := filepath.Join(projectDir, "SKILL.md")
			err = os.WriteFile(skillPath, []byte(skillContent), 0644)
			require.NoError(t, err)
		}
		
		// Test discovery service
		service := NewOpenClawDiscoveryService()
		discoveryResults, err := service.DiscoverLocalOpenClawProjects([]string{tempDir})
		require.NoError(t, err)
		
		// Should find all 3 projects
		assert.Len(t, discoveryResults, 3)
		
		// Verify each project was discovered
		foundProjects := make(map[string]bool)
		for _, result := range discoveryResults {
			foundProjects[result.Name] = true
			assert.Equal(t, "local", result.Source)
			assert.Contains(t, result.URL, "project-")
			assert.Equal(t, 1, result.ComponentCount)
		}
		
		assert.True(t, foundProjects["project-1"])
		assert.True(t, foundProjects["project-2"])
		assert.True(t, foundProjects["project-3"])
	})
}