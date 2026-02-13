---
id: smft-ruq7
status: closed
deps: [smft-vi2p]
links: []
created: 2026-02-12T23:55:15Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Replace Interface panel with Philosophy callout

Replace the placeholder Interface panel with a philosophy callout.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html

**Find (lines 87-98):**
```html
<section>
    <div>
        <h2>Understanding the Interface</h2>
        ...
    </div>
    <figure>
        ...
    </figure>
</section>
```

**Replace with:**
```html
<section>
    <blockquote data-type="note">
        <p>Ayo extends Unix philosophy to AI: simple tools that compose, text as interface, files as configuration.</p>
    </blockquote>
</section>
```

