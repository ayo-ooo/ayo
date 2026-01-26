// Package paths provides directory paths for ayo.
//
// Directory Priority Order (first found wins for lookups):
//  1. ./.config/ayo (local project config)
//  2. ./.local/share/ayo (local project data)
//  3. ~/.config/ayo (user config)
//  4. ~/.local/share/ayo (user data / built-ins)
//
// For writes, ayo uses:
//   - User agents/skills: ~/.config/ayo (or ./.config/ayo with --dev)
//   - Built-in installation: ~/.local/share/ayo (or ./.local/share/ayo with --dev)
package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	devRoot     string
	devRootOnce sync.Once
)

// IsDevMode returns true if ayo is running from a source checkout.
// In dev mode, built-in data is stored in {repo}/.local/share/ayo/ instead of ~/.local/share/ayo/.
func IsDevMode() bool {
	return getDevRoot() != ""
}

// DevRoot returns the repository root if running in dev mode, or empty string otherwise.
func DevRoot() string {
	return getDevRoot()
}

// getDevRoot finds the repository root by checking:
// 1. Walking up from executable location (for built binaries in repo)
// 2. Walking up from current working directory (for go run)
// looking for a go.mod file with "module ayo".
func getDevRoot() string {
	devRootOnce.Do(func() {
		// Try from executable first (handles ./ayo built binary)
		if root := findDevRootFrom(executableDir()); root != "" {
			devRoot = root
			return
		}

		// Try from current working directory (handles go run)
		if wd, err := os.Getwd(); err == nil {
			if root := findDevRootFrom(wd); root != "" {
				devRoot = root
				return
			}
		}
	})
	return devRoot
}

// executableDir returns the directory containing the executable, or empty if unknown.
func executableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
}

// findDevRootFrom walks up from the given directory looking for a go.mod with the ayo module.
func findDevRootFrom(startDir string) string {
	if startDir == "" {
		return ""
	}

	dir := startDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(goModPath); err == nil {
			// Check if this is the ayo module (matches both "module ayo" and "module github.com/.../ayo")
			content := string(data)
			if strings.HasPrefix(content, "module ayo") ||
				strings.Contains(content, "\nmodule ayo") ||
				strings.Contains(content, "/ayo\n") {
				return dir
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached filesystem root
		}
		dir = parent
	}
	return ""
}

// DataDir returns the data directory for ayo.
//
// Dev mode: {repo}/.local/share/ayo (project-local built-ins)
// Production Unix: ~/.local/share/ayo (XDG compliant)
// Production Windows: %LOCALAPPDATA%\ayo
//
// This directory stores built-in agents, built-in skills, and version markers.
// In dev mode, each checkout has its own isolated built-ins.
func DataDir() string {
	// Dev mode: use project-local .local/share/ayo
	if root := getDevRoot(); root != "" {
		return filepath.Join(root, ".local", "share", "ayo")
	}

	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "ayo")
	}
	// Unix (macOS, Linux, etc.) - XDG compliant
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "ayo")
}

// ConfigDir returns the config directory for ayo.
//
// Dev mode: {repo}/.config/ayo (project-local config)
// Production Unix: ~/.config/ayo
// Production Windows: %LOCALAPPDATA%\ayo
//
// This directory stores user configuration and user-created content:
// ayo.json, user agents, user skills, and system prompts.
func ConfigDir() string {
	// Dev mode: use project-local .config/ayo
	if root := getDevRoot(); root != "" {
		return filepath.Join(root, ".config", "ayo")
	}

	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "ayo")
	}
	// Unix (macOS, Linux, etc.)
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ayo")
}

// AgentsDir returns the directory for user-created agents.
// Location: ~/.config/ayo/agents (Unix) or %LOCALAPPDATA%\ayo\agents (Windows)
// This is always the global user directory, even in dev mode.
func AgentsDir() string {
	return filepath.Join(ConfigDir(), "agents")
}

// BuiltinAgentsDir returns the directory for installed built-in agents.
// Dev mode: {repo}/.local/share/ayo/agents
// Production: ~/.local/share/ayo/agents (Unix) or %LOCALAPPDATA%\ayo\agents (Windows)
func BuiltinAgentsDir() string {
	return filepath.Join(DataDir(), "agents")
}

// SkillsDir returns the directory for user shared skills.
// Location: ~/.config/ayo/skills (Unix) or %LOCALAPPDATA%\ayo\skills (Windows)
// This is always the global user directory, even in dev mode.
func SkillsDir() string {
	return filepath.Join(ConfigDir(), "skills")
}

// BuiltinSkillsDir returns the directory for installed built-in skills.
// Dev mode: {repo}/.local/share/ayo/skills
// Production: ~/.local/share/ayo/skills (Unix) or %LOCALAPPDATA%\ayo\skills (Windows)
func BuiltinSkillsDir() string {
	return filepath.Join(DataDir(), "skills")
}

// ConfigFile returns the path to the main config file.
// Location: ~/.config/ayo/ayo.json (Unix) or %LOCALAPPDATA%\ayo\ayo.json (Windows)
// This is always the global user config, even in dev mode.
func ConfigFile() string {
	return filepath.Join(ConfigDir(), "ayo.json")
}

// ConfigSchemaFile returns the path to the config JSON schema file.
// Location: ~/.config/ayo/ayo-schema.json (Unix) or %LOCALAPPDATA%\ayo\ayo-schema.json (Windows)
// The schema is installed during setup and enables IDE validation/autocomplete.
func ConfigSchemaFile() string {
	return filepath.Join(ConfigDir(), "ayo-schema.json")
}

// SystemPromptsDir returns the directory for system prompt files.
// Location: ~/.config/ayo/prompts (Unix) or %LOCALAPPDATA%\ayo\prompts (Windows)
// This is always the global user directory, even in dev mode.
func SystemPromptsDir() string {
	return filepath.Join(ConfigDir(), "prompts")
}

// VersionFile returns the path to the builtin version marker.
// Dev mode: {repo}/.ayo/.builtin-version
// Production: ~/.local/share/ayo/.builtin-version (Unix) or %LOCALAPPDATA%\ayo\.builtin-version (Windows)
func VersionFile() string {
	return filepath.Join(DataDir(), ".builtin-version")
}

// LocalConfigDir returns the local project config directory (./.config/ayo).
// Returns empty string if not in a directory context or on Windows.
func LocalConfigDir() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(wd, ".config", "ayo")
}

// LocalDataDir returns the local project data directory (./.local/share/ayo).
// Returns empty string if not in a directory context or on Windows.
func LocalDataDir() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(wd, ".local", "share", "ayo")
}

// UserConfigDir returns the global user config directory (~/.config/ayo).
// On Windows, returns %LOCALAPPDATA%\ayo.
func UserConfigDir() string {
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "ayo")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ayo")
}

// UserDataDir returns the global user data directory (~/.local/share/ayo).
// On Windows, returns %LOCALAPPDATA%\ayo.
func UserDataDir() string {
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "ayo")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "ayo")
}

// HasLocalConfig returns true if a local config directory exists (./.config/ayo).
func HasLocalConfig() bool {
	dir := LocalConfigDir()
	if dir == "" {
		return false
	}
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// HasLocalData returns true if a local data directory exists (./.local/share/ayo).
func HasLocalData() bool {
	dir := LocalDataDir()
	if dir == "" {
		return false
	}
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// DataDirs returns all data directories where built-in agents/skills could be installed.
// This includes both dev mode locations and production locations.
// Used to determine if an agent should be treated as built-in.
func DataDirs() []string {
	var dirs []string
	seen := make(map[string]bool)

	add := func(dir string) {
		if dir != "" && !seen[dir] {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}

	// Dev mode: {repo}/.local/share/ayo
	if root := getDevRoot(); root != "" {
		add(filepath.Join(root, ".local", "share", "ayo"))
	}

	// Local data: ./.local/share/ayo (from cwd)
	add(LocalDataDir())

	// Production: ~/.local/share/ayo
	add(UserDataDir())

	return dirs
}

// AgentsDirs returns all agent directories in lookup priority order.
// Order: local config, local data, user config, user data (built-in).
// Only includes directories that exist.
func AgentsDirs() []string {
	var dirs []string
	check := func(base string) {
		if base == "" {
			return
		}
		dir := filepath.Join(base, "agents")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			dirs = append(dirs, dir)
		}
	}

	check(LocalConfigDir())
	check(LocalDataDir())
	check(UserConfigDir())
	check(UserDataDir())

	return dirs
}

// SkillsDirs returns all skills directories in lookup priority order.
// Order: local config, local data, user config, user data (built-in).
// Only includes directories that exist.
func SkillsDirs() []string {
	var dirs []string
	check := func(base string) {
		if base == "" {
			return
		}
		dir := filepath.Join(base, "skills")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			dirs = append(dirs, dir)
		}
	}

	check(LocalConfigDir())
	check(LocalDataDir())
	check(UserConfigDir())
	check(UserDataDir())

	return dirs
}

// BuiltinPromptsDir returns the directory for built-in prompts.
// Dev mode: {repo}/.ayo/prompts or ./.local/share/ayo/prompts
// Production: ~/.local/share/ayo/prompts
func BuiltinPromptsDir() string {
	return filepath.Join(DataDir(), "prompts")
}

// UserPromptsDir returns the user prompts directory for overrides.
// Location: ~/.config/ayo/prompts or ./.config/ayo/prompts
func UserPromptsDir() string {
	return filepath.Join(ConfigDir(), "prompts")
}

// FindPromptFile looks for a prompt file in priority order:
// 1. ./.config/ayo/prompts/{name}
// 2. ~/.config/ayo/prompts/{name}
// 3. ./.local/share/ayo/prompts/{name}
// 4. ~/.local/share/ayo/prompts/{name}
// Returns empty string if not found.
func FindPromptFile(name string) string {
	// Priority order: local config, user config, local data, user data
	candidates := []string{
		filepath.Join(LocalConfigDir(), "prompts", name),
		filepath.Join(UserConfigDir(), "prompts", name),
		filepath.Join(LocalDataDir(), "prompts", name),
		filepath.Join(UserDataDir(), "prompts", name),
	}

	for _, path := range candidates {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// DatabasePath returns the path to the SQLite database file.
//
// Local dev mode: ./.local/share/ayo/ayo.db
// Dev mode: {repo}/.ayo/ayo.db
// Production: ~/.local/share/ayo/ayo.db
//
// The database stores session history and messages.
func DatabasePath() string {
	return filepath.Join(DataDir(), "ayo.db")
}

// ToolsDataDir returns the base directory for tool-specific data storage.
// Location: ~/.local/share/ayo/tools (Unix) or %LOCALAPPDATA%\ayo\tools (Windows)
// Each stateful tool gets its own subdirectory for isolated storage.
func ToolsDataDir() string {
	return filepath.Join(DataDir(), "tools")
}

// ToolDataDir returns the data directory for a specific tool.
// Location: ~/.local/share/ayo/tools/{toolName}
// Tools can store their own SQLite databases and other data here.
func ToolDataDir(toolName string) string {
	return filepath.Join(ToolsDataDir(), toolName)
}

// ToolDatabasePath returns the path to a tool's SQLite database.
// Location: ~/.local/share/ayo/tools/{toolName}/{toolName}.db
func ToolDatabasePath(toolName string) string {
	return filepath.Join(ToolDataDir(toolName), toolName+".db")
}

// PluginsDir returns the directory where plugins are installed.
// Location: ~/.local/share/ayo/plugins (Unix) or %LOCALAPPDATA%\ayo\plugins (Windows)
func PluginsDir() string {
	return filepath.Join(DataDir(), "plugins")
}

// PluginDir returns the directory for a specific plugin.
func PluginDir(name string) string {
	return filepath.Join(PluginsDir(), name)
}

// PluginsRegistry returns the path to the plugins registry file.
// Location: ~/.local/share/ayo/packages.json
func PluginsRegistry() string {
	return filepath.Join(DataDir(), "packages.json")
}

// PluginAgentsDir returns the agents directory within a plugin.
func PluginAgentsDir(pluginName string) string {
	return filepath.Join(PluginDir(pluginName), "agents")
}

// PluginSkillsDir returns the skills directory within a plugin.
func PluginSkillsDir(pluginName string) string {
	return filepath.Join(PluginDir(pluginName), "skills")
}

// PluginToolsDir returns the tools directory within a plugin.
func PluginToolsDir(pluginName string) string {
	return filepath.Join(PluginDir(pluginName), "tools")
}

// AllPluginAgentsDirs returns agent directories from all installed plugins.
// This is used when discovering agents.
func AllPluginAgentsDirs() []string {
	pluginsDir := PluginsDir()
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			agentsDir := filepath.Join(pluginsDir, entry.Name(), "agents")
			if info, err := os.Stat(agentsDir); err == nil && info.IsDir() {
				dirs = append(dirs, agentsDir)
			}
		}
	}
	return dirs
}

// AllPluginSkillsDirs returns skill directories from all installed plugins.
// This is used when discovering skills.
func AllPluginSkillsDirs() []string {
	pluginsDir := PluginsDir()
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			skillsDir := filepath.Join(pluginsDir, entry.Name(), "skills")
			if info, err := os.Stat(skillsDir); err == nil && info.IsDir() {
				dirs = append(dirs, skillsDir)
			}
		}
	}
	return dirs
}

// DirectoryConfigFile returns the path to a directory-level config file.
// This is .ayo.json in the given directory.
func DirectoryConfigFile(dir string) string {
	return filepath.Join(dir, ".ayo.json")
}

// FindDirectoryConfig searches for .ayo.json starting from dir and walking up.
// Returns empty string if not found.
func FindDirectoryConfig(dir string) string {
	for {
		configPath := DirectoryConfigFile(dir)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached filesystem root
		}
		dir = parent
	}
	return ""
}

// UserFlowsDir returns the directory for user-created flows.
// Location: ~/.config/ayo/flows (Unix) or %LOCALAPPDATA%\ayo\flows (Windows)
func UserFlowsDir() string {
	return filepath.Join(ConfigDir(), "flows")
}

// BuiltinFlowsDir returns the directory for built-in flows.
// Location: ~/.local/share/ayo/flows (Unix) or %LOCALAPPDATA%\ayo\flows (Windows)
func BuiltinFlowsDir() string {
	return filepath.Join(DataDir(), "flows")
}

// ProjectFlowsDir returns the project-specific flows directory (.ayo/flows).
// Returns empty string if no project .ayo directory exists.
func ProjectFlowsDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Check for .ayo/flows in current directory or parent directories
	dir := wd
	for {
		flowsDir := filepath.Join(dir, ".ayo", "flows")
		if info, err := os.Stat(flowsDir); err == nil && info.IsDir() {
			return flowsDir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached filesystem root
		}
		dir = parent
	}
	return ""
}

// FlowsDirs returns all flows directories in lookup priority order.
// Order: project (.ayo/flows), user config, builtin.
// Only includes directories that exist.
func FlowsDirs() []string {
	var dirs []string
	seen := make(map[string]bool)

	add := func(dir string) {
		if dir != "" && !seen[dir] {
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				seen[dir] = true
				dirs = append(dirs, dir)
			}
		}
	}

	// Project flows first (.ayo/flows)
	add(ProjectFlowsDir())

	// User flows (~/.config/ayo/flows)
	add(UserFlowsDir())

	// Built-in flows (~/.local/share/ayo/flows)
	add(BuiltinFlowsDir())

	return dirs
}
