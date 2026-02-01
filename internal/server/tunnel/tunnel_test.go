package tunnel

import (
	"testing"
)

func TestCloudflaredProviderName(t *testing.T) {
	p := &CloudflaredProvider{}
	if p.Name() != "cloudflared" {
		t.Errorf("expected 'cloudflared', got %q", p.Name())
	}
}

func TestCloudflaredProviderAvailable(t *testing.T) {
	p := &CloudflaredProvider{}
	// Just verify it doesn't panic
	_ = p.Available()
}

func TestListProviders(t *testing.T) {
	providers := ListProviders()
	if len(providers) == 0 {
		t.Error("expected at least one provider")
	}

	found := false
	for _, p := range providers {
		if p.Name == "cloudflared" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected cloudflared in providers list")
	}
}

func TestDefaultProvider(t *testing.T) {
	// Just verify it doesn't panic
	_ = DefaultProvider()
}
