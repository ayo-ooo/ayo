package publish

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
)

func TestValidateDestination(t *testing.T) {
	projectPath := "/home/user/project"

	tests := []struct {
		name     string
		dest     string
		allowed  []string
		wantErr  bool
	}{
		{
			name:    "within project",
			dest:    "/home/user/project/output",
			wantErr: false,
		},
		{
			name:    "project root",
			dest:    "/home/user/project",
			wantErr: false,
		},
		{
			name:    "outside project",
			dest:    "/home/user/other",
			wantErr: true,
		},
		{
			name:    "allowed list match",
			dest:    "/tmp/output",
			allowed: []string{"/tmp"},
			wantErr: false,
		},
		{
			name:    "allowed wildcard",
			dest:    "/anywhere/at/all",
			allowed: []string{"*"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ToolConfig{
				HostProjectPath:     projectPath,
				AllowedDestinations: tt.allowed,
			}
			err := validateDestination(tt.dest, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDestination(%q) error = %v, wantErr %v", tt.dest, err, tt.wantErr)
			}
		})
	}
}

func TestValidateFile(t *testing.T) {
	cfg := ToolConfig{
		BlockedFilePatterns: DefaultBlockedPatterns(),
	}

	tests := []struct {
		file    string
		wantErr bool
	}{
		{"output.txt", false},
		{"dist/bundle.js", false},
		{"../outside", true},       // path traversal
		{".env", true},              // blocked
		{".git/config", true},       // blocked
		{"secrets/api.key", true},   // blocked
		{"normal.key", true},        // blocked by *.key
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			err := validateFile(tt.file, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFile(%q) error = %v, wantErr %v", tt.file, err, tt.wantErr)
			}
		})
	}
}

func TestValidateFile_AllowedPatterns(t *testing.T) {
	cfg := ToolConfig{
		AllowedFilePatterns: []string{"*.go", "*.md"},
		BlockedFilePatterns: DefaultBlockedPatterns(),
	}

	tests := []struct {
		file    string
		wantErr bool
	}{
		{"main.go", false},
		{"README.md", false},
		{"config.yaml", true},  // not in allowed
		{"script.sh", true},    // not in allowed
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			err := validateFile(tt.file, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFile(%q) error = %v, wantErr %v", tt.file, err, tt.wantErr)
			}
		})
	}
}

func TestDefaultBlockedPatterns(t *testing.T) {
	patterns := DefaultBlockedPatterns()

	// Check essential patterns
	expected := []string{".env", ".git/*"}
	for _, exp := range expected {
		found := false
		for _, p := range patterns {
			if p == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pattern %q not found in defaults", exp)
		}
	}
}

func TestPublishResult_String(t *testing.T) {
	r := PublishResult{
		Published: []string{"file1.go", "file2.go"},
		Skipped:   []string{"file3.go: already exists"},
		Errors:    []string{"file4.go: not found"},
	}

	s := r.String()

	if s == "" {
		t.Error("String() returned empty string")
	}
	if !contains(s, "file1.go") {
		t.Error("String() missing published file")
	}
	if !contains(s, "already exists") {
		t.Error("String() missing skipped reason")
	}
	if !contains(s, "not found") {
		t.Error("String() missing error")
	}
}

func TestPublishResult_Empty(t *testing.T) {
	r := PublishResult{}
	s := r.String()
	if s != "No files published" {
		t.Errorf("String() = %q, want 'No files published'", s)
	}
}

func TestNewPublishTool(t *testing.T) {
	mockProvider := sandbox.NewMockProvider()
	ctx := context.Background()

	// Create sandbox
	sb, err := mockProvider.Create(ctx, providers.SandboxCreateOptions{
		Name: "test-sandbox",
	})
	if err != nil {
		t.Fatal(err)
	}

	hostDir := t.TempDir()

	cfg := ToolConfig{
		Provider:        mockProvider,
		SandboxID:       sb.ID,
		HostProjectPath: hostDir,
	}

	tool := NewPublishTool(cfg)

	// Verify tool is created
	_ = tool
}

func TestCopyFileFromSandbox(t *testing.T) {
	mockProvider := sandbox.NewMockProvider()
	ctx := context.Background()

	// Create sandbox
	sb, err := mockProvider.Create(ctx, providers.SandboxCreateOptions{
		Name: "test-sandbox",
	})
	if err != nil {
		t.Fatal(err)
	}

	hostDir := t.TempDir()
	hostPath := filepath.Join(hostDir, "output.txt")

	// Mock exec to return file content
	mockProvider.ExecFunc = func(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
		if contains(opts.Command, "cat") {
			return providers.ExecResult{
				Stdout:   "file content here",
				ExitCode: 0,
			}, nil
		}
		if contains(opts.Command, "stat") {
			return providers.ExecResult{
				Stdout:   "644\n",
				ExitCode: 0,
			}, nil
		}
		return providers.ExecResult{ExitCode: 0}, nil
	}

	err = copyFileFromSandbox(ctx, mockProvider, sb.ID, "/workspace/output.txt", hostPath)
	if err != nil {
		t.Fatalf("copyFileFromSandbox failed: %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(hostPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(content) != "file content here" {
		t.Errorf("content = %q, want 'file content here'", string(content))
	}
}

func TestCopyFileFromSandbox_NotFound(t *testing.T) {
	mockProvider := sandbox.NewMockProvider()
	ctx := context.Background()

	// Create sandbox
	sb, err := mockProvider.Create(ctx, providers.SandboxCreateOptions{
		Name: "test-sandbox",
	})
	if err != nil {
		t.Fatal(err)
	}

	hostDir := t.TempDir()
	hostPath := filepath.Join(hostDir, "missing.txt")

	// Mock exec to return error
	mockProvider.ExecFunc = func(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
		return providers.ExecResult{
			Stderr:   "cat: /workspace/missing.txt: No such file or directory",
			ExitCode: 1,
		}, nil
	}

	err = copyFileFromSandbox(ctx, mockProvider, sb.ID, "/workspace/missing.txt", hostPath)
	if err == nil {
		t.Error("copyFileFromSandbox should have failed for missing file")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
