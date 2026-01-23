package crush

import (
	"os"
	"path/filepath"
)

// DefaultCrushDataDir returns the default Crush data directory.
// Crush stores its data in .crush/ relative to the working directory.
const CrushDataDirName = ".crush"
const CrushDatabaseName = "crush.db"

// FindCrushDatabase returns the path to Crush's SQLite database.
// It looks in the working directory's .crush/ folder.
func FindCrushDatabase(workingDir string) string {
	if workingDir == "" {
		workingDir, _ = os.Getwd()
	}
	return filepath.Join(workingDir, CrushDataDirName, CrushDatabaseName)
}

// CrushDatabaseExists returns true if Crush's database exists at the expected location.
func CrushDatabaseExists(workingDir string) bool {
	dbPath := FindCrushDatabase(workingDir)
	_, err := os.Stat(dbPath)
	return err == nil
}

// SessionMapping tracks the relationship between an ayo session and a Crush session.
type SessionMapping struct {
	AyoSessionID   string // Session ID in ayo's database
	CrushSessionID string // Session ID in Crush's database (may be empty if not resolved)
	CrushTitle     string // Title of the Crush session (for matching)
	StartedAt      int64  // Unix timestamp when Crush was invoked
	WorkingDir     string // Working directory where Crush was run
}

// NonInteractiveTitlePrefix is the prefix Crush uses for headless session titles.
const NonInteractiveTitlePrefix = "Non-interactive: "
