package ollama

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_IsAvailable(t *testing.T) {
	// Test with a server that returns 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"models":[]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(WithHost(server.URL))
	if !client.IsAvailable(context.Background()) {
		t.Error("expected IsAvailable to return true")
	}

	// Test with a server that returns 500
	server500 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server500.Close()

	client500 := NewClient(WithHost(server500.URL))
	if client500.IsAvailable(context.Background()) {
		t.Error("expected IsAvailable to return false for 500 response")
	}
}

func TestClient_ListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"models": [
					{"name": "llama3.2:latest", "size": 1000000},
					{"name": "nomic-embed-text:latest", "size": 500000}
				]
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(WithHost(server.URL))
	models, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}

	if models[0].Name != "llama3.2:latest" {
		t.Errorf("expected first model to be llama3.2:latest, got %s", models[0].Name)
	}
}

func TestClient_HasModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"models": [
					{"name": "llama3.2:latest", "size": 1000000},
					{"name": "nomic-embed-text:latest", "size": 500000}
				]
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(WithHost(server.URL))

	// Test with exact match
	if !client.HasModel(context.Background(), "llama3.2:latest") {
		t.Error("expected HasModel to return true for llama3.2:latest")
	}

	// Test with normalized name (without :latest)
	if !client.HasModel(context.Background(), "llama3.2") {
		t.Error("expected HasModel to return true for llama3.2")
	}

	// Test with non-existent model
	if client.HasModel(context.Background(), "nonexistent") {
		t.Error("expected HasModel to return false for nonexistent model")
	}
}

func TestClient_GetVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/version" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"version": "0.5.1"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(WithHost(server.URL))
	version, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if version != "0.5.1" {
		t.Errorf("expected version 0.5.1, got %s", version)
	}
}

func TestClient_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"model": "llama3.2",
				"message": {"role": "assistant", "content": "Hello! How can I help you?"},
				"done": true
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(WithHost(server.URL))
	resp, err := client.Chat(context.Background(), "llama3.2", []Message{
		{Role: "user", Content: "Hello"},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Message.Content != "Hello! How can I help you?" {
		t.Errorf("unexpected response content: %s", resp.Message.Content)
	}
}

func TestClient_Embed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/embed" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"model": "nomic-embed-text",
				"embeddings": [[0.1, 0.2, 0.3, 0.4]]
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(WithHost(server.URL))
	embedding, err := client.Embed(context.Background(), "nomic-embed-text", "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(embedding) != 4 {
		t.Fatalf("expected 4 dimensions, got %d", len(embedding))
	}

	if embedding[0] != 0.1 {
		t.Errorf("expected first value to be 0.1, got %f", embedding[0])
	}
}

func TestNormalizeModelName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"llama3.2:latest", "llama3.2"},
		{"llama3.2", "llama3.2"},
		{"nomic-embed-text:latest", "nomic-embed-text"},
		{"model:v1", "model:v1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeModelName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeModelName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
