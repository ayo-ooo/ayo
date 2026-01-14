---
name: project-summary
description: Generate a comprehensive summary of a project's structure, dependencies, and key files. Use when the user asks for a project overview, wants to understand a codebase, or needs to document project structure.
metadata:
  author: ayo
  version: "1.0"
---

# Project Summary Skill

Generate a structured summary of a software project.

## When to Use

Activate this skill when the user:
- Asks "what is this project?"
- Wants to understand a new codebase
- Needs project documentation
- Asks about project structure or architecture
- Says "give me an overview" or similar

## Summary Process

### 1. Identify Project Type

Check for these files to determine language/framework:
- `package.json` → Node.js/JavaScript
- `go.mod` → Go
- `Cargo.toml` → Rust
- `pyproject.toml` or `setup.py` → Python
- `pom.xml` or `build.gradle` → Java
- `Gemfile` → Ruby
- `*.csproj` → .NET

### 2. Gather Key Information

```bash
# List root directory
ls -la

# Read README if present
cat README.md 2>/dev/null || cat readme.md 2>/dev/null || echo "No README found"

# Check dependency manifest (pick appropriate one)
cat package.json 2>/dev/null | head -50
cat go.mod 2>/dev/null | head -30
cat Cargo.toml 2>/dev/null | head -30
cat pyproject.toml 2>/dev/null | head -30
```

### 3. Analyze Structure

- Count source files by extension
- Identify main entry points (main.go, index.js, etc.)
- Note test file locations
- Check for CI/CD configuration (.github/, .gitlab-ci.yml)

### 4. Output Format

Present findings in this structure:

```markdown
## Project: {name}

**Type**: {language/framework}
**Description**: {from README or manifest}

### Structure
- `src/` - Source code
- `tests/` - Test files
- `docs/` - Documentation

### Key Files
- Entry point: {main file}
- Config: {config files}

### Dependencies
- {dependency}: {purpose if clear}
- ...

### Build & Run
- Build: `{build command}`
- Test: `{test command}`
- Run: `{run command}`
```

## Examples

**User**: "What is this project?"

Activate this skill, run the analysis commands, and present findings in the structured format above.

**User**: "Give me an overview of the codebase"

Focus on architecture, main modules, and how they connect. Use `tree` or directory listing to show structure.

**User**: "Document this project for me"

Generate a more detailed summary including all sections above, plus any notable patterns or conventions observed.
