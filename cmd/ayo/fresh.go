package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newFreshCmd() *cobra.Command {
	var agentName, description, model string
	var template string

	cmd := &cobra.Command{
		Use:   "fresh [directory]",
		Short: "Create a new agent project",
		Long: `Create a new agent project directory with a template configuration.

This command creates:
- config.toml - Agent configuration
- skills/     - Directory for agent-specific skills
- tools/      - Directory for custom Go tools
- prompts/    - Directory for prompt templates

After initialization, edit config.toml to customize your agent, then run:
  ayo build [directory]

Examples:
  ayo fresh myreviewer
  ayo fresh myagent --template simple
  ayo fresh ./agents/code-reviewer --name reviewer --description "Code reviewer agent"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine project directory
			var dir string
			if len(args) > 0 {
				dir = args[0]
			} else if agentName != "" {
				dir = agentName
			} else {
				return fmt.Errorf("must specify directory or --name")
			}

			// Set defaults
			if agentName == "" {
				agentName = filepath.Base(dir)
			}
			if description == "" {
				description = fmt.Sprintf("AI agent: %s", agentName)
			}
			if model == "" {
				model = "claude-3-5-sonnet" // Default model
			}
			if template == "" {
				template = "standard" // Default template
			}

			return runFresh(dir, agentName, description, model, template)
		},
	}

	cmd.Flags().StringVar(&agentName, "name", "", "Agent name (default: directory name)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Agent description")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Default model (default: claude-3-5-sonnet)")
	cmd.Flags().StringVar(&template, "template", "standard", "Template to use: standard, simple, advanced")

	return cmd
}

func runFresh(dir, name, description, model, template string) error {
	// Resolve to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve directory: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); err == nil {
		return fmt.Errorf("directory already exists: %s", absDir)
	}

	// Create directory structure
	dirs := []string{
		absDir,
		filepath.Join(absDir, "skills"),
		filepath.Join(absDir, "tools"),
		filepath.Join(absDir, "prompts"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", d, err)
		}
	}

	// Generate config.toml based on template
	configContent, err := generateConfigTemplate(name, description, model, template)
	if err != nil {
		return fmt.Errorf("generate config template: %w", err)
	}

	// Write config.toml
	configPath := filepath.Join(absDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("write config.toml: %w", err)
	}

	// Create example system prompt
	systemPrompt := generateSystemPrompt(template)
	systemPath := filepath.Join(absDir, "prompts", "system.md")
	if err := os.WriteFile(systemPath, []byte(systemPrompt), 0644); err != nil {
		return fmt.Errorf("write system.md: %w", err)
	}

	// Create example skill if template supports it
	if template == "standard" || template == "advanced" {
		skillContent := generateExampleSkill(name)
		skillPath := filepath.Join(absDir, "skills", "custom", "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(skillPath), 0755); err != nil {
			return fmt.Errorf("create skills/custom directory: %w", err)
		}
		if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
			return fmt.Errorf("write example skill: %w", err)
		}
	}

	// Create .gitkeep in tools/ to ensure embed works when empty
	toolsGitkeep := filepath.Join(absDir, "tools", ".gitkeep")
	if err := os.WriteFile(toolsGitkeep, []byte("# Keep this directory for embed\n"), 0644); err != nil {
		return fmt.Errorf("write tools/.gitkeep: %w", err)
	}

	// Create .gitkeep in skills/ to ensure embed works when empty
	skillsGitkeep := filepath.Join(absDir, "skills", ".gitkeep")
	if err := os.WriteFile(skillsGitkeep, []byte("# Keep this directory for embed\n"), 0644); err != nil {
		return fmt.Errorf("write skills/.gitkeep: %w", err)
	}

	// Print success message
	fmt.Printf("Created new agent project: %s\n", absDir)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Edit %s to customize your agent\n", configPath)
	fmt.Printf("  2. Add custom tools in %s/ (optional)\n", filepath.Join(absDir, "tools"))
	fmt.Printf("  3. Add skills in %s/ (optional)\n", filepath.Join(absDir, "skills"))
	fmt.Printf("  4. Build your agent: ayo build %s\n", dir)
	fmt.Printf("  5. Run: ./%s\n", name)

	return nil
}

func generateConfigTemplate(name, description, model, template string) (string, error) {
	var baseConfig string

	switch template {
	case "simple":
		baseConfig = `[agent]
name = "%s"
description = "%s"
model = "%s"

[cli]
mode = "freeform"
description = "%s"

[cli.flags]

[agent.tools]
allowed = ["bash", "file_read", "file_write"]

[agent.memory]
enabled = true
scope = "agent"

[agent.sandbox]
network = false
host_path = "."

[triggers]
watch = []
schedule = ""
events = []
`
	case "standard":
		baseConfig = `[agent]
name = "%s"
description = "%s"
model = "%s"

[cli]
mode = "hybrid"
description = "%s"

[cli.flags]

[agent.tools]
allowed = ["bash", "file_read", "file_write", "git"]

[agent.memory]
enabled = true
scope = "agent"

[agent.sandbox]
network = false
host_path = "."

[triggers]
watch = []
schedule = ""
events = []
`
	case "advanced":
		baseConfig = `[agent]
name = "%s"
description = "%s"
model = "%s"

[cli]
mode = "structured"
description = "%s"

[cli.flags]

[agent.tools]
allowed = ["bash", "file_read", "file_write", "git", "web_search"]

[agent.memory]
enabled = true
scope = "session"

[agent.sandbox]
network = true
host_path = "."

[triggers]
watch = []
schedule = ""
events = []
`
	default:
		baseConfig = `[agent]
name = "%s"
description = "%s"
model = "%s"

[cli]
mode = "hybrid"
description = "%s"

[cli.flags]

[agent.tools]
allowed = ["bash", "file_read", "file_write", "git"]

[agent.memory]
enabled = true
scope = "agent"

[agent.sandbox]
network = false
host_path = "."

[triggers]
watch = []
schedule = ""
events = []
`
	}

	return fmt.Sprintf(baseConfig, name, description, model, description), nil
}

func generateSystemPrompt(template string) string {
	switch template {
	case "simple":
		return `You are a helpful AI assistant. Use your tools to complete tasks efficiently.

Follow these guidelines:
- Be concise and direct
- Use tools when needed
- Ask clarifying questions if requirements are unclear`
	case "advanced":
		return `You are an expert AI assistant with access to multiple tools and capabilities.

## Core Principles
1. Be precise and thorough in your responses
2. Always verify information before presenting it
3. Use appropriate tools for each task
4. Maintain context across the conversation
5. Think step-by-step for complex problems

## Tool Usage
- Use bash commands for system operations
- Read files to understand context
- Write files to produce outputs
- Use git for version control operations
- Use web_search for external information when needed

## Communication Style
- Provide clear explanations
- Include reasoning for important decisions
- Flag assumptions you're making
- Suggest alternatives when appropriate`
	default: // standard
		return `You are a capable AI assistant. Help users accomplish their tasks using available tools.

## Guidelines
- Understand the user's goal before acting
- Use tools appropriately and efficiently
- Provide clear, helpful responses
- Ask questions when requirements are unclear
- Maintain context throughout the conversation

## Available Tools
You have access to various tools for file operations, bash commands, and more. Use them as needed to complete tasks.`
	}
}

func generateExampleSkill(name string) string {
	return `# Custom Skill: %[1]s Specific

This skill provides additional capabilities specific to the %[1]s agent.

## Behavior

When processing requests:

1. **Analyze the Request**: Understand what the user is asking for
2. **Determine the Approach**: Choose the best tools and methods
3. **Execute**: Perform the task using available tools
4. **Verify**: Check the results
5. **Report**: Provide clear feedback to the user

## Special Instructions

- Be thorough in your analysis
- Prioritize correctness over speed
- Document any assumptions you make
- Suggest improvements when appropriate

## Examples

### Example 1
User: "Help me understand this code"
Action: Read the file, analyze its structure, explain key components

### Example 2
User: "Fix this bug"
Action: Read the code, identify the issue, propose and implement a fix

---
Agent: %[1]s
Purpose: Custom behavior specific to this agent
`[1:]
}
