// Package tunnel provides local tunneling for remote access to ayo.
package tunnel

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Tunnel represents an active tunnel connection.
type Tunnel interface {
	// URL returns the public URL of the tunnel.
	URL() string

	// Stop closes the tunnel.
	Stop() error

	// Wait blocks until the tunnel exits.
	Wait() error
}

// Provider creates tunnels.
type Provider interface {
	// Name returns the provider name (e.g., "cloudflared", "ngrok").
	Name() string

	// Available returns true if the provider is installed and usable.
	Available() bool

	// Start creates a new tunnel to the given local address.
	Start(ctx context.Context, localAddr string) (Tunnel, error)
}

// CloudflaredProvider provides tunnels via Cloudflare's cloudflared.
type CloudflaredProvider struct{}

// cloudflaredTunnel is an active cloudflared tunnel.
type cloudflaredTunnel struct {
	cmd     *exec.Cmd
	url     string
	cancel  context.CancelFunc
	done    chan struct{}
	err     error
	mu      sync.Mutex
}

// Name returns "cloudflared".
func (p *CloudflaredProvider) Name() string {
	return "cloudflared"
}

// Available returns true if cloudflared is installed.
func (p *CloudflaredProvider) Available() bool {
	_, err := exec.LookPath("cloudflared")
	return err == nil
}

// Start creates a new tunnel.
func (p *CloudflaredProvider) Start(ctx context.Context, localAddr string) (Tunnel, error) {
	if !p.Available() {
		return nil, errors.New("cloudflared not found in PATH")
	}

	// Ensure localAddr has the http scheme
	if !strings.HasPrefix(localAddr, "http://") && !strings.HasPrefix(localAddr, "https://") {
		localAddr = "http://" + localAddr
	}

	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "--url", localAddr)

	// Capture stderr for URL
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	t := &cloudflaredTunnel{
		cmd:    cmd,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start cloudflared: %w", err)
	}

	// Parse URL from stderr in background
	urlCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(stderr)
		urlPattern := regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)

		for scanner.Scan() {
			line := scanner.Text()
			if matches := urlPattern.FindString(line); matches != "" {
				urlCh <- matches
				return
			}
		}
		errCh <- errors.New("failed to find tunnel URL in cloudflared output")
	}()

	// Wait for URL with timeout
	select {
	case url := <-urlCh:
		t.url = url
	case err := <-errCh:
		cmd.Process.Kill()
		cancel()
		return nil, err
	case <-time.After(30 * time.Second):
		cmd.Process.Kill()
		cancel()
		return nil, errors.New("timeout waiting for cloudflared tunnel URL")
	case <-ctx.Done():
		cmd.Process.Kill()
		cancel()
		return nil, ctx.Err()
	}

	// Monitor process in background
	go func() {
		t.err = cmd.Wait()
		close(t.done)
	}()

	return t, nil
}

// URL returns the public tunnel URL.
func (t *cloudflaredTunnel) URL() string {
	return t.url
}

// Stop closes the tunnel.
func (t *cloudflaredTunnel) Stop() error {
	t.cancel()
	return t.Wait()
}

// Wait blocks until the tunnel exits.
func (t *cloudflaredTunnel) Wait() error {
	<-t.done
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.err
}

// DefaultProvider returns the first available tunnel provider.
// Currently only supports cloudflared.
func DefaultProvider() Provider {
	cf := &CloudflaredProvider{}
	if cf.Available() {
		return cf
	}
	return nil
}

// ListProviders returns all known tunnel providers with availability status.
func ListProviders() []struct {
	Name      string
	Available bool
} {
	return []struct {
		Name      string
		Available bool
	}{
		{"cloudflared", (&CloudflaredProvider{}).Available()},
	}
}
