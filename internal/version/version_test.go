package version

import (
	"testing"
)

func TestVersionFormat(t *testing.T) {
	// Version should be in semver format
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Should start with a digit (semver)
	if Version[0] < '0' || Version[0] > '9' {
		t.Errorf("Version %q should start with a digit", Version)
	}
}

func TestVersionValue(t *testing.T) {
	// Verify current version matches expected
	expected := "0.3.0"
	if Version != expected {
		t.Errorf("Version = %q, want %q", Version, expected)
	}
}
