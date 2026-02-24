package prompts

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed defaults/*
var embeddedPrompts embed.FS

// InstallDefaultPrompts installs embedded prompts to the prompts directory.
// If force is true, overwrites existing files.
func InstallDefaultPrompts(force bool) error {
	promptsDir := DefaultBaseDir()

	// Check if already installed
	if !force && dirExists(promptsDir) {
		return nil
	}

	return fs.WalkDir(embeddedPrompts, "defaults", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root "defaults" directory
		if path == "defaults" {
			return nil
		}

		// Get relative path without "defaults/" prefix
		relPath, err := filepath.Rel("defaults", path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(promptsDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Skip if file exists and not forcing
		if !force && fileExists(destPath) {
			return nil
		}

		content, err := embeddedPrompts.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading embedded prompt %s: %w", path, err)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		return os.WriteFile(destPath, content, 0644)
	})
}

// ValidatePrompts checks that all required prompts exist.
func ValidatePrompts() []string {
	var missing []string

	required := []string{
		PathGuardrailsDefault,
		PathSandwichPrefix,
		PathSandwichSuffix,
	}

	loader := Default()
	for _, path := range required {
		if !loader.Exists(path) {
			missing = append(missing, path)
		}
	}

	return missing
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
