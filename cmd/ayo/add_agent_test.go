package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateAgentConfigTemplate(t *testing.T) {
	tests := []struct {
		name        string
		agentName   string
		description string
		model       string
		template    string
		wantErr     bool
	}{
		{
			name:        "standard template",
			agentName:   "test-agent",
			description: "Test agent description",
			model:       "claude-3-5-sonnet",
			template:    "standard",
			wantErr:     false,
		},
		{
			name:        "simple template",
			agentName:   "simple-agent",
			description: "Simple agent",
			model:       "gpt-4o",
			template:    "simple",
			wantErr:     false,
		},
		{
			name:        "advanced template",
			agentName:   "advanced-agent",
			description: "Advanced agent",
			model:       "claude-3-opus",
			template:    "advanced",
			wantErr:     false,
		},
		{
			name:        "default template",
			agentName:   "default-agent",
			description: "Default agent",
			model:       "gpt-4",
			template:    "unknown",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generateAgentConfigTemplate(tt.agentName, tt.description, tt.model, tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateAgentConfigTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify agent name is in the config
				if !strings.Contains(result, tt.agentName) {
					t.Errorf("config template missing agent name %q", tt.agentName)
				}

				// Verify description is in the config
				if !strings.Contains(result, tt.description) {
					t.Errorf("config template missing description %q", tt.description)
				}

				// Verify model is in the config
				if !strings.Contains(result, tt.model) {
					t.Errorf("config template missing model %q", tt.model)
				}

				// Verify [agent] section exists
				if !strings.Contains(result, "[agent]") {
					t.Error("config template missing [agent] section")
				}

				// Verify [cli] section exists
				if !strings.Contains(result, "[cli]") {
					t.Error("config template missing [cli] section")
				}
			}
		})
	}
}

func TestGenerateAgentSystemPrompt(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantLen  int
	}{
		{
			name:     "simple template",
			template: "simple",
			wantLen:  10, // At least some content
		},
		{
			name:     "standard template",
			template: "standard",
			wantLen:  10,
		},
		{
			name:     "advanced template",
			template: "advanced",
			wantLen:  10,
		},
		{
			name:     "unknown template (defaults to standard)",
			template: "unknown",
			wantLen:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateAgentSystemPrompt(tt.template)
			if len(result) < tt.wantLen {
				t.Errorf("generateAgentSystemPrompt() returned too short result, got len %d, want >= %d", len(result), tt.wantLen)
			}

			// Simple template should have "concise and direct"
			if tt.template == "simple" && !strings.Contains(result, "concise") {
				t.Error("simple template should contain 'concise'")
			}

			// Advanced template should have "Core Principles"
			if tt.template == "advanced" && !strings.Contains(result, "Core Principles") {
				t.Error("advanced template should contain 'Core Principles'")
			}
		})
	}
}

func TestGenerateAgentExampleSkill(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		wantSub   string
	}{
		{
			name:      "standard agent",
			agentName: "code-analyzer",
			wantSub:   "code-analyzer",
		},
		{
			name:      "agent with spaces",
			agentName: "data processor",
			wantSub:   "data processor",
		},
		{
			name:      "single word name",
			agentName: "agent",
			wantSub:   "agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateAgentExampleSkill(tt.agentName)

			if !strings.Contains(result, tt.wantSub) {
				t.Errorf("example skill should contain agent name %q", tt.wantSub)
			}

			// Should contain some expected sections
			if !strings.Contains(result, "## Behavior") {
				t.Error("example skill should contain '## Behavior' section")
			}

			if !strings.Contains(result, "## Special Instructions") {
				t.Error("example skill should contain '## Special Instructions' section")
			}

			if !strings.Contains(result, "## Examples") {
				t.Error("example skill should contain '## Examples' section")
			}
		})
	}
}

func TestRunAddAgentToSingleAgentProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a single-agent project (has config.toml)
	configContent := `[agent]
name = "existing-agent"
description = "Existing agent"
model = "gpt-4o"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Add a new agent (won't promote to team since agents/ doesn't exist)
	err := runAddAgent(tmpDir, "new-agent", "New agent description", "gpt-4o", "standard")
	if err != nil {
		t.Fatalf("runAddAgent failed: %v", err)
	}

	// Verify agent directory was created
	agentDir := filepath.Join(tmpDir, "agents", "new-agent")
	if _, err := os.Stat(agentDir); os.IsNotExist(err) {
		t.Error("agent directory was not created")
	}

	// Verify config.toml was created
	agentConfigPath := filepath.Join(agentDir, "config.toml")
	if _, err := os.Stat(agentConfigPath); os.IsNotExist(err) {
		t.Error("agent config.toml was not created")
	}

	// Verify config content
	content, err := os.ReadFile(agentConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "new-agent") {
		t.Error("config doesn't contain agent name")
	}
	if !strings.Contains(contentStr, "New agent description") {
		t.Error("config doesn't contain description")
	}

	// Verify system prompt was created
	systemPromptPath := filepath.Join(agentDir, "prompts", "system.md")
	if _, err := os.Stat(systemPromptPath); os.IsNotExist(err) {
		t.Error("system.md was not created")
	}

	// Verify example skill was created
	skillPath := filepath.Join(agentDir, "skills", "custom", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Error("example SKILL.md was not created")
	}
}

func TestRunAddAgentToTeamProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a team project (has team.toml)
	teamConfigContent := `name = "myteam"
description = "My team"

[agents]
agent1 = { path = "agents/agent1" }
`
	teamConfigPath := filepath.Join(tmpDir, "team.toml")
	if err := os.WriteFile(teamConfigPath, []byte(teamConfigContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Add a new agent to the team
	err := runAddAgent(tmpDir, "new-agent", "New team agent", "gpt-4o", "standard")
	if err != nil {
		t.Fatalf("runAddAgent failed: %v", err)
	}

	// Verify agent directory was created
	agentDir := filepath.Join(tmpDir, "agents", "new-agent")
	if _, err := os.Stat(agentDir); os.IsNotExist(err) {
		t.Error("agent directory was not created")
	}

	// Verify team.toml was updated
	content, err := os.ReadFile(teamConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "new-agent") {
		t.Error("team.toml was not updated with new agent")
	}
}

func TestRunAddAgentSimpleTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a single-agent project
	configContent := `[agent]
name = "existing-agent"
description = "Existing agent"
model = "gpt-4o"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Add an agent with simple template
	err := runAddAgent(tmpDir, "simple-agent", "Simple agent", "gpt-4o", "simple")
	if err != nil {
		t.Fatalf("runAddAgent failed: %v", err)
	}

	// Verify agent was created
	agentDir := filepath.Join(tmpDir, "agents", "simple-agent")
	if _, err := os.Stat(agentDir); os.IsNotExist(err) {
		t.Error("agent directory was not created")
	}

	// Verify config content
	agentConfigPath := filepath.Join(agentDir, "config.toml")
	content, err := os.ReadFile(agentConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, `mode = "freeform"`) {
		t.Error("simple template should use freeform mode")
	}
}

func TestRunAddAgentAdvancedTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a single-agent project
	configContent := `[agent]
name = "existing-agent"
description = "Existing agent"
model = "gpt-4o"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Add an agent with advanced template
	err := runAddAgent(tmpDir, "advanced-agent", "Advanced agent", "claude-3-opus", "advanced")
	if err != nil {
		t.Fatalf("runAddAgent failed: %v", err)
	}

	// Verify agent was created
	agentDir := filepath.Join(tmpDir, "agents", "advanced-agent")
	if _, err := os.Stat(agentDir); os.IsNotExist(err) {
		t.Error("agent directory was not created")
	}

	// Verify config content
	agentConfigPath := filepath.Join(agentDir, "config.toml")
	content, err := os.ReadFile(agentConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, `mode = "structured"`) {
		t.Error("advanced template should use structured mode")
	}
	if !strings.Contains(contentStr, "web_search") {
		t.Error("advanced template should include web_search tool")
	}
}

func TestRunAddAgentErrors(t *testing.T) {
	// Test with non-existent directory
	err := runAddAgent("/nonexistent/directory", "agent", "desc", "gpt-4o", "standard")
	if err == nil {
		t.Error("runAddAgent should error for non-existent directory")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected 'does not exist' error, got: %v", err)
	}

	// Test with directory that's not an ayo project
	tmpDir := t.TempDir()
	err = runAddAgent(tmpDir, "agent", "desc", "gpt-4o", "standard")
	if err == nil {
		t.Error("runAddAgent should error for non-ayo project")
	}
	if !strings.Contains(err.Error(), "not a valid ayo project") {
		t.Errorf("expected 'not a valid ayo project' error, got: %v", err)
	}
}

func TestRunAddAgentAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a single-agent project
	configContent := `[agent]
name = "existing-agent"
description = "Existing agent"
model = "gpt-4o"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create an existing agent directory
	agentsDir := filepath.Join(tmpDir, "agents", "existing")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Try to add an agent that already exists
	err := runAddAgent(tmpDir, "existing", "desc", "gpt-4o", "standard")
	if err == nil {
		t.Error("runAddAgent should error when agent already exists")
	}
	if !strings.Contains(err.Error(), "agent already exists") {
		t.Errorf("expected 'agent already exists' error, got: %v", err)
	}
}

func TestUpdateTeamConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an existing team.toml
	teamConfigContent := `name = "myteam"
description = "My team"

[agents]
agent1 = { path = "agents/agent1" }
`
	teamConfigPath := filepath.Join(tmpDir, "team.toml")
	if err := os.WriteFile(teamConfigPath, []byte(teamConfigContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Update the team config with a new agent
	err := updateTeamConfig(teamConfigPath, "agent2")
	if err != nil {
		t.Fatalf("updateTeamConfig failed: %v", err)
	}

	// Verify the file was updated
	content, err := os.ReadFile(teamConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "agent2") {
		t.Error("team.toml was not updated with new agent")
	}
}

func TestUpdateTeamConfigNoAgentsSection(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a team.toml without agents section
	teamConfigContent := `name = "myteam"
description = "My team"
`
	teamConfigPath := filepath.Join(tmpDir, "team.toml")
	if err := os.WriteFile(teamConfigPath, []byte(teamConfigContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Update the team config
	err := updateTeamConfig(teamConfigPath, "agent1")
	if err != nil {
		t.Fatalf("updateTeamConfig failed: %v", err)
	}

	// Verify the file was updated with agents section
	content, err := os.ReadFile(teamConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "[agents]") {
		t.Error("team.toml was not updated with [agents] section")
	}
	if !strings.Contains(contentStr, "agent1") {
		t.Error("team.toml was not updated with new agent")
	}
}

func TestCreateTeamProjectFromSingleAgent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a single-agent project
	configContent := `[agent]
name = "existing-agent"
description = "Existing agent"
model = "gpt-4o"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create agents directory with an existing agent
	agentsDir := filepath.Join(tmpDir, "agents", "agent1")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create team project
	err := createTeamProjectFromSingleAgent("myteam", tmpDir, []string{"agent1", "agent2"})
	if err != nil {
		t.Fatalf("createTeamProjectFromSingleAgent failed: %v", err)
	}

	// Verify team.toml was created
	teamConfigPath := filepath.Join(tmpDir, "team.toml")
	if _, err := os.Stat(teamConfigPath); os.IsNotExist(err) {
		t.Error("team.toml was not created")
	}

	// Verify workspace directory was created
	workspaceDir := filepath.Join(tmpDir, "workspace")
	if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
		t.Error("workspace directory was not created")
	}

	// Verify SQUAD.md was created
	squadPath := filepath.Join(tmpDir, "SQUAD.md")
	if _, err := os.Stat(squadPath); os.IsNotExist(err) {
		t.Error("SQUAD.md was not created")
	}

	// Verify team.toml content
	content, err := os.ReadFile(teamConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "agent1") {
		t.Error("team.toml doesn't contain agent1")
	}
	if !strings.Contains(contentStr, "agent2") {
		t.Error("team.toml doesn't contain agent2")
	}
}
