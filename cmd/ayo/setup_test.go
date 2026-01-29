package main

import (
	"os"
	"testing"

	"github.com/alexcabrera/ayo/internal/config"
)

func TestProviderDetectionEnvVars(t *testing.T) {
	// Save original env
	originalAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	originalOpenAIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		if originalAnthropicKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalAnthropicKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
		if originalOpenAIKey != "" {
			os.Setenv("OPENAI_API_KEY", originalOpenAIKey)
		} else {
			os.Unsetenv("OPENAI_API_KEY")
		}
	}()

	// Test with no keys
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")

	providers := config.DetectProviders()
	for _, p := range providers {
		if p.ID == "anthropic" && p.HasKey {
			t.Error("anthropic should not have key")
		}
		if p.ID == "openai" && p.HasKey {
			t.Error("openai should not have key")
		}
	}

	if config.HasAnyProvider() {
		// This might be true if other env vars are set, which is fine
		// Only fail if specifically the ones we unset are detected
		for _, p := range config.GetProvidersWithCredentials() {
			if p.ID == "anthropic" || p.ID == "openai" {
				t.Errorf("unexpected provider with credentials: %s", p.ID)
			}
		}
	}

	// Test with anthropic key
	os.Setenv("ANTHROPIC_API_KEY", "test-key")

	providers = config.DetectProviders()
	foundAnthropicWithKey := false
	for _, p := range providers {
		if p.ID == "anthropic" && p.HasKey {
			foundAnthropicWithKey = true
		}
	}
	if !foundAnthropicWithKey {
		t.Error("anthropic should have key after setting env var")
	}

	if !config.HasAnyProvider() {
		t.Error("HasAnyProvider should return true when ANTHROPIC_API_KEY is set")
	}

	configuredProviders := config.GetProvidersWithCredentials()
	foundAnthropicConfigured := false
	for _, p := range configuredProviders {
		if p.ID == "anthropic" {
			foundAnthropicConfigured = true
		}
	}
	if !foundAnthropicConfigured {
		t.Error("anthropic should be in configured providers list")
	}
}

func TestCredentialStoreRoundTrip(t *testing.T) {
	// Create temp directory for credentials
	tmpDir := t.TempDir()
	originalConfigDir := os.Getenv("AYO_CONFIG_DIR")
	os.Setenv("AYO_CONFIG_DIR", tmpDir)
	defer func() {
		if originalConfigDir != "" {
			os.Setenv("AYO_CONFIG_DIR", originalConfigDir)
		} else {
			os.Unsetenv("AYO_CONFIG_DIR")
		}
		config.ClearCredentialCache()
	}()

	// Clear cache to ensure fresh state
	config.ClearCredentialCache()

	// Store a credential
	err := config.StoreCredential("anthropic", "test-api-key-12345")
	if err != nil {
		t.Fatalf("StoreCredential failed: %v", err)
	}

	// Clear cache and reload
	config.ClearCredentialCache()
	creds, err := config.LoadStoredCredentials()
	if err != nil {
		t.Fatalf("LoadStoredCredentials failed: %v", err)
	}

	anthCred, ok := creds.Credentials["anthropic"]
	if !ok {
		t.Fatal("anthropic credential not found after store")
	}

	if anthCred.APIKey != "test-api-key-12345" {
		t.Errorf("stored API key = %q, want %q", anthCred.APIKey, "test-api-key-12345")
	}

	if anthCred.AddedAt.IsZero() {
		t.Error("AddedAt should be set")
	}
}

func TestInjectCredentialsDoesNotOverwrite(t *testing.T) {
	// Create temp directory for credentials
	tmpDir := t.TempDir()
	originalConfigDir := os.Getenv("AYO_CONFIG_DIR")
	os.Setenv("AYO_CONFIG_DIR", tmpDir)
	originalAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalConfigDir != "" {
			os.Setenv("AYO_CONFIG_DIR", originalConfigDir)
		} else {
			os.Unsetenv("AYO_CONFIG_DIR")
		}
		if originalAnthropicKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalAnthropicKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
		config.ClearCredentialCache()
	}()

	config.ClearCredentialCache()

	// Store a credential
	err := config.StoreCredential("anthropic", "stored-key")
	if err != nil {
		t.Fatalf("StoreCredential failed: %v", err)
	}

	// Set env var to a different value
	os.Setenv("ANTHROPIC_API_KEY", "env-key")

	// Clear cache and inject
	config.ClearCredentialCache()
	err = config.InjectCredentials()
	if err != nil {
		t.Fatalf("InjectCredentials failed: %v", err)
	}

	// Env var should still be "env-key" (not overwritten)
	if os.Getenv("ANTHROPIC_API_KEY") != "env-key" {
		t.Errorf("InjectCredentials should not overwrite existing env var, got %q", os.Getenv("ANTHROPIC_API_KEY"))
	}
}

func TestInjectCredentialsSetsEnvVar(t *testing.T) {
	// Create temp directory for credentials
	tmpDir := t.TempDir()
	originalConfigDir := os.Getenv("AYO_CONFIG_DIR")
	os.Setenv("AYO_CONFIG_DIR", tmpDir)
	originalAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalConfigDir != "" {
			os.Setenv("AYO_CONFIG_DIR", originalConfigDir)
		} else {
			os.Unsetenv("AYO_CONFIG_DIR")
		}
		if originalAnthropicKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalAnthropicKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
		config.ClearCredentialCache()
	}()

	config.ClearCredentialCache()

	// Store a credential
	err := config.StoreCredential("anthropic", "stored-key-inject")
	if err != nil {
		t.Fatalf("StoreCredential failed: %v", err)
	}

	// Clear the env var
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Clear cache and inject
	config.ClearCredentialCache()
	err = config.InjectCredentials()
	if err != nil {
		t.Fatalf("InjectCredentials failed: %v", err)
	}

	// Env var should now be set to stored value
	if os.Getenv("ANTHROPIC_API_KEY") != "stored-key-inject" {
		t.Errorf("InjectCredentials should set env var, got %q", os.Getenv("ANTHROPIC_API_KEY"))
	}
}
