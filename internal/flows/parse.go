package flows

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ParseYAMLFlow parses a YAML flow file from the given path.
func ParseYAMLFlow(path string) (*YAMLFlow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read flow file: %w", err)
	}

	flow, err := ParseYAMLFlowBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", filepath.Base(path), err)
	}

	return flow, nil
}

// ParseYAMLFlowBytes parses a YAML flow from raw bytes.
func ParseYAMLFlowBytes(data []byte) (*YAMLFlow, error) {
	var flow YAMLFlow
	if err := yaml.Unmarshal(data, &flow); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	return &flow, nil
}

// IsYAMLFlowFile returns true if the file extension indicates a YAML flow.
func IsYAMLFlowFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".yaml" || ext == ".yml"
}

// DiscoverYAMLFlows finds all YAML flow files in a directory.
func DiscoverYAMLFlows(dir string) ([]*YAMLFlow, error) {
	var flows []*YAMLFlow

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !IsYAMLFlowFile(entry.Name()) {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		flow, err := ParseYAMLFlow(path)
		if err != nil {
			// Skip invalid flows but continue discovering
			continue
		}

		if err := flow.Validate(); err != nil {
			// Skip invalid flows
			continue
		}

		flows = append(flows, flow)
	}

	return flows, nil
}
