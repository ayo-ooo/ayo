package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the request body for /api/chat.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Format   string    `json:"format,omitempty"` // "json" for JSON mode
	Options  *Options  `json:"options,omitempty"`
}

// Options are model-specific options.
type Options struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
}

// ChatResponse is the response from /api/chat.
type ChatResponse struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`

	// Final response includes these
	TotalDuration      int64 `json:"total_duration,omitempty"`
	LoadDuration       int64 `json:"load_duration,omitempty"`
	PromptEvalCount    int   `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64 `json:"prompt_eval_duration,omitempty"`
	EvalCount          int   `json:"eval_count,omitempty"`
	EvalDuration       int64 `json:"eval_duration,omitempty"`
}

// Chat sends a chat request and returns the complete response.
func (c *Client) Chat(ctx context.Context, model string, messages []Message, opts *Options) (*ChatResponse, error) {
	reqBody := ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Options:  opts,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chat request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chat request: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// ChatJSON sends a chat request expecting a JSON response.
func (c *Client) ChatJSON(ctx context.Context, model string, messages []Message, opts *Options) (json.RawMessage, error) {
	reqBody := ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Format:   "json",
		Options:  opts,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chat request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chat request: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return json.RawMessage(result.Message.Content), nil
}
