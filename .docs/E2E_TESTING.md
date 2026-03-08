# Ayo Build System - End-to-End Testing Guide

This document provides exhaustive step-by-step testing instructions for the Ayo build system. Each test is broken down to its atomic level with exact commands, expected outputs, and validation criteria.

---

## Table of Contents

1. [Test Environment Setup](#test-environment-setup)
2. [Core Functionality Tests](#core-functionality-tests)
3. [Progressive Team Creation Tests](#progressive-team-creation-tests)
4. [Configuration Validation Tests](#configuration-validation-tests)
5. [Error Handling Tests](#error-handling-tests)
6. [Command Integration Tests](#command-integration-tests)
7. [Performance Testing](#performance-testing)
8. [Edge Case Testing](#edge-case-testing)
9. [Regression Testing](#regression-testing)
10. [Cleanup and Verification](#cleanup-and-verification)

---

## Test Environment Setup

### Prerequisites Check

**Test 1.1: Verify Go installation**
```bash
go version
```
**Expected Output**: `go version go1.21+` (or similar)
**Validation**: Exit code 0, version ≥ 1.20

**Test 1.2: Verify Git installation**
```bash
git --version
```
**Expected Output**: `git version 2.x+` (or similar)
**Validation**: Exit code 0, version ≥ 2.0

**Test 1.3: Check working directory**
```bash
pwd && ls -la
```
**Expected Output**: Should be in Ayo repository root
**Validation**: Should see `cmd/`, `internal/`, `docs/` directories

**Test 1.4: Build Ayo binary**
```bash
go build -o /tmp/ayo-test ./cmd/ayo/
```
**Expected Output**: No errors
**Validation**: `/tmp/ayo-test` file exists and is executable

**Test 1.5: Verify binary works**
```bash
/tmp/ayo-test --help
```
**Expected Output**: Help text with available commands
**Validation**: Exit code 0, shows command list

---

## Core Functionality Tests

### Single Agent Project Creation

**Test 2.1: Create single agent project**
```bash
cd /tmp && rm -rf test-project-1 && /tmp/ayo-test fresh test-project-1
```
**Expected Output**: 
```
Created new agent project: test-project-1
Agent configuration: test-project-1/agents/main/config.toml
```
**Validation**: 
- `test-project-1/` directory exists
- `test-project-1/config.toml` exists
- `test-project-1/agents/main/config.toml` exists
- `test-project-1/agents/main/prompts/system.md` exists

**Test 2.2: Verify project structure**
```bash
find test-project-1 -type f | sort
```
**Expected Output**:
```
test-project-1/agents/main/config.toml
test-project-1/agents/main/prompts/system.md
test-project-1/agents/main/skills/custom/SKILL.md
test-project-1/config.toml
```
**Validation**: All expected files present

**Test 2.3: Check config.toml content**
```bash
cat test-project-1/config.toml
```
**Expected Output**: Valid TOML with agent configuration
**Validation**: Contains `[agent]`, `[cli]`, `[agent.tools]` sections

**Test 2.4: Check agent config.toml content**
```bash
cat test-project-1/agents/main/config.toml
```
**Expected Output**: Valid TOML with agent-specific configuration
**Validation**: Contains `name = "main"`, model specification

---

## Progressive Team Creation Tests

### Automatic Team Promotion

**Test 3.1: Add second agent (triggers team promotion)**
```bash
cd /tmp && rm -rf test-project-2 && /tmp/ayo-test fresh test-project-2
/tmp/ayo-test add-agent test-project-2 reviewer
```
**Expected Output**:
```
Promoted project to team format with 2 agents
Added agent 'reviewer' to project: test-project-2
```
**Validation**:
- `test-project-2/team.toml` exists
- `test-project-2/SQUAD.md` exists
- `test-project-2/workspace/` exists
- `test-project-2/agents/reviewer/` exists

**Test 3.2: Verify team.toml structure**
```bash
cat test-project-2/team.toml
```
**Expected Output**:
```toml
[team]
name = "test-project-2"
description = "Team description"
coordination = "sequential"

[agents]
[agents.main]
path = "agents/main"

[agents.reviewer]
path = "agents/reviewer"

[workspace]
shared_path = "workspace"
output_path = "workspace/results"

[coordination]
strategy = "round-robin"
max_iterations = 5
```
**Validation**: Contains all required sections and both agents

**Test 3.3: Verify SQUAD.md structure**
```bash
cat test-project-2/SQUAD.md
```
**Expected Output**:
```markdown
# Team: test-project-2

## Mission

[Describe what this team is trying to accomplish in 1-2 paragraphs.]

## Context

[Background information all agents need: project constraints, technical decisions,
external dependencies, deadlines, or any shared knowledge.]

## Agents

### main
**Role**: [Define this agent's role]
**Responsibilities**:
- [Responsibility 1]
- [Responsibility 2]

### reviewer
**Role**: [Define this agent's role]
**Responsibilities**:
- [Responsibility 1]
- [Responsibility 2]

## Coordination

[How agents should work together: handoff protocols, communication patterns,
dependency chains, blocking rules.]

## Guidelines

[Specific rules or preferences for this team: coding style, testing requirements,
commit conventions, review process.]
```
**Validation**: Contains team name, mission, context, agents, coordination, guidelines

**Test 3.4: Add third agent to existing team**
```bash
/tmp/ayo-test add-agent test-project-2 security-analyst
```
**Expected Output**:
```
Added agent 'security-analyst' to project: test-project-2
```
**Validation**:
- `test-project-2/agents/security-analyst/` exists
- `test-project-2/team.toml` updated with third agent
- No "Promoted" message (already a team)

**Test 3.5: Verify team.toml updated correctly**
```bash
grep -A 2 "security-analyst" test-project-2/team.toml
```
**Expected Output**:
```toml
[agents.security-analyst]
path = "agents/security-analyst"
```
**Validation**: Third agent properly added to team.toml

---

## Configuration Validation Tests

### Project Validation

**Test 4.1: Validate single-agent project**
```bash
cd /tmp && rm -rf test-project-3 && /tmp/ayo-test fresh test-project-3
/tmp/ayo-test checkit test-project-3
```
**Expected Output**: `Validation successful` or similar
**Validation**: Exit code 0, no errors

**Test 4.2: Validate team project**
```bash
/tmp/ayo-test add-agent test-project-3 agent2
/tmp/ayo-test checkit test-project-3
```
**Expected Output**: `Validation successful` or similar
**Validation**: Exit code 0, no errors

**Test 4.3: Validate non-existent project**
```bash
/tmp/ayo-test checkit non-existent-project 2>&1
```
**Expected Output**: Error message about project not found
**Validation**: Exit code non-zero, clear error message

**Test 4.4: Validate invalid project structure**
```bash
mkdir -p /tmp/invalid-project && echo "invalid" > /tmp/invalid-project/config.toml
/tmp/ayo-test checkit /tmp/invalid-project 2>&1
```
**Expected Output**: Validation error
**Validation**: Exit code non-zero, specific error about invalid config

---

## Error Handling Tests

### Command-Specific Errors

**Test 5.1: Capabilities command (disabled)**
```bash
/tmp/ayo-test agents capabilities 2>&1
```
**Expected Output**:
```
error: agent capabilities are no longer supported in the build system. Capabilities are now determined at build time and embedded in executables
```
**Validation**: Exit code non-zero, clear error message

**Test 5.2: Agent promotion (disabled)**
```bash
/tmp/ayo-test agents promote agent1 agent2 2>&1
```
**Expected Output**:
```
error: agent promotion is no longer supported in the build system. Use 'ayo fresh' to create new agents
```
**Validation**: Exit code non-zero, clear error message

**Test 5.3: Team chat without team.toml**
```bash
cd /tmp && rm -rf test-project-4 && /tmp/ayo-test fresh test-project-4
cd test-project-4 && /tmp/ayo-test chat . 2>&1
```
**Expected Output**:
```
error: no team.toml found in current directory. Add more agents to automatically create a team project
```
**Validation**: Exit code non-zero, helpful error message

**Test 5.4: Invalid agent name in add-agent**
```bash
/tmp/ayo-test add-agent test-project-4 "invalid/name" 2>&1
```
**Expected Output**: Error about invalid agent name
**Validation**: Exit code non-zero, clear error about naming

---

## Command Integration Tests

### Workflow Testing

**Test 6.1: Complete single-agent workflow**
```bash
cd /tmp && rm -rf workflow-1 && /tmp/ayo-test fresh workflow-1
/tmp/ayo-test checkit workflow-1
/tmp/ayo-test doctor
```
**Expected Output**: All commands succeed
**Validation**: Exit code 0 for all commands

**Test 6.2: Complete team workflow**
```bash
/tmp/ayo-test add-agent workflow-1 agent2
/tmp/ayo-test add-agent workflow-1 agent3
/tmp/ayo-test checkit workflow-1
```
**Expected Output**: All commands succeed, team structure validated
**Validation**: Exit code 0, team.toml contains 3 agents

**Test 6.3: Mixed command sequence**
```bash
/tmp/ayo-test fresh workflow-2
/tmp/ayo-test checkit workflow-2
/tmp/ayo-test add-agent workflow-2 reviewer
/tmp/ayo-test checkit workflow-2
/tmp/ayo-test add-agent workflow-2 security
/tmp/ayo-test checkit workflow-2
```
**Expected Output**: All commands succeed, validation passes at each step
**Validation**: Exit code 0 for all commands

---

## Performance Testing

### Timing Measurements

**Test 7.1: Single agent creation time**
```bash
cd /tmp && rm -rf perf-test-1
time /tmp/ayo-test fresh perf-test-1
```
**Expected Output**: < 1 second
**Validation**: Real time < 1.0s

**Test 7.2: Team promotion time**
```bash
time /tmp/ayo-test add-agent perf-test-1 agent2
```
**Expected Output**: < 0.5 seconds
**Validation**: Real time < 0.5s

**Test 7.3: Agent addition to existing team**
```bash
time /tmp/ayo-test add-agent perf-test-1 agent3
```
**Expected Output**: < 0.2 seconds
**Validation**: Real time < 0.2s

**Test 7.4: Validation time**
```bash
time /tmp/ayo-test checkit perf-test-1
```
**Expected Output**: < 0.1 seconds
**Validation**: Real time < 0.1s

---

## Edge Case Testing

### Boundary Conditions

**Test 8.1: Empty project name**
```bash
/tmp/ayo-test fresh "" 2>&1
```
**Expected Output**: Error about invalid project name
**Validation**: Exit code non-zero

**Test 8.2: Project name with special characters**
```bash
/tmp/ayo-test fresh "test-project-with-dashes" && ls | grep test-project-with-dashes
```
**Expected Output**: Project created successfully
**Validation**: Directory created with correct name

**Test 8.3: Agent name with spaces**
```bash
cd /tmp && rm -rf edge-test && /tmp/ayo-test fresh edge-test
/tmp/ayo-test add-agent edge-test "agent with spaces" 2>&1
```
**Expected Output**: Error or warning about agent name format
**Validation**: Clear error message

**Test 8.4: Adding agent to non-existent project**
```bash
/tmp/ayo-test add-agent non-existent-project agent1 2>&1
```
**Expected Output**: Error about project not found
**Validation**: Exit code non-zero, clear error

---

## Regression Testing

### Backward Compatibility

**Test 9.1: Existing single-agent projects still work**
```bash
cd /tmp && rm -rf regression-1 && /tmp/ayo-test fresh regression-1
/tmp/ayo-test checkit regression-1
```
**Expected Output**: Validation successful
**Validation**: Exit code 0

**Test 9.2: Team promotion is automatic**
```bash
/tmp/ayo-test add-agent regression-1 agent2
ls regression-1/ | grep team.toml
```
**Expected Output**: team.toml exists
**Validation**: Automatic promotion occurred

**Test 9.3: Error messages updated**
```bash
/tmp/ayo-test agents capabilities 2>&1 | grep "build system"
```
**Expected Output**: Error message mentions build system
**Validation**: No references to old framework

---

## Cleanup and Verification

### Test Environment Cleanup

**Test 10.1: Remove test projects**
```bash
cd /tmp && rm -rf test-project-* workflow-* perf-test-* edge-test regression-* invalid-project
```
**Expected Output**: No errors
**Validation**: `ls` shows no test directories

**Test 10.2: Verify no orphaned files**
```bash
ls /tmp/ | grep -E "ayo|test|project|workflow|perf|edge|regression" || echo "Clean"
```
**Expected Output**: "Clean"
**Validation**: No test files remaining

**Test 10.3: Final binary verification**
```bash
/tmp/ayo-test --version
```
**Expected Output**: Version information
**Validation**: Exit code 0

---

## Test Result Summary

### Expected Outcomes

**Passing Tests**:
- ✅ All single-agent project creation tests
- ✅ All progressive team creation tests
- ✅ All configuration validation tests
- ✅ All error handling tests
- ✅ All command integration tests
- ✅ All performance tests (within thresholds)
- ✅ All regression tests
- ✅ All cleanup tests

**Expected Failures** (documented limitations):
- ❌ Build compilation (`ayo dunn`) - Not yet implemented
- ❌ Runtime execution - Needs refactoring
- ❌ Some edge cases with special characters

### Test Coverage Metrics

**Areas Covered**:
- ✅ Project creation and structure
- ✅ Team promotion and configuration
- ✅ Configuration validation
- ✅ Error handling and user guidance
- ✅ Command integration and workflows
- ✅ Performance characteristics
- ✅ Edge cases and boundary conditions
- ✅ Regression prevention

**Areas Not Covered** (future work):
- ❌ Actual build compilation
- ❌ Runtime execution
- ❌ Cross-compilation
- ❌ Remote agent support
- ❌ Advanced coordination strategies

---

## Automated Testing Script

For convenience, here's a script to run all tests automatically:

```bash
#!/bin/bash

# Ayo E2E Testing Script
set -e

echo "=== Ayo Build System E2E Tests ==="

# Setup
cd /tmp
rm -rf ayo-e2e-test-*
echo "✓ Test environment cleaned"

# Build binary
go build -o /tmp/ayo-test ./cmd/ayo/
echo "✓ Ayo binary built"

# Test 1: Single agent creation
/tmp/ayo-test fresh ayo-e2e-test-1
test -f ayo-e2e-test-1/config.toml
test -f ayo-e2e-test-1/agents/main/config.toml
echo "✓ Single agent project creation works"

# Test 2: Team promotion
/tmp/ayo-test add-agent ayo-e2e-test-1 reviewer
test -f ayo-e2e-test-1/team.toml
test -f ayo-e2e-test-1/SQUAD.md
echo "✓ Team promotion works"

# Test 3: Validation
/tmp/ayo-test checkit ayo-e2e-test-1
echo "✓ Project validation works"

# Test 4: Error handling
/tmp/ayo-test agents capabilities 2>&1 | grep -q "build system"
echo "✓ Error handling works"

# Test 5: Multiple agents
/tmp/ayo-test add-agent ayo-e2e-test-1 agent3
grep -q "agent3" ayo-e2e-test-1/team.toml
echo "✓ Multiple agent addition works"

# Cleanup
rm -rf ayo-e2e-test-*
echo "✓ Test cleanup completed"

echo "=== All E2E Tests Passed ==="
```

Save as `run-e2e-tests.sh` and execute:
```bash
chmod +x run-e2e-tests.sh
./run-e2e-tests.sh
```

---

## Manual Verification Checklist

Use this checklist for manual testing:

- [ ] Single agent project creation works
- [ ] Team promotion triggers automatically on second agent
- [ ] team.toml contains correct structure
- [ ] SQUAD.md contains proper template
- [ ] Validation passes for both single and team projects
- [ ] Error messages are clear and helpful
- [ ] Disabled commands show appropriate errors
- [ ] Performance is within expected ranges
- [ ] Edge cases handled gracefully
- [ ] No regressions from previous functionality

---

## Troubleshooting Test Failures

### Common Issues and Solutions

**Issue**: `command not found: /tmp/ayo-test`
**Solution**: Rebuild binary with `go build -o /tmp/ayo-test ./cmd/ayo/`

**Issue**: Permission denied errors
**Solution**: Run `chmod +x /tmp/ayo-test`

**Issue**: Tests failing with database errors
**Solution**: Ensure all framework-specific commands are disabled

**Issue**: Team promotion not triggering
**Solution**: Verify you're adding the second agent to a single-agent project

**Issue**: Validation errors on clean projects
**Solution**: Check TOML syntax and file permissions

---

## Test Environment Requirements

### Minimum Requirements

- Go 1.20+
- Git 2.0+
- Unix-like operating system (Linux, macOS)
- 1GB free disk space
- 512MB RAM

### Recommended Requirements

- Go 1.21+
- Git 2.30+
- Linux/macOS with latest updates
- SSD storage
- 2GB RAM

---

## Continuous Integration Setup

### GitHub Actions Example

```yaml
name: Ayo E2E Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Build
      run: go build -o /tmp/ayo-test ./cmd/ayo/
    - name: Run E2E Tests
      run: |
        cd /tmp
        /tmp/ayo-test fresh e2e-test-1
        /tmp/ayo-test add-agent e2e-test-1 reviewer
        /tmp/ayo-test checkit e2e-test-1
        /tmp/ayo-test agents capabilities 2>&1 | grep -q "build system"
        rm -rf e2e-test-1
```

### GitLab CI Example

```yaml
stages:
  - build
  - test

build:
  stage: build
  script:
    - go build -o /tmp/ayo-test ./cmd/ayo/
  artifacts:
    paths:
      - /tmp/ayo-test

test:
  stage: test
  script:
    - cd /tmp
    - /tmp/ayo-test fresh gitlab-test-1
    - /tmp/ayo-test add-agent gitlab-test-1 reviewer
    - /tmp/ayo-test checkit gitlab-test-1
    - rm -rf gitlab-test-1
```

---

## Test Data Generation

For automated testing, use this script to generate test data:

```bash
#!/bin/bash

# Generate test projects for various scenarios

echo "Generating test data..."

# Single agent project
/tmp/ayo-test fresh /tmp/test-data/single-agent

# Small team (2 agents)
/tmp/ayo-test fresh /tmp/test-data/small-team
/tmp/ayo-test add-agent /tmp/test-data/small-team reviewer

# Medium team (5 agents)
/tmp/ayo-test fresh /tmp/test-data/medium-team
for i in 2 3 4 5; do
  /tmp/ayo-test add-agent /tmp/test-data/medium-team "agent$i"
done

# Invalid projects for error testing
mkdir -p /tmp/test-data/invalid-no-config
echo "not toml" > /tmp/test-data/invalid-no-config/config.txt

mkdir -p /tmp/test-data/invalid-toml
echo "invalid [ toml" > /tmp/test-data/invalid-toml/config.toml

echo "Test data generated in /tmp/test-data/"
```

---

## Performance Benchmarking

Use these commands to benchmark performance:

```bash
# Single agent creation (run 10 times)
echo "Single agent creation:"
for i in {1..10}; do
  rm -rf /tmp/bench-single-$i
  time /tmp/ayo-test fresh /tmp/bench-single-$i
done

# Team promotion (run 10 times)
echo "Team promotion:"
for i in {1..10}; do
  rm -rf /tmp/bench-team-$i
  /tmp/ayo-test fresh /tmp/bench-team-$i
  time /tmp/ayo-test add-agent /tmp/bench-team-$i agent2
done

# Validation (run 10 times)
echo "Project validation:"
for i in {1..10}; do
  time /tmp/ayo-test checkit /tmp/bench-team-$i
done
```

---

## Security Testing

Basic security checks:

```bash
# Check file permissions
echo "File permissions:"
ls -la /tmp/ayo-test
find /tmp/test-* -type f -exec ls -la {} \; | head -10

# Check for sensitive data in configs
echo "Checking for secrets in configs:"
grep -r "password\|secret\|api_key" /tmp/test-* || echo "No secrets found"

# Verify no network calls
echo "Network activity check:"
strace -e trace=network /tmp/ayo-test --help 2>&1 | grep -i "socket\|connect" || echo "No network activity"
```

---

## Compliance Testing

Check compliance with standards:

```bash
# TOML validation
echo "TOML validation:"
find /tmp/test-* -name "*.toml" -exec sh -c 'echo "Checking $1"; go run github.com/pelletier/go-toml/v2/cmd/tomllint@latest lint "$1"' sh {} \;

# Configuration schema validation
echo "Schema validation:"
/tmp/ayo-test checkit /tmp/test-* --strict

# Documentation completeness
echo "Documentation check:"
test -f docs/REFERENCE.md && echo "✓ Reference documentation exists"
test -f .docs/E2E_TESTING.md && echo "✓ Testing documentation exists"
```

---

## Test Reporting

Generate a test report:

```bash
#!/bin/bash

echo "=== Ayo Build System Test Report ==="
echo "Generated: $(date)"
echo

echo "1. ENVIRONMENT"
echo "Go version: $(go version)"
echo "OS: $(uname -a)"
echo "Ayo binary: $(ls -la /tmp/ayo-test | awk '{print $5}') bytes"
echo

echo "2. CORE FUNCTIONALITY"
cd /tmp
rm -rf report-test-*

# Test single agent
/tmp/ayo-test fresh report-test-1 2>&1 > /dev/null && echo "✓ Single agent creation" || echo "✗ Single agent creation"

# Test team promotion
/tmp/ayo-test add-agent report-test-1 agent2 2>&1 > /dev/null && echo "✓ Team promotion" || echo "✗ Team promotion"

# Test validation
/tmp/ayo-test checkit report-test-1 2>&1 > /dev/null && echo "✓ Project validation" || echo "✗ Project validation"

# Test error handling
/tmp/ayo-test agents capabilities 2>&1 | grep -q "build system" && echo "✓ Error handling" || echo "✗ Error handling"

echo

echo "3. PERFORMANCE"
echo "Single agent: $( { time /tmp/ayo-test fresh report-test-2 2>&1 > /dev/null; } 2>&1 | grep real )"
echo "Team promotion: $( { time /tmp/ayo-test add-agent report-test-2 agent2 2>&1 > /dev/null; } 2>&1 | grep real )"
echo "Validation: $( { time /tmp/ayo-test checkit report-test-2 2>&1 > /dev/null; } 2>&1 | grep real )"

echo

echo "4. FILE STRUCTURE"
echo "Single agent files: $(find report-test-1 -type f | wc -l)"
echo "Team files: $(find report-test-2 -type f | wc -l)"

rm -rf report-test-*
echo
echo "=== Report Complete ==="
```

---

## Conclusion

This comprehensive E2E testing guide provides atomic-level testing instructions for every aspect of the Ayo build system. Each test is designed to be:

- **Reproducible**: Exact commands with expected outputs
- **Isolated**: Independent tests that don't interfere
- **Comprehensive**: Covers all functionality and edge cases
- **Automatable**: Can be scripted for CI/CD integration
- **Documented**: Clear validation criteria and troubleshooting

Use this guide to:
1. Verify new implementations
2. Test bug fixes
3. Validate refactoring
4. Ensure backward compatibility
5. Performance benchmarking
6. Security auditing
7. Compliance checking

The tests cover the complete lifecycle from project creation to team coordination, ensuring the build system works as designed and meets all requirements.
