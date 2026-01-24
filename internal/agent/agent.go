package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"charm.land/fantasy/schema"

	"github.com/alexcabrera/ayo/internal/builtin"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/delegates"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/skills"
)

// Schema is a type alias for the fantasy schema type.
type Schema = schema.Schema

type Config struct {
	Model       string   `json:"model"`
	SystemFile  string   `json:"system_file"`
	Description string   `json:"description,omitempty"`
	AllowedTools []string `json:"allowed_tools,omitempty"`
	
	// System prompt configuration
	NoSystemWrapper bool `json:"no_system_wrapper,omitempty"` // Skip prefix/suffix wrapping
	
	// Skill configuration
	Skills            []string `json:"skills,omitempty"`             // Explicit include list
	ExcludeSkills     []string `json:"exclude_skills,omitempty"`     // Explicit exclude list
	IgnoreBuiltinSkills bool     `json:"ignore_builtin_skills,omitempty"`
	IgnoreSharedSkills  bool     `json:"ignore_shared_skills,omitempty"`

	// Delegation configuration
	// Maps task types (e.g., "coding", "research") to agent handles (e.g., "@crush")
	Delegates map[string]string `json:"delegates,omitempty"`
}

type Agent struct {
	Handle          string
	Dir             string
	Model           string
	System          string
	CombinedSystem  string
	Skills          []skills.Metadata
	SkillsWarnings  []string
	SkillsPrompt    string
	ToolsPrompt     string
	DelegateContext string // XML block with configured delegates
	Config          Config
	BuiltIn         bool
	InputSchema     *schema.Schema // JSON schema for input validation (optional)
	OutputSchema    *schema.Schema // JSON schema for output formatting (optional)
}

func NormalizeHandle(handle string) string {
	if strings.HasPrefix(handle, "@") {
		return handle
	}
	return "@" + handle
}

func ListHandles(cfg config.Config) ([]string, error) {
	handleSet := make(map[string]struct{})

	// Add built-in agents first
	for _, h := range builtin.ListAgents() {
		handleSet[h] = struct{}{}
	}

	// Scan all agent directories in priority order (local, user config, user data)
	for _, dir := range paths.AgentsDirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "@") {
				handleSet[entry.Name()] = struct{}{}
			}
		}
	}

	// Also check cfg.AgentsDir (may be custom location)
	if cfg.AgentsDir != "" {
		entries, err := os.ReadDir(cfg.AgentsDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), "@") {
					handleSet[entry.Name()] = struct{}{}
				}
			}
		}
	}

	// Add plugin agents
	for _, dir := range paths.AllPluginAgentsDirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "@") {
				handleSet[entry.Name()] = struct{}{}
			}
		}
	}

	// Convert to sorted slice
	handles := make([]string, 0, len(handleSet))
	for h := range handleSet {
		handles = append(handles, h)
	}
	sort.Strings(handles)
	return handles, nil
}

func Load(cfg config.Config, handle string) (Agent, error) {
	normalized := NormalizeHandle(handle)

	// Build directory list in priority order:
	// 1. ./.config/ayo/agents (local project)
	// 2. ./.local/share/ayo/agents (local project data)
	// 3. cfg.AgentsDir (user config, typically ~/.config/ayo/agents)
	// 4. ~/.local/share/ayo/agents (user data / built-in)
	// 5. Plugin agent directories
	dirs := paths.AgentsDirs()

	// Add cfg.AgentsDir if not already included
	cfgAgentsDirIncluded := false
	for _, d := range dirs {
		if d == cfg.AgentsDir {
			cfgAgentsDirIncluded = true
			break
		}
	}
	if !cfgAgentsDirIncluded && cfg.AgentsDir != "" {
		// Insert cfg.AgentsDir before builtin dir
		builtinDir := builtin.InstallDir()
		var newDirs []string
		inserted := false
		for _, d := range dirs {
			if d == builtinDir && !inserted {
				newDirs = append(newDirs, cfg.AgentsDir)
				inserted = true
			}
			newDirs = append(newDirs, d)
		}
		if !inserted {
			newDirs = append(newDirs, cfg.AgentsDir)
		}
		dirs = newDirs
	}

	// Add plugin directories at the end (lower priority than user/builtin)
	dirs = append(dirs, paths.AllPluginAgentsDirs()...)

	// Build set of data directories where builtins live
	builtinDirs := paths.DataDirs()

	// Try each directory
	for _, dir := range dirs {
		isBuiltIn := isDataDir(dir, builtinDirs)
		agent, err := loadFromDir(cfg, normalized, dir, isBuiltIn)
		if err == nil {
			return agent, nil
		}
		if !errors.Is(err, os.ErrNotExist) && !strings.Contains(err.Error(), "no such file") {
			return Agent{}, err
		}
	}

	// If not found in any directory, check if it's a built-in that needs installing
	if builtin.HasAgent(normalized) {
		if installErr := builtin.Install(); installErr != nil {
			return Agent{}, fmt.Errorf("install built-in agents: %w", installErr)
		}
		return loadFromDir(cfg, normalized, builtin.InstallDir(), true)
	}

	return Agent{}, fmt.Errorf("agent not found: %s", normalized)
}

func loadFromDir(cfg config.Config, normalized string, baseDir string, isBuiltIn bool) (Agent, error) {
	var agent Agent
	// All agents use @ prefix in directory name (e.g., @ayo, @myagent)
	dir := filepath.Join(baseDir, normalized)
	info, err := os.Stat(dir)
	if err != nil {
		return agent, err
	}
	if !info.IsDir() {
		return agent, errors.New("agent path is not a directory")
	}

	agentConfig, err := loadAgentConfig(dir)
	if err != nil {
		return agent, err
	}

	systemPath := agentConfig.SystemFile
	if systemPath == "" {
		systemPath = filepath.Join(dir, "system.md")
	} else if !filepath.IsAbs(systemPath) {
		systemPath = filepath.Join(dir, systemPath)
	}
	systemBytes, err := os.ReadFile(systemPath)
	if err != nil {
		return agent, err
	}
	agentSystem := strings.TrimSpace(string(systemBytes))

	// Load prefix and suffix using priority-based lookup (unless NoSystemWrapper is set)
	var prefix, suffix string
	if !agentConfig.NoSystemWrapper {
		if cfg.SystemPrefix != "" {
			prefix = strings.TrimSpace(readOptional(cfg.SystemPrefix))
		} else if prefixPath := paths.FindPromptFile("system-prefix.md"); prefixPath != "" {
			prefix = strings.TrimSpace(readOptional(prefixPath))
		}
		if cfg.SystemSuffix != "" {
			suffix = strings.TrimSpace(readOptional(cfg.SystemSuffix))
		} else if suffixPath := paths.FindPromptFile("system-suffix.md"); suffixPath != "" {
			suffix = strings.TrimSpace(readOptional(suffixPath))
		}
	}

	// Build environment context block (placed at top of system prompt)
	envContext := buildEnvContext()

	combinedParts := make([]string, 0, 4)
	combinedParts = append(combinedParts, envContext)
	if prefix != "" {
		combinedParts = append(combinedParts, prefix)
	}
	combinedParts = append(combinedParts, agentSystem)
	if suffix != "" {
		combinedParts = append(combinedParts, suffix)
	}
	combined := strings.TrimSpace(strings.Join(combinedParts, "\n\n"))

	agentSkillsDir := filepath.Join(dir, "skills")
	discovery := skills.DiscoverForAgent(
		agentSkillsDir,
		paths.SkillsDirs(),
		skills.DiscoveryFilterConfig{
			IncludeSkills: agentConfig.Skills,
			ExcludeSkills: agentConfig.ExcludeSkills,
			IgnoreBuiltin: agentConfig.IgnoreBuiltinSkills,
			IgnoreShared:  agentConfig.IgnoreSharedSkills,
		},
	)
	skillsPrompt := buildSkillsPrompt(discovery.Skills)
	toolsPrompt := BuildToolsPrompt(agentConfig.AllowedTools)

	// Load input schema if present
	inputSchema, err := loadInputSchema(dir)
	if err != nil {
		return agent, fmt.Errorf("load input schema: %w", err)
	}

	// Load output schema if present
	outputSchema, err := loadOutputSchema(dir)
	if err != nil {
		return agent, fmt.Errorf("load output schema: %w", err)
	}

	// Build delegate context from all sources
	delegateContext := buildDelegateContext(cfg, agentConfig.Delegates)

	agent = Agent{
		Handle:          normalized,
		Dir:             dir,
		Model:           resolveModel(cfg, agentConfig),
		System:          agentSystem,
		CombinedSystem:  combined,
		Skills:          discovery.Skills,
		SkillsWarnings:  discovery.Warnings,
		SkillsPrompt:    skillsPrompt,
		ToolsPrompt:     toolsPrompt,
		DelegateContext: delegateContext,
		Config:          agentConfig,
		BuiltIn:         isBuiltIn,
		InputSchema:     inputSchema,
		OutputSchema:    outputSchema,
	}
	return agent, nil
}

func resolveModel(cfg config.Config, agentConfig Config) string {
	if agentConfig.Model != "" {
		return agentConfig.Model
	}
	return cfg.DefaultModel
}

func readOptional(path string) string {
	if path == "" {
		return ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

// buildDelegateContext builds an XML block with configured delegates.
// This tells the agent which other agents handle specific task types.
func buildDelegateContext(cfg config.Config, agentDelegates map[string]string) string {
	allDelegates := delegates.GetAllDelegates(agentDelegates, cfg)
	if len(allDelegates) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("<delegate_context>\n")
	b.WriteString("The following task types have configured delegate agents:\n\n")

	// Sort keys for consistent output
	keys := make([]string, 0, len(allDelegates))
	for k := range allDelegates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, taskType := range keys {
		agent := allDelegates[taskType]
		b.WriteString(fmt.Sprintf("- %s: %s\n", taskType, agent))
	}

	b.WriteString("\nUse agent_call with the appropriate agent for these task types.\n")
	b.WriteString("</delegate_context>")
	return b.String()
}

// buildEnvContext returns environment information for the system prompt.
// This is placed at the top so the model has immediate context about
// the runtime environment and current time.
func buildEnvContext() string {
	var b strings.Builder
	b.WriteString("<environment>\n")

	// Current datetime
	now := time.Now()
	b.WriteString(fmt.Sprintf("datetime: %s\n", now.Format("2006-01-02 15:04:05 MST")))

	// Platform info
	b.WriteString(fmt.Sprintf("os: %s\n", runtime.GOOS))
	b.WriteString(fmt.Sprintf("arch: %s\n", runtime.GOARCH))

	// Working directory
	if wd, err := os.Getwd(); err == nil {
		b.WriteString(fmt.Sprintf("cwd: %s\n", wd))
	}

	// Shell (from environment)
	if shell := os.Getenv("SHELL"); shell != "" {
		b.WriteString(fmt.Sprintf("shell: %s\n", shell))
	}

	// Home directory
	if home, err := os.UserHomeDir(); err == nil {
		b.WriteString(fmt.Sprintf("home: %s\n", home))
	}

	b.WriteString("</environment>")
	return b.String()
}

// DefaultAgent is the default agent handle used when no agent is specified.
const DefaultAgent = "@ayo"

// ErrReservedNamespace is returned when a user tries to create an agent with a reserved namespace prefix.
var ErrReservedNamespace = errors.New("agent name cannot use reserved 'ayo.' namespace")

// IsReservedNamespace checks if a handle uses a reserved namespace (e.g., ayo.).
func IsReservedNamespace(handle string) bool {
	name := strings.TrimPrefix(handle, "@")
	return strings.HasPrefix(name, "ayo.") || name == "ayo"
}

// isDataDir checks if dir is within one of the data directories (where builtins live).
// The dir parameter is an agents directory, e.g., "./.ayo/agents" or "~/.local/share/ayo/agents".
func isDataDir(dir string, dataDirs []string) bool {
	for _, dataDir := range dataDirs {
		agentsDir := filepath.Join(dataDir, "agents")
		if dir == agentsDir {
			return true
		}
	}
	return false
}

func Save(cfg config.Config, handle string, cfgData Config, systemMessage string) (Agent, error) {
	return SaveWithSchemas(cfg, handle, cfgData, systemMessage, "", "")
}

// SaveWithSchemas creates a new agent with optional input/output schema files.
// If inputSchemaFile or outputSchemaFile are provided, their contents are copied
// into the agent directory as input.jsonschema and output.jsonschema respectively.
func SaveWithSchemas(cfg config.Config, handle string, cfgData Config, systemMessage string, inputSchemaFile, outputSchemaFile string) (Agent, error) {
	normalized := NormalizeHandle(handle)

	// Prevent users from creating agents in reserved namespaces
	if IsReservedNamespace(normalized) {
		return Agent{}, ErrReservedNamespace
	}

	// User agents use @ prefix in directory name (e.g., @myagent)
	agentDir := filepath.Join(cfg.AgentsDir, normalized)
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		return Agent{}, err
	}

	systemPath := cfgData.SystemFile
	if systemPath == "" {
		systemPath = filepath.Join(agentDir, "system.md")
	} else if !filepath.IsAbs(systemPath) {
		systemPath = filepath.Join(agentDir, systemPath)
	}

	if err := os.WriteFile(systemPath, []byte(strings.TrimSpace(systemMessage)+"\n"), 0o644); err != nil {
		return Agent{}, err
	}

	configPath := filepath.Join(agentDir, "config.json")
	cfgBytes, err := json.MarshalIndent(cfgData, "", "  ")
	if err != nil {
		return Agent{}, fmt.Errorf("marshal agent config: %w", err)
	}
	if err := os.WriteFile(configPath, cfgBytes, 0o644); err != nil {
		return Agent{}, fmt.Errorf("write agent config: %w", err)
	}

	// Copy input schema if provided
	if inputSchemaFile != "" {
		if err := copySchemaFile(inputSchemaFile, agentDir, "input.jsonschema"); err != nil {
			return Agent{}, fmt.Errorf("copy input schema: %w", err)
		}
	}

	// Copy output schema if provided
	if outputSchemaFile != "" {
		if err := copySchemaFile(outputSchemaFile, agentDir, "output.jsonschema"); err != nil {
			return Agent{}, fmt.Errorf("copy output schema: %w", err)
		}
	}

	return Load(cfg, normalized)
}

// copySchemaFile reads a JSON schema from srcPath and writes it to destName in agentDir.
// It validates that the file contains valid JSON before copying.
func copySchemaFile(srcPath, agentDir, destName string) error {
	// Expand ~ to home directory
	if strings.HasPrefix(srcPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		srcPath = strings.Replace(srcPath, "~", home, 1)
	}

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read schema file: %w", err)
	}

	// Validate JSON
	var js json.RawMessage
	if err := json.Unmarshal(data, &js); err != nil {
		return fmt.Errorf("invalid JSON in schema file: %w", err)
	}

	destPath := filepath.Join(agentDir, destName)
	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		return fmt.Errorf("write schema file: %w", err)
	}

	return nil
}

func loadAgentConfig(dir string) (Config, error) {
	cfg := Config{}
	configPath := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// loadInputSchema loads the input.jsonschema file from the agent directory if it exists.
// Returns nil if no schema file is present.
func loadInputSchema(dir string) (*schema.Schema, error) {
	schemaPath := filepath.Join(dir, "input.jsonschema")
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil // No schema file is fine
		}
		return nil, err
	}

	var s schema.Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse input schema: %w", err)
	}
	return &s, nil
}

// InputValidationError is returned when input doesn't match the agent's input schema.
type InputValidationError struct {
	Input      string
	ParseError error
	Schema     *schema.Schema
}

func (e *InputValidationError) Error() string {
	var msg strings.Builder

	msg.WriteString("This agent requires structured JSON input.\n\n")

	// Check if it's a JSON parse error vs schema validation error
	if e.ParseError != nil && strings.Contains(e.ParseError.Error(), "input must be valid JSON") {
		msg.WriteString("Your input is not valid JSON.\n\n")
	} else if e.ParseError != nil {
		msg.WriteString(fmt.Sprintf("Validation error: %v\n\n", e.ParseError))
	}

	// Show expected schema structure
	if e.Schema != nil && e.Schema.Properties != nil {
		msg.WriteString("Expected format:\n")
		msg.WriteString("  {\n")

		// Collect property names for consistent ordering
		var propNames []string
		for name := range e.Schema.Properties {
			propNames = append(propNames, name)
		}
		sort.Strings(propNames)

		for i, name := range propNames {
			prop := e.Schema.Properties[name]
			required := ""
			for _, r := range e.Schema.Required {
				if r == name {
					required = " (required)"
					break
				}
			}

			// Determine example value based on type
			example := "..."
			switch prop.Type {
			case "string":
				if len(prop.Enum) > 0 {
					// Show enum values
					var enumStrs []string
					for _, v := range prop.Enum {
						enumStrs = append(enumStrs, fmt.Sprintf("%q", v))
					}
					example = enumStrs[0]
				} else {
					example = "\"...\""
				}
			case "number", "integer":
				example = "0"
			case "boolean":
				example = "true"
			case "array":
				example = "[...]"
			case "object":
				example = "{...}"
			}

			comma := ","
			if i == len(propNames)-1 {
				comma = ""
			}

			desc := ""
			if prop.Description != "" {
				desc = fmt.Sprintf("  // %s", prop.Description)
			}

			msg.WriteString(fmt.Sprintf("    %q: %s%s%s%s\n", name, example, comma, required, desc))
		}
		msg.WriteString("  }\n")
	}

	return msg.String()
}

// ValidateInput validates the given input against the agent's input schema.
// Returns nil if the agent has no input schema or if validation passes.
// Returns an error if the input is not valid JSON or doesn't match the schema.
func (a *Agent) ValidateInput(input string) error {
	if a.InputSchema == nil {
		return nil // No schema means no validation required
	}

	// Try to parse input as JSON
	var parsed any
	if err := json.Unmarshal([]byte(input), &parsed); err != nil {
		return &InputValidationError{
			Input:      input,
			ParseError: fmt.Errorf("input must be valid JSON: %w", err),
			Schema:     a.InputSchema,
		}
	}

	// Validate against schema
	if err := schema.ValidateAgainstSchema(parsed, *a.InputSchema); err != nil {
		return &InputValidationError{
			Input:      input,
			ParseError: err,
			Schema:     a.InputSchema,
		}
	}

	return nil
}

// HasInputSchema returns true if the agent has an input schema defined.
func (a *Agent) HasInputSchema() bool {
	return a.InputSchema != nil
}

// loadOutputSchema loads the output.jsonschema file from the agent directory if it exists.
// Returns nil if no schema file is present.
func loadOutputSchema(dir string) (*schema.Schema, error) {
	schemaPath := filepath.Join(dir, "output.jsonschema")
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil // No schema file is fine
		}
		return nil, err
	}

	var s schema.Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse output schema: %w", err)
	}
	return &s, nil
}

// HasOutputSchema returns true if the agent has an output schema defined.
func (a *Agent) HasOutputSchema() bool {
	return a.OutputSchema != nil
}

// ValidateOutput validates the given output against the agent's output schema.
// Returns nil if the agent has no output schema or if validation passes.
func (a *Agent) ValidateOutput(output string) error {
	if a.OutputSchema == nil {
		return nil
	}

	var parsed any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		return fmt.Errorf("output must be valid JSON: %w", err)
	}

	if err := schema.ValidateAgainstSchema(parsed, *a.OutputSchema); err != nil {
		return fmt.Errorf("output validation failed: %w", err)
	}

	return nil
}

// CompatibilityTier represents how well two agents are compatible for chaining.
type CompatibilityTier int

const (
	// CompatibilityNone means the agents are not compatible.
	CompatibilityNone CompatibilityTier = iota
	// CompatibilityFreeform means the target accepts any input (no schema).
	CompatibilityFreeform
	// CompatibilityStructural means the output contains all required fields of input.
	CompatibilityStructural
	// CompatibilityExact means the schemas are identical.
	CompatibilityExact
)

func (c CompatibilityTier) String() string {
	switch c {
	case CompatibilityExact:
		return "exact"
	case CompatibilityStructural:
		return "structural"
	case CompatibilityFreeform:
		return "freeform"
	default:
		return "none"
	}
}

// ChainableAgent contains an agent and its compatibility information.
type ChainableAgent struct {
	Agent         Agent
	Compatibility CompatibilityTier
}

// CanChainTo checks if this agent's output can be consumed by the target agent.
func (a *Agent) CanChainTo(target *Agent) CompatibilityTier {
	// Source must have output schema to chain
	if a.OutputSchema == nil {
		return CompatibilityNone
	}

	// Target with no input schema accepts anything (freeform)
	if target.InputSchema == nil {
		return CompatibilityFreeform
	}

	// Check if output schema is compatible with input schema
	return checkSchemaCompatibility(a.OutputSchema, target.InputSchema)
}

// CanChainFrom checks if this agent can receive input from the source agent.
func (a *Agent) CanChainFrom(source *Agent) CompatibilityTier {
	return source.CanChainTo(a)
}

// checkSchemaCompatibility checks if an output schema is compatible with an input schema.
// Returns the compatibility tier.
func checkSchemaCompatibility(output, input *schema.Schema) CompatibilityTier {
	if output == nil || input == nil {
		return CompatibilityNone
	}

	// Both must be objects for structural comparison
	if output.Type != "object" || input.Type != "object" {
		// For non-object types, require exact match
		if schemasEqual(output, input) {
			return CompatibilityExact
		}
		return CompatibilityNone
	}

	// Check if all required input fields are present in output
	if input.Properties == nil {
		return CompatibilityExact // No properties required
	}

	if output.Properties == nil {
		return CompatibilityNone // Input requires properties but output has none
	}

	// Check required fields
	for _, requiredField := range input.Required {
		outputProp, exists := output.Properties[requiredField]
		if !exists {
			return CompatibilityNone
		}

		inputProp := input.Properties[requiredField]
		if inputProp == nil {
			continue
		}

		// Basic type compatibility check
		if outputProp.Type != inputProp.Type {
			return CompatibilityNone
		}
	}

	// Check if schemas are exactly equal
	if schemasEqual(output, input) {
		return CompatibilityExact
	}

	return CompatibilityStructural
}

// schemasEqual checks if two schemas are identical.
func schemasEqual(a, b *schema.Schema) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare basic fields
	if a.Type != b.Type {
		return false
	}

	// Compare required fields
	if len(a.Required) != len(b.Required) {
		return false
	}
	reqA := make(map[string]bool)
	for _, r := range a.Required {
		reqA[r] = true
	}
	for _, r := range b.Required {
		if !reqA[r] {
			return false
		}
	}

	// Compare properties
	if len(a.Properties) != len(b.Properties) {
		return false
	}
	for name, propA := range a.Properties {
		propB, exists := b.Properties[name]
		if !exists {
			return false
		}
		if !schemasEqual(propA, propB) {
			return false
		}
	}

	return true
}

// FindDownstreamAgents finds all agents that can receive output from the given agent.
func FindDownstreamAgents(cfg config.Config, source Agent) ([]ChainableAgent, error) {
	if source.OutputSchema == nil {
		return nil, nil // No output schema, can't chain
	}

	handles, err := ListHandles(cfg)
	if err != nil {
		return nil, err
	}

	var compatible []ChainableAgent
	for _, handle := range handles {
		if handle == source.Handle {
			continue // Skip self
		}

		target, err := Load(cfg, handle)
		if err != nil {
			continue // Skip agents that fail to load
		}

		tier := source.CanChainTo(&target)
		if tier != CompatibilityNone {
			compatible = append(compatible, ChainableAgent{
				Agent:         target,
				Compatibility: tier,
			})
		}
	}

	// Sort by compatibility tier (exact > structural > freeform)
	sortChainableAgents(compatible)
	return compatible, nil
}

// FindUpstreamAgents finds all agents whose output this agent can receive.
func FindUpstreamAgents(cfg config.Config, target Agent) ([]ChainableAgent, error) {
	handles, err := ListHandles(cfg)
	if err != nil {
		return nil, err
	}

	var compatible []ChainableAgent
	for _, handle := range handles {
		if handle == target.Handle {
			continue // Skip self
		}

		source, err := Load(cfg, handle)
		if err != nil {
			continue
		}

		if source.OutputSchema == nil {
			continue // Source must have output schema
		}

		tier := source.CanChainTo(&target)
		if tier != CompatibilityNone {
			compatible = append(compatible, ChainableAgent{
				Agent:         source,
				Compatibility: tier,
			})
		}
	}

	sortChainableAgents(compatible)
	return compatible, nil
}

// sortChainableAgents sorts agents by compatibility tier (highest first).
func sortChainableAgents(agents []ChainableAgent) {
	sort.Slice(agents, func(i, j int) bool {
		// Higher tier comes first
		if agents[i].Compatibility != agents[j].Compatibility {
			return agents[i].Compatibility > agents[j].Compatibility
		}
		// Then alphabetically by handle
		return agents[i].Agent.Handle < agents[j].Agent.Handle
	})
}

// IsChainable returns true if the agent has either input or output schema.
func (a *Agent) IsChainable() bool {
	return a.InputSchema != nil || a.OutputSchema != nil
}

// ListChainableAgents returns all agents that have input or output schemas.
func ListChainableAgents(cfg config.Config) ([]Agent, error) {
	handles, err := ListHandles(cfg)
	if err != nil {
		return nil, err
	}

	var chainable []Agent
	for _, handle := range handles {
		ag, err := Load(cfg, handle)
		if err != nil {
			continue
		}
		if ag.IsChainable() {
			chainable = append(chainable, ag)
		}
	}
	return chainable, nil
}
