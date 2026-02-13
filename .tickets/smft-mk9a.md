---
id: smft-mk9a
status: closed
deps: [smft-jgxf]
links: []
created: 2026-02-12T23:56:02Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Replace Keyboard Shortcuts with Interaction Modes

Replace the Keyboard Shortcuts table with Interaction Modes.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html

**Find (lines 181-198):**
```html
<section data-table="striped">
    <h2>Keyboard Shortcuts</h2>
    <table>
        ...
    </table>
</section>
```

**Replace with:**
```html
<section data-table="striped">
    <h2>Interaction Modes</h2>
    <table>
        <thead>
            <tr><th>Command</th><th>What it does</th></tr>
        </thead>
        <tbody>
            <tr><td><code>ayo</code></td><td>Interactive chat session</td></tr>
            <tr><td><code>ayo "prompt"</code></td><td>Single question, then exit</td></tr>
            <tr><td><code>ayo -a file</code></td><td>Attach file as context</td></tr>
            <tr><td><code>ayo -c</code></td><td>Continue last conversation</td></tr>
            <tr><td><code>ayo @agent</code></td><td>Use specific agent</td></tr>
        </tbody>
    </table>
</section>
```

