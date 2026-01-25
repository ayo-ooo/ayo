// Package plugins provides functionality for managing ayo plugins.
// Plugins are distributed via git repositories with the naming convention
// ayo-plugins-<name> and can contain agents, skills, and tools.
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

	// AyoVersion specifies the minimum ayo version required.
	// Uses semver constraints (e.g., ">=0.2.0").
	AyoVersion string `json:"ayo_version,omitempty"`
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
	ErrManifestNotFound    = errors.New("manifest.json not found")
	ErrInvalidManifest     = errors.New("invalid manifest")
	ErrMissingName         = errors.New("manifest: name is required")
	ErrMissingVersion      = errors.New("manifest: version is required")
	ErrMissingDescription  = errors.New("manifest: description is required")
	ErrInvalidName         = errors.New("manifest: name must be lowercase alphanumeric with hyphens")
	ErrInvalidVersion      = errors.New("manifest: version must be valid semver (e.g., 1.0.0)")
	ErrAgentNotFound       = errors.New("manifest: declared agent not found in agents/ directory")
	ErrSkillNotFound       = errors.New("manifest: declared skill not found in skills/ directory")
	ErrToolNotFound        = errors.New("manifest: declared tool not found in tools/ directory")
	ErrInvalidPluginRef    = errors.New("invalid plugin reference: must be a full git URL (https:// or git@)")
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
