package plugins

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
)

// UpdateOptions configures plugin update behavior.
type UpdateOptions struct {
	// Force updates even if at the same version.
	Force bool

	// DryRun shows what would be updated without making changes.
	DryRun bool
}

// UpdateResult contains information about an update operation.
type UpdateResult struct {
	Plugin      *InstalledPlugin
	OldVersion  string
	NewVersion  string
	OldCommit   string
	NewCommit   string
	WasUpdated  bool
	SkipReason  string
}

// Update updates a single plugin to the latest version.
func Update(name string, opts *UpdateOptions) (*UpdateResult, error) {
	if opts == nil {
		opts = &UpdateOptions{}
	}

	// Load registry
	registry, err := LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("load registry: %w", err)
	}

	// Get plugin info
	plugin, err := registry.Get(name)
	if err != nil {
		return nil, err
	}

	result := &UpdateResult{
		Plugin:     plugin,
		OldVersion: plugin.Version,
		OldCommit:  plugin.GitCommit,
	}

	// Can't update local installs
	if strings.HasPrefix(plugin.GitURL, "local:") {
		result.SkipReason = "local installation (reinstall from source to update)"
		return result, nil
	}

	pluginDir := paths.PluginDir(name)

	// Fetch latest from remote
	if err := gitFetch(pluginDir); err != nil {
		return nil, fmt.Errorf("fetch updates: %w", err)
	}

	// Check if there are updates
	localCommit, err := getGitCommit(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("get local commit: %w", err)
	}

	remoteCommit, err := getGitRemoteCommit(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("get remote commit: %w", err)
	}

	result.NewCommit = remoteCommit

	// Check if update is needed
	if localCommit == remoteCommit && !opts.Force {
		result.SkipReason = "already at latest version"
		result.NewVersion = plugin.Version
		return result, nil
	}

	// Dry run - don't actually update
	if opts.DryRun {
		result.WasUpdated = false
		result.NewVersion = "(pending)"
		return result, nil
	}

	// Pull latest changes
	if err := gitPull(pluginDir); err != nil {
		return nil, fmt.Errorf("pull updates: %w", err)
	}

	// Reload manifest to get new version
	manifest, err := LoadManifest(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("load updated manifest: %w", err)
	}

	result.NewVersion = manifest.Version
	result.WasUpdated = true

	// Update registry
	plugin.Version = manifest.Version
	plugin.GitCommit = remoteCommit
	plugin.UpdatedAt = time.Now()
	plugin.Agents = manifest.Agents
	plugin.Skills = manifest.Skills
	plugin.Tools = manifest.Tools

	if err := registry.Update(plugin); err != nil {
		return nil, fmt.Errorf("update registry: %w", err)
	}

	return result, nil
}

// UpdateAll updates all installed plugins.
func UpdateAll(opts *UpdateOptions) ([]*UpdateResult, error) {
	registry, err := LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("load registry: %w", err)
	}

	plugins := registry.List()
	results := make([]*UpdateResult, 0, len(plugins))

	for _, plugin := range plugins {
		result, err := Update(plugin.Name, opts)
		if err != nil {
			// Add error result but continue with other plugins
			results = append(results, &UpdateResult{
				Plugin:     plugin,
				SkipReason: fmt.Sprintf("error: %v", err),
			})
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// CheckForUpdates checks which plugins have updates available.
func CheckForUpdates() ([]*UpdateResult, error) {
	return UpdateAll(&UpdateOptions{DryRun: true})
}

// gitFetch fetches from the remote without merging.
func gitFetch(repoDir string) error {
	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = repoDir
	return cmd.Run()
}

// gitPull pulls latest changes from the remote.
func gitPull(repoDir string) error {
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = repoDir
	return cmd.Run()
}

// getGitRemoteCommit returns the commit hash of origin/HEAD.
func getGitRemoteCommit(repoDir string) (string, error) {
	// First, determine the default branch
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "origin/HEAD")
	cmd.Dir = repoDir

	output, err := cmd.Output()
	if err != nil {
		// Fallback to origin/main or origin/master
		for _, branch := range []string{"origin/main", "origin/master"} {
			cmd := exec.Command("git", "rev-parse", branch)
			cmd.Dir = repoDir
			output, err = cmd.Output()
			if err == nil {
				return strings.TrimSpace(string(output)), nil
			}
		}
		return "", err
	}

	// Get the commit for the remote branch
	remoteBranch := strings.TrimSpace(string(output))
	cmd = exec.Command("git", "rev-parse", remoteBranch)
	cmd.Dir = repoDir

	output, err = cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// GetInstalledVersion returns the installed version of a plugin.
func GetInstalledVersion(name string) (string, error) {
	registry, err := LoadRegistry()
	if err != nil {
		return "", err
	}

	plugin, err := registry.Get(name)
	if err != nil {
		return "", err
	}

	return plugin.Version, nil
}

// GetAvailableVersion fetches and returns the latest available version.
func GetAvailableVersion(name string) (string, error) {
	registry, err := LoadRegistry()
	if err != nil {
		return "", err
	}

	plugin, err := registry.Get(name)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(plugin.GitURL, "local:") {
		return plugin.Version, nil
	}

	pluginDir := paths.PluginDir(name)

	// Fetch and check manifest on remote
	if err := gitFetch(pluginDir); err != nil {
		return "", err
	}

	// Read manifest from origin
	cmd := exec.Command("git", "show", "origin/HEAD:manifest.json")
	cmd.Dir = pluginDir

	output, err := cmd.Output()
	if err != nil {
		// Try main/master
		for _, branch := range []string{"origin/main", "origin/master"} {
			cmd = exec.Command("git", "show", branch+":manifest.json")
			cmd.Dir = pluginDir
			output, err = cmd.Output()
			if err == nil {
				break
			}
		}
		if err != nil {
			return "", fmt.Errorf("read remote manifest: %w", err)
		}
	}

	// Parse manifest
	var manifest Manifest
	if err := parseManifestBytes(output, &manifest); err != nil {
		return "", err
	}

	return manifest.Version, nil
}

// parseManifestBytes parses manifest JSON bytes.
func parseManifestBytes(data []byte, m *Manifest) error {
	return json.Unmarshal(data, m)
}

// Disable disables a plugin without uninstalling it.
func Disable(name string) error {
	registry, err := LoadRegistry()
	if err != nil {
		return err
	}

	plugin, err := registry.Get(name)
	if err != nil {
		return err
	}

	plugin.Disabled = true
	return registry.Update(plugin)
}

// Enable enables a disabled plugin.
func Enable(name string) error {
	registry, err := LoadRegistry()
	if err != nil {
		return err
	}

	plugin, err := registry.Get(name)
	if err != nil {
		return err
	}

	plugin.Disabled = false
	return registry.Update(plugin)
}

// IsDisabled checks if a plugin is disabled.
func IsDisabled(name string) (bool, error) {
	registry, err := LoadRegistry()
	if err != nil {
		return false, err
	}

	plugin, err := registry.Get(name)
	if err != nil {
		return false, err
	}

	return plugin.Disabled, nil
}


