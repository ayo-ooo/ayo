---
id: smft-hroz
status: closed
deps: [smft-dbrk]
links: []
created: 2026-02-12T23:54:55Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Replace Ch1 Installation panel

Replace the Installation panel with ayo-specific content.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html

**Find (lines 49-58):**
```html
<section>
    <h2>Installation</h2>
    <ol>
        <li>Download Ayo from the official website</li>
        ...
    </ol>
</section>
```

**Replace with:**
```html
<section>
    <h2>Installation</h2>
    <pre data-copy><code># Install ayo
go install github.com/alexcabrera/ayo/cmd/ayo@latest

# Set your API key
export ANTHROPIC_API_KEY="sk-..."</code></pre>
    <samp>ayo installed successfully</samp>
</section>
```

