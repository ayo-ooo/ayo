---
id: smft-fnex
status: closed
deps: [smft-ti1b]
links: []
created: 2026-02-12T23:54:43Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Replace placeholder Introduction section

Replace the placeholder Introduction section in index.html.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html

**Find (lines 26-39):**
The current Introduction section with placeholder content.

**Replace with:**
```html
<section data-tab="Introduction">
    <article>
        <h2>What This Manual Teaches</h2>
        <p>Ayo lets you create AI agents that can execute commands, remember context, and work together. This manual starts simple and builds up.</p>
        <p>By the end, you'll know how to create custom agents, teach them skills, automate workflows, and prototype agent teams.</p>
        <p>Scroll through the panels. Each fold is a self-contained lesson.</p>
    </article>
    <aside>
        <h3>Navigation</h3>
        <p>↑↓ Scroll or arrow keys</p>
        <p>Space: next panel</p>
        <p>Shift+Space: previous</p>
    </aside>
</section>
```

