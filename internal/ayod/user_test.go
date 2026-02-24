package ayod

import (
	"testing"
)

func TestSanitizeUsername(t *testing.T) {
	tests := []struct {
		handle string
		want   string
	}{
		{"@ayo", "ayo"},
		{"@crush", "crush"},
		{"@Reviewer", "reviewer"},
		{"ayo", "ayo"},
		{"@test-agent", "test-agent"},
		{"@test_agent", "test_agent"},
		{"@Agent123", "agent123"},
		{"@123agent", "agent"}, // Can't start with number
		{"@!!!", "agent"},      // All invalid chars
		{"@Test Agent", "test_agent"},
		{"@very-long-username-that-exceeds-unix-limits-of-32-characters", "very-long-username-that-exceeds"},
		{"", "agent"},
		{"@", "agent"},
	}

	for _, tt := range tests {
		t.Run(tt.handle, func(t *testing.T) {
			got := SanitizeUsername(tt.handle)
			if got != tt.want {
				t.Errorf("SanitizeUsername(%q) = %q, want %q", tt.handle, got, tt.want)
			}
		})
	}
}

func TestParentDir(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/home/user/file.txt", "/home/user"},
		{"/home/user", "/home"},
		{"/home", "/"},
		{"/", "/"},
		{"file.txt", "."},
		{"dir/file.txt", "dir"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := parentDir(tt.path)
			if got != tt.want {
				t.Errorf("parentDir(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
