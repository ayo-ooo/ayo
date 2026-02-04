package zettelkasten

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/alexcabrera/ayo/internal/db"
	"github.com/alexcabrera/ayo/internal/providers"
)

// MigrationResult summarizes the migration operation.
type MigrationResult struct {
	Migrated int      // Number of memories migrated
	Skipped  int      // Number already existing
	Failed   int      // Number that failed
	Errors   []string // Error messages
}

// MigrateFromSQLite migrates memories from the SQLite database to zettelkasten files.
// It reads all memories from the database and creates corresponding markdown files.
// Existing files are skipped unless overwrite is true.
func MigrateFromSQLite(ctx context.Context, queries *db.Queries, provider *Provider, overwrite bool) (*MigrationResult, error) {
	result := &MigrationResult{}

	// Get all memories from SQLite
	dbMemories, err := queries.ListMemories(ctx, db.ListMemoriesParams{
		Status: sql.NullString{String: "", Valid: false}, // All statuses
		Off:    0,
		Lim:    10000, // High limit to get all
	})
	if err != nil {
		return nil, fmt.Errorf("list memories: %w", err)
	}

	for _, dbMem := range dbMemories {
		// Check if already exists in zettelkasten
		if !overwrite {
			if _, err := provider.Get(ctx, dbMem.ID); err == nil {
				result.Skipped++
				continue
			}
		}

		// Convert to providers.Memory
		mem := dbMemoryToProviderMemory(dbMem)

		// Create in zettelkasten
		if _, err := provider.Create(ctx, mem); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", dbMem.ID, err))
			continue
		}

		result.Migrated++
	}

	return result, nil
}

// dbMemoryToProviderMemory converts a database memory row to a providers.Memory.
func dbMemoryToProviderMemory(dbMem db.Memory) providers.Memory {
	mem := providers.Memory{
		ID:          dbMem.ID,
		Content:     dbMem.Content,
		Category:    providers.MemoryCategory(dbMem.Category),
		AgentHandle: nullStringValue(dbMem.AgentHandle),
		PathScope:   nullStringValue(dbMem.PathScope),
		Confidence:  nullFloat64Value(dbMem.Confidence, 1.0),
		Status:      providers.MemoryStatus(nullStringValueDefault(dbMem.Status, "active")),
	}

	// Convert timestamps (stored as unix seconds in SQLite)
	mem.CreatedAt = time.Unix(dbMem.CreatedAt, 0).UTC()
	mem.UpdatedAt = time.Unix(dbMem.UpdatedAt, 0).UTC()

	if dbMem.LastAccessedAt.Valid {
		mem.LastAccessedAt = time.Unix(dbMem.LastAccessedAt.Int64, 0).UTC()
	}
	mem.AccessCount = nullInt64Value(dbMem.AccessCount)

	// Source provenance
	mem.SourceSessionID = nullStringValue(dbMem.SourceSessionID)
	mem.SourceMessageID = nullStringValue(dbMem.SourceMessageID)

	// Supersession chain
	mem.SupersedesID = nullStringValue(dbMem.SupersedesID)
	mem.SupersededByID = nullStringValue(dbMem.SupersededByID)
	mem.SupersessionReason = nullStringValue(dbMem.SupersessionReason)

	// Embedding (stored as blob)
	if len(dbMem.Embedding) > 0 {
		mem.Embedding = bytesToFloat32(dbMem.Embedding)
	}

	return mem
}

// Helper functions for handling nullable database fields

func nullStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullStringValueDefault(ns sql.NullString, def string) string {
	if ns.Valid && ns.String != "" {
		return ns.String
	}
	return def
}

func nullFloat64Value(nf sql.NullFloat64, def float64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return def
}

func nullInt64Value(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}
