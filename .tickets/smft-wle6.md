---
id: smft-wle6
status: closed
deps: [smft-53x5]
links: []
created: 2026-02-12T23:55:33Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Replace Ch2 Note callout with Agent Directory figure

Replace the Note blockquote with an Agent Directory structure.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html

**Find (lines 123-128):**
```html
<section>
    <blockquote data-type="note">
        <p>Nodes are the heart of Ayo...</p>
    </blockquote>
</section>
```

**Replace with:**
```html
<section>
    <h2>Agent Directory Structure</h2>
    <pre><code>@greeter/
├── config.json    # Settings
└── system.md      # Instructions</code></pre>
</section>
```

