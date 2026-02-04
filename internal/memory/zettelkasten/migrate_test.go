package zettelkasten

import (
	"database/sql"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/db"
)

func TestDbMemoryToProviderMemory(t *testing.T) {
	now := time.Now().Unix()

	dbMem := db.Memory{
		ID:                 "test-id",
		Content:            "Test content",
		Category:           "preference",
		AgentHandle:        sql.NullString{String: "@ayo", Valid: true},
		PathScope:          sql.NullString{String: "/project", Valid: true},
		SourceSessionID:    sql.NullString{String: "sess-123", Valid: true},
		SourceMessageID:    sql.NullString{String: "msg-456", Valid: true},
		CreatedAt:          now,
		UpdatedAt:          now,
		Confidence:         sql.NullFloat64{Float64: 0.95, Valid: true},
		LastAccessedAt:     sql.NullInt64{Int64: now, Valid: true},
		AccessCount:        sql.NullInt64{Int64: 5, Valid: true},
		SupersedesID:       sql.NullString{String: "old-id", Valid: true},
		SupersededByID:     sql.NullString{},
		SupersessionReason: sql.NullString{String: "updated info", Valid: true},
		Status:             sql.NullString{String: "active", Valid: true},
	}

	mem := dbMemoryToProviderMemory(dbMem)

	if mem.ID != "test-id" {
		t.Errorf("unexpected ID: %s", mem.ID)
	}
	if mem.Content != "Test content" {
		t.Errorf("unexpected content: %s", mem.Content)
	}
	if string(mem.Category) != "preference" {
		t.Errorf("unexpected category: %s", mem.Category)
	}
	if mem.AgentHandle != "@ayo" {
		t.Errorf("unexpected agent handle: %s", mem.AgentHandle)
	}
	if mem.PathScope != "/project" {
		t.Errorf("unexpected path scope: %s", mem.PathScope)
	}
	if mem.Confidence != 0.95 {
		t.Errorf("unexpected confidence: %f", mem.Confidence)
	}
	if mem.AccessCount != 5 {
		t.Errorf("unexpected access count: %d", mem.AccessCount)
	}
	if mem.SupersedesID != "old-id" {
		t.Errorf("unexpected supersedes ID: %s", mem.SupersedesID)
	}
	if string(mem.Status) != "active" {
		t.Errorf("unexpected status: %s", mem.Status)
	}
}

func TestDbMemoryToProviderMemoryNulls(t *testing.T) {
	now := time.Now().Unix()

	// Test with all nullable fields as null
	dbMem := db.Memory{
		ID:        "test-id-null",
		Content:   "Test content",
		Category:  "fact",
		CreatedAt: now,
		UpdatedAt: now,
		// All optional fields are zero/null
	}

	mem := dbMemoryToProviderMemory(dbMem)

	if mem.AgentHandle != "" {
		t.Errorf("expected empty agent handle, got: %s", mem.AgentHandle)
	}
	if mem.PathScope != "" {
		t.Errorf("expected empty path scope, got: %s", mem.PathScope)
	}
	if mem.Confidence != 1.0 {
		t.Errorf("expected default confidence 1.0, got: %f", mem.Confidence)
	}
	if mem.AccessCount != 0 {
		t.Errorf("expected 0 access count, got: %d", mem.AccessCount)
	}
	if string(mem.Status) != "active" {
		t.Errorf("expected default status 'active', got: %s", mem.Status)
	}
}

func TestNullStringValue(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullString
		expected string
	}{
		{"valid string", sql.NullString{String: "hello", Valid: true}, "hello"},
		{"null string", sql.NullString{Valid: false}, ""},
		{"empty valid", sql.NullString{String: "", Valid: true}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nullStringValue(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNullFloat64Value(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullFloat64
		def      float64
		expected float64
	}{
		{"valid float", sql.NullFloat64{Float64: 0.75, Valid: true}, 1.0, 0.75},
		{"null uses default", sql.NullFloat64{Valid: false}, 1.0, 1.0},
		{"zero valid", sql.NullFloat64{Float64: 0.0, Valid: true}, 1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nullFloat64Value(tt.input, tt.def)
			if result != tt.expected {
				t.Errorf("got %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestNullInt64Value(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullInt64
		expected int64
	}{
		{"valid int", sql.NullInt64{Int64: 42, Valid: true}, 42},
		{"null uses zero", sql.NullInt64{Valid: false}, 0},
		{"zero valid", sql.NullInt64{Int64: 0, Valid: true}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nullInt64Value(tt.input)
			if result != tt.expected {
				t.Errorf("got %d, want %d", result, tt.expected)
			}
		})
	}
}
