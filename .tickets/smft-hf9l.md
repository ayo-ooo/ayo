---
id: smft-hf9l
status: closed
deps: [smft-wle6]
links: []
created: 2026-02-12T23:55:41Z
type: task
priority: 2
assignee: Alex Cabrera
parent: smft-2r6n
---
# Website: Replace Ch2 Flow steps with Create Agent steps

Replace the Creating Your First Flow steps with Create Your First Agent.

**File:** /Users/acabrera/Code/ayo-ooo/ayo-ooo-website/index.html

**Find (lines 130-151):**
```html
<section data-width="wide">
    <h2>Creating Your First Flow</h2>
    <ol data-flow>
        ...
    </ol>
</section>
```

**Replace with:**
```html
<section data-width="wide">
    <h2>Create Your First Agent</h2>
    <ol data-flow>
        <li>
            <h3>1. Create Directory</h3>
            <p><code>mkdir -p ~/.config/ayo/agents/@greeter</code></p>
        </li>
        <li>
            <h3>2. Write config.json</h3>
            <p>Model, tools, description</p>
        </li>
        <li>
            <h3>3. Write system.md</h3>
            <p>Personality and instructions</p>
        </li>
        <li>
            <h3>4. Test It</h3>
            <p><code>ayo @greeter "hello"</code></p>
        </li>
    </ol>
</section>
```

