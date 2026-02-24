---
id: ayo-prss
status: open
deps: [ayo-pltg]
links: []
created: 2026-02-23T12:00:00Z
type: epic
priority: 2
assignee: Alex Cabrera
tags: [plugins, triggers, external-repo, open-standards]
---
# Epic: RSS/Atom Feed Trigger Plugin

## Summary

Create `ayo-plugins-rss` - a trigger plugin for RSS and Atom feed monitoring. RSS (Really Simple Syndication) and Atom are open standards for web feed syndication, enabling agents to react to new content from any feed source.

## Open Standards Focus

- **RSS 2.0** - Most widely used feed format
- **Atom 1.0 (RFC 4287)** - Modern alternative to RSS
- **JSON Feed** - Modern JSON-based feed format
- Universal compatibility with millions of feeds

## Use Cases

1. **News Monitoring** - Track industry news, trigger summaries
2. **Blog Aggregation** - Collect and summarize blog posts
3. **Release Tracking** - Monitor GitHub releases, package updates
4. **Security Alerts** - Track CVE feeds, security advisories
5. **Content Curation** - Aggregate content for newsletters
6. **Price Monitoring** - Track deal/price feeds
7. **Research Monitoring** - Track arxiv, academic feeds

## Plugin Components

```
ayo-plugins-rss/
├── manifest.json
├── triggers/
│   └── rss/
│       ├── trigger.json
│       └── rss-trigger       # Binary
├── tools/
│   ├── rss-fetch/
│   ├── rss-list/
│   └── opml-import/
├── agents/
│   └── @feed-reader/
└── skills/
    ├── news-summary.md
    └── content-curation.md
```

## Trigger Specification

### Configuration Schema

```json
{
  "type": "rss",
  "config": {
    "feeds": {
      "type": "array",
      "required": true,
      "items": {
        "type": "object",
        "properties": {
          "url": { "type": "string", "required": true },
          "name": { "type": "string" },
          "tags": { "type": "array", "items": { "type": "string" } }
        }
      }
    },
    "opml_file": {
      "type": "string",
      "description": "Path to OPML file with feed subscriptions"
    },
    "poll_interval": {
      "type": "duration",
      "default": "15m"
    },
    "dedupe_window": {
      "type": "duration",
      "default": "7d",
      "description": "How long to remember seen items"
    },
    "filter": {
      "type": "object",
      "properties": {
        "title_contains": { "type": "string" },
        "title_regex": { "type": "string" },
        "author": { "type": "string" },
        "min_age": { "type": "duration" },
        "max_age": { "type": "duration" }
      }
    },
    "batch": {
      "type": "boolean",
      "default": false,
      "description": "If true, batch multiple items into single trigger"
    },
    "batch_size": {
      "type": "integer",
      "default": 10
    }
  }
}
```

### Event Payload (Single Item)

```json
{
  "event_type": "rss.new_item",
  "feed": {
    "url": "https://example.com/feed.xml",
    "name": "Example Blog",
    "tags": ["tech", "news"]
  },
  "item": {
    "id": "https://example.com/post/123",
    "title": "New Feature Released",
    "link": "https://example.com/post/123",
    "description": "We're excited to announce...",
    "content": "Full HTML content...",
    "author": "John Doe",
    "published": "2026-02-23T10:00:00Z",
    "updated": "2026-02-23T10:00:00Z",
    "categories": ["release", "announcement"],
    "enclosures": [
      {
        "url": "https://example.com/podcast.mp3",
        "type": "audio/mpeg",
        "length": 12345678
      }
    ]
  }
}
```

### Event Payload (Batched)

```json
{
  "event_type": "rss.batch",
  "items": [
    { "feed": {...}, "item": {...} },
    { "feed": {...}, "item": {...} }
  ],
  "count": 5
}
```

## Tool Specifications

### rss-fetch

```json
{
  "name": "rss-fetch",
  "description": "Fetch items from an RSS/Atom feed",
  "parameters": {
    "url": { "type": "string", "required": true },
    "limit": { "type": "integer", "default": 10 },
    "since": { "type": "string", "description": "ISO 8601 datetime" },
    "include_content": { "type": "boolean", "default": true }
  }
}
```

### rss-list

```json
{
  "name": "rss-list",
  "description": "List configured feed subscriptions",
  "parameters": {
    "tags": { "type": "array", "items": { "type": "string" } }
  }
}
```

### opml-import

```json
{
  "name": "opml-import",
  "description": "Import feed subscriptions from OPML file",
  "parameters": {
    "file": { "type": "string", "required": true },
    "tags": { "type": "array", "items": { "type": "string" } }
  }
}
```

## Agent: @feed-reader

```markdown
# @feed-reader

You monitor and analyze RSS/Atom feeds for relevant content.

## Capabilities

- Fetch and parse RSS/Atom/JSON feeds
- Summarize new content
- Identify trending topics across feeds
- Filter by relevance to user interests
- Generate daily/weekly digests

## Guidelines

1. Respect feed update frequencies
2. Deduplicate content across feeds
3. Extract key information efficiently
4. Link back to original sources
5. Identify and flag potential spam/low-quality content
```

## Implementation Steps

1. [ ] Create repository `ayo-plugins-rss`
2. [ ] Implement feed parser (RSS 2.0, Atom 1.0, JSON Feed)
3. [ ] Implement OPML import/export
4. [ ] Create polling trigger with deduplication
5. [ ] Implement state persistence (seen items)
6. [ ] Implement rss-fetch tool
7. [ ] Implement rss-list tool
8. [ ] Implement opml-import tool
9. [ ] Create @feed-reader agent
10. [ ] Create news-summary skill
11. [ ] Create content-curation skill
12. [ ] Write documentation
13. [ ] Add tests with sample feeds

## Dependencies

- Depends on: `ayo-pltg` (trigger plugin architecture)
- Go libraries:
  - `github.com/mmcdole/gofeed` - Universal feed parser
  - Built-in `encoding/xml` for OPML

## Feed Discovery

Optional enhancement: Auto-discover feeds from URLs:
- Look for `<link rel="alternate" type="application/rss+xml">`
- Check common paths: `/feed`, `/rss`, `/atom.xml`, `/feed.xml`

---

*Created: 2026-02-23*
