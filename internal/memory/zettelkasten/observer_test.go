package zettelkasten

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
)

// MockSmallModel implements SmallModelProvider for testing.
type MockSmallModel struct {
	responses map[string]string
	calls     []string
	mu        sync.Mutex
}

func NewMockSmallModel() *MockSmallModel {
	return &MockSmallModel{
		responses: make(map[string]string),
	}
}

func (m *MockSmallModel) Complete(ctx context.Context, prompt string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, prompt)

	// Check for specific responses
	for key, response := range m.responses {
		if containsSubstring(prompt, key) {
			return response, nil
		}
	}
	return "NONE", nil
}

func (m *MockSmallModel) SetResponse(promptContains, response string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[promptContains] = response
}

func (m *MockSmallModel) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

// MockMemoryProvider implements MemoryProvider for testing.
type MockMemoryProvider struct {
	memories []providers.Memory
	mu       sync.Mutex
}

func NewMockMemoryProvider() *MockMemoryProvider {
	return &MockMemoryProvider{}
}

func (m *MockMemoryProvider) Name() string                         { return "mock" }
func (m *MockMemoryProvider) Type() providers.ProviderType         { return providers.ProviderTypeMemory }
func (m *MockMemoryProvider) Init(context.Context, map[string]any) error { return nil }
func (m *MockMemoryProvider) Close() error                         { return nil }

func (m *MockMemoryProvider) Create(ctx context.Context, mem providers.Memory) (providers.Memory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.memories = append(m.memories, mem)
	return mem, nil
}

func (m *MockMemoryProvider) Get(context.Context, string) (providers.Memory, error) {
	return providers.Memory{}, nil
}
func (m *MockMemoryProvider) Search(context.Context, string, providers.SearchOptions) ([]providers.SearchResult, error) {
	return nil, nil
}
func (m *MockMemoryProvider) List(context.Context, providers.ListOptions) ([]providers.Memory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.memories, nil
}
func (m *MockMemoryProvider) Update(context.Context, providers.Memory) error { return nil }
func (m *MockMemoryProvider) Forget(context.Context, string) error           { return nil }
func (m *MockMemoryProvider) Supersede(context.Context, string, providers.Memory, string) (providers.Memory, error) {
	return providers.Memory{}, nil
}
func (m *MockMemoryProvider) Topics(context.Context) ([]string, error)   { return nil, nil }
func (m *MockMemoryProvider) Link(context.Context, string, string) error   { return nil }
func (m *MockMemoryProvider) Unlink(context.Context, string, string) error { return nil }
func (m *MockMemoryProvider) Reindex(context.Context) error                { return nil }

func (m *MockMemoryProvider) MemoryCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.memories)
}

func TestObserver_Interface(t *testing.T) {
	var _ providers.ObserverProvider = (*Observer)(nil)
}

func TestObserver_StartStop(t *testing.T) {
	obs := NewObserver(ObserverConfig{})
	ctx := context.Background()

	if err := obs.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Double start should be idempotent
	if err := obs.Start(ctx); err != nil {
		t.Fatalf("Start again: %v", err)
	}

	if err := obs.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	// Double stop should be idempotent
	if err := obs.Stop(); err != nil {
		t.Fatalf("Stop again: %v", err)
	}
}

func TestObserver_OnMessage(t *testing.T) {
	memProvider := NewMockMemoryProvider()
	smallModel := NewMockSmallModel()
	smallModel.SetResponse("dark mode", "preference|User prefers dark mode")

	obs := NewObserver(ObserverConfig{
		MemoryProvider: memProvider,
		SmallModel:     smallModel,
		BatchSize:      1,
		BatchWait:      10 * time.Millisecond,
	})

	ctx := context.Background()
	obs.Start(ctx)
	defer obs.Stop()

	// Send a message
	obs.OnMessage(ctx, providers.MessageEvent{
		SessionID:   "sess-1",
		MessageID:   "msg-1",
		Role:        "user",
		Content:     "Please use dark mode for all my editors",
		AgentHandle: "@ayo",
		Timestamp:   time.Now(),
	})

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Should have extracted a memory
	if memProvider.MemoryCount() != 1 {
		t.Errorf("MemoryCount = %d, want 1", memProvider.MemoryCount())
	}
}

func TestObserver_SkipsShortMessages(t *testing.T) {
	memProvider := NewMockMemoryProvider()
	smallModel := NewMockSmallModel()

	obs := NewObserver(ObserverConfig{
		MemoryProvider: memProvider,
		SmallModel:     smallModel,
		BatchSize:      1,
		BatchWait:      10 * time.Millisecond,
	})

	ctx := context.Background()
	obs.Start(ctx)
	defer obs.Stop()

	// Send a very short message
	obs.OnMessage(ctx, providers.MessageEvent{
		SessionID: "sess-1",
		MessageID: "msg-1",
		Role:      "user",
		Content:   "hi",
	})

	time.Sleep(50 * time.Millisecond)

	// Should not have called the model
	if smallModel.CallCount() != 0 {
		t.Errorf("CallCount = %d, want 0 (short message skipped)", smallModel.CallCount())
	}
}

func TestObserver_SkipsSystemMessages(t *testing.T) {
	memProvider := NewMockMemoryProvider()
	smallModel := NewMockSmallModel()

	obs := NewObserver(ObserverConfig{
		MemoryProvider: memProvider,
		SmallModel:     smallModel,
		BatchSize:      1,
		BatchWait:      10 * time.Millisecond,
	})

	ctx := context.Background()
	obs.Start(ctx)
	defer obs.Stop()

	// Send a system message
	obs.OnMessage(ctx, providers.MessageEvent{
		SessionID: "sess-1",
		MessageID: "msg-1",
		Role:      "system",
		Content:   "This is a long system prompt with lots of instructions",
	})

	time.Sleep(50 * time.Millisecond)

	// Should not have called the model
	if smallModel.CallCount() != 0 {
		t.Errorf("CallCount = %d, want 0 (system message skipped)", smallModel.CallCount())
	}
}

func TestParseExtractionResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     int
		category string
	}{
		{
			name:     "none",
			response: "NONE",
			want:     0,
		},
		{
			name:     "empty",
			response: "",
			want:     0,
		},
		{
			name:     "single preference",
			response: "preference|User prefers dark mode",
			want:     1,
			category: "preference",
		},
		{
			name:     "single fact",
			response: "fact|Project uses Go 1.22",
			want:     1,
			category: "fact",
		},
		{
			name:     "multiple",
			response: "preference|Dark mode\nfact|Uses Go\ncorrection|Should use tabs",
			want:     3,
		},
		{
			name:     "invalid category defaults to fact",
			response: "invalid|Some content",
			want:     1,
			category: "fact",
		},
		{
			name:     "empty content skipped",
			response: "fact|",
			want:     0,
		},
		{
			name:     "no separator",
			response: "just some text without separator",
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memories := parseExtractionResponse(tt.response)
			if len(memories) != tt.want {
				t.Errorf("got %d memories, want %d", len(memories), tt.want)
			}
			if tt.category != "" && len(memories) > 0 {
				if string(memories[0].Category) != tt.category {
					t.Errorf("category = %q, want %q", memories[0].Category, tt.category)
				}
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("truncate", func(t *testing.T) {
		if truncate("hello", 10) != "hello" {
			t.Error("short string should not be truncated")
		}
		if truncate("hello world", 5) != "hello..." {
			t.Error("long string should be truncated")
		}
	})

	t.Run("splitLines", func(t *testing.T) {
		lines := splitLines("a\nb\nc")
		if len(lines) != 3 {
			t.Errorf("splitLines = %d, want 3", len(lines))
		}
	})

	t.Run("splitOnce", func(t *testing.T) {
		parts := splitOnce("a|b|c", "|")
		if len(parts) != 2 || parts[0] != "a" || parts[1] != "b|c" {
			t.Errorf("splitOnce = %v, want [a, b|c]", parts)
		}

		parts = splitOnce("no separator", "|")
		if len(parts) != 1 {
			t.Errorf("splitOnce no sep = %v, want [no separator]", parts)
		}
	})

	t.Run("trimSpace", func(t *testing.T) {
		if trimSpace("  hello  ") != "hello" {
			t.Error("trimSpace failed")
		}
		if trimSpace("\t\nhello\r\n") != "hello" {
			t.Error("trimSpace with various whitespace failed")
		}
	})
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
