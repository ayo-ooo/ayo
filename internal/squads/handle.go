package squads

import "strings"

// SquadPrefix is the prefix used to identify squad handles.
const SquadPrefix = "#"

// NormalizeHandle ensures a squad handle has the # prefix.
// If the handle already has the prefix, it is returned unchanged.
func NormalizeHandle(handle string) string {
	if handle == "" {
		return ""
	}
	if strings.HasPrefix(handle, SquadPrefix) {
		return handle
	}
	return SquadPrefix + handle
}

// StripPrefix removes the # prefix from a squad handle.
// If the handle doesn't have the prefix, it is returned unchanged.
func StripPrefix(handle string) string {
	return strings.TrimPrefix(handle, SquadPrefix)
}

// IsSquadHandle returns true if the string starts with #.
func IsSquadHandle(s string) bool {
	return strings.HasPrefix(s, SquadPrefix)
}

// ValidateHandle checks if a handle is valid.
// Valid handles are non-empty and contain only alphanumeric characters,
// hyphens, and underscores after the optional # prefix.
func ValidateHandle(handle string) bool {
	name := StripPrefix(handle)
	if name == "" {
		return false
	}
	for i, r := range name {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '-' || r == '_' {
			// Can't be first character
			if i == 0 {
				return false
			}
			continue
		}
		return false
	}
	return true
}
