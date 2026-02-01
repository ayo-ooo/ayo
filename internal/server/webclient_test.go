package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexcabrera/ayo/internal/config"
)

func TestWebClientServing(t *testing.T) {
	srv := New(config.Config{}, Options{})

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected Content-Type text/html, got %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("response should contain HTML doctype")
	}
	if !strings.Contains(body, "ayo") {
		t.Error("response should contain ayo title")
	}
}

func TestWebClientCacheControl(t *testing.T) {
	srv := New(config.Config{}, Options{})

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	srv.mux.ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("expected Cache-Control no-cache, got %s", cacheControl)
	}
}
