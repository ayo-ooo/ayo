package daemon

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestTriggerEngine_StartStop(t *testing.T) {
	engine := NewTriggerEngine(TriggerEngineConfig{
		Logger: slog.Default(),
	})

	ctx := context.Background()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if err := engine.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestTriggerEngine_RegisterCron(t *testing.T) {
	var fired atomic.Int32

	engine := NewTriggerEngine(TriggerEngineConfig{
		Logger: slog.Default(),
		Callback: func(event TriggerEvent) {
			fired.Add(1)
		},
	})

	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer engine.Stop(ctx)

	trigger := &Trigger{
		ID:      "test-cron",
		Type:    TriggerTypeCron,
		Agent:   "@test",
		Enabled: true,
		Config: TriggerConfig{
			Schedule: "*/1 * * * * *", // Every second
		},
		Prompt: "Test prompt",
	}

	if err := engine.Register(trigger); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Wait for at least one trigger
	time.Sleep(1500 * time.Millisecond)

	if fired.Load() == 0 {
		t.Error("cron trigger should have fired at least once")
	}
}

func TestTriggerEngine_RegisterWatch(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "trigger-test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	var fired atomic.Int32
	var lastEvent TriggerEvent

	engine := NewTriggerEngine(TriggerEngineConfig{
		Logger: slog.Default(),
		Callback: func(event TriggerEvent) {
			fired.Add(1)
			lastEvent = event
		},
	})

	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer engine.Stop(ctx)

	trigger := &Trigger{
		ID:      "test-watch",
		Type:    TriggerTypeWatch,
		Agent:   "@test",
		Enabled: true,
		Config: TriggerConfig{
			Path:     tmpDir,
			Patterns: []string{"*.txt"},
			Events:   []string{"create", "modify"},
		},
		Prompt: "File changed",
	}

	if err := engine.Register(trigger); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Wait for watcher to be ready
	time.Sleep(100 * time.Millisecond)

	// Create a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// Wait for trigger
	time.Sleep(200 * time.Millisecond)

	if fired.Load() == 0 {
		t.Error("watch trigger should have fired")
	}

	if lastEvent.Agent != "@test" {
		t.Errorf("expected agent @test, got %s", lastEvent.Agent)
	}
}

func TestTriggerEngine_List(t *testing.T) {
	engine := NewTriggerEngine(TriggerEngineConfig{
		Logger: slog.Default(),
	})

	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer engine.Stop(ctx)

	trigger1 := &Trigger{
		ID:      "trigger-1",
		Type:    TriggerTypeCron,
		Agent:   "@agent1",
		Enabled: true,
		Config: TriggerConfig{
			Schedule: "0 0 * * * *",
		},
	}

	trigger2 := &Trigger{
		ID:      "trigger-2",
		Type:    TriggerTypeCron,
		Agent:   "@agent2",
		Enabled: true,
		Config: TriggerConfig{
			Schedule: "0 0 * * * *",
		},
	}

	engine.Register(trigger1)
	engine.Register(trigger2)

	triggers := engine.List()
	if len(triggers) != 2 {
		t.Errorf("expected 2 triggers, got %d", len(triggers))
	}
}

func TestTriggerEngine_Unregister(t *testing.T) {
	engine := NewTriggerEngine(TriggerEngineConfig{
		Logger: slog.Default(),
	})

	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer engine.Stop(ctx)

	trigger := &Trigger{
		ID:      "test-trigger",
		Type:    TriggerTypeCron,
		Agent:   "@test",
		Enabled: true,
		Config: TriggerConfig{
			Schedule: "0 0 * * * *",
		},
	}

	engine.Register(trigger)

	if len(engine.List()) != 1 {
		t.Error("expected 1 trigger after register")
	}

	engine.Unregister("test-trigger")

	if len(engine.List()) != 0 {
		t.Error("expected 0 triggers after unregister")
	}
}

func TestTriggerEngine_Get(t *testing.T) {
	engine := NewTriggerEngine(TriggerEngineConfig{
		Logger: slog.Default(),
	})

	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer engine.Stop(ctx)

	trigger := &Trigger{
		ID:      "test-trigger",
		Type:    TriggerTypeCron,
		Agent:   "@test",
		Enabled: true,
		Config: TriggerConfig{
			Schedule: "0 0 * * * *",
		},
	}

	engine.Register(trigger)

	got, err := engine.Get("test-trigger")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Agent != "@test" {
		t.Errorf("expected agent @test, got %s", got.Agent)
	}

	_, err = engine.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent trigger")
	}
}

func TestTriggerEngine_DisabledTrigger(t *testing.T) {
	engine := NewTriggerEngine(TriggerEngineConfig{
		Logger: slog.Default(),
	})

	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer engine.Stop(ctx)

	trigger := &Trigger{
		ID:      "disabled-trigger",
		Type:    TriggerTypeCron,
		Agent:   "@test",
		Enabled: false,
		Config: TriggerConfig{
			Schedule: "0 0 * * * *",
		},
	}

	engine.Register(trigger)

	// Disabled triggers shouldn't be registered
	if len(engine.List()) != 0 {
		t.Error("disabled trigger should not be registered")
	}
}

func TestGenerateTriggerID(t *testing.T) {
	id1 := GenerateTriggerID()
	id2 := GenerateTriggerID()

	if id1 == "" {
		t.Error("trigger ID should not be empty")
	}

	if id1 == id2 {
		t.Error("trigger IDs should be unique")
	}

	if len(id1) < 5 || id1[:5] != "trig_" {
		t.Errorf("trigger ID should have 'trig_' prefix, got: %s", id1)
	}
}
