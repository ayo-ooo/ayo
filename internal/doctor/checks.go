// Package doctor provides system health checks for ayo.
package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	// Removed daemon import during daemon infrastructure removal
	// "github.com/alexcabrera/ayo/internal/daemon"
	"github.com/alexcabrera/ayo/internal/paths"
)

// Status represents the result of a health check.
type Status string

const (
	StatusPass Status = "pass"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

// CheckResult represents the outcome of a single health check.
type CheckResult struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Status   Status `json:"status"`
	Message  string `json:"message"`
	Fix      string `json:"fix,omitempty"`
	DocsURL  string `json:"docs_url,omitempty"`
	Fixable  bool   `json:"fixable,omitempty"`
}

// Summary contains aggregate check results.
type Summary struct {
	Passed   int           `json:"passed"`
	Warnings int           `json:"warnings"`
	Errors   int           `json:"errors"`
	Results  []CheckResult `json:"results"`
}

// Checker runs health checks.
type Checker struct {
	results []CheckResult
}

// NewChecker creates a new health checker.
func NewChecker() *Checker {
	return &Checker{}
}

// Add adds a check result.
func (c *Checker) Add(result CheckResult) {
	c.results = append(c.results, result)
}

// Summary returns the aggregate results.
func (c *Checker) Summary() Summary {
	var s Summary
	s.Results = c.results
	for _, r := range c.results {
		switch r.Status {
		case StatusPass:
			s.Passed++
		case StatusWarn:
			s.Warnings++
		case StatusFail:
			s.Errors++
		}
	}
	return s
}

// CheckSystemRequirements verifies system dependencies.
func (c *Checker) CheckSystemRequirements(ctx context.Context) {
	// Check OS version
	if runtime.GOOS == "darwin" {
		osVer := getMacOSVersion()
		if osVer != "" {
			c.Add(CheckResult{
				Name:     "macOS Version",
				Category: "System Requirements",
				Status:   StatusPass,
				Message:  osVer,
			})
		} else {
			c.Add(CheckResult{
				Name:     "macOS Version",
				Category: "System Requirements",
				Status:   StatusWarn,
				Message:  "could not determine version",
			})
		}
	} else if runtime.GOOS == "linux" {
		c.Add(CheckResult{
			Name:     "Linux",
			Category: "System Requirements",
			Status:   StatusPass,
			Message:  runtime.GOARCH,
		})
	}

	// Check Go version
	goVer := runtime.Version()
	c.Add(CheckResult{
		Name:     "Go Version",
		Category: "System Requirements",
		Status:   StatusPass,
		Message:  goVer,
	})

	// Check git
	if gitPath, err := exec.LookPath("git"); err == nil {
		gitVer := getGitVersion()
		c.Add(CheckResult{
			Name:     "git",
			Category: "System Requirements",
			Status:   StatusPass,
			Message:  fmt.Sprintf("%s (%s)", gitVer, gitPath),
		})
	} else {
		c.Add(CheckResult{
			Name:     "git",
			Category: "System Requirements",
			Status:   StatusFail,
			Message:  "not found in PATH",
			Fix:      "Install git from https://git-scm.com",
		})
	}
}

// CheckDaemon is disabled during daemon infrastructure removal.
func (c *Checker) CheckDaemon(ctx context.Context) {
	// Daemon infrastructure is being removed
	c.Add(CheckResult{
		Name:     "Daemon",
		Category: "Daemon",
		Status:   StatusWarn,
		Message:  "daemon infrastructure removed - moving to standalone executables",
	})
}

// CheckAPIKeys verifies API key configuration.
func (c *Checker) CheckAPIKeys(ctx context.Context) {
	apiKeys := map[string]string{
		"ANTHROPIC_API_KEY": "Anthropic (Claude)",
		"OPENAI_API_KEY":    "OpenAI",
		"GOOGLE_API_KEY":    "Google (Gemini)",
	}

	hasAny := false
	for envVar, name := range apiKeys {
		if os.Getenv(envVar) != "" {
			hasAny = true
			c.Add(CheckResult{
				Name:     name,
				Category: "API Keys",
				Status:   StatusPass,
				Message:  envVar + " set",
			})
		}
	}

	if !hasAny {
		c.Add(CheckResult{
			Name:     "API Keys",
			Category: "API Keys",
			Status:   StatusWarn,
			Message:  "no API keys configured",
			Fix:      "Set ANTHROPIC_API_KEY, OPENAI_API_KEY, or GOOGLE_API_KEY",
			DocsURL:  "https://docs.ayo.dev/setup#api-keys",
		})
	}
}

// CheckPaths verifies ayo directories and files.
func (c *Checker) CheckPaths(ctx context.Context) {
	// Config directory
	configDir := paths.ConfigDir()
	if info, err := os.Stat(configDir); err == nil && info.IsDir() {
		c.Add(CheckResult{
			Name:     "Config Directory",
			Category: "Paths",
			Status:   StatusPass,
			Message:  configDir,
		})
	} else {
		c.Add(CheckResult{
			Name:     "Config Directory",
			Category: "Paths",
			Status:   StatusWarn,
			Message:  configDir + " (not found)",
			Fix:      "Run: ayo setup",
			Fixable:  true,
		})
	}

	// Data directory
	dataDir := paths.DataDir()
	if info, err := os.Stat(dataDir); err == nil && info.IsDir() {
		c.Add(CheckResult{
			Name:     "Data Directory",
			Category: "Paths",
			Status:   StatusPass,
			Message:  dataDir,
		})
	} else {
		c.Add(CheckResult{
			Name:     "Data Directory",
			Category: "Paths",
			Status:   StatusWarn,
			Message:  dataDir + " (not found)",
			Fix:      "Run: ayo setup",
			Fixable:  true,
		})
	}

	// Database
	dbPath := paths.DatabasePath()
	if info, err := os.Stat(dbPath); err == nil && !info.IsDir() {
		c.Add(CheckResult{
			Name:     "Database",
			Category: "Paths",
			Status:   StatusPass,
			Message:  dbPath,
		})
	} else {
		c.Add(CheckResult{
			Name:     "Database",
			Category: "Paths",
			Status:   StatusWarn,
			Message:  dbPath + " (not found)",
			Fix:      "Run: ayo setup",
			Fixable:  true,
		})
	}
}

// CheckSquads lists squad status.
func (c *Checker) CheckSquads(ctx context.Context) {
	// Check for old squad system (deprecated)
	squads, err := paths.ListSquads()
	if err != nil {
		c.Add(CheckResult{
			Name:     "Squads",
			Category: "Squads",
			Status:   StatusWarn,
			Message:  "error checking squads: " + err.Error(),
		})
		return
	}

	// Check for new team projects in current directory
	teamProjects := 0
	if TeamConfigExists(".") {
		teamProjects = 1
	}

	if len(squads) == 0 && teamProjects == 0 {
		c.Add(CheckResult{
			Name:     "Squads",
			Category: "Squads",
			Status:   StatusPass,
			Message:  "none configured (use 'ayo init' to create a project, then 'ayo add-agent' to add more agents)",
		})
		return
	}

	// Report both old squads and new team projects
	var messages []string
	if len(squads) > 0 {
		messages = append(messages, fmt.Sprintf("%d legacy squads: %s", len(squads), strings.Join(squads, ", ")))
	}
	if teamProjects > 0 {
		messages = append(messages, "1 team project (team.toml)")
	}

	c.Add(CheckResult{
		Name:     "Squads",
		Category: "Squads",
		Status:   StatusPass,
		Message:  strings.Join(messages, "; "),
	})
}

// TeamConfigExists checks if team.toml exists in a directory.
func TeamConfigExists(dir string) bool {
	teamPath := filepath.Join(dir, "team.toml")
	_, err := os.Stat(teamPath)
	return err == nil
}

// Helper functions

func getMacOSVersion() string {
	cmd := exec.Command("sw_vers", "-productVersion")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return "macOS " + strings.TrimSpace(string(out))
}

func getGitVersion() string {
	cmd := exec.Command("git", "--version")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	// "git version 2.43.0" -> "2.43.0"
	ver := strings.TrimSpace(string(out))
	ver = strings.TrimPrefix(ver, "git version ")
	return ver
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	if mins > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dh", hours)
}
