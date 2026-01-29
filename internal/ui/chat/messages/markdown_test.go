package messages

import (
	"sync"
	"testing"
)

func TestGetMarkdownRenderer_CacheHit(t *testing.T) {
	ClearRendererCache()

	// First call creates renderer
	r1 := GetMarkdownRenderer(80)
	if r1 == nil {
		t.Fatal("expected non-nil renderer")
	}

	// Second call with same width returns same instance
	r2 := GetMarkdownRenderer(80)
	if r1 != r2 {
		t.Error("expected same renderer instance for same width")
	}
}

func TestGetMarkdownRenderer_DifferentWidths(t *testing.T) {
	ClearRendererCache()

	r80 := GetMarkdownRenderer(80)
	r100 := GetMarkdownRenderer(100)

	if r80 == nil || r100 == nil {
		t.Fatal("expected non-nil renderers")
	}

	// Different widths get different renderers
	if r80 == r100 {
		t.Error("expected different renderers for different widths")
	}
}

func TestGetMarkdownRenderer_WidthClamping(t *testing.T) {
	ClearRendererCache()

	// Too small should clamp to 40
	r1 := GetMarkdownRenderer(10)
	r2 := GetMarkdownRenderer(40)
	if r1 != r2 {
		t.Error("expected width 10 to clamp to 40")
	}

	// Too large should clamp to 200
	r3 := GetMarkdownRenderer(500)
	r4 := GetMarkdownRenderer(200)
	if r3 != r4 {
		t.Error("expected width 500 to clamp to 200")
	}
}

func TestGetMarkdownRenderer_Concurrent(t *testing.T) {
	ClearRendererCache()

	var wg sync.WaitGroup
	renderers := make(chan *interface{}, 100)

	// Spawn 100 goroutines all requesting width 80
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := GetMarkdownRenderer(80)
			if r != nil {
				var iface interface{} = r
				renderers <- &iface
			}
		}()
	}

	wg.Wait()
	close(renderers)

	// All should have gotten a renderer
	count := 0
	for range renderers {
		count++
	}
	if count != 100 {
		t.Errorf("expected 100 renderers, got %d", count)
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		v, min, max, want int
	}{
		{50, 0, 100, 50},
		{-10, 0, 100, 0},
		{150, 0, 100, 100},
		{40, 40, 200, 40},
		{200, 40, 200, 200},
	}

	for _, tt := range tests {
		got := clamp(tt.v, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("clamp(%d, %d, %d) = %d, want %d", tt.v, tt.min, tt.max, got, tt.want)
		}
	}
}
