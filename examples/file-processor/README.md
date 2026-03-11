# Example: File Processor

This agent can read and process files from the filesystem.

## Files

- `config.toml` - Bot configuration with file tool permissions
- `prompts/system.txt` - System prompt
- `prompts/user.txt` - User prompt template

## Usage

```bash
# Build the bot
../../build .

# Run it - it can now read files
./.build/bin/file-processor "Summarize the contents of test.txt"
```

## Features

This example demonstrates:
- Enabling file tools (`file_read`, `file_list`)
- Using tools to access the filesystem
- Processing file contents with AI
