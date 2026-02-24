package squads

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSchemasFromDir_NoSchemas(t *testing.T) {
	dir := t.TempDir()

	schemas, err := loadSchemasFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schemas.Input != nil {
		t.Error("expected nil input schema")
	}
	if schemas.Output != nil {
		t.Error("expected nil output schema")
	}
	if schemas.HasInputSchema() {
		t.Error("HasInputSchema should return false")
	}
	if schemas.HasOutputSchema() {
		t.Error("HasOutputSchema should return false")
	}
}

func TestLoadSchemasFromDir_InputOnly(t *testing.T) {
	dir := t.TempDir()

	inputSchema := `{
		"type": "object",
		"properties": {
			"task": {"type": "string"}
		},
		"required": ["task"]
	}`

	if err := os.WriteFile(filepath.Join(dir, "input.jsonschema"), []byte(inputSchema), 0o644); err != nil {
		t.Fatal(err)
	}

	schemas, err := loadSchemasFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schemas.Input == nil {
		t.Fatal("expected input schema")
	}
	if schemas.Output != nil {
		t.Error("expected nil output schema")
	}
	if !schemas.HasInputSchema() {
		t.Error("HasInputSchema should return true")
	}
	if schemas.HasOutputSchema() {
		t.Error("HasOutputSchema should return false")
	}
}

func TestLoadSchemasFromDir_OutputOnly(t *testing.T) {
	dir := t.TempDir()

	outputSchema := `{
		"type": "object",
		"properties": {
			"result": {"type": "string"}
		},
		"required": ["result"]
	}`

	if err := os.WriteFile(filepath.Join(dir, "output.jsonschema"), []byte(outputSchema), 0o644); err != nil {
		t.Fatal(err)
	}

	schemas, err := loadSchemasFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schemas.Input != nil {
		t.Error("expected nil input schema")
	}
	if schemas.Output == nil {
		t.Fatal("expected output schema")
	}
	if schemas.HasInputSchema() {
		t.Error("HasInputSchema should return false")
	}
	if !schemas.HasOutputSchema() {
		t.Error("HasOutputSchema should return true")
	}
}

func TestLoadSchemasFromDir_BothSchemas(t *testing.T) {
	dir := t.TempDir()

	inputSchema := `{
		"type": "object",
		"properties": {
			"task": {"type": "string"}
		}
	}`

	outputSchema := `{
		"type": "object",
		"properties": {
			"result": {"type": "string"}
		}
	}`

	if err := os.WriteFile(filepath.Join(dir, "input.jsonschema"), []byte(inputSchema), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "output.jsonschema"), []byte(outputSchema), 0o644); err != nil {
		t.Fatal(err)
	}

	schemas, err := loadSchemasFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schemas.Input == nil {
		t.Fatal("expected input schema")
	}
	if schemas.Output == nil {
		t.Fatal("expected output schema")
	}
	if !schemas.HasInputSchema() {
		t.Error("HasInputSchema should return true")
	}
	if !schemas.HasOutputSchema() {
		t.Error("HasOutputSchema should return true")
	}
}

func TestLoadSchemasFromDir_InvalidInputSchema(t *testing.T) {
	dir := t.TempDir()

	// Invalid JSON
	if err := os.WriteFile(filepath.Join(dir, "input.jsonschema"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := loadSchemasFromDir(dir)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}

	// Check it's a SchemaParseError
	parseErr, ok := err.(*SchemaParseError)
	if !ok {
		t.Errorf("expected *SchemaParseError, got %T", err)
	}
	if parseErr.Path != filepath.Join(dir, "input.jsonschema") {
		t.Errorf("expected path %s, got %s", filepath.Join(dir, "input.jsonschema"), parseErr.Path)
	}
}

func TestLoadSchemasFromDir_InvalidOutputSchema(t *testing.T) {
	dir := t.TempDir()

	// Valid input schema
	inputSchema := `{"type": "object"}`
	if err := os.WriteFile(filepath.Join(dir, "input.jsonschema"), []byte(inputSchema), 0o644); err != nil {
		t.Fatal(err)
	}

	// Invalid output schema
	if err := os.WriteFile(filepath.Join(dir, "output.jsonschema"), []byte("{invalid}"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := loadSchemasFromDir(dir)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}

	// Check it's a SchemaParseError
	parseErr, ok := err.(*SchemaParseError)
	if !ok {
		t.Errorf("expected *SchemaParseError, got %T", err)
	}
	if parseErr.Path != filepath.Join(dir, "output.jsonschema") {
		t.Errorf("expected path %s, got %s", filepath.Join(dir, "output.jsonschema"), parseErr.Path)
	}
}

func TestSchemaParseError(t *testing.T) {
	err := &SchemaParseError{
		Path: "/path/to/schema.json",
		Err:  os.ErrNotExist,
	}

	// Test Error()
	expected := "parse schema /path/to/schema.json: file does not exist"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	// Test Unwrap()
	if err.Unwrap() != os.ErrNotExist {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), os.ErrNotExist)
	}
}

func TestSquadSchemas_NilReceiver(t *testing.T) {
	var schemas *SquadSchemas

	if schemas.HasInputSchema() {
		t.Error("nil SquadSchemas.HasInputSchema should return false")
	}
	if schemas.HasOutputSchema() {
		t.Error("nil SquadSchemas.HasOutputSchema should return false")
	}
}

func TestLoadSchemaFile_ReadError(t *testing.T) {
	// Test with a directory instead of a file
	dir := t.TempDir()
	schemaDir := filepath.Join(dir, "input.jsonschema")
	if err := os.MkdirAll(schemaDir, 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := loadSchemaFile(schemaDir)
	if err == nil {
		t.Error("expected error when reading directory as file")
	}
}

func TestLoadSchemasFromDir_NewSchemasPath(t *testing.T) {
	// Test that schemas/input.json and schemas/output.json are preferred
	dir := t.TempDir()

	// Create schemas subdirectory
	schemasDir := filepath.Join(dir, "schemas")
	if err := os.MkdirAll(schemasDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Use new path: schemas/input.json
	inputSchema := `{
		"type": "object",
		"properties": {
			"action": {"type": "string"}
		},
		"required": ["action"]
	}`
	outputSchema := `{
		"type": "object",
		"properties": {
			"status": {"type": "string"}
		}
	}`

	if err := os.WriteFile(filepath.Join(schemasDir, "input.json"), []byte(inputSchema), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(schemasDir, "output.json"), []byte(outputSchema), 0o644); err != nil {
		t.Fatal(err)
	}

	schemas, err := loadSchemasFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schemas.Input == nil {
		t.Fatal("expected input schema from schemas/input.json")
	}
	if schemas.Output == nil {
		t.Fatal("expected output schema from schemas/output.json")
	}
}

func TestLoadSchemasFromDir_NewPathOverridesLegacy(t *testing.T) {
	// New path takes precedence over legacy path
	dir := t.TempDir()

	// Create schemas subdirectory
	schemasDir := filepath.Join(dir, "schemas")
	if err := os.MkdirAll(schemasDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// New path schema with "new" in required
	newSchema := `{
		"type": "object",
		"required": ["new_field"]
	}`
	if err := os.WriteFile(filepath.Join(schemasDir, "input.json"), []byte(newSchema), 0o644); err != nil {
		t.Fatal(err)
	}

	// Legacy path schema with "legacy" in required
	legacySchema := `{
		"type": "object",
		"required": ["legacy_field"]
	}`
	if err := os.WriteFile(filepath.Join(dir, "input.jsonschema"), []byte(legacySchema), 0o644); err != nil {
		t.Fatal(err)
	}

	schemas, err := loadSchemasFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schemas.Input == nil {
		t.Fatal("expected input schema")
	}

	// Should use new path (has "new_field" in required)
	if len(schemas.Input.Required) != 1 || schemas.Input.Required[0] != "new_field" {
		t.Errorf("expected new path schema with required=[new_field], got %v", schemas.Input.Required)
	}
}

func TestLoadSchemasFromDir_FallbackToLegacy(t *testing.T) {
	// Falls back to legacy path when new path doesn't exist
	dir := t.TempDir()

	legacySchema := `{
		"type": "object",
		"required": ["legacy_field"]
	}`
	if err := os.WriteFile(filepath.Join(dir, "input.jsonschema"), []byte(legacySchema), 0o644); err != nil {
		t.Fatal(err)
	}

	schemas, err := loadSchemasFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schemas.Input == nil {
		t.Fatal("expected input schema from legacy path")
	}

	if len(schemas.Input.Required) != 1 || schemas.Input.Required[0] != "legacy_field" {
		t.Errorf("expected legacy schema with required=[legacy_field], got %v", schemas.Input.Required)
	}
}
