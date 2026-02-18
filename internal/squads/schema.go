// Package squads provides squad management for agent team coordination.
package squads

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"charm.land/fantasy/schema"

	"github.com/alexcabrera/ayo/internal/paths"
)

// SquadSchemas holds input and output JSON schemas for a squad.
// Both fields are optional - nil means free-form input/output is allowed.
type SquadSchemas struct {
	// Input is the JSON schema for validating squad input.
	// If nil, any input is accepted.
	Input *schema.Schema

	// Output is the JSON schema for validating squad output.
	// If nil, any output format is accepted.
	Output *schema.Schema
}

// HasInputSchema returns true if the squad has an input schema defined.
func (s *SquadSchemas) HasInputSchema() bool {
	return s != nil && s.Input != nil
}

// HasOutputSchema returns true if the squad has an output schema defined.
func (s *SquadSchemas) HasOutputSchema() bool {
	return s != nil && s.Output != nil
}

// LoadSquadSchemas loads the input.jsonschema and output.jsonschema files
// from the squad directory. Returns a SquadSchemas struct with nil fields
// for any missing schema files. Returns an error only if a schema file
// exists but is invalid JSON.
func LoadSquadSchemas(squadName string) (*SquadSchemas, error) {
	squadDir := paths.SquadDir(squadName)
	return loadSchemasFromDir(squadDir)
}

// loadSchemasFromDir loads schemas from a specific directory.
// This is separated for testability.
func loadSchemasFromDir(dir string) (*SquadSchemas, error) {
	schemas := &SquadSchemas{}

	// Load input schema
	inputSchema, err := loadSchemaFile(filepath.Join(dir, "input.jsonschema"))
	if err != nil {
		return nil, err
	}
	schemas.Input = inputSchema

	// Load output schema
	outputSchema, err := loadSchemaFile(filepath.Join(dir, "output.jsonschema"))
	if err != nil {
		return nil, err
	}
	schemas.Output = outputSchema

	return schemas, nil
}

// loadSchemaFile loads a JSON schema from a file path.
// Returns nil if the file doesn't exist, error if file exists but is invalid.
func loadSchemaFile(path string) (*schema.Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil // No schema file is fine
		}
		return nil, err
	}

	var s schema.Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, &SchemaParseError{Path: path, Err: err}
	}
	return &s, nil
}

// SchemaParseError is returned when a schema file exists but contains invalid JSON.
type SchemaParseError struct {
	Path string
	Err  error
}

func (e *SchemaParseError) Error() string {
	return "parse schema " + e.Path + ": " + e.Err.Error()
}

func (e *SchemaParseError) Unwrap() error {
	return e.Err
}
