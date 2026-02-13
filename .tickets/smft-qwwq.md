---
id: smft-qwwq
status: closed
deps: [smft-hroz]
links: []
created: 2026-02-12T23:55:02Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Replace Ch1 Features panel with First Run

Replace the placeholder Features panel with First Run content.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html

**Find (lines 60-71):**
```html
<section>
    <h2>Key Features</h2>
    <ul>
        <li>Intuitive drag-and-drop interface</li>
        ...
    </ul>
</section>
```

**Replace with:**
```html
<section>
    <article>
        <h2>First Run</h2>
        <p>Open your terminal and type:</p>
        <pre data-copy><code>ayo</code></pre>
        <p>That's it. You're now talking to an AI agent. Ask it anything.</p>
        <p>Try: <code>ayo "what can you do?"</code></p>
    </article>
    <aside>
        <h3>What just happened?</h3>
        <p>Ayo connected to an LLM, set up a conversation, and gave you a chat interface.</p>
    </aside>
</section>
```

