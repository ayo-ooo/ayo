// Package crush provides integration with the Crush coding agent CLI.
package crush

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

// MinVersion is the minimum supported Crush version.
// Crush headless mode was introduced in this version.
const MinVersion = "0.1.0"

// CrushNotFoundError indicates the crush binary was not found in PATH.
type CrushNotFoundError struct {
	Err error
}

func (e *CrushNotFoundError) Error() string {
	return fmt.Sprintf("crush binary not found in PATH: %v", e.Err)
}

func (e *CrushNotFoundError) Unwrap() error {
	return e.Err
}

// CrushVersionError indicates the crush binary version is incompatible.
type CrushVersionError struct {
	Found    string
	Required string
}

func (e *CrushVersionError) Error() string {
	return fmt.Sprintf("crush version %s is below minimum required %s", e.Found, e.Required)
}

// BinaryInfo contains information about the detected Crush binary.
type BinaryInfo struct {
	Path    string // Full path to the crush executable
	Version string // Parsed version string (e.g., "0.1.0")
}

// FindBinary locates the crush executable in PATH and validates its version.
// Returns CrushNotFoundError if not found, CrushVersionError if version is too old.
func FindBinary(ctx context.Context) (*BinaryInfo, error) {
	// Look up crush in PATH
	path, err := exec.LookPath("crush")
	if err != nil {
		return nil, &CrushNotFoundError{Err: err}
	}

	// Get version
	version, err := getVersion(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get crush version: %w", err)
	}

	// Validate version
	if err := validateVersion(version); err != nil {
		return nil, err
	}

	return &BinaryInfo{
		Path:    path,
		Version: version,
	}, nil
}

// FindBinaryPath locates the crush executable without version validation.
// Useful when you just need the path and will validate separately.
func FindBinaryPath() (string, error) {
	path, err := exec.LookPath("crush")
	if err != nil {
		return "", &CrushNotFoundError{Err: err}
	}
	return path, nil
}

// getVersion runs 'crush --version' and parses the output.
// Expected format: "crush version X.Y.Z" or "crush X.Y.Z"
func getVersion(ctx context.Context, binaryPath string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run crush --version: %w", err)
	}

	return parseVersion(string(output))
}

// parseVersion extracts a semver-compatible version from crush --version output.
// Handles formats like:
//   - "crush version 0.1.0"
//   - "crush 0.1.0"
//   - "0.1.0"
//   - "v0.1.0"
//   - "devel" (development builds)
func parseVersion(output string) (string, error) {
	output = strings.TrimSpace(output)

	// Handle "devel" builds - treat as latest
	if strings.Contains(output, "devel") {
		return "999.999.999", nil
	}

	// Try to find a semver pattern
	// Match X.Y.Z or vX.Y.Z with optional prerelease/build metadata
	versionRe := regexp.MustCompile(`v?(\d+\.\d+\.\d+(?:-[a-zA-Z0-9.]+)?(?:\+[a-zA-Z0-9.]+)?)`)
	matches := versionRe.FindStringSubmatch(output)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse version from: %q", output)
	}

	return matches[1], nil
}

// validateVersion checks if the found version meets minimum requirements.
func validateVersion(found string) error {
	foundVer, err := semver.NewVersion(found)
	if err != nil {
		return fmt.Errorf("invalid version %q: %w", found, err)
	}

	minVer, err := semver.NewVersion(MinVersion)
	if err != nil {
		return fmt.Errorf("invalid minimum version %q: %w", MinVersion, err)
	}

	if foundVer.LessThan(minVer) {
		return &CrushVersionError{
			Found:    found,
			Required: MinVersion,
		}
	}

	return nil
}

// IsAvailable returns true if crush is installed and meets version requirements.
func IsAvailable(ctx context.Context) bool {
	_, err := FindBinary(ctx)
	return err == nil
}
