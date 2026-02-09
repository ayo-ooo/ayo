---
id: ase-zd7g
status: closed
deps: [ase-qsc7]
links: []
created: 2026-02-09T03:13:00Z
type: task
priority: 2
assignee: Alex Cabrera
parent: ase-y48y
---
# Unit tests for capability inference

## Background

The capability inference system analyzes agent definitions using an LLM to extract structured capabilities. These unit tests verify the inference logic works correctly across various agent types.

## Why This Matters

Capability inference is core to @ayo's agent selection. Tests ensure:
- Diverse agent types produce meaningful capabilities
- Confidence scores are calibrated correctly
- Edge cases (empty prompts, skill-only agents) handled
- LLM prompt produces parseable output

## Implementation Details

### Test Structure

```
internal/
  capabilities/
    inference.go
    inference_test.go
    embeddings.go
    embeddings_test.go
    search.go
    search_test.go
    similarity.go
    similarity_test.go
```

### Mock LLM Provider

For unit tests, mock the LLM to return predictable responses:

```go
// internal/capabilities/mock_provider_test.go
type MockProvider struct {
    responses map[string]string
}

func (m *MockProvider) Complete(prompt string) (string, error) {
    // Return canned response based on prompt hash or keywords
    if strings.Contains(prompt, "code reviewer") {
        return `[{"name":"code-review","description":"Reviews source code","confidence":0.95}]`, nil
    }
    return "[]", nil
}
```

### Inference Test Cases

**inference_test.go:**

```go
func TestInferCapabilities_CodeReviewer(t *testing.T) {
    input := InferenceInput{
        SystemPrompt: "You are a code reviewer who focuses on code quality and best practices.",
    }
    
    result, err := inferrer.Infer(input)
    
    assert.NoError(t, err)
    assert.Len(t, result.Capabilities, >= 1)
    assert.Contains(t, capabilityNames(result), "code-review")
}

func TestInferCapabilities_WithSkills(t *testing.T) {
    input := InferenceInput{
        SystemPrompt: "You are a general assistant.",
        SkillNames:   []string{"git-operations", "file-search"},
        SkillContents: []string{
            "Git commands and repository management",
            "Search files by name and content",
        },
    }
    
    result, err := inferrer.Infer(input)
    
    // Should infer capabilities from skills
    assert.Contains(t, capabilityNames(result), "git-operations")
}

func TestInferCapabilities_EmptyPrompt(t *testing.T) {
    input := InferenceInput{
        SystemPrompt: "",
    }
    
    result, err := inferrer.Infer(input)
    
    assert.NoError(t, err)
    assert.Empty(t, result.Capabilities)
}

func TestInferCapabilities_SecurityAgent(t *testing.T) {
    input := InferenceInput{
        SystemPrompt: "You are a security auditor. You analyze code for vulnerabilities including SQL injection, XSS, and authentication bypasses.",
    }
    
    result, err := inferrer.Infer(input)
    
    assert.Contains(t, capabilityNames(result), "security-analysis")
    // Confidence should be high for primary purpose
    secCap := findCapability(result, "security-analysis")
    assert.GreaterOrEqual(t, secCap.Confidence, 0.8)
}

func TestInferCapabilities_MultiPurpose(t *testing.T) {
    input := InferenceInput{
        SystemPrompt: "You are a senior developer who reviews code, writes documentation, and mentors junior developers.",
    }
    
    result, err := inferrer.Infer(input)
    
    // Should have multiple capabilities
    assert.GreaterOrEqual(t, len(result.Capabilities), 2)
}

func TestInferCapabilities_ExplicitRestrictions(t *testing.T) {
    input := InferenceInput{
        SystemPrompt: "You are a documentation writer. You do NOT write code, only documentation.",
    }
    
    result, err := inferrer.Infer(input)
    
    // Should NOT have code-writing capability
    assert.NotContains(t, capabilityNames(result), "code-writing")
}

func TestInferCapabilities_InvalidLLMResponse(t *testing.T) {
    // Mock LLM returns non-JSON
    mockProvider.SetResponse("This is not JSON")
    
    _, err := inferrer.Infer(input)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "parse")
}

func TestInferCapabilities_CacheHit(t *testing.T) {
    // Same input should not call LLM again
    input := InferenceInput{SystemPrompt: "Test agent"}
    
    inferrer.Infer(input) // First call
    inferrer.Infer(input) // Second call
    
    assert.Equal(t, 1, mockProvider.CallCount())
}

func TestInferCapabilities_CacheInvalidation(t *testing.T) {
    // Different input should call LLM
    input1 := InferenceInput{SystemPrompt: "Agent A"}
    input2 := InferenceInput{SystemPrompt: "Agent B"}
    
    inferrer.Infer(input1)
    inferrer.Infer(input2)
    
    assert.Equal(t, 2, mockProvider.CallCount())
}
```

### Embeddings Test Cases

**embeddings_test.go:**

```go
func TestEmbedding_GeneratesVector(t *testing.T) {
    text := "code review and analysis"
    
    embedding, err := embedder.Embed(text)
    
    assert.NoError(t, err)
    assert.Len(t, embedding, 1536) // Typical embedding size
}

func TestEmbedding_SimilarTextsSimilarVectors(t *testing.T) {
    e1, _ := embedder.Embed("code review")
    e2, _ := embedder.Embed("reviewing code")
    e3, _ := embedder.Embed("making pizza")
    
    sim12 := cosineSimilarity(e1, e2)
    sim13 := cosineSimilarity(e1, e3)
    
    assert.Greater(t, sim12, sim13)
}
```

### Search Test Cases

**search_test.go:**

```go
func TestSearch_FindsRelevant(t *testing.T) {
    // Seed with known capabilities
    setupTestCapabilities(t, db)
    
    results, err := searcher.Search("review code for bugs", 3)
    
    assert.NoError(t, err)
    assert.Equal(t, "code-reviewer", results[0].AgentName)
}

func TestSearch_ExcludesUnrestricted(t *testing.T) {
    // Add unrestricted agent with matching capability
    addUnrestrictedAgent(t, db, "secret-agent", "code-review")
    
    results, err := searcher.Search("code review", 10)
    
    for _, r := range results {
        assert.NotEqual(t, "secret-agent", r.AgentName)
    }
}

func TestSearch_RespectLimit(t *testing.T) {
    results, err := searcher.Search("anything", 5)
    
    assert.LessOrEqual(t, len(results), 5)
}

func TestSearch_EmptyResults(t *testing.T) {
    results, err := searcher.Search("quantum teleportation", 3)
    
    assert.NoError(t, err)
    assert.Empty(t, results)
}
```

## Acceptance Criteria

- [ ] Mock LLM provider for deterministic tests
- [ ] Inference tests for diverse agent types
- [ ] Tests for skill-based capability inference
- [ ] Tests for empty/minimal agents
- [ ] Tests for explicit restrictions in prompts
- [ ] Cache hit/miss tests
- [ ] Embedding vector generation tests
- [ ] Embedding similarity tests
- [ ] Search tests with seeded data
- [ ] Unrestricted agent exclusion tests
- [ ] All tests pass with `go test ./internal/capabilities/...`
- [ ] Code coverage > 80%

