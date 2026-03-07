package build

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"charm.land/catwalk/pkg/catwalk"
	"charm.land/fantasy"
	"charm.land/fantasy/providers/openai"
	"charm.land/fantasy/providers/anthropic"
	"charm.land/fantasy/providers/google"
	"charm.land/fantasy/providers/openaicompat"
	"charm.land/fantasy/providers/openrouter"
)

// JudgeResult represents the result of evaluating an output
type JudgeResult struct {
	Score     float64 // 0-10 score
	Reasoning string  // Explanation from LLM
	Passed    bool    // Whether it meets threshold
}

// Judge uses LLM to evaluate agent outputs
type Judge struct {
	agent   fantasy.Agent
	model    string
	criteria string
}

// JudgeResponse is the expected response from the LLM
type JudgeResponse struct {
	Score     float64 `json:"score"`
	Reasoning string  `json:"reasoning"`
}

// NewJudge creates a new judge for evaluating outputs
func NewJudge(catwalkProvider catwalk.Provider, judgeModel, criteria string) (*Judge, error) {
	if judgeModel == "" {
		return nil, fmt.Errorf("judge_model not specified in [evals] config")
	}

	// Create fantasy provider from catwalk provider
	fantasyProvider, err := newFantasyProvider(catwalkProvider)
	if err != nil {
		return nil, fmt.Errorf("create fantasy provider: %w", err)
	}

	// Create language model
	ctx := context.Background()
	model, err := fantasyProvider.LanguageModel(ctx, judgeModel)
	if err != nil {
		return nil, fmt.Errorf("create language model: %w", err)
	}

	// Create judge agent
	agent := fantasy.NewAgent(model,
		fantasy.WithSystemPrompt("You are an evaluator that assesses the quality of AI outputs."),
	)

	return &Judge{
		agent:    agent,
		model:    judgeModel,
		criteria: criteria,
	}, nil
}

// newFantasyProvider creates a Fantasy provider from Catwalk configuration.
// This is similar to the function in internal/run/fantasy_provider.go
func newFantasyProvider(p catwalk.Provider) (fantasy.Provider, error) {
	apiKey := getProviderAPIKey(p)

	switch p.Type {
	case catwalk.TypeOpenAI:
		opts := []openai.Option{
			openai.WithAPIKey(apiKey),
			openai.WithUseResponsesAPI(),
		}
		if p.APIEndpoint != "" {
			opts = append(opts, openai.WithBaseURL(p.APIEndpoint))
		}
		if len(p.DefaultHeaders) > 0 {
			opts = append(opts, openai.WithHeaders(p.DefaultHeaders))
		}
		return openai.New(opts...)

	case catwalk.TypeOpenAICompat:
		opts := []openaicompat.Option{
			openaicompat.WithAPIKey(apiKey),
		}
		if p.APIEndpoint != "" {
			opts = append(opts, openaicompat.WithBaseURL(p.APIEndpoint))
		}
		if len(p.DefaultHeaders) > 0 {
			opts = append(opts, openaicompat.WithHeaders(p.DefaultHeaders))
		}
		return openaicompat.New(opts...)

	case catwalk.TypeAnthropic:
		opts := []anthropic.Option{anthropic.WithAPIKey(apiKey)}
		if p.APIEndpoint != "" {
			opts = append(opts, anthropic.WithBaseURL(p.APIEndpoint))
		}
		return anthropic.New(opts...)

	case catwalk.TypeGoogle:
		opts := []google.Option{google.WithGeminiAPIKey(apiKey)}
		return google.New(opts...)

	case catwalk.TypeOpenRouter:
		opts := []openrouter.Option{openrouter.WithAPIKey(apiKey)}
		return openrouter.New(opts...)

	default:
		// For unknown types, try OpenAI-compatible
		if p.APIEndpoint != "" {
			opts := []openaicompat.Option{
				openaicompat.WithAPIKey(apiKey),
				openaicompat.WithBaseURL(p.APIEndpoint),
			}
			return openaicompat.New(opts...)
		}
		return nil, fmt.Errorf("unsupported provider type: %s", p.Type)
	}
}

// getProviderAPIKey retrieves the API key from config or environment.
func getProviderAPIKey(p catwalk.Provider) string {
	if p.APIKey != "" {
		return p.APIKey
	}
	envKey := strings.ToUpper(string(p.ID)) + "_API_KEY"
	return os.Getenv(envKey)
}

// Compare evaluates an actual output against expected output
func (j *Judge) Compare(ctx context.Context, input, expected, actual map[string]any, criteriaOverride string) (JudgeResult, error) {
	// Use override criteria if provided
	criteria := j.criteria
	if criteriaOverride != "" {
		criteria = criteriaOverride
	}

	// Marshal inputs to JSON
	inputJSON, _ := json.MarshalIndent(input, "", "  ")
	expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
	actualJSON, _ := json.MarshalIndent(actual, "", "  ")

	// Build prompt
	systemPrompt := "You are an evaluator. Compare the actual output to the expected output based on the following criteria: " + criteria + ".\n\nRate the accuracy on a scale of 0 to 10, where:\n- 10: Perfect match, completely correct\n- 7-9: Minor differences, functionally correct\n- 4-6: Partially correct, significant issues\n- 1-3: Major errors, mostly incorrect\n- 0: Completely incorrect or failed\n\nProvide your reasoning and score in JSON format: {\"score\": <0-10>, \"reasoning\": \"<explanation>\"}"

	userPrompt := fmt.Sprintf("Input:\n%s\n\nExpected output:\n%s\n\nActual output:\n%s\n\nEvaluate the accuracy and provide your score and reasoning in JSON format.", string(inputJSON), string(expectedJSON), string(actualJSON))

	// Get LLM response - build prompt with system and user messages
	var content strings.Builder
	promptText := systemPrompt + "\n\n" + userPrompt
	_, callErr := j.agent.Stream(ctx, fantasy.AgentStreamCall{
		Prompt: promptText,
		OnTextDelta: func(id, text string) error { content.WriteString(text); return nil },
	})
	if callErr != nil {
		return JudgeResult{}, fmt.Errorf("get LLM response: %w", callErr)
	}

	resp := content.String()

	// Parse response
	var judgeResp JudgeResponse
	if err := json.Unmarshal([]byte(resp), &judgeResp); err != nil {
		return JudgeResult{}, fmt.Errorf("parse judge response: %w (response: %s)", err, resp)
	}

	// Default threshold is 7/10
	return JudgeResult{
		Score:     judgeResp.Score,
		Reasoning: judgeResp.Reasoning,
		Passed:    judgeResp.Score >= 7.0,
	}, nil
}

// Score is a convenience method that returns just the score
func (j *Judge) Score(ctx context.Context, input, expected, actual map[string]any, criteria string) (float64, string, error) {
	result, err := j.Compare(ctx, input, expected, actual, criteria)
	if err != nil {
		return 0, "", err
	}
	return result.Score, result.Reasoning, nil
}
