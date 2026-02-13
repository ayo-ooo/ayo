---
id: smft-4nxx
status: closed
deps: [smft-er0g]
links: []
created: 2026-02-12T23:56:41Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Chapter 4 - Agent Conventions panels

Add Chapter 4 content panels after the Ch4 title panel.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html
**Insert after:** Ch4 title panel (Agent Conventions)

**Panels to add:**

1. The @ Prefix panel:
```html
<section>
    <article>
        <h2>The @ Prefix</h2>
        <p>Agent names start with @ by convention. It's not syntax—just a visual marker that makes agents easy to spot.</p>
        <p><code>@ayo</code>, <code>@reviewer</code>, <code>@researcher</code></p>
    </article>
    <aside>
        <h3>Built-in</h3>
        <p>Ayo ships with one agent: <code>@ayo</code></p>
    </aside>
</section>
```

2. Agent Directories panel:
```html
<section>
    <h2>Where Agents Live</h2>
    <dl>
        <dt>~/.config/ayo/agents/</dt>
        <dd>Your custom agents</dd>
        <dt>~/.local/share/ayo/agents/</dt>
        <dd>Built-in agents</dd>
    </dl>
</section>
```

3. config.json fields table:
```html
<section data-table="striped">
    <h2>config.json Fields</h2>
    <table>
        <thead>
            <tr><th>Field</th><th>Purpose</th></tr>
        </thead>
        <tbody>
            <tr><td><code>description</code></td><td>Brief agent description</td></tr>
            <tr><td><code>model</code></td><td>LLM to use</td></tr>
            <tr><td><code>allowed_tools</code></td><td>Tools agent can call</td></tr>
            <tr><td><code>skills</code></td><td>Skills to load</td></tr>
            <tr><td><code>guardrails</code></td><td>Safety constraints (default: true)</td></tr>
        </tbody>
    </table>
</section>
```

