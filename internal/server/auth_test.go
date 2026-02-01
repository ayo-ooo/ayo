package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthGenerateToken(t *testing.T) {
	auth := NewAuth(false)
	token := auth.Token()

	if token == "" {
		t.Error("expected non-empty token")
	}

	// Token should start with prefix
	if len(token) < 4 || token[:4] != "ayo_" {
		t.Errorf("token should start with 'ayo_', got %q", token)
	}

	// Each instance should have a unique token
	auth2 := NewAuth(false)
	if auth.Token() == auth2.Token() {
		t.Error("expected unique tokens for different instances")
	}
}

func TestAuthValidateLocalhost(t *testing.T) {
	auth := NewAuth(true) // localhost only

	tests := []struct {
		name       string
		remoteAddr string
		token      string
		want       bool
	}{
		{
			name:       "localhost IPv4 no token",
			remoteAddr: "127.0.0.1:12345",
			token:      "",
			want:       true,
		},
		{
			name:       "localhost IPv6 no token",
			remoteAddr: "[::1]:12345",
			token:      "",
			want:       true,
		},
		{
			name:       "remote no token",
			remoteAddr: "192.168.1.100:12345",
			token:      "",
			want:       false,
		},
		{
			name:       "remote with valid token",
			remoteAddr: "192.168.1.100:12345",
			token:      auth.Token(),
			want:       true,
		},
		{
			name:       "remote with invalid token",
			remoteAddr: "192.168.1.100:12345",
			token:      "wrong_token",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			got := auth.Validate(req)
			if got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthValidateNoLocalhostBypass(t *testing.T) {
	auth := NewAuth(false) // require auth for all

	tests := []struct {
		name       string
		remoteAddr string
		token      string
		want       bool
	}{
		{
			name:       "localhost no token - should fail",
			remoteAddr: "127.0.0.1:12345",
			token:      "",
			want:       false,
		},
		{
			name:       "localhost with valid token",
			remoteAddr: "127.0.0.1:12345",
			token:      auth.Token(),
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			got := auth.Validate(req)
			if got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthBearerParsing(t *testing.T) {
	auth := NewAuth(false)

	tests := []struct {
		name   string
		header string
		want   bool
	}{
		{
			name:   "valid bearer",
			header: "Bearer " + auth.Token(),
			want:   true,
		},
		{
			name:   "missing bearer prefix",
			header: auth.Token(),
			want:   false,
		},
		{
			name:   "wrong prefix",
			header: "Basic " + auth.Token(),
			want:   false,
		},
		{
			name:   "empty header",
			header: "",
			want:   false,
		},
		{
			name:   "bearer only",
			header: "Bearer ",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "192.168.1.100:12345" // non-localhost
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			got := auth.Validate(req)
			if got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		remoteAddr string
		want       bool
	}{
		{"127.0.0.1:12345", true},
		{"127.0.0.1", true},
		{"[::1]:12345", true},
		{"::1", true},
		{"192.168.1.100:12345", false},
		{"10.0.0.1:8080", false},
		{"8.8.8.8:443", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.remoteAddr, func(t *testing.T) {
			req := &http.Request{RemoteAddr: tt.remoteAddr}
			got := isLocalhost(req)
			if got != tt.want {
				t.Errorf("isLocalhost(%q) = %v, want %v", tt.remoteAddr, got, tt.want)
			}
		})
	}
}
