---
id: smft-7h2s
status: closed
deps: [smft-vvoc]
links: []
created: 2026-02-12T23:57:41Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Chapter 10 - Sandboxes title and panels

Add Chapter 10: Sandboxes (Advanced) section.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html
**Insert after:** Chapter 9 panels

**Content to add:**

1. Chapter title:
```html
<section data-width="narrow" data-crease="chapter" data-tab="Sandboxes">
    <hgroup>
        <h2>Chapter 10</h2>
        <p>Sandboxes</p>
    </hgroup>
</section>
```

2. Why Isolation:
```html
<section>
    <article>
        <h2>Why Isolation Matters</h2>
        <p>When agents run commands, they have real power. Sandboxes contain that power: isolated filesystem, controlled network, resource limits.</p>
        <p>Even if an agent misbehaves, damage is contained.</p>
    </article>
    <aside>
        <h3>Not Docker</h3>
        <p>Ayo uses native containers: Apple Container (macOS) or systemd-nspawn (Linux).</p>
    </aside>
</section>
```

3. Sharing Directories:
```html
<section>
    <h2>Sharing Host Directories</h2>
    <pre data-copy><code># Share a directory with the sandbox
ayo share ~/Code/myproject

# Now accessible at /workspace/myproject inside

# List shares
ayo share list

# Remove share
ayo share rm myproject</code></pre>
</section>
```

4. Sandbox Commands:
```html
<section data-table="striped">
    <h2>Sandbox Commands</h2>
    <table>
        <thead><tr><th>Command</th><th>Action</th></tr></thead>
        <tbody>
            <tr><td><code>ayo sandbox list</code></td><td>List running sandboxes</td></tr>
            <tr><td><code>ayo sandbox exec</code></td><td>Run command in sandbox</td></tr>
            <tr><td><code>ayo sandbox login</code></td><td>Interactive shell</td></tr>
        </tbody>
    </table>
</section>
```

