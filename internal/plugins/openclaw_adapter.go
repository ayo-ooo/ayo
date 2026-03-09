// Package plugins provides OpenClaw plugin integration for Ayo build system.
//
// This adapter enables OpenClaw plugins to be installed and used natively
// in Ayo projects, allowing seamless integration with the OpenClaw ecosystem.
package plugins

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// OpenClawPluginManifest represents an OpenClaw plugin manifest.
// OpenClaw uses a different manifest format than Ayo.
type OpenClawPluginManifest struct {
	// Name is the plugin identifier.
	Name string `json:"name" yaml:"name" toml:"name"`

	// Version is the semantic version.
	Version string `json:"version" yaml:"version" toml:"version"`

	// Description describes what the plugin provides.
	Description string `json:"description" yaml:"description" toml:"description"`

	// Author is the plugin author or organization.
	Author string `json:"author,omitempty" yaml:"author,omitempty" toml:"author,omitempty"`

	// Repository is the source repository URL.
	Repository string `json:"repository,omitempty" yaml:"repository,omitempty" toml:"repository,omitempty"`

	// License is the SPDX license identifier.
	License string `json:"license,omitempty" yaml:"license,omitempty" toml:"license,omitempty"`

	// Components lists the components provided by this plugin.
	// OpenClaw plugins can provide multiple component types.
	Components []OpenClawComponent `json:"components" yaml:"components" toml:"components"`

	// Dependencies lists other OpenClaw plugins this plugin depends on.
	Dependencies []OpenClawDependency `json:"dependencies,omitempty" yaml:"dependencies,omitempty" toml:"dependencies,omitempty"`

	// Compatibility specifies version requirements.
	Compatibility string `json:"compatibility,omitempty" yaml:"compatibility,omitempty" toml:"compatibility,omitempty"`

	// AyoVersion specifies the minimum ayo version required (if adapted).
	AyoVersion string `json:"ayo_version,omitempty" yaml:"ayo_version,omitempty" toml:"ayo_version,omitempty"`
}

// OpenClawComponent represents a component in an OpenClaw plugin.
type OpenClawComponent struct {
	// Name is the component identifier.
	Name string `json:"name" yaml:"name" toml:"name"`

	// Type is the component type (skill, tool, agent, etc.).
	Type string `json:"type" yaml:"type" toml:"type"`

	// Description describes what the component does.
	Description string `json:"description" yaml:"description" toml:"description"`

	// EntryPoint is the path to the component implementation.
	EntryPoint string `json:"entry_point" yaml:"entry_point" toml:"entry_point"`

	// Config contains component-specific configuration.
	Config map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty" toml:"config,omitempty"`

	// Metadata contains additional component metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty" toml:"metadata,omitempty"`
}

// OpenClawDependency represents a plugin dependency.
type OpenClawDependency struct {
	// Name is the dependency identifier.
	Name string `json:"name" yaml:"name" toml:"name"`

	// Version specifies the version requirement.
	Version string `json:"version" yaml:"version" toml:"version"`

	// Type specifies the dependency type (optional, required, etc.).
	Type string `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
}

// OpenClawPluginAdapter converts OpenClaw plugins to Ayo's plugin format.
type OpenClawPluginAdapter struct {
	// PluginDir is the directory containing the OpenClaw plugin.
	PluginDir string

	// ManifestPath is the path to the OpenClaw manifest file.
	ManifestPath string

	// FS allows using embedded filesystems (for testing).
	FS fs.FS
}

// NewOpenClawPluginAdapter creates a new OpenClaw plugin adapter.
func NewOpenClawPluginAdapter(pluginDir, manifestPath string) *OpenClawPluginAdapter {
	return &OpenClawPluginAdapter{
		PluginDir:    pluginDir,
		ManifestPath: manifestPath,
	}
}

// Adapt converts an OpenClaw plugin to Ayo's plugin format.
func (a *OpenClawPluginAdapter) Adapt() (*Manifest, error) {
	// Load the OpenClaw manifest
	openClawManifest, err := a.loadOpenClawManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenClaw manifest: %w", err)
	}

	// Convert to Ayo manifest
	ayoManifest, err := a.convertToAyoManifest(openClawManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to Ayo manifest: %w", err)
	}

	return ayoManifest, nil
}

// loadOpenClawManifest loads an OpenClaw manifest from file.
func (a *OpenClawPluginAdapter) loadOpenClawManifest() (*OpenClawPluginManifest, error) {
	// Read the manifest file
	data, err := os.ReadFile(a.ManifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenClaw manifest: %w", err)
	}

	// Determine the format based on file extension
	ext := filepath.Ext(a.ManifestPath)
	var manifest OpenClawPluginManifest

	switch ext {
	case ".json":
		err = json.Unmarshal(data, &manifest)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &manifest)
	case ".toml":
		err = toml.Unmarshal(data, &manifest)
	default:
		return nil, fmt.Errorf("unsupported manifest format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenClaw manifest: %w", err)
	}

	return &manifest, nil
}

// convertToAyoManifest converts an OpenClaw manifest to Ayo's format.
func (a *OpenClawPluginAdapter) convertToAyoManifest(openClawManifest *OpenClawPluginManifest) (*Manifest, error) {
	ayoManifest := &Manifest{
		Name:        fmt.Sprintf("openclaw-%s", openClawManifest.Name),
		Version:     openClawManifest.Version,
		Description: fmt.Sprintf("OpenClaw plugin: %s", openClawManifest.Description),
		Author:      openClawManifest.Author,
		Repository:  openClawManifest.Repository,
		License:     openClawManifest.License,
		AyoVersion:  openClawManifest.AyoVersion,
	}

	// Convert dependencies
	if len(openClawManifest.Dependencies) > 0 {
		ayoManifest.Dependencies = &Dependencies{}
		
		for _, dep := range openClawManifest.Dependencies {
			// Format dependency as "name@version" or just "name"
			depStr := fmt.Sprintf("openclaw-%s", dep.Name)
			if dep.Version != "" {
				depStr += "@" + dep.Version
			}
			ayoManifest.Dependencies.Plugins = append(ayoManifest.Dependencies.Plugins, depStr)
		}
	}

	// Convert components to Ayo plugin types
	for _, component := range openClawManifest.Components {
		switch component.Type {
		case "skill":
			ayoManifest.Skills = append(ayoManifest.Skills, component.Name)
			
		case "tool":
			ayoManifest.Tools = append(ayoManifest.Tools, component.Name)
			
		case "agent":
			ayoManifest.Agents = append(ayoManifest.Agents, component.Name)
			
		case "memory":
			ayoManifest.Providers = append(ayoManifest.Providers, ProviderDef{
				Name:        component.Name,
				Type:        PluginTypeMemory,
				Description: component.Description,
				EntryPoint:  component.EntryPoint,
				Config:      component.Config,
			})
			
		case "embedding":
			ayoManifest.Providers = append(ayoManifest.Providers, ProviderDef{
				Name:        component.Name,
				Type:        PluginTypeEmbedding,
				Description: component.Description,
				EntryPoint:  component.EntryPoint,
				Config:      component.Config,
			})
			
		case "observer":
			ayoManifest.Providers = append(ayoManifest.Providers, ProviderDef{
				Name:        component.Name,
				Type:        PluginTypeObserver,
				Description: component.Description,
				EntryPoint:  component.EntryPoint,
				Config:      component.Config,
			})
			
		case "planner":
			// Default to near-term planner
			plannerType := PlannerTypeNear
			if compType, ok := component.Metadata["planner_type"].(string); ok {
				if compType == "long" {
					plannerType = PlannerTypeLong
				}
			}
			
			ayoManifest.Planners = append(ayoManifest.Planners, PlannerDef{
				Name:        component.Name,
				Type:        plannerType,
				Description: component.Description,
				EntryPoint:  component.EntryPoint,
				Config:      component.Config,
			})
			
		default:
			// Unknown component type - skip for now
			// Could be handled as custom components in future
		}
	}

	return ayoManifest, nil
}



// AdaptOpenClawPluginFromDirectory adapts an OpenClaw plugin from a directory.
// This function searches for OpenClaw manifest files and converts them.
func AdaptOpenClawPluginFromDirectory(pluginDir string) (*Manifest, error) {
	// Search for OpenClaw manifest files
	manifestPaths := findOpenClawManifests(pluginDir)
	if len(manifestPaths) == 0 {
		return nil, fmt.Errorf("no OpenClaw manifest files found in %s", pluginDir)
	}

	// Use the first manifest found
	manifestPath := manifestPaths[0]
	adapter := NewOpenClawPluginAdapter(pluginDir, manifestPath)

	return adapter.Adapt()
}

// findOpenClawManifests searches for OpenClaw manifest files in a directory.
func findOpenClawManifests(dir string) []string {
	var manifestPaths []string
	
	manifestNames := []string{
		"manifest.json",
		"manifest.yaml",
		"manifest.yml", 
		"manifest.toml",
		"plugin.json",
		"plugin.yaml",
		"plugin.yml",
		"plugin.toml",
		"openclaw.json",
		"openclaw.yaml",
		"openclaw.yml",
		"openclaw.toml",
	}
	
	for _, name := range manifestNames {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			manifestPaths = append(manifestPaths, path)
		}
	}
	
	return manifestPaths
}

// InstallOpenClawPlugin installs an OpenClaw plugin by adapting it and installing as an Ayo plugin.
func InstallOpenClawPlugin(pluginRef string, opts *InstallOptions) (*InstallResult, error) {
	// First, install as a regular git plugin
	result, err := Install(pluginRef, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to install OpenClaw plugin: %w", err)
	}

	// Get the plugin directory
	pluginDir := PluginDir(result.Plugin.Name)

	// Adapt the OpenClaw plugin to Ayo format
	ayoManifest, err := AdaptOpenClawPluginFromDirectory(pluginDir)
	if err != nil {
		// If adaptation fails, it might not be an OpenClaw plugin
		// That's okay - it can still work as a regular plugin
		return result, nil
	}

	// Save the adapted manifest
	manifestPath := filepath.Join(pluginDir, "manifest.json")
	manifestData, err := json.MarshalIndent(ayoManifest, "", "  ")
	if err != nil {
		return result, fmt.Errorf("failed to save adapted manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return result, fmt.Errorf("failed to write adapted manifest: %w", err)
	}

	// Update the result with the adapted plugin information
	result.Manifest = ayoManifest
	result.IsOpenClawAdapted = true

	return result, nil
}

