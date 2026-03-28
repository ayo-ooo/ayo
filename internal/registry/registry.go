package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// Entry represents a registered agent in the ayo registry.
type Entry struct {
	Name         string    `toml:"name"`
	Version      string    `toml:"version"`
	Description  string    `toml:"description"`
	SourcePath   string    `toml:"source_path"`
	BinaryPath   string    `toml:"binary_path"`
	Type         string    `toml:"type"` // "tool" or "conversational"
	RegisteredAt time.Time `toml:"registered_at"`
}

// Registry manages the collection of registered agents.
type Registry struct {
	Agents []Entry `toml:"agents"`
	path   string
}

// DefaultPath returns the default registry file path.
func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting config dir: %w", err)
	}
	return filepath.Join(configDir, "ayo", "registry.toml"), nil
}

// Load reads the registry from disk. Returns an empty registry if the file doesn't exist.
func Load() (*Registry, error) {
	path, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	return LoadFrom(path)
}

// LoadFrom reads the registry from a specific path.
func LoadFrom(path string) (*Registry, error) {
	r := &Registry{path: path}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return r, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading registry: %w", err)
	}

	if _, err := toml.Decode(string(data), r); err != nil {
		return nil, fmt.Errorf("parsing registry: %w", err)
	}

	return r, nil
}

// Save writes the registry to disk.
func (r *Registry) Save() error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0755); err != nil {
		return fmt.Errorf("creating registry directory: %w", err)
	}

	f, err := os.Create(r.path)
	if err != nil {
		return fmt.Errorf("creating registry file: %w", err)
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("writing registry: %w", err)
	}

	return nil
}

// Register adds or updates an agent entry.
func (r *Registry) Register(entry Entry) {
	entry.RegisteredAt = time.Now()

	for i, existing := range r.Agents {
		if existing.Name == entry.Name {
			r.Agents[i] = entry
			return
		}
	}
	r.Agents = append(r.Agents, entry)
}

// Remove deletes an agent entry by name. Returns true if found and removed.
func (r *Registry) Remove(name string) bool {
	for i, entry := range r.Agents {
		if entry.Name == name {
			r.Agents = append(r.Agents[:i], r.Agents[i+1:]...)
			return true
		}
	}
	return false
}

// Get returns an agent entry by name, or nil if not found.
func (r *Registry) Get(name string) *Entry {
	for i := range r.Agents {
		if r.Agents[i].Name == name {
			return &r.Agents[i]
		}
	}
	return nil
}

// List returns all registered agents, optionally filtered by type.
func (r *Registry) List(filterType string) []Entry {
	if filterType == "" {
		return r.Agents
	}
	var filtered []Entry
	for _, entry := range r.Agents {
		if entry.Type == filterType {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
