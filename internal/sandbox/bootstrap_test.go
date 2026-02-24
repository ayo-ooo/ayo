package sandbox

import (
	"testing"
)

func TestGetAyodBinaryPath(t *testing.T) {
	// This test verifies that getAyodBinaryPath doesn't panic
	// The actual path lookup depends on the build state
	path, err := getAyodBinaryPath()
	if err != nil {
		// Expected if binary hasn't been built
		t.Logf("ayod binary not found (expected in dev): %v", err)
		return
	}
	if path == "" {
		t.Error("expected non-empty path when no error")
	}
	t.Logf("ayod binary found: %s", path)
}
