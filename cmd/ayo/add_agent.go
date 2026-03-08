package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

func newAddAgentCmd() *cobra.Command {
	var agentName, description, model string
	var template string

	cmd := &cobra.Command{
		Use:   "add-agent [directory] [agent-name]",
		Short: "Add an agent to an existing project",
		Long: `Add an additional agent to an existing ayo project (single-agent or team project).` + "\n\n" +
			`This command adds a new agent configuration to an existing project. For single-agent projects,
		it converts the project to a multi-agent structure. For team projects, it adds the agent
		to the existing team configuration.` + "\n\n" +
			`Examples:
  ayo add-agent myproject reviewer
  ayo add-agent myproject security-agent --template advanced
  ayo add-agent ./myteam code-analyzer --name analyzer --description "Code analysis agent"`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine project directory and agent name
			var dir, agentNameArg string
			if len(args) == 2 {
				dir = args[0]
				agentNameArg = args[1]
			} else if len(args) == 1 {
				if agentName != "" {
					dir = args[0]
					agentNameArg = agentName
				} else {
					return fmt.Errorf("must specify agent name via second argument or --name flag")
				}
			} else {
				return fmt.Errorf("must specify directory and agent name")
			}

			// Set defaults
			if agentName == "" {
				agentName = agentNameArg
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

			return runAddAgent(dir, agentName, description, model, template)
		},
	}

	cmd.Flags().StringVar(&agentName, "name", "", "Agent name (default: second argument)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Agent description")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Default model (default: claude-3-5-sonnet)")
	cmd.Flags().StringVar(&template, "template", "standard", "Template to use: standard, simple, advanced")

	return cmd
}

func runAddAgent(dir, name, description, model, template string) error {
	// Resolve to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve directory: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); err != nil {
		return fmt.Errorf("project directory does not exist: %s", absDir)
	}

	// Check if this is a valid ayo project
	configPath := filepath.Join(absDir, "config.toml")
	teamConfigPath := filepath.Join(absDir, "team.toml")

	isSingleAgentProject := false
	isTeamProject := false

	if _, err := os.Stat(configPath); err == nil {
		isSingleAgentProject = true
	}

	if _, err := os.Stat(teamConfigPath); err == nil {
		isTeamProject = true
	}

	if !isSingleAgentProject && !isTeamProject {
		return fmt.Errorf("not a valid ayo project: neither config.toml nor team.toml found in %s", absDir)
	}

	// Check if we need to promote single-agent project to team project
	if isSingleAgentProject && !isTeamProject {
		// Count existing agents
		agentsDir := filepath.Join(absDir, "agents")
		if _, err := os.Stat(agentsDir); err == nil {
			// Count subdirectories in agents/
			entries, err := os.ReadDir(agentsDir)
			if err == nil {
				agentCount := 0
				for _, entry := range entries {
					if entry.IsDir() {
						agentCount++
					}
				}

				// If we already have 1 agent and are adding another, promote to team
				if agentCount >= 1 {
					// Create team project
					teamName := filepath.Base(absDir)

					// Get existing agent names
					existingAgents := []string{}
					for _, entry := range entries {
						if entry.IsDir() {
							existingAgents = append(existingAgents, entry.Name())
						}
					}

					// Add the new agent to the list
					existingAgents = append(existingAgents, name)

					// Create team project
					if err := createTeamProjectFromSingleAgent(teamName, absDir, existingAgents); err != nil {
						return fmt.Errorf("failed to create team project: %w", err)
					}

					isTeamProject = true
					fmt.Printf("Promoted project to team format with %d agents\n", len(existingAgents))
				}
			}
		}
	}

	// Create agents directory if it doesn't exist
	agentsDir := filepath.Join(absDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("create agents directory: %w", err)
	}

	// Create agent subdirectory
	agentDir := filepath.Join(agentsDir, name)
	if _, err := os.Stat(agentDir); err == nil {
		return fmt.Errorf("agent already exists: %s", agentDir)
	}

	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return fmt.Errorf("create agent directory: %w", err)
	}

	// Create agent subdirectories
	dirs := []string{
		filepath.Join(agentDir, "skills"),
		filepath.Join(agentDir, "tools"),
		filepath.Join(agentDir, "prompts"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", d, err)
		}
	}

	// Generate config.toml for the agent
	configContent, err := generateAgentConfigTemplate(name, description, model, template)
	if err != nil {
		return fmt.Errorf("generate config template: %w", err)
	}

	// Write config.toml
	agentConfigPath := filepath.Join(agentDir, "config.toml")
	if err := os.WriteFile(agentConfigPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("write config.toml: %w", err)
	}

	// Create example system prompt
	systemPrompt := generateAgentSystemPrompt(template)
	systemPath := filepath.Join(agentDir, "prompts", "system.md")
	if err := os.WriteFile(systemPath, []byte(systemPrompt), 0644); err != nil {
		return fmt.Errorf("write system.md: %w", err)
	}

	// Create example skill if template supports it
	if template == "standard" || template == "advanced" {
		skillContent := generateAgentExampleSkill(name)
		skillPath := filepath.Join(agentDir, "skills", "custom", "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(skillPath), 0755); err != nil {
			return fmt.Errorf("create skills/custom directory: %w", err)
		}
		if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
			return fmt.Errorf("write example skill: %w", err)
		}
	}

	// Update team.toml if this is a team project
	if isTeamProject {
		if err := updateTeamConfig(teamConfigPath, name); err != nil {
			return fmt.Errorf("update team.toml: %w", err)
		}
	}

	// Print success message
	fmt.Printf("Added agent '%s' to project: %s\n", name, absDir)
	fmt.Printf("Agent configuration: %s\n", agentConfigPath)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Edit %s to customize your agent\n", agentConfigPath)
	fmt.Printf("  2. Add custom tools in %s/ (optional)\n", filepath.Join(agentDir, "tools"))
	fmt.Printf("  3. Add skills in %s/ (optional)\n", filepath.Join(agentDir, "skills"))
	if isTeamProject {
		fmt.Printf("  4. Build your team: ayo build %s\n", absDir)
	} else {
		fmt.Printf("  4. Build your agent: ayo build %s --agent %s\n", absDir, name)
	}

	return nil
}

func generateAgentConfigTemplate(name, description, model, template string) (string, error) {
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
host_path = ".."

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
host_path = ".."

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
host_path = ".."

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
host_path = ".."

[triggers]
watch = []
schedule = ""
events = []
`
	}

	return fmt.Sprintf(baseConfig, name, description, model, description), nil
}

func generateAgentSystemPrompt(template string) string {
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

func generateAgentExampleSkill(name string) string {
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

func createTeamProjectFromSingleAgent(teamName, teamDir string, agentNames []string) error {
	// Create team.toml
	teamConfig := struct {
		Team struct {
			Name         string `toml:"name"`
			Description  string `toml:"description"`
			Coordination string `toml:"coordination"`
		} `toml:"team"`
		Agents map[string]struct {
			Path string `toml:"path"`
		} `toml:"agents"`
		Workspace struct {
			SharedPath string `toml:"shared_path"`
			OutputPath string `toml:"output_path"`
		} `toml:"workspace"`
		Coordination struct {
			Strategy      string `toml:"strategy"`
			MaxIterations int    `toml:"max_iterations"`
		} `toml:"coordination"`
	}{
		Team: struct {
			Name         string `toml:"name"`
			Description  string `toml:"description"`
			Coordination string `toml:"coordination"`
		}{
			Name:         teamName,
			Description:  "Team description",
			Coordination: "sequential",
		},
		Agents: make(map[string]struct {
			Path string `toml:"path"`
		}),
		Workspace: struct {
			SharedPath string `toml:"shared_path"`
			OutputPath string `toml:"output_path"`
		}{
			SharedPath: "workspace",
			OutputPath: "workspace/results",
		},
		Coordination: struct {
			Strategy      string `toml:"strategy"`
			MaxIterations int    `toml:"max_iterations"`
		}{
			Strategy:      "round-robin",
			MaxIterations: 5,
		},
	}

	// Add agents
	for _, agentName := range agentNames {
		teamConfig.Agents[agentName] = struct {
			Path string `toml:"path"`
		}{
			Path: filepath.Join("agents", agentName),
		}
	}

	// Create directories
	if err := os.MkdirAll(filepath.Join(teamDir, "workspace"), 0755); err != nil {
		return fmt.Errorf("create workspace directory: %w", err)
	}

	// Marshal to TOML
	data, err := toml.Marshal(teamConfig)
	if err != nil {
		return fmt.Errorf("marshal team.toml: %w", err)
	}

	// Write team.toml
	teamPath := filepath.Join(teamDir, "team.toml")
	if err := os.WriteFile(teamPath, data, 0644); err != nil {
		return fmt.Errorf("write team.toml: %w", err)
	}

	// Create default constitution
	var agentSection strings.Builder
	for _, agent := range agentNames {
		agentSection.WriteString(fmt.Sprintf(`### %s
**Role**: [Define this agent's role]
**Responsibilities**:
- [Responsibility 1]
- [Responsibility 2]

`, agent))
	}

	template := fmt.Sprintf(`# Team: %s

## Mission

[Describe what this team is trying to accomplish in 1-2 paragraphs.]

## Context

[Background information all agents need: project constraints, technical decisions,
external dependencies, deadlines, or any shared knowledge.]

## Agents

%s
## Coordination

[How agents should work together: handoff protocols, communication patterns,
dependency chains, blocking rules.]

## Guidelines

[Specific rules or preferences for this team: coding style, testing requirements,
commit conventions, review process.]
`, teamName, agentSection.String())

	constPath := filepath.Join(teamDir, "SQUAD.md")
	if err := os.WriteFile(constPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("write SQUAD.md: %w", err)
	}

	return nil
}

func updateTeamConfig(teamConfigPath, agentName string) error {
	// Read existing team.toml
	content, err := os.ReadFile(teamConfigPath)
	if err != nil {
		return fmt.Errorf("read team.toml: %w", err)
	}

	contentStr := string(content)

	// Check if agents section exists
	if !strings.Contains(contentStr, "[agents]") {
		// Add agents section
		contentStr += "\n[agents]\n"
	}

	// Add agent to the list (simple append approach)
	// In a real implementation, this would parse TOML and add properly
	contentStr += fmt.Sprintf("  %s = { path = \"agents/%s\" }\n", agentName, agentName)

	// Write back
	if err := os.WriteFile(teamConfigPath, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("write team.toml: %w", err)
	}

	return nil
}
