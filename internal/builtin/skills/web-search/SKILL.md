---
name: web-search
description: Search the web for live, verifiable information using SearXNG via curl. Use when the user needs current data, facts, news, or information that may have changed since your training cutoff.
compatibility: Requires curl, internet access, and SEARXNG_ENDPOINT environment variable
metadata:
  author: ayo
  version: "1.0"
---

# Web Search Skill

Search the web for live, verifiable information using SearXNG—a privacy-respecting metasearch engine. This skill uses `curl` to query SearXNG's JSON API.

## Required Configuration

**The `SEARXNG_ENDPOINT` environment variable MUST be set to use this skill.**

Before performing any search, check that it's configured:

```bash
echo "${SEARXNG_ENDPOINT:-NOT_SET}"
```

- **If output is `NOT_SET`**: Stop and inform the user they need to set `SEARXNG_ENDPOINT`
- **If output is a hostname**: Proceed with the search using that endpoint

Example user configuration:
```bash
export SEARXNG_ENDPOINT=search.example.com
```

## When to Use

SearXNG is a privacy-respecting metasearch engine that aggregates results from multiple sources. Results may be delayed by several days compared to direct Google searches.

Activate this skill when:
- User needs to verify facts, claims, or research topics
- User asks about events or news older than a few days
- User wants to find documentation, articles, or references
- User asks "find information about..." or "search for..."
- User needs multiple perspectives on a topic
- Information needs to be cited with sources
- User asks about something that may have changed since your training

**Less suitable for:**
- Breaking news from the last 24-48 hours (results may be delayed)
- Real-time information (stock prices, live scores, etc.)
- Rapidly developing situations in progress

## Performing Searches

### Step-by-Step Process

1. **Verify endpoint is configured**:
   ```bash
   echo "${SEARXNG_ENDPOINT:-NOT_SET}"
   ```
   If `NOT_SET`, tell the user to configure `SEARXNG_ENDPOINT` and stop.

2. **Execute the search**:
   ```bash
   curl -s "https://${SEARXNG_ENDPOINT}/search?q=YOUR+QUERY&format=json" | head -c 50000
   ```

### Search Parameters

| Parameter | Values | Example |
|-----------|--------|---------|
| `q` | Search query (required) | `q=golang+error+handling` |
| `format` | `json` (required for API) | `format=json` |
| `categories` | `general`, `news`, `images`, `videos`, `science`, `it`, `files`, `social media`, `music`, `map` | `categories=news` |
| `time_range` | `day`, `week`, `month`, `year` | `time_range=week` |
| `language` | Language code | `language=en` |
| `engines` | Specific engines | `engines=google,duckduckgo` |
| `pageno` | Page number | `pageno=2` |

### Example Searches

**General search:**
```bash
curl -s "https://${SEARXNG_ENDPOINT}/search?q=kubernetes+pod+networking&format=json" | head -c 50000
```

**News from the past week:**
```bash
curl -s "https://${SEARXNG_ENDPOINT}/search?q=AI+regulation&format=json&categories=news&time_range=week" | head -c 50000
```

**Technical search with specific engines:**
```bash
curl -s "https://${SEARXNG_ENDPOINT}/search?q=react+hooks+performance&format=json&categories=it&engines=stackoverflow,github" | head -c 50000
```

## Response Format

The JSON response contains:

```json
{
  "query": "your search query",
  "number_of_results": 12345,
  "results": [
    {
      "url": "https://example.com/page",
      "title": "Page Title",
      "content": "Snippet or description text",
      "engine": "google",
      "publishedDate": "2024-01-15T10:30:00Z"
    }
  ],
  "answers": ["Direct answers if available"],
  "suggestions": ["related", "search", "queries"],
  "infoboxes": [
    {
      "infobox": "Topic Name",
      "content": "Summary information"
    }
  ]
}
```

### Key Fields

| Field | Description |
|-------|-------------|
| `results` | Array of search results with url, title, content, engine |
| `answers` | Direct answers (e.g., calculations, simple facts) |
| `suggestions` | Related search queries to explore |
| `infoboxes` | Knowledge panel-style summaries (Wikipedia, etc.) |

## Search Strategies

### For Current Events/News

```bash
curl -s "https://${SEARXNG_ENDPOINT}/search?q=TOPIC&format=json&categories=news&time_range=day" | head -c 50000
```

### For Technical Topics

```bash
curl -s "https://${SEARXNG_ENDPOINT}/search?q=TOPIC&format=json&categories=it&engines=stackoverflow,github,google" | head -c 50000
```

### For Fact Verification

Run multiple searches with different phrasings:
1. Search the exact claim
2. Search for counterarguments
3. Search for primary sources

## Parsing Results with jq

For more precise extraction:

```bash
# Extract just titles and URLs
curl -s "https://${SEARXNG_ENDPOINT}/search?q=QUERY&format=json" | jq -r '.results[:10] | .[] | "\(.title)\n  \(.url)\n"'

# Get suggestions
curl -s "https://${SEARXNG_ENDPOINT}/search?q=QUERY&format=json" | jq -r '.suggestions[]'
```

## Error Handling

| Error | Cause | Solution |
|-------|-------|----------|
| `NOT_SET` from endpoint check | `SEARXNG_ENDPOINT` not configured | User must set the environment variable |
| HTML response | JSON API disabled on instance | User must enable JSON API on their SearXNG instance |
| Empty results | Query too specific | Broaden search terms |
| Connection refused | Instance offline or misconfigured | User must check their SearXNG instance |

## Best Practices

1. **Always verify endpoint first** - Check `SEARXNG_ENDPOINT` before any search
2. **Be specific in queries** - Include relevant context and keywords
3. **Use appropriate categories** - `news` for current events, `it` for technical topics
4. **Check multiple sources** - Don't rely on a single result
5. **Note publication dates** - Prioritize recent sources for time-sensitive topics
6. **Cite your sources** - Include URLs when presenting information
7. **Limit output** - Always pipe through `head -c 50000`

## Output Format

When presenting search results to users:

```markdown
Based on my web search:

**Summary**: [Your synthesis of the findings]

**Key Sources**:
- [Title](URL) - [Brief note on relevance]
- [Title](URL) - [Brief note on relevance]

**Note**: [Any caveats about the information]
```

When endpoint is not configured:

```markdown
Web search is not available. Please set the `SEARXNG_ENDPOINT` environment variable to your SearXNG instance:

export SEARXNG_ENDPOINT=search.example.com
```
