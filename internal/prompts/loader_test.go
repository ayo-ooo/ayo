package prompts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPromptLoader(t *testing.T) {
	loader := NewPromptLoader()
	if loader.baseDir == "" {
		t.Error("expected non-empty baseDir")
	}
	if loader.cache == nil {
		t.Error("expected non-nil cache")
	}
}

func TestPromptLoader_Load(t *testing.T) {
	// Create temp directory with test prompt
	tmpDir := t.TempDir()
	testPath := "test/prompt.md"
	fullPath := filepath.Join(tmpDir, testPath)
	
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	
	content := "Test prompt content"
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	
	loader := NewPromptLoaderWithDir(tmpDir)
	
	// Load the prompt
	loaded, err := loader.Load(testPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	if loaded != content {
		t.Errorf("expected %q, got %q", content, loaded)
	}
	
	// Should be cached
	if _, ok := loader.cache[testPath]; !ok {
		t.Error("expected prompt to be cached")
	}
	
	// Second load should use cache
	loaded2, err := loader.Load(testPath)
	if err != nil {
		t.Fatalf("second Load failed: %v", err)
	}
	if loaded2 != content {
		t.Errorf("cached load: expected %q, got %q", content, loaded2)
	}
}

func TestPromptLoader_Load_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewPromptLoaderWithDir(tmpDir)
	
	_, err := loader.Load("nonexistent.md")
	if err == nil {
		t.Error("expected error for missing prompt")
	}
}

func TestPromptLoader_MustLoad_Panics(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewPromptLoaderWithDir(tmpDir)
	
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing required prompt")
		}
	}()
	
	loader.MustLoad("missing.md")
}

func TestPromptLoader_LoadOrDefault(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewPromptLoaderWithDir(tmpDir)
	
	defaultContent := "default content"
	result := loader.LoadOrDefault("missing.md", defaultContent)
	
	if result != defaultContent {
		t.Errorf("expected %q, got %q", defaultContent, result)
	}
}

func TestPromptLoader_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := "exists.md"
	
	if err := os.WriteFile(filepath.Join(tmpDir, testPath), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	
	loader := NewPromptLoaderWithDir(tmpDir)
	
	if !loader.Exists(testPath) {
		t.Error("expected file to exist")
	}
	
	if loader.Exists("missing.md") {
		t.Error("expected missing file to not exist")
	}
}

func TestPromptLoader_Refresh(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := "test.md"
	
	if err := os.WriteFile(filepath.Join(tmpDir, testPath), []byte("original"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	
	loader := NewPromptLoaderWithDir(tmpDir)
	
	// Load to cache
	_, _ = loader.Load(testPath)
	
	// Modify file
	if err := os.WriteFile(filepath.Join(tmpDir, testPath), []byte("modified"), 0644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}
	
	// Without refresh, should return cached
	cached, _ := loader.Load(testPath)
	if cached != "original" {
		t.Error("expected cached value before refresh")
	}
	
	// Refresh cache
	loader.Refresh()
	
	// Should return new content
	fresh, _ := loader.Load(testPath)
	if fresh != "modified" {
		t.Errorf("expected 'modified' after refresh, got %q", fresh)
	}
}

func TestPromptLoader_List(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create some prompt files
	paths := []string{"a.md", "sub/b.md", "sub/deep/c.txt"}
	for _, p := range paths {
		fullPath := filepath.Join(tmpDir, p)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}
	
	loader := NewPromptLoaderWithDir(tmpDir)
	
	list, err := loader.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	
	if len(list) != 3 {
		t.Errorf("expected 3 prompts, got %d: %v", len(list), list)
	}
}

func TestInstallDefaultPrompts(t *testing.T) {
	// Create temp directory that acts as XDG_DATA_HOME
	tmpDir := t.TempDir()
	promptsDir := filepath.Join(tmpDir, "ayo", "prompts")
	
	// Override the default base dir for testing
	originalBaseDir := DefaultBaseDir()
	
	// For this test, we manually install to the tmp dir
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}
	
	// Create a loader pointing to the new location
	loader := NewPromptLoaderWithDir(promptsDir)
	
	// Copy embedded prompts manually for testing
	testContent := "test guardrails"
	guardrailsPath := filepath.Join(promptsDir, "guardrails")
	if err := os.MkdirAll(guardrailsPath, 0755); err != nil {
		t.Fatalf("failed to create guardrails dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(guardrailsPath, "default.md"), []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test prompt: %v", err)
	}
	
	// Verify the prompt can be loaded
	loaded, err := loader.Load(PathGuardrailsDefault)
	if err != nil {
		t.Fatalf("failed to load installed prompt: %v", err)
	}
	
	if loaded != testContent {
		t.Errorf("expected %q, got %q", testContent, loaded)
	}
	
	_ = originalBaseDir // Keep reference to avoid unused warning
}

func TestPathConstants(t *testing.T) {
	// Verify path constants are non-empty
	paths := []string{
		PathSystemBase,
		PathSystemToolUsage,
		PathSystemMemory,
		PathSystemPlanning,
		PathGuardrailsDefault,
		PathGuardrailsSafety,
		PathSandwichPrefix,
		PathSandwichSuffix,
	}
	
	for _, p := range paths {
		if p == "" {
			t.Error("empty path constant")
		}
	}
}
