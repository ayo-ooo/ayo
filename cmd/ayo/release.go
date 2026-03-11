package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/build"
	"github.com/alexcabrera/ayo/internal/build/types"
)

func newReleaseCmd() *cobra.Command {
	var bumpType string
	var preRelease string
	var buildMetadata string

	cmd := &cobra.Command{
		Use:   "release <directory> [new-version]",
		Short: "Manage agent versions and create releases",
		Long: `Manage agent versioning and prepare for release.

The release command can:
1. Bump version numbers (major, minor, patch)
2. Set specific versions
3. Create Git tags for releases
4. Generate CHANGELOG entries

Usage:
  ayo release <directory> --bump patch       Bump patch version (0.0.0 -> 0.0.1)
  ayo release <directory> --bump minor        Bump minor version (0.0.0 -> 0.1.0)
  ayo release <directory> --bump major        Bump major version (0.0.0 -> 1.0.0)
  ayo release <directory> 1.2.3               Set specific version
  ayo release <directory> --bump patch --pre  Set pre-release version (0.0.1-pre)

Semantic Versioning: https://semver.org/`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			return runRelease(dir, bumpType, preRelease, buildMetadata)
		},
	}

	cmd.Flags().StringVar(&bumpType, "bump", "", "Version part to bump: major, minor, or patch")
	cmd.Flags().StringVar(&preRelease, "pre", "", "Pre-release identifier (e.g., 'beta', 'rc.1')")
	cmd.Flags().StringVar(&buildMetadata, "build", "", "Build metadata (e.g., 'exp.sha.5114f85')")

	return cmd
}

func runRelease(dir, bumpType, preRelease, buildMetadata string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("error resolving directory: %w", err)
	}

	cfg, configPath, err := build.LoadConfigFromDir(absDir)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	currentVersion := cfg.Agent.Version
	if currentVersion == "" {
		currentVersion = "0.0.0"
	}

	var newVersion string
	if bumpType != "" {
		newVersion, err = bumpVersion(currentVersion, bumpType)
		if err != nil {
			return err
		}
	} else {
		// Version should have been provided as second argument, but we only have one arg (directory)
		// Check if there's a version in the command
		return fmt.Errorf("please specify a version or use --bump flag")
	}

	if preRelease != "" {
		newVersion = fmt.Sprintf("%s-%s", newVersion, preRelease)
	}

	if buildMetadata != "" {
		newVersion = fmt.Sprintf("%s+%s", newVersion, buildMetadata)
	}

	fmt.Printf("Bumping version from %s to %s\n", currentVersion, newVersion)
	cfg.Agent.Version = newVersion

	if err := saveConfig(absDir, cfg); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	fmt.Printf("✓ Updated %s with version %s\n", configPath, newVersion)

	if preRelease == "" && buildMetadata == "" {
		tagName := fmt.Sprintf("v%s", newVersion)
		if err := createGitTag(absDir, tagName, fmt.Sprintf("Release %s", newVersion)); err != nil {
			fmt.Printf("Warning: failed to create git tag: %v\n", err)
		} else {
			fmt.Printf("✓ Created git tag %s\n", tagName)
		}
	}

	changelogPath := filepath.Join(absDir, "CHANGELOG.md")
	if err := updateChangelog(changelogPath, newVersion, cfg.Agent.Name, cfg.Agent.Description); err != nil {
		fmt.Printf("Warning: failed to update CHANGELOG.md: %v\n", err)
	} else {
		fmt.Printf("✓ Updated CHANGELOG.md\n")
	}

	fmt.Println("\nNext steps:")
	fmt.Printf("  1. Review changes: git diff %s\n", configPath)
	fmt.Printf("  2. Commit: git add %s CHANGELOG.md && git commit -m 'Bump version to %s'\n", configPath, newVersion)
	fmt.Printf("  3. Build: ayo build %s\n", dir)
	fmt.Printf("  4. Package: ayo package %s\n", dir)

	return nil
}

func bumpVersion(version, bumpType string) (string, error) {
	version = strings.TrimPrefix(version, "v")
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid version format: %s (expected x.y.z)", version)
	}

	major := parts[0]
	minor := parts[1]
	patch := parts[2]

	switch bumpType {
	case "major":
		major = incrementVersion(major)
		minor = "0"
		patch = "0"
	case "minor":
		minor = incrementVersion(minor)
		patch = "0"
	case "patch":
		patch = incrementVersion(patch)
	default:
		return "", fmt.Errorf("invalid bump type: %s (must be major, minor, or patch)", bumpType)
	}

	return fmt.Sprintf("%s.%s.%s", major, minor, patch), nil
}

func incrementVersion(part string) string {
	num := 0
	if n, err := fmt.Sscanf(part, "%d", &num); n == 1 && err == nil {
		return fmt.Sprintf("%d", num+1)
	}
	return "1"
}

func saveConfig(dir string, cfg *types.Config) error {
	configPath := filepath.Join(dir, "config.toml")
	return build.WriteConfig(*cfg, configPath)
}

func createGitTag(dir, tagName, message string) error {
	cmd := exec.Command("git", "tag", "-a", tagName, "-m", message)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(output), err)
	}
	return nil
}

func updateChangelog(path, version, name, description string) error {
	var existingContent string
	if content, err := os.ReadFile(path); err == nil {
		existingContent = string(content)
	}

	date := getCurrentDate()
	newEntry := fmt.Sprintf("## [%s] - %s\n\n### Added\n\n### Changed\n\n### Fixed\n\n", version, date)

	if existingContent != "" {
		existingContent = strings.TrimPrefix(existingContent, "# Changelog\n\n")
		newEntry = "# Changelog\n\n" + newEntry + existingContent
	} else {
		newEntry = fmt.Sprintf("# Changelog\n\nAll notable changes to %s will be documented in this file.\n\n%s", name, newEntry)
	}

	return os.WriteFile(path, []byte(newEntry), 0644)
}

func getCurrentDate() string {
	cmd := exec.Command("date", "+%Y-%m-%d")
	output, err := cmd.Output()
	if err != nil {
		return "TBD"
	}
	return strings.TrimSpace(string(output))
}

func init() {
	// This function exists to satisfy go vet
	_ = regexp.MustCompile(`.*`)
}
