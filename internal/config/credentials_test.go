package config

import (
	"os"
	"testing"
)

func TestDetectProviders(t *testing.T) {
	// Clear any cached credentials
	ClearCredentialCache()

	// Clear test env vars
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

	// Test with no env vars set
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")

	providers := DetectProviders()
	if len(providers) == 0 {
		t.Fatal("expected at least one provider")
	}

	// All should have HasKey = false
	for _, p := range providers {
		if p.HasKey {
			t.Errorf("provider %s should not have key when env is empty", p.ID)
		}
	}

	// Set one env var
	os.Setenv("ANTHROPIC_API_KEY", "test-key")

	providers = DetectProviders()
	var found bool
	for _, p := range providers {
		if p.ID == "anthropic" {
			if !p.HasKey {
				t.Error("anthropic should have key after setting env")
			}
			found = true
		}
	}
	if !found {
		t.Error("anthropic provider not found")
	}
}

func TestHasAnyProvider(t *testing.T) {
	ClearCredentialCache()

	// Clear test env vars
	originalKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENAI_API_KEY", originalKey)
		} else {
			os.Unsetenv("OPENAI_API_KEY")
		}
	}()

	// Save and clear all known env vars
	savedEnvs := make(map[string]string)
	for _, p := range knownProviders {
		savedEnvs[p.EnvVar] = os.Getenv(p.EnvVar)
		os.Unsetenv(p.EnvVar)
	}
	defer func() {
		for envVar, value := range savedEnvs {
			if value != "" {
				os.Setenv(envVar, value)
			}
		}
	}()

	if HasAnyProvider() {
		t.Error("should not have any provider when all env vars are empty")
	}

	os.Setenv("OPENAI_API_KEY", "test-key")

	if !HasAnyProvider() {
		t.Error("should have provider after setting OPENAI_API_KEY")
	}
}

func TestGetProvidersWithCredentials(t *testing.T) {
	ClearCredentialCache()

	// Save and clear all known env vars
	savedEnvs := make(map[string]string)
	for _, p := range knownProviders {
		savedEnvs[p.EnvVar] = os.Getenv(p.EnvVar)
		os.Unsetenv(p.EnvVar)
	}
	defer func() {
		for envVar, value := range savedEnvs {
			if value != "" {
				os.Setenv(envVar, value)
			}
		}
	}()

	providers := GetProvidersWithCredentials()
	if len(providers) != 0 {
		t.Errorf("expected 0 providers, got %d", len(providers))
	}

	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	os.Setenv("OPENAI_API_KEY", "test-key-2")

	providers = GetProvidersWithCredentials()
	if len(providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(providers))
	}
}

func TestCredentialStorage(t *testing.T) {
	ClearCredentialCache()

	// Skip storage tests in dev mode since we can't control the path
	path := credentialsPath()
	if _, err := os.Stat(path); err == nil {
		// Credentials file already exists from dev environment
		// Test the loading behavior instead
		creds, err := LoadStoredCredentials()
		if err != nil {
			t.Fatalf("LoadStoredCredentials failed: %v", err)
		}
		if creds.Version != 1 {
			t.Errorf("expected version 1, got %d", creds.Version)
		}
		t.Log("skipping full storage test - credentials file exists in dev environment")
		return
	}

	// Load empty credentials
	creds, err := LoadStoredCredentials()
	if err != nil {
		t.Fatalf("LoadStoredCredentials failed: %v", err)
	}

	if creds.Version != 1 {
		t.Errorf("expected version 1, got %d", creds.Version)
	}

	if len(creds.Credentials) != 0 {
		t.Error("expected empty credentials")
	}

	// Store a credential
	if err := StoreCredential("anthropic", "test-api-key"); err != nil {
		t.Fatalf("StoreCredential failed: %v", err)
	}

	// Clear cache and reload
	ClearCredentialCache()

	creds, err = LoadStoredCredentials()
	if err != nil {
		t.Fatalf("LoadStoredCredentials failed: %v", err)
	}

	if _, ok := creds.Credentials["anthropic"]; !ok {
		t.Error("expected anthropic credential after store")
	}

	if creds.Credentials["anthropic"].APIKey != "test-api-key" {
		t.Error("stored key doesn't match")
	}
}

func TestInjectCredentials(t *testing.T) {
	ClearCredentialCache()

	// This test verifies InjectCredentials reads from stored credentials
	// and sets environment variables

	// Clear test env var
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	// Store a credential (this writes to the dev config dir)
	if err := StoreCredential("anthropic", "injected-key"); err != nil {
		t.Fatalf("StoreCredential failed: %v", err)
	}
	// Clean up after test
	defer func() {
		creds, _ := LoadStoredCredentials()
		delete(creds.Credentials, "anthropic")
		SaveCredentials(creds)
	}()

	// Clear cache
	ClearCredentialCache()

	// Inject credentials
	if err := InjectCredentials(); err != nil {
		t.Fatalf("InjectCredentials failed: %v", err)
	}

	// Check env var is set
	if os.Getenv("ANTHROPIC_API_KEY") != "injected-key" {
		t.Error("ANTHROPIC_API_KEY should be set after InjectCredentials")
	}
}

func TestInjectCredentials_NoOverwrite(t *testing.T) {
	ClearCredentialCache()

	// Set env var first
	originalKey := os.Getenv("OPENAI_API_KEY")
	os.Setenv("OPENAI_API_KEY", "existing-key")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENAI_API_KEY", originalKey)
		} else {
			os.Unsetenv("OPENAI_API_KEY")
		}
	}()

	// Store a different credential
	if err := StoreCredential("openai", "stored-key"); err != nil {
		t.Fatalf("StoreCredential failed: %v", err)
	}
	// Clean up after test
	defer func() {
		creds, _ := LoadStoredCredentials()
		delete(creds.Credentials, "openai")
		SaveCredentials(creds)
	}()

	ClearCredentialCache()

	// Inject credentials
	if err := InjectCredentials(); err != nil {
		t.Fatalf("InjectCredentials failed: %v", err)
	}

	// Env var should NOT be overwritten
	if os.Getenv("OPENAI_API_KEY") != "existing-key" {
		t.Error("OPENAI_API_KEY should not be overwritten by stored credential")
	}
}

func TestCredentialFilePermissions(t *testing.T) {
	ClearCredentialCache()

	// Store a credential and check its permissions
	if err := StoreCredential("test-provider-perm", "test-key"); err != nil {
		t.Fatalf("StoreCredential failed: %v", err)
	}
	// Clean up after test
	defer func() {
		creds, _ := LoadStoredCredentials()
		delete(creds.Credentials, "test-provider-perm")
		SaveCredentials(creds)
	}()

	// Check file permissions - use the actual path the module uses
	path := credentialsPath()
	info, err := os.Stat(path)
	if err != nil {
		t.Skipf("skipping permission test - file not at expected path: %v", err)
	}

	// Check permissions are 0600 (owner read/write only)
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected permissions 0600, got %o", perm)
	}
}
