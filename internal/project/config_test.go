package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ayo-ooo/ayo/internal/testutil"
)

func TestParseConfig_Valid(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	os.WriteFile(configPath, []byte(testutil.MinimalConfig("test-agent")), 0644)

	got, err := ParseConfig(configPath)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if got.Name != "test-agent" {
		t.Errorf("Name = %q, want %q", got.Name, "test-agent")
	}

	if got.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", got.Version, "1.0.0")
	}

	if got.Description != "Test agent" {
		t.Errorf("Description = %q, want %q", got.Description, "Test agent")
	}
}

func TestParseConfig_MissingFile(t *testing.T) {
	_, err := ParseConfig("/nonexistent/config.toml")
	if err == nil {
		t.Error("ParseConfig() expected error for missing file")
	}
}

func TestParseConfig_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	os.WriteFile(configPath, []byte(`invalid [[toml`), 0644)

	_, err := ParseConfig(configPath)
	if err == nil {
		t.Error("ParseConfig() expected error for invalid TOML")
	}
}

func TestParseConfig_AllFields(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	fullConfig := `[agent]
name = "full-agent"
version = "2.0.0"
description = "A fully configured agent"

[model]
requires_structured_output = true
requires_tools = true
requires_vision = true
suggested = ["claude-3-opus", "gpt-4"]
default = "claude-3-opus"

[defaults]
temperature = 0.5
max_tokens = 8192
`
	os.WriteFile(configPath, []byte(fullConfig), 0644)

	got, err := ParseConfig(configPath)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if got.Name != "full-agent" {
		t.Errorf("Name = %q, want %q", got.Name, "full-agent")
	}

	if got.Model.RequiresStructuredOutput != true {
		t.Error("Model.RequiresStructuredOutput should be true")
	}

	if got.Model.RequiresTools != true {
		t.Error("Model.RequiresTools should be true")
	}

	if got.Model.RequiresVision != true {
		t.Error("Model.RequiresVision should be true")
	}

	if len(got.Model.Suggested) != 2 {
		t.Errorf("Model.Suggested length = %d, want 2", len(got.Model.Suggested))
	}

	if got.Model.Default != "claude-3-opus" {
		t.Errorf("Model.Default = %q, want %q", got.Model.Default, "claude-3-opus")
	}

	if got.Defaults.Temperature != 0.5 {
		t.Errorf("Defaults.Temperature = %v, want 0.5", got.Defaults.Temperature)
	}

	if got.Defaults.MaxTokens != 8192 {
		t.Errorf("Defaults.MaxTokens = %d, want 8192", got.Defaults.MaxTokens)
	}
}

func TestParseConfig_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	os.WriteFile(configPath, []byte{}, 0644)

	got, err := ParseConfig(configPath)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if got.Name != "" {
		t.Errorf("Name = %q, want empty", got.Name)
	}
}

func TestParseConfig_PartialFields(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	partialConfig := `[agent]
name = "partial-agent"
`
	os.WriteFile(configPath, []byte(partialConfig), 0644)

	got, err := ParseConfig(configPath)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	if got.Name != "partial-agent" {
		t.Errorf("Name = %q, want %q", got.Name, "partial-agent")
	}

	if got.Version != "" {
		t.Errorf("Version = %q, want empty", got.Version)
	}
}
