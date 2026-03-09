// Package capabilities provides OpenClaw discovery service for Ayo build system.
//
// This service enables discovery of OpenClaw components from various sources
// including GitHub repositories, local directories, and the OpenClaw registry.
package capabilities

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// OpenClawDiscoveryService provides discovery of OpenClaw components.
type OpenClawDiscoveryService struct {
	// Client is the HTTP client for making requests.
	Client *http.Client

	// GitHubClient is the GitHub API client.
	GitHubClient *github.Client

	// CacheDir is the directory for caching discovery results.
	CacheDir string

	// RegistryURL is the OpenClaw registry URL.
	RegistryURL string

	// RegistryAPIKey is the API key for authenticated registry access.
	RegistryAPIKey string

	// Timeout is the request timeout duration.
	Timeout time.Duration

	// UserAgent is the HTTP user agent for requests.
	UserAgent string

	// EnableCache controls whether to use caching.
	EnableCache bool
}

// NewOpenClawDiscoveryService creates a new OpenClaw discovery service.
func NewOpenClawDiscoveryService() *OpenClawDiscoveryService {
	return &OpenClawDiscoveryService{
		Client:      &http.Client{Timeout: 30 * time.Second},
		CacheDir:    filepath.Join("~", ".cache", "ayo", "openclaw-discovery"),
		RegistryURL: "https://registry.openclaw.io", // Default OpenClaw registry
		Timeout:     30 * time.Second,
		UserAgent:   "AyoOpenClawDiscovery/1.0",
		EnableCache: true,
	}
}

// WithGitHubToken configures the discovery service with a GitHub token.
func (s *OpenClawDiscoveryService) WithGitHubToken(token string) *OpenClawDiscoveryService {
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		s.GitHubClient = github.NewClient(oauth2.NewClient(context.Background(), ts))
	}
	return s
}

// WithRegistryURL configures a custom OpenClaw registry URL.
func (s *OpenClawDiscoveryService) WithRegistryURL(registryURL string) *OpenClawDiscoveryService {
	s.RegistryURL = registryURL
	return s
}

// WithRegistryAPIKey configures the OpenClaw registry API key for authentication.
func (s *OpenClawDiscoveryService) WithRegistryAPIKey(apiKey string) *OpenClawDiscoveryService {
	s.RegistryAPIKey = apiKey
	return s
}

// WithTimeout configures the request timeout.
func (s *OpenClawDiscoveryService) WithTimeout(timeout time.Duration) *OpenClawDiscoveryService {
	s.Timeout = timeout
	s.Client.Timeout = timeout
	return s
}

// WithUserAgent configures the HTTP user agent.
func (s *OpenClawDiscoveryService) WithUserAgent(userAgent string) *OpenClawDiscoveryService {
	s.UserAgent = userAgent
	return s
}

// WithCacheEnablement controls whether caching is enabled.
func (s *OpenClawDiscoveryService) WithCacheEnablement(enable bool) *OpenClawDiscoveryService {
	s.EnableCache = enable
	return s
}

// createRegistryRequest creates an HTTP request for the OpenClaw registry with proper headers.
func (s *OpenClawDiscoveryService) createRegistryRequest(method, path string, body io.Reader) (*http.Request, error) {
	url := s.RegistryURL + path
	if !strings.HasSuffix(s.RegistryURL, "/") && !strings.HasPrefix(path, "/") {
		url = s.RegistryURL + "/" + path
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", s.UserAgent)
	req.Header.Set("Accept", "application/json")
	
	if s.RegistryAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.RegistryAPIKey)
	}

	return req, nil
}

// DiscoverFromGitHub searches GitHub for OpenClaw plugins.
func (s *OpenClawDiscoveryService) DiscoverFromGitHub(query string, limit int) ([]OpenClawDiscoveryResult, error) {
	if s.GitHubClient == nil {
		// Use unauthenticated client if no token provided
		s.GitHubClient = github.NewClient(nil)
	}

	// Search for repositories matching the query
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: limit},
	}

	results, _, err := s.GitHubClient.Search.Repositories(context.Background(), query, opt)
	if err != nil {
		return nil, fmt.Errorf("GitHub search failed: %w", err)
	}

	var discoveries []OpenClawDiscoveryResult
	for _, repo := range results.Repositories {
		// Check if this looks like an OpenClaw plugin
		if isOpenClawRepository(repo) {
			discoveries = append(discoveries, OpenClawDiscoveryResult{
				Source:      "github",
				Name:        repo.GetName(),
				Description: repo.GetDescription(),
				URL:         repo.GetHTMLURL(),
				CloneURL:    repo.GetCloneURL(),
				Stars:       repo.GetStargazersCount(),
				UpdatedAt:   repo.GetUpdatedAt().Time,
				Metadata: map[string]interface{}{
					"owner": repo.GetOwner().GetLogin(),
				},
			})
		}
	}

	return discoveries, nil
}

// isOpenClawRepository checks if a GitHub repository looks like an OpenClaw plugin.
func isOpenClawRepository(repo *github.Repository) bool {
	// Check for OpenClaw-related keywords in name, description, or topics
	repoName := strings.ToLower(repo.GetName())
	repoDesc := strings.ToLower(repo.GetDescription())
	
	keywords := []string{"openclaw", "skill", "plugin", "component", "agent"}
	
	for _, keyword := range keywords {
		if strings.Contains(repoName, keyword) || strings.Contains(repoDesc, keyword) {
			return true
		}
	}
	
	// Check topics
	for _, topic := range repo.Topics {
		topicLower := strings.ToLower(topic)
		for _, keyword := range keywords {
			if strings.Contains(topicLower, keyword) {
				return true
			}
		}
	}
	
	return false
}

// DiscoverFromLocalDirectory searches a local directory for OpenClaw components.
func (s *OpenClawDiscoveryService) DiscoverFromLocalDirectory(dir string) ([]OpenClawDiscoveryResult, error) {
	var discoveries []OpenClawDiscoveryResult

	// Look for OpenClaw manifest files
	manifestFiles := findOpenClawManifests(dir)
	
	for _, manifestPath := range manifestFiles {
		// Try to load the manifest
		provider := NewOpenClawSkillProvider(dir)
		skills, err := provider.DiscoverSkills()
		if err != nil {
			continue // Skip directories with invalid manifests
		}

		if len(skills) > 0 {
			discoveries = append(discoveries, OpenClawDiscoveryResult{
				Source:      "local",
				Name:        filepath.Base(dir),
				Description: fmt.Sprintf("Local OpenClaw components in %s", dir),
				URL:         dir,
				ComponentCount: len(skills),
				Metadata: map[string]interface{}{
					"manifest_path": manifestPath,
					"skills":        len(skills),
				},
			})
		}
	}

	return discoveries, nil
}

// DiscoverFromRegistry queries the OpenClaw registry for components.
func (s *OpenClawDiscoveryService) DiscoverFromRegistry(query string, limit int) ([]OpenClawDiscoveryResult, error) {
	// Check cache first if enabled
	if s.EnableCache {
		if cachedResults, err := s.LoadCachedDiscoveryResults("registry-" + query); err == nil && len(cachedResults) > 0 {
			return cachedResults, nil
		}
	}

	// Build the registry API path
	apiPath := fmt.Sprintf("/api/search?q=%s&limit=%d", url.QueryEscape(query), limit)

	// Create authenticated request
	req, err := s.createRegistryRequest("GET", apiPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry request: %w", err)
	}

	// Execute request with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	req = req.WithContext(ctx)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registry request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle different status codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Success - parse response
		
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("registry authentication failed. Please check your API key")
		
	case http.StatusNotFound:
		return nil, fmt.Errorf("no results found for query: %s", query)
		
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("registry rate limit exceeded. Please try again later")
		
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("registry server error. Please try again later")
		
	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("registry returned unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var registryResults []OpenClawRegistryResult
	err = json.NewDecoder(resp.Body).Decode(&registryResults)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry response: %w", err)
	}

	var discoveries []OpenClawDiscoveryResult
	for _, result := range registryResults {
		// Validate component data
		if result.Name == "" {
			continue // Skip invalid entries
		}

		componentCount := len(result.Components)
		if componentCount == 0 && result.Metadata != nil {
			if count, ok := result.Metadata["component_count"].(float64); ok {
				componentCount = int(count)
			}
		}

		discoveries = append(discoveries, OpenClawDiscoveryResult{
			Source:        "registry",
			Name:          result.Name,
			Description:   result.Description,
			URL:           result.URL,
			Version:       result.Version,
			Downloads:     result.Downloads,
			ComponentCount: componentCount,
			UpdatedAt:     result.UpdatedAt,
			Metadata:      result.Metadata,
		})
	}

	// Cache results if enabled
	if s.EnableCache && len(discoveries) > 0 {
		err = s.CacheDiscoveryResults("registry-"+query, discoveries)
		if err != nil {
			// Log cache error but don't fail the discovery
		}
	}

	return discoveries, nil
}

// OpenClawComponent represents a component in an OpenClaw plugin.
// This is duplicated from the adapter to avoid circular imports.
type OpenClawComponent struct {
	// Name is the component identifier.
	Name string `json:"name"`

	// Type is the component type (skill, tool, agent, etc.).
	Type string `json:"type"`

	// Description describes what the component does.
	Description string `json:"description"`

	// EntryPoint is the path to the component implementation.
	EntryPoint string `json:"entry_point"`

	// Config contains component-specific configuration.
	Config map[string]interface{} `json:"config,omitempty"`

	// Metadata contains additional component metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// OpenClawDiscoveryResult represents a discovered OpenClaw component or plugin.
type OpenClawDiscoveryResult struct {
	// Source indicates where the component was discovered (github, local, registry).
	Source string `json:"source"`

	// Name is the component or plugin name.
	Name string `json:"name"`

	// Description describes what the component does.
	Description string `json:"description,omitempty"`

	// URL is the location where the component can be accessed.
	URL string `json:"url"`

	// CloneURL is the git clone URL (for GitHub sources).
	CloneURL string `json:"clone_url,omitempty"`

	// Version is the component version.
	Version string `json:"version,omitempty"`

	// ComponentCount is the number of components in a plugin.
	ComponentCount int `json:"component_count,omitempty"`

	// Downloads is the number of downloads (for registry sources).
	Downloads int `json:"downloads,omitempty"`

	// Stars is the number of GitHub stars (for GitHub sources).
	Stars int `json:"stars,omitempty"`

	// UpdatedAt is when the component was last updated.
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// Metadata contains additional source-specific information.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// OpenClawRegistryResult represents a result from the OpenClaw registry API.
type OpenClawRegistryResult struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	URL         string                 `json:"url"`
	Version     string                 `json:"version"`
	Downloads   int                    `json:"downloads"`
	Components  []OpenClawComponent    `json:"components"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SearchOpenClawComponents searches multiple sources for OpenClaw components.
func (s *OpenClawDiscoveryService) SearchOpenClawComponents(query string, limit int) ([]OpenClawDiscoveryResult, error) {
	var allResults []OpenClawDiscoveryResult
	var errors []error

	// Search GitHub
	githubResults, err := s.DiscoverFromGitHub(query, limit)
	if err != nil {
		errors = append(errors, fmt.Errorf("GitHub search failed: %w", err))
	} else {
		allResults = append(allResults, githubResults...)
	}

	// Search registry
	registryResults, err := s.DiscoverFromRegistry(query, limit)
	if err != nil {
		errors = append(errors, fmt.Errorf("registry search failed: %w", err))
	} else {
		allResults = append(allResults, registryResults...)
	}

	// If we have results, return them even if some sources failed
	if len(allResults) > 0 {
		return allResults, nil
	}

	// If no results and we have errors, return the first error
	if len(errors) > 0 {
		return nil, errors[0]
	}

	return allResults, nil
}

// DiscoverLocalOpenClawProjects finds OpenClaw projects in local directories.
func (s *OpenClawDiscoveryService) DiscoverLocalOpenClawProjects(searchPaths []string) ([]OpenClawDiscoveryResult, error) {
	var allResults []OpenClawDiscoveryResult

	for _, searchPath := range searchPaths {
		results, err := s.discoverInPath(searchPath)
		if err != nil {
			return nil, fmt.Errorf("error searching %s: %w", searchPath, err)
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// discoverInPath searches for OpenClaw projects in a specific directory.
func (s *OpenClawDiscoveryService) discoverInPath(basePath string) ([]OpenClawDiscoveryResult, error) {
	var results []OpenClawDiscoveryResult

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		// Skip hidden directories
		if strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// Check for OpenClaw manifest files
		manifestFiles := findOpenClawManifests(path)
		if len(manifestFiles) > 0 {
			// Found an OpenClaw project
			provider := NewOpenClawSkillProvider(path)
			skills, err := provider.DiscoverSkills()
			if err != nil {
				return nil // Skip directories with invalid manifests
			}

			results = append(results, OpenClawDiscoveryResult{
				Source:        "local",
				Name:          filepath.Base(path),
				Description:   fmt.Sprintf("OpenClaw project at %s", path),
				URL:           path,
				ComponentCount: len(skills),
				Metadata: map[string]interface{}{
					"path":         path,
					"manifests":    manifestFiles,
					"skill_count": len(skills),
				},
			})
		}

		return nil
	})

	return results, err
}

// GetOpenClawComponentInfo retrieves detailed information about a specific component from the registry.
func (s *OpenClawDiscoveryService) GetOpenClawComponentInfo(componentName string) (*OpenClawRegistryResult, error) {
	// Check cache first if enabled
	if s.EnableCache {
		if cachedResults, err := s.LoadCachedDiscoveryResults("component-" + componentName); err == nil && len(cachedResults) > 0 {
			// Convert cached discovery result to registry result
			if len(cachedResults) > 0 {
				return &OpenClawRegistryResult{
					Name:        cachedResults[0].Name,
					Description: cachedResults[0].Description,
					URL:         cachedResults[0].URL,
					Version:     cachedResults[0].Version,
					Downloads:   cachedResults[0].Downloads,
					UpdatedAt:   cachedResults[0].UpdatedAt,
					Metadata:    cachedResults[0].Metadata,
				}, nil
			}
		}
	}

	// Create API request
	apiPath := fmt.Sprintf("/api/components/%s", url.PathEscape(componentName))

	req, err := s.createRegistryRequest("GET", apiPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create component info request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	req = req.WithContext(ctx)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("component info request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get component info: status %d - %s", resp.StatusCode, string(body))
	}

	var componentInfo OpenClawRegistryResult
	err = json.NewDecoder(resp.Body).Decode(&componentInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse component info: %w", err)
	}

	// Cache result if enabled
	if s.EnableCache {
		discoveryResult := OpenClawDiscoveryResult{
			Source:        "registry",
			Name:          componentInfo.Name,
			Description:   componentInfo.Description,
			URL:           componentInfo.URL,
			Version:       componentInfo.Version,
			Downloads:     componentInfo.Downloads,
			ComponentCount: len(componentInfo.Components),
			UpdatedAt:     componentInfo.UpdatedAt,
			Metadata:      componentInfo.Metadata,
		}
		err = s.CacheDiscoveryResults("component-"+componentName, []OpenClawDiscoveryResult{discoveryResult})
		if err != nil {
			// Log cache error but don't fail
		}
	}

	return &componentInfo, nil
}

// GetOpenClawComponentInstallationURL gets the installation URL for a component.
func (s *OpenClawDiscoveryService) GetOpenClawComponentInstallationURL(componentName, version string) (string, error) {
	componentInfo, err := s.GetOpenClawComponentInfo(componentName)
	if err != nil {
		return "", fmt.Errorf("failed to get component info: %w", err)
	}

	// Use specific version if provided, otherwise use latest
	if version != "" {
		return fmt.Sprintf("%s@%s", componentInfo.URL, version), nil
	}

	return componentInfo.URL, nil
}

// SearchOpenClawComponentsWithFilters searches with additional filters.
func (s *OpenClawDiscoveryService) SearchOpenClawComponentsWithFilters(query string, limit int, filters map[string]string) ([]OpenClawDiscoveryResult, error) {
	// Build query string with filters
	queryParams := url.Values{}
	queryParams.Set("q", query)
	queryParams.Set("limit", strconv.Itoa(limit))

	for key, value := range filters {
		queryParams.Set(key, value)
	}

	apiPath := "/api/search?" + queryParams.Encode()

	req, err := s.createRegistryRequest("GET", apiPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create filtered search request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	req = req.WithContext(ctx)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("filtered search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("filtered search failed: status %d - %s", resp.StatusCode, string(body))
	}

	var registryResults []OpenClawRegistryResult
	err = json.NewDecoder(resp.Body).Decode(&registryResults)
	if err != nil {
		return nil, fmt.Errorf("failed to parse filtered search results: %w", err)
	}

	var discoveries []OpenClawDiscoveryResult
	for _, result := range registryResults {
		componentCount := len(result.Components)
		if componentCount == 0 && result.Metadata != nil {
			if count, ok := result.Metadata["component_count"].(float64); ok {
				componentCount = int(count)
			}
		}

		discoveries = append(discoveries, OpenClawDiscoveryResult{
			Source:        "registry",
			Name:          result.Name,
			Description:   result.Description,
			URL:           result.URL,
			Version:       result.Version,
			Downloads:     result.Downloads,
			ComponentCount: componentCount,
			UpdatedAt:     result.UpdatedAt,
			Metadata:      result.Metadata,
		})
	}

	return discoveries, nil
}

// CacheDiscoveryResults caches discovery results to avoid repeated searches.
func (s *OpenClawDiscoveryService) CacheDiscoveryResults(query string, results []OpenClawDiscoveryResult) error {
	// Ensure cache directory exists
	err := os.MkdirAll(s.CacheDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create cache file path
	cacheFile := filepath.Join(s.CacheDir, fmt.Sprintf("discovery-%s.json", sanitizeFilename(query)))

	// Serialize results
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize discovery results: %w", err)
	}

	// Write to cache
	err = os.WriteFile(cacheFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// LoadCachedDiscoveryResults loads cached discovery results.
func (s *OpenClawDiscoveryService) LoadCachedDiscoveryResults(query string) ([]OpenClawDiscoveryResult, error) {
	// Create cache file path
	cacheFile := filepath.Join(s.CacheDir, fmt.Sprintf("discovery-%s.json", sanitizeFilename(query)))

	// Check if cache file exists
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return nil, nil // No cached results
	}

	// Read cache file
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	// Deserialize results
	var results []OpenClawDiscoveryResult
	err = json.Unmarshal(data, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize cache data: %w", err)
	}

	return results, nil
}

// sanitizeFilename creates a safe filename from a query string.
func sanitizeFilename(query string) string {
	safe := strings.NewReplacer(
		" ", "-",
		"/", "-",
		`\`, "-",
		":", "-",
		"*", "-",
		"?", "-",
		"<", "-",
		">", "-",
		"|", "-",
		"\"", "-",
	).Replace(query)
	
	// Limit length
	if len(safe) > 100 {
		safe = safe[:100]
	}
	
	return safe
}

// findOpenClawManifests is a helper function to find OpenClaw manifest files.
// This is duplicated from the adapter to avoid circular imports.
func findOpenClawManifests(dir string) []string {
	var manifestPaths []string
	
	manifestNames := []string{
		"manifest.json",
		"manifest.yaml",
		"manifest.yml",
		"manifest.toml",
		"plugin.json",
		"plugin.yaml",
		"plugin.yml",
		"plugin.toml",
		"openclaw.json",
		"openclaw.yaml",
		"openclaw.yml",
		"openclaw.toml",
	}
	
	for _, name := range manifestNames {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			manifestPaths = append(manifestPaths, path)
		}
	}
	
	return manifestPaths
}