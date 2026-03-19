# Model Selection TUI

The model selection TUI provides an interactive first-run experience for configuring which AI provider and model to use.

## When Triggered

The TUI is displayed when:

1. No config file exists (`~/.config/agents/{agent-name}.toml`)
2. Config exists but is incomplete (missing provider or model)
3. User didn't specify `--provider` and `--model` flags

### Decision Flow

```
ensureConfig() called
    │
    ├─► Config exists with provider + model? ──► YES ──► Return (use config)
    │
    NO
    │
    ├─► --provider AND --model flags set? ──► YES ──► Save to config, return
    │
    NO
    │
    └─► Launch selectModel() TUI
```

## Environment Scanner

The TUI scans environment variables to detect available providers:

| Provider | Environment Variable |
|----------|---------------------|
| Anthropic | `ANTHROPIC_API_KEY` |
| OpenAI | `OPENAI_API_KEY` |
| Gemini | `GEMINI_API_KEY` |
| Groq | `GROQ_API_KEY` |
| OpenRouter | `OPENROUTER_API_KEY` |
| Zai | `ZAI_API_KEY |

### Scanner Interface

```go
type ProviderInfo struct {
    Name        string
    APIKeyEnv   string
    APIKey      string // Empty if not set
    IsAvailable bool
    Models      []ModelInfo
}

type ModelInfo struct {
    ID          string
    Name        string
    Description string
}

func ScanProviders() []ProviderInfo
func FetchModels(provider string, apiKey string) ([]ModelInfo, error)
```

## TUI Flow

### Step 1: Provider Selection

```
╭────────────────────────────────────────╮
│ Select a Provider                      │
├────────────────────────────────────────┤
│ > Anthropic (API key detected)         │
│   OpenAI (API key detected)            │
│   OpenRouter (API key detected)        │
│   Gemini                               │
│   Groq                                 │
│   Enter manually...                    │
╰────────────────────────────────────────╯
```

- Providers with API keys shown first with indicator
- Providers without keys shown but dimmed
- "Enter manually" option for custom providers

### Step 2: Model Selection

```
╭────────────────────────────────────────╮
│ Select a Model (Anthropic)             │
├────────────────────────────────────────┤
│ > claude-3-5-sonnet-20241022           │
│   claude-3-5-haiku-20241022            │
│   claude-3-opus-20240229               │
│   Enter custom model ID...             │
╰────────────────────────────────────────╯
```

- Models fetched from provider API when possible
- Fallback to static model list if API unavailable
- Custom model ID option always available

### Step 3: Confirmation

```
╭────────────────────────────────────────╮
│ Configuration Summary                  │
├────────────────────────────────────────┤
│ Provider: anthropic                    │
│ Model: claude-3-5-sonnet-20241022      │
│                                        │
│ Config will be saved to:               │
│ ~/.config/agents/{agent-name}.toml     │
│                                        │
│ > Save and continue                    │
│   Back                                 │
│   Cancel                               │
╰────────────────────────────────────────╯
```

## Config Saving

After selection, config is saved to:

```
~/.config/agents/{agent-name}.toml
```

Format:
```toml
provider = "anthropic"
model = "claude-3-5-sonnet-20241022"
api_key = ""  # Not stored, read from env at runtime
```

**Important**: API keys are NOT stored in config. They are read from environment variables at runtime.

## Cancel Behavior

- Pressing `Esc` or selecting "Cancel" exits cleanly with code 0
- No error message displayed
- No config file created
- User can run again with flags or set environment variables

## Error Handling

| Error | Behavior |
|-------|----------|
| No providers available | Show message directing user to set API keys, exit 1 |
| Model fetch fails | Fall back to static model list |
| Config save fails | Show error, offer retry or continue without saving |
| Invalid selection | Show inline error, allow retry |

## Implementation Notes

### Generated Code Location

The `selectModel()` function is generated in `agent.go`:

```go
func selectModel() error {
    return fmt.Errorf("no model configured - use --provider and --model flags or set up config interactively (TUI not yet implemented)")
}
```

### Required Changes

1. Create `internal/tui/select.go` with TUI implementation using Bubble Tea
2. Create `internal/generate/tui.go` to generate TUI-related code
3. Update `GenerateAgent` to include TUI code instead of error return
4. Add model fetching logic for each provider

### Static Model Lists

Fallback when API unavailable:

```go
var defaultModels = map[string][]ModelInfo{
    "anthropic": {
        {ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet"},
        {ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku"},
        {ID: "claude-3-opus-20240229", Name: "Claude 3 Opus"},
    },
    "openai": {
        {ID: "gpt-4o", Name: "GPT-4o"},
        {ID: "gpt-4o-mini", Name: "GPT-4o Mini"},
        {ID: "gpt-4-turbo", Name: "GPT-4 Turbo"},
    },
    // ... other providers
}
```
