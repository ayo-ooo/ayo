---
id: smft-u3tp
status: closed
deps: [smft-7h2s]
links: []
created: 2026-02-12T23:57:53Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Chapter 11 - Prototyping title and panels

Add Chapter 11: Prototyping Agent Systems section.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html
**Insert after:** Chapter 10 panels

**Content to add:**

1. Chapter title:
```html
<section data-width="narrow" data-crease="chapter" data-tab="Prototyping">
    <hgroup>
        <h2>Chapter 11</h2>
        <p>Prototyping Agent Systems</p>
    </hgroup>
</section>
```

2. Experimental Mindset:
```html
<section>
    <article>
        <h2>The Experimental Mindset</h2>
        <p>Ayo makes it easy to try different approaches. Create variants, test them, iterate. Agents are just directories—cheap to create, easy to modify.</p>
    </article>
    <aside>
        <h3>Key Insight</h3>
        <p>You're not just building agents. You're discovering what works.</p>
    </aside>
</section>
```

3. Team Structures:
```html
<section>
    <h2>Team Structures</h2>
    <dl>
        <dt>Hierarchical</dt>
        <dd>One coordinator routes to specialists</dd>
        <dt>Flat</dt>
        <dd>Peer agents collaborate via shared workspace</dd>
        <dt>Pipeline</dt>
        <dd>Agents chain output to input</dd>
    </dl>
</section>
```

4. Harnesses:
```html
<section>
    <article>
        <h2>Harnesses</h2>
        <p>A harness coordinates multiple agents for complex tasks. Think of it as the test framework for your agent system.</p>
        <p>Flows + delegates + shared sandboxes = harness.</p>
    </article>
    <aside>
        <h3>Advanced</h3>
        <p>Start simple. Add complexity only when needed.</p>
    </aside>
</section>
```

