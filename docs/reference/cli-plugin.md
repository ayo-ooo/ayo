# ayo plugin

Manage plugins to extend ayo with community packages.

## Synopsis

```
ayo plugin <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List installed plugins |
| `install` | Install a plugin |
| `update` | Update a plugin |
| `remove` | Remove a plugin |
| `search` | Search for plugins |

---

## ayo plugin list

List all installed plugins.

```bash
$ ayo plugin list
NAME           VERSION  TYPE     DESCRIPTION
github-tools   1.2.0    tools    GitHub API tools
slack-notify   1.0.0    tools    Slack notifications
custom-model   2.0.0    provider Custom model provider
```

---

## ayo plugin install

Install a plugin from the registry or URL.

```bash
$ ayo plugin install <name|url> [--version <v>]
```

### Examples

```bash
$ ayo plugin install github-tools
Installed github-tools@1.2.0

$ ayo plugin install https://example.com/my-plugin.tar.gz
Installed my-plugin@1.0.0
```

---

## ayo plugin update

Update a plugin to the latest version.

```bash
$ ayo plugin update <name> [--version <v>]
```

### Example

```bash
$ ayo plugin update github-tools
Updated github-tools: 1.2.0 → 1.3.0
```

---

## ayo plugin remove

Remove an installed plugin.

```bash
$ ayo plugin remove <name>
```

---

## ayo plugin search

Search for plugins in the registry.

```bash
$ ayo plugin search <query>
```

### Example

```bash
$ ayo plugin search "notification"
NAME           VERSION  DESCRIPTION
slack-notify   1.0.0    Slack notifications
discord-bot    1.1.0    Discord integration
email-alerts   2.0.0    Email notifications
```

---

## Plugin Types

| Type | Description |
|------|-------------|
| `tools` | Additional tools for agents |
| `provider` | Custom LLM providers |
| `skill` | Bundled skills |
| `flow` | Reusable flows |

## See Also

- [Plugins Guide](../plugins.md) - Creating and using plugins
