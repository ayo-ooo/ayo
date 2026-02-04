package providers

import (
	"context"
	"testing"
	"time"
)

// MockMemoryProvider is a test implementation of MemoryProvider.
type MockMemoryProvider struct {
	memories map[string]Memory
	name     string
}

func NewMockMemoryProvider(name string) *MockMemoryProvider {
	return &MockMemoryProvider{
		memories: make(map[string]Memory),
		name:     name,
	}
}

func (m *MockMemoryProvider) Name() string            { return m.name }
func (m *MockMemoryProvider) Type() ProviderType     { return ProviderTypeMemory }
func (m *MockMemoryProvider) Init(ctx context.Context, config map[string]any) error { return nil }
func (m *MockMemoryProvider) Close() error           { return nil }

func (m *MockMemoryProvider) Create(ctx context.Context, mem Memory) (Memory, error) {
	if mem.ID == "" {
		mem.ID = "mock-" + time.Now().Format("20060102150405")
	}
	mem.CreatedAt = time.Now()
	mem.UpdatedAt = mem.CreatedAt
	if mem.Status == "" {
		mem.Status = MemoryStatusActive
	}
	m.memories[mem.ID] = mem
	return mem, nil
}

func (m *MockMemoryProvider) Get(ctx context.Context, id string) (Memory, error) {
	if mem, ok := m.memories[id]; ok {
		return mem, nil
	}
	return Memory{}, nil
}

func (m *MockMemoryProvider) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	var results []SearchResult
	for _, mem := range m.memories {
		if mem.Status != MemoryStatusActive {
			continue
		}
		results = append(results, SearchResult{Memory: mem, Similarity: 0.5, MatchType: "mock"})
	}
	return results, nil
}

func (m *MockMemoryProvider) List(ctx context.Context, opts ListOptions) ([]Memory, error) {
	var results []Memory
	for _, mem := range m.memories {
		results = append(results, mem)
	}
	return results, nil
}

func (m *MockMemoryProvider) Update(ctx context.Context, mem Memory) error {
	mem.UpdatedAt = time.Now()
	m.memories[mem.ID] = mem
	return nil
}

func (m *MockMemoryProvider) Forget(ctx context.Context, id string) error {
	if mem, ok := m.memories[id]; ok {
		mem.Status = MemoryStatusForgotten
		m.memories[id] = mem
	}
	return nil
}

func (m *MockMemoryProvider) Supersede(ctx context.Context, oldID string, newMem Memory, reason string) (Memory, error) {
	if old, ok := m.memories[oldID]; ok {
		old.Status = MemoryStatusSuperseded
		m.memories[oldID] = old
	}
	newMem.SupersedesID = oldID
	return m.Create(ctx, newMem)
}

func (m *MockMemoryProvider) Topics(ctx context.Context) ([]string, error) {
	topicSet := make(map[string]bool)
	for _, mem := range m.memories {
		for _, t := range mem.Topics {
			topicSet[t] = true
		}
	}
	var topics []string
	for t := range topicSet {
		topics = append(topics, t)
	}
	return topics, nil
}

func (m *MockMemoryProvider) Link(ctx context.Context, id1, id2 string) error   { return nil }
func (m *MockMemoryProvider) Unlink(ctx context.Context, id1, id2 string) error { return nil }
func (m *MockMemoryProvider) Reindex(ctx context.Context) error                 { return nil }

// MockSandboxProvider is a test implementation of SandboxProvider.
type MockSandboxProvider struct {
	sandboxes map[string]Sandbox
	name      string
}

func NewMockSandboxProvider(name string) *MockSandboxProvider {
	return &MockSandboxProvider{
		sandboxes: make(map[string]Sandbox),
		name:      name,
	}
}

func (m *MockSandboxProvider) Name() string            { return m.name }
func (m *MockSandboxProvider) Type() ProviderType     { return ProviderTypeSandbox }
func (m *MockSandboxProvider) Init(ctx context.Context, config map[string]any) error { return nil }
func (m *MockSandboxProvider) Close() error           { return nil }

func (m *MockSandboxProvider) Create(ctx context.Context, opts SandboxCreateOptions) (Sandbox, error) {
	id := "sandbox-" + time.Now().Format("20060102150405")
	sb := Sandbox{
		ID:        id,
		Name:      opts.Name,
		Image:     opts.Image,
		Status:    SandboxStatusRunning,
		Pool:      opts.Pool,
		CreatedAt: time.Now(),
		Mounts:    opts.Mounts,
		Resources: opts.Resources,
	}
	m.sandboxes[id] = sb
	return sb, nil
}

func (m *MockSandboxProvider) Get(ctx context.Context, id string) (Sandbox, error) {
	if sb, ok := m.sandboxes[id]; ok {
		return sb, nil
	}
	return Sandbox{}, nil
}

func (m *MockSandboxProvider) List(ctx context.Context) ([]Sandbox, error) {
	var results []Sandbox
	for _, sb := range m.sandboxes {
		results = append(results, sb)
	}
	return results, nil
}

func (m *MockSandboxProvider) Start(ctx context.Context, id string) error {
	if sb, ok := m.sandboxes[id]; ok {
		sb.Status = SandboxStatusRunning
		m.sandboxes[id] = sb
	}
	return nil
}

func (m *MockSandboxProvider) Stop(ctx context.Context, id string, opts SandboxStopOptions) error {
	if sb, ok := m.sandboxes[id]; ok {
		sb.Status = SandboxStatusStopped
		m.sandboxes[id] = sb
	}
	return nil
}

func (m *MockSandboxProvider) Delete(ctx context.Context, id string, force bool) error {
	delete(m.sandboxes, id)
	return nil
}

func (m *MockSandboxProvider) Exec(ctx context.Context, id string, opts ExecOptions) (ExecResult, error) {
	return ExecResult{
		Stdout:   "mock output",
		ExitCode: 0,
		Duration: 10 * time.Millisecond,
	}, nil
}

func (m *MockSandboxProvider) Status(ctx context.Context, id string) (SandboxStatus, error) {
	if sb, ok := m.sandboxes[id]; ok {
		return sb.Status, nil
	}
	return SandboxStatusFailed, nil
}

// MockEmbeddingProvider is a test implementation of EmbeddingProvider.
type MockEmbeddingProvider struct {
	name       string
	dimensions int
}

func NewMockEmbeddingProvider(name string, dims int) *MockEmbeddingProvider {
	return &MockEmbeddingProvider{name: name, dimensions: dims}
}

func (m *MockEmbeddingProvider) Name() string            { return m.name }
func (m *MockEmbeddingProvider) Type() ProviderType     { return ProviderTypeEmbedding }
func (m *MockEmbeddingProvider) Init(ctx context.Context, config map[string]any) error { return nil }
func (m *MockEmbeddingProvider) Close() error           { return nil }

func (m *MockEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	vec := make([]float32, m.dimensions)
	for i := range vec {
		vec[i] = float32(i) / float32(m.dimensions)
	}
	return vec, nil
}

func (m *MockEmbeddingProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i := range texts {
		vec, _ := m.Embed(ctx, texts[i])
		results[i] = vec
	}
	return results, nil
}

func (m *MockEmbeddingProvider) Dimensions() int { return m.dimensions }
func (m *MockEmbeddingProvider) Model() string   { return "mock-embed" }

// Tests

func TestProviderTypes(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		wantType ProviderType
	}{
		{"memory", NewMockMemoryProvider("test-memory"), ProviderTypeMemory},
		{"sandbox", NewMockSandboxProvider("test-sandbox"), ProviderTypeSandbox},
		{"embedding", NewMockEmbeddingProvider("test-embed", 384), ProviderTypeEmbedding},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.provider.Type(); got != tt.wantType {
				t.Errorf("Type() = %v, want %v", got, tt.wantType)
			}
		})
	}
}

func TestMemoryProviderInterface(t *testing.T) {
	ctx := context.Background()
	mp := NewMockMemoryProvider("test")

	// Test Create
	mem, err := mp.Create(ctx, Memory{
		Content:  "User prefers Go",
		Category: MemoryCategoryPreference,
		Topics:   []string{"go", "languages"},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if mem.ID == "" {
		t.Error("Create() should generate ID")
	}
	if mem.Status != MemoryStatusActive {
		t.Errorf("Create() status = %v, want active", mem.Status)
	}

	// Test Get
	got, err := mp.Get(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Content != mem.Content {
		t.Errorf("Get() content = %v, want %v", got.Content, mem.Content)
	}

	// Test Search
	results, err := mp.Search(ctx, "Go", SearchOptions{})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Search() got %d results, want 1", len(results))
	}

	// Test Forget
	if err := mp.Forget(ctx, mem.ID); err != nil {
		t.Fatalf("Forget() error = %v", err)
	}
	got, _ = mp.Get(ctx, mem.ID)
	if got.Status != MemoryStatusForgotten {
		t.Errorf("Forget() status = %v, want forgotten", got.Status)
	}
}

func TestSandboxProviderInterface(t *testing.T) {
	ctx := context.Background()
	sp := NewMockSandboxProvider("test")

	// Test Create
	sb, err := sp.Create(ctx, SandboxCreateOptions{
		Name:  "test-sandbox",
		Image: "busybox",
		Pool:  "default",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if sb.ID == "" {
		t.Error("Create() should generate ID")
	}
	if sb.Status != SandboxStatusRunning {
		t.Errorf("Create() status = %v, want running", sb.Status)
	}

	// Test Exec
	result, err := sp.Exec(ctx, sb.ID, ExecOptions{
		Command: "echo",
		Args:    []string{"hello"},
	})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Exec() exit code = %d, want 0", result.ExitCode)
	}

	// Test Stop
	if err := sp.Stop(ctx, sb.ID, SandboxStopOptions{}); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	status, _ := sp.Status(ctx, sb.ID)
	if status != SandboxStatusStopped {
		t.Errorf("Stop() status = %v, want stopped", status)
	}

	// Test Delete
	if err := sp.Delete(ctx, sb.ID, false); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	list, _ := sp.List(ctx)
	if len(list) != 0 {
		t.Errorf("Delete() list len = %d, want 0", len(list))
	}
}

func TestEmbeddingProviderInterface(t *testing.T) {
	ctx := context.Background()
	ep := NewMockEmbeddingProvider("test", 384)

	// Test Embed
	vec, err := ep.Embed(ctx, "test text")
	if err != nil {
		t.Fatalf("Embed() error = %v", err)
	}
	if len(vec) != 384 {
		t.Errorf("Embed() dimensions = %d, want 384", len(vec))
	}

	// Test EmbedBatch
	vecs, err := ep.EmbedBatch(ctx, []string{"text1", "text2"})
	if err != nil {
		t.Fatalf("EmbedBatch() error = %v", err)
	}
	if len(vecs) != 2 {
		t.Errorf("EmbedBatch() count = %d, want 2", len(vecs))
	}

	// Test Dimensions
	if dims := ep.Dimensions(); dims != 384 {
		t.Errorf("Dimensions() = %d, want 384", dims)
	}
}

func TestMemoryCategories(t *testing.T) {
	categories := []MemoryCategory{
		MemoryCategoryPreference,
		MemoryCategoryFact,
		MemoryCategoryCorrection,
		MemoryCategoryPattern,
	}

	for _, c := range categories {
		if c == "" {
			t.Errorf("Category should not be empty")
		}
	}
}

func TestMemoryStatus(t *testing.T) {
	statuses := []MemoryStatus{
		MemoryStatusActive,
		MemoryStatusSuperseded,
		MemoryStatusArchived,
		MemoryStatusForgotten,
	}

	for _, s := range statuses {
		if s == "" {
			t.Errorf("Status should not be empty")
		}
	}
}

func TestSandboxStatus(t *testing.T) {
	statuses := []SandboxStatus{
		SandboxStatusCreating,
		SandboxStatusRunning,
		SandboxStatusStopped,
		SandboxStatusFailed,
	}

	for _, s := range statuses {
		if s == "" {
			t.Errorf("Status should not be empty")
		}
	}
}

func TestMountModes(t *testing.T) {
	modes := []MountMode{
		MountModeVirtioFS,
		MountModeBind,
		MountModeOverlay,
		MountModeTmpfs,
	}

	for _, m := range modes {
		if m == "" {
			t.Errorf("MountMode should not be empty")
		}
	}
}
