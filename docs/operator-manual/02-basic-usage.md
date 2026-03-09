# Basic Usage Guide

## Building on the Foundation

Now that you've created your first agent, let's dive deeper into customization and practical usage patterns.

## Table of Contents

1. [Configuration Deep Dive](#configuration-deep-dive)
2. [Working with Tools](#working-with-tools)
3. [Input/Output Patterns](#inputoutput-patterns)
4. [Memory Configuration](#memory-configuration)
5. [CLI Flags and Arguments](#cli-flags-and-arguments)
6. [Building for Different Platforms](#building-for-different-platforms)
7. [Debugging and Validation](#debugging-and-validation)

## Configuration Deep Dive

### Agent Configuration Options

```toml
[agent]
name = "my-agent"                    # Required: Agent name
description = "My AI assistant"       # Required: Description
model = "gpt-4o"                     # Required: LLM model
temperature = 0.7                    # Optional: Creativity (0.0-1.0)
max_tokens = 4096                     # Optional: Max tokens
stop_sequences = ["\n\n"]            # Optional: Stop sequences
presence_penalty = 0.0               # Optional: Presence penalty
frequency_penalty = 0.0              # Optional: Frequency penalty
```

**Temperature Guide:**
- `0.0-0.3`: Deterministic, factual responses
- `0.4-0.7`: Balanced creativity (default)
- `0.8-1.0`: Highly creative, diverse responses

### Model Selection Guide

| Model | Best For | Temperature Range |
|-------|----------|-------------------|
| `gpt-4o` | General purpose | 0.3-0.7 |
| `claude-3-opus` | Complex reasoning | 0.2-0.6 |
| `gemini-1.5-pro` | Multimodal tasks | 0.4-0.8 |
| `llama-3-70b` | Open-source alternative | 0.5-0.9 |

## Working with Tools

### Built-in Tools Reference

| Tool | Description | Example Usage |
|------|-------------|---------------|
| `bash` | Execute shell commands | `bash "ls -la"` |
| `file_read` | Read file contents | `file_read "/path/to/file.txt"` |
| `file_write` | Write to files | `file_write "/path/output.txt" "content"` |
| `http_request` | Make HTTP requests | `http_request GET "https://api.example.com"` |
| `python` | Execute Python code | `python "import math; print(math.pi)"` |
| `node` | Execute Node.js code | `node "console.log('hello')"` |

### Configuring Tools

```toml
[agent.tools]
# Enable built-in tools
allowed = ["bash", "file_read", "file_write", "http_request"]

# Configure tool timeouts
timeout = 30  # Seconds

# Configure tool retries
max_retries = 3
retry_delay = 2  # Seconds
```

### Adding Custom Tools

```bash
# Step 1: Create tool directory
mkdir -p my-agent/tools/my_custom_tool

# Step 2: Create executable script
cat > my-agent/tools/my_custom_tool/tool.sh << 'EOF'
#!/bin/bash
# Custom tool implementation
echo "Custom tool output: $@"
EOF

# Step 3: Make executable
chmod +x my-agent/tools/my_custom_tool/tool.sh

# Step 4: Update config.toml
[agent.tools.custom]
my_custom_tool = { 
  command = "./tools/my_custom_tool/tool.sh",
  description = "My custom tool",
  timeout = 10
}
```

## Input/Output Patterns

### Basic Input/Output

```toml
[input]
schema = '''
{
  "type": "object",
  "properties": {
    "query": { "type": "string" },
    "context": { "type": "string", "description": "Optional context" }
  },
  "required": ["query"]
}
'''

[output]
schema = '''
{
  "type": "object",
  "properties": {
    "answer": { "type": "string" },
    "confidence": { "type": "number", "minimum": 0, "maximum": 1 }
  },
  "required": ["answer"]
}
'''
```

### Complex Data Structures

```toml
[input]
schema = '''
{
  "type": "object",
  "properties": {
    "user": { 
      "type": "object",
      "properties": {
        "id": { "type": "string" },
        "name": { "type": "string" },
        "preferences": { "type": "object" }
      },
      "required": ["id", "name"]
    },
    "items": { 
      "type": "array",
      "items": { 
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "quantity": { "type": "integer" }
        },
        "required": ["id", "quantity"]
      }
    }
  },
  "required": ["user", "items"]
}
'''
```

### Validation Patterns

```toml
[input]
schema = '''
{
  "type": "object",
  "properties": {
    "email": { 
      "type": "string",
      "format": "email",
      "pattern": "^[^@]+@[^@]+\\.[^@]+$"
    },
    "age": { 
      "type": "integer",
      "minimum": 18,
      "maximum": 120
    },
    "password": { 
      "type": "string",
      "minLength": 8,
      "maxLength": 64
    }
  },
  "required": ["email", "age", "password"]
}
'''
```

## Memory Configuration

### Memory Types Explained

| Type | Duration | Use Case |
|------|----------|----------|
| `conversation` | Single interaction | Short-term context |
| `session` | Multiple interactions | Multi-turn conversations |
| `persistent` | Across sessions | Long-term learning |

### Memory Configuration Examples

```toml
# Basic conversation memory
[agent.memory]
enabled = true
scope = "conversation"
max_history = 50

# Session memory with embedding
[agent.memory]
enabled = true
scope = "session"
max_history = 200
embedding_model = "text-embedding-ada-002"
similarity_threshold = 0.7

# Persistent memory with vector storage
[agent.memory]
enabled = true
scope = "persistent"
max_history = 1000
storage = "vector"
persist_path = "./memory.db"
```

### Memory Best Practices

1. **Conversation Memory**: Use for single-question agents
2. **Session Memory**: Use for multi-turn interactions
3. **Persistent Memory**: Use for agents that learn over time
4. **Limit History**: Start with 50-200 items, adjust as needed
5. **Use Embeddings**: For better semantic search in large memories

## CLI Flags and Arguments

### Basic CLI Configuration

```toml
[cli]
mode = "interactive"  # interactive, batch, or service
description = "My CLI agent"

[cli.flags]
# Positional arguments
input_file = { 
  type = "string", 
  description = "Input file path",
  position = 0
}

# Optional flags
verbose = { 
  type = "bool", 
  description = "Enable verbose output",
  short = "v"
}
quiet = { 
  type = "bool", 
  description = "Suppress output",
  short = "q"
}
```

### Advanced CLI Patterns

```toml
[cli]
mode = "interactive"
description = "Advanced CLI agent"

[cli.flags]
# Required flags
config = { 
  type = "string",
  description = "Configuration file",
  required = true,
  short = "c"
}

# Flags with defaults
timeout = { 
  type = "integer",
  description = "Operation timeout in seconds",
  default = 30,
  short = "t"
}

# Enum flags
log_level = { 
  type = "string",
  description = "Log level",
  enum = ["debug", "info", "warn", "error"],
  default = "info"
}

# Array flags
tags = { 
  type = "array",
  description = "Tags for categorization",
  item_type = "string"
}
```

### CLI Usage Examples

```bash
# Basic usage
./my-agent/main "Hello"

# With flags
./my-agent/main --verbose --timeout 60 "Process this"

# Short flags
./my-agent/main -v -t 60 "Process this"

# Required flags
./my-agent/main --config config.json "Run"

# Array flags
./my-agent/main --tags tag1 --tags tag2 "Categorize"
```

## Building for Different Platforms

### Cross-Platform Build Commands

```bash
# Build for current platform
ao build my-agent

# Build for Linux (64-bit)
GOOS=linux GOARCH=amd64 ayo build my-agent

# Build for Windows (64-bit)
GOOS=windows GOARCH=amd64 ayo build my-agent

# Build for macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 ayo build my-agent

# Build for macOS (Intel)
GOOS=darwin GOARCH=amd64 ayo build my-agent
```

### Platform-Specific Considerations

| Platform | Notes |
|----------|-------|
| **Linux** | Works on most distributions |
| **Windows** | Requires Windows 10+ |
| **macOS** | Supports Intel & Apple Silicon |
| **Docker** | Use `linux/amd64` base image |

### Distribution Strategies

```bash
# Create a distribution package
tar -czvf my-agent-linux.tar.gz my-agent/

# Create platform-specific packages
mkdir -p dist/linux dist/windows dist/macos
GOOS=linux GOARCH=amd64 ayo build my-agent -o dist/linux/my-agent
GOOS=windows GOARCH=amd64 ayo build my-agent -o dist/windows/my-agent.exe
GOOS=darwin GOARCH=arm64 ayo build my-agent -o dist/macos/my-agent

# Create zip archives
cd dist/linux && tar -czvf ../my-agent-linux.tar.gz . && cd ../..
cd dist/windows && zip -r ../my-agent-windows.zip . && cd ../..
cd dist/macos && tar -czvf ../my-agent-macos.tar.gz . && cd ../..
```

## Debugging and Validation

### Configuration Validation

```bash
# Basic validation
ao checkit my-agent

# Verbose validation
ao checkit my-agent --verbose

# Check specific aspects
ayo checkit my-agent --config-only
ayo checkit my-agent --schema-only
```

### Common Validation Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `Invalid TOML` | Syntax error in config.toml | Fix TOML syntax |
| `Missing required field` | Required field not specified | Add missing field |
| `Invalid schema` | Malformed JSON Schema | Fix schema syntax |
| `Unknown tool` | Tool not in allowed list | Add tool to config |

### Debugging Build Issues

```bash
# Clean build
ao build my-agent --clean

# Verbose build
ayo build my-agent --verbose

# Check dependencies
go mod tidy
ayo build my-agent
```

### Runtime Debugging

```bash
# Verbose mode
./my-agent/main --verbose "test"

# Debug mode
./my-agent/main --debug "test"

# Trace mode (very detailed)
./my-agent/main --trace "test"
```

## Practical Examples

### Example 1: File Processing Agent

```bash
# Create agent
ao fresh file-processor

# Edit config.toml
[agent]
name = "file-processor"
description = "Process text files"
model = "gpt-4o"

[agent.tools]
allowed = ["file_read", "file_write", "bash"]

[cli]
flags = {
  input = { type = "string", description = "Input file", position = 0, required = true },
  output = { type = "string", description = "Output file", position = 1, required = true },
  operation = { type = "string", description = "Operation to perform", position = 2, 
                enum = ["summarize", "analyze", "transform"], default = "summarize" }
}

# Build and test
ao build file-processor
./file-processor/main input.txt output.txt analyze
```

### Example 2: Web Scraper Agent

```bash
# Create agent
ao fresh web-scraper

# Edit config.toml
[agent]
name = "web-scraper"
description = "Extract content from websites"
model = "gpt-4o"
temperature = 0.3

[agent.tools]
allowed = ["http_request", "bash"]

[agent.memory]
enabled = true
scope = "session"

[cli]
flags = {
  url = { type = "string", description = "URL to scrape", position = 0, required = true },
  selector = { type = "string", description = "CSS selector", position = 1 },
  depth = { type = "integer", description = "Scraping depth", default = 1 }
}

# Build and test
ao build web-scraper
./web-scraper/main "https://example.com" ".content" --depth 2
```

### Example 3: Data Analysis Agent

```bash
# Create agent
ao fresh data-analyst

# Edit config.toml
[agent]
name = "data-analyst"
description = "Analyze structured data"
model = "gpt-4o"

[agent.tools]
allowed = ["file_read", "python"]

[input]
schema = '''
{
  "type": "object",
  "properties": {
    "data": { "type": "array" },
    "analysis_type": { 
      "type": "string",
      "enum": ["statistics", "trends", "anomalies", "correlations"]
    }
  },
  "required": ["data", "analysis_type"]
}
'''

[output]
schema = '''
{
  "type": "object",
  "properties": {
    "results": { "type": "object" },
    "visualization": { "type": "string" },
    "insights": { "type": "array", "items": { "type": "string" } }
  },
  "required": ["results", "insights"]
}
'''

# Build and test
ao build data-analyst
./data-analyst/main '{"data": [1,2,3,4,5], "analysis_type": "statistics"}'
```

## Best Practices for Basic Usage

### Configuration Organization

1. **Start Simple**: Begin with minimal configuration
2. **Add Gradually**: Add features as needed
3. **Validate Often**: Run `ayo checkit` frequently
4. **Document**: Add comments to your config.toml

### Tool Usage

1. **Limit Tools**: Only enable what you need
2. **Set Timeouts**: Prevent hanging operations
3. **Handle Errors**: Configure retry logic
4. **Test Individually**: Verify each tool works

### Performance Tips

1. **Optimize Temperature**: Lower for factual, higher for creative
2. **Limit Token Usage**: Set appropriate max_tokens
3. **Use Efficient Models**: Choose model based on task
4. **Cache Responses**: For repeated similar queries

## Troubleshooting Common Issues

### Build Problems

**Issue**: `build failed: exit status 1`

**Solution**:
```bash
# Check for syntax errors
ao checkit my-agent --verbose

# Clean and rebuild
ayo build my-agent --clean
```

**Issue**: Missing dependencies

**Solution**:
```bash
# Update Go modules
go mod tidy

# Rebuild
ao build my-agent
```

### Runtime Problems

**Issue**: Agent crashes on startup

**Solution**:
```bash
# Run in verbose mode
./my-agent/main --verbose "test"

# Check configuration
ayo checkit my-agent
```

**Issue**: Tool execution fails

**Solution**:
```bash
# Test tool individually
./my-agent/tools/my_tool/tool.sh "test"

# Check permissions
chmod +x my-agent/tools/my_tool/tool.sh
```

### Configuration Problems

**Issue**: Invalid schema

**Solution**:
```bash
# Validate schema separately
ao checkit my-agent --schema-only

# Use JSON schema validator
ajv validate -s config.toml
```

**Issue**: Missing required fields

**Solution**:
```bash
# Check configuration
ao checkit my-agent

# Add missing fields to config.toml
```

## Summary

✅ **Mastered agent configuration**
✅ **Learned tool system**
✅ **Understood I/O patterns**
✅ **Configured memory systems**
✅ **Created CLI interfaces**
✅ **Built for multiple platforms**
✅ **Debugged common issues**

You're now ready for more advanced techniques! 🚀

**Next**: [Intermediate Techniques](03-intermediate-techniques.md) → Learn about prompts, skills, and advanced configuration