package flows

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kaptinlin/jsonschema"
)

// SchemaValidationError represents a schema validation failure.
type SchemaValidationError struct {
	Message string
	Details []string
}

func (e *SchemaValidationError) Error() string {
	if len(e.Details) == 0 {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Details)
}

// ValidateInput validates the input JSON against the flow's input schema.
// Returns nil if no schema is defined or validation passes.
func ValidateInput(flow *Flow, input string) error {
	if !flow.HasInputSchema() {
		return nil
	}

	schemaData, err := os.ReadFile(flow.InputSchemaPath)
	if err != nil {
		return fmt.Errorf("read input schema: %w", err)
	}

	return validateJSON(input, schemaData, "input")
}

// ValidateOutput validates the output JSON against the flow's output schema.
// Returns warnings, not errors - output is still valid even if schema check fails.
func ValidateOutput(flow *Flow, output string) []string {
	if !flow.HasOutputSchema() {
		return nil
	}

	schemaData, err := os.ReadFile(flow.OutputSchemaPath)
	if err != nil {
		return []string{fmt.Sprintf("read output schema: %v", err)}
	}

	if err := validateJSON(output, schemaData, "output"); err != nil {
		if schemaErr, ok := err.(*SchemaValidationError); ok {
			return schemaErr.Details
		}
		return []string{err.Error()}
	}

	return nil
}

// validateJSON validates a JSON string against a JSON schema.
func validateJSON(jsonStr string, schemaData []byte, context string) error {
	// Parse the schema
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(schemaData)
	if err != nil {
		return fmt.Errorf("compile %s schema: %w", context, err)
	}

	// Parse the JSON
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return &SchemaValidationError{
			Message: fmt.Sprintf("invalid %s JSON", context),
			Details: []string{err.Error()},
		}
	}

	// Validate
	result := schema.Validate(data)
	if !result.IsValid() {
		var details []string
		for _, detail := range result.Errors {
			details = append(details, detail.Message)
		}
		return &SchemaValidationError{
			Message: fmt.Sprintf("%s validation failed", context),
			Details: details,
		}
	}

	return nil
}
