// Package plugins provides functionality for managing ayo plugins.
// Plugins are distributed via git repositories with the naming convention
// ayo-plugins-<name> and can contain agents, skills, tools, and providers.
package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PluginType categorizes what a plugin provides.
type PluginType string

const (
	// PluginTypeAgent indicates the plugin provides agents.
	PluginTypeAgent PluginType = "agent"
	// PluginTypeSkill indicates the plugin provides skills.
	PluginTypeSkill PluginType = "skill"
	// PluginTypeTool indicates the plugin provides tools.
	PluginTypeTool PluginType = "tool"
	// PluginTypeMemory indicates the plugin provides a memory provider.
	PluginTypeMemory PluginType = "memory"
	// PluginTypeSandbox indicates the plugin provides a sandbox provider.
	PluginTypeSandbox PluginType = "sandbox"
	// PluginTypeEmbedding indicates the plugin provides an embedding provider.
	PluginTypeEmbedding PluginType = "embedding"
	// PluginTypeObserver indicates the plugin provides an observer provider.
	PluginTypeObserver PluginType = "observer"
	// PluginTypePlanner indicates the plugin provides a planner.
	PluginTypePlanner PluginType = "planner"
)

// Manifest represents the plugin manifest (manifest.json).
// Every plugin must have a manifest at the root of its repository.
type Manifest struct {
	// Name is the plugin identifier (e.g., "crush").
	// Must match the repository name pattern: ayo-plugins-<name>
	Name string `json:"name"`

	// Version is the semantic version of the plugin (e.g., "1.0.0").
	Version string `json:"version"`

	// Description briefly describes what the plugin provides.
	Description string `json:"description"`

	// Author is the plugin author or organization.
	Author string `json:"author,omitempty"`

	// Repository is the git repository URL.
	Repository string `json:"repository,omitempty"`

	// License is the SPDX license identifier.
	License string `json:"license,omitempty"`

	// Agents lists the agent handles provided by this plugin.
	// These must exist in the agents/ directory.
	Agents []string `json:"agents,omitempty"`

	// Skills lists the shared skill names provided by this plugin.
	// These must exist in the skills/ directory.
	Skills []string `json:"skills,omitempty"`

	// Tools lists the tool names provided by this plugin.
	// These must exist in the tools/ directory.
	Tools []string `json:"tools,omitempty"`

	// Delegates declares task types this plugin's agents can handle.
	// On install, user is prompted to set these as global defaults.
	// Example: {"coding": "@crush"}
	Delegates map[string]string `json:"delegates,omitempty"`

	// DefaultTools declares tool type aliases this plugin's tools can provide.
	// On install, user is prompted to set these as default tool mappings.
	// Example: {"search": "searxng"}
	DefaultTools map[string]string `json:"default_tools,omitempty"`

	// Dependencies specifies external requirements.
	Dependencies *Dependencies `json:"dependencies,omitempty"`

	// PostInstall is an optional script to run after installation.
	// Path is relative to plugin root. Script receives plugin directory as first arg.
	// Example: "scripts/post-install.sh"
	PostInstall string `json:"post_install,omitempty"`

	// AyoVersion specifies the minimum ayo version required.
	// Uses semver constraints (e.g., ">=0.2.0").
	AyoVersion string `json:"ayo_version,omitempty"`

	// Providers lists provider implementations this plugin provides.
	// Each provider must have a unique name within its type (memory, sandbox, etc.).
	Providers []ProviderDef `json:"providers,omitempty"`

	// Planners lists planner implementations this plugin provides.
	// Each planner must have a unique name and specify its type (near or long).
	Planners []PlannerDef `json:"planners,omitempty"`
}

// ProviderDef describes a provider implementation in a plugin.
type ProviderDef struct {
	// Name is the unique identifier for this provider (e.g., "zettelkasten", "apple-container").
	Name string `json:"name"`

	// Type is the provider category. Must be one of: memory, sandbox, embedding, observer.
	Type PluginType `json:"type"`

	// Description briefly describes what this provider does.
	Description string `json:"description,omitempty"`

	// EntryPoint is the path to the provider implementation (Go plugin or binary).
	// For built-in providers, this may be empty.
	EntryPoint string `json:"entry_point,omitempty"`

	// Config contains provider-specific default configuration.
	// These values are merged with user config when the provider is activated.
	Config map[string]any `json:"config,omitempty"`
}

// PlannerType specifies the type of planner (near-term or long-term).
type PlannerType string

const (
	// PlannerTypeNear is for near-term planning (session-scoped todos).
	PlannerTypeNear PlannerType = "near"
	// PlannerTypeLong is for long-term planning (persistent tickets).
	PlannerTypeLong PlannerType = "long"
)

// PlannerDef describes a planner implementation in a plugin.
type PlannerDef struct {
	// Name is the unique identifier for this planner (e.g., "ayo-todos", "ayo-tickets").
	Name string `json:"name"`

	// Type specifies whether this is a near-term or long-term planner.
	// Must be "near" or "long".
	Type PlannerType `json:"type"`

	// Description briefly describes what this planner does.
	Description string `json:"description,omitempty"`

	// EntryPoint is the path to the planner implementation (Go plugin .so or binary).
	// For built-in planners, this may be empty.
	EntryPoint string `json:"entry_point,omitempty"`

	// Config contains planner-specific default configuration.
	Config map[string]any `json:"config,omitempty"`
}

// ValidPlannerType checks if a PlannerType is valid.
func ValidPlannerType(t PlannerType) bool {
	return t == PlannerTypeNear || t == PlannerTypeLong
}

// Dependencies specifies external requirements for a plugin.
type Dependencies struct {
	// Binaries lists executable names that must be in PATH.
	// Can be simple strings (just the binary name) or BinaryDep objects
	// with installation instructions.
	Binaries []BinaryDep `json:"-"` // Custom unmarshaling

	// Plugins lists other ayo plugins that must be installed.
	Plugins []string `json:"plugins,omitempty"`
}

// BinaryDep describes a binary dependency with optional installation instructions.
type BinaryDep struct {
	// Name is the executable name to look for in PATH.
	Name string `json:"name"`

	// InstallHint is a human-readable message explaining how to install.
	// Example: "Install with: brew install foo"
	InstallHint string `json:"install_hint,omitempty"`

	// InstallURL is a URL with installation instructions.
	InstallURL string `json:"install_url,omitempty"`

	// InstallCmd is a command that can be run to install the binary.
	// Example: "go install github.com/foo/bar@latest"
	InstallCmd string `json:"install_cmd,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Dependencies.
// It supports both simple string arrays and mixed string/object arrays:
//
//	"binaries": ["foo", "bar"]
//	"binaries": [{"name": "foo", "install_hint": "brew install foo"}, "bar"]
func (d *Dependencies) UnmarshalJSON(data []byte) error {
	// Use an alias to avoid infinite recursion
	type Alias Dependencies
	aux := &struct {
		Binaries []json.RawMessage `json:"binaries,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Parse binaries - each can be a string or an object
	for _, raw := range aux.Binaries {
		// Try as string first
		var name string
		if err := json.Unmarshal(raw, &name); err == nil {
			d.Binaries = append(d.Binaries, BinaryDep{Name: name})
			continue
		}

		// Try as object
		var dep BinaryDep
		if err := json.Unmarshal(raw, &dep); err != nil {
			return fmt.Errorf("invalid binary dependency: must be string or object: %w", err)
		}
		if dep.Name == "" {
			return errors.New("invalid binary dependency: name is required")
		}
		d.Binaries = append(d.Binaries, dep)
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling for Dependencies.
// It serializes simple dependencies (name only) as strings for cleaner output.
func (d Dependencies) MarshalJSON() ([]byte, error) {
	type Alias struct {
		Binaries []any    `json:"binaries,omitempty"`
		Plugins  []string `json:"plugins,omitempty"`
	}

	aux := Alias{
		Plugins: d.Plugins,
	}

	for _, dep := range d.Binaries {
		if dep.InstallHint == "" && dep.InstallURL == "" && dep.InstallCmd == "" {
			// Simple dependency - just the name
			aux.Binaries = append(aux.Binaries, dep.Name)
		} else {
			// Full dependency object
			aux.Binaries = append(aux.Binaries, dep)
		}
	}

	return json.Marshal(aux)
}

// GetBinaryNames returns just the binary names for backwards compatibility.
func (d *Dependencies) GetBinaryNames() []string {
	if d == nil {
		return nil
	}
	names := make([]string, len(d.Binaries))
	for i, b := range d.Binaries {
		names[i] = b.Name
	}
	return names
}

// ManifestFile is the expected filename for plugin manifests.
const ManifestFile = "manifest.json"

// PluginPrefix is the required prefix for plugin repository names.
const PluginPrefix = "ayo-plugins-"

// Validation errors
var (
	ErrManifestNotFound      = errors.New("manifest.json not found")
	ErrInvalidManifest       = errors.New("invalid manifest")
	ErrMissingName           = errors.New("manifest: name is required")
	ErrMissingVersion        = errors.New("manifest: version is required")
	ErrMissingDescription    = errors.New("manifest: description is required")
	ErrInvalidName           = errors.New("manifest: name must be lowercase alphanumeric with hyphens")
	ErrInvalidVersion        = errors.New("manifest: version must be valid semver (e.g., 1.0.0)")
	ErrAgentNotFound         = errors.New("manifest: declared agent not found in agents/ directory")
	ErrSkillNotFound         = errors.New("manifest: declared skill not found in skills/ directory")
	ErrToolNotFound          = errors.New("manifest: declared tool not found in tools/ directory")
	ErrInvalidPluginRef      = errors.New("invalid plugin reference: must be a full git URL (https:// or git@)")
	ErrInvalidProviderType   = errors.New("manifest: provider type must be memory, sandbox, embedding, or observer")
	ErrMissingProviderName   = errors.New("manifest: provider name is required")
	ErrDuplicateProviderName = errors.New("manifest: duplicate provider name")
	ErrInvalidPlannerType    = errors.New("manifest: planner type must be 'near' or 'long'")
	ErrMissingPlannerName    = errors.New("manifest: planner name is required")
	ErrDuplicatePlannerName  = errors.New("manifest: duplicate planner name")
)

// namePattern validates plugin names: lowercase letters, numbers, hyphens.
var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$|^[a-z]$`)

// versionPattern validates semver versions.
var versionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$`)

// LoadManifest reads and validates a manifest from the given plugin directory.
func LoadManifest(pluginDir string) (*Manifest, error) {
	manifestPath := filepath.Join(pluginDir, ManifestFile)

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrManifestNotFound
		}
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidManifest, err)
	}

	if err := m.Validate(); err != nil {
		return nil, err
	}

	if err := m.ValidateContents(pluginDir); err != nil {
		return nil, err
	}

	return &m, nil
}

// Validate checks that the manifest has all required fields with valid values.
func (m *Manifest) Validate() error {
	if m.Name == "" {
		return ErrMissingName
	}
	if !namePattern.MatchString(m.Name) {
		return fmt.Errorf("%w: got %q", ErrInvalidName, m.Name)
	}

	if m.Version == "" {
		return ErrMissingVersion
	}
	if !versionPattern.MatchString(m.Version) {
		return fmt.Errorf("%w: got %q", ErrInvalidVersion, m.Version)
	}

	if m.Description == "" {
		return ErrMissingDescription
	}

	// Validate providers
	if err := m.validateProviders(); err != nil {
		return err
	}

	// Validate planners
	if err := m.validatePlanners(); err != nil {
		return err
	}

	return nil
}

// validateProviders checks that provider definitions are valid.
func (m *Manifest) validateProviders() error {
	seen := make(map[string]bool)
	validTypes := map[PluginType]bool{
		PluginTypeMemory:    true,
		PluginTypeSandbox:   true,
		PluginTypeEmbedding: true,
		PluginTypeObserver:  true,
	}

	for i, p := range m.Providers {
		if p.Name == "" {
			return fmt.Errorf("%w (provider %d)", ErrMissingProviderName, i)
		}

		if !validTypes[p.Type] {
			return fmt.Errorf("%w: got %q for provider %q", ErrInvalidProviderType, p.Type, p.Name)
		}

		// Check for duplicates within the same type
		key := string(p.Type) + ":" + p.Name
		if seen[key] {
			return fmt.Errorf("%w: %s/%s", ErrDuplicateProviderName, p.Type, p.Name)
		}
		seen[key] = true
	}

	return nil
}

// validatePlanners checks that planner definitions are valid.
func (m *Manifest) validatePlanners() error {
	seen := make(map[string]bool)

	for i, p := range m.Planners {
		if p.Name == "" {
			return fmt.Errorf("%w (planner %d)", ErrMissingPlannerName, i)
		}

		if !ValidPlannerType(p.Type) {
			return fmt.Errorf("%w: got %q for planner %q", ErrInvalidPlannerType, p.Type, p.Name)
		}

		// Check for duplicates within the same type
		key := string(p.Type) + ":" + p.Name
		if seen[key] {
			return fmt.Errorf("%w: %s/%s", ErrDuplicatePlannerName, p.Type, p.Name)
		}
		seen[key] = true
	}

	return nil
}

// ValidateContents checks that declared agents, skills, and tools exist.
func (m *Manifest) ValidateContents(pluginDir string) error {
	// Check agents
	for _, agent := range m.Agents {
		agentDir := filepath.Join(pluginDir, "agents", agent)
		if _, err := os.Stat(agentDir); os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrAgentNotFound, agent)
		}
	}

	// Check skills
	for _, skill := range m.Skills {
		skillDir := filepath.Join(pluginDir, "skills", skill)
		if _, err := os.Stat(skillDir); os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrSkillNotFound, skill)
		}
	}

	// Check tools
	for _, tool := range m.Tools {
		toolDir := filepath.Join(pluginDir, "tools", tool)
		if _, err := os.Stat(toolDir); os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrToolNotFound, tool)
		}
	}

	return nil
}

// ParsePluginURL parses a plugin reference into a git URL and plugin name.
// Only full git URLs are accepted (https:// or git@).
// The plugin name is extracted from the repository name by removing the
// ayo-plugins- prefix if present.
func ParsePluginURL(ref string) (gitURL string, name string, err error) {
	ref = strings.TrimSpace(ref)

	// Only accept full URLs
	if !strings.HasPrefix(ref, "https://") && !strings.HasPrefix(ref, "git@") {
		return "", "", ErrInvalidPluginRef
	}

	gitURL = ref
	// Extract name from URL
	parts := strings.Split(strings.TrimSuffix(ref, ".git"), "/")
	repoName := parts[len(parts)-1]
	if strings.HasPrefix(repoName, PluginPrefix) {
		name = strings.TrimPrefix(repoName, PluginPrefix)
	} else {
		name = repoName
	}
	return gitURL, name, nil
}

// ExtractNameFromRepo extracts the plugin name from a repository name.
// Returns the name without the ayo-plugins- prefix.
func ExtractNameFromRepo(repoName string) string {
	repoName = strings.TrimSuffix(repoName, ".git")
	if strings.HasPrefix(repoName, PluginPrefix) {
		return strings.TrimPrefix(repoName, PluginPrefix)
	}
	return repoName
}

// Types returns the list of content types this plugin provides.
// This is useful for displaying what a plugin offers.
func (m *Manifest) Types() []PluginType {
	var types []PluginType
	seen := make(map[PluginType]bool)

	if len(m.Agents) > 0 && !seen[PluginTypeAgent] {
		types = append(types, PluginTypeAgent)
		seen[PluginTypeAgent] = true
	}
	if len(m.Skills) > 0 && !seen[PluginTypeSkill] {
		types = append(types, PluginTypeSkill)
		seen[PluginTypeSkill] = true
	}
	if len(m.Tools) > 0 && !seen[PluginTypeTool] {
		types = append(types, PluginTypeTool)
		seen[PluginTypeTool] = true
	}
	for _, p := range m.Providers {
		if !seen[p.Type] {
			types = append(types, p.Type)
			seen[p.Type] = true
		}
	}

	return types
}

// ProvidersByType returns providers filtered by the given type.
func (m *Manifest) ProvidersByType(t PluginType) []ProviderDef {
	var result []ProviderDef
	for _, p := range m.Providers {
		if p.Type == t {
			result = append(result, p)
		}
	}
	return result
}

// HasProviders returns true if the plugin provides any providers.
func (m *Manifest) HasProviders() bool {
	return len(m.Providers) > 0
}

// IsProviderType returns true if the given type is a provider type
// (memory, sandbox, embedding, observer) as opposed to a content type
// (agent, skill, tool).
func IsProviderType(t PluginType) bool {
	switch t {
	case PluginTypeMemory, PluginTypeSandbox, PluginTypeEmbedding, PluginTypeObserver:
		return true
	default:
		return false
	}
}
