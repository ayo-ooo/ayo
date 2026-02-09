package agent

import (
	"testing"
)

func TestHashPrompt(t *testing.T) {
	// Test deterministic hashing
	hash1 := hashPrompt("You are a helpful assistant.")
	hash2 := hashPrompt("You are a helpful assistant.")
	hash3 := hashPrompt("Different prompt")

	if hash1 != hash2 {
		t.Error("same prompt should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("different prompts should produce different hashes")
	}
	if len(hash1) != 64 {
		t.Errorf("expected 64-character hex hash, got %d", len(hash1))
	}
}

func TestCreateOptions(t *testing.T) {
	opts := CreateOptions{
		Handle:         "test-agent",
		SystemPrompt:   "You are a test assistant.",
		Description:    "Test agent",
		Skills:         []string{"coding"},
		AllowedTools:   []string{"bash"},
		CreatedBy:      "@ayo",
		CreationReason: "Testing",
	}

	if opts.Handle != "test-agent" {
		t.Errorf("expected handle 'test-agent', got %q", opts.Handle)
	}
	if opts.CreatedBy != "@ayo" {
		t.Errorf("expected created_by '@ayo', got %q", opts.CreatedBy)
	}
}

func TestSqlNullString(t *testing.T) {
	// Test empty string
	ns := sqlNullString("")
	if ns.Valid {
		t.Error("empty string should produce invalid NullString")
	}

	// Test non-empty string
	ns = sqlNullString("test")
	if !ns.Valid {
		t.Error("non-empty string should produce valid NullString")
	}
	if ns.String != "test" {
		t.Errorf("expected 'test', got %q", ns.String)
	}
}

func TestSqlNullInt64(t *testing.T) {
	ni := sqlNullInt64(42)
	if !ni.Valid {
		t.Error("should produce valid NullInt64")
	}
	if ni.Int64 != 42 {
		t.Errorf("expected 42, got %d", ni.Int64)
	}
}
