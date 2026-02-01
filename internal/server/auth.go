package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net"
	"net/http"
	"strings"
)

// Auth handles authentication for the HTTP API.
type Auth struct {
	token         string
	localhostOnly bool
}

// NewAuth creates a new Auth instance with a randomly generated token.
// If localhostOnly is true, requests from localhost bypass token validation.
func NewAuth(localhostOnly bool) *Auth {
	return &Auth{
		token:         generateToken(),
		localhostOnly: localhostOnly,
	}
}

// Token returns the authentication token.
func (a *Auth) Token() string {
	return a.token
}

// Validate checks if the request is authenticated.
// Returns true if:
// - Request is from localhost and localhostOnly is true
// - Request has valid Bearer token
func (a *Auth) Validate(r *http.Request) bool {
	// Check if request is from localhost
	if a.localhostOnly && isLocalhost(r) {
		return true
	}

	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	// Parse Bearer token
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return false
	}
	token := strings.TrimPrefix(authHeader, prefix)

	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(token), []byte(a.token)) == 1
}

// isLocalhost checks if the request originates from localhost.
func isLocalhost(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// Try without port
		host = r.RemoteAddr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	return ip.IsLoopback()
}

// generateToken creates a cryptographically secure random token.
func generateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("failed to generate random token: " + err.Error())
	}
	return "ayo_" + base64.RawURLEncoding.EncodeToString(b)
}
