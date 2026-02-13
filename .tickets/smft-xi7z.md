---
id: smft-xi7z
status: closed
deps: [smft-cd0y]
links: []
created: 2026-02-12T23:56:16Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Replace Platform tabs with Attachments example

Replace Platform-Specific Notes tabs with Attachments example.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html

**Find (lines 216-240):**
The tabbed Platform-Specific Notes section.

**Replace with:**
```html
<section>
    <h2>Working with Files</h2>
    <pre data-copy><code># Attach a single file
ayo -a main.go "review this code"

# Attach multiple files
ayo -a src/*.go "find the bug"

# Attachments + specific agent
ayo @reviewer -a api.go "security audit"</code></pre>
</section>
```

