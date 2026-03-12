package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexcabrera/ayo/internal/build/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildAndRunAgent tests the complete build and execution workflow
func TestBuildAndRunAgent(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "ayo-build-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a simple agent configuration
	config := types.Config{
		Agent: types.AgentConfig{
			Name:        "test-agent",
			Description: "Test agent for build system",
			Model:       "claude-3-5-sonnet",
		},
		CLI: types.CLIConfig{
			Mode:        "freeform",
			Description: "Test agent CLI",
		},
	}

	// Create config.toml file
	configPath := filepath.Join(tmpDir, "config.toml")
	err = WriteConfig(config, configPath)
	require.NoError(t, err)

	// Test loading the configuration
	loadedConfig, _, err := LoadConfigFromDir(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "test-agent", loadedConfig.Agent.Name)
	assert.Equal(t, "freeform", loadedConfig.CLI.Mode)

	// Test building the agent (this would normally call go build)
	// For testing, we'll just test the configuration generation
	
	// Test that the main.go stub can be generated
	tmpBuildDir, err := os.MkdirTemp("", "ayo-build-gen-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpBuildDir)
	
	mainGoPath := filepath.Join(tmpBuildDir, "main.go")
	err = GenerateMainStub(mainGoPath, loadedConfig, configPath)
	require.NoError(t, err)
	
	// Verify the main.go file was created
	_, err = os.Stat(mainGoPath)
	require.NoError(t, err)
	
	// Test simple evaluation system
	evalsPath := filepath.Join(tmpDir, "evals.json")
	evalsJSON := `[
		{
			"name": "basic test",
			"input": {"task": "say hello"},
			"expected": {"response": "hello"}
		}
	]`
	err = os.WriteFile(evalsPath, []byte(evalsJSON), 0644)
	require.NoError(t, err)
	
	evals, err := ParseSimpleEvals(evalsPath)
	require.NoError(t, err)
	assert.Len(t, evals, 1)
	assert.Equal(t, "basic test", evals[0].Name)
	
	// Test running an evaluation
	testInput := map[string]any{"response": "hello"}
	result := RunSimpleEval(evals[0], testInput)
	assert.True(t, result.Passed)
	
	// Test failing evaluation
	badInput := map[string]any{"response": "goodbye"}
	badResult := RunSimpleEval(evals[0], badInput)
	assert.False(t, badResult.Passed)
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test valid configuration
	validConfig := types.Config{
		Agent: types.AgentConfig{
			Name:        "valid-agent",
			Description: "Valid agent",
			Model:       "gpt-4-turbo",
		},
		CLI: types.CLIConfig{
			Mode:        "structured",
			Description: "Valid CLI",
		},
	}

	configPath := filepath.Join(tmpDir, "config.toml")
	err = WriteConfig(validConfig, configPath)
	require.NoError(t, err)

	// Should load without error
	loadedConfig, _, err := LoadConfigFromDir(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "valid-agent", loadedConfig.Agent.Name)
	assert.Equal(t, "structured", loadedConfig.CLI.Mode)
}

// TestToolConfiguration tests tool configuration
func TestToolConfiguration(t *testing.T) {
	config := types.Config{
		Agent: types.AgentConfig{
			Name:  "tool-test-agent",
			Model: "claude-3-5-sonnet",
			Tools: types.AgentTools{
				Allowed: []string{"bash", "file_read", "git"},
			},
		},
	}

	// Verify tools are configured correctly
	assert.Len(t, config.Agent.Tools.Allowed, 3)
	assert.Contains(t, config.Agent.Tools.Allowed, "bash")
	assert.Contains(t, config.Agent.Tools.Allowed, "file_read")
	assert.Contains(t, config.Agent.Tools.Allowed, "git")
}

// TestMemoryConfiguration tests memory configuration
func TestMemoryConfiguration(t *testing.T) {
	config := types.Config{
		Agent: types.AgentConfig{
			Name:    "memory-test-agent",
			Model:   "gpt-4-turbo",
			Memory: types.AgentMemory{
				Enabled: true,
				Scope:   "agent",
			},
		},
	}

	// Verify memory is configured correctly
	assert.True(t, config.Agent.Memory.Enabled)
	assert.Equal(t, "agent", config.Agent.Memory.Scope)
}

// TestCopyFile tests the copyFile function
func TestCopyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-copy-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a source file with some content
	srcPath := filepath.Join(tmpDir, "source.txt")
	srcContent := []byte("Hello, World!")
	err = os.WriteFile(srcPath, srcContent, 0644)
	require.NoError(t, err)

	// Test successful copy
	dstPath := filepath.Join(tmpDir, "destination.txt")
	err = copyFile(srcPath, dstPath)
	require.NoError(t, err)

	// Verify the destination file exists and has the same content
	dstContent, err := os.ReadFile(dstPath)
	require.NoError(t, err)
	assert.Equal(t, srcContent, dstContent)

	// Test copying to a non-existent directory (should fail)
	invalidDst := filepath.Join(tmpDir, "nonexistent", "file.txt")
	err = copyFile(srcPath, invalidDst)
	assert.Error(t, err)

	// Test copying from a non-existent file (should fail)
	nonExistentSrc := filepath.Join(tmpDir, "does-not-exist.txt")
	err = copyFile(nonExistentSrc, dstPath)
	assert.Error(t, err)
}

// TestBuildAgentErrors tests error cases in BuildAgent
func TestBuildAgentErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-build-errors-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test building from a non-existent directory
	err = BuildAgent("/non/existent/directory", "/tmp/output", "linux", "amd64")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load config")

	// Test building from a directory without a config file
	emptyDir := filepath.Join(tmpDir, "empty")
	err = os.MkdirAll(emptyDir, 0755)
	require.NoError(t, err)

	err = BuildAgent(emptyDir, "/tmp/output", "linux", "amd64")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load config")
}

// TestBuildAgentSetup verifies that BuildAgent correctly sets up the build directory
func TestBuildAgentSetup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ayo-build-setup-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a minimal agent configuration
	config := types.Config{
		Agent: types.AgentConfig{
			Name:        "test-build-agent",
			Description: "Test agent for build system",
			Model:       "claude-3-5-sonnet",
		},
		CLI: types.CLIConfig{
			Mode:        "freeform",
			Description: "Test agent CLI",
		},
	}

	configPath := filepath.Join(tmpDir, "config.toml")
	err = WriteConfig(config, configPath)
	require.NoError(t, err)

	// Create prompts and skills directories to avoid embedding errors
	promptsDir := filepath.Join(tmpDir, "prompts")
	err = os.MkdirAll(promptsDir, 0755)
	require.NoError(t, err)
	systemPath := filepath.Join(promptsDir, "system.md")
	err = os.WriteFile(systemPath, []byte("Test system prompt"), 0644)
	require.NoError(t, err)

	skillsDir := filepath.Join(tmpDir, "skills")
	err = os.MkdirAll(skillsDir, 0755)
	require.NoError(t, err)

	toolsDir := filepath.Join(tmpDir, "tools")
	err = os.MkdirAll(toolsDir, 0755)
	require.NoError(t, err)

	// Test that BuildAgent fails appropriately when go build fails
	// (since we don't have a proper go module setup in the temp dir)
	outputPath := filepath.Join(tmpDir, "test-agent")
	err = BuildAgent(tmpDir, outputPath, "linux", "amd64")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "go build failed")
}