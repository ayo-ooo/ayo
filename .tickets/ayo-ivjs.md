---
id: ayo-ivjs
status: closed
deps: [ayo-fpxg]
links: []
created: 2026-02-06T22:14:34Z
type: task
priority: 1
assignee: Alex Cabrera
tags: [testing, manual-test]
---
# Test Section 2: Pre-Flight Verification

Verify system requirements and ayo doctor checks pass.

## Scope
- Run debug/system-info.sh and verify output
- Run ayo doctor and analyze results
- Run ayo doctor -v for verbose diagnostics

## Commands
```bash
./debug/system-info.sh
ayo doctor
ayo doctor -v
```

## Verification
- [ ] OS detected correctly (darwin or linux)
- [ ] Docker shows as running
- [ ] ayo version displayed
- [ ] Config paths shown
- [ ] Doctor shows Ayo Version
- [ ] Doctor shows Paths section
- [ ] Doctor shows Ollama section (if applicable)
- [ ] Doctor shows Database section
- [ ] Doctor shows Configuration section
- [ ] Doctor shows Sandbox provider
- [ ] Verbose mode shows additional sandbox test

## Analysis Required
- Document any FAIL status items
- Document any unexpected WARN status items
- Record sandbox provider type (apple-container or linux)
- Record ayo version for test report

## Cleanup
None

## Exit Criteria
System requirements verified, all critical checks pass


## Notes

**2026-02-06T22:17:37Z**

PASSED: System info collected (darwin/arm64, Docker 29.2.1 running). Doctor shows all OK except config file WARN (using defaults). Sandbox provider: apple-container. Verbose mode shows sandbox test passed (create, exec, cleanup).
