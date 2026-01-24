package skills

import (
	"sort"

	"github.com/alexcabrera/ayo/internal/paths"
)

// DiscoveryOptions configures skill discovery behavior.
type DiscoveryOptions struct {
	// AgentSkillsDir is the agent-specific skills directory (highest priority).
	AgentSkillsDir string
	// SharedDirs are additional shared skills directories (in priority order).
	// These are scanned after agent-specific but before built-in.
	SharedDirs []string
	// UserSharedDir is the user's shared skills directory (~/.config/ayo/skills).
	// Deprecated: Use SharedDirs for multiple directories.
	UserSharedDir string
	// BuiltinDir is the directory for installed built-in skills (~/.local/share/ayo/skills).
	BuiltinDir string

	// IncludeSkills limits discovery to only these skill names (empty = all).
	IncludeSkills []string
	// ExcludeSkills excludes these skill names from discovery.
	ExcludeSkills []string
	// IgnoreBuiltin skips built-in skills.
	IgnoreBuiltin bool
	// IgnoreShared skips user shared skills.
	IgnoreShared bool
}

// DiscoverAll scans all configured directories for skills in priority order.
// Earlier sources take priority over later sources with the same skill name.
// Priority: agent-specific > shared dirs (in order) > user shared > built-in > plugins
// Skills are filtered by include/exclude lists and ignore flags.
func DiscoverAll(opts DiscoveryOptions) DiscoveryResult {
	// Build source list in priority order
	var sources []SkillSourceDir

	// 1. Agent-specific skills (highest priority)
	if opts.AgentSkillsDir != "" {
		sources = append(sources, SkillSourceDir{
			Path:   opts.AgentSkillsDir,
			Source: SourceAgentSpecific,
			Label:  "agent",
		})
	}

	// 2. Additional shared directories (in priority order)
	if !opts.IgnoreShared {
		for _, dir := range opts.SharedDirs {
			if dir != "" {
				sources = append(sources, SkillSourceDir{
					Path:   dir,
					Source: SourceUserShared,
					Label:  "shared",
				})
			}
		}
	}

	// 3. User shared skills (~/.config/ayo/skills) - legacy single dir support
	if opts.UserSharedDir != "" && !opts.IgnoreShared {
		sources = append(sources, SkillSourceDir{
			Path:   opts.UserSharedDir,
			Source: SourceUserShared,
			Label:  "user",
		})
	}

	// 4. Built-in skills (~/.local/share/ayo/skills)
	if opts.BuiltinDir != "" && !opts.IgnoreBuiltin {
		sources = append(sources, SkillSourceDir{
			Path:   opts.BuiltinDir,
			Source: SourceBuiltIn,
			Label:  "builtin",
		})
	}

	// 5. Plugin skills (lowest priority)
	for _, dir := range paths.AllPluginSkillsDirs() {
		sources = append(sources, SkillSourceDir{
			Path:   dir,
			Source: SourcePlugin,
			Label:  "plugin",
		})
	}

	// Discover from all sources
	result := DiscoverWithSources(sources)

	// Apply include/exclude filters
	result.Skills = filterSkills(result.Skills, opts.IncludeSkills, opts.ExcludeSkills)

	// Sort by name for consistent ordering
	sort.Slice(result.Skills, func(i, j int) bool {
		return result.Skills[i].Name < result.Skills[j].Name
	})

	return result
}

// filterSkills applies include/exclude filters to a skill list.
func filterSkills(skills []Metadata, include, exclude []string) []Metadata {
	// Build lookup maps
	includeSet := make(map[string]bool)
	for _, name := range include {
		includeSet[name] = true
	}
	excludeSet := make(map[string]bool)
	for _, name := range exclude {
		excludeSet[name] = true
	}

	var filtered []Metadata
	for _, skill := range skills {
		// Check exclude first
		if excludeSet[skill.Name] {
			continue
		}

		// If include list is specified, only include those
		if len(includeSet) > 0 && !includeSet[skill.Name] {
			continue
		}

		filtered = append(filtered, skill)
	}

	return filtered
}

// DiscoverForAgent is a convenience function that discovers skills for an agent
// using the standard directory structure.
// skillsDirs is a list of shared skill directories in priority order.
func DiscoverForAgent(agentDir string, skillsDirs []string, cfg DiscoveryFilterConfig) DiscoveryResult {
	return DiscoverAll(DiscoveryOptions{
		AgentSkillsDir: agentDir,
		SharedDirs:     skillsDirs,
		IncludeSkills:  cfg.IncludeSkills,
		ExcludeSkills:  cfg.ExcludeSkills,
		IgnoreBuiltin:  cfg.IgnoreBuiltin,
		IgnoreShared:   cfg.IgnoreShared,
	})
}

// DiscoveryFilterConfig contains the filter configuration from an agent's config.
type DiscoveryFilterConfig struct {
	IncludeSkills []string
	ExcludeSkills []string
	IgnoreBuiltin bool
	IgnoreShared  bool
}
