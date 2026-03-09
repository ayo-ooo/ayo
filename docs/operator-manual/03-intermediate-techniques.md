# Intermediate Techniques

## Elevating Your Agent Development

Now that you're comfortable with basic agent creation, let's explore intermediate techniques that will make your agents more powerful and sophisticated.

## Table of Contents

1. [Prompt Engineering](#prompt-engineering)
2. [Skill Development](#skill-development)
3. [Advanced Tool Integration](#advanced-tool-integration)
4. [Memory Optimization](#memory-optimization)
5. [Error Handling and Recovery](#error-handling-and-recovery)
6. [Performance Tuning](#performance-tuning)
7. [Testing Strategies](#testing-strategies)

## Prompt Engineering

### Prompt Structure

```toml
# Basic prompt structure
[agent.prompts]
system = "You are a helpful AI assistant."
user = "{{.Input.query}}"
assistant = "I can help with that!"
```

### Advanced Prompt Techniques

#### Few-Shot Learning

```toml
[agent.prompts]
system = '''
You are an expert data analyzer.

Examples:
User: Analyze sales data
Assistant: Here's the sales trend analysis...

User: Find anomalies
Assistant: I found these anomalies...
'''
```

#### Chain-of-Thought Prompting

```toml
[agent.prompts]
system = '''
Solve problems step by step:
1. Understand the problem
2. Break it into parts
3. Solve each part
4. Combine solutions
5. Verify result

Problem: {{.Input.query}}
'''
```

#### Role-Based Prompting

```toml
[agent.prompts]
system = '''
Role: Senior Data Scientist
Expertise: Statistical analysis, machine learning, data visualization
Approach: Methodical, evidence-based, clear communication

Task: {{.Input.query}}
'''
```

### Prompt Templates

```bash
# Create prompt templates
mkdir -p my-agent/prompts

# system.md - System prompt
cat > my-agent/prompts/system.md << 'EOF'
You are an expert {{.Role}} specializing in {{.Expertise}}.
Always provide clear, concise responses with evidence.
EOF

# user.md - User prompt template
cat > my-agent/prompts/user.md << 'EOF'
User Query: {{.Input.query}}
Context: {{.Input.context}}
Additional Info: {{.Input.details}}
EOF

# Update config.toml
[agent.prompts]
system_template = "prompts/system.md"
user_template = "prompts/user.md"
```

## Skill Development

### Creating Skills

```bash
# Create a skill directory
mkdir -p my-agent/skills/analyze_data

# Create skill definition
cat > my-agent/skills/analyze_data/skill.toml << 'EOF'
name = "analyze_data"
description = "Analyze structured data"
input_schema = '''
{
  "type": "object",
  "properties": {
    "data": { "type": "array" },
    "analysis_type": { "type": "string" }
  },
  "required": ["data", "analysis_type"]
}
'''
output_schema = '''
{
  "type": "object",
  "properties": {
    "results": { "type": "object" },
    "insights": { "type": "array" }
  }
}
'''
EOF

# Create skill implementation
cat > my-agent/skills/analyze_data/skill.py << 'EOF'
#!/usr/bin/env python3
import json
import sys

def analyze(data, analysis_type):
    # Implement analysis logic
    if analysis_type == "statistics":
        return {"mean": sum(data)/len(data), "min": min(data), "max": max(data)}
    # ... other analysis types
    
def main():
    input_data = json.load(sys.stdin)
    result = analyze(input_data["data"], input_data["analysis_type"])
    print(json.dumps(result))

if __name__ == "__main__":
    main()
EOF

chmod +x my-agent/skills/analyze_data/skill.py
```

### Skill Configuration

```toml
[agent.skills]
enabled = ["analyze_data"]

[agent.skills.analyze_data]
command = "./skills/analyze_data/skill.py"
timeout = 30
retry_on_failure = true
max_retries = 2
```

### Skill Chaining

```toml
[agent.workflow]
steps = [
  { skill = "validate_input", next = "analyze_data" },
  { skill = "analyze_data", next = "generate_report" },
  { skill = "generate_report", next = "format_output" }
]
```

## Advanced Tool Integration

### Tool Chaining

```toml
[agent.tools.workflows]
process_file = [
  { tool = "file_read", args = ["{{.Input.file}}"] },
  { tool = "analyze_content", args = ["{{.PreviousResult}}"] },
  { tool = "file_write", args = ["{{.Input.output}}", "{{.PreviousResult}}"] }
]
```

### Conditional Tool Execution

```toml
[agent.tools.conditions]
use_advanced_analysis = '''
{{if eq .Input.analysis_type "advanced"}}
  ["advanced_analyzer"]
{{else}}
  ["basic_analyzer"]
{{end}}
'''
```

### Tool Result Processing

```toml
[agent.tools.post_processing]
format_json = '''
{
  "formatted": {{.Result | toJson}},
  "timestamp": {{now | date "2006-01-02T15:04:05Z07:00"}},
  "agent": "{{.Agent.Name}}"
}
'''
```

## Memory Optimization

### Memory Strategies

```toml
[agent.memory.strategies]
# Basic strategy
simple = { 
  type = "fifo",
  max_items = 100
}

# Advanced strategy
advanced = { 
  type = "semantic",
  embedding_model = "text-embedding-ada-002",
  similarity_threshold = 0.7,
  max_items = 500,
  eviction_policy = "lru"
}

# Hybrid strategy
hybrid = { 
  type = "hybrid",
  short_term = { type = "fifo", max_items = 50 },
  long_term = { type = "vector", max_items = 1000 }
}
```

### Memory Compression

```toml
[agent.memory.compression]
enabled = true
method = "semantic"  # semantic, keyword, or hybrid
compression_ratio = 0.7
min_retention_score = 0.8
```

### Context Window Management

```toml
[agent.memory.context]
max_tokens = 4096
sliding_window = true
window_size = 2000
overlap = 500
summary_strategy = "auto"
```

## Error Handling and Recovery

### Error Handling Configuration

```toml
[agent.error_handling]
max_retries = 3
retry_delay = 2
backoff_factor = 2.0
retry_codes = ["timeout", "rate_limit", "temporary_failure"]

[agent.error_handling.fallbacks]
tool_failure = "use_alternative_tool"
api_failure = "use_cached_response"
validation_failure = "request_correction"
```

### Recovery Strategies

```toml
[agent.recovery]
strategies = [
  { 
    name = "tool_timeout",
    condition = "tool_execution_timeout",
    action = "retry_with_higher_timeout",
    max_attempts = 2,
    timeout_increase = 1.5
  },
  { 
    name = "api_rate_limit",
    condition = "api_rate_limit_exceeded",
    action = "exponential_backoff",
    initial_delay = 5,
    max_delay = 60
  }
]
```

### Circuit Breakers

```toml
[agent.circuit_breakers]
tool_execution = { 
  failure_threshold = 3,
  recovery_timeout = 60,
  half_open_retries = 1
}

api_calls = { 
  failure_threshold = 5,
  recovery_timeout = 120,
  half_open_retries = 2
}
```

## Performance Tuning

### Model Optimization

```toml
[agent.model_optimization]
# Dynamic temperature adjustment
temperature = { 
  min = 0.3,
  max = 0.9,
  adjustment_rate = 0.1,
  cooldown = 5
}

# Token management
tokens = { 
  max_input = 4096,
  max_output = 2048,
  buffer = 512,
  compression_threshold = 0.9
}

# Caching
cache = { 
  enabled = true,
  size = 100,
  ttl = 3600,
  similarity_threshold = 0.9
}
```

### Parallel Execution

```toml
[agent.parallel_execution]
enabled = true
max_concurrent = 4
queue_size = 10
timeout = 30
strategy = "round_robin"  # round_robin, priority, or adaptive
```

### Resource Management

```toml
[agent.resources]
memory_limit = "512MB"
cpu_limit = 0.8
concurrency_limit = 4
graceful_shutdown_timeout = 30
```

## Testing Strategies

### Unit Testing

```bash
# Create test directory
mkdir -p my-agent/tests

# Create test configuration
cat > my-agent/tests/config_test.toml << 'EOF'
[agent]
name = "test-agent"
model = "gpt-4o"

[input]
schema = { type = "object", properties = { test_input = { type = "string" } } }

[output]
schema = { type = "object", properties = { test_output = { type = "string" } } }
EOF

# Create test cases
cat > my-agent/tests/test_cases.json << 'EOF'
[
  {
    "name": "Basic functionality",
    "input": { "test_input": "hello" },
    "expected": { "test_output": "world" },
    "description": "Test basic input/output"
  },
  {
    "name": "Error handling",
    "input": { "test_input": "error" },
    "expected_error": "invalid_input",
    "description": "Test error handling"
  }
]
EOF

# Run tests
ayo test my-agent --test-config tests/config_test.toml --test-cases tests/test_cases.json
```

### Integration Testing

```toml
[agent.testing.integration]
enabled = true
test_suite = "tests/integration"
coverage_threshold = 0.8
timeout = 60

[agent.testing.integration.environment]
TEST_ENV = "integration"
DEBUG = "true"
```

### Performance Testing

```toml
[agent.testing.performance]
enabled = true
concurrency_levels = [1, 4, 8, 16]
duration = 60
target_rps = 100

[agent.testing.performance.metrics]
latency_p99 = 500
error_rate = 0.01
throughput = 50
```

## Practical Intermediate Examples

### Example 1: Multi-Step Data Pipeline

```bash
# Create pipeline agent
ayo fresh data-pipeline

# Edit config.toml
[agent]
name = "data-pipeline"
description = "Multi-step data processing"
model = "gpt-4o"

[agent.tools]
allowed = ["file_read", "python", "file_write"]

[agent.workflow]
steps = [
  { 
    name = "load_data",
    tool = "file_read",
    args = ["{{.Input.file}}"]
  },
  { 
    name = "clean_data",
    tool = "python",
    args = ["clean.py", "{{.PreviousResult}}"]
  },
  { 
    name = "analyze_data",
    tool = "python",
    args = ["analyze.py", "{{.PreviousResult}}"]
  },
  { 
    name = "save_results",
    tool = "file_write",
    args = ["{{.Input.output}}", "{{.PreviousResult}}"]
  }
]

[agent.memory]
enabled = true
scope = "session"
max_history = 100

# Build and test
ayo build data-pipeline
./data-pipeline/main --file input.csv --output results.json
```

### Example 2: Web Automation Agent

```bash
# Create web automation agent
ayo fresh web-automation

# Edit config.toml
[agent]
name = "web-automation"
description = "Automated web interactions"
model = "gpt-4o"
temperature = 0.2

[agent.tools]
allowed = ["http_request", "bash", "file_write"]

[agent.tools.workflows]
scrape_and_analyze = [
  { 
    tool = "http_request",
    args = ["GET", "{{.Input.url}}"]
  },
  { 
    tool = "bash",
    args = ["extract.sh", "{{.PreviousResult}}"]
  },
  { 
    tool = "file_write",
    args = ["{{.Input.output}}", "{{.PreviousResult}}"]
  }
]

[agent.error_handling]
max_retries = 3
retry_delay = 5

# Build and test
ayo build web-automation
./web-automation/main --url "https://example.com" --output results.json
```

### Example 3: Document Processing System

```bash
# Create document processor
ayo fresh document-processor

# Edit config.toml
[agent]
name = "document-processor"
description = "Advanced document processing"
model = "gpt-4o"

[agent.tools]
allowed = ["file_read", "python", "http_request"]

[agent.skills]
enabled = ["ocr", "summarization", "translation"]

[agent.memory]
enabled = true
scope = "persistent"
max_history = 500
embedding_model = "text-embedding-ada-002"

[agent.prompts]
system = '''
You are an expert document processing system.
Handle documents with care and precision.
'''

# Build and test
ayo build document-processor
./document-processor/main --file document.pdf --language en --output summary.txt
```

## Best Practices for Intermediate Users

### Prompt Engineering

1. **Be Specific**: Clearly define the agent's role and constraints
2. **Use Examples**: Provide few-shot examples for complex tasks
3. **Structure Matters**: Use clear sections and formatting
4. **Iterate**: Test and refine prompts based on results

### Skill Development

1. **Single Responsibility**: Each skill should do one thing well
2. **Clear Interfaces**: Well-defined input/output schemas
3. **Error Handling**: Graceful degradation on failures
4. **Documentation**: Clear descriptions and usage examples

### Tool Integration

1. **Limit Scope**: Only enable necessary tools
2. **Set Timeouts**: Prevent hanging operations
3. **Validate Inputs**: Check tool arguments before execution
4. **Handle Errors**: Implement retry and fallback logic

### Memory Management

1. **Right Size**: Choose appropriate memory scope
2. **Limit Growth**: Set reasonable max_history values
3. **Compress**: Use semantic compression for large memories
4. **Persist**: Save important memories between sessions

### Performance Optimization

1. **Profile**: Measure before optimizing
2. **Cache**: Cache frequent similar requests
3. **Parallelize**: Use concurrent execution where possible
4. **Monitor**: Track resource usage over time

## Troubleshooting Intermediate Issues

### Workflow Problems

**Issue**: Workflow steps not executing in order

**Solution**:
```bash
# Check workflow configuration
ao checkit my-agent --workflow-only

# Validate step dependencies
ayo validate my-agent --steps
```

**Issue**: Tool chaining fails

**Solution**:
```bash
# Test individual tools
ayo test my-agent --tool file_read
ayo test my-agent --tool python

# Check tool dependencies
ayo checkit my-agent --tools
```

### Memory Problems

**Issue**: Memory grows too large

**Solution**:
```bash
# Adjust memory limits
[agent.memory]
max_history = 100  # Reduce from default

# Enable compression
[agent.memory.compression]
enabled = true
```

**Issue**: Memory retrieval slow

**Solution**:
```bash
# Use vector memory for large datasets
[agent.memory]
type = "vector"
embedding_model = "text-embedding-ada-002"

# Add indexing
[agent.memory.indexing]
enabled = true
index_type = "hnsw"
```

### Performance Problems

**Issue**: Agent responses too slow

**Solution**:
```bash
# Check model settings
[agent]
model = "gpt-4o"  # Faster than gpt-4-turbo
max_tokens = 2048  # Reduce from 4096

# Enable caching
[agent.cache]
enabled = true
size = 50
```

**Issue**: High resource usage

**Solution**:
```bash
# Limit concurrency
[agent.resources]
concurrency_limit = 2  # Reduce from 4
memory_limit = "256MB"  # Reduce from 512MB

# Optimize workflows
[agent.workflow]
parallel_steps = false  # Change from true
```

## Summary

✅ **Mastered prompt engineering**
✅ **Developed custom skills**
✅ **Implemented advanced tool integration**
✅ **Optimized memory usage**
✅ **Implemented error handling**
✅ **Tuned performance**
✅ **Established testing strategies**

You're now ready for advanced patterns and expert techniques! 🚀

**Next**: [Advanced Patterns](04-advanced-patterns.md) → Learn about multi-agent systems, complex workflows, and production deployment