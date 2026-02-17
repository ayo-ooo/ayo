# ayo skill

Manage skills—reusable instruction modules that extend agent capabilities.

## Synopsis

```
ayo skill <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List available skills |
| `show` | Show skill details |
| `search` | Search for skills |
| `install` | Install a skill |
| `create` | Create a new skill |

---

## ayo skill list

List all available skills.

```bash
$ ayo skill list
NAME           VERSION  DESCRIPTION
debugging      1.2.0    Systematic debugging techniques
code-review    1.0.0    Code review best practices
security       2.1.0    Security analysis guidelines
testing        1.1.0    Test writing strategies
```

---

## ayo skill show

Show detailed skill information.

```bash
$ ayo skill show <name>
```

### Example

```bash
$ ayo skill show debugging
Name:        debugging
Version:     1.2.0
Author:      ayo-community
Description: Systematic debugging techniques

Instructions:
  When debugging an issue:
  1. Reproduce the problem consistently
  2. Isolate the failing component
  3. Form and test hypotheses
  ...
```

---

## ayo skill search

Search for skills by keyword.

```bash
$ ayo skill search <query>
```

### Example

```bash
$ ayo skill search "security"
NAME        DESCRIPTION
security    Security analysis guidelines
auth        Authentication best practices
crypto      Cryptography guidelines
```

---

## ayo skill install

Install a skill from the registry.

```bash
$ ayo skill install <name> [--version <v>]
```

### Example

```bash
$ ayo skill install debugging
Installed debugging@1.2.0
```

---

## ayo skill create

Create a new skill.

```bash
$ ayo skill create <name>
```

Creates a skill template in `~/.config/ayo/skills/<name>/`.

---

## Skill File Format

Skills follow the [agentskills.org](https://agentskills.org) specification:

```markdown
---
name: debugging
version: 1.2.0
description: Systematic debugging techniques
author: ayo-community
tags: [development, debugging]
---
# Debugging

When debugging an issue, follow this systematic approach:

## Reproduce

First, reproduce the problem consistently...

## Isolate

Narrow down the failing component...

## Hypothesize

Form hypotheses about the root cause...
```

## See Also

- [Skills Guide](../skills.md) - Conceptual overview
- [ayo agents](cli-agents.md) - Adding skills to agents
