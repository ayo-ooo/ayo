package server

import (
	_ "embed"
	"net/http"
)

//go:embed webclient/index.html
var webclientHTML []byte

// handleWebClient serves the embedded web client.
func (s *Server) handleWebClient(w http.ResponseWriter, r *http.Request) {
	// Only serve at root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(webclientHTML)
}
