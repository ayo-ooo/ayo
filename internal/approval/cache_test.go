package approval

import (
	"testing"
)

func TestCache_ApproveAll(t *testing.T) {
	cache := NewCache("session123")

	// Initially not approved
	if cache.IsApproved("/path/to/file.go") {
		t.Error("should not be approved initially")
	}

	// Approve all
	cache.ApproveAll()

	// Now should be approved
	if !cache.IsApproved("/path/to/file.go") {
		t.Error("should be approved after ApproveAll")
	}
	if !cache.IsApproved("/any/other/path.txt") {
		t.Error("any path should be approved after ApproveAll")
	}
}

func TestCache_PatternMatching(t *testing.T) {
	cache := NewCache("session123")

	// Add pattern for markdown files
	cache.AddPattern("*.md", "")

	// Should match
	if !cache.IsApproved("/docs/readme.md") {
		t.Error("should match *.md pattern")
	}

	// Should not match
	if cache.IsApproved("/docs/readme.txt") {
		t.Error("should not match *.md pattern")
	}
}

func TestCache_DirectoryScopedPattern(t *testing.T) {
	cache := NewCache("session123")

	// Add pattern scoped to directory
	cache.AddPattern("*.go", "/home/user/project/src")

	// Should match within directory
	if !cache.IsApproved("/home/user/project/src/main.go") {
		t.Error("should match within scoped directory")
	}

	// Should not match outside directory
	if cache.IsApproved("/home/user/other/main.go") {
		t.Error("should not match outside scoped directory")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache("session123")

	cache.ApproveAll()
	cache.AddPattern("*.md", "")

	if !cache.IsApproved("/any/file.md") {
		t.Error("should be approved")
	}

	cache.Clear()

	if cache.IsApproved("/any/file.md") {
		t.Error("should not be approved after Clear")
	}
}

func TestManager_GetCache(t *testing.T) {
	manager := NewManager()

	cache1 := manager.GetCache("session1")
	cache2 := manager.GetCache("session1")
	cache3 := manager.GetCache("session2")

	// Same session should return same cache
	if cache1 != cache2 {
		t.Error("same session should return same cache")
	}

	// Different session should return different cache
	if cache1 == cache3 {
		t.Error("different sessions should return different caches")
	}
}

func TestManager_ClearSession(t *testing.T) {
	manager := NewManager()

	cache := manager.GetCache("session1")
	cache.ApproveAll()

	manager.ClearSession("session1")

	// New cache should be fresh
	newCache := manager.GetCache("session1")
	if newCache.IsApproved("/any/file") {
		t.Error("new cache should not have old approvals")
	}
}

func TestGlobalManager(t *testing.T) {
	// Get global manager
	manager := GetManager()

	if manager == nil {
		t.Fatal("global manager should not be nil")
	}

	// Should return same instance
	manager2 := GetManager()
	if manager != manager2 {
		t.Error("should return same global manager instance")
	}
}

func TestIsApproved_Convenience(t *testing.T) {
	// Clear any existing state
	GetManager().ClearSession("test-session")

	// Initially not approved
	if IsApproved("test-session", "/path/to/file") {
		t.Error("should not be approved initially")
	}

	// Approve via cache
	cache := GetManager().GetCache("test-session")
	cache.ApproveAll()

	// Now should be approved
	if !IsApproved("test-session", "/path/to/file") {
		t.Error("should be approved after ApproveAll")
	}

	// Cleanup
	GetManager().ClearSession("test-session")
}
