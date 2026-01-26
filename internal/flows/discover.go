package flows

import (
	"os"
	"path/filepath"
	"strings"
)

// Discover finds all flows in the given directories.
// Directories are searched in order; the first flow found with a given name wins.
func Discover(dirs []string) ([]Flow, error) {
	seen := make(map[string]bool)
	var flows []Flow

	for _, dir := range dirs {
		source := sourceFromPath(dir)
		discovered, err := discoverInDir(dir, source)
		if err != nil {
			// Skip directories that don't exist or can't be read
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		for _, f := range discovered {
			if !seen[f.Name] {
				seen[f.Name] = true
				flows = append(flows, f)
			}
		}
	}

	return flows, nil
}

// DiscoverOne loads a single flow from a path.
// The path can be a .sh file or a directory containing flow.sh.
func DiscoverOne(path string) (*Flow, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return loadPackagedFlow(path, FlowSourceUser)
	}
	return loadSimpleFlow(path, FlowSourceUser)
}

// discoverInDir finds all flows in a single directory.
func discoverInDir(dir string, source FlowSource) ([]Flow, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var flows []Flow

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Check for packaged flow (dir with flow.sh)
			flow, err := loadPackagedFlow(fullPath, source)
			if err != nil {
				// Not a valid flow package, skip
				continue
			}
			flows = append(flows, *flow)
		} else if strings.HasSuffix(entry.Name(), ".sh") {
			// Check for simple flow
			flow, err := loadSimpleFlow(fullPath, source)
			if err != nil {
				// Not a valid flow, skip
				continue
			}
			flows = append(flows, *flow)
		}
	}

	return flows, nil
}

// loadSimpleFlow loads a flow from a single .sh file.
func loadSimpleFlow(path string, source FlowSource) (*Flow, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	raw, err := ParseFrontmatter(content)
	if err != nil {
		return nil, err
	}

	if err := ValidateFrontmatter(raw.Frontmatter); err != nil {
		return nil, err
	}

	flow := &Flow{
		Name:        raw.Frontmatter["name"],
		Description: raw.Frontmatter["description"],
		Path:        path,
		Dir:         filepath.Dir(path),
		Source:      source,
		Metadata: FlowMetadata{
			Version: raw.Frontmatter["version"],
			Author:  raw.Frontmatter["author"],
		},
		Raw: raw,
	}

	return flow, nil
}

// loadPackagedFlow loads a flow from a directory containing flow.sh.
func loadPackagedFlow(dir string, source FlowSource) (*Flow, error) {
	flowPath := filepath.Join(dir, "flow.sh")

	flow, err := loadSimpleFlow(flowPath, source)
	if err != nil {
		return nil, err
	}

	// Check for schemas
	inputSchema := filepath.Join(dir, "input.jsonschema")
	if _, err := os.Stat(inputSchema); err == nil {
		flow.InputSchemaPath = inputSchema
	}

	outputSchema := filepath.Join(dir, "output.jsonschema")
	if _, err := os.Stat(outputSchema); err == nil {
		flow.OutputSchemaPath = outputSchema
	}

	return flow, nil
}

// sourceFromPath determines the FlowSource based on the directory path.
// This is a simple heuristic that can be overridden by callers.
func sourceFromPath(dir string) FlowSource {
	// Check for common patterns
	if strings.Contains(dir, ".local/share") {
		return FlowSourceBuiltin
	}
	if strings.Contains(dir, ".config") {
		return FlowSourceUser
	}
	if strings.Contains(dir, ".ayo/flows") {
		return FlowSourceProject
	}
	return FlowSourceUser
}
