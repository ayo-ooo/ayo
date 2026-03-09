# Ayo Operator Manual

## Table of Contents

1. [Getting Started](#getting-started)
2. [Basic Usage](#basic-usage)
3. [Configuration Reference](#configuration-reference)
4. [Advanced Configuration](#advanced-configuration)
5. [Tool System](#tool-system)
6. [Memory System](#memory-system)
7. [Building and Distribution](#building-and-distribution)
8. [Troubleshooting](#troubleshooting)
9. [Best Practices](#best-practices)

## Getting Started

### Installation

```bash
# Install via Homebrew
brew tap alexcabrera/ayo
brew install ayo

# Or build from source
git clone https://github.com/alexcabrera/ayo.git
cd ayo
make install
```

### Quick Start

```bash
# Create a new agent
ayo fresh my-agent

# Build the agent
ayo build my-agent

# Run the agent
./my-agent/main "Hello, world!"
```

## Basic Usage

### Creating Agents

```bash
# Create a new agent project
ayo fresh agent-name

# This creates:
# agent-name/
#   ├── config.toml      # Main configuration
#   ├── prompts/         # Prompt templates
#   ├── skills/          # Agent skills
#   ├── tools/           # Custom tools
#   └── main             # Built executable
```

### Configuration Structure

```toml
# config.toml
[agent]
name = "my-agent"
description = "My AI agent"
model = "gpt-4o"

[cli]
mode = "interactive"
description = "My CLI agent"

[input]
schema = { type = "object", properties = { query = { type = "string" } } }

[output]
schema = { type = "object", properties = { result = { type = "string" } } }

[agent.tools]
allowed = ["bash", "file_read", "file_write"]

[agent.memory]
enabled = true
scope = "conversation"
```

### Building Agents

```bash
# Build agent
ayo build agent-name

# Build with specific output name
ayo build agent-name -o my-executable

# Build for different platforms
GOOS=linux GOARCH=amd64 ayo build agent-name
```

## Configuration Reference

### Agent Configuration

```toml
[agent]
name = "string"           # Agent name
description = "string"    # Agent description
model = "string"          # LLM model (gpt-4o, claude-3-opus, etc.)
temperature = 0.7          # Creativity (0.0-1.0)
max_tokens = 4096          # Maximum tokens
```

### CLI Configuration

```toml
[cli]
mode = "interactive"      # Mode: interactive, batch, or service
description = "string"    # CLI description

[cli.flags]
# Define CLI flags
verbose = { type = "bool", description = "Enable verbose output" }
quiet = { type = "bool", description = "Suppress output" }
```

### Input/Output Schemas

```toml
[input]
# JSON Schema for input validation
schema = '''
{
  "type": "object",
  "properties": {
    "query": { "type": "string" },
    "context": { "type": "string" }
  },
  "required": ["query"]
}
'''

[output]
# JSON Schema for output validation
schema = '''
{
  "type": "object",
  "properties": {
    "result": { "type": "string" },
    "confidence": { "type": "number", "minimum": 0, "maximum": 1 }
  },
  "required": ["result"]
}
'''
```

## Advanced Configuration

### Tool Configuration

```toml
[agent.tools]
# Built-in tools
allowed = ["bash", "file_read", "file_write", "http_request"]

# Custom tools
[agent.tools.custom]
my_tool = { 
  command = "/path/to/tool",
  description = "My custom tool"
}
```

### Memory Configuration

```toml
[agent.memory]
enabled = true
scope = "conversation"  # conversation, session, or persistent
max_history = 100       # Maximum memory items
embedding_model = "text-embedding-ada-002"
```

### Performance Optimization

```toml
[performance]
concurrency = 4         # Maximum concurrent operations
cache_size = 100        # Response cache size
timeout = 30            # Operation timeout (seconds)
```

## Tool System

### Built-in Tools

| Tool | Description | Usage |
|------|-------------|-------|
| `bash` | Execute shell commands | `bash "ls -la"` |
| `file_read` | Read file contents | `file_read "/path/to/file"` |
| `file_write` | Write to files | `file_write "/path/to/file" "content"` |
| `http_request` | Make HTTP requests | `http_request GET "https://api.example.com"` |
| `python` | Execute Python code | `python "print('hello')"` |

### Custom Tools

```bash
# Add custom tool to tools/ directory
mkdir -p my-agent/tools/my_tool

# Create tool script (any executable)
#!/bin/bash
echo "Custom tool output"

# Make executable
chmod +x my-agent/tools/my_tool/tool.sh
```

## Memory System

### Memory Types

1. **Conversation Memory**: Short-term memory within a single interaction
2. **Session Memory**: Persists across multiple interactions in a session
3. **Persistent Memory**: Long-term memory stored between sessions

### Memory Usage

```toml
[agent.memory]
enabled = true
type = "vector"        # vector, keyword, or hybrid
persist = true         # Persist between sessions
max_items = 1000       # Maximum memory items
```

## Building and Distribution

### Cross-Platform Building

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 ayo build my-agent

# Build for Windows
GOOS=windows GOARCH=amd64 ayo build my-agent

# Build for macOS
GOOS=darwin GOARCH=arm64 ayo build my-agent
```

### Distribution

```bash
# Create distribution package
tar -czvf my-agent.tar.gz my-agent/

# Install on target system
tar -xzvf my-agent.tar.gz
cd my-agent
./main "Hello"
```

## Troubleshooting

### Common Issues

**Build fails with missing dependencies:**
```bash
go mod tidy
ayo build my-agent
```

**Agent fails to start:**
```bash
# Check configuration
ayo checkit my-agent

# Run in verbose mode
./main --verbose "test"
```

**Permission denied:**
```bash
chmod +x my-agent/main
./my-agent/main "test"
```

## Best Practices

### Configuration

1. **Use environment variables for secrets:**
   ```toml
   [agent]
   api_key = "${API_KEY}"
   ```

2. **Validate schemas thoroughly:**
   ```bash
   ayo checkit my-agent --verbose
   ```

3. **Start with simple configurations:**
   ```toml
   [agent]
   name = "simple-agent"
   model = "gpt-4o"
   ```

### Development Workflow

```bash
# 1. Create agent
ayo fresh my-agent

# 2. Test configuration
ayo checkit my-agent

# 3. Build and test
ayo build my-agent
./my-agent/main "test input"

# 4. Iterate and improve
# Edit config.toml, prompts/, skills/
# Repeat build and test
```

### Performance Optimization

1. **Limit tool usage:**
   ```toml
   [agent.tools]
   allowed = ["bash", "file_read"]
   ```

2. **Use appropriate memory scope:**
   ```toml
   [agent.memory]
   scope = "conversation"  # For short interactions
   ```

3. **Optimize LLM settings:**
   ```toml
   [agent]
   temperature = 0.3  # Lower for deterministic outputs
   max_tokens = 2048 # Balance quality and cost
   ```
