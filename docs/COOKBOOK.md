# Ayo Cookbook

## Table of Contents

1. [Basic Examples](#basic-examples)
2. [File Processing](#file-processing)
3. [Web Automation](#web-automation)
4. [Data Analysis](#data-analysis)
5. [System Monitoring](#system-monitoring)
6. [Advanced Patterns](#advanced-patterns)

## Basic Examples

### Hello World Agent

```bash
# Create agent
ayo fresh hello-agent

# Edit config.toml
[agent]
name = "hello-agent"
description = "Simple greeting agent"
model = "gpt-4o"

[cli]
mode = "interactive"
description = "Greeting agent"

# Build and run
ayo build hello-agent
./hello-agent/main "Hello"
```

### Q&A Agent

```toml
# config.toml
[agent]
name = "qa-agent"
description = "Question answering agent"
model = "gpt-4o"
temperature = 0.3

[input]
schema = '''
{
  "type": "object",
  "properties": {
    "question": { "type": "string" },
    "context": { "type": "string" }
  },
  "required": ["question"]
}
'''

[output]
schema = '''
{
  "type": "object",
  "properties": {
    "answer": { "type": "string" },
    "confidence": { "type": "number" }
  },
  "required": ["answer"]
}
'''
```

## File Processing

### File Analyzer

```bash
# Create agent
ayo fresh file-analyzer

# Edit config.toml
[agent]
name = "file-analyzer"
description = "Analyze file contents"
model = "gpt-4o"

[agent.tools]
allowed = ["file_read", "bash"]

[cli]
flags = {
  file = { type = "string", description = "File to analyze", position = 0 }
}

# Build and use
ayo build file-analyzer
./file-analyzer/main --file document.txt
```

### Document Summarizer

```toml
# config.toml
[agent]
name = "summarizer"
description = "Document summarization agent"
model = "gpt-4o"

[agent.tools]
allowed = ["file_read"]

[agent.memory]
enabled = true
scope = "conversation"

[input]
schema = '''
{
  "type": "object",
  "properties": {
    "document": { "type": "string" },
    "length": { "type": "string", "enum": ["short", "medium", "long"] }
  },
  "required": ["document"]
}
'''
```

## Web Automation

### Web Scraper

```bash
# Create web scraper agent
ayo fresh web-scraper

# Edit config.toml
[agent]
name = "web-scraper"
description = "Web content extraction agent"
model = "gpt-4o"

[agent.tools]
allowed = ["http_request", "bash"]

[cli]
flags = {
  url = { type = "string", description = "URL to scrape", position = 0 }
}

# Build and use
ayo build web-scraper
./web-scraper/main --url "https://example.com"
```

### API Client

```toml
# config.toml
[agent]
name = "api-client"
description = "REST API interaction agent"
model = "gpt-4o"

[agent.tools]
allowed = ["http_request"]

[input]
schema = '''
{
  "type": "object",
  "properties": {
    "method": { "type": "string", "enum": ["GET", "POST", "PUT", "DELETE"] },
    "url": { "type": "string" },
    "headers": { "type": "object" },
    "body": { "type": "object" }
  },
  "required": ["method", "url"]
}
'''
```

## Data Analysis

### CSV Analyzer

```bash
# Create CSV analyzer
ayo fresh csv-analyzer

# Edit config.toml
[agent]
name = "csv-analyzer"
description = "CSV data analysis agent"
model = "gpt-4o"

[agent.tools]
allowed = ["file_read", "python"]

[cli]
flags = {
  file = { type = "string", description = "CSV file to analyze", position = 0 },
  operation = { type = "string", description = "Operation to perform", position = 1 }
}

# Build and use
ayo build csv-analyzer
./csv-analyzer/main --file data.csv --operation summarize
```

### Data Transformer

```toml
# config.toml
[agent]
name = "data-transformer"
description = "Data transformation agent"
model = "gpt-4o"

[agent.tools]
allowed = ["file_read", "file_write", "python"]

[input]
schema = '''
{
  "type": "object",
  "properties": {
    "input_file": { "type": "string" },
    "output_file": { "type": "string" },
    "transformation": { "type": "string" }
  },
  "required": ["input_file", "output_file", "transformation"]
}
'''
```

## System Monitoring

### Log Analyzer

```bash
# Create log analyzer
ayo fresh log-analyzer

# Edit config.toml
[agent]
name = "log-analyzer"
description = "Log file analysis agent"
model = "gpt-4o"

[agent.tools]
allowed = ["file_read", "bash"]

[agent.memory]
enabled = true
scope = "session"

[cli]
flags = {
  log_file = { type = "string", description = "Log file to analyze", position = 0 },
  pattern = { type = "string", description = "Pattern to search for", position = 1 }
}

# Build and use
ayo build log-analyzer
./log-analyzer/main --log_file app.log --pattern "ERROR"
```

### System Monitor

```toml
# config.toml
[agent]
name = "system-monitor"
description = "System monitoring agent"
model = "gpt-4o"

[agent.tools]
allowed = ["bash", "file_read"]

[input]
schema = '''
{
  "type": "object",
  "properties": {
    "metrics": { 
      "type": "array",
      "items": { "type": "string", "enum": ["cpu", "memory", "disk", "network"] }
    },
    "threshold": { "type": "number" }
  },
  "required": ["metrics"]
}
'''
```

## Advanced Patterns

### Multi-Agent Workflow

```bash
# Create coordinator agent
ayo fresh workflow-coordinator

# Edit config.toml
[agent]
name = "workflow-coordinator"
description = "Multi-agent workflow coordinator"
model = "gpt-4o"

[agent.tools]
allowed = ["bash", "file_read", "file_write"]

[agent.memory]
enabled = true
scope = "persistent"

# Build and use
ayo build workflow-coordinator
./workflow-coordinator/main "start workflow"
```

### Conditional Execution

```toml
# config.toml
[agent]
name = "conditional-agent"
description = "Conditional execution agent"
model = "gpt-4o"

[agent.tools]
allowed = ["bash", "http_request"]

[input]
schema = '''
{
  "type": "object",
  "properties": {
    "condition": { "type": "string" },
    "true_action": { "type": "string" },
    "false_action": { "type": "string" }
  },
  "required": ["condition", "true_action", "false_action"]
}
'''
```

### Batch Processor

```bash
# Create batch processor
ayo fresh batch-processor

# Edit config.toml
[agent]
name = "batch-processor"
description = "Batch file processing agent"
model = "gpt-4o"

[agent.tools]
allowed = ["file_read", "file_write", "bash"]

[cli]
flags = {
  input_dir = { type = "string", description = "Input directory", position = 0 },
  output_dir = { type = "string", description = "Output directory", position = 1 },
  batch_size = { type = "integer", description = "Batch size", default = 10 }
}

# Build and use
ayo build batch-processor
./batch-processor/main --input_dir input/ --output_dir output/
```

### Real-time Monitor

```toml
# config.toml
[agent]
name = "real-time-monitor"
description = "Real-time system monitoring agent"
model = "gpt-4o"

[agent.tools]
allowed = ["bash", "file_read"]

[agent.memory]
enabled = true
scope = "persistent"
max_history = 1000

[input]
schema = '''
{
  "type": "object",
  "properties": {
    "interval": { "type": "integer", "minimum": 1 },
    "metrics": { "type": "array", "items": { "type": "string" } }
  },
  "required": ["interval", "metrics"]
}
'''
```

## Usage Patterns

### Interactive Mode

```bash
# Start interactive session
./my-agent/main

# Type commands interactively
> analyze this document
> summarize the results
> exit
```

### Batch Mode

```bash
# Process multiple inputs
cat inputs.txt | while read input; do
  ./my-agent/main "$input"
done
```

### Service Mode

```bash
# Run as background service
nohup ./my-agent/main --service &>

# Check status
ps aux | grep my-agent
```

### API Integration

```bash
# Use with curl
curl -X POST http://localhost:8080/api 
  -H "Content-Type: application/json" 
  -d '{"query": "analyze this data"}'
```
