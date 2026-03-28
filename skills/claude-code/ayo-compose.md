# ayo-compose: Orchestrate multi-agent workflows

Use this skill when the user wants to chain agents together, build multi-step workflows, or compose agents via shell pipelines.

## Overview

Ayo agents are Unix-native: they read stdin, write stdout, and return exit codes. This makes them composable with standard shell patterns — pipes, subshells, and scripts.

## Composition Patterns

### 1. Shell Pipeline

Chain agents where one's output feeds the next's input:

```bash
# Extract -> Summarize -> Translate
ayo run extractor '{"file": "report.pdf"}' --non-interactive | \
  ayo run summarizer --non-interactive | \
  ayo run translator '{"target_language": "es"}' --non-interactive
```

For tool agents (with input_schema), pipe JSON between them:

```bash
ayo run parser '{"file": "data.csv"}' --non-interactive | \
  ayo run validator --non-interactive | \
  ayo run formatter '{"output_format": "json"}' --non-interactive
```

### 2. Sequential with Intermediate Processing

Run agents in sequence, processing results between steps:

```bash
# Step 1: Analyze
analysis=$(ayo run analyzer '{"code": "src/"}' --non-interactive)

# Step 2: Process the result
critical_issues=$(echo "$analysis" | jq '[.issues[] | select(.severity == "critical")]')

# Step 3: Generate fix suggestions only if critical issues found
if [ "$(echo "$critical_issues" | jq 'length')" -gt 0 ]; then
  ayo run fixer "{\"issues\": $critical_issues}" --non-interactive
fi
```

### 3. Fan-Out (Parallel Execution)

Run multiple agents on the same input and collect results:

```bash
input='{"text": "The quarterly results show a 15% increase in revenue."}'

# Run multiple agents in parallel
security_review=$(ayo run security-reviewer "$input" --non-interactive &)
style_review=$(ayo run style-checker "$input" --non-interactive &)
fact_check=$(ayo run fact-checker "$input" --non-interactive &)
wait

# Combine results
echo "Security: $security_review"
echo "Style: $style_review"
echo "Facts: $fact_check"
```

### 4. Conditional Routing

Pick the right agent based on the task:

```bash
# Determine file type
file_type=$(file --mime-type -b "$1")

case "$file_type" in
  application/json)
    ayo run json-processor --non-interactive < "$1"
    ;;
  text/csv)
    ayo run csv-analyzer --non-interactive < "$1"
    ;;
  text/markdown)
    ayo run doc-reviewer --non-interactive < "$1"
    ;;
  *)
    ayo run general-analyzer --non-interactive < "$1"
    ;;
esac
```

### 5. Agent Discovery + Dynamic Dispatch

Use the registry to find the right agent dynamically:

```bash
# Find all tool agents
agents=$(ayo list --json --type tool)

# Find an agent that can handle the task
for name in $(echo "$agents" | jq -r '.[].Name'); do
  desc=$(ayo describe "$name" --json)
  has_input=$(echo "$desc" | jq 'has("input_schema")')
  if [ "$has_input" = "true" ]; then
    echo "Available tool: $name - $(echo "$desc" | jq -r '.description')"
  fi
done
```

## Real-World Workflow Examples

### Code Review Pipeline

```bash
#!/bin/bash
# review-pipeline.sh — multi-agent code review

file="$1"
language=$(echo "$file" | sed 's/.*\.//')

echo "Reviewing $file..."

# Step 1: Security audit
security=$(ayo run security-auditor \
  "{\"file\": \"$file\", \"language\": \"$language\"}" \
  --non-interactive 2>/dev/null)

# Step 2: Style check
style=$(ayo run style-checker \
  "{\"file\": \"$file\", \"language\": \"$language\"}" \
  --non-interactive 2>/dev/null)

# Step 3: Synthesize into final report
ayo run report-writer \
  "{\"security\": $(echo "$security" | jq -c), \"style\": $(echo "$style" | jq -c), \"file\": \"$file\"}" \
  --non-interactive
```

### Research Workflow

```bash
#!/bin/bash
# research.sh — gather, summarize, translate

topic="$1"
target_lang="${2:-en}"

# Step 1: Research the topic
research=$(ayo run researcher "$topic" --non-interactive)

# Step 2: Summarize findings
summary=$(echo "$research" | ayo run summarizer --non-interactive)

# Step 3: Translate if needed
if [ "$target_lang" != "en" ]; then
  summary=$(echo "{\"text\": $(echo "$summary" | jq -Rs), \"target_language\": \"$target_lang\"}" | \
    ayo run translator --non-interactive)
fi

echo "$summary"
```

### Data Pipeline

```bash
#!/bin/bash
# etl.sh — extract, transform, load

input_file="$1"
output_file="${2:-output.json}"

# Extract
ayo run extractor "{\"file\": \"$input_file\"}" --non-interactive | \
# Transform
ayo run transformer '{"rules": "normalize_dates,trim_whitespace"}' --non-interactive | \
# Validate
ayo run validator --non-interactive | \
# Format and save
ayo run formatter '{"output_format": "json"}' --non-interactive -o "$output_file"

echo "Pipeline complete: $output_file"
```

## Tips

- Always use `--non-interactive` when composing agents programmatically
- Redirect stderr to capture errors separately: `2>/dev/null` or `2>/tmp/errors`
- Use `jq` to reshape JSON between agents
- For long-running workflows, consider writing an orchestrator agent that calls other agents via its sandboxed shell
- Agents can be composed into new ayo agents: create a system.md that instructs the agent to use its shell tool to invoke other agents by name
