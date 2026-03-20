package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const configTemplate = `[agent]
name = "%s"
version = "1.0.0"
description = "An AI agent"

[model]
requires_structured_output = false
requires_tools = false
requires_vision = false
suggested = ["claude-sonnet-4-6", "gpt-4o", "gemini-2.5-pro"]
default = "claude-sonnet-4-6"

[defaults]
temperature = 0.7
max_tokens = 4096
`

const systemTemplate = `# Agent Instructions

You are a helpful AI assistant. Your primary goal is to assist users with their requests in a clear, accurate, and helpful manner.

## Guidelines

- Be concise but thorough
- Ask clarifying questions when needed
- Provide examples when helpful
- Admit uncertainty when appropriate
`

const gitignoreTemplate = `# Ayo
*.exe
*.exe~
*.dll
*.so
*.dylib

# Generated binary (uncomment with actual name)
# %s
`

var freshCmd = &cobra.Command{
	Use:   "fresh <name>",
	Short: "Create a new agent project",
	Long: `Create a new agent project directory with template files.

The command creates a directory with the agent name containing:
  - config.toml: Agent configuration with default settings
  - system.md: Template system message
  - .gitignore: Common ignore patterns`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		
		if err := createProject(name); err != nil {
			exitError(err.Error())
		}
		
		printSuccess(fmt.Sprintf("Created agent project: %s", name))
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Edit %s/system.md to define agent behavior\n", name)
		fmt.Printf("  2. Run 'ayo checkit %s' to validate\n", name)
		fmt.Printf("  3. Run 'ayo runthat %s' to compile\n", name)
	},
}

func init() {
	rootCmd.AddCommand(freshCmd)

	// Hidden alias
	newAlias := &cobra.Command{
		Use:    "new <name>",
		Hidden: true,
		Run:    freshCmd.Run,
	}
	rootCmd.AddCommand(newAlias)
}

func createProject(name string) error {
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("directory '%s' already exists", name)
	}

	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	agentName := filepath.Base(name)
	configContent := fmt.Sprintf(configTemplate, agentName)
	if err := os.WriteFile(filepath.Join(name, "config.toml"), []byte(configContent), 0644); err != nil {
		return fmt.Errorf("creating config.toml: %w", err)
	}

	if err := os.WriteFile(filepath.Join(name, "system.md"), []byte(systemTemplate), 0644); err != nil {
		return fmt.Errorf("creating system.md: %w", err)
	}

	gitignoreContent := fmt.Sprintf(gitignoreTemplate, agentName)
	if err := os.WriteFile(filepath.Join(name, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("creating .gitignore: %w", err)
	}

	return nil
}
