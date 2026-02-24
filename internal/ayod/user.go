package ayod

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// UserManager handles user creation and lookup inside the sandbox.
type UserManager struct {
	mu    sync.RWMutex
	users map[string]*userInfo
}

type userInfo struct {
	Username string
	UID      int
	GID      int
	Home     string
}

// NewUserManager creates a new user manager.
func NewUserManager() *UserManager {
	return &UserManager{
		users: make(map[string]*userInfo),
	}
}

// AddUser creates a new user in the sandbox.
// This uses the busybox/Alpine adduser command.
func (m *UserManager) AddUser(req UserAddRequest) (*UserAddResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if user already exists
	if info, ok := m.users[req.Username]; ok {
		return &UserAddResponse{
			UID:  info.UID,
			GID:  info.GID,
			Home: info.Home,
		}, nil
	}

	// Check if user exists in system
	if u, err := user.Lookup(req.Username); err == nil {
		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)
		info := &userInfo{
			Username: req.Username,
			UID:      uid,
			GID:      gid,
			Home:     u.HomeDir,
		}
		m.users[req.Username] = info
		return &UserAddResponse{
			UID:  info.UID,
			GID:  info.GID,
			Home: info.Home,
		}, nil
	}

	// Create user with adduser (busybox/Alpine style)
	shell := req.Shell
	if shell == "" {
		shell = "/bin/sh"
	}

	// adduser -D disables password (no interactive prompt)
	// -s sets shell, -h sets home directory
	home := filepath.Join("/home", req.Username)
	cmd := exec.Command("adduser", "-D", "-s", shell, "-h", home, req.Username)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("adduser failed: %w: %s", err, string(output))
	}

	// Look up the created user to get UID/GID
	u, err := user.Lookup(req.Username)
	if err != nil {
		return nil, fmt.Errorf("lookup user after creation: %w", err)
	}

	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)

	info := &userInfo{
		Username: req.Username,
		UID:      uid,
		GID:      gid,
		Home:     home,
	}
	m.users[req.Username] = info

	// Extract dotfiles if provided
	if len(req.Dotfiles) > 0 {
		if err := m.extractDotfiles(info, req.Dotfiles); err != nil {
			// Log but don't fail - user was created
			fmt.Fprintf(os.Stderr, "warning: failed to extract dotfiles: %v\n", err)
		}
	}

	return &UserAddResponse{
		UID:  info.UID,
		GID:  info.GID,
		Home: info.Home,
	}, nil
}

// extractDotfiles extracts a tar archive of dotfiles to the user's home directory.
func (m *UserManager) extractDotfiles(info *userInfo, dotfiles []byte) error {
	// Write tar to temp file
	tmpFile, err := os.CreateTemp("", "dotfiles-*.tar")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(dotfiles); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	// Extract to home directory
	cmd := exec.Command("tar", "-xf", tmpFile.Name(), "-C", info.Home)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tar extract: %w: %s", err, string(output))
	}

	// Fix ownership
	cmd = exec.Command("chown", "-R", fmt.Sprintf("%d:%d", info.UID, info.GID), info.Home)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("chown: %w: %s", err, string(output))
	}

	return nil
}

// GetUser returns user info by username.
func (m *UserManager) GetUser(username string) (*userInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	info, ok := m.users[username]
	return info, ok
}

// ListUsers returns all created usernames.
func (m *UserManager) ListUsers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.users))
	for name := range m.users {
		names = append(names, name)
	}
	return names
}

// SanitizeUsername converts an agent handle to a valid Unix username.
// Handles like "@ayo" become "ayo", with special characters replaced.
func SanitizeUsername(handle string) string {
	// Remove @ prefix
	name := strings.TrimPrefix(handle, "@")

	// Replace invalid characters with underscore
	var result strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			result.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			result.WriteRune(r + 32) // lowercase
		case r >= '0' && r <= '9' && result.Len() > 0:
			result.WriteRune(r)
		case (r == '_' || r == '-') && result.Len() > 0:
			result.WriteRune(r)
		default:
			// Skip invalid characters at start, otherwise add underscore
			// but only if we have content and last char isn't already underscore
			if result.Len() > 0 {
				s := result.String()
				if s[len(s)-1] != '_' {
					result.WriteRune('_')
				}
			}
		}
	}

	username := result.String()
	// Trim trailing underscore
	username = strings.TrimSuffix(username, "_")

	if username == "" {
		username = "agent"
	}

	// Truncate to 32 characters (Unix limit)
	if len(username) > 32 {
		username = username[:32]
	}

	// Trim trailing underscore after truncation
	username = strings.TrimSuffix(username, "_")
	username = strings.TrimSuffix(username, "-")

	return username
}
