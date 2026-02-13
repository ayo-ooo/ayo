---
id: smft-gujv
status: closed
deps: [smft-hf9l]
links: []
created: 2026-02-12T23:55:49Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Replace Ch2 Comparison spread with config.json example

Replace the Manual vs Ayo Workflow spread with config.json and system.md examples.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html

**Find (lines 153-171):**
```html
<section data-width="spread" data-crease="spread">
    <div>
        <h2>Manual Workflow</h2>
        ...
    </div>
    <div>
        <h2>Ayo Workflow</h2>
        ...
    </div>
</section>
```

**Replace with:**
```html
<section data-width="spread" data-crease="spread">
    <div>
        <h2>config.json</h2>
        <pre data-copy><code>{
  "description": "A friendly greeter",
  "model": "claude-sonnet-4-20250514",
  "allowed_tools": ["bash"]
}</code></pre>
    </div>
    <div>
        <h2>system.md</h2>
        <pre data-copy><code># Greeter

You are a friendly assistant who greets users warmly.

When someone says hello, respond with enthusiasm and ask how you can help today.</code></pre>
    </div>
</section>
```

