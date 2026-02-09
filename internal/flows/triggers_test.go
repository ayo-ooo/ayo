package flows

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestTriggerLoader_LoadCronTrigger(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a flow with a cron trigger
	flowContent := `version: 1
name: daily-report
description: Generate daily report
steps:
  - id: generate
    type: shell
    run: echo "generating report"
triggers:
  - id: daily
    type: cron
    schedule: "0 9 * * *"
    params:
      format: pdf
`
	if err := os.WriteFile(filepath.Join(tmpDir, "daily.yaml"), []byte(flowContent), 0644); err != nil {
		t.Fatal(err)
	}

	var addedMu sync.Mutex
	var addedTriggers []*ExtractedTrigger

	loader := NewTriggerLoader(TriggerLoaderConfig{
		Dirs: []string{tmpDir},
		OnTriggersChanged: func(added, removed []*ExtractedTrigger) {
			addedMu.Lock()
			addedTriggers = append(addedTriggers, added...)
			addedMu.Unlock()
		},
	})

	ctx := context.Background()
	if err := loader.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer loader.Stop(ctx)

	// Give it time to load
	time.Sleep(50 * time.Millisecond)

	triggers := loader.GetTriggers()
	if len(triggers) != 1 {
		t.Fatalf("Expected 1 trigger, got %d", len(triggers))
	}

	trigger := triggers[0]
	if trigger.FlowName != "daily-report" {
		t.Errorf("FlowName = %q, want 'daily-report'", trigger.FlowName)
	}
	if trigger.Schedule != "0 9 * * *" {
		t.Errorf("Schedule = %q, want '0 9 * * *'", trigger.Schedule)
	}
	if !trigger.Enabled {
		t.Error("Expected trigger to be enabled")
	}
}

func TestTriggerLoader_LoadWatchTrigger(t *testing.T) {
	tmpDir := t.TempDir()

	flowContent := `version: 1
name: file-processor
description: Process files on change
steps:
  - id: process
    type: shell
    run: echo "processing"
triggers:
  - id: on-change
    type: watch
    path: /data/inbox
    patterns:
      - "*.csv"
      - "*.json"
    recursive: true
    events:
      - create
      - modify
`
	if err := os.WriteFile(filepath.Join(tmpDir, "processor.yaml"), []byte(flowContent), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewTriggerLoader(TriggerLoaderConfig{
		Dirs: []string{tmpDir},
	})

	ctx := context.Background()
	if err := loader.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer loader.Stop(ctx)

	time.Sleep(50 * time.Millisecond)

	triggers := loader.GetTriggers()
	if len(triggers) != 1 {
		t.Fatalf("Expected 1 trigger, got %d", len(triggers))
	}

	trigger := triggers[0]
	if trigger.Path != "/data/inbox" {
		t.Errorf("Path = %q, want '/data/inbox'", trigger.Path)
	}
	if len(trigger.Patterns) != 2 {
		t.Errorf("Patterns = %v, want 2 patterns", trigger.Patterns)
	}
	if !trigger.Recursive {
		t.Error("Expected recursive to be true")
	}
}

func TestTriggerLoader_MultipleFlows(t *testing.T) {
	tmpDir := t.TempDir()

	flow1 := `version: 1
name: flow1
steps:
  - id: step1
    type: shell
    run: echo "1"
triggers:
  - id: t1
    type: cron
    schedule: "0 * * * *"
`
	flow2 := `version: 1
name: flow2
steps:
  - id: step1
    type: shell
    run: echo "2"
triggers:
  - id: t2
    type: cron
    schedule: "*/5 * * * *"
  - id: t3
    type: cron
    schedule: "0 0 * * *"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "flow1.yaml"), []byte(flow1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "flow2.yaml"), []byte(flow2), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewTriggerLoader(TriggerLoaderConfig{
		Dirs: []string{tmpDir},
	})

	ctx := context.Background()
	if err := loader.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer loader.Stop(ctx)

	time.Sleep(50 * time.Millisecond)

	triggers := loader.GetTriggers()
	if len(triggers) != 3 {
		t.Fatalf("Expected 3 triggers, got %d", len(triggers))
	}
}

func TestTriggerLoader_DisabledTrigger(t *testing.T) {
	tmpDir := t.TempDir()

	flowContent := `version: 1
name: disabled-flow
steps:
  - id: step1
    type: shell
    run: echo "test"
triggers:
  - id: disabled
    type: cron
    schedule: "0 * * * *"
    enabled: false
`
	if err := os.WriteFile(filepath.Join(tmpDir, "disabled.yaml"), []byte(flowContent), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewTriggerLoader(TriggerLoaderConfig{
		Dirs: []string{tmpDir},
	})

	ctx := context.Background()
	if err := loader.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer loader.Stop(ctx)

	time.Sleep(50 * time.Millisecond)

	triggers := loader.GetTriggers()
	if len(triggers) != 1 {
		t.Fatalf("Expected 1 trigger, got %d", len(triggers))
	}

	if triggers[0].Enabled {
		t.Error("Expected trigger to be disabled")
	}
}

func TestTriggerLoader_FlowWithoutTriggers(t *testing.T) {
	tmpDir := t.TempDir()

	flowContent := `version: 1
name: no-triggers
steps:
  - id: step1
    type: shell
    run: echo "test"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "no-triggers.yaml"), []byte(flowContent), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewTriggerLoader(TriggerLoaderConfig{
		Dirs: []string{tmpDir},
	})

	ctx := context.Background()
	if err := loader.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer loader.Stop(ctx)

	time.Sleep(50 * time.Millisecond)

	triggers := loader.GetTriggers()
	if len(triggers) != 0 {
		t.Fatalf("Expected 0 triggers, got %d", len(triggers))
	}
}

func TestTriggerLoader_RunsBeforePermanent(t *testing.T) {
	tmpDir := t.TempDir()

	flowContent := `version: 1
name: trial-flow
steps:
  - id: step1
    type: shell
    run: echo "test"
triggers:
  - id: trial
    type: cron
    schedule: "0 * * * *"
    runs_before_permanent: 5
    params:
      key: value
`
	if err := os.WriteFile(filepath.Join(tmpDir, "trial.yaml"), []byte(flowContent), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewTriggerLoader(TriggerLoaderConfig{
		Dirs: []string{tmpDir},
	})

	ctx := context.Background()
	if err := loader.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer loader.Stop(ctx)

	time.Sleep(50 * time.Millisecond)

	triggers := loader.GetTriggers()
	if len(triggers) != 1 {
		t.Fatalf("Expected 1 trigger, got %d", len(triggers))
	}

	trigger := triggers[0]
	if trigger.RunsBeforePermanent != 5 {
		t.Errorf("RunsBeforePermanent = %d, want 5", trigger.RunsBeforePermanent)
	}
	if trigger.Params["key"] != "value" {
		t.Errorf("Params = %v, want key=value", trigger.Params)
	}
}

func TestTriggerLoader_InvalidFlow(t *testing.T) {
	tmpDir := t.TempDir()

	// Invalid YAML
	if err := os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), []byte("not: valid: yaml: {{"), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewTriggerLoader(TriggerLoaderConfig{
		Dirs: []string{tmpDir},
	})

	ctx := context.Background()
	if err := loader.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer loader.Stop(ctx)

	time.Sleep(50 * time.Millisecond)

	// Should not crash, just skip invalid flow
	triggers := loader.GetTriggers()
	if len(triggers) != 0 {
		t.Fatalf("Expected 0 triggers from invalid flow, got %d", len(triggers))
	}
}

func TestTriggerLoader_NonexistentDir(t *testing.T) {
	loader := NewTriggerLoader(TriggerLoaderConfig{
		Dirs: []string{"/nonexistent/path"},
	})

	ctx := context.Background()
	if err := loader.Start(ctx); err != nil {
		t.Fatalf("Start should not fail for nonexistent dir: %v", err)
	}
	defer loader.Stop(ctx)

	triggers := loader.GetTriggers()
	if len(triggers) != 0 {
		t.Fatalf("Expected 0 triggers, got %d", len(triggers))
	}
}
