package memory

import (
	"testing"
)

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "with heading",
			content:  "# My Title\n\nSome content here.",
			expected: "My Title",
		},
		{
			name:     "no heading short",
			content:  "Just some text",
			expected: "Just some text",
		},
		{
			name:     "no heading long",
			content:  "This is a very long piece of content that exceeds fifty characters and should be truncated",
			expected: "This is a very long piece of content that exceeds ...",
		},
		{
			name:     "empty",
			content:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTitle(tt.content)
			if result != tt.expected {
				t.Errorf("extractTitle() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractTags(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "with tags",
			content:  "# Title\n\nTags: golang, testing, memory\n\nContent here.",
			expected: []string{"golang", "testing", "memory"},
		},
		{
			name:     "no tags",
			content:  "# Title\n\nContent without tags.",
			expected: nil,
		},
		{
			name:     "empty tags",
			content:  "Tags: \n\nContent",
			expected: []string{},
		},
		{
			name:     "single tag",
			content:  "Tags: single\n\nContent",
			expected: []string{"single"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTags(tt.content)
			if len(result) != len(tt.expected) {
				t.Errorf("extractTags() length = %d, want %d", len(result), len(tt.expected))
				return
			}
			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("extractTags()[%d] = %q, want %q", i, tag, tt.expected[i])
				}
			}
		})
	}
}
