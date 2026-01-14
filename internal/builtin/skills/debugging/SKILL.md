---
name: debugging
description: Systematic debugging techniques for diagnosing issues. Use when troubleshooting errors, unexpected behavior, test failures, or when the user is stuck on a problem.
compatibility: Requires bash and common Unix utilities
metadata:
  author: ayo
  version: "1.0"
---

# Debugging Skill

A systematic approach to diagnosing and fixing issues.

## When to Use

Activate this skill when:
- User reports an error or unexpected behavior
- Tests are failing
- User says "it's not working" or "help me debug"
- User is stuck on a problem
- Something that "should work" doesn't

## Systematic Debugging Process

### 1. Reproduce the Issue

First, confirm the problem consistently:
```bash
# Run the failing command/test
{the command that fails}

# Check exit code
echo "Exit code: $?"
```

**Key questions:**
- Does it fail every time?
- Does it fail in the same way?
- When did it start failing?

### 2. Gather Information

```bash
# Check recent changes (if in git repo)
git log --oneline -10
git diff HEAD~3

# Check environment
env | grep -i relevant_var

# Check logs
tail -100 /path/to/log 2>/dev/null

# Check running processes if relevant
ps aux | grep process_name
```

### 3. Isolate the Problem

**Binary search approach:**
1. Find the last known good state
2. Check the midpoint between good and bad
3. Narrow down to the specific change/line

```bash
# Git bisect for regression hunting
git bisect start
git bisect bad HEAD
git bisect good {known_good_commit}
# Then test at each bisect point
```

### 4. Form Hypotheses

Common causes to check:
- **Configuration**: Wrong environment, missing config
- **Dependencies**: Version mismatch, missing package
- **State**: Stale cache, corrupted data
- **Permissions**: File access, network access
- **Race conditions**: Timing issues in concurrent code

### 5. Test Hypotheses

For each hypothesis:
1. Make a prediction about what you'll find
2. Run a specific test to confirm/deny
3. Document the result

```bash
# Example: Testing if it's a cache issue
rm -rf .cache/ && {command}

# Example: Testing if it's a dependency issue
{package_manager} install --force && {command}
```

### 6. Fix and Verify

After identifying the cause:
1. Make the minimal fix
2. Verify the original issue is resolved
3. Check for regressions (run related tests)
4. Document what the problem was

## Common Debugging Commands

```bash
# Verbose output
{command} -v 2>&1 | tee debug.log

# Trace system calls (Linux)
strace -f {command}

# Network debugging
curl -v {url}
nc -zv {host} {port}

# File permissions
ls -la {path}
stat {file}

# Process inspection
lsof -p {pid}
```

## Output Format

When reporting findings:

```markdown
## Debug Summary

**Issue**: {description}
**Root Cause**: {what was wrong}
**Solution**: {what fixed it}

### Investigation Steps
1. {step 1}
2. {step 2}
...

### Lessons Learned
- {insight for future}
```
