---
id: smft-qzpp
status: closed
deps: [smft-9dox]
links: []
created: 2026-02-12T23:57:03Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Chapter 6 - Skills title and panels

Add Chapter 6: Skills section.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html
**Insert after:** Chapter 5 panels

**Content to add:**

1. Chapter title:
```html
<section data-width="narrow" data-crease="chapter" data-tab="Skills">
    <hgroup>
        <h2>Chapter 6</h2>
        <p>Skills</p>
    </hgroup>
</section>
```

2. What Skills Provide:
```html
<section>
    <article>
        <h2>What Skills Provide</h2>
        <p>System prompt = personality. Skills = modular knowledge you can attach.</p>
        <p>Skills are Markdown files with instructions that get injected into the agent's context.</p>
    </article>
    <aside>
        <h3>Composable</h3>
        <p>Mix and match skills across agents.</p>
    </aside>
</section>
```

3. SKILL.md format:
```html
<section>
    <h2>SKILL.md Format</h2>
    <pre data-copy><code>---
name: debugging
description: Systematic bug-finding
---

# Debugging Methodology

1. Reproduce the bug
2. Isolate the cause
3. Form a hypothesis
4. Test and fix</code></pre>
</section>
```

4. Attaching Skills:
```html
<section>
    <h2>Attaching Skills</h2>
    <pre data-copy><code>// In config.json
{
  "skills": ["debugging", "coding"]
}</code></pre>
</section>
```

