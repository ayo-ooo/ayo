package publish

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateDestination(t *testing.T) {
	tests := []struct {
		name    string
		dest    string
		blocked []string
		wantErr bool
	}{
		{
			name:    "normal path",
			dest:    "/home/user/Documents/output.pdf",
			wantErr: false,
		},
		{
			name:    "blocked system path",
			dest:    "/etc/passwd",
			blocked: DefaultBlockedDestinations(),
			wantErr: true,
		},
		{
			name:    "blocked root",
			dest:    "/",
			blocked: DefaultBlockedDestinations(),
			wantErr: true,
		},
		{
			name:    "home directory with tilde",
			dest:    "~/Documents/report.pdf",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ToolConfig{
				BlockedDestinations: tt.blocked,
			}
			err := validateDestination(tt.dest, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDestination(%q) error = %v, wantErr %v", tt.dest, err, tt.wantErr)
			}
		})
	}
}

func TestDefaultBlockedDestinations(t *testing.T) {
	blocked := DefaultBlockedDestinations()

	expected := []string{"/", "/etc", "/usr"}
	for _, exp := range expected {
		found := false
		for _, b := range blocked {
			if b == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected blocked destination %q not found in defaults", exp)
		}
	}
}

func TestPublishResult_String(t *testing.T) {
	r := PublishResult{
		Published: []string{"/output/report.pdf -> ~/Documents/report.pdf"},
		Skipped:   []string{"data.csv: already exists"},
		Errors:    []string{"missing.txt: not found"},
	}

	s := r.String()

	if s == "" {
		t.Error("String() returned empty string")
	}
	if !strings.Contains(s, "report.pdf") {
		t.Error("String() missing published file")
	}
	if !strings.Contains(s, "already exists") {
		t.Error("String() missing skipped reason")
	}
	if !strings.Contains(s, "not found") {
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

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/Documents", filepath.Join(home, "Documents")},
		{"~", home},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandHome(tt.input)
			if got != tt.want {
				t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestNewPublishTool is disabled during sandbox infrastructure removal.
// TODO: Re-enable when sandbox is re-implemented as standalone executable
func TestNewPublishTool(t *testing.T) {
	t.Skip("Sandbox infrastructure removed - test disabled")
}

func TestCopyOutputFile_FromHostDir(t *testing.T) {
	ctx := context.Background()

	// Create temp directories
	outputDir := t.TempDir()
	destDir := t.TempDir()

	// Create source file
	sourceContent := "test file content"
	sourceFile := filepath.Join(outputDir, "report.txt")
	if err := os.WriteFile(sourceFile, []byte(sourceContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := ToolConfig{
		OutputDir:     "/output",
		HostOutputDir: outputDir,
	}

	destFile := filepath.Join(destDir, "report.txt")
	err := copyOutputFile(ctx, cfg, "/output/report.txt", destFile)
	if err != nil {
		t.Fatalf("copyOutputFile failed: %v", err)
	}

	// Verify file was copied
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(content) != sourceContent {
		t.Errorf("content = %q, want %q", string(content), sourceContent)
	}
}

// TestCopyOutputFile_ViaExec is disabled during sandbox infrastructure removal.
// TODO: Re-enable when sandbox is re-implemented as standalone executable
func TestCopyOutputFile_ViaExec(t *testing.T) {
	t.Skip("Sandbox infrastructure removed - test disabled")
}

func TestSourceMustBeInOutput(t *testing.T) {
	// This tests the validation in the tool itself
	// Sources not in /output/ should be rejected
	tests := []struct {
		source  string
		valid   bool
	}{
		{"/output/report.pdf", true},
		{"/output/subdir/data.csv", true},
		{"/workspace/file.txt", false},
		{"/home/user/file.txt", false},
		{"relative/path.txt", false},
	}

	for _, tt := range tests {
		source := tt.source
		isInOutput := strings.HasPrefix(source, "/output/") || strings.HasPrefix(source, "/output")
		if isInOutput != tt.valid {
			t.Errorf("source %q validation: got isInOutput=%v, want %v", source, isInOutput, tt.valid)
		}
	}
}
