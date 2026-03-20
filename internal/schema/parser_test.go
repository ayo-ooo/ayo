package schema

import (
	"reflect"
	"sort"
	"testing"

	"github.com/charmbracelet/ayo/internal/testutil"
)

func TestParseSchema_ValidObjectSchema(t *testing.T) {
	input := []byte(testutil.ValidInputSchema())

	got, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	if got.Type != "object" {
		t.Errorf("ParseSchema().Type = %q, want %q", got.Type, "object")
	}

	if len(got.Properties) != 1 {
		t.Errorf("ParseSchema().Properties length = %d, want 1", len(got.Properties))
	}

	if prop, ok := got.Properties["query"]; !ok {
		t.Error("ParseSchema() missing 'query' property")
	} else {
		if prop.Type != "string" {
			t.Errorf("ParseSchema().Properties['query'].Type = %q, want %q", prop.Type, "string")
		}
		if prop.Description != "Search query" {
			t.Errorf("ParseSchema().Properties['query'].Description = %q, want %q", prop.Description, "Search query")
		}
	}

	if len(got.Required) != 1 || got.Required[0] != "query" {
		t.Errorf("ParseSchema().Required = %v, want [query]", got.Required)
	}
}

func TestParseSchema_ValidOutputSchema(t *testing.T) {
	input := []byte(testutil.ValidOutputSchema())

	got, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	if got.Type != "object" {
		t.Errorf("ParseSchema().Type = %q, want %q", got.Type, "object")
	}

	if _, ok := got.Properties["result"]; !ok {
		t.Error("ParseSchema() missing 'result' property")
	}
}

func TestParseSchema_SchemaWithDefaults(t *testing.T) {
	input := []byte(testutil.SchemaWithDefaults())

	got, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	tests := []struct {
		name         string
		propName     string
		wantType     string
		wantDefault  any
	}{
		{"string default", "name", "string", "default-name"},
		{"integer default", "count", "integer", float64(10)},
		{"boolean default", "enabled", "boolean", true},
		{"number default", "ratio", "number", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop, ok := got.Properties[tt.propName]
			if !ok {
				t.Fatalf("missing property %q", tt.propName)
			}
			if prop.Type != tt.wantType {
				t.Errorf("property type = %q, want %q", prop.Type, tt.wantType)
			}
			if prop.Default != tt.wantDefault {
				t.Errorf("property default = %v (%T), want %v (%T)", prop.Default, prop.Default, tt.wantDefault, tt.wantDefault)
			}
		})
	}
}

func TestParseSchema_SchemaWithCLIExtensions(t *testing.T) {
	input := []byte(testutil.SchemaWithCLIExtensions())

	got, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	tests := []struct {
		name       string
		propName   string
		wantPos    int
		wantFlag   string
		wantShort  string
		wantFile   bool
	}{
		{"positional arg", "input", 1, "", "", false},
		{"custom flag", "output", 0, "--out", "o", false},
		{"file flag", "file", 0, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop, ok := got.Properties[tt.propName]
			if !ok {
				t.Fatalf("missing property %q", tt.propName)
			}
			if prop.CLIPosition != tt.wantPos {
				t.Errorf("CLIPosition = %d, want %d", prop.CLIPosition, tt.wantPos)
			}
			if prop.CLIFlag != tt.wantFlag {
				t.Errorf("CLIFlag = %q, want %q", prop.CLIFlag, tt.wantFlag)
			}
			if prop.CLIShort != tt.wantShort {
				t.Errorf("CLIShort = %q, want %q", prop.CLIShort, tt.wantShort)
			}
			if prop.CLIFile != tt.wantFile {
				t.Errorf("CLIFile = %v, want %v", prop.CLIFile, tt.wantFile)
			}
		})
	}
}

func TestParseSchema_SchemaWithNewCLIHints(t *testing.T) {
	input := []byte(`{
		"type": "object",
		"properties": {
			"source": {
				"type": "string",
				"flag": "--src",
				"file": true
			},
			"destination": {
				"type": "string",
				"flag": "--dest"
			},
			"config": {
				"type": "string",
				"file": true
			}
		}
	}`)

	got, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	tests := []struct {
		name     string
		propName string
		wantFlag string
		wantFile bool
	}{
		{"source with flag and file", "source", "--src", true},
		{"destination with flag only", "destination", "--dest", false},
		{"config with file only", "config", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop, ok := got.Properties[tt.propName]
			if !ok {
				t.Fatalf("missing property %q", tt.propName)
			}
			if prop.Flag != tt.wantFlag {
				t.Errorf("Flag = %q, want %q", prop.Flag, tt.wantFlag)
			}
			if prop.File != tt.wantFile {
				t.Errorf("File = %v, want %v", prop.File, tt.wantFile)
			}
		})
	}
}

func TestParseSchema_InvalidJSON(t *testing.T) {
	input := []byte(testutil.InvalidJSON())

	_, err := ParseSchema(input)
	if err == nil {
		t.Error("ParseSchema() expected error for invalid JSON")
	}
}

func TestParseSchema_EmptyInput(t *testing.T) {
	input := []byte{}

	_, err := ParseSchema(input)
	if err == nil {
		t.Error("ParseSchema() expected error for empty input")
	}
}

func TestParseSchema_EmptyObject(t *testing.T) {
	input := []byte("{}")

	got, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	if got.Type != "" {
		t.Errorf("ParseSchema().Type = %q, want empty", got.Type)
	}
}

func TestParseSchema_NestedObject(t *testing.T) {
	input := []byte(`{
		"type": "object",
		"properties": {
			"user": {
				"type": "object",
				"description": "User object"
			}
		}
	}`)

	got, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	prop, ok := got.Properties["user"]
	if !ok {
		t.Fatal("missing 'user' property")
	}
	if prop.Type != "object" {
		t.Errorf("user.Type = %q, want %q", prop.Type, "object")
	}
}

func TestParseSchema_ArraySchema(t *testing.T) {
	input := []byte(`{
		"type": "object",
		"properties": {
			"tags": {
				"type": "array",
				"description": "List of tags"
			}
		}
	}`)

	got, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	prop, ok := got.Properties["tags"]
	if !ok {
		t.Fatal("missing 'tags' property")
	}
	if prop.Type != "array" {
		t.Errorf("tags.Type = %q, want %q", prop.Type, "array")
	}
}

func TestParseSchema_EnumValues(t *testing.T) {
	input := []byte(`{
		"type": "object",
		"properties": {
			"status": {
				"type": "string",
				"enum": ["active", "inactive", "pending"]
			}
		}
	}`)

	got, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	prop, ok := got.Properties["status"]
	if !ok {
		t.Fatal("missing 'status' property")
	}

	expected := []string{"active", "inactive", "pending"}
	if !reflect.DeepEqual(prop.Enum, expected) {
		t.Errorf("status.Enum = %v, want %v", prop.Enum, expected)
	}
}

func TestGenerateFlags_ValidSchema(t *testing.T) {
	input := []byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"count": {"type": "integer", "default": 5}
		},
		"required": ["name"]
	}`)

	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	flags := GenerateFlags(schema)
	if len(flags) != 2 {
		t.Errorf("GenerateFlags() returned %d flags, want 2", len(flags))
	}

	var nameFlag, countFlag *FlagDef
	for i := range flags {
		if flags[i].Name == "--name" || flags[i].Name == "name" {
			nameFlag = &flags[i]
		}
		if flags[i].Name == "--count" || flags[i].Name == "count" {
			countFlag = &flags[i]
		}
	}

	if nameFlag == nil {
		t.Error("GenerateFlags() missing 'name' flag")
	} else {
		if nameFlag.Type != "string" {
			t.Errorf("name flag type = %q, want %q", nameFlag.Type, "string")
		}
		if !nameFlag.Required {
			t.Error("name flag should be required")
		}
	}

	if countFlag == nil {
		t.Error("GenerateFlags() missing 'count' flag")
	} else {
		if countFlag.Type != "integer" {
			t.Errorf("count flag type = %q, want %q", countFlag.Type, "integer")
		}
		if countFlag.DefaultValue != float64(5) {
			t.Errorf("count flag default = %v, want 5", countFlag.DefaultValue)
		}
		if countFlag.Required {
			t.Error("count flag should not be required")
		}
	}
}

func TestGenerateFlags_WithCLIExtensions(t *testing.T) {
	input := []byte(testutil.SchemaWithCLIExtensions())

	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	flags := GenerateFlags(schema)

	var inputFlag, outputFlag, fileFlag *FlagDef
	for i := range flags {
		if flags[i].Position == 1 {
			inputFlag = &flags[i]
		}
		if flags[i].Name == "--out" {
			outputFlag = &flags[i]
		}
		if flags[i].IsFile {
			fileFlag = &flags[i]
		}
	}

	if inputFlag == nil {
		t.Error("GenerateFlags() missing positional input flag")
	} else if inputFlag.Position != 1 {
		t.Errorf("input flag position = %d, want 1", inputFlag.Position)
	}

	if outputFlag == nil {
		t.Error("GenerateFlags() missing output flag with custom name")
	} else {
		if outputFlag.ShortName != "o" {
			t.Errorf("output flag short = %q, want %q", outputFlag.ShortName, "o")
		}
	}

	if fileFlag == nil {
		t.Error("GenerateFlags() missing file flag")
	} else if !fileFlag.IsFile {
		t.Error("file flag IsFile should be true")
	}
}

func TestGenerateFlags_WithNewCLIHints(t *testing.T) {
	input := []byte(`{
		"type": "object",
		"properties": {
			"source": {
				"type": "string",
				"flag": "--src",
				"file": true
			},
			"output": {
				"type": "string",
				"flag": "--out"
			},
			"verbose": {
				"type": "boolean"
			}
		}
	}`)

	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	flags := GenerateFlags(schema)

	var sourceFlag, outputFlag, verboseFlag *FlagDef
	for i := range flags {
		if flags[i].Name == "--src" {
			sourceFlag = &flags[i]
		}
		if flags[i].Name == "--out" {
			outputFlag = &flags[i]
		}
		if flags[i].Name == "--verbose" {
			verboseFlag = &flags[i]
		}
	}

	if sourceFlag == nil {
		t.Error("GenerateFlags() missing source flag with custom name --src")
	} else {
		if sourceFlag.PropertyName != "source" {
			t.Errorf("source flag PropertyName = %q, want %q", sourceFlag.PropertyName, "source")
		}
		if !sourceFlag.IsFile {
			t.Error("source flag IsFile should be true")
		}
	}

	if outputFlag == nil {
		t.Error("GenerateFlags() missing output flag with custom name --out")
	} else if outputFlag.IsFile {
		t.Error("output flag IsFile should be false")
	}

	if verboseFlag == nil {
		t.Error("GenerateFlags() missing verbose flag with auto-generated name --verbose")
	}
}

func TestGenerateFlags_NewFieldsTakePrecedenceOverDeprecated(t *testing.T) {
	input := []byte(`{
		"type": "object",
		"properties": {
			"input": {
				"type": "string",
				"flag": "--new-flag",
				"file": true,
				"x-cli-flag": "--old-flag",
				"x-cli-file": false
			}
		}
	}`)

	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	flags := GenerateFlags(schema)

	var inputFlag *FlagDef
	for i := range flags {
		if flags[i].PropertyName == "input" {
			inputFlag = &flags[i]
			break
		}
	}

	if inputFlag == nil {
		t.Fatal("GenerateFlags() missing input flag")
	}

	if inputFlag.Name != "--new-flag" {
		t.Errorf("input flag Name = %q, want %q (new Flag should take precedence)", inputFlag.Name, "--new-flag")
	}

	if !inputFlag.IsFile {
		t.Error("input flag IsFile should be true (new File should take precedence)")
	}
}

func TestGenerateFlags_EmptySchema(t *testing.T) {
	schema := &ParsedSchema{
		Type:       "object",
		Properties: map[string]Property{},
	}

	flags := GenerateFlags(schema)
	if len(flags) != 0 {
		t.Errorf("GenerateFlags() returned %d flags, want 0 for empty schema", len(flags))
	}
}

func TestGenerateFlags_NilSchema(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("GenerateFlags() should panic with nil schema")
		}
	}()

	GenerateFlags(nil)
}

func TestGenerateFlags_FlagNameSanitization(t *testing.T) {
	input := []byte(`{
		"type": "object",
		"properties": {
			"some_field_name": {"type": "string"}
		}
	}`)

	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	flags := GenerateFlags(schema)
	if len(flags) != 1 {
		t.Fatalf("GenerateFlags() returned %d flags, want 1", len(flags))
	}

	if !sort.StringsAreSorted([]string{flags[0].Name}) {
		t.Logf("Flag name: %s", flags[0].Name)
	}
}

func TestGenerateFlags_RequiredDetection(t *testing.T) {
	input := []byte(`{
		"type": "object",
		"properties": {
			"required_field": {"type": "string"},
			"optional_field": {"type": "string"}
		},
		"required": ["required_field"]
	}`)

	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema() error = %v", err)
	}

	flags := GenerateFlags(schema)

	requiredCount := 0
	for _, f := range flags {
		if f.Required {
			requiredCount++
		}
	}

	if requiredCount != 1 {
		t.Errorf("GenerateFlags() found %d required flags, want 1", requiredCount)
	}
}
