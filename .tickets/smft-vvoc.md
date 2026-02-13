---
id: smft-vvoc
status: closed
deps: [smft-olgi]
links: []
created: 2026-02-12T23:57:31Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Chapter 9 - Flows & Automation title and panels

Add Chapter 9: Flows & Automation section.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html
**Insert after:** Chapter 8 panels

**Content to add:**

1. Chapter title:
```html
<section data-width="narrow" data-crease="chapter" data-tab="Flows">
    <hgroup>
        <h2>Chapter 9</h2>
        <p>Flows & Automation</p>
    </hgroup>
</section>
```

2. What Flows Provide:
```html
<section>
    <article>
        <h2>Multi-Step Workflows</h2>
        <p>Flows combine multiple agents and shell commands into repeatable workflows. Two types: shell scripts and YAML definitions.</p>
    </article>
    <aside>
        <h3>Beyond Chaining</h3>
        <p>Flows add parallel steps, conditions, and triggers.</p>
    </aside>
</section>
```

3. Shell Flow example:
```html
<section>
    <h2>Shell Flows</h2>
    <pre data-copy><code>#!/usr/bin/env bash
# ayo:flow
# name: daily-summary

git log --oneline --since="1 day ago" | \
  ayo @ayo "summarize these commits"</code></pre>
    <samp>ayo flows run daily-summary</samp>
</section>
```

4. Triggers:
```html
<section>
    <h2>Triggers</h2>
    <pre data-copy><code># Run at 9 AM daily
ayo triggers add --cron "0 9 * * *" \
  --agent @ayo --prompt "daily standup"

# Watch for file changes  
ayo triggers add --watch ./src \
  --agent @ayo --prompt "run tests"</code></pre>
</section>
```

