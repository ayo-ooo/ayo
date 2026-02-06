package filerequest

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
)

func TestValidatePath(t *testing.T) {
	cfg := ToolConfig{
		BlockedPatterns: DefaultBlockedPatterns(),
	}

	tests := []struct {
		path    string
		wantErr bool
	}{
		{"src/main.go", false},
		{"README.md", false},
		{"../outside", true}, // path traversal
		{".env", true},       // blocked
		{".env.local", true}, // blocked
		{"secrets/api.key", true},
		{".git/config", true},
		{"config/app.yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			err := validatePath(tt.path, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePath_AllowedPrefixes(t *testing.T) {
	cfg := ToolConfig{
		BlockedPatterns: DefaultBlockedPatterns(),
		AllowedPrefixes: []string{"src/", "lib/"},
	}

	tests := []struct {
		path    string
		wantErr bool
	}{
		{"src/main.go", false},
		{"src/util/helpers.go", false},
		{"lib/utils.go", false},
		{"config/app.yaml", true}, // not in allowed prefixes
		{"README.md", true},       // not in allowed prefixes
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			err := validatePath(tt.path, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestDefaultBlockedPatterns(t *testing.T) {
	patterns := DefaultBlockedPatterns()

	// Check that essential patterns are present
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

func TestFileRequestResult_String(t *testing.T) {
	r := FileRequestResult{
		Copied:  []string{"file1.go", "file2.go"},
		Skipped: []string{"file3.go: already exists"},
		Errors:  []string{"file4.go: permission denied"},
	}

	s := r.String()

	if s == "" {
		t.Error("String() returned empty string")
	}
	if !contains(s, "file1.go") {
		t.Error("String() missing copied file")
	}
	if !contains(s, "already exists") {
		t.Error("String() missing skipped reason")
	}
	if !contains(s, "permission denied") {
		t.Error("String() missing error")
	}
}

func TestFileRequestResult_Empty(t *testing.T) {
	r := FileRequestResult{}
	s := r.String()
	if s != "No files requested" {
		t.Errorf("String() = %q, want 'No files requested'", s)
	}
}

func TestNewFileRequestTool(t *testing.T) {
	mockProvider := sandbox.NewMockProvider()
	ctx := context.Background()

	// Create sandbox
	sb, err := mockProvider.Create(ctx, providers.SandboxCreateOptions{
		Name: "test-sandbox",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create temp host directory with test files
	hostDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(hostDir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Track exec calls
	mockProvider.ExecFunc = func(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
		return providers.ExecResult{ExitCode: 0}, nil
	}

	cfg := ToolConfig{
		Provider:        mockProvider,
		SandboxID:       sb.ID,
		HostProjectPath: hostDir,
	}

	tool := NewFileRequestTool(cfg)

	// Verify tool is created (it's an interface, can't easily check name without calling)
	_ = tool
}

func TestCopyFileToSandbox(t *testing.T) {
	mockProvider := sandbox.NewMockProvider()
	ctx := context.Background()

	// Create sandbox
	sb, err := mockProvider.Create(ctx, providers.SandboxCreateOptions{
		Name: "test-sandbox",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create temp file
	hostDir := t.TempDir()
	testFile := filepath.Join(hostDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	// Track exec calls
	var execCalls []string
	mockProvider.ExecFunc = func(ctx context.Context, id string, opts providers.ExecOptions) (providers.ExecResult, error) {
		execCalls = append(execCalls, opts.Command)
		return providers.ExecResult{ExitCode: 0}, nil
	}

	err = copyFileToSandbox(ctx, mockProvider, sb.ID, testFile, "/workspace/test.txt")
	if err != nil {
		t.Fatalf("copyFileToSandbox failed: %v", err)
	}

	// Verify mkdir and cat commands were called
	if len(execCalls) < 2 {
		t.Errorf("Expected at least 2 exec calls, got %d", len(execCalls))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
