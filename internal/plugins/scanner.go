package plugins

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// SecurityScanner scans plugins for adversarial content.
type SecurityScanner struct {
	// MaxFileSize is the maximum file size to scan (default 50KB).
	MaxFileSize int64
}

// NewSecurityScanner creates a new scanner with default settings.
func NewSecurityScanner() *SecurityScanner {
	return &SecurityScanner{
		MaxFileSize: 50 * 1024, // 50KB
	}
}

// ScanResult contains the results of a security scan.
type ScanResult struct {
	Allowed    bool          // Whether the plugin should be allowed
	Reason     string        // Human-readable reason if blocked
	Confidence float64       // 0.0-1.0 confidence in the decision
	Matches    []MatchResult // Pattern matches found
	Warnings   []string      // Non-blocking warnings
}

// Scan scans a plugin directory for security issues.
func (s *SecurityScanner) Scan(pluginPath string) (*ScanResult, error) {
	result := &ScanResult{
		Allowed:    true,
		Confidence: 1.0,
		Matches:    []MatchResult{},
		Warnings:   []string{},
	}

	// Check if directory exists
	info, err := os.Stat(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("access plugin directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", pluginPath)
	}

	// Structural checks
	if err := s.checkStructure(pluginPath, result); err != nil {
		return nil, err
	}

	// Scan text files for adversarial patterns
	if err := s.scanFiles(pluginPath, result); err != nil {
		return nil, err
	}

	// Aggregate results
	s.aggregateResults(result)

	return result, nil
}

// checkStructure performs structural security checks.
func (s *SecurityScanner) checkStructure(pluginPath string, result *ScanResult) error {
	// Check for hidden files
	entries, err := os.ReadDir(pluginPath)
	if err != nil {
		return fmt.Errorf("read directory: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()

		// Hidden files (excluding .gitignore, .gitkeep)
		if strings.HasPrefix(name, ".") && name != ".gitignore" && name != ".gitkeep" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("hidden file: %s", name))
		}

		// Suspicious extensions
		if isSuspiciousExtension(name) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("suspicious file type: %s", name))
		}
	}

	return nil
}

// scanFiles scans all text files for adversarial patterns.
func (s *SecurityScanner) scanFiles(pluginPath string, result *ScanResult) error {
	return filepath.WalkDir(pluginPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Skip non-text files
		if !isTextFile(d.Name()) {
			return nil
		}

		// Check file size
		info, err := d.Info()
		if err != nil {
			return nil // Skip files we can't stat
		}
		if info.Size() > s.MaxFileSize {
			result.Warnings = append(result.Warnings, fmt.Sprintf("large file skipped: %s (%d bytes)", d.Name(), info.Size()))
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		// Scan for patterns
		matches := ScanForPatterns(string(content))
		for _, m := range matches {
			// Add relative path info
			relPath, _ := filepath.Rel(pluginPath, path)
			m.Match = fmt.Sprintf("[%s] %s", relPath, m.Match)
			result.Matches = append(result.Matches, m)
		}

		return nil
	})
}

// aggregateResults determines overall scan result from matches.
func (s *SecurityScanner) aggregateResults(result *ScanResult) {
	if len(result.Matches) == 0 {
		result.Allowed = true
		result.Confidence = 1.0
		result.Reason = "No adversarial patterns detected"
		return
	}

	// Count severity levels
	var highCount, mediumCount, lowCount int
	for _, m := range result.Matches {
		switch m.Pattern.Severity {
		case SeverityHigh:
			highCount++
		case SeverityMedium:
			mediumCount++
		case SeverityLow:
			lowCount++
		}
	}

	// Block on high severity matches
	if highCount > 0 {
		result.Allowed = false
		result.Confidence = 0.95
		result.Reason = fmt.Sprintf("Detected %d high-severity adversarial pattern(s)", highCount)
		return
	}

	// Block on multiple medium severity matches
	if mediumCount >= 3 {
		result.Allowed = false
		result.Confidence = 0.85
		result.Reason = fmt.Sprintf("Detected %d medium-severity adversarial patterns", mediumCount)
		return
	}

	// Warn on medium matches but allow
	if mediumCount > 0 {
		result.Allowed = true
		result.Confidence = 0.7
		result.Reason = fmt.Sprintf("Detected %d potentially concerning pattern(s) - review recommended", mediumCount)
		for _, m := range result.Matches {
			if m.Pattern.Severity == SeverityMedium {
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s: %s", m.Pattern.Name, m.Pattern.Description))
			}
		}
		return
	}

	// Low severity only - allow with minor warnings
	result.Allowed = true
	result.Confidence = 0.9
	result.Reason = "Minor patterns detected, likely benign"
}

// isTextFile checks if a file is likely a text file based on extension.
func isTextFile(name string) bool {
	textExts := []string{
		".md", ".txt", ".json", ".yaml", ".yml", ".toml",
		".go", ".py", ".js", ".ts", ".sh", ".bash",
		".html", ".css", ".xml", ".csv",
	}
	lower := strings.ToLower(name)
	for _, ext := range textExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	// Also check files without extension (like SKILL.md, README)
	if !strings.Contains(filepath.Base(name), ".") {
		return true
	}
	return false
}

// isSuspiciousExtension checks for potentially dangerous file types.
func isSuspiciousExtension(name string) bool {
	suspiciousExts := []string{
		".exe", ".dll", ".so", ".dylib", ".bin",
		".bat", ".cmd", ".ps1", ".vbs",
		".jar", ".class",
	}
	lower := strings.ToLower(name)
	for _, ext := range suspiciousExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}
