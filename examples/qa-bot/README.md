# Example: Simple Q&A Bot

This is a simple Q&A bot that answers questions using GPT-4.

## Files

- `config.toml` - Bot configuration
- `prompts/system.txt` - System prompt defining the bot's behavior
- `prompts/user.txt` - User prompt template

## Usage

```bash
# Build the bot
../../build .

# Run it
./.build/bin/qa-bot "What is the capital of France?"

# Or in dev mode
../../dev . --run
```

## Configuration

This example uses:
- Freeform CLI mode for simple text input/output
- No custom skills or tools
- Default GPT-4 model
