package mounts

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/alexcabrera/ayo/internal/providers"
)

func TestMountValidator_Validate(t *testing.T) {
	v := NewMountValidator()
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		mount   providers.Mount
		wantErr bool
	}{
		{
			name: "valid directory mount",
			mount: providers.Mount{
				Source:      tmpDir,
				Destination: "/workspace",
				Mode:        providers.MountModeBind,
			},
			wantErr: false,
		},
		{
			name: "empty source",
			mount: providers.Mount{
				Source:      "",
				Destination: "/workspace",
			},
			wantErr: true,
		},
		{
			name: "empty destination",
			mount: providers.Mount{
				Source:      tmpDir,
				Destination: "",
			},
			wantErr: true,
		},
		{
			name: "relative destination",
			mount: providers.Mount{
				Source:      tmpDir,
				Destination: "workspace",
			},
			wantErr: true,
		},
		{
			name: "nonexistent source",
			mount: providers.Mount{
				Source:      "/nonexistent/path/12345",
				Destination: "/workspace",
			},
			wantErr: true,
		},
		{
			name: "blocked root path",
			mount: providers.Mount{
				Source:      "/",
				Destination: "/workspace",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.mount)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMountValidator_VirtioFSRequiresDirectory(t *testing.T) {
	v := NewMountValidator()

	// Create a file
	tmpFile := filepath.Join(t.TempDir(), "testfile")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	mount := providers.Mount{
		Source:      tmpFile,
		Destination: "/workspace/file",
		Mode:        providers.MountModeVirtioFS,
	}

	err := v.Validate(mount)
	if err == nil {
		t.Error("Validate() should fail for virtiofs mount of file")
	}
}

func TestPrepareMount(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		mount   providers.Mount
		wantErr bool
	}{
		{
			name: "absolute path",
			mount: providers.Mount{
				Source:      tmpDir,
				Destination: "/workspace",
			},
			wantErr: false,
		},
		{
			name: "relative path",
			mount: providers.Mount{
				Source:      ".",
				Destination: "/workspace",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PrepareMount(tt.mount)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrepareMount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !filepath.IsAbs(result.Source) {
				t.Errorf("PrepareMount() source not absolute: %s", result.Source)
			}
		})
	}
}

func TestPrepareMount_HomeExpansion(t *testing.T) {
	home := os.Getenv("HOME")
	if home == "" {
		t.Skip("HOME not set")
	}

	mount := providers.Mount{
		Source:      "~/",
		Destination: "/home",
	}

	result, err := PrepareMount(mount)
	if err != nil {
		t.Fatalf("PrepareMount() error = %v", err)
	}

	if result.Source != home {
		t.Errorf("PrepareMount() source = %s, want %s", result.Source, home)
	}
}

func TestDefaultMountMode(t *testing.T) {
	mode := DefaultMountMode()

	// Should return a valid mode
	switch mode {
	case providers.MountModeVirtioFS, providers.MountModeBind:
		// Valid
	default:
		t.Errorf("DefaultMountMode() = %v, want virtiofs or bind", mode)
	}
}

func TestIsVirtioFSAvailable(t *testing.T) {
	// Just verify it doesn't panic
	_ = IsVirtioFSAvailable()
}

func TestAppleContainerMountArgs(t *testing.T) {
	tests := []struct {
		name  string
		mount providers.Mount
		want  []string
	}{
		{
			name: "virtiofs mount",
			mount: providers.Mount{
				Source:      "/host/path",
				Destination: "/container/path",
				Mode:        providers.MountModeVirtioFS,
			},
			want: []string{"--volume", "/host/path:/container/path"},
		},
		{
			name: "readonly mount",
			mount: providers.Mount{
				Source:      "/host/path",
				Destination: "/container/path",
				Mode:        providers.MountModeVirtioFS,
				ReadOnly:    true,
			},
			want: []string{"--volume", "/host/path:/container/path:ro"},
		},
		{
			name: "tmpfs",
			mount: providers.Mount{
				Destination: "/tmp/cache",
				Mode:        providers.MountModeTmpfs,
			},
			want: []string{"--tmpfs", "/tmp/cache"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AppleContainerMountArgs(tt.mount)
			if len(got) != len(tt.want) {
				t.Errorf("AppleContainerMountArgs() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("AppleContainerMountArgs()[%d] = %s, want %s", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLinuxMountArgs(t *testing.T) {
	tests := []struct {
		name  string
		mount providers.Mount
		want  []string
	}{
		{
			name: "bind mount",
			mount: providers.Mount{
				Source:      "/host/path",
				Destination: "/container/path",
				Mode:        providers.MountModeBind,
			},
			want: []string{"--bind=/host/path:/container/path"},
		},
		{
			name: "readonly bind",
			mount: providers.Mount{
				Source:      "/host/path",
				Destination: "/container/path",
				Mode:        providers.MountModeBind,
				ReadOnly:    true,
			},
			want: []string{"--bind-ro=/host/path:/container/path"},
		},
		{
			name: "tmpfs",
			mount: providers.Mount{
				Destination: "/tmp/cache",
				Mode:        providers.MountModeTmpfs,
			},
			want: []string{"--tmpfs=/tmp/cache"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LinuxMountArgs(tt.mount)
			if len(got) != len(tt.want) {
				t.Errorf("LinuxMountArgs() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("LinuxMountArgs()[%d] = %s, want %s", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestProjectMount(t *testing.T) {
	tmpDir := t.TempDir()

	mount, err := ProjectMount(tmpDir)
	if err != nil {
		t.Fatalf("ProjectMount() error = %v", err)
	}

	if mount.Destination != "/workspace" {
		t.Errorf("Destination = %s, want /workspace", mount.Destination)
	}
	if mount.ReadOnly {
		t.Error("ProjectMount should not be read-only")
	}
}

func TestReadOnlyProjectMount(t *testing.T) {
	tmpDir := t.TempDir()

	mount, err := ReadOnlyProjectMount(tmpDir)
	if err != nil {
		t.Fatalf("ReadOnlyProjectMount() error = %v", err)
	}

	if !mount.ReadOnly {
		t.Error("ReadOnlyProjectMount should be read-only")
	}
}

func TestHomeMount(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("HOME-based test not applicable on Windows")
	}

	home := os.Getenv("HOME")
	if home == "" {
		t.Skip("HOME not set")
	}

	// Test with nonexistent subdir
	_, err := HomeMount("nonexistent-subdir-12345", "/data")
	if err == nil {
		t.Error("HomeMount should fail for nonexistent subdir")
	}
}

func TestTmpfsMount(t *testing.T) {
	mount := TmpfsMount("/tmp/cache")

	if mount.Mode != providers.MountModeTmpfs {
		t.Errorf("Mode = %v, want tmpfs", mount.Mode)
	}
	if mount.Destination != "/tmp/cache" {
		t.Errorf("Destination = %s, want /tmp/cache", mount.Destination)
	}
}
