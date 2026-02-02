---
id: am-hm8w
status: open
deps: []
links: []
created: 2026-02-02T02:56:44Z
type: task
priority: 3
assignee: Alex Cabrera
---
# Apply gopls modernize suggestions

Apply Go modernization suggestions from gopls:
- Use slices.Contains instead of manual loops (chain.go:358, resolve.go:221, fantasy_tools.go:223, run.go:491)
- Use strings.CutPrefix instead of HasPrefix+TrimPrefix (frontmatter.go:70)
- Replace interface{} with any (validate.go:71, template_renderer.go)
- Use range over int (run.go:1046, services_test.go:139)
- Use strings.SplitSeq for efficiency (qrcode.go:72)
- Replace deprecated strings.Title with golang.org/x/text/cases (template_renderer.go:134)

