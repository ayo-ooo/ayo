package builtin

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alexcabrera/ayo/internal/paths"
)

// Version is the current version of built-in agents and skills.
// Bump this when built-in content changes to trigger reinstallation.
const Version = "16"

// ModifiedAgent represents an installed agent that has local modifications
type ModifiedAgent struct {
	Handle        string   // e.g., "@ayo"
	InstalledDir  string   // Full path to installed agent dir
	ModifiedFiles []string // List of modified file names
}

// CheckModifiedAgents compares installed agents against embedded versions
// and returns a list of agents that have been modified locally.
func CheckModifiedAgents() ([]ModifiedAgent, error) {
	installDir := InstallDir()
	var modified []ModifiedAgent

	// Get list of embedded agents
	entries, err := agentsFS.ReadDir("agents")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		handle := entry.Name() // e.g., "@ayo"
		agentInstallDir := filepath.Join(installDir, handle)

		// Check if agent is installed
		if _, err := os.Stat(agentInstallDir); os.IsNotExist(err) {
			continue // Not installed, nothing to compare
		}

		// Compare files
		modifiedFiles, err := compareAgentFiles(handle, agentInstallDir)
		if err != nil {
			continue // Skip on error
		}

		if len(modifiedFiles) > 0 {
			modified = append(modified, ModifiedAgent{
				Handle:        handle,
				InstalledDir:  agentInstallDir,
				ModifiedFiles: modifiedFiles,
			})
		}
	}

	return modified, nil
}

// compareAgentFiles compares installed files against embedded files for an agent
func compareAgentFiles(handle, installedDir string) ([]string, error) {
	var modifiedFiles []string
	embeddedBase := filepath.Join("agents", handle)

	err := fs.WalkDir(agentsFS, embeddedBase, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Get relative path from agent dir
		relPath, _ := filepath.Rel(embeddedBase, path)
		installedPath := filepath.Join(installedDir, relPath)

		// Read embedded file
		embeddedData, err := agentsFS.ReadFile(path)
		if err != nil {
			return nil // Skip
		}

		// Read installed file
		installedData, err := os.ReadFile(installedPath)
		if err != nil {
			if os.IsNotExist(err) {
				// File was deleted - consider it modified
				modifiedFiles = append(modifiedFiles, relPath+" (deleted)")
			}
			return nil
		}

		// Compare content
		if !bytes.Equal(embeddedData, installedData) {
			modifiedFiles = append(modifiedFiles, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Also check for extra files that don't exist in embedded
	err = filepath.WalkDir(installedDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(installedDir, path)
		embeddedPath := filepath.Join(embeddedBase, relPath)

		if _, err := agentsFS.ReadFile(embeddedPath); err != nil {
			// File exists in installed but not in embedded - it's an addition
			modifiedFiles = append(modifiedFiles, relPath+" (added)")
		}

		return nil
	})

	return modifiedFiles, err
}

// Install extracts built-in agents and skills to the platform-specific install directory.
// It only reinstalls if the version has changed or content is missing.
func Install() error {
	versionFile := VersionFile()

	// Check if already installed with current version
	if !needsInstall(versionFile) {
		return nil
	}

	_, err := ForceInstall()
	return err
}

// ForceInstall extracts built-in agents and skills regardless of version.
// Returns the install directory path.
func ForceInstall() (string, error) {
	installDir := InstallDir()
	skillsInstallDir := SkillsInstallDir()
	promptsInstallDir := PromptsInstallDir()
	versionFile := VersionFile()

	// Create install directories
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return "", fmt.Errorf("create agents install dir: %w", err)
	}
	if err := os.MkdirAll(skillsInstallDir, 0o755); err != nil {
		return "", fmt.Errorf("create skills install dir: %w", err)
	}
	if err := os.MkdirAll(promptsInstallDir, 0o755); err != nil {
		return "", fmt.Errorf("create prompts install dir: %w", err)
	}

	// Extract all agents (including agent-specific skills)
	if err := extractAgents(installDir); err != nil {
		return "", fmt.Errorf("extract agents: %w", err)
	}

	// Extract shared built-in skills
	if err := extractSkills(skillsInstallDir); err != nil {
		return "", fmt.Errorf("extract skills: %w", err)
	}

	// Extract built-in prompts
	if err := extractPrompts(promptsInstallDir); err != nil {
		return "", fmt.Errorf("extract prompts: %w", err)
	}

	// Install config schema to user config directory
	if err := InstallConfigSchema(); err != nil {
		return "", fmt.Errorf("install config schema: %w", err)
	}

	// Write version marker
	if err := os.WriteFile(versionFile, []byte(Version), 0o644); err != nil {
		return "", fmt.Errorf("write version file: %w", err)
	}

	return installDir, nil
}

// needsInstall checks if installation is required
func needsInstall(versionFile string) bool {
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return true
	}
	return string(data) != Version
}

// extractAgents copies all embedded agents to the install directory
func extractAgents(installDir string) error {
	return fs.WalkDir(agentsFS, "agents", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root "agents" directory itself
		if path == "agents" {
			return nil
		}

		// Calculate destination path (strip "agents/" prefix)
		relPath, _ := filepath.Rel("agents", path)
		destPath := filepath.Join(installDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		// Read embedded file
		data, err := agentsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}

		// Write to destination
		if err := os.WriteFile(destPath, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", destPath, err)
		}

		return nil
	})
}

// extractSkills copies all shared embedded skills to the install directory
func extractSkills(installDir string) error {
	return fs.WalkDir(skillsFS, "skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root "skills" directory itself
		if path == "skills" {
			return nil
		}

		// Calculate destination path (strip "skills/" prefix)
		relPath, _ := filepath.Rel("skills", path)
		destPath := filepath.Join(installDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		// Read embedded file
		data, err := skillsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}

		// Write to destination
		if err := os.WriteFile(destPath, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", destPath, err)
		}

		return nil
	})
}

// Uninstall removes installed built-in agents and skills
func Uninstall() error {
	installDir := InstallDir()
	skillsInstallDir := SkillsInstallDir()
	versionFile := VersionFile()

	os.Remove(versionFile)
	os.RemoveAll(skillsInstallDir)
	return os.RemoveAll(installDir)
}

// InstalledAgentDir returns the path to an installed built-in agent
func InstalledAgentDir(handle string) string {
	// Normalize to include @ prefix to match directory name
	if len(handle) > 0 && handle[0] != '@' {
		handle = "@" + handle
	}
	return filepath.Join(InstallDir(), handle)
}

// IsInstalled checks if a built-in agent is installed on the filesystem
func IsInstalled(handle string) bool {
	dir := InstalledAgentDir(handle)
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// InstallDir returns the directory where built-in agents are installed.
// Location: ~/.local/share/ayo/agents
func InstallDir() string {
	return paths.BuiltinAgentsDir()
}

// SkillsInstallDir returns the directory where shared built-in skills are installed.
// Location: ~/.local/share/ayo/skills
func SkillsInstallDir() string {
	return paths.BuiltinSkillsDir()
}

// VersionFile returns the path to the version marker file.
func VersionFile() string {
	return paths.VersionFile()
}

// InstalledSkillDir returns the path to an installed built-in skill
func InstalledSkillDir(name string) string {
	return filepath.Join(SkillsInstallDir(), name)
}

// IsSkillInstalled checks if a built-in skill is installed on the filesystem
func IsSkillInstalled(name string) bool {
	dir := InstalledSkillDir(name)
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// PromptsInstallDir returns the directory where built-in prompts are installed.
// Location: ~/.local/share/ayo/prompts
func PromptsInstallDir() string {
	return paths.BuiltinPromptsDir()
}

// extractPrompts copies all embedded prompts to the install directory
func extractPrompts(installDir string) error {
	return fs.WalkDir(promptsFS, "prompts", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root "prompts" directory itself
		if path == "prompts" {
			return nil
		}

		// Calculate destination path (strip "prompts/" prefix)
		relPath, _ := filepath.Rel("prompts", path)
		destPath := filepath.Join(installDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		// Read embedded file
		data, err := promptsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}

		// Write to destination
		if err := os.WriteFile(destPath, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", destPath, err)
		}
		return nil
	})
}

// ModifiedSkill represents an installed skill that has local modifications
type ModifiedSkill struct {
	Name          string   // e.g., "debugging"
	InstalledDir  string   // Full path to installed skill dir
	ModifiedFiles []string // List of modified file names
}

// CheckModifiedSkills compares installed skills against embedded versions
// and returns a list of skills that have been modified locally.
func CheckModifiedSkills() ([]ModifiedSkill, error) {
	installDir := SkillsInstallDir()
	var modified []ModifiedSkill

	// Get list of embedded skills
	entries, err := skillsFS.ReadDir("skills")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		skillInstallDir := filepath.Join(installDir, skillName)

		// Check if skill is installed
		if _, err := os.Stat(skillInstallDir); os.IsNotExist(err) {
			continue // Not installed, nothing to compare
		}

		// Compare files
		modifiedFiles, err := compareSkillFiles(skillName, skillInstallDir)
		if err != nil {
			continue // Skip on error
		}

		if len(modifiedFiles) > 0 {
			modified = append(modified, ModifiedSkill{
				Name:          skillName,
				InstalledDir:  skillInstallDir,
				ModifiedFiles: modifiedFiles,
			})
		}
	}

	return modified, nil
}

// compareSkillFiles compares installed files against embedded files for a skill
func compareSkillFiles(skillName, installedDir string) ([]string, error) {
	var modifiedFiles []string
	embeddedBase := filepath.Join("skills", skillName)

	err := fs.WalkDir(skillsFS, embeddedBase, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Get relative path from skill dir
		relPath, _ := filepath.Rel(embeddedBase, path)
		installedPath := filepath.Join(installedDir, relPath)

		// Read embedded file
		embeddedData, err := skillsFS.ReadFile(path)
		if err != nil {
			return nil // Skip
		}

		// Read installed file
		installedData, err := os.ReadFile(installedPath)
		if err != nil {
			if os.IsNotExist(err) {
				// File was deleted - consider it modified
				modifiedFiles = append(modifiedFiles, relPath+" (deleted)")
			}
			return nil
		}

		// Compare content
		if !bytes.Equal(embeddedData, installedData) {
			modifiedFiles = append(modifiedFiles, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Also check for extra files that don't exist in embedded
	err = filepath.WalkDir(installedDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(installedDir, path)
		embeddedPath := filepath.Join(embeddedBase, relPath)

		if _, err := skillsFS.ReadFile(embeddedPath); err != nil {
			// File exists in installed but not in embedded - it's an addition
			modifiedFiles = append(modifiedFiles, relPath+" (added)")
		}

		return nil
	})

	return modifiedFiles, err
}

// ConfigSchemaFile returns the path where the config schema is installed.
// Location: ~/.config/ayo/ayo-schema.json (Unix) or %LOCALAPPDATA%\ayo\ayo-schema.json (Windows)
func ConfigSchemaFile() string {
	return paths.ConfigSchemaFile()
}

// InstallConfigSchema writes the embedded config schema to the user config directory.
func InstallConfigSchema() error {
	if len(ConfigSchema) == 0 {
		return nil // No schema embedded
	}

	schemaPath := ConfigSchemaFile()

	// Ensure config directory exists
	if err := os.MkdirAll(paths.ConfigDir(), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	// Write schema file
	if err := os.WriteFile(schemaPath, ConfigSchema, 0o644); err != nil {
		return fmt.Errorf("write schema: %w", err)
	}

	return nil
}
