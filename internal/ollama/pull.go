package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// PullProgress represents progress during model download.
type PullProgress struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// ProgressFunc is called with progress updates during pull.
type ProgressFunc func(progress PullProgress)

// PullModel downloads a model from the Ollama registry.
// The progress callback is called for each status update.
func (c *Client) PullModel(ctx context.Context, name string, progress ProgressFunc) error {
	body := map[string]string{"name": name}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/pull", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Use a client without timeout for long downloads
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("pull model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("pull model: %s", errResp.Error)
		}
		return fmt.Errorf("pull model: status %d", resp.StatusCode)
	}

	// Stream progress updates
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var p PullProgress
		if err := json.Unmarshal(line, &p); err != nil {
			continue // Skip malformed lines
		}

		if progress != nil {
			progress(p)
		}

		// Check for completion
		if p.Status == "success" {
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read progress: %w", err)
	}

	return nil
}

// FormatProgress returns a human-readable progress string.
func FormatProgress(p PullProgress) string {
	if p.Total > 0 {
		percent := float64(p.Completed) / float64(p.Total) * 100
		return fmt.Sprintf("%s: %.1f%% (%s / %s)",
			p.Status,
			percent,
			formatBytes(p.Completed),
			formatBytes(p.Total))
	}
	return p.Status
}

// formatBytes converts bytes to a human-readable string.
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
