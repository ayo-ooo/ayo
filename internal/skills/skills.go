package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillSource indicates where a skill was discovered from.
type SkillSource int

const (
	// SourceAgentSpecific is a skill from the agent's own skills directory.
	SourceAgentSpecific SkillSource = iota
	// SourceUserShared is a skill from the user's shared skills directory (~/.config/ayo/skills).
	SourceUserShared
	// SourceBuiltIn is a skill from the built-in skills directory (~/.local/share/ayo/skills).
	SourceBuiltIn
	// SourcePlugin is a skill from an installed plugin.
	SourcePlugin
)

// String returns a human-readable name for the skill source.
func (s SkillSource) String() string {
	switch s {
	case SourceAgentSpecific:
		return "agent"
	case SourceUserShared:
		return "user"
	case SourceBuiltIn:
		return "built-in"
	case SourcePlugin:
		return "plugin"
	default:
		return "unknown"
	}
}

// Metadata contains the parsed frontmatter and location info for a skill.
type Metadata struct {
	// Required fields (agentskills spec)
	Name        string
	Description string

	// Optional fields (agentskills spec)
	License      string
	Compatibility string
	AllowedTools string
	RawMetadata  map[string]string

	// Internal fields
	Path       string      // Absolute path to SKILL.md
	Source     SkillSource // Where this skill came from
	HasScripts bool        // scripts/ directory exists
	HasRefs    bool        // references/ directory exists
	HasAssets  bool        // assets/ directory exists
}

// Version returns the version from metadata, if present.
func (m Metadata) Version() string {
	if m.RawMetadata != nil {
		return m.RawMetadata["version"]
	}
	return ""
}

// Author returns the author from metadata, if present.
func (m Metadata) Author() string {
	if m.RawMetadata != nil {
		return m.RawMetadata["author"]
	}
	return ""
}

// Skill contains the full skill definition including body content.
type Skill struct {
	Metadata Metadata
	Body     string
}

// DiscoveryResult contains the results of skill discovery.
type DiscoveryResult struct {
	Skills   []Metadata
	Warnings []string
}

var namePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Discover scans agent and shared skill directories, returning all found skills.
// Agent skills take priority over shared skills with the same name.
func Discover(agentSkillsDir, sharedSkillsDir string) DiscoveryResult {
	return DiscoverWithSources([]SkillSourceDir{
		{Path: agentSkillsDir, Source: SourceAgentSpecific, Label: "agent"},
		{Path: sharedSkillsDir, Source: SourceUserShared, Label: "shared"},
	})
}

// SkillSourceDir represents a directory to scan for skills.
type SkillSourceDir struct {
	Path   string
	Source SkillSource
	Label  string
}

// DiscoverWithSources scans multiple directories for skills in priority order.
// Earlier sources take priority over later sources with the same skill name.
func DiscoverWithSources(sources []SkillSourceDir) DiscoveryResult {
	results := make(map[string]Metadata)
	warnings := make([]string, 0)

	appendWarning := func(msg string) {
		warnings = append(warnings, msg)
	}

	for _, src := range sources {
		if src.Path == "" {
			continue
		}
		entries, err := os.ReadDir(src.Path)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			dirPath := filepath.Join(src.Path, entry.Name())
			skillPath := filepath.Join(dirPath, "SKILL.md")
			data, err := os.ReadFile(skillPath)
			if err != nil {
				appendWarning(fmt.Sprintf("%s: missing SKILL.md", dirPath))
				continue
			}
			meta, _, err := parseSkill(skillPath, entry.Name(), string(data))
			if err != nil {
				appendWarning(fmt.Sprintf("%s: %v", skillPath, err))
				continue
			}
			if _, exists := results[meta.Name]; exists {
				appendWarning(fmt.Sprintf("duplicate skill %s from %s ignored", meta.Name, src.Label))
				continue
			}
			meta.Path = skillPath
			meta.Source = src.Source
			
			// Check for optional directories
			meta.HasScripts = dirExists(filepath.Join(dirPath, "scripts"))
			meta.HasRefs = dirExists(filepath.Join(dirPath, "references"))
			meta.HasAssets = dirExists(filepath.Join(dirPath, "assets"))
			
			results[meta.Name] = meta
		}
	}

	skills := make([]Metadata, 0, len(results))
	for _, meta := range results {
		skills = append(skills, meta)
	}

	return DiscoveryResult{Skills: skills, Warnings: warnings}
}

// dirExists checks if a directory exists.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// Load reads the full skill content from disk.
func Load(meta Metadata) (Skill, error) {
	var s Skill
	if meta.Path == "" {
		return s, fmt.Errorf("missing skill path for %s", meta.Name)
	}
	data, err := os.ReadFile(meta.Path)
	if err != nil {
		return s, err
	}
	parsedMeta, body, err := parseSkill(meta.Path, filepath.Base(filepath.Dir(meta.Path)), string(data))
	if err != nil {
		return s, err
	}
	parsedMeta.Path = meta.Path
	parsedMeta.Source = meta.Source
	parsedMeta.HasScripts = meta.HasScripts
	parsedMeta.HasRefs = meta.HasRefs
	parsedMeta.HasAssets = meta.HasAssets
	return Skill{Metadata: parsedMeta, Body: body}, nil
}

func parseSkill(path, dirName, content string) (Metadata, string, error) {
	var meta Metadata
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 || parts[0] != "" {
		return meta, "", fmt.Errorf("missing frontmatter")
	}
	yamlPart := strings.TrimSpace(parts[1])
	body := strings.TrimSpace(parts[2])

	var raw map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlPart), &raw); err != nil {
		return meta, "", fmt.Errorf("invalid frontmatter: %w", err)
	}

	// Parse required field: name
	nameAny, ok := raw["name"]
	if !ok {
		return meta, "", fmt.Errorf("missing name")
	}
	name, ok := nameAny.(string)
	if !ok || strings.TrimSpace(name) == "" {
		return meta, "", fmt.Errorf("invalid name")
	}
	if len(name) > 64 || len(name) < 1 || !namePattern.MatchString(name) {
		return meta, "", fmt.Errorf("invalid name")
	}
	if name != dirName {
		return meta, "", fmt.Errorf("name must match directory %s", dirName)
	}

	// Parse required field: description
	descAny, ok := raw["description"]
	if !ok {
		return meta, "", fmt.Errorf("missing description")
	}
	desc, ok := descAny.(string)
	if !ok {
		return meta, "", fmt.Errorf("invalid description")
	}
	trimmedDesc := strings.TrimSpace(desc)
	if trimmedDesc == "" || len(trimmedDesc) > 1024 {
		return meta, "", fmt.Errorf("invalid description")
	}

	meta.Name = name
	meta.Description = trimmedDesc

	// Parse optional field: license
	if licenseAny, ok := raw["license"]; ok {
		if license, ok := licenseAny.(string); ok {
			meta.License = strings.TrimSpace(license)
		}
	}

	// Parse optional field: compatibility
	if compatAny, ok := raw["compatibility"]; ok {
		if compat, ok := compatAny.(string); ok {
			trimmed := strings.TrimSpace(compat)
			if len(trimmed) <= 500 {
				meta.Compatibility = trimmed
			}
		}
	}

	// Parse optional field: allowed-tools
	if toolsAny, ok := raw["allowed-tools"]; ok {
		if tools, ok := toolsAny.(string); ok {
			meta.AllowedTools = strings.TrimSpace(tools)
		}
	}

	// Parse optional field: metadata (map[string]string)
	if metaAny, ok := raw["metadata"]; ok {
		if metaMap, ok := metaAny.(map[string]interface{}); ok {
			meta.RawMetadata = make(map[string]string)
			for k, v := range metaMap {
				if strVal, ok := v.(string); ok {
					meta.RawMetadata[k] = strVal
				} else {
					// Convert non-string values to string
					meta.RawMetadata[k] = fmt.Sprintf("%v", v)
				}
			}
		}
	}

	return meta, body, nil
}
