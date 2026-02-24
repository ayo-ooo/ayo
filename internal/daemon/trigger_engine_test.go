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

func TestTriggerEngine_RegisterOnceTrigger(t *testing.T) {
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

	// Schedule a one-time trigger 2 seconds from now
	// We use seconds because RFC3339 doesn't preserve sub-second precision
	futureTime := time.Now().Add(2 * time.Second).Format(time.RFC3339)

	trigger := &Trigger{
		ID:      "test-once",
		Type:    TriggerTypeOnce,
		Agent:   "@test",
		Enabled: true,
		Config: TriggerConfig{
			At: futureTime,
		},
	}

	if err := engine.Register(trigger); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Should be registered
	triggers := engine.List()
	if len(triggers) != 1 {
		t.Errorf("expected 1 trigger, got %d", len(triggers))
	}

	// Wait for trigger to fire
	time.Sleep(3 * time.Second)

	if fired.Load() != 1 {
		t.Errorf("expected trigger to fire once, got %d", fired.Load())
	}

	// Trigger should be auto-removed after execution
	time.Sleep(50 * time.Millisecond) // Give cleanup goroutine time
	triggers = engine.List()
	if len(triggers) != 0 {
		t.Errorf("expected trigger to be removed after execution, got %d", len(triggers))
	}
}

func TestTriggerEngine_OnceTrigerPastTimeRejected(t *testing.T) {
	engine := NewTriggerEngine(TriggerEngineConfig{
		Logger: slog.Default(),
	})

	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer engine.Stop(ctx)

	// Try to register a trigger in the past
	pastTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)

	trigger := &Trigger{
		ID:      "test-past",
		Type:    TriggerTypeOnce,
		Agent:   "@test",
		Enabled: true,
		Config: TriggerConfig{
			At: pastTime,
		},
	}

	err := engine.Register(trigger)
	if err == nil {
		t.Error("expected error for past time")
	}
}

func TestTriggerEngine_OnceTriggerMissingAt(t *testing.T) {
	engine := NewTriggerEngine(TriggerEngineConfig{
		Logger: slog.Default(),
	})

	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer engine.Stop(ctx)

	trigger := &Trigger{
		ID:      "test-no-at",
		Type:    TriggerTypeOnce,
		Agent:   "@test",
		Enabled: true,
		Config:  TriggerConfig{}, // Missing At
	}

	err := engine.Register(trigger)
	if err == nil {
		t.Error("expected error for missing 'at' field")
	}
}

func TestTriggerEngine_OnceTriggerInvalidTime(t *testing.T) {
	engine := NewTriggerEngine(TriggerEngineConfig{
		Logger: slog.Default(),
	})

	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer engine.Stop(ctx)

	trigger := &Trigger{
		ID:      "test-invalid",
		Type:    TriggerTypeOnce,
		Agent:   "@test",
		Enabled: true,
		Config: TriggerConfig{
			At: "not-a-valid-time",
		},
	}

	err := engine.Register(trigger)
	if err == nil {
		t.Error("expected error for invalid time format")
	}
}
