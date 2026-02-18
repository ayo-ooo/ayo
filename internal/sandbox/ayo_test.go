package sandbox

import (
	"os"
	"strings"
	"testing"

	"github.com/alexcabrera/ayo/internal/providers"
)

func TestEnsureAgentHome(t *testing.T) {
	handle := "@test-ayo-sandbox-agent"

	dir, err := EnsureAgentHome(handle)
	if err != nil {
		t.Fatalf("EnsureAgentHome() error = %v", err)
	}

	// Should end with agent name (without @)
	if !strings.HasSuffix(dir, "test-ayo-sandbox-agent") {
		t.Errorf("EnsureAgentHome() = %s, want to end with test-ayo-sandbox-agent", dir)
	}

	// Directory should exist
	info, err := os.Stat(dir)
	if err != nil {
		t.Errorf("directory should exist: %v", err)
	} else if !info.IsDir() {
		t.Error("path should be a directory")
	}

	// Cleanup
	os.RemoveAll(dir)
}

func TestAgentHomeMount(t *testing.T) {
	tests := []struct {
		handle            string
		wantContainerPath string
	}{
		{"@myagent", "/home/myagent"},
		{"@test-agent", "/home/test-agent"},
		{"myagent", "/home/myagent"},
	}

	for _, tt := range tests {
		t.Run(tt.handle, func(t *testing.T) {
			mount, containerPath, err := AgentHomeMount(tt.handle)
			if err != nil {
				t.Fatalf("AgentHomeMount() error = %v", err)
			}

			// Check container path
			if containerPath != tt.wantContainerPath {
				t.Errorf("containerPath = %s, want %s", containerPath, tt.wantContainerPath)
			}

			// Check mount destination
			if mount.Destination != tt.wantContainerPath {
				t.Errorf("mount.Destination = %s, want %s", mount.Destination, tt.wantContainerPath)
			}

			// Check mount mode
			if mount.Mode != providers.MountModeVirtioFS {
				t.Errorf("mount.Mode = %v, want VirtioFS", mount.Mode)
			}

			// Check mount is not read-only
			if mount.ReadOnly {
				t.Error("mount should not be read-only")
			}

			// Check source path exists
			info, err := os.Stat(mount.Source)
			if err != nil {
				t.Errorf("source directory should exist: %v", err)
			} else if !info.IsDir() {
				t.Error("source should be a directory")
			}

			// Cleanup
			os.RemoveAll(mount.Source)
		})
	}
}

func TestAgentHomeContainerPath(t *testing.T) {
	tests := []struct {
		handle string
		want   string
	}{
		{"@myagent", "/home/myagent"},
		{"@test-agent", "/home/test-agent"},
		{"myagent", "/home/myagent"},
		{"@ayo", "/home/ayo"},
		{"", "/home/"},
	}

	for _, tt := range tests {
		t.Run(tt.handle, func(t *testing.T) {
			got := AgentHomeContainerPath(tt.handle)
			if got != tt.want {
				t.Errorf("AgentHomeContainerPath(%q) = %s, want %s", tt.handle, got, tt.want)
			}
		})
	}
}
