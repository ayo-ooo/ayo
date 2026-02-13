---
id: smft-53x5
status: closed
deps: [smft-zr6y]
links: []
created: 2026-02-12T23:55:27Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Replace Ch2 Terminology with What is an Agent

Replace the Terminology definition list with What is an Agent.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html

**Find (lines 108-121):**
```html
<section>
    <h2>Terminology</h2>
    <dl>
        <dt>Node</dt>
        ...
    </dl>
</section>
```

**Replace with:**
```html
<section>
    <h2>What is an Agent?</h2>
    <dl>
        <dt>Agent</dt>
        <dd>A directory containing configuration files that define an AI assistant's behavior.</dd>
        <dt>config.json</dt>
        <dd>Settings: which model to use, which tools to allow, which skills to load.</dd>
        <dt>system.md</dt>
        <dd>Instructions: the agent's personality, expertise, and guidelines.</dd>
    </dl>
</section>
```

