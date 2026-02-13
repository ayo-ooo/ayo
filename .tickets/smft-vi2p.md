---
id: smft-vi2p
status: closed
deps: [smft-qwwq]
links: []
created: 2026-02-12T23:55:09Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Replace Ch1 Workspace panel with Your First Prompt

Replace the placeholder Workspace panel with Your First Prompt content.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html

**Find (lines 73-85):**
```html
<section>
    <article>
        <h2>The Workspace</h2>
        ...
    </article>
    <aside>
        ...
    </aside>
</section>
```

**Replace with:**
```html
<section>
    <h2>Your First Prompt</h2>
    <pre data-copy><code># Single task
ayo "help me debug this test"

# With a file
ayo -a main.go "review this code"

# Continue the conversation
ayo -c "what about edge cases?"</code></pre>
</section>
```

