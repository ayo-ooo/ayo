package squads

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/paths"
)

// MigrationResult contains the result of a migration operation.
type MigrationResult struct {
	// SquadName is the name of the squad.
	SquadName string

	// Migrated indicates whether the migration was performed.
	Migrated bool

	// Skipped indicates whether the migration was skipped (already migrated).
	Skipped bool

	// Error contains any error that occurred during migration.
	Error error

	// AyoJSONPath is the path to the generated ayo.json.
	AyoJSONPath string

	// SquadMDPath is the path to the stripped SQUAD.md.
	SquadMDPath string
}

// MigrateOptions configures migration behavior.
type MigrateOptions struct {
	// DryRun if true, shows what would change without modifying files.
	DryRun bool

	// Force if true, overwrites existing ayo.json.
	Force bool
}

// MigrateSquadConfig migrates a squad from SQUAD.md frontmatter to ayo.json format.
// Returns a MigrationResult describing what was done.
func MigrateSquadConfig(squadName string, opts MigrateOptions) MigrationResult {
	squadDir := paths.SquadDir(squadName)
	return migrateSquadDir(squadName, squadDir, opts)
}

// migrateSquadDir performs migration on a specific directory.
// This is the internal implementation that can be tested with arbitrary paths.
func migrateSquadDir(squadName, squadDir string, opts MigrateOptions) MigrationResult {
	result := MigrationResult{
		SquadName: squadName,
	}

	squadMDPath := filepath.Join(squadDir, "SQUAD.md")
	ayoJSONPath := filepath.Join(squadDir, "ayo.json")

	result.SquadMDPath = squadMDPath
	result.AyoJSONPath = ayoJSONPath

	// Check if ayo.json already exists
	if _, err := os.Stat(ayoJSONPath); err == nil && !opts.Force {
		result.Skipped = true
		debug.Log("squad already migrated", "squad", squadName)
		return result
	}

	// Read SQUAD.md
	data, err := os.ReadFile(squadMDPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Skipped = true
			result.Error = fmt.Errorf("SQUAD.md not found")
			return result
		}
		result.Error = fmt.Errorf("read SQUAD.md: %w", err)
		return result
	}

	// Parse frontmatter
	frontmatter, body, err := parseFrontmatter(string(data))
	if err != nil {
		result.Error = fmt.Errorf("parse frontmatter: %w", err)
		return result
	}

	// Check if there's any frontmatter to migrate
	if frontmatter.Lead == "" && len(frontmatter.Agents) == 0 &&
		frontmatter.Planners.IsEmpty() && frontmatter.InputAccepts == "" {
		result.Skipped = true
		debug.Log("no frontmatter to migrate", "squad", squadName)
		return result
	}

	// Convert to ayo.json format
	squadCfg := &config.SquadConfig{
		Name:         squadName,
		Lead:         frontmatter.Lead,
		Agents:       frontmatter.Agents,
		InputAccepts: frontmatter.InputAccepts,
	}

	// Convert planners
	if !frontmatter.Planners.IsEmpty() {
		squadCfg.Planners = &config.SquadPlannersConfig{
			NearTerm: frontmatter.Planners.NearTerm,
			LongTerm: frontmatter.Planners.LongTerm,
		}
	}

	ayoCfg := AyoConfig{
		Schema:  "https://ayo.dev/schemas/ayo.json",
		Version: "1",
		Squad:   squadCfg,
	}

	// Dry run - just return what would happen
	if opts.DryRun {
		result.Migrated = false
		return result
	}

	// Write ayo.json
	cfgData, err := json.MarshalIndent(ayoCfg, "", "  ")
	if err != nil {
		result.Error = fmt.Errorf("marshal ayo.json: %w", err)
		return result
	}

	if err := os.WriteFile(ayoJSONPath, cfgData, 0644); err != nil {
		result.Error = fmt.Errorf("write ayo.json: %w", err)
		return result
	}

	// Strip frontmatter from SQUAD.md
	cleanBody := strings.TrimSpace(body)
	if err := os.WriteFile(squadMDPath, []byte(cleanBody+"\n"), 0644); err != nil {
		result.Error = fmt.Errorf("write SQUAD.md: %w", err)
		return result
	}

	result.Migrated = true
	debug.Log("migrated squad config", "squad", squadName)
	return result
}

// MigrateAllSquads migrates all squads from SQUAD.md frontmatter to ayo.json format.
func MigrateAllSquads(opts MigrateOptions) ([]MigrationResult, error) {
	names, err := config.ListSquadConfigs()
	if err != nil {
		return nil, fmt.Errorf("list squads: %w", err)
	}

	var results []MigrationResult
	for _, name := range names {
		result := MigrateSquadConfig(name, opts)
		results = append(results, result)
	}

	return results, nil
}

// NeedsMigration returns true if a squad has SQUAD.md frontmatter but no ayo.json.
func NeedsMigration(squadName string) bool {
	squadDir := paths.SquadDir(squadName)
	return needsMigrationDir(squadDir)
}

// needsMigrationDir checks if a directory needs migration.
func needsMigrationDir(squadDir string) bool {
	ayoJSONPath := filepath.Join(squadDir, "ayo.json")
	squadMDPath := filepath.Join(squadDir, "SQUAD.md")

	// If ayo.json exists, no migration needed
	if _, err := os.Stat(ayoJSONPath); err == nil {
		return false
	}

	// Check if SQUAD.md has frontmatter
	data, err := os.ReadFile(squadMDPath)
	if err != nil {
		return false
	}

	frontmatter, _, err := parseFrontmatter(string(data))
	if err != nil {
		return false
	}

	// Has frontmatter if any fields are set
	return frontmatter.Lead != "" || len(frontmatter.Agents) > 0 ||
		!frontmatter.Planners.IsEmpty() || frontmatter.InputAccepts != ""
}

// DeprecationWarning returns a deprecation message if the squad needs migration.
func DeprecationWarning(squadName string) string {
	if NeedsMigration(squadName) {
		return fmt.Sprintf("⚠️  Squad '%s' uses deprecated SQUAD.md frontmatter.\n"+
			"    Run 'ayo migrate squad %s' to update.", squadName, squadName)
	}
	return ""
}

// deprecationWarningDir returns a deprecation message for a directory.
func deprecationWarningDir(squadName, squadDir string) string {
	if needsMigrationDir(squadDir) {
		return fmt.Sprintf("⚠️  Squad '%s' uses deprecated SQUAD.md frontmatter.\n"+
			"    Run 'ayo migrate squad %s' to update.", squadName, squadName)
	}
	return ""
}
