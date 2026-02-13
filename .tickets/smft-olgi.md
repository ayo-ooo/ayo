---
id: smft-olgi
status: closed
deps: [smft-0jzw]
links: []
created: 2026-02-12T23:57:20Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Chapter 8 - Delegation title and panels

Add Chapter 8: Delegation section.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html
**Insert after:** Chapter 7 panels

**Content to add:**

1. Chapter title:
```html
<section data-width="narrow" data-crease="chapter" data-tab="Delegation">
    <hgroup>
        <h2>Chapter 8</h2>
        <p>Delegation</p>
    </hgroup>
</section>
```

2. Task Routing:
```html
<section>
    <article>
        <h2>Task Routing</h2>
        <p>A generalist agent can route specialized tasks to experts. Configure which agent handles which task type.</p>
    </article>
    <aside>
        <h3>Philosophy</h3>
        <p>Do one thing well. Each agent has a focused purpose.</p>
    </aside>
</section>
```

3. Configuring Delegates:
```html
<section>
    <h2>Configuring Delegates</h2>
    <pre data-copy><code>// .ayo.json in your project
{
  "delegates": {
    "coding": "@crush",
    "research": "@researcher"
  }
}</code></pre>
    <samp>Now @ayo routes coding tasks to @crush automatically</samp>
</section>
```

