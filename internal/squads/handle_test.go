package squads

import "testing"

func TestNormalizeHandle(t *testing.T) {
	tests := []struct {
		name     string
		handle   string
		expected string
	}{
		{
			name:     "empty string returns empty",
			handle:   "",
			expected: "",
		},
		{
			name:     "handle without prefix gets prefix added",
			handle:   "frontend",
			expected: "#frontend",
		},
		{
			name:     "handle with prefix unchanged",
			handle:   "#frontend",
			expected: "#frontend",
		},
		{
			name:     "single character handle",
			handle:   "a",
			expected: "#a",
		},
		{
			name:     "handle with hyphens",
			handle:   "my-squad",
			expected: "#my-squad",
		},
		{
			name:     "handle with underscores",
			handle:   "my_squad",
			expected: "#my_squad",
		},
		{
			name:     "handle with numbers",
			handle:   "squad123",
			expected: "#squad123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeHandle(tt.handle)
			if result != tt.expected {
				t.Errorf("NormalizeHandle(%q) = %q, want %q", tt.handle, result, tt.expected)
			}
		})
	}
}

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		name     string
		handle   string
		expected string
	}{
		{
			name:     "handle with prefix gets stripped",
			handle:   "#frontend",
			expected: "frontend",
		},
		{
			name:     "handle without prefix unchanged",
			handle:   "frontend",
			expected: "frontend",
		},
		{
			name:     "empty string",
			handle:   "",
			expected: "",
		},
		{
			name:     "just the prefix",
			handle:   "#",
			expected: "",
		},
		{
			name:     "multiple prefix chars only strips first",
			handle:   "##double",
			expected: "#double",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripPrefix(tt.handle)
			if result != tt.expected {
				t.Errorf("StripPrefix(%q) = %q, want %q", tt.handle, result, tt.expected)
			}
		})
	}
}

func TestIsSquadHandle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "handle with prefix is squad handle",
			input:    "#frontend",
			expected: true,
		},
		{
			name:     "handle without prefix is not squad handle",
			input:    "frontend",
			expected: false,
		},
		{
			name:     "empty string is not squad handle",
			input:    "",
			expected: false,
		},
		{
			name:     "just prefix is squad handle",
			input:    "#",
			expected: true,
		},
		{
			name:     "at symbol is not squad handle",
			input:    "@agent",
			expected: false,
		},
		{
			name:     "prefix in middle is not squad handle",
			input:    "not#squad",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSquadHandle(tt.input)
			if result != tt.expected {
				t.Errorf("IsSquadHandle(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateHandle(t *testing.T) {
	tests := []struct {
		name     string
		handle   string
		expected bool
	}{
		{
			name:     "valid handle without prefix",
			handle:   "frontend",
			expected: true,
		},
		{
			name:     "valid handle with prefix",
			handle:   "#frontend",
			expected: true,
		},
		{
			name:     "valid handle with hyphen",
			handle:   "my-squad",
			expected: true,
		},
		{
			name:     "valid handle with underscore",
			handle:   "my_squad",
			expected: true,
		},
		{
			name:     "valid handle with numbers",
			handle:   "squad123",
			expected: true,
		},
		{
			name:     "valid handle uppercase",
			handle:   "Frontend",
			expected: true,
		},
		{
			name:     "valid handle mixed case",
			handle:   "FrontEnd",
			expected: true,
		},
		{
			name:     "empty string is invalid",
			handle:   "",
			expected: false,
		},
		{
			name:     "just prefix is invalid",
			handle:   "#",
			expected: false,
		},
		{
			name:     "hyphen at start is invalid",
			handle:   "-squad",
			expected: false,
		},
		{
			name:     "hyphen at start with prefix is invalid",
			handle:   "#-squad",
			expected: false,
		},
		{
			name:     "underscore at start is invalid",
			handle:   "_squad",
			expected: false,
		},
		{
			name:     "space in handle is invalid",
			handle:   "my squad",
			expected: false,
		},
		{
			name:     "special char is invalid",
			handle:   "squad!",
			expected: false,
		},
		{
			name:     "dot is invalid",
			handle:   "squad.name",
			expected: false,
		},
		{
			name:     "slash is invalid",
			handle:   "squad/name",
			expected: false,
		},
		{
			name:     "single character is valid",
			handle:   "a",
			expected: true,
		},
		{
			name:     "single digit is valid",
			handle:   "1",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateHandle(tt.handle)
			if result != tt.expected {
				t.Errorf("ValidateHandle(%q) = %v, want %v", tt.handle, result, tt.expected)
			}
		})
	}
}
