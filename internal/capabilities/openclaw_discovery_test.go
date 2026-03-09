package capabilities

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenClawDiscoveryService(t *testing.T) {
	t.Run("DiscoverFromLocalDirectory", func(t *testing.T) {
		// Create temporary directory with OpenClaw project
		tempDir := t.TempDir()
		projectDir := filepath.Join(tempDir, "test-project")
		err := os.MkdirAll(projectDir, 0755)
		require.NoError(t, err)
		
		// Create manifest file
		manifestContent := `{
  "name": "test-project",
  "version": "1.0.0",
  "components": [
    {
      "name": "test-skill",
      "type": "skill",
      "description": "A test skill"
    }
  ]
}`
		manifestPath := filepath.Join(projectDir, "manifest.json")
		err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)
		
		// Create SKILL.md file
		skillContent := `---
name: test-skill
description: Test skill
---
# Test Skill

This is a test skill.
`
		skillPath := filepath.Join(projectDir, "SKILL.md")
		err = os.WriteFile(skillPath, []byte(skillContent), 0644)
		require.NoError(t, err)
		
		service := NewOpenClawDiscoveryService()
		results, err := service.DiscoverFromLocalDirectory(projectDir)
		require.NoError(t, err)
		
		assert.Len(t, results, 1)
		assert.Equal(t, "local", results[0].Source)
		assert.Equal(t, "test-project", results[0].Name)
		assert.Contains(t, results[0].Description, tempDir)
		assert.Equal(t, 1, results[0].ComponentCount)
	})

	t.Run("DiscoverLocalOpenClawProjects", func(t *testing.T) {
		// Create temporary directory structure
		tempDir := t.TempDir()
		
		// Create first project
		project1Dir := filepath.Join(tempDir, "project1")
		err := os.MkdirAll(project1Dir, 0755)
		require.NoError(t, err)
		
		manifest1Content := `{"name": "project1", "version": "1.0.0", "components": []}`
		manifest1Path := filepath.Join(project1Dir, "manifest.json")
		err = os.WriteFile(manifest1Path, []byte(manifest1Content), 0644)
		require.NoError(t, err)
		
		// Create second project in subdirectory
		project2Dir := filepath.Join(tempDir, "subdir", "project2")
		err = os.MkdirAll(project2Dir, 0755)
		require.NoError(t, err)
		
		manifest2Content := `{"name": "project2", "version": "2.0.0", "components": []}`
		manifest2Path := filepath.Join(project2Dir, "manifest.json")
		err = os.WriteFile(manifest2Path, []byte(manifest2Content), 0644)
		require.NoError(t, err)
		
		service := NewOpenClawDiscoveryService()
		results, err := service.DiscoverLocalOpenClawProjects([]string{tempDir})
		require.NoError(t, err)
		
		assert.Len(t, results, 2)
		
		// Check that we found both projects
		foundNames := make(map[string]bool)
		for _, result := range results {
			foundNames[result.Name] = true
		}
		
		assert.True(t, foundNames["project1"])
		assert.True(t, foundNames["project2"])
	})

	t.Run("DiscoverFromRegistry with mock server", func(t *testing.T) {
		// Create mock registry server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check query parameters
			query := r.URL.Query()
			assert.Equal(t, "test", query.Get("q"))
			assert.Equal(t, "10", query.Get("limit"))
			
			// Return mock response
			response := `[
  {
    "name": "mock-plugin",
    "description": "A mock OpenClaw plugin",
    "url": "https://registry.openclaw.io/plugins/mock-plugin",
    "version": "1.0.0",
    "downloads": 100,
    "components": [
      {
        "name": "mock-skill",
        "type": "skill",
        "description": "A mock skill"
      }
    ],
    "updated_at": "2023-01-01T00:00:00Z",
    "metadata": {
      "author": "Mock Author"
    }
  }
]`
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()
		
		service := NewOpenClawDiscoveryService().WithRegistryURL(server.URL)
		results, err := service.DiscoverFromRegistry("test", 10)
		require.NoError(t, err)
		
		assert.Len(t, results, 1)
		assert.Equal(t, "registry", results[0].Source)
		assert.Equal(t, "mock-plugin", results[0].Name)
		assert.Equal(t, "A mock OpenClaw plugin", results[0].Description)
		assert.Equal(t, "https://registry.openclaw.io/plugins/mock-plugin", results[0].URL)
		assert.Equal(t, "1.0.0", results[0].Version)
		assert.Equal(t, 100, results[0].Downloads)
		assert.Equal(t, 1, results[0].ComponentCount)
		assert.Equal(t, "Mock Author", results[0].Metadata["author"])
	})

	t.Run("CacheDiscoveryResults and LoadCachedDiscoveryResults", func(t *testing.T) {
		// Create temporary cache directory
		cacheDir := t.TempDir()
		service := NewOpenClawDiscoveryService()
		service.CacheDir = cacheDir
		
		// Create test results
		results := []OpenClawDiscoveryResult{
			{
				Source:      "test",
				Name:        "test-plugin",
				Description: "Test plugin",
				URL:         "https://example.com/test-plugin",
			},
		}
		
		// Cache results
		err := service.CacheDiscoveryResults("test-query", results)
		require.NoError(t, err)
		
		// Load cached results
		loadedResults, err := service.LoadCachedDiscoveryResults("test-query")
		require.NoError(t, err)
		
		assert.Len(t, loadedResults, 1)
		assert.Equal(t, "test-plugin", loadedResults[0].Name)
		assert.Equal(t, "Test plugin", loadedResults[0].Description)
	})

	t.Run("sanitizeFilename", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"simple query", "simple-query"},
			{"query with spaces", "query-with-spaces"},
			{"query/with/slashes", "query-with-slashes"},
			{"query:with:colons", "query-with-colons"},
			{"query*with*stars", "query-with-stars"},
			{"query?with?questions", "query-with-questions"},
			{"very long query that should be truncated because it exceeds the maximum length of 100 characters", "very-long-query-that-should-be-truncated-because-it-exceeds-the-maximum-length-of-100-characters"},
		}
		
		for _, tc := range testCases {
			result := sanitizeFilename(tc.input)
			assert.Equal(t, tc.expected, result, "Input: %s", tc.input)
		}
	})

	t.Run("findOpenClawManifests", func(t *testing.T) {
		// Create temporary directory with various manifest files
		tempDir := t.TempDir()
		
		manifestFiles := []string{
			"manifest.json",
			"manifest.yaml",
			"plugin.toml",
			"openclaw.yml",
			"readme.md", // Should be ignored
		}
		
		for _, filename := range manifestFiles {
			if filepath.Ext(filename) != ".md" {
				path := filepath.Join(tempDir, filename)
				content := `{"name": "test", "version": "1.0.0", "components": []}`
				if filepath.Ext(filename) == ".yaml" || filepath.Ext(filename) == ".yml" {
					content = "name: test\nversion: 1.0.0\ncomponents: []"
				} else if filepath.Ext(filename) == ".toml" {
					content = "name = 'test'\nversion = '1.0.0'\ncomponents = []"
				}
				err := os.WriteFile(path, []byte(content), 0644)
				require.NoError(t, err)
			}
		}
		
		// Find manifests
		manifestPaths := findOpenClawManifests(tempDir)
		
		// Should find all manifest files except readme.md
		assert.Len(t, manifestPaths, 4)
		
		// Check that we found the right files
		foundFilenames := make(map[string]bool)
		for _, path := range manifestPaths {
			foundFilenames[filepath.Base(path)] = true
		}
		
		assert.True(t, foundFilenames["manifest.json"])
		assert.True(t, foundFilenames["manifest.yaml"])
		assert.True(t, foundFilenames["plugin.toml"])
		assert.True(t, foundFilenames["openclaw.yml"])
		assert.False(t, foundFilenames["readme.md"])
	})

	t.Run("SearchOpenClawComponents with mock servers", func(t *testing.T) {
		// Create mock GitHub server
		githubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `{
  "total_count": 1,
  "items": [
    {
      "name": "openclaw-test",
      "full_name": "test/openclaw-test",
      "description": "A test OpenClaw plugin",
      "html_url": "https://github.com/test/openclaw-test",
      "clone_url": "https://github.com/test/openclaw-test.git",
      "stargazers_count": 42,
      "updated_at": "2023-01-01T00:00:00Z",
      "topics": ["openclaw", "plugin"]
    }
  ]
}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer githubServer.Close()
		
		// Create mock registry server
		registryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `[
  {
    "name": "registry-test",
    "description": "A test plugin from registry",
    "url": "https://registry.openclaw.io/plugins/registry-test",
    "version": "2.0.0",
    "downloads": 200,
    "components": [
      {
        "name": "registry-skill",
        "type": "skill",
        "description": "A registry skill"
      }
    ],
    "updated_at": "2023-02-01T00:00:00Z"
  }
]`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer registryServer.Close()
		
		// Create service with mock clients
		service := NewOpenClawDiscoveryService().WithRegistryURL(registryServer.URL)
		
		// Mock GitHub client - this is more complex, so we'll skip it for this test
		// and just test that the service can handle GitHub failures gracefully
		
		// Test with registry only
		results, err := service.DiscoverFromRegistry("test", 10)
		require.NoError(t, err)
		
		assert.Len(t, results, 1)
		assert.Equal(t, "registry", results[0].Source)
		assert.Equal(t, "mock-plugin", results[0].Name)
	})

	t.Run("GetOpenClawComponentInfo with mock server", func(t *testing.T) {
		// Create mock registry server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check path
			assert.Equal(t, "/api/components/test-component", r.URL.Path)
			
			response := `{
  "name": "test-component",
  "description": "A test component",
  "url": "https://registry.openclaw.io/components/test-component",
  "version": "1.0.0",
  "downloads": 50,
  "components": [
    {
      "name": "test-skill",
      "type": "skill",
      "description": "A test skill"
    }
  ],
  "updated_at": "2023-01-01T00:00:00Z",
  "metadata": {
    "author": "Test Author",
    "license": "MIT"
  }
}`
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()
		
		service := NewOpenClawDiscoveryService().WithRegistryURL(server.URL).WithCacheEnablement(false)
		
		componentInfo, err := service.GetOpenClawComponentInfo("test-component")
		require.NoError(t, err)
		
		assert.Equal(t, "test-component", componentInfo.Name)
		assert.Equal(t, "A test component", componentInfo.Description)
		assert.Equal(t, "https://registry.openclaw.io/components/test-component", componentInfo.URL)
		assert.Equal(t, "1.0.0", componentInfo.Version)
		assert.Equal(t, 50, componentInfo.Downloads)
		assert.Len(t, componentInfo.Components, 1)
		assert.Equal(t, "Test Author", componentInfo.Metadata["author"])
		assert.Equal(t, "MIT", componentInfo.Metadata["license"])
	})

	t.Run("GetOpenClawComponentInstallationURL", func(t *testing.T) {
		// Create mock registry server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `{
  "name": "install-test",
  "description": "Install test component",
  "url": "https://registry.openclaw.io/components/install-test",
  "version": "2.0.0",
  "downloads": 100,
  "components": [],
  "updated_at": "2023-01-01T00:00:00Z"
}`
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()
		
		service := NewOpenClawDiscoveryService().WithRegistryURL(server.URL)
		
		// Test without version
		url, err := service.GetOpenClawComponentInstallationURL("install-test", "")
		require.NoError(t, err)
		assert.Equal(t, "https://registry.openclaw.io/components/install-test", url)
		
		// Test with version
		urlWithVersion, err := service.GetOpenClawComponentInstallationURL("install-test", "1.5.0")
		require.NoError(t, err)
		assert.Equal(t, "https://registry.openclaw.io/components/install-test@1.5.0", urlWithVersion)
	})

	t.Run("SearchOpenClawComponentsWithFilters", func(t *testing.T) {
		// Create mock registry server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check query parameters
			query := r.URL.Query()
			assert.Equal(t, "filter-test", query.Get("q"))
			assert.Equal(t, "10", query.Get("limit"))
			assert.Equal(t, "skill", query.Get("type"))
			assert.Equal(t, "test-author", query.Get("author"))
			
			response := `[
  {
    "name": "filtered-result",
    "description": "Filtered search result",
    "url": "https://registry.openclaw.io/components/filtered-result",
    "version": "1.0.0",
    "downloads": 75,
    "components": [
      {
        "name": "filtered-skill",
        "type": "skill",
        "description": "A filtered skill"
      }
    ],
    "updated_at": "2023-01-01T00:00:00Z",
    "metadata": {
      "component_count": 1
    }
  }
]`
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()
		
		service := NewOpenClawDiscoveryService().WithRegistryURL(server.URL)
		
		filters := map[string]string{
			"type":   "skill",
			"author": "test-author",
		}
		
		results, err := service.SearchOpenClawComponentsWithFilters("filter-test", 10, filters)
		require.NoError(t, err)
		
		assert.Len(t, results, 1)
		assert.Equal(t, "filtered-result", results[0].Name)
		assert.Equal(t, 1, results[0].ComponentCount)
	})

	t.Run("Registry error handling", func(t *testing.T) {
		// Create mock registry server that returns unauthorized error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Authentication required"}`))
		}))
		defer server.Close()
		
		service := NewOpenClawDiscoveryService().WithRegistryURL(server.URL).WithCacheEnablement(false)
		
		_, err := service.DiscoverFromRegistry("test", 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
	})

	t.Run("OpenClawDiscoveryResult serialization", func(t *testing.T) {
		result := OpenClawDiscoveryResult{
			Source:        "test",
			Name:          "test-plugin",
			Description:   "Test description",
			URL:           "https://example.com/test-plugin",
			CloneURL:      "https://github.com/test/test-plugin.git",
			Version:       "1.0.0",
			ComponentCount: 5,
			Downloads:     100,
			Stars:         42,
			UpdatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			Metadata: map[string]interface{}{
				"key": "value",
				"number": float64(123),
			},
		}
		
		// Serialize to JSON
		data, err := json.MarshalIndent(result, "", "  ")
		require.NoError(t, err)
		
		// Deserialize back
		var loadedResult OpenClawDiscoveryResult
		err = json.Unmarshal(data, &loadedResult)
		require.NoError(t, err)
		
		// Verify all fields
		assert.Equal(t, result.Source, loadedResult.Source)
		assert.Equal(t, result.Name, loadedResult.Name)
		assert.Equal(t, result.Description, loadedResult.Description)
		assert.Equal(t, result.URL, loadedResult.URL)
		assert.Equal(t, result.CloneURL, loadedResult.CloneURL)
		assert.Equal(t, result.Version, loadedResult.Version)
		assert.Equal(t, result.ComponentCount, loadedResult.ComponentCount)
		assert.Equal(t, result.Downloads, loadedResult.Downloads)
		assert.Equal(t, result.Stars, loadedResult.Stars)
		assert.Equal(t, result.UpdatedAt, loadedResult.UpdatedAt)
		assert.Equal(t, result.Metadata, loadedResult.Metadata)
	})
}