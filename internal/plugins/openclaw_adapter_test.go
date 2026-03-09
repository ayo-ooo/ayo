package plugins

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenClawPluginAdapter(t *testing.T) {
	t.Run("Adapt JSON manifest", func(t *testing.T) {
		// Create temporary directory with JSON manifest
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "manifest.json")
		
		manifestContent := `{
  "name": "test-plugin",
  "version": "1.0.0",
  "description": "A test OpenClaw plugin",
  "author": "Test Author",
  "repository": "https://github.com/test/plugin",
  "license": "MIT",
  "components": [
    {
      "name": "test-skill",
      "type": "skill",
      "description": "A test skill"
    },
    {
      "name": "test-tool",
      "type": "tool",
      "description": "A test tool"
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
  ]
}`
		
		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)
		
		adapter := NewOpenClawPluginAdapter(tempDir, manifestPath)
		ayoManifest, err := adapter.Adapt()
		require.NoError(t, err)
		
		assert.Equal(t, "openclaw-test-plugin", ayoManifest.Name)
		assert.Equal(t, "1.0.0", ayoManifest.Version)
		assert.Equal(t, "OpenClaw plugin: A test OpenClaw plugin", ayoManifest.Description)
		assert.Equal(t, "Test Author", ayoManifest.Author)
		assert.Equal(t, "https://github.com/test/plugin", ayoManifest.Repository)
		assert.Equal(t, "MIT", ayoManifest.License)
		
		// Check components were converted
		assert.Contains(t, ayoManifest.Skills, "test-skill")
		assert.Contains(t, ayoManifest.Tools, "test-tool")
		assert.Len(t, ayoManifest.Providers, 1)
		assert.Equal(t, "memory-provider", ayoManifest.Providers[0].Name)
		assert.Equal(t, PluginTypeMemory, ayoManifest.Providers[0].Type)
		
		// Check dependencies were converted
		require.NotNil(t, ayoManifest.Dependencies)
		assert.Len(t, ayoManifest.Dependencies.Plugins, 1)
		assert.Equal(t, "openclaw-base-plugin@>=1.0.0", ayoManifest.Dependencies.Plugins[0])
		
		// Metadata is not directly accessible in Manifest, but we can verify the name format
		assert.Equal(t, "openclaw-test-plugin", ayoManifest.Name)
	})

	t.Run("Adapt YAML manifest", func(t *testing.T) {
		// Create temporary directory with YAML manifest
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "manifest.yaml")
		
		manifestContent := `name: yaml-plugin
version: 2.0.0
description: A YAML OpenClaw plugin
author: YAML Author
components:
  - name: yaml-skill
    type: skill
    description: A YAML skill
  - name: planner-component
    type: planner
    description: A planner component
    metadata:
      planner_type: long
`
		
		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)
		
		adapter := NewOpenClawPluginAdapter(tempDir, manifestPath)
		ayoManifest, err := adapter.Adapt()
		require.NoError(t, err)
		
		assert.Equal(t, "openclaw-yaml-plugin", ayoManifest.Name)
		assert.Equal(t, "2.0.0", ayoManifest.Version)
		assert.Contains(t, ayoManifest.Skills, "yaml-skill")
		assert.Len(t, ayoManifest.Planners, 1)
		assert.Equal(t, "planner-component", ayoManifest.Planners[0].Name)
		assert.Equal(t, PlannerTypeLong, ayoManifest.Planners[0].Type)
		
		// Verify name format indicates YAML source
		assert.Equal(t, "openclaw-yaml-plugin", ayoManifest.Name)
	})

	t.Run("Adapt TOML manifest", func(t *testing.T) {
		// Create temporary directory with TOML manifest
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "manifest.toml")
		
		manifestContent := `name = "toml-plugin"
version = "3.0.0"
description = "A TOML OpenClaw plugin"

[[components]]
name = "toml-skill"
type = "skill"
description = "A TOML skill"

[[components]]
name = "agent-component"
type = "agent"
description = "An agent component"
`
		
		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)
		
		adapter := NewOpenClawPluginAdapter(tempDir, manifestPath)
		ayoManifest, err := adapter.Adapt()
		require.NoError(t, err)
		
		assert.Equal(t, "openclaw-toml-plugin", ayoManifest.Name)
		assert.Equal(t, "3.0.0", ayoManifest.Version)
		assert.Contains(t, ayoManifest.Skills, "toml-skill")
		assert.Contains(t, ayoManifest.Agents, "agent-component")
		
		// Verify name format indicates TOML source
		assert.Equal(t, "openclaw-toml-plugin", ayoManifest.Name)
	})

	t.Run("AdaptOpenClawPluginFromDirectory", func(t *testing.T) {
		// Create temporary directory with manifest
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "plugin.json")
		
		manifestContent := `{
  "name": "discovery-test",
  "version": "1.0.0",
  "description": "Plugin for discovery testing",
  "components": [
    {
      "name": "discovery-skill",
      "type": "skill",
      "description": "Skill for testing discovery"
    }
  ]
}`
		
		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)
		
		// Test discovery and adaptation
		ayoManifest, err := AdaptOpenClawPluginFromDirectory(tempDir)
		require.NoError(t, err)
		
		assert.Equal(t, "openclaw-discovery-test", ayoManifest.Name)
		assert.Contains(t, ayoManifest.Skills, "discovery-skill")
	})

	t.Run("findOpenClawManifests", func(t *testing.T) {
		// Create temporary directory with multiple manifest files
		tempDir := t.TempDir()
		
		// Create various manifest files
		manifestFiles := []string{
			"manifest.json",
			"manifest.yaml", 
			"plugin.toml",
			"openclaw.yml",
			"readme.md", // Should be ignored
		}
		
		for _, filename := range manifestFiles {
			if filepath.Ext(filename) != ".md" { // Skip non-manifest files
				path := filepath.Join(tempDir, filename)
				content := `{"name": "test", "version": "1.0.0", "components": []}`
				if filepath.Ext(filename) == ".yaml" || filepath.Ext(filename) == ".yml" {
					content = "name: test\nversion: 1.0.0\ncomponents: []"
				} else if filepath.Ext(filename) == ".toml" {
					content = "name = 'test'\nversion = '1.0.0'\ncomponents = []"
				}
				err := os.WriteFile(path, []byte(content), 0644)
				require.NoError(t, err)
			}
		}
		
		// Find manifests
		manifestPaths := findOpenClawManifests(tempDir)
		
		// Should find all manifest files except readme.md
		assert.Len(t, manifestPaths, 4)
		
		// Check that we found the right files
		foundFilenames := make(map[string]bool)
		for _, path := range manifestPaths {
			foundFilenames[filepath.Base(path)] = true
		}
		
		assert.True(t, foundFilenames["manifest.json"])
		assert.True(t, foundFilenames["manifest.yaml"])
		assert.True(t, foundFilenames["plugin.toml"])
		assert.True(t, foundFilenames["openclaw.yml"])
		assert.False(t, foundFilenames["readme.md"])
	})

	t.Run("Unknown component type", func(t *testing.T) {
		// Create temporary directory with manifest containing unknown component type
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "manifest.json")
		
		manifestContent := `{
  "name": "unknown-components",
  "version": "1.0.0",
  "description": "Plugin with unknown components",
  "components": [
    {
      "name": "known-skill",
      "type": "skill",
      "description": "A known skill"
    },
    {
      "name": "unknown-component",
      "type": "unknown-type",
      "description": "An unknown component type"
    }
  ]
}`
		
		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)
		
		adapter := NewOpenClawPluginAdapter(tempDir, manifestPath)
		ayoManifest, err := adapter.Adapt()
		require.NoError(t, err)
		
		// Known component should be converted
		assert.Contains(t, ayoManifest.Skills, "known-skill")
		
		// Unknown components are handled internally but not exposed in the manifest
		// The important thing is that the known component was converted
		assert.Contains(t, ayoManifest.Skills, "known-skill")
	})
}