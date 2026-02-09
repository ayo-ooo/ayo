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
	ErrSecurityScanFailed = errors.New("security scan failed")
)

// InstallOptions configures plugin installation behavior.
type InstallOptions struct {
	// Force overwrites an existing installation.
	Force bool

	// Renames maps original agent handles to new names for conflict resolution.
	Renames map[string]string

	// SkipDependencyCheck skips checking for required binaries.
	SkipDependencyCheck bool

	// SkipSecurityScan skips the security scan (dangerous - use with caution).
	SkipSecurityScan bool
}

// InstallResult contains information about a successful installation.
type InstallResult struct {
	Plugin       *InstalledPlugin
	Manifest     *Manifest
	MissingDeps  []BinaryDep   // Dependencies that are missing (with install hints)
	SecurityScan *ScanResult   // Security scan results (nil if skipped)
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

	// Run security scan unless skipped
	var scanResult *ScanResult
	if !opts.SkipSecurityScan {
		scanner := NewSecurityScanner()
		scanResult, err = scanner.Scan(pluginDir)
		if err != nil {
			os.RemoveAll(pluginDir)
			return nil, fmt.Errorf("security scan error: %w", err)
		}
		if !scanResult.Allowed {
			os.RemoveAll(pluginDir)
			return nil, fmt.Errorf("%w: %s (use --force to skip security scan)", ErrSecurityScanFailed, scanResult.Reason)
		}
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

	// Run post-install hook if present
	if manifest.PostInstall != "" {
		if err := runPostInstallHook(pluginDir, manifest.PostInstall); err != nil {
			// Log warning but don't fail installation
			fmt.Fprintf(os.Stderr, "Warning: post-install hook failed: %v\n", err)
		}
	}

	return &InstallResult{
		Plugin:       plugin,
		Manifest:     manifest,
		MissingDeps:  missingDeps,
		SecurityScan: scanResult,
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

	// Run security scan on source directory unless skipped
	var scanResult *ScanResult
	if !opts.SkipSecurityScan {
		scanner := NewSecurityScanner()
		var err error
		scanResult, err = scanner.Scan(localPath)
		if err != nil {
			return nil, fmt.Errorf("security scan error: %w", err)
		}
		if !scanResult.Allowed {
			return nil, fmt.Errorf("%w: %s (use --force to skip security scan)", ErrSecurityScanFailed, scanResult.Reason)
		}
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

	// Run post-install hook if present
	if manifest.PostInstall != "" {
		if err := runPostInstallHook(pluginDir, manifest.PostInstall); err != nil {
			// Log warning but don't fail installation
			fmt.Fprintf(os.Stderr, "Warning: post-install hook failed: %v\n", err)
		}
	}

	return &InstallResult{
		Plugin:       plugin,
		Manifest:     manifest,
		MissingDeps:  missingDeps,
		SecurityScan: scanResult,
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

// runPostInstallHook runs a post-install script if it exists.
// The script receives the plugin directory as its first argument.
func runPostInstallHook(pluginDir, hookPath string) error {
	fullPath := filepath.Join(pluginDir, hookPath)

	// Check if hook exists
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("post-install script not found: %s", hookPath)
		}
		return err
	}

	// Ensure it's executable (on Unix) or just exists (on Windows)
	if info.IsDir() {
		return fmt.Errorf("post-install path is a directory: %s", hookPath)
	}

	// Run the script with plugin directory as argument
	cmd := exec.Command("/bin/sh", fullPath, pluginDir)
	cmd.Dir = pluginDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
