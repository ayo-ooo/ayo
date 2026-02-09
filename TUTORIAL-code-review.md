# Tutorial: Code Review Pipeline

This tutorial shows you how to build a multi-agent code review pipeline that analyzes code, identifies issues, and generates actionable feedback.

## What You'll Build

A code review system with three specialized agents:
1. **@analyzer** - Identifies code patterns and potential issues
2. **@security** - Focuses on security vulnerabilities
3. **@reviewer** - Synthesizes findings into actionable review

## Prerequisites

- ayo installed and configured
- Sandbox service running (`ayo sandbox service start`)
- Mount access granted to your code directory

## Step 1: Grant File Access

First, grant ayo access to the code you want to review:

```bash
# Grant access to your project
ayo mount add /path/to/your/project

# Verify the grant
ayo mount list
```

## Step 2: Create the Analyzer Agent

Create a specialized code analysis agent:

```bash
mkdir -p ~/.config/ayo/agents/@analyzer
```

Create `~/.config/ayo/agents/@analyzer/config.json`:

```json
{
  "model": "gpt-4",
  "description": "Analyzes code for patterns, complexity, and maintainability",
  "tools": ["bash"]
}
```

Create `~/.config/ayo/agents/@analyzer/system.md`:

```markdown
# Code Analyzer

You analyze code for:
- Code complexity and readability
- Design patterns (good and bad)
- Potential bugs and edge cases
- Performance concerns
- Test coverage gaps

When analyzing, provide structured output:
- File: filename
- Line: line number (if applicable)
- Category: complexity|pattern|bug|performance|testing
- Severity: info|warning|error
- Description: what you found
- Suggestion: how to improve

Focus on actionable insights, not style nitpicks.
```

## Step 3: Create the Security Agent

Create a security-focused review agent:

```bash
mkdir -p ~/.config/ayo/agents/@security
```

Create `~/.config/ayo/agents/@security/config.json`:

```json
{
  "model": "gpt-4",
  "description": "Reviews code for security vulnerabilities",
  "tools": ["bash"]
}
```

Create `~/.config/ayo/agents/@security/system.md`:

```markdown
# Security Reviewer

You review code for security issues:
- Input validation vulnerabilities
- SQL injection, XSS, CSRF risks
- Authentication/authorization flaws
- Secrets and credential exposure
- Insecure dependencies
- Path traversal vulnerabilities
- Race conditions

Output format:
- Vulnerability: name
- Severity: low|medium|high|critical
- Location: file:line
- Description: what's wrong
- Remediation: how to fix
- References: CWE/OWASP if applicable

Only report genuine security concerns, not theoretical issues.
```

## Step 4: Create the Review Flow

Create `~/.config/ayo/flows/code-review.yaml`:

```yaml
version: 1
name: code-review
description: Multi-agent code review pipeline

params:
  path:
    type: string
    default: "."
  files:
    type: string
    default: "*.go"

steps:
  # Gather code files
  - id: gather-files
    type: shell
    run: |
      find {{ params.path }} -name "{{ params.files }}" -type f | head -20

  # Read file contents
  - id: read-code
    type: shell
    run: |
      for file in $(cat << 'EOF'
      {{ steps.gather-files.stdout }}
      EOF
      ); do
        echo "=== $file ==="
        cat "$file" 2>/dev/null | head -200
        echo ""
      done
    depends_on: [gather-files]

  # Run static analysis (if available)
  - id: static-analysis
    type: shell
    run: |
      if command -v golangci-lint &> /dev/null; then
        golangci-lint run {{ params.path }}/... 2>&1 | head -50
      elif command -v go &> /dev/null; then
        go vet {{ params.path }}/... 2>&1
      else
        echo "No static analyzer available"
      fi
    continue_on_error: true

  # Code analysis
  - id: analyze
    type: agent
    agent: "@analyzer"
    prompt: |
      Analyze this code for quality issues:
      
      {{ steps.read-code.stdout }}
      
      Static analysis output:
      {{ steps.static-analysis.stdout }}
      
      Provide structured findings as JSON array.
    depends_on: [read-code, static-analysis]
    timeout: 5m

  # Security review
  - id: security
    type: agent
    agent: "@security"
    prompt: |
      Review this code for security vulnerabilities:
      
      {{ steps.read-code.stdout }}
      
      Report findings as JSON array, empty if no issues.
    depends_on: [read-code]
    timeout: 5m

  # Synthesize review
  - id: synthesize
    type: agent
    agent: "@ayo"
    prompt: |
      Create a code review summary from these analyses:
      
      ## Code Quality Analysis
      {{ steps.analyze.output }}
      
      ## Security Review
      {{ steps.security.output }}
      
      Format as a markdown code review with:
      1. Executive Summary (1-2 sentences)
      2. Critical Issues (must fix)
      3. Recommendations (should fix)
      4. Suggestions (nice to have)
      5. Overall Assessment
    depends_on: [analyze, security]
    timeout: 3m

  # Save report
  - id: save-report
    type: shell
    run: |
      mkdir -p .ayo/reviews
      cat > .ayo/reviews/review-$(date +%Y%m%d-%H%M%S).md << 'EOF'
      # Code Review Report
      Generated: $(date)
      
      {{ steps.synthesize.output }}
      EOF
      echo "Report saved to .ayo/reviews/"
    depends_on: [synthesize]
```

## Step 5: Run the Review

```bash
# Review current directory
cd /path/to/your/project
ayo flows run code-review

# Review specific path and file types
ayo flows run code-review --param path=./src --param files="*.ts"

# Review with JSON output
ayo flows run code-review --json
```

## Step 6: Add PR Integration

Create a GitHub-aware version:

```yaml
version: 1
name: pr-review
description: Review code changes in a pull request

params:
  base:
    type: string
    default: "main"

steps:
  - id: get-changed-files
    type: shell
    run: |
      git diff --name-only {{ params.base }}...HEAD | grep -E '\.(go|ts|js|py)$'

  - id: get-diff
    type: shell
    run: |
      git diff {{ params.base }}...HEAD --unified=5

  - id: review-changes
    type: agent
    agent: "@ayo"
    prompt: |
      Review this pull request diff:
      
      Changed files:
      {{ steps.get-changed-files.stdout }}
      
      Diff:
      {{ steps.get-diff.stdout }}
      
      Focus on:
      - Breaking changes
      - Logic errors
      - Missing error handling
      - Test coverage
      
      Format as PR review comments.
    depends_on: [get-changed-files, get-diff]
```

## Step 7: Automate with Triggers

### Watch for Changes

```yaml
triggers:
  - id: on-save
    type: watch
    path: ./src
    patterns: ["*.go", "*.ts"]
    debounce: 30s
```

### Pre-Commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
ayo flows run code-review --param path=. 2>&1 | tee .ayo/last-review.md

# Check for critical issues
if grep -q "Critical Issues" .ayo/last-review.md; then
  echo "Critical issues found. Review .ayo/last-review.md"
  exit 1
fi
```

## Customization

### Add Language-Specific Agents

Create `@go-reviewer`, `@ts-reviewer`, etc. with language-specific expertise.

### Integrate with CI/CD

```yaml
# GitHub Actions example
- name: Run ayo code review
  run: |
    ayo flows run pr-review --param base=${{ github.base_ref }}
```

### Add Metrics Tracking

```yaml
- id: track-metrics
  type: shell
  run: |
    echo '{"date": "'$(date -Iseconds)'", "issues": ...}' >> .ayo/review-metrics.jsonl
  depends_on: [synthesize]
```

## Troubleshooting

### Agent not finding files

Verify mount access:
```bash
ayo mount list
ayo sandbox exec ls /path/to/project
```

### Review taking too long

- Reduce file count with more specific patterns
- Increase timeout in step definition
- Split into multiple focused reviews

### Inconsistent results

- Add more specific instructions to agent prompts
- Use lower temperature models for consistency
- Cache common analysis patterns

## Complete Example

See the full flow with all features:

```bash
ayo flows show code-review
```

## Next Steps

- [TUTORIAL-daily-digest.md](TUTORIAL-daily-digest.md) - Build a daily digest flow
- [TUTORIAL-file-watcher.md](TUTORIAL-file-watcher.md) - Set up file watch triggers
