---
id: ase-fb0m
status: closed
deps: []
links: []
created: 2026-02-06T04:09:46Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ase-ka3q
---
# Switch from busybox to Alpine base image

Replace the busybox base image with Alpine Linux for the sandbox. Alpine provides a real package manager (apk), musl libc, and better compatibility.

## Design

## Current State
The sandbox currently uses 'docker.io/library/busybox:stable' as the base image.

## Changes
1. Update default image in internal/sandbox/apple.go from busybox to Alpine
2. Update internal/sandbox/linux.go similarly
3. Update documentation in internal/sandbox/images/busybox.md to alpine.md
4. Update any tests that reference busybox

## Alpine Benefits
- apk package manager for installing tools
- More complete userspace
- Better compatibility with standard Linux tools
- ngircd available as package

## Acceptance Criteria

- Default image is alpine:latest or alpine:3.19
- Existing functionality unchanged
- Tests pass with new image

