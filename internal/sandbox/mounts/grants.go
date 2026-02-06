package mounts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
)

// GrantMode represents the access mode for a mount grant.
type GrantMode string

const (
	// GrantModeReadOnly allows read-only access.
	GrantModeReadOnly GrantMode = "readonly"
	// GrantModeReadWrite allows read and write access.
	GrantModeReadWrite GrantMode = "readwrite"
)

// Permission represents a granted filesystem access.
type Permission struct {
	Path      string    `json:"path"`
	Mode      GrantMode `json:"mode"`
	GrantedAt time.Time `json:"granted_at"`
	GrantedBy string    `json:"granted_by"`
}

// GrantsFile represents the mounts.json file structure.
type GrantsFile struct {
	Version     int          `json:"version"`
	Permissions []Permission `json:"permissions"`
}

// GrantService manages filesystem mount permissions.
type GrantService struct {
	mu       sync.RWMutex
	filePath string
	grants   *GrantsFile
}

// NewGrantService creates a new grant service.
func NewGrantService() *GrantService {
	return &GrantService{
		filePath: grantsFilePath(),
	}
}

// grantsFilePath returns the path to the mounts.json file.
func grantsFilePath() string {
	return filepath.Join(paths.DataDir(), "mounts.json")
}

// Load reads the grants file from disk.
func (s *GrantService) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize with empty grants
			s.grants = &GrantsFile{
				Version:     1,
				Permissions: []Permission{},
			}
			return nil
		}
		return fmt.Errorf("read grants file: %w", err)
	}

	var grants GrantsFile
	if err := json.Unmarshal(data, &grants); err != nil {
		return fmt.Errorf("parse grants file: %w", err)
	}

	s.grants = &grants
	return nil
}

// Save writes the grants file to disk.
func (s *GrantService) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.grants == nil {
		return fmt.Errorf("no grants loaded")
	}

	// Ensure directory exists
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create grants directory: %w", err)
	}

	data, err := json.MarshalIndent(s.grants, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal grants: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("write grants file: %w", err)
	}

	return nil
}

// Grant adds a permission for a path.
func (s *GrantService) Grant(path string, mode GrantMode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.grants == nil {
		s.grants = &GrantsFile{Version: 1, Permissions: []Permission{}}
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Check if already granted
	for i, p := range s.grants.Permissions {
		if p.Path == absPath {
			// Update existing grant
			s.grants.Permissions[i].Mode = mode
			s.grants.Permissions[i].GrantedAt = time.Now()
			return nil
		}
	}

	// Add new grant
	s.grants.Permissions = append(s.grants.Permissions, Permission{
		Path:      absPath,
		Mode:      mode,
		GrantedAt: time.Now(),
		GrantedBy: "user",
	})

	return nil
}

// Revoke removes a permission for a path.
func (s *GrantService) Revoke(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.grants == nil {
		return nil
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Find and remove the grant
	for i, p := range s.grants.Permissions {
		if p.Path == absPath {
			s.grants.Permissions = append(
				s.grants.Permissions[:i],
				s.grants.Permissions[i+1:]...,
			)
			return nil
		}
	}

	return nil // Not found is not an error
}

// List returns all granted permissions.
func (s *GrantService) List() []Permission {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.grants == nil {
		return nil
	}

	// Return a copy to prevent modification
	result := make([]Permission, len(s.grants.Permissions))
	copy(result, s.grants.Permissions)
	return result
}

// IsGranted checks if a path has the requested access mode.
// It checks parent directories as well - a grant for /Users/alex/Code
// covers /Users/alex/Code/project.
func (s *GrantService) IsGranted(path string, mode GrantMode) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.grants == nil {
		return false
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Check each grant
	for _, p := range s.grants.Permissions {
		// Check if path is under granted path
		if isUnderPath(absPath, p.Path) {
			// Check mode
			if p.Mode == GrantModeReadWrite {
				return true // Read-write covers everything
			}
			if p.Mode == GrantModeReadOnly && mode == GrantModeReadOnly {
				return true // Read-only grant matches read-only request
			}
		}
	}

	return false
}

// GetGrant returns the permission for a specific path, or nil if not found.
func (s *GrantService) GetGrant(path string) *Permission {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.grants == nil {
		return nil
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil
	}

	for _, p := range s.grants.Permissions {
		if p.Path == absPath {
			return &p
		}
	}

	return nil
}

// isUnderPath checks if checkPath is under or equal to basePath.
func isUnderPath(checkPath, basePath string) bool {
	// Normalize paths
	checkPath = filepath.Clean(checkPath)
	basePath = filepath.Clean(basePath)

	// Exact match
	if checkPath == basePath {
		return true
	}

	// Check if checkPath is under basePath
	// Ensure we're checking full path components (not /Users/alex vs /Users/alexo)
	if strings.HasPrefix(checkPath, basePath+string(filepath.Separator)) {
		return true
	}

	return false
}

// LoadGrants is a convenience function to load grants from the default location.
func LoadGrants() (*GrantService, error) {
	service := NewGrantService()
	if err := service.Load(); err != nil {
		return nil, err
	}
	return service, nil
}

// ConfigMount represents a resolved mount from .ayo.json.
type ConfigMount struct {
	Path       string    // Absolute path to mount
	Mode       GrantMode // readonly or readwrite
	ConfigPath string    // Path to .ayo.json that defined this mount
}

// ResolveProjectMounts resolves paths from a .ayo.json mounts map.
// Paths are resolved relative to configDir:
//   - Relative paths (e.g., ".", "./lib") are relative to configDir
//   - ~/ paths are expanded to user home directory
//   - Absolute paths are used as-is
//
// Returns an error if a path cannot be resolved.
func ResolveProjectMounts(configDir string, mounts map[string]string) ([]ConfigMount, error) {
	if len(mounts) == 0 {
		return nil, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}

	result := make([]ConfigMount, 0, len(mounts))

	for path, modeStr := range mounts {
		// Validate mode
		mode := GrantMode(modeStr)
		if mode != GrantModeReadOnly && mode != GrantModeReadWrite {
			return nil, fmt.Errorf("invalid mount mode %q for path %q (must be 'readonly' or 'readwrite')", modeStr, path)
		}

		// Resolve path
		var absPath string
		if strings.HasPrefix(path, "~/") {
			// Expand home directory
			absPath = filepath.Join(homeDir, path[2:])
		} else if filepath.IsAbs(path) {
			// Already absolute
			absPath = path
		} else {
			// Relative to config directory
			absPath = filepath.Join(configDir, path)
		}

		// Clean and resolve the path
		absPath = filepath.Clean(absPath)

		result = append(result, ConfigMount{
			Path:       absPath,
			Mode:       mode,
			ConfigPath: filepath.Join(configDir, ".ayo.json"),
		})
	}

	return result, nil
}

// ValidateProjectMounts checks that all project mounts are covered by persistent grants.
// Returns a list of paths that are NOT granted (security violation).
// This enforces the rule: .ayo.json can only restrict, not grant new access.
func ValidateProjectMounts(projectMounts []ConfigMount, grants *GrantService) []string {
	if grants == nil || len(projectMounts) == 0 {
		return nil
	}

	var violations []string
	for _, pm := range projectMounts {
		if !grants.IsGranted(pm.Path, pm.Mode) {
			violations = append(violations, pm.Path)
		}
	}

	return violations
}

// MergedMount represents a mount with its effective mode after merging.
type MergedMount struct {
	Path   string
	Mode   GrantMode
	Source string // "cli", "project", or "grants"
}

// MergeMounts combines mounts from multiple sources with priority:
//  1. CLI --mount flags (highest)
//  2. .ayo.json project mounts (middle)
//  3. mounts.json persistent grants (lowest)
//
// Higher priority sources can RESTRICT access (downgrade readwrite to readonly)
// but cannot GRANT new access (a path must be in grants to be usable).
func MergeMounts(cliMounts map[string]GrantMode, projectMounts []ConfigMount, grants *GrantService) []MergedMount {
	// Start with all grants as base
	result := make(map[string]MergedMount)

	if grants != nil {
		for _, p := range grants.List() {
			result[p.Path] = MergedMount{
				Path:   p.Path,
				Mode:   p.Mode,
				Source: "grants",
			}
		}
	}

	// Apply project mounts (can only restrict, not add)
	for _, pm := range projectMounts {
		if existing, ok := result[pm.Path]; ok {
			// Can only restrict (readwrite -> readonly), not upgrade
			if pm.Mode == GrantModeReadOnly && existing.Mode == GrantModeReadWrite {
				result[pm.Path] = MergedMount{
					Path:   pm.Path,
					Mode:   GrantModeReadOnly,
					Source: "project",
				}
			}
			// If project specifies readwrite but grants only has readonly, keep readonly
		}
		// If path not in grants, it's ignored (security: can't grant new access)
	}

	// Apply CLI mounts (can only restrict, not add)
	for path, mode := range cliMounts {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		if existing, ok := result[absPath]; ok {
			// Can only restrict (readwrite -> readonly), not upgrade
			if mode == GrantModeReadOnly && existing.Mode == GrantModeReadWrite {
				result[absPath] = MergedMount{
					Path:   absPath,
					Mode:   GrantModeReadOnly,
					Source: "cli",
				}
			}
		}
		// If path not in grants, it's ignored (security: can't grant new access)
	}

	// Convert to slice
	mounts := make([]MergedMount, 0, len(result))
	for _, m := range result {
		mounts = append(mounts, m)
	}

	return mounts
}
