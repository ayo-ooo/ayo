package flows

import (
	"errors"
	"strings"
)

// Frontmatter parsing errors.
var (
	ErrNoShebang    = errors.New("flow must start with #!/usr/bin/env bash")
	ErrNoFlowMarker = errors.New("flow must contain # ayo:flow marker")
	ErrMissingName  = errors.New("flow missing required field: name")
	ErrMissingDesc  = errors.New("flow missing required field: description")
)

const (
	shebangLine = "#!/usr/bin/env bash"
	flowMarker  = "# ayo:flow"
)

// ParseFrontmatter extracts metadata and script from a flow file.
// It expects the file to start with a shebang, followed by the ayo:flow marker,
// then key: value metadata lines, and finally the script content.
func ParseFrontmatter(content []byte) (FlowRaw, error) {
	raw := FlowRaw{
		Frontmatter: make(map[string]string),
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		return raw, ErrNoShebang
	}

	// First line must be shebang
	if strings.TrimSpace(lines[0]) != shebangLine {
		return raw, ErrNoShebang
	}

	// Find the flow marker
	markerIdx := -1
	for i := 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == flowMarker {
			markerIdx = i
			break
		}
		// Skip empty lines and comments before marker
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Non-comment line before marker means not a flow
		return raw, ErrNoFlowMarker
	}

	if markerIdx == -1 {
		return raw, ErrNoFlowMarker
	}

	// Parse metadata after marker
	scriptStartIdx := len(lines) // default to end if no script
	for i := markerIdx + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])

		// Empty lines are allowed in frontmatter
		if trimmed == "" {
			continue
		}

		// Check if it's a metadata comment
		if strings.HasPrefix(trimmed, "#") {
			rest := strings.TrimPrefix(trimmed, "#")
			rest = strings.TrimSpace(rest)

			if idx := strings.Index(rest, ":"); idx > 0 {
				key := strings.TrimSpace(rest[:idx])
				value := strings.TrimSpace(rest[idx+1:])
				if isValidMetadataKey(key) {
					raw.Frontmatter[key] = value
					continue
				}
			}
			// Comment that's not metadata - this is script start
			scriptStartIdx = i
			break
		}

		// Non-comment line - script starts here
		scriptStartIdx = i
		break
	}

	// Extract script, trimming leading empty lines
	if scriptStartIdx < len(lines) {
		scriptLines := lines[scriptStartIdx:]
		
		// Trim leading empty lines from script
		start := 0
		for start < len(scriptLines) && strings.TrimSpace(scriptLines[start]) == "" {
			start++
		}
		
		if start < len(scriptLines) {
			raw.Script = strings.Join(scriptLines[start:], "\n")
		}
	}

	return raw, nil
}

// ValidateFrontmatter checks that required fields are present.
func ValidateFrontmatter(fm map[string]string) error {
	if fm["name"] == "" {
		return ErrMissingName
	}
	if fm["description"] == "" {
		return ErrMissingDesc
	}
	return nil
}

// isValidMetadataKey checks if a string is a valid frontmatter key.
// Keys should be lowercase alphanumeric with optional hyphens/underscores.
func isValidMetadataKey(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}
