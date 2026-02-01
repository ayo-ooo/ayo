package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexcabrera/ayo/internal/config"
)

func TestServerHealth(t *testing.T) {
	srv := New(config.Config{}, Options{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", resp["status"])
	}
}

func TestServerConnect(t *testing.T) {
	srv := New(config.Config{}, Options{})

	req := httptest.NewRequest(http.MethodGet, "/connect", nil)
	w := httptest.NewRecorder()

	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["version"] != float64(1) {
		t.Errorf("expected version 1, got %v", resp["version"])
	}

	token, ok := resp["token"].(string)
	if !ok || token == "" {
		t.Error("expected non-empty token")
	}

	if token != srv.Token() {
		t.Errorf("token mismatch: response=%q, server=%q", token, srv.Token())
	}
}

func TestServerAuthLocalhostBypass(t *testing.T) {
	srv := New(config.Config{}, Options{
		AllowRemote: false, // localhost only
	})

	// Request from localhost should succeed without token
	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for localhost, got %d", w.Code)
	}
}

func TestServerAuthTokenRequired(t *testing.T) {
	srv := New(config.Config{}, Options{
		AllowRemote: true, // require auth even for localhost
	})

	// Request without token should fail
	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	// Request with valid token should succeed
	req = httptest.NewRequest(http.MethodGet, "/agents", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Authorization", "Bearer "+srv.Token())
	w = httptest.NewRecorder()

	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 with token, got %d", w.Code)
	}
}

func TestServerAuthInvalidToken(t *testing.T) {
	srv := New(config.Config{}, Options{
		AllowRemote: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	req.RemoteAddr = "192.168.1.100:12345" // non-localhost
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()

	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for invalid token, got %d", w.Code)
	}
}

func TestServerCORS(t *testing.T) {
	srv := New(config.Config{}, Options{})

	// Test CORS preflight on various endpoints
	endpoints := []string{"/", "/agents", "/agents/ayo", "/agents/ayo/chat", "/sessions"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, endpoint, nil)
			req.RemoteAddr = "127.0.0.1:12345"
			w := httptest.NewRecorder()

			srv.mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200 for OPTIONS %s, got %d", endpoint, w.Code)
			}

			if w.Header().Get("Access-Control-Allow-Origin") != "*" {
				t.Errorf("expected CORS Allow-Origin header for %s", endpoint)
			}
		})
	}
}

func TestServerStartStop(t *testing.T) {
	srv := New(config.Config{}, Options{
		Addr: "127.0.0.1:0",
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Start server in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	// Wait for server to start
	time.Sleep(50 * time.Millisecond)

	addr := srv.Addr()
	if addr == "" {
		t.Fatal("expected non-empty address after start")
	}

	// Make a request
	resp, err := http.Get("http://" + addr + "/health")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Stop server
	cancel()

	// Wait for server to stop
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("server did not stop in time")
	}
}

func TestServerRoutes(t *testing.T) {
	srv := New(config.Config{}, Options{})

	tests := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/health", http.StatusOK},
		{http.MethodGet, "/connect", http.StatusOK},
		{http.MethodGet, "/agents", http.StatusOK},
		{http.MethodGet, "/agents/ayo", http.StatusOK},                           // @ayo is a built-in agent
		{http.MethodGet, "/sessions", http.StatusServiceUnavailable},             // No session service configured
		{http.MethodPost, "/agents/ayo/chat", http.StatusBadRequest},             // Missing request body
		{http.MethodPost, "/agents/ayo/sessions/123", http.StatusBadRequest},     // Missing request body
		{http.MethodGet, "/agents/ayo/sessions/123", http.StatusServiceUnavailable}, // No session service
		{http.MethodDelete, "/agents/ayo/sessions/123", http.StatusServiceUnavailable}, // No session service
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.RemoteAddr = "127.0.0.1:12345"
			w := httptest.NewRecorder()

			srv.mux.ServeHTTP(w, req)

			if w.Code != tt.want {
				t.Errorf("expected status %d, got %d", tt.want, w.Code)
			}
		})
	}
}
