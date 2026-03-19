package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/ayo/internal/testutil"
)

func TestParseProject_MinimalValid(t *testing.T) {
	projectDir := testutil.CreateProject(t, "test-agent")

	got, err := ParseProject(projectDir)
	if err != nil {
		t.Fatalf("ParseProject() error = %v", err)
	}

	if got.Config.Name != "test-agent" {
		t.Errorf("Config.Name = %q, want %q", got.Config.Name, "test-agent")
	}

	if got.Config.Version != "1.0.0" {
		t.Errorf("Config.Version = %q, want %q", got.Config.Version, "1.0.0")
	}

	if got.System == "" {
		t.Error("System should not be empty")
	}

	if got.Hooks == nil {
		t.Error("Hooks should be initialized (not nil)")
	}
}

func TestParseProject_MissingConfig(t *testing.T) {
	dir := t.TempDir()

	_, err := ParseProject(dir)
	if err == nil {
		t.Error("ParseProject() expected error for missing config.toml")
	}
	if !strings.Contains(err.Error(), "config") {
		t.Errorf("Error should mention config, got: %v", err)
	}
}

func TestParseProject_MissingSystemMd(t *testing.T) {
	dir := t.TempDir()

	config := testutil.MinimalConfig("test")
	os.WriteFile(filepath.Join(dir, "config.toml"), []byte(config), 0644)

	_, err := ParseProject(dir)
	if err == nil {
		t.Error("ParseProject() expected error for missing system.md")
	}
	if !strings.Contains(err.Error(), "system.md") {
		t.Errorf("Error should mention system.md, got: %v", err)
	}
}

func TestParseProject_NonExistentPath(t *testing.T) {
	_, err := ParseProject("/nonexistent/path")
	if err == nil {
		t.Error("ParseProject() expected error for non-existent path")
	}
}

func TestParseProject_PathIsFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	_, err := ParseProject(tmpFile)
	if err == nil {
		t.Error("ParseProject() expected error when path is a file")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("Error should mention 'not a directory', got: %v", err)
	}
}

func TestParseProject_WithInputSchema(t *testing.T) {
	projectDir := testutil.CreateProjectWithFiles(t, "schema-agent", map[string]string{
		"input.jsonschema": testutil.ValidInputSchema(),
	})

	got, err := ParseProject(projectDir)
	if err != nil {
		t.Fatalf("ParseProject() error = %v", err)
	}

	if got.Input == nil {
		t.Fatal("Input should not be nil")
	}

	if got.Input.Parsed == nil {
		t.Error("Input.Parsed should not be nil")
	}
}

func TestParseProject_WithOutputSchema(t *testing.T) {
	projectDir := testutil.CreateProjectWithFiles(t, "schema-agent", map[string]string{
		"output.jsonschema": testutil.ValidOutputSchema(),
	})

	got, err := ParseProject(projectDir)
	if err != nil {
		t.Fatalf("ParseProject() error = %v", err)
	}

	if got.Output == nil {
		t.Fatal("Output should not be nil")
	}

	if got.Output.Parsed == nil {
		t.Error("Output.Parsed should not be nil")
	}
}

func TestParseProject_WithInvalidInputSchema(t *testing.T) {
	projectDir := testutil.CreateProjectWithFiles(t, "bad-schema", map[string]string{
		"input.jsonschema": testutil.InvalidJSON(),
	})

	_, err := ParseProject(projectDir)
	if err == nil {
		t.Error("ParseProject() expected error for invalid input.jsonschema")
	}
	if !strings.Contains(err.Error(), "input.jsonschema") {
		t.Errorf("Error should mention input.jsonschema, got: %v", err)
	}
}

func TestParseProject_WithInvalidOutputSchema(t *testing.T) {
	projectDir := testutil.CreateProjectWithFiles(t, "bad-schema", map[string]string{
		"output.jsonschema": testutil.InvalidJSON(),
	})

	_, err := ParseProject(projectDir)
	if err == nil {
		t.Error("ParseProject() expected error for invalid output.jsonschema")
	}
	if !strings.Contains(err.Error(), "output.jsonschema") {
		t.Errorf("Error should mention output.jsonschema, got: %v", err)
	}
}

func TestParseProject_WithPromptTemplate(t *testing.T) {
	promptContent := "Hello {{.input.query}}"
	projectDir := testutil.CreateProjectWithFiles(t, "prompt-agent", map[string]string{
		"prompt.tmpl": promptContent,
	})

	got, err := ParseProject(projectDir)
	if err != nil {
		t.Fatalf("ParseProject() error = %v", err)
	}

	if got.Prompt == nil {
		t.Fatal("Prompt should not be nil")
	}

	if *got.Prompt != promptContent {
		t.Errorf("Prompt = %q, want %q", *got.Prompt, promptContent)
	}
}

func TestParseProject_WithSkills(t *testing.T) {
	projectDir := testutil.CreateProject(t, "skill-agent")

	skillsDir := filepath.Join(projectDir, "skills", "test-skill")
	os.MkdirAll(skillsDir, 0755)
	os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte("# Test Skill\n\nA skill for testing."), 0644)

	got, err := ParseProject(projectDir)
	if err != nil {
		t.Fatalf("ParseProject() error = %v", err)
	}

	if len(got.Skills) != 1 {
		t.Fatalf("len(Skills) = %d, want 1", len(got.Skills))
	}

	if got.Skills[0].Name != "test-skill" {
		t.Errorf("Skills[0].Name = %q, want %q", got.Skills[0].Name, "test-skill")
	}

	if got.Skills[0].Description != "Test Skill" {
		t.Errorf("Skills[0].Description = %q, want %q", got.Skills[0].Description, "Test Skill")
	}
}

func TestParseProject_WithHooks(t *testing.T) {
	projectDir := testutil.CreateProject(t, "hook-agent")

	hooksDir := filepath.Join(projectDir, "hooks")
	os.MkdirAll(hooksDir, 0755)
	os.WriteFile(filepath.Join(hooksDir, "agent-start"), []byte("#!/bin/sh\necho start"), 0755)
	os.WriteFile(filepath.Join(hooksDir, "agent-finish"), []byte("#!/bin/sh\necho finish"), 0755)
	os.WriteFile(filepath.Join(hooksDir, "invalid-hook"), []byte("#!/bin/sh\necho invalid"), 0755)

	got, err := ParseProject(projectDir)
	if err != nil {
		t.Fatalf("ParseProject() error = %v", err)
	}

	if len(got.Hooks) != 2 {
		t.Errorf("len(Hooks) = %d, want 2 (invalid-hook should be ignored)", len(got.Hooks))
	}

	if _, ok := got.Hooks[HookAgentStart]; !ok {
		t.Error("Hooks should contain agent-start")
	}

	if _, ok := got.Hooks[HookAgentFinish]; !ok {
		t.Error("Hooks should contain agent-finish")
	}
}

func TestParseSkill(t *testing.T) {
	skillDir := filepath.Join(t.TempDir(), "my-skill")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# My Skill\n\nDescription here."), 0644)

	got, err := parseSkill(skillDir)
	if err != nil {
		t.Fatalf("parseSkill() error = %v", err)
	}

	if got.Name != "my-skill" {
		t.Errorf("Name = %q, want %q", got.Name, "my-skill")
	}

	if got.Description != "My Skill" {
		t.Errorf("Description = %q, want %q", got.Description, "My Skill")
	}
}

func TestParseSkill_NoSkillMd(t *testing.T) {
	skillDir := filepath.Join(t.TempDir(), "no-md-skill")
	os.MkdirAll(skillDir, 0755)

	got, err := parseSkill(skillDir)
	if err != nil {
		t.Fatalf("parseSkill() error = %v", err)
	}

	if got.Description != "" {
		t.Errorf("Description = %q, want empty", got.Description)
	}
}

func TestExtractDescription(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "simple header",
			content: "# My Title\n\nSome content",
			want:    "My Title",
		},
		{
			name:    "header with extra spaces",
			content: "#   My Title  \n\nContent",
			want:    "  My Title",
		},
		{
			name:    "no header",
			content: "No header here\nJust content",
			want:    "",
		},
		{
			name:    "header not first line",
			content: "Some intro\n# Title\nContent",
			want:    "Title",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDescription(tt.content)
			if got != tt.want {
				t.Errorf("extractDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsValidHookType(t *testing.T) {
	validHooks := []string{
		"agent-start", "agent-finish", "agent-error",
		"step-start", "step-finish",
		"text-start", "text-delta", "text-end",
		"reasoning-start", "reasoning-delta", "reasoning-end",
		"tool-input-start", "tool-input-delta", "tool-input-end",
		"tool-call", "tool-result",
		"source", "stream-finish", "warnings",
	}

	for _, hook := range validHooks {
		t.Run("valid_"+hook, func(t *testing.T) {
			if !isValidHookType(hook) {
				t.Errorf("isValidHookType(%q) = false, want true", hook)
			}
		})
	}

	invalidHooks := []string{
		"invalid-hook", "agent-started", "text-output", "", "AGENT-START",
	}

	for _, hook := range invalidHooks {
		t.Run("invalid_"+hook, func(t *testing.T) {
			if isValidHookType(hook) {
				t.Errorf("isValidHookType(%q) = true, want false", hook)
			}
		})
	}
}

func TestValidateProject_Valid(t *testing.T) {
	p := &Project{
		Config: AgentConfig{
			Name:    "test",
			Version: "1.0.0",
		},
		System: "You are helpful.",
	}

	errors := ValidateProject(p)
	if len(errors) != 0 {
		t.Errorf("ValidateProject() returned %d errors, want 0", len(errors))
		for _, e := range errors {
			t.Logf("  - %s: %s", e.File, e.Message)
		}
	}
}

func TestValidateProject_MissingName(t *testing.T) {
	p := &Project{
		Config: AgentConfig{
			Version: "1.0.0",
		},
		System: "System",
	}

	errors := ValidateProject(p)
	if len(errors) == 0 {
		t.Fatal("ValidateProject() expected error for missing name")
	}

	found := false
	for _, e := range errors {
		if e.File == "config.toml" && strings.Contains(e.Message, "name") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected error about missing name in config.toml")
	}
}

func TestValidateProject_MissingVersion(t *testing.T) {
	p := &Project{
		Config: AgentConfig{
			Name: "test",
		},
		System: "System",
	}

	errors := ValidateProject(p)
	if len(errors) == 0 {
		t.Fatal("ValidateProject() expected error for missing version")
	}
}

func TestValidateProject_EmptySystem(t *testing.T) {
	p := &Project{
		Config: AgentConfig{
			Name:    "test",
			Version: "1.0.0",
		},
		System: "",
	}

	errors := ValidateProject(p)
	if len(errors) == 0 {
		t.Fatal("ValidateProject() expected error for empty system")
	}
}
