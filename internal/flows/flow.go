// Package flows provides flow discovery, parsing, and execution.
package flows

// FlowSource indicates where a flow was discovered from.
type FlowSource string

const (
	FlowSourceBuiltin FlowSource = "built-in"
	FlowSourceUser    FlowSource = "user"
	FlowSourceProject FlowSource = "project"
)

// FlowMetadata contains optional metadata fields.
type FlowMetadata struct {
	Version string
	Author  string
}

// FlowRaw contains the raw parsed content of a flow file.
type FlowRaw struct {
	Frontmatter map[string]string
	Script      string
}

// Flow represents a discovered flow.
type Flow struct {
	Name        string
	Description string
	Path        string     // Absolute path to flow.sh or name.sh
	Dir         string     // Parent directory
	Source      FlowSource // Where the flow was discovered

	// Optional schemas (nil if not present)
	InputSchemaPath  string // Path to input.jsonschema
	OutputSchemaPath string // Path to output.jsonschema

	// Metadata
	Metadata FlowMetadata

	// Parsed content
	Raw FlowRaw
}

// HasInputSchema returns true if the flow has an input schema defined.
func (f *Flow) HasInputSchema() bool {
	return f.InputSchemaPath != ""
}

// HasOutputSchema returns true if the flow has an output schema defined.
func (f *Flow) HasOutputSchema() bool {
	return f.OutputSchemaPath != ""
}
