---
id: smft-cd0y
status: closed
deps: [smft-mk9a]
links: []
created: 2026-02-12T23:56:10Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Replace Configuration Example with Sessions example

Replace Configuration Example with Sessions usage.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html

**Find (lines 200-214):**
```html
<section>
    <h2>Configuration Example</h2>
    <pre data-copy><code>{
  "project": "my-flow",
  ...
}</code></pre>
    <samp>Configuration loaded successfully.</samp>
</section>
```

**Replace with:**
```html
<section>
    <article>
        <h2>Continuing Conversations</h2>
        <p>Every conversation is saved as a session. Pick up where you left off:</p>
        <pre data-copy><code># Continue last session
ayo -c "what about the edge cases?"

# List all sessions
ayo sessions list

# Continue specific session
ayo -s abc123 "one more thing"</code></pre>
    </article>
    <aside>
        <h3>Session Storage</h3>
        <p>Sessions are stored in SQLite at <code>~/.local/share/ayo/ayo.db</code></p>
    </aside>
</section>
```

