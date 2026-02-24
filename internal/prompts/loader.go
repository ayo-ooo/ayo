// Package prompts provides runtime-loaded prompt templates for ayo.
// All prompts are sourced from ~/.local/share/ayo/prompts/ at runtime,
// with embedded defaults for installation.
package prompts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// PromptLoader loads prompts from the filesystem with caching.
type PromptLoader struct {
	baseDir string
	cache   map[string]string
	mu      sync.RWMutex
}

// DefaultBaseDir returns the default prompts directory.
func DefaultBaseDir() string {
	// Use XDG_DATA_HOME if set, otherwise ~/.local/share
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, _ := os.UserHomeDir()
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "ayo", "prompts")
}

// NewPromptLoader creates a new prompt loader.
func NewPromptLoader() *PromptLoader {
	return &PromptLoader{
		baseDir: DefaultBaseDir(),
		cache:   make(map[string]string),
	}
}

// NewPromptLoaderWithDir creates a loader with a custom base directory.
func NewPromptLoaderWithDir(baseDir string) *PromptLoader {
	return &PromptLoader{
		baseDir: baseDir,
		cache:   make(map[string]string),
	}
}

// Load returns prompt content, with caching.
// The path is relative to the base directory (e.g., "guardrails/default.md").
func (l *PromptLoader) Load(path string) (string, error) {
	l.mu.RLock()
	if cached, ok := l.cache[path]; ok {
		l.mu.RUnlock()
		return cached, nil
	}
	l.mu.RUnlock()

	fullPath := filepath.Join(l.baseDir, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("prompt not found: %s", path)
	}

	result := strings.TrimSpace(string(content))

	l.mu.Lock()
	l.cache[path] = result
	l.mu.Unlock()

	return result, nil
}

// MustLoad loads a prompt and panics if not found.
// Use this for required prompts.
func (l *PromptLoader) MustLoad(path string) string {
	content, err := l.Load(path)
	if err != nil {
		panic(fmt.Sprintf("required prompt missing: %s", path))
	}
	return content
}

// LoadOrDefault loads a prompt, returning the default if not found.
func (l *PromptLoader) LoadOrDefault(path, defaultContent string) string {
	content, err := l.Load(path)
	if err != nil {
		return defaultContent
	}
	return content
}

// LoadWithEmbeddedFallback loads a prompt from filesystem, falling back to embedded defaults.
// This is the preferred method as it avoids duplicating prompt content in Go code.
func (l *PromptLoader) LoadWithEmbeddedFallback(path string) string {
	// Try filesystem first
	content, err := l.Load(path)
	if err == nil {
		return content
	}

	// Fall back to embedded
	return EmbeddedDefault(path)
}

// EmbeddedDefault returns the embedded default for a prompt path.
// Returns empty string if not found in embedded defaults.
func EmbeddedDefault(path string) string {
	content, err := embeddedPrompts.ReadFile("defaults/" + path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

// Exists checks if a prompt file exists.
func (l *PromptLoader) Exists(path string) bool {
	fullPath := filepath.Join(l.baseDir, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// Refresh clears the cache.
func (l *PromptLoader) Refresh() {
	l.mu.Lock()
	l.cache = make(map[string]string)
	l.mu.Unlock()
}

// List returns all prompt files in the base directory.
func (l *PromptLoader) List() ([]string, error) {
	var prompts []string
	err := filepath.Walk(l.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".txt")) {
			rel, _ := filepath.Rel(l.baseDir, path)
			prompts = append(prompts, rel)
		}
		return nil
	})
	return prompts, err
}

// BaseDir returns the base directory.
func (l *PromptLoader) BaseDir() string {
	return l.baseDir
}

// Global default loader instance
var defaultLoader *PromptLoader
var defaultLoaderOnce sync.Once

// Default returns the global prompt loader.
func Default() *PromptLoader {
	defaultLoaderOnce.Do(func() {
		defaultLoader = NewPromptLoader()
	})
	return defaultLoader
}

// Load loads a prompt using the global loader.
func Load(path string) (string, error) {
	return Default().Load(path)
}

// MustLoad loads a prompt using the global loader, panicking if not found.
func MustLoad(path string) string {
	return Default().MustLoad(path)
}

// Well-known prompt paths
const (
	PathSystemBase       = "system/base.md"
	PathSystemToolUsage  = "system/tool-usage.md"
	PathSystemMemory     = "system/memory-usage.md"
	PathSystemPlanning   = "system/planning.md"
	PathGuardrailsDefault = "guardrails/default.md"
	PathGuardrailsSafety = "guardrails/safety.md"
	PathSandwichPrefix   = "sandwich/prefix.md"
	PathSandwichSuffix   = "sandwich/suffix.md"
)
