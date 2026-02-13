package tickets

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

// GenerateID creates a unique ticket ID in the format {prefix}-{random4}.
// The prefix is derived from the directory name (first 4 chars of parent dir).
// The random suffix is 4 hex characters.
func GenerateID(ticketsDir string) (string, error) {
	prefix := derivePrefix(ticketsDir)
	suffix, err := randomSuffix(4)
	if err != nil {
		return "", err
	}
	return prefix + "-" + suffix, nil
}

// GenerateUniqueID generates an ID and ensures it doesn't conflict with existing tickets.
func GenerateUniqueID(ticketsDir string) (string, error) {
	for i := 0; i < 100; i++ { // Max attempts to find unique ID
		id, err := GenerateID(ticketsDir)
		if err != nil {
			return "", err
		}

		// Check if file already exists
		path := filepath.Join(ticketsDir, id+".md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return id, nil
		}
	}

	// Fall back to longer suffix if we can't find unique short ID
	prefix := derivePrefix(ticketsDir)
	suffix, err := randomSuffix(8)
	if err != nil {
		return "", err
	}
	return prefix + "-" + suffix, nil
}

// derivePrefix extracts a short prefix from the tickets directory path.
// Uses the parent directory name (session ID or project name) truncated to 4 chars.
func derivePrefix(ticketsDir string) string {
	// Get parent directory of .tickets
	parent := filepath.Dir(ticketsDir)
	name := filepath.Base(parent)

	// Clean the name - remove common prefixes
	name = strings.TrimPrefix(name, "ses_")
	name = strings.TrimPrefix(name, "session_")

	// Take first 4 alphanumeric characters
	var prefix strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			prefix.WriteRune(r)
			if prefix.Len() >= 4 {
				break
			}
		}
	}

	// Default prefix if name is too short
	if prefix.Len() < 2 {
		return "tk"
	}

	return prefix.String()
}

// randomSuffix generates n random hex characters.
func randomSuffix(n int) (string, error) {
	bytes := make([]byte, (n+1)/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:n], nil
}
