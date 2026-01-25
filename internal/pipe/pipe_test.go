package pipe

import (
	"os"
	"testing"
)

func TestChainContext(t *testing.T) {
	// Save original env and restore after test
	original := os.Getenv(ChainContextEnvVar)
	defer os.Setenv(ChainContextEnvVar, original)

	t.Run("GetChainContext returns nil when not set", func(t *testing.T) {
		os.Unsetenv(ChainContextEnvVar)
		ctx := GetChainContext()
		if ctx != nil {
			t.Errorf("expected nil, got %+v", ctx)
		}
	})

	t.Run("GetChainContext parses valid context", func(t *testing.T) {
		os.Setenv(ChainContextEnvVar, `{"depth":2,"source":"@ayo.test","source_description":"Test agent"}`)
		ctx := GetChainContext()
		if ctx == nil {
			t.Fatal("expected context, got nil")
		}
		if ctx.Depth != 2 {
			t.Errorf("expected depth 2, got %d", ctx.Depth)
		}
		if ctx.Source != "@ayo.test" {
			t.Errorf("expected source @ayo.test, got %s", ctx.Source)
		}
		if ctx.SourceDescription != "Test agent" {
			t.Errorf("expected description 'Test agent', got %s", ctx.SourceDescription)
		}
	})

	t.Run("GetChainContext returns nil for invalid JSON", func(t *testing.T) {
		os.Setenv(ChainContextEnvVar, "not json")
		ctx := GetChainContext()
		if ctx != nil {
			t.Errorf("expected nil for invalid JSON, got %+v", ctx)
		}
	})
}
