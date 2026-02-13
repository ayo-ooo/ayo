---
id: smft-9dox
status: closed
deps: [smft-4nxx]
links: []
created: 2026-02-12T23:56:53Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Chapter 5 - Tools title and panels

Add Chapter 5: Tools section.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html
**Insert after:** Chapter 4 panels

**Content to add:**

1. Chapter title:
```html
<section data-width="narrow" data-crease="chapter" data-tab="Tools">
    <hgroup>
        <h2>Chapter 5</h2>
        <p>Tools & Capabilities</p>
    </hgroup>
</section>
```

2. What Tools Provide:
```html
<section>
    <article>
        <h2>What Tools Provide</h2>
        <p>An agent without tools can only talk. Tools let agents act: execute commands, remember things, delegate to other agents.</p>
    </article>
    <aside>
        <h3>Principle</h3>
        <p>Agents get only the tools they need. Least privilege.</p>
    </aside>
</section>
```

3. Built-in Tools table:
```html
<section data-table="striped">
    <h2>Built-in Tools</h2>
    <table>
        <thead><tr><th>Tool</th><th>What it does</th></tr></thead>
        <tbody>
            <tr><td><code>bash</code></td><td>Execute shell commands</td></tr>
            <tr><td><code>todo</code></td><td>Track multi-step tasks</td></tr>
            <tr><td><code>memory</code></td><td>Store/retrieve knowledge</td></tr>
            <tr><td><code>agent_call</code></td><td>Delegate to other agents</td></tr>
        </tbody>
    </table>
</section>
```

4. bash tool detail:
```html
<section>
    <h2>The bash Tool</h2>
    <p>Most important tool. Runs commands in the sandbox (isolated from your system).</p>
    <pre><code>Tool: bash
Command: go test ./...

Output:
PASS
ok  mypackage 0.003s</code></pre>
</section>
```

