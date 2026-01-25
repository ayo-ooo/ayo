package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ToolDefinition describes an external tool provided by a plugin.
// This is stored in tools/<tool-name>/tool.json.
type ToolDefinition struct {
	// Name is the tool identifier used in agent configs.
	Name string `json:"name"`

	// Description briefly describes what the tool does.
	Description string `json:"description"`

	// Command is the executable to run.
	// Can be a binary name (looked up in PATH) or absolute path.
	Command string `json:"command"`

	// Args are default arguments passed to the command.
	// Use {{param}} placeholders for parameter substitution.
	Args []string `json:"args,omitempty"`

	// Parameters defines the input schema for the tool.
	Parameters []ToolParameter `json:"parameters,omitempty"`

	// Timeout is the default timeout in seconds (0 = no timeout).
	Timeout int `json:"timeout,omitempty"`

	// WorkingDir specifies the working directory behavior.
	// "inherit" (default) = use caller's cwd
	// "plugin" = use plugin directory
	// "param" = use working_dir parameter
	WorkingDir string `json:"working_dir,omitempty"`

	// AllowAnyDir allows the working_dir parameter to be any directory,
	// not just subdirectories of the base directory. Use with caution -
	// only for trusted tools that manage their own security.
	AllowAnyDir bool `json:"allow_any_dir,omitempty"`

	// Quiet suppresses command output in the UI.
	Quiet bool `json:"quiet,omitempty"`

	// StreamOutput streams output as it's produced.
	StreamOutput bool `json:"stream_output,omitempty"`

	// Env sets additional environment variables.
	Env map[string]string `json:"env,omitempty"`

	// DependsOn lists binaries that must be available.
	DependsOn []string `json:"depends_on,omitempty"`

	// SpinnerStyle controls the spinner animation during tool execution.
	// "default" (or empty) = standard tool spinner (dots)
	// "crush" = fancy scrambling hex/symbol animation (for coding tools)
	// "none" = no spinner (tool manages its own output)
	SpinnerStyle string `json:"spinner_style,omitempty"`
}

// ToolParameter defines a parameter for an external tool.
type ToolParameter struct {
	// Name is the parameter identifier.
	Name string `json:"name"`

	// Description describes the parameter for the LLM.
	Description string `json:"description"`

	// Type is the JSON schema type (string, number, boolean, array, object).
	Type string `json:"type"`

	// Required indicates if the parameter must be provided.
	Required bool `json:"required,omitempty"`

	// Default is the default value if not provided.
	Default any `json:"default,omitempty"`

	// Enum lists allowed values for string parameters.
	Enum []string `json:"enum,omitempty"`

	// Items describes array element type (for type=array).
	Items *ToolParameter `json:"items,omitempty"`

	// ArgTemplate is how this param maps to command args.
	// Examples: "--flag={{value}}", "{{value}}", "--flag", "{{value}}"
	// If empty, uses "--name={{value}}" for non-boolean, "--name" for boolean.
	ArgTemplate string `json:"arg_template,omitempty"`

	// Position is for positional args (0 = first positional).
	// If set, ArgTemplate should be "{{value}}" or similar.
	Position *int `json:"position,omitempty"`

	// OmitIfEmpty skips the arg if value is empty/false/nil.
	OmitIfEmpty bool `json:"omit_if_empty,omitempty"`
}

// ToolFile is the expected filename for tool definitions.
const ToolFile = "tool.json"

// Tool definition errors
var (
	ErrToolDefNotFound      = errors.New("tool.json not found")
	ErrInvalidToolDef       = errors.New("invalid tool definition")
	ErrMissingToolName      = errors.New("tool: name is required")
	ErrMissingToolDesc      = errors.New("tool: description is required")
	ErrMissingToolCommand   = errors.New("tool: command is required")
	ErrMissingParamName     = errors.New("tool parameter: name is required")
	ErrMissingParamDesc     = errors.New("tool parameter: description is required")
	ErrMissingParamType     = errors.New("tool parameter: type is required")
	ErrInvalidParamType     = errors.New("tool parameter: invalid type")
)

// ValidParamTypes are the allowed parameter types.
var ValidParamTypes = map[string]bool{
	"string":  true,
	"number":  true,
	"integer": true,
	"boolean": true,
	"array":   true,
	"object":  true,
}

// LoadToolDefinition reads and validates a tool definition from a plugin.
func LoadToolDefinition(pluginDir, toolName string) (*ToolDefinition, error) {
	toolPath := filepath.Join(pluginDir, "tools", toolName, ToolFile)

	data, err := os.ReadFile(toolPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrToolDefNotFound, toolName)
		}
		return nil, fmt.Errorf("read tool definition: %w", err)
	}

	var td ToolDefinition
	if err := json.Unmarshal(data, &td); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToolDef, err)
	}

	if err := td.Validate(); err != nil {
		return nil, err
	}

	return &td, nil
}

// LoadAllToolDefinitions loads all tool definitions from a plugin.
func LoadAllToolDefinitions(pluginDir string) ([]*ToolDefinition, error) {
	toolsDir := filepath.Join(pluginDir, "tools")

	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No tools directory is fine
		}
		return nil, fmt.Errorf("read tools dir: %w", err)
	}

	var tools []*ToolDefinition
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		td, err := LoadToolDefinition(pluginDir, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("load tool %s: %w", entry.Name(), err)
		}
		tools = append(tools, td)
	}

	return tools, nil
}

// Validate checks that the tool definition has all required fields.
func (td *ToolDefinition) Validate() error {
	if td.Name == "" {
		return ErrMissingToolName
	}
	if td.Description == "" {
		return ErrMissingToolDesc
	}
	if td.Command == "" {
		return ErrMissingToolCommand
	}

	for i, param := range td.Parameters {
		if err := param.Validate(); err != nil {
			return fmt.Errorf("parameter %d: %w", i, err)
		}
	}

	return nil
}

// Validate checks that a tool parameter is valid.
func (p *ToolParameter) Validate() error {
	if p.Name == "" {
		return ErrMissingParamName
	}
	if p.Description == "" {
		return ErrMissingParamDesc
	}
	if p.Type == "" {
		return ErrMissingParamType
	}
	if !ValidParamTypes[p.Type] {
		return fmt.Errorf("%w: %s", ErrInvalidParamType, p.Type)
	}

	// Validate nested items for arrays
	if p.Type == "array" && p.Items != nil {
		if err := p.Items.Validate(); err != nil {
			return fmt.Errorf("items: %w", err)
		}
	}

	return nil
}

// ToParameters converts the tool definition to a Fantasy-compatible parameters map.
// This returns just the properties map (not a full JSON schema), which is what
// ToolInfo.Parameters expects. The required fields go in ToolInfo.Required separately.
func (td *ToolDefinition) ToParameters() map[string]any {
	properties := make(map[string]any)

	for _, param := range td.Parameters {
		properties[param.Name] = param.ToSchemaProperty()
	}

	return properties
}

// ToSchemaProperty converts a parameter to a JSON schema property.
func (p *ToolParameter) ToSchemaProperty() map[string]any {
	prop := map[string]any{
		"type":        p.Type,
		"description": p.Description,
	}

	if p.Default != nil {
		prop["default"] = p.Default
	}

	if len(p.Enum) > 0 {
		prop["enum"] = p.Enum
	}

	if p.Type == "array" && p.Items != nil {
		prop["items"] = p.Items.ToSchemaProperty()
	}

	return prop
}

// GetRequiredParams returns a list of required parameter names.
func (td *ToolDefinition) GetRequiredParams() []string {
	var required []string
	for _, param := range td.Parameters {
		if param.Required {
			required = append(required, param.Name)
		}
	}
	return required
}

// GetParamByName returns a parameter by name.
func (td *ToolDefinition) GetParamByName(name string) *ToolParameter {
	for i := range td.Parameters {
		if td.Parameters[i].Name == name {
			return &td.Parameters[i]
		}
	}
	return nil
}
