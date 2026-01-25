package smallmodel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestService_ExtractMemory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.WriteHeader(http.StatusOK)
			resp := map[string]any{
				"model": "granite4:3b",
				"message": map[string]string{
					"role":    "assistant",
					"content": `{"should_remember": true, "content": "User prefers TypeScript", "category": "preference", "confidence": 0.9, "reason": "explicit preference statement"}`,
				},
				"done": true,
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	svc := NewService(Config{Host: server.URL})
	result, err := svc.ExtractMemory(context.Background(), "remember that I prefer TypeScript")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.ShouldRemember {
		t.Error("expected ShouldRemember to be true")
	}
	if result.Content != "User prefers TypeScript" {
		t.Errorf("unexpected content: %s", result.Content)
	}
	if result.Category != "preference" {
		t.Errorf("unexpected category: %s", result.Category)
	}
}

func TestService_ExtractMemory_NoMemory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.WriteHeader(http.StatusOK)
			resp := map[string]any{
				"model": "granite4:3b",
				"message": map[string]string{
					"role":    "assistant",
					"content": `{"should_remember": false, "reason": "just a greeting"}`,
				},
				"done": true,
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	svc := NewService(Config{Host: server.URL})
	result, err := svc.ExtractMemory(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ShouldRemember {
		t.Error("expected ShouldRemember to be false")
	}
}

func TestService_CheckDuplicate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.WriteHeader(http.StatusOK)
			resp := map[string]any{
				"model": "granite4:3b",
				"message": map[string]string{
					"role":    "assistant",
					"content": `{"action": "duplicate", "reason": "same preference already stored"}`,
				},
				"done": true,
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	svc := NewService(Config{Host: server.URL})
	result, err := svc.CheckDuplicate(context.Background(), "prefers TypeScript", []ExistingMemory{
		{ID: "abc12345-1234-1234-1234-123456789012", Content: "User prefers TypeScript over JavaScript"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Action != "duplicate" {
		t.Errorf("expected action 'duplicate', got %s", result.Action)
	}
}

func TestService_CheckDuplicate_NoExisting(t *testing.T) {
	svc := NewService(Config{})
	result, err := svc.CheckDuplicate(context.Background(), "prefers TypeScript", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Action != "new" {
		t.Errorf("expected action 'new', got %s", result.Action)
	}
}

func TestService_GenerateTitle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.WriteHeader(http.StatusOK)
			resp := map[string]any{
				"model": "granite4:3b",
				"message": map[string]string{
					"role":    "assistant",
					"content": "React TypeScript Setup",
				},
				"done": true,
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	svc := NewService(Config{Host: server.URL})
	title, err := svc.GenerateTitle(context.Background(), "Help me set up a React project with TypeScript")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if title != "React TypeScript Setup" {
		t.Errorf("unexpected title: %s", title)
	}
}
