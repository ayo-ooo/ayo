package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestWebhookServer_StartStop(t *testing.T) {
	var fired bool
	server := NewWebhookServer(WebhookServerConfig{
		Callback: func(event TriggerEvent) {
			fired = true
		},
	})

	ctx := context.Background()

	// Start
	if err := server.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Check address
	addr := server.Addr()
	if addr == "" {
		t.Error("Expected non-empty address")
	}

	port := server.Port()
	if port == 0 {
		t.Error("Expected non-zero port")
	}

	// Stop
	if err := server.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	if fired {
		t.Error("Callback should not have been called")
	}
}

func TestWebhookServer_RegisterTrigger(t *testing.T) {
	var firedAgent string
	server := NewWebhookServer(WebhookServerConfig{
		Callback: func(event TriggerEvent) {
			firedAgent = event.Agent
		},
	})

	ctx := context.Background()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer server.Stop(ctx)

	// Register trigger
	trigger := &WebhookTrigger{
		ID:    "test-webhook",
		Path:  "/hooks/test",
		Agent: "@webhook-agent",
	}
	if err := server.Register(trigger); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Send webhook request
	url := "http://" + server.Addr() + "/hooks/test"
	payload := map[string]any{"test": true}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Wait for callback
	time.Sleep(100 * time.Millisecond)

	if firedAgent != "@webhook-agent" {
		t.Errorf("Agent: got %q, want %q", firedAgent, "@webhook-agent")
	}
}

func TestWebhookServer_Health(t *testing.T) {
	server := NewWebhookServer(WebhookServerConfig{})

	ctx := context.Background()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer server.Stop(ctx)

	// Call health endpoint
	url := "http://" + server.Addr() + "/health"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if result["status"] != "healthy" {
		t.Errorf("Status: got %v, want healthy", result["status"])
	}
}

func TestWebhookServer_GenericEndpoint(t *testing.T) {
	var firedSource string
	server := NewWebhookServer(WebhookServerConfig{
		Callback: func(event TriggerEvent) {
			if ctx := event.Context; ctx != nil {
				if source, ok := ctx["source"].(string); ok {
					firedSource = source
				}
			}
		},
	})

	ctx := context.Background()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer server.Stop(ctx)

	// Register generic trigger
	trigger := &WebhookTrigger{
		ID:     "generic-trigger",
		Path:   "/hooks/generic",
		Agent:  "@generic-agent",
		Format: "generic",
	}
	if err := server.Register(trigger); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Send to generic endpoint
	url := "http://" + server.Addr() + "/hooks/generic"
	payload := map[string]any{"action": "test"}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Wait for callback
	time.Sleep(100 * time.Millisecond)

	if firedSource != "generic" {
		t.Errorf("Source: got %q, want %q", firedSource, "generic")
	}
}

func TestWebhookServer_NotFound(t *testing.T) {
	server := NewWebhookServer(WebhookServerConfig{})

	ctx := context.Background()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer server.Stop(ctx)

	// Try non-existent hook
	url := "http://" + server.Addr() + "/hooks/nonexistent"
	resp, err := http.Post(url, "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Status: got %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWebhookServer_Unregister(t *testing.T) {
	server := NewWebhookServer(WebhookServerConfig{})

	ctx := context.Background()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer server.Stop(ctx)

	// Register and unregister
	trigger := &WebhookTrigger{
		ID:    "temp-trigger",
		Path:  "/hooks/temp",
		Agent: "@temp-agent",
	}
	if err := server.Register(trigger); err != nil {
		t.Fatalf("Register: %v", err)
	}

	server.Unregister("/hooks/temp")

	// Should now return 404
	url := "http://" + server.Addr() + "/hooks/temp"
	resp, err := http.Post(url, "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Status: got %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}
