package capabilities

import (
	"context"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/squads"
)

func TestEntityIndex_GetAgent(t *testing.T) {
	idx := NewEntityIndex()
	idx.Agents = []IndexedAgent{
		{Handle: "@crush", Description: "Coding agent"},
		{Handle: "@writer", Description: "Writing agent"},
	}

	// Found
	agent := idx.GetAgent("@crush")
	if agent == nil {
		t.Fatal("expected to find @crush")
	}
	if agent.Description != "Coding agent" {
		t.Errorf("Description = %q, want %q", agent.Description, "Coding agent")
	}

	// Not found
	agent = idx.GetAgent("@unknown")
	if agent != nil {
		t.Error("expected nil for unknown agent")
	}
}

func TestEntityIndex_GetSquad(t *testing.T) {
	idx := NewEntityIndex()
	idx.Squads = []IndexedSquad{
		{Name: "dev-team", Mission: "Build software"},
		{Name: "qa-team", Mission: "Test software"},
	}

	// Found
	squad := idx.GetSquad("dev-team")
	if squad == nil {
		t.Fatal("expected to find dev-team")
	}
	if squad.Mission != "Build software" {
		t.Errorf("Mission = %q, want %q", squad.Mission, "Build software")
	}

	// Not found
	squad = idx.GetSquad("unknown-team")
	if squad != nil {
		t.Error("expected nil for unknown squad")
	}
}

func TestEntityIndex_UpsertAgent(t *testing.T) {
	idx := NewEntityIndex()

	// Add new
	idx.UpsertAgent(IndexedAgent{Handle: "@crush", Description: "v1"})
	if idx.AgentCount() != 1 {
		t.Errorf("AgentCount = %d, want 1", idx.AgentCount())
	}

	// Update existing
	idx.UpsertAgent(IndexedAgent{Handle: "@crush", Description: "v2"})
	if idx.AgentCount() != 1 {
		t.Errorf("AgentCount = %d, want 1 (should update, not add)", idx.AgentCount())
	}
	if idx.GetAgent("@crush").Description != "v2" {
		t.Error("expected description to be updated")
	}

	// Add another
	idx.UpsertAgent(IndexedAgent{Handle: "@writer", Description: "writer"})
	if idx.AgentCount() != 2 {
		t.Errorf("AgentCount = %d, want 2", idx.AgentCount())
	}
}

func TestEntityIndex_UpsertSquad(t *testing.T) {
	idx := NewEntityIndex()

	// Add new
	idx.UpsertSquad(IndexedSquad{Name: "dev-team", Mission: "v1"})
	if idx.SquadCount() != 1 {
		t.Errorf("SquadCount = %d, want 1", idx.SquadCount())
	}

	// Update existing
	idx.UpsertSquad(IndexedSquad{Name: "dev-team", Mission: "v2"})
	if idx.SquadCount() != 1 {
		t.Errorf("SquadCount = %d, want 1 (should update, not add)", idx.SquadCount())
	}
	if idx.GetSquad("dev-team").Mission != "v2" {
		t.Error("expected mission to be updated")
	}
}

func TestEntityIndex_RemoveAgent(t *testing.T) {
	idx := NewEntityIndex()
	idx.Agents = []IndexedAgent{
		{Handle: "@crush"},
		{Handle: "@writer"},
	}

	idx.RemoveAgent("@crush")
	if idx.AgentCount() != 1 {
		t.Errorf("AgentCount = %d, want 1", idx.AgentCount())
	}
	if idx.GetAgent("@crush") != nil {
		t.Error("@crush should be removed")
	}
	if idx.GetAgent("@writer") == nil {
		t.Error("@writer should still exist")
	}

	// Remove non-existent (should not panic)
	idx.RemoveAgent("@unknown")
	if idx.AgentCount() != 1 {
		t.Errorf("AgentCount should still be 1")
	}
}

func TestEntityIndex_RemoveSquad(t *testing.T) {
	idx := NewEntityIndex()
	idx.Squads = []IndexedSquad{
		{Name: "dev-team"},
		{Name: "qa-team"},
	}

	idx.RemoveSquad("dev-team")
	if idx.SquadCount() != 1 {
		t.Errorf("SquadCount = %d, want 1", idx.SquadCount())
	}
	if idx.GetSquad("dev-team") != nil {
		t.Error("dev-team should be removed")
	}

	// Remove non-existent (should not panic)
	idx.RemoveSquad("unknown")
	if idx.SquadCount() != 1 {
		t.Errorf("SquadCount should still be 1")
	}
}

func TestEntityIndex_Counts(t *testing.T) {
	idx := NewEntityIndex()
	if idx.TotalCount() != 0 {
		t.Error("empty index should have 0 total")
	}

	idx.UpsertAgent(IndexedAgent{Handle: "@a"})
	idx.UpsertAgent(IndexedAgent{Handle: "@b"})
	idx.UpsertSquad(IndexedSquad{Name: "s1"})

	if idx.AgentCount() != 2 {
		t.Errorf("AgentCount = %d, want 2", idx.AgentCount())
	}
	if idx.SquadCount() != 1 {
		t.Errorf("SquadCount = %d, want 1", idx.SquadCount())
	}
	if idx.TotalCount() != 3 {
		t.Errorf("TotalCount = %d, want 3", idx.TotalCount())
	}
}

func TestIndexedAgent_NeedsUpdate(t *testing.T) {
	agent := &IndexedAgent{
		Handle:      "@crush",
		ContentHash: "abc123",
	}

	if agent.NeedsUpdate("abc123") {
		t.Error("same hash should not need update")
	}
	if !agent.NeedsUpdate("def456") {
		t.Error("different hash should need update")
	}
}

func TestIndexedSquad_NeedsUpdate(t *testing.T) {
	squad := &IndexedSquad{
		Name:        "dev-team",
		ContentHash: "abc123",
	}

	if squad.NeedsUpdate("abc123") {
		t.Error("same hash should not need update")
	}
	if !squad.NeedsUpdate("def456") {
		t.Error("different hash should need update")
	}
}

func TestIndexedAgent_HasEmbedding(t *testing.T) {
	agent := &IndexedAgent{}
	if agent.HasEmbedding() {
		t.Error("empty embedding should return false")
	}

	agent.Embedding = []float32{0.1, 0.2, 0.3}
	if !agent.HasEmbedding() {
		t.Error("non-empty embedding should return true")
	}
}

func TestIndexedSquad_HasEmbedding(t *testing.T) {
	squad := &IndexedSquad{}
	if squad.HasEmbedding() {
		t.Error("empty embedding should return false")
	}

	squad.Embedding = []float32{0.1, 0.2, 0.3}
	if !squad.HasEmbedding() {
		t.Error("non-empty embedding should return true")
	}
}

func TestNewEntityIndex(t *testing.T) {
	idx := NewEntityIndex()

	if idx == nil {
		t.Fatal("NewEntityIndex returned nil")
	}
	if idx.Agents == nil {
		t.Error("Agents should be initialized")
	}
	if idx.Squads == nil {
		t.Error("Squads should be initialized")
	}
	if len(idx.Agents) != 0 {
		t.Error("Agents should be empty")
	}
	if len(idx.Squads) != 0 {
		t.Error("Squads should be empty")
	}
}

func TestIndexedAgent_Fields(t *testing.T) {
	now := time.Now()
	agent := IndexedAgent{
		Handle:          "@crush",
		Description:     "A coding agent",
		ContentHash:     "abc123",
		Embedding:       []float32{0.1, 0.2},
		HasInputSchema:  true,
		HasOutputSchema: false,
		UpdatedAt:       now,
	}

	if agent.Handle != "@crush" {
		t.Error("Handle mismatch")
	}
	if agent.Description != "A coding agent" {
		t.Error("Description mismatch")
	}
	if agent.ContentHash != "abc123" {
		t.Error("ContentHash mismatch")
	}
	if len(agent.Embedding) != 2 {
		t.Error("Embedding length mismatch")
	}
	if !agent.HasInputSchema {
		t.Error("HasInputSchema should be true")
	}
	if agent.HasOutputSchema {
		t.Error("HasOutputSchema should be false")
	}
	if agent.UpdatedAt != now {
		t.Error("UpdatedAt mismatch")
	}
}

func TestIndexedSquad_Fields(t *testing.T) {
	now := time.Now()
	squad := IndexedSquad{
		Name:            "dev-team",
		Mission:         "Build great software",
		ContentHash:     "def456",
		Embedding:       []float32{0.3, 0.4},
		HasInputSchema:  false,
		HasOutputSchema: true,
		UpdatedAt:       now,
	}

	if squad.Name != "dev-team" {
		t.Error("Name mismatch")
	}
	if squad.Mission != "Build great software" {
		t.Error("Mission mismatch")
	}
	if squad.ContentHash != "def456" {
		t.Error("ContentHash mismatch")
	}
	if len(squad.Embedding) != 2 {
		t.Error("Embedding length mismatch")
	}
	if squad.HasInputSchema {
		t.Error("HasInputSchema should be false")
	}
	if !squad.HasOutputSchema {
		t.Error("HasOutputSchema should be true")
	}
}

// testEmbedder is a mock embedder for testing lazy invalidation.
type testEmbedder struct {
	embedCount int
	embedding  []float32
}

func (e *testEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	e.embedCount++
	return e.embedding, nil
}

func (e *testEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i], _ = e.Embed(ctx, texts[i])
	}
	return result, nil
}

func (e *testEmbedder) Dimension() int {
	return len(e.embedding)
}

func (e *testEmbedder) Close() error {
	return nil
}

func TestLazyEntityIndex_GetAgentEmbedding_CacheHit(t *testing.T) {
	embedder := &testEmbedder{embedding: []float32{1.0, 2.0, 3.0}}
	idx := NewLazyEntityIndex(embedder)

	ag := agent.Agent{
		Handle: "@test",
		Config: agent.Config{Description: "Test agent"},
		System: "Test system",
	}

	// First call - should embed
	emb1, err := idx.GetAgentEmbedding(context.Background(), ag)
	if err != nil {
		t.Fatalf("GetAgentEmbedding failed: %v", err)
	}
	if embedder.embedCount != 1 {
		t.Errorf("Expected 1 embed call, got %d", embedder.embedCount)
	}
	if len(emb1) != 3 {
		t.Errorf("Expected 3-dim embedding, got %d", len(emb1))
	}

	// Second call with same content - should hit cache
	emb2, err := idx.GetAgentEmbedding(context.Background(), ag)
	if err != nil {
		t.Fatalf("GetAgentEmbedding failed: %v", err)
	}
	if embedder.embedCount != 1 {
		t.Errorf("Expected still 1 embed call (cache hit), got %d", embedder.embedCount)
	}
	if len(emb2) != 3 {
		t.Errorf("Expected 3-dim embedding, got %d", len(emb2))
	}
}

func TestLazyEntityIndex_GetAgentEmbedding_Invalidation(t *testing.T) {
	embedder := &testEmbedder{embedding: []float32{1.0, 2.0, 3.0}}
	idx := NewLazyEntityIndex(embedder)

	ag := agent.Agent{
		Handle: "@test",
		Config: agent.Config{Description: "Test agent"},
		System: "Test system",
	}

	// First call - should embed
	_, err := idx.GetAgentEmbedding(context.Background(), ag)
	if err != nil {
		t.Fatalf("GetAgentEmbedding failed: %v", err)
	}
	if embedder.embedCount != 1 {
		t.Errorf("Expected 1 embed call, got %d", embedder.embedCount)
	}

	// Modify agent content
	ag.System = "Modified system"

	// Second call with different content - should re-embed
	_, err = idx.GetAgentEmbedding(context.Background(), ag)
	if err != nil {
		t.Fatalf("GetAgentEmbedding failed: %v", err)
	}
	if embedder.embedCount != 2 {
		t.Errorf("Expected 2 embed calls (invalidation), got %d", embedder.embedCount)
	}
}

func TestLazyEntityIndex_GetSquadEmbedding_CacheHit(t *testing.T) {
	embedder := &testEmbedder{embedding: []float32{4.0, 5.0, 6.0}}
	idx := NewLazyEntityIndex(embedder)

	constitution := &squads.Constitution{
		Raw:       "# Test Squad\n\nTest mission",
		SquadName: "test-squad",
	}

	// First call - should embed
	emb1, err := idx.GetSquadEmbedding(context.Background(), constitution, "test-squad")
	if err != nil {
		t.Fatalf("GetSquadEmbedding failed: %v", err)
	}
	if embedder.embedCount != 1 {
		t.Errorf("Expected 1 embed call, got %d", embedder.embedCount)
	}
	if len(emb1) != 3 {
		t.Errorf("Expected 3-dim embedding, got %d", len(emb1))
	}

	// Second call with same content - should hit cache
	emb2, err := idx.GetSquadEmbedding(context.Background(), constitution, "test-squad")
	if err != nil {
		t.Fatalf("GetSquadEmbedding failed: %v", err)
	}
	if embedder.embedCount != 1 {
		t.Errorf("Expected still 1 embed call (cache hit), got %d", embedder.embedCount)
	}
	if len(emb2) != 3 {
		t.Errorf("Expected 3-dim embedding, got %d", len(emb2))
	}
}
