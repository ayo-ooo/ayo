package zettelkasten

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/google/uuid"
)

// Provider implements providers.MemoryProvider using the Zettelkasten file format.
type Provider struct {
	mu        sync.RWMutex
	structure *Structure
	cache     map[string]*MemoryFile // In-memory cache by ID
	inited    bool
}

// NewProvider creates a new zettelkasten memory provider.
func NewProvider() *Provider {
	return &Provider{
		cache: make(map[string]*MemoryFile),
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "zettelkasten"
}

// Type returns the provider type.
func (p *Provider) Type() providers.ProviderType {
	return providers.ProviderTypeMemory
}

// Init initializes the provider with configuration.
func (p *Provider) Init(ctx context.Context, config map[string]any) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Get root directory from config or use default
	root := ""
	if r, ok := config["root"].(string); ok && r != "" {
		root = r
	}

	p.structure = NewStructure(root)

	// Create directory structure
	if err := p.structure.Initialize(); err != nil {
		return fmt.Errorf("initialize structure: %w", err)
	}

	// Load existing memories into cache
	if err := p.loadCache(); err != nil {
		return fmt.Errorf("load cache: %w", err)
	}

	p.inited = true
	return nil
}

// Close releases resources.
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.cache = make(map[string]*MemoryFile)
	p.inited = false
	return nil
}

// loadCache loads all memory files into the in-memory cache.
func (p *Provider) loadCache() error {
	files, err := p.structure.ListAllMemories()
	if err != nil {
		return err
	}

	for _, path := range files {
		mf, err := ParseFile(path)
		if err != nil {
			// Log but don't fail on individual file errors
			continue
		}
		p.cache[mf.Frontmatter.ID] = mf
	}

	return nil
}

// Create stores a new memory.
func (p *Provider) Create(ctx context.Context, m providers.Memory) (providers.Memory, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.inited {
		return providers.Memory{}, fmt.Errorf("provider not initialized")
	}

	// Generate ID if not provided
	if m.ID == "" {
		m.ID = "mem-" + uuid.New().String()[:8]
	}

	// Set timestamps
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now

	// Set defaults
	if m.Status == "" {
		m.Status = providers.MemoryStatusActive
	}
	if m.Confidence == 0 {
		m.Confidence = 1.0
	}

	// Convert to MemoryFile
	mf := providerMemoryToFile(m)

	// Validate
	if err := mf.Frontmatter.Validate(); err != nil {
		return providers.Memory{}, fmt.Errorf("validate: %w", err)
	}

	// Write to disk
	path := p.structure.MemoryPath(m.ID, string(m.Category))
	if err := mf.WriteFile(path); err != nil {
		return providers.Memory{}, fmt.Errorf("write file: %w", err)
	}

	// Update cache
	p.cache[m.ID] = mf

	// Create topic symlinks
	for _, topic := range m.Topics {
		if err := p.structure.LinkToTopic(m.ID, string(m.Category), topic); err != nil {
			// Log but don't fail
			continue
		}
	}

	return fileToProviderMemory(mf), nil
}

// Get retrieves a memory by ID.
func (p *Provider) Get(ctx context.Context, id string) (providers.Memory, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.inited {
		return providers.Memory{}, fmt.Errorf("provider not initialized")
	}

	mf, ok := p.cache[id]
	if !ok {
		return providers.Memory{}, fmt.Errorf("memory not found: %s", id)
	}

	return fileToProviderMemory(mf), nil
}

// Search finds memories matching the query.
// Note: Full semantic search requires the embedding provider and derived index.
// This implementation does basic text matching.
func (p *Provider) Search(ctx context.Context, query string, opts providers.SearchOptions) ([]providers.SearchResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.inited {
		return nil, fmt.Errorf("provider not initialized")
	}

	var results []providers.SearchResult
	queryLower := strings.ToLower(query)

	for _, mf := range p.cache {
		m := fileToProviderMemory(mf)

		// Filter by status (default to active only)
		if len(opts.Status) == 0 {
			if m.Status != providers.MemoryStatusActive {
				continue
			}
		} else {
			match := false
			for _, s := range opts.Status {
				if m.Status == s {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		// Filter by agent
		if opts.AgentHandle != "" && m.AgentHandle != opts.AgentHandle {
			continue
		}

		// Filter by path scope
		if opts.PathScope != "" && !strings.HasPrefix(opts.PathScope, m.PathScope) {
			continue
		}

		// Filter by categories
		if len(opts.Categories) > 0 {
			match := false
			for _, c := range opts.Categories {
				if m.Category == c {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		// Filter by topics
		if len(opts.Topics) > 0 {
			match := false
			for _, t := range opts.Topics {
				for _, mt := range m.Topics {
					if t == mt {
						match = true
						break
					}
				}
				if match {
					break
				}
			}
			if !match {
				continue
			}
		}

		// Basic text matching
		contentLower := strings.ToLower(mf.Content)
		if strings.Contains(contentLower, queryLower) {
			// Simple similarity score based on match position and length
			similarity := float32(len(queryLower)) / float32(len(contentLower)+1)
			if similarity > 1 {
				similarity = 1
			}

			if opts.Threshold > 0 && similarity < opts.Threshold {
				continue
			}

			results = append(results, providers.SearchResult{
				Memory:     m,
				Similarity: similarity,
				MatchType:  "text",
			})
		}
	}

	// Apply limit
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// List returns all memories matching the filter.
func (p *Provider) List(ctx context.Context, opts providers.ListOptions) ([]providers.Memory, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.inited {
		return nil, fmt.Errorf("provider not initialized")
	}

	var memories []providers.Memory

	for _, mf := range p.cache {
		m := fileToProviderMemory(mf)

		// Filter by status (default to active only)
		if len(opts.Status) == 0 {
			if m.Status != providers.MemoryStatusActive {
				continue
			}
		} else {
			match := false
			for _, s := range opts.Status {
				if m.Status == s {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		// Filter by agent
		if opts.AgentHandle != "" && m.AgentHandle != opts.AgentHandle {
			continue
		}

		// Filter by path scope
		if opts.PathScope != "" && m.PathScope != opts.PathScope {
			continue
		}

		// Filter by categories
		if len(opts.Categories) > 0 {
			match := false
			for _, c := range opts.Categories {
				if m.Category == c {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		// Filter by topics
		if len(opts.Topics) > 0 {
			match := false
			for _, t := range opts.Topics {
				for _, mt := range m.Topics {
					if t == mt {
						match = true
						break
					}
				}
				if match {
					break
				}
			}
			if !match {
				continue
			}
		}

		memories = append(memories, m)
	}

	// Apply offset and limit
	if opts.Offset > 0 {
		if opts.Offset >= len(memories) {
			return nil, nil
		}
		memories = memories[opts.Offset:]
	}

	if opts.Limit > 0 && len(memories) > opts.Limit {
		memories = memories[:opts.Limit]
	}

	return memories, nil
}

// Update modifies an existing memory.
func (p *Provider) Update(ctx context.Context, m providers.Memory) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.inited {
		return fmt.Errorf("provider not initialized")
	}

	// Check if memory exists
	existing, ok := p.cache[m.ID]
	if !ok {
		return fmt.Errorf("memory not found: %s", m.ID)
	}

	// Update timestamp
	m.UpdatedAt = time.Now().UTC()

	// Convert to file
	mf := providerMemoryToFile(m)

	// Validate
	if err := mf.Frontmatter.Validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	// Handle category change - need to move file
	oldCategory := existing.Frontmatter.Category
	newCategory := string(m.Category)

	oldPath := p.structure.MemoryPath(m.ID, oldCategory)
	newPath := p.structure.MemoryPath(m.ID, newCategory)

	if oldCategory != newCategory {
		// Remove old file
		os.Remove(oldPath)

		// Remove old topic links
		for _, topic := range existing.Frontmatter.Topics {
			p.structure.UnlinkFromTopic(m.ID, topic)
		}
	}

	// Write new file
	if err := mf.WriteFile(newPath); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// Update cache
	p.cache[m.ID] = mf

	// Update topic symlinks
	for _, topic := range m.Topics {
		if err := p.structure.LinkToTopic(m.ID, newCategory, topic); err != nil {
			continue
		}
	}

	return nil
}

// Forget soft-deletes a memory.
func (p *Provider) Forget(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.inited {
		return fmt.Errorf("provider not initialized")
	}

	mf, ok := p.cache[id]
	if !ok {
		return fmt.Errorf("memory not found: %s", id)
	}

	// Update status to forgotten
	mf.Frontmatter.Status = "forgotten"
	mf.Frontmatter.Updated = time.Now().UTC()

	// Write updated file
	path := p.structure.MemoryPath(id, mf.Frontmatter.Category)
	if err := mf.WriteFile(path); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// Update cache
	p.cache[id] = mf

	return nil
}

// Supersede replaces an old memory with a new one.
func (p *Provider) Supersede(ctx context.Context, oldID string, newMemory providers.Memory, reason string) (providers.Memory, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.inited {
		return providers.Memory{}, fmt.Errorf("provider not initialized")
	}

	// Get old memory
	oldMF, ok := p.cache[oldID]
	if !ok {
		return providers.Memory{}, fmt.Errorf("memory not found: %s", oldID)
	}

	// Generate ID for new memory
	if newMemory.ID == "" {
		newMemory.ID = "mem-" + uuid.New().String()[:8]
	}

	now := time.Now().UTC()
	newMemory.CreatedAt = now
	newMemory.UpdatedAt = now
	newMemory.SupersedesID = oldID
	newMemory.SupersessionReason = reason

	if newMemory.Status == "" {
		newMemory.Status = providers.MemoryStatusActive
	}
	if newMemory.Confidence == 0 {
		newMemory.Confidence = 1.0
	}

	// Create new memory file
	newMF := providerMemoryToFile(newMemory)
	if err := newMF.Frontmatter.Validate(); err != nil {
		return providers.Memory{}, fmt.Errorf("validate: %w", err)
	}

	// Write new memory
	newPath := p.structure.MemoryPath(newMemory.ID, string(newMemory.Category))
	if err := newMF.WriteFile(newPath); err != nil {
		return providers.Memory{}, fmt.Errorf("write new memory: %w", err)
	}

	// Update old memory
	oldMF.Frontmatter.Status = "superseded"
	oldMF.Frontmatter.Updated = now
	oldMF.Frontmatter.Supersession.SupersededBy = newMemory.ID
	oldMF.Frontmatter.Supersession.Reason = reason

	oldPath := p.structure.MemoryPath(oldID, oldMF.Frontmatter.Category)
	if err := oldMF.WriteFile(oldPath); err != nil {
		return providers.Memory{}, fmt.Errorf("update old memory: %w", err)
	}

	// Update cache
	p.cache[oldID] = oldMF
	p.cache[newMemory.ID] = newMF

	// Create topic symlinks for new memory
	for _, topic := range newMemory.Topics {
		p.structure.LinkToTopic(newMemory.ID, string(newMemory.Category), topic)
	}

	return fileToProviderMemory(newMF), nil
}

// Topics returns all known topics.
func (p *Provider) Topics(ctx context.Context) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.inited {
		return nil, fmt.Errorf("provider not initialized")
	}

	return p.structure.ListTopics()
}

// Link creates a bidirectional link between two memories.
func (p *Provider) Link(ctx context.Context, id1, id2 string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.inited {
		return fmt.Errorf("provider not initialized")
	}

	mf1, ok := p.cache[id1]
	if !ok {
		return fmt.Errorf("memory not found: %s", id1)
	}

	mf2, ok := p.cache[id2]
	if !ok {
		return fmt.Errorf("memory not found: %s", id2)
	}

	now := time.Now().UTC()

	// Add link to mf1
	if !contains(mf1.Frontmatter.Links.Related, id2) {
		mf1.Frontmatter.Links.Related = append(mf1.Frontmatter.Links.Related, id2)
		mf1.Frontmatter.Updated = now
		path := p.structure.MemoryPath(id1, mf1.Frontmatter.Category)
		if err := mf1.WriteFile(path); err != nil {
			return fmt.Errorf("write file %s: %w", id1, err)
		}
	}

	// Add link to mf2
	if !contains(mf2.Frontmatter.Links.Related, id1) {
		mf2.Frontmatter.Links.Related = append(mf2.Frontmatter.Links.Related, id1)
		mf2.Frontmatter.Updated = now
		path := p.structure.MemoryPath(id2, mf2.Frontmatter.Category)
		if err := mf2.WriteFile(path); err != nil {
			return fmt.Errorf("write file %s: %w", id2, err)
		}
	}

	return nil
}

// Unlink removes a link between two memories.
func (p *Provider) Unlink(ctx context.Context, id1, id2 string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.inited {
		return fmt.Errorf("provider not initialized")
	}

	mf1, ok := p.cache[id1]
	if !ok {
		return fmt.Errorf("memory not found: %s", id1)
	}

	mf2, ok := p.cache[id2]
	if !ok {
		return fmt.Errorf("memory not found: %s", id2)
	}

	now := time.Now().UTC()

	// Remove link from mf1
	mf1.Frontmatter.Links.Related = removeString(mf1.Frontmatter.Links.Related, id2)
	mf1.Frontmatter.Updated = now
	path1 := p.structure.MemoryPath(id1, mf1.Frontmatter.Category)
	if err := mf1.WriteFile(path1); err != nil {
		return fmt.Errorf("write file %s: %w", id1, err)
	}

	// Remove link from mf2
	mf2.Frontmatter.Links.Related = removeString(mf2.Frontmatter.Links.Related, id1)
	mf2.Frontmatter.Updated = now
	path2 := p.structure.MemoryPath(id2, mf2.Frontmatter.Category)
	if err := mf2.WriteFile(path2); err != nil {
		return fmt.Errorf("write file %s: %w", id2, err)
	}

	return nil
}

// Reindex rebuilds the in-memory cache from files.
func (p *Provider) Reindex(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.inited {
		return fmt.Errorf("provider not initialized")
	}

	// Clear cache
	p.cache = make(map[string]*MemoryFile)

	// Reload
	return p.loadCache()
}

// Helper functions

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

// providerMemoryToFile converts a providers.Memory to a MemoryFile.
func providerMemoryToFile(m providers.Memory) *MemoryFile {
	return &MemoryFile{
		Frontmatter: Frontmatter{
			ID:         m.ID,
			Created:    m.CreatedAt,
			Updated:    m.UpdatedAt,
			Category:   string(m.Category),
			Status:     string(m.Status),
			Topics:     m.Topics,
			Confidence: m.Confidence,
			Source: SourceSection{
				SessionID: m.SourceSessionID,
				MessageID: m.SourceMessageID,
			},
			Scope: ScopeSection{
				Agent: m.AgentHandle,
				Path:  m.PathScope,
			},
			Access: AccessSection{
				LastAccessed: m.LastAccessedAt,
				AccessCount:  m.AccessCount,
			},
			Supersession: SupersessionSection{
				Supersedes:   m.SupersedesID,
				SupersededBy: m.SupersededByID,
				Reason:       m.SupersessionReason,
			},
			Links: LinksSection{
				Related: m.LinkedIDs,
			},
			Unclear: UnclearSection{
				Flagged: m.Unclear,
				Reason:  m.UnclearReason,
			},
		},
		Content: m.Content,
	}
}

// fileToProviderMemory converts a MemoryFile to a providers.Memory.
func fileToProviderMemory(mf *MemoryFile) providers.Memory {
	fm := mf.Frontmatter
	return providers.Memory{
		ID:                 fm.ID,
		Content:            mf.Content,
		Category:           providers.MemoryCategory(fm.Category),
		Topics:             fm.Topics,
		AgentHandle:        fm.Scope.Agent,
		PathScope:          fm.Scope.Path,
		SourceSessionID:    fm.Source.SessionID,
		SourceMessageID:    fm.Source.MessageID,
		CreatedAt:          fm.Created,
		UpdatedAt:          fm.Updated,
		Confidence:         fm.Confidence,
		LastAccessedAt:     fm.Access.LastAccessed,
		AccessCount:        fm.Access.AccessCount,
		SupersedesID:       fm.Supersession.Supersedes,
		SupersededByID:     fm.Supersession.SupersededBy,
		SupersessionReason: fm.Supersession.Reason,
		Status:             providers.MemoryStatus(fm.Status),
		Unclear:            fm.Unclear.Flagged,
		UnclearReason:      fm.Unclear.Reason,
		LinkedIDs:          fm.Links.Related,
	}
}

// Ensure Provider implements MemoryProvider
var _ providers.MemoryProvider = (*Provider)(nil)
