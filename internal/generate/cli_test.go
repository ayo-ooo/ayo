package generate

import (
	"strings"
	"testing"

	"github.com/charmbracelet/ayo/internal/project"
	"github.com/charmbracelet/ayo/internal/schema"
	"github.com/charmbracelet/ayo/internal/testutil"
)

func TestGenerateCLI_Basic(t *testing.T) {
	parsedSchema := mustParseSchema(testutil.ValidInputSchema())
	proj := &project.Project{
		Config: project.AgentConfig{
			Name:        "test-cli",
			Description: "A test CLI",
		},
		Input: &project.Schema{
			Content: []byte(testutil.ValidInputSchema()),
			Parsed:  parsedSchema,
		},
	}

	code, err := GenerateCLI(proj, "main")
	if err != nil {
		t.Fatalf("GenerateCLI() error = %v", err)
	}

	if !strings.Contains(code, "package main") {
		t.Error("Generated code should contain 'package main'")
	}

	if !strings.Contains(code, "import") {
		t.Error("Generated code should contain imports")
	}

	if !strings.Contains(code, "rootCmd") {
		t.Error("Generated code should contain rootCmd")
	}

	if !strings.Contains(code, "Execute()") {
		t.Error("Generated code should contain Execute function")
	}
}

func TestGenerateCLI_WithOutputSchema(t *testing.T) {
	parsedInput := mustParseSchema(testutil.ValidInputSchema())
	parsedOutput := mustParseSchema(testutil.ValidOutputSchema())

	proj := &project.Project{
		Config: project.AgentConfig{
			Name:        "test-cli",
			Description: "A test CLI",
		},
		Input: &project.Schema{
			Content: []byte(testutil.ValidInputSchema()),
			Parsed:  parsedInput,
		},
		Output: &project.Schema{
			Content: []byte(testutil.ValidOutputSchema()),
			Parsed:  parsedOutput,
		},
	}

	code, err := GenerateCLI(proj, "cmd")
	if err != nil {
		t.Fatalf("GenerateCLI() error = %v", err)
	}

	if !strings.Contains(code, "package cmd") {
		t.Error("Generated code should use specified package name")
	}

	if !strings.Contains(code, "encoding/json") {
		t.Error("Generated code should import encoding/json when output schema exists")
	}
}

func TestGenerateCLI_WithFlags(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"verbose": {
				"type": "boolean",
				"x-cli-flag": "--verbose",
				"x-cli-short": "-v",
				"description": "Enable verbose output"
			},
			"count": {
				"type": "integer",
				"x-cli-flag": "--count",
				"default": 10,
				"description": "Number of results"
			},
			"rate": {
				"type": "number",
				"x-cli-flag": "--rate",
				"default": 1.5,
				"description": "Rate limit"
			}
		}
	}`

	parsedSchema := mustParseSchema(schemaJSON)
	proj := &project.Project{
		Config: project.AgentConfig{
			Name:        "flag-test",
			Description: "Flag test",
		},
		Input: &project.Schema{
			Content: []byte(schemaJSON),
			Parsed:  parsedSchema,
		},
	}

	code, err := GenerateCLI(proj, "main")
	if err != nil {
		t.Fatalf("GenerateCLI() error = %v", err)
	}

	if !strings.Contains(code, "BoolVar") {
		t.Error("Generated code should use BoolVar for boolean flags")
	}

	if !strings.Contains(code, "IntVar") {
		t.Error("Generated code should use IntVar for integer flags")
	}

	if !strings.Contains(code, "Float64Var") {
		t.Error("Generated code should use Float64Var for number flags")
	}

	if !strings.Contains(code, "Shorthand") {
		t.Error("Generated code should set shorthand for short flags")
	}
}

func TestGenerateCLI_WithRequiredFlag(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"required": ["api-key"],
		"properties": {
			"api-key": {
				"type": "string",
				"x-cli-flag": "--api-key",
				"description": "API Key"
			}
		}
	}`

	parsedSchema := mustParseSchema(schemaJSON)
	proj := &project.Project{
		Config: project.AgentConfig{
			Name:        "required-test",
			Description: "Required test",
		},
		Input: &project.Schema{
			Content: []byte(schemaJSON),
			Parsed:  parsedSchema,
		},
	}

	code, err := GenerateCLI(proj, "main")
	if err != nil {
		t.Fatalf("GenerateCLI() error = %v", err)
	}

	if !strings.Contains(code, "MarkFlagRequired") {
		t.Error("Generated code should mark required flags")
	}
}

func TestGenerateCLI_WithPositionalArgs(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"file": {
				"type": "string",
				"x-cli-position": 1
			},
			"output": {
				"type": "string",
				"x-cli-position": 2
			}
		}
	}`

	parsedSchema := mustParseSchema(schemaJSON)
	proj := &project.Project{
		Config: project.AgentConfig{
			Name:        "positional-test",
			Description: "Positional test",
		},
		Input: &project.Schema{
			Content: []byte(schemaJSON),
			Parsed:  parsedSchema,
		},
	}

	code, err := GenerateCLI(proj, "main")
	if err != nil {
		t.Fatalf("GenerateCLI() error = %v", err)
	}

	if !strings.Contains(code, "input.File = args[0]") {
		t.Error("Generated code should assign first positional arg to File")
	}

	if !strings.Contains(code, "input.Output = args[1]") {
		t.Error("Generated code should assign second positional arg to Output")
	}
}

func TestGenerateCLI_NoInput(t *testing.T) {
	proj := &project.Project{
		Config: project.AgentConfig{
			Name:        "no-input-test",
			Description: "No input test",
		},
	}

	code, err := GenerateCLI(proj, "main")
	if err != nil {
		t.Fatalf("GenerateCLI() error = %v", err)
	}

	if !strings.Contains(code, "\tinput string") {
		t.Error("Generated code should have default input string variable when no input schema")
	}

	if !strings.Contains(code, "buildInput(args []string) string {") {
		t.Error("Generated code should have buildInput returning string when no input schema")
	}
}

func TestGenerateCLI_GeneratedHeader(t *testing.T) {
	parsedSchema := mustParseSchema(testutil.ValidInputSchema())
	proj := &project.Project{
		Config: project.AgentConfig{
			Name:        "header-test",
			Description: "Header test",
		},
		Input: &project.Schema{
			Content: []byte(testutil.ValidInputSchema()),
			Parsed:  parsedSchema,
		},
	}

	code, err := GenerateCLI(proj, "main")
	if err != nil {
		t.Fatalf("GenerateCLI() error = %v", err)
	}

	if !strings.Contains(code, "// Code generated by ayo. DO NOT EDIT.") {
		t.Error("Generated code should contain generation header")
	}
}

func TestGenerateFlags_Basic(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"x-cli-flag": "--name"
			}
		}
	}`

	parsed := mustParseSchema(schemaJSON)
	flags := schema.GenerateFlags(parsed)

	if len(flags) != 1 {
		t.Fatalf("Expected 1 flag, got %d", len(flags))
	}

	if flags[0].Name != "--name" {
		t.Errorf("Expected flag name '--name', got %q", flags[0].Name)
	}
}

func TestGenerateFlags_NilSchema(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("GenerateFlags(nil) should panic")
		}
	}()

	schema.GenerateFlags(nil)
}

func TestGenerateFlags_DefaultFlagName(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"query": {
				"type": "string"
			}
		}
	}`

	parsed := mustParseSchema(schemaJSON)
	flags := schema.GenerateFlags(parsed)

	if len(flags) != 1 {
		t.Fatalf("Expected 1 flag, got %d", len(flags))
	}

	if flags[0].Name != "--query" {
		t.Errorf("Expected flag name '--query', got %q", flags[0].Name)
	}
}

func TestGenerateFlags_PositionalSkipped(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"file": {
				"type": "string",
				"x-cli-position": 1
			}
		}
	}`

	parsed := mustParseSchema(schemaJSON)
	flags := schema.GenerateFlags(parsed)

	if len(flags) != 1 {
		t.Fatalf("Expected 1 flag, got %d", len(flags))
	}

	if flags[0].Position != 1 {
		t.Errorf("Expected position 1, got %d", flags[0].Position)
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Name", "name"},
		{"APIKey", "aPIKey"},
		{"verbose", "verbose"},
		{"A", "a"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toCamelCase(tt.input)
			if got != tt.expected {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEscapeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`hello "world"`, `hello \"world\"`},
		{"no quotes", "no quotes"},
		{`"quoted"`, `\"quoted\"`},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeString(tt.input)
			if got != tt.expected {
				t.Errorf("escapeString(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func mustParseSchema(json string) *schema.ParsedSchema {
	s, err := schema.ParseSchema([]byte(json))
	if err != nil {
		panic(err)
	}
	return s
}
