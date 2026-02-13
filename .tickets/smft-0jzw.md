---
id: smft-0jzw
status: closed
deps: [smft-qzpp]
links: []
created: 2026-02-12T23:57:12Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Chapter 7 - Memory & Sessions title and panels

Add Chapter 7: Memory & Sessions section.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html
**Insert after:** Chapter 6 panels

**Content to add:**

1. Chapter title:
```html
<section data-width="narrow" data-crease="chapter" data-tab="Memory">
    <hgroup>
        <h2>Chapter 7</h2>
        <p>Memory & Sessions</p>
    </hgroup>
</section>
```

2. Two Types spread:
```html
<section data-width="spread" data-crease="spread">
    <div>
        <h2>Sessions</h2>
        <p>Full conversation history. Tied to one interaction. Use to continue a specific conversation.</p>
        <pre><code>ayo sessions list
ayo -c "follow up"</code></pre>
    </div>
    <div>
        <h2>Memory</h2>
        <p>Distilled facts and preferences. Persists across all sessions. Used automatically.</p>
        <pre><code>ayo memory store "I prefer tabs"
ayo memory search "preferences"</code></pre>
    </div>
</section>
```

3. When to use which:
```html
<section>
    <blockquote data-type="tip">
        <p>Sessions: "Continue this conversation." Memory: "Remember this forever."</p>
    </blockquote>
</section>
```

