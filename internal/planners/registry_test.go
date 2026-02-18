package planners

import (
	"errors"
	"sync"
	"testing"
)

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	factory := func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{name: "test"}, nil
	}

	r.Register("test-planner", factory)

	if !r.Has("test-planner") {
		t.Error("expected planner to be registered")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry()

	factory := func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{name: "test"}, nil
	}

	r.Register("test-planner", factory)

	if !r.Unregister("test-planner") {
		t.Error("expected Unregister to return true")
	}

	if r.Has("test-planner") {
		t.Error("expected planner to be unregistered")
	}

	if r.Unregister("test-planner") {
		t.Error("expected Unregister to return false for non-existent planner")
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()

	factory := func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{name: "test"}, nil
	}

	r.Register("test-planner", factory)

	f, ok := r.Get("test-planner")
	if !ok {
		t.Fatal("expected to get factory")
	}
	if f == nil {
		t.Error("factory should not be nil")
	}

	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("expected Get to return false for nonexistent planner")
	}
}

func TestRegistry_Has(t *testing.T) {
	r := NewRegistry()

	if r.Has("test") {
		t.Error("expected Has to return false for empty registry")
	}

	r.Register("test", func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{}, nil
	})

	if !r.Has("test") {
		t.Error("expected Has to return true after registration")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()

	// Empty registry
	if len(r.List()) != 0 {
		t.Error("expected empty list")
	}

	// Add some planners
	r.Register("zebra", func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{}, nil
	})
	r.Register("alpha", func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{}, nil
	})
	r.Register("beta", func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{}, nil
	})

	list := r.List()
	if len(list) != 3 {
		t.Errorf("expected 3 planners, got %d", len(list))
	}

	// Should be sorted
	expected := []string{"alpha", "beta", "zebra"}
	for i, name := range expected {
		if list[i] != name {
			t.Errorf("expected list[%d] = %s, got %s", i, name, list[i])
		}
	}
}

func TestRegistry_Count(t *testing.T) {
	r := NewRegistry()

	if r.Count() != 0 {
		t.Errorf("expected 0, got %d", r.Count())
	}

	r.Register("one", func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{}, nil
	})
	r.Register("two", func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{}, nil
	})

	if r.Count() != 2 {
		t.Errorf("expected 2, got %d", r.Count())
	}
}

func TestRegistry_Instantiate(t *testing.T) {
	r := NewRegistry()

	r.Register("test-planner", func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{
			name:     "test-planner",
			stateDir: ctx.StateDir,
		}, nil
	})

	ctx := PlannerContext{
		SandboxName: "test-sandbox",
		StateDir:    "/tmp/state",
	}

	planner, err := r.Instantiate("test-planner", ctx)
	if err != nil {
		t.Fatalf("Instantiate failed: %v", err)
	}

	if planner.Name() != "test-planner" {
		t.Errorf("expected name 'test-planner', got '%s'", planner.Name())
	}
	if planner.StateDir() != "/tmp/state" {
		t.Errorf("expected stateDir '/tmp/state', got '%s'", planner.StateDir())
	}
}

func TestRegistry_Instantiate_NotFound(t *testing.T) {
	r := NewRegistry()

	_, err := r.Instantiate("nonexistent", PlannerContext{})
	if err == nil {
		t.Error("expected error for nonexistent planner")
	}
	if err.Error() != "planner not found: nonexistent" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegistry_Instantiate_FactoryError(t *testing.T) {
	r := NewRegistry()

	r.Register("failing-planner", func(ctx PlannerContext) (PlannerPlugin, error) {
		return nil, errors.New("factory failed")
	})

	_, err := r.Instantiate("failing-planner", PlannerContext{})
	if err == nil {
		t.Error("expected error from failing factory")
	}
	if err.Error() != "failed to create planner failing-planner: factory failed" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegistry_MustInstantiate_Panics(t *testing.T) {
	r := NewRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from MustInstantiate")
		}
	}()

	r.MustInstantiate("nonexistent", PlannerContext{})
}

func TestRegistry_MustInstantiate_Success(t *testing.T) {
	r := NewRegistry()

	r.Register("test", func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{name: "test"}, nil
	})

	// Should not panic
	planner := r.MustInstantiate("test", PlannerContext{})
	if planner.Name() != "test" {
		t.Errorf("expected name 'test', got '%s'", planner.Name())
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := NewRegistry()

	// Register some base planners
	for i := 0; i < 10; i++ {
		name := string(rune('a' + i))
		r.Register(name, func(ctx PlannerContext) (PlannerPlugin, error) {
			return &mockPlanner{name: name}, nil
		})
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 100)

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.List()
			_ = r.Has("a")
			_, _ = r.Get("b")
			_ = r.Count()
		}()
	}

	// Concurrent writes
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := string(rune('z' - i))
			r.Register(name, func(ctx PlannerContext) (PlannerPlugin, error) {
				return &mockPlanner{name: name}, nil
			})
		}(i)
	}

	// Concurrent instantiations
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := r.Instantiate("a", PlannerContext{StateDir: "/tmp"})
			if err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent instantiation error: %v", err)
	}
}

func TestDefaultRegistry(t *testing.T) {
	// Save and restore default registry
	saved := DefaultRegistry
	defer func() { DefaultRegistry = saved }()

	DefaultRegistry = NewRegistry()

	// Test convenience functions
	Register("convenience-test", func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{name: "convenience-test"}, nil
	})

	if _, ok := Get("convenience-test"); !ok {
		t.Error("expected Get to find registered planner")
	}

	list := List()
	if len(list) != 1 || list[0] != "convenience-test" {
		t.Errorf("expected list ['convenience-test'], got %v", list)
	}

	planner, err := Instantiate("convenience-test", PlannerContext{})
	if err != nil {
		t.Fatalf("Instantiate failed: %v", err)
	}
	if planner.Name() != "convenience-test" {
		t.Errorf("expected name 'convenience-test', got '%s'", planner.Name())
	}
}

func TestRegistry_OverwriteExisting(t *testing.T) {
	r := NewRegistry()

	r.Register("test", func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{name: "original"}, nil
	})

	planner1, _ := r.Instantiate("test", PlannerContext{})
	if planner1.Name() != "original" {
		t.Errorf("expected 'original', got '%s'", planner1.Name())
	}

	// Overwrite
	r.Register("test", func(ctx PlannerContext) (PlannerPlugin, error) {
		return &mockPlanner{name: "replacement"}, nil
	})

	planner2, _ := r.Instantiate("test", PlannerContext{})
	if planner2.Name() != "replacement" {
		t.Errorf("expected 'replacement', got '%s'", planner2.Name())
	}
}

// mockPlanner is defined in planners_test.go but we need it here too
// for the registry tests. The test file will be in the same package
// so we can reuse it.

func init() {
	// Ensure mockPlanner satisfies PlannerPlugin
	var _ PlannerPlugin = (*mockPlanner)(nil)
}
