---
id: smft-qfj5
status: closed
deps: [smft-u3tp]
links: []
created: 2026-02-12T23:58:03Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Quick Reference section

Add Quick Reference section and final End panel.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html
**Insert after:** Chapter 11 panels, replacing the placeholder End section

**Content to add:**

1. Quick Reference title:
```html
<section data-width="narrow" data-crease="chapter" data-tab="Reference">
    <hgroup>
        <h2>Quick Reference</h2>
        <p>Commands at a glance</p>
    </hgroup>
</section>
```

2. Core Commands:
```html
<section data-table="striped">
    <h2>Core Commands</h2>
    <table>
        <thead><tr><th>Command</th><th>Action</th></tr></thead>
        <tbody>
            <tr><td><code>ayo</code></td><td>Interactive chat</td></tr>
            <tr><td><code>ayo "prompt"</code></td><td>Single prompt</td></tr>
            <tr><td><code>ayo @agent</code></td><td>Use specific agent</td></tr>
            <tr><td><code>ayo -a file</code></td><td>Attach file</td></tr>
            <tr><td><code>ayo -c</code></td><td>Continue session</td></tr>
        </tbody>
    </table>
</section>
```

3. File Locations:
```html
<section>
    <h2>File Locations</h2>
    <dl data-dl="inline">
        <dt>Config</dt><dd>~/.config/ayo/</dd>
        <dt>Data</dt><dd>~/.local/share/ayo/</dd>
        <dt>Agents</dt><dd>~/.config/ayo/agents/</dd>
        <dt>Database</dt><dd>~/.local/share/ayo/ayo.db</dd>
    </dl>
</section>
```

4. Final panel:
```html
<section data-bg="dark">
    <hgroup>
        <h1>End of Manual</h1>
        <p>Full documentation at github.com/alexcabrera/ayo</p>
    </hgroup>
</section>
```

