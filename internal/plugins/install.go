package plugins

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
)

// Installation errors
var (
	ErrGitNotFound       = errors.New("git is not installed or not in PATH")
	ErrCloneFailed       = errors.New("failed to clone repository")
	ErrDependencyMissing = errors.New("required dependency is missing")
)

// InstallOptions configures plugin installation behavior.
type InstallOptions struct {
	// Force overwrites an existing installation.
	Force bool

	// Renames maps original agent handles to new names for conflict resolution.
	Renames map[string]string

	// SkipDependencyCheck skips checking for required binaries.
	SkipDependencyCheck bool
}

// InstallResult contains information about a successful installation.
type InstallResult struct {
	Plugin      *InstalledPlugin
	Manifest    *Manifest
	MissingDeps []BinaryDep // Dependencies that are missing (with install hints)
}

// Install installs a plugin from a git repository.
func Install(pluginRef string, opts *InstallOptions) (*InstallResult, error) {
	if opts == nil {
		opts = &InstallOptions{}
	}

	// Check git is available
	if _, err := exec.LookPath("git"); err != nil {
		return nil, ErrGitNotFound
	}

	// Parse the plugin reference
	gitURL, name, err := ParsePluginURL(pluginRef)
	if err != nil {
		return nil, fmt.Errorf("parse plugin reference: %w", err)
	}

	// Load registry
	registry, err := LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("load registry: %w", err)
	}

	// Check if already installed
	if registry.Has(name) && !opts.Force {
		return nil, fmt.Errorf("%w: %s (use --force to reinstall)", ErrPluginExists, name)
	}

	// Create plugins directory
	pluginsDir := paths.PluginsDir()
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create plugins dir: %w", err)
	}

	// Target directory for this plugin
	pluginDir := paths.PluginDir(name)

	// Remove existing if force install
	if opts.Force && registry.Has(name) {
		if err := os.RemoveAll(pluginDir); err != nil {
			return nil, fmt.Errorf("remove existing plugin: %w", err)
		}
		registry.Remove(name)
	}

	// Clone the repository
	commit, err := gitClone(gitURL, pluginDir)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCloneFailed, err)
	}

	// Load and validate manifest
	manifest, err := LoadManifest(pluginDir)
	if err != nil {
		// Clean up on failure
		os.RemoveAll(pluginDir)
		return nil, fmt.Errorf("validate manifest: %w", err)
	}

	// Check dependencies
	var missingDeps []BinaryDep
	if !opts.SkipDependencyCheck && manifest.Dependencies != nil {
		missingDeps = CheckMissingDependencies(manifest)
	}

	// Create installed plugin record
	plugin := &InstalledPlugin{
		Name:        name,
		Version:     manifest.Version,
		GitURL:      gitURL,
		GitCommit:   commit,
		InstalledAt: time.Now(),
		Path:        pluginDir,
		Agents:      manifest.Agents,
		Skills:      manifest.Skills,
		Tools:       manifest.Tools,
		Renames:     opts.Renames,
	}

	// Add to registry
	if err := registry.Add(plugin); err != nil {
		os.RemoveAll(pluginDir)
		return nil, fmt.Errorf("register plugin: %w", err)
	}

	return &InstallResult{
		Plugin:      plugin,
		Manifest:    manifest,
		MissingDeps: missingDeps,
	}, nil
}

// gitClone clones a repository and returns the commit hash.
func gitClone(url, dest string) (string, error) {
	// Clone with depth 1 for faster downloads
	cmd := exec.Command("git", "clone", "--depth", "1", url, dest)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return "", err
	}

	// Get the commit hash
	commit, err := getGitCommit(dest)
	if err != nil {
		return "", err
	}

	return commit, nil
}

// getGitCommit returns the current HEAD commit hash.
func getGitCommit(repoDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// InstallFromLocal installs a plugin from a local directory (for development).
func InstallFromLocal(localPath string, opts *InstallOptions) (*InstallResult, error) {
	if opts == nil {
		opts = &InstallOptions{}
	}

	// Load and validate manifest from local path
	manifest, err := LoadManifest(localPath)
	if err != nil {
		return nil, fmt.Errorf("validate manifest: %w", err)
	}

	name := manifest.Name

	// Load registry
	registry, err := LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("load registry: %w", err)
	}

	// Check if already installed
	if registry.Has(name) && !opts.Force {
		return nil, fmt.Errorf("%w: %s (use --force to reinstall)", ErrPluginExists, name)
	}

	// Create plugins directory
	pluginsDir := paths.PluginsDir()
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create plugins dir: %w", err)
	}

	// Target directory for this plugin
	pluginDir := paths.PluginDir(name)

	// Remove existing if force install
	if opts.Force && registry.Has(name) {
		if err := os.RemoveAll(pluginDir); err != nil {
			return nil, fmt.Errorf("remove existing plugin: %w", err)
		}
		registry.Remove(name)
	}

	// Copy the local directory to plugins dir
	if err := copyDir(localPath, pluginDir); err != nil {
		return nil, fmt.Errorf("copy plugin: %w", err)
	}

	// Check dependencies
	var missingDeps []BinaryDep
	if !opts.SkipDependencyCheck && manifest.Dependencies != nil {
		missingDeps = CheckMissingDependencies(manifest)
	}

	// Get git commit if it's a git repo
	commit := ""
	if _, err := os.Stat(filepath.Join(localPath, ".git")); err == nil {
		commit, _ = getGitCommit(localPath)
	}

	// Create installed plugin record
	plugin := &InstalledPlugin{
		Name:        name,
		Version:     manifest.Version,
		GitURL:      "local:" + localPath,
		GitCommit:   commit,
		InstalledAt: time.Now(),
		Path:        pluginDir,
		Agents:      manifest.Agents,
		Skills:      manifest.Skills,
		Tools:       manifest.Tools,
		Renames:     opts.Renames,
	}

	// Add to registry
	if err := registry.Add(plugin); err != nil {
		os.RemoveAll(pluginDir)
		return nil, fmt.Errorf("register plugin: %w", err)
	}

	return &InstallResult{
		Plugin:      plugin,
		Manifest:    manifest,
		MissingDeps: missingDeps,
	}, nil
}

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}

// CheckDependencies checks if all required dependencies are available.
// Returns the names of missing binaries (for backwards compatibility).
func CheckDependencies(manifest *Manifest) []string {
	deps := CheckMissingDependencies(manifest)
	names := make([]string, len(deps))
	for i, d := range deps {
		names[i] = d.Name
	}
	return names
}

// CheckMissingDependencies checks for missing dependencies and returns
// the full BinaryDep objects with installation hints.
func CheckMissingDependencies(manifest *Manifest) []BinaryDep {
	var missing []BinaryDep

	if manifest.Dependencies == nil {
		return missing
	}

	for _, binary := range manifest.Dependencies.Binaries {
		if _, err := exec.LookPath(binary.Name); err != nil {
			missing = append(missing, binary)
		}
	}

	return missing
}
