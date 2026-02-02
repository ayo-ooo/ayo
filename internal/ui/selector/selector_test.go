package selector

import (
	"testing"
)

func TestNew(t *testing.T) {
	items := []string{"a", "b", "c"}
	s := New(items)

	if s.Len() != 3 {
		t.Errorf("Len() = %d, want 3", s.Len())
	}
	if s.Index() != 0 {
		t.Errorf("Index() = %d, want 0", s.Index())
	}
}

func TestSelector_Items(t *testing.T) {
	items := []int{1, 2, 3}
	s := New(items)

	got := s.Items()
	if len(got) != len(items) {
		t.Errorf("len(Items()) = %d, want %d", len(got), len(items))
	}
}

func TestSelector_SetItems(t *testing.T) {
	s := New([]string{"a", "b", "c"})
	s.SetIndex(2)

	s.SetItems([]string{"x", "y"})

	if s.Len() != 2 {
		t.Errorf("Len() = %d, want 2", s.Len())
	}
	// Index should be clamped
	if s.Index() != 1 {
		t.Errorf("Index() = %d, want 1 (clamped)", s.Index())
	}
}

func TestSelector_SetItemsEmpty(t *testing.T) {
	s := New([]string{"a", "b"})
	s.SetItems([]string{})

	if s.Len() != 0 {
		t.Errorf("Len() = %d, want 0", s.Len())
	}
	if s.Index() != 0 {
		t.Errorf("Index() = %d, want 0", s.Index())
	}
}

func TestSelector_Selected(t *testing.T) {
	s := New([]string{"a", "b", "c"})

	if s.Selected() != "a" {
		t.Errorf("Selected() = %q, want %q", s.Selected(), "a")
	}

	s.SetIndex(2)
	if s.Selected() != "c" {
		t.Errorf("Selected() = %q, want %q", s.Selected(), "c")
	}
}

func TestSelector_SelectedEmpty(t *testing.T) {
	s := New([]string{})

	if s.Selected() != "" {
		t.Errorf("Selected() = %q, want empty string", s.Selected())
	}
}

func TestSelector_SetIndex(t *testing.T) {
	s := New([]string{"a", "b", "c"})

	s.SetIndex(1)
	if s.Index() != 1 {
		t.Errorf("Index() = %d, want 1", s.Index())
	}

	// Test clamping
	s.SetIndex(-5)
	if s.Index() != 0 {
		t.Errorf("Index() = %d, want 0 (clamped)", s.Index())
	}

	s.SetIndex(100)
	if s.Index() != 2 {
		t.Errorf("Index() = %d, want 2 (clamped)", s.Index())
	}
}

func TestSelector_Next(t *testing.T) {
	s := New([]string{"a", "b", "c"})

	if !s.Next() {
		t.Error("Next() should return true when not at end")
	}
	if s.Index() != 1 {
		t.Errorf("Index() = %d, want 1", s.Index())
	}

	s.Next()
	if s.Next() {
		t.Error("Next() should return false at end")
	}
	if s.Index() != 2 {
		t.Errorf("Index() = %d, want 2", s.Index())
	}
}

func TestSelector_Prev(t *testing.T) {
	s := New([]string{"a", "b", "c"})
	s.SetIndex(2)

	if !s.Prev() {
		t.Error("Prev() should return true when not at start")
	}
	if s.Index() != 1 {
		t.Errorf("Index() = %d, want 1", s.Index())
	}

	s.Prev()
	if s.Prev() {
		t.Error("Prev() should return false at start")
	}
	if s.Index() != 0 {
		t.Errorf("Index() = %d, want 0", s.Index())
	}
}

func TestSelector_OnFirst(t *testing.T) {
	s := New([]string{"a", "b", "c"})

	if !s.OnFirst() {
		t.Error("OnFirst() should be true at start")
	}

	s.Next()
	if s.OnFirst() {
		t.Error("OnFirst() should be false after Next()")
	}
}

func TestSelector_OnLast(t *testing.T) {
	s := New([]string{"a", "b", "c"})

	if s.OnLast() {
		t.Error("OnLast() should be false at start")
	}

	s.SetIndex(2)
	if !s.OnLast() {
		t.Error("OnLast() should be true at end")
	}
}

func TestSelector_OnLastEmpty(t *testing.T) {
	s := New([]string{})

	if !s.OnLast() {
		t.Error("OnLast() should be true for empty list")
	}
}

func TestSelector_Range(t *testing.T) {
	s := New([]int{10, 20, 30})

	var sum int
	s.Range(func(i int, item int) bool {
		sum += item
		return true
	})

	if sum != 60 {
		t.Errorf("sum = %d, want 60", sum)
	}
}

func TestSelector_RangeEarlyExit(t *testing.T) {
	s := New([]int{1, 2, 3, 4, 5})

	var count int
	s.Range(func(i int, item int) bool {
		count++
		return i < 2 // Stop at index 2
	})

	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

func TestSelector_Find(t *testing.T) {
	s := New([]string{"apple", "banana", "cherry"})

	idx := s.Find(func(item string) bool {
		return item == "banana"
	})

	if idx != 1 {
		t.Errorf("Find() = %d, want 1", idx)
	}
}

func TestSelector_FindNotFound(t *testing.T) {
	s := New([]string{"a", "b", "c"})

	idx := s.Find(func(item string) bool {
		return item == "z"
	})

	if idx != -1 {
		t.Errorf("Find() = %d, want -1", idx)
	}
}

func TestSelector_Get(t *testing.T) {
	s := New([]string{"a", "b", "c"})

	if s.Get(0) != "a" {
		t.Errorf("Get(0) = %q, want %q", s.Get(0), "a")
	}
	if s.Get(2) != "c" {
		t.Errorf("Get(2) = %q, want %q", s.Get(2), "c")
	}
}

func TestSelector_GetOutOfBounds(t *testing.T) {
	s := New([]string{"a", "b"})

	if s.Get(-1) != "" {
		t.Errorf("Get(-1) = %q, want empty", s.Get(-1))
	}
	if s.Get(10) != "" {
		t.Errorf("Get(10) = %q, want empty", s.Get(10))
	}
}

func TestSelector_Add(t *testing.T) {
	s := New([]string{"a", "b"})
	s.Add("c")

	if s.Len() != 3 {
		t.Errorf("Len() = %d, want 3", s.Len())
	}
	if s.Get(2) != "c" {
		t.Errorf("Get(2) = %q, want %q", s.Get(2), "c")
	}
}

func TestSelector_Remove(t *testing.T) {
	s := New([]string{"a", "b", "c"})
	s.Remove(1)

	if s.Len() != 2 {
		t.Errorf("Len() = %d, want 2", s.Len())
	}
	if s.Get(1) != "c" {
		t.Errorf("Get(1) = %q, want %q", s.Get(1), "c")
	}
}

func TestSelector_RemoveAdjustsIndex(t *testing.T) {
	s := New([]string{"a", "b", "c"})
	s.SetIndex(2)
	s.Remove(2)

	if s.Index() != 1 {
		t.Errorf("Index() = %d, want 1 (adjusted)", s.Index())
	}
}

func TestSelector_RemoveOutOfBounds(t *testing.T) {
	s := New([]string{"a", "b"})
	s.Remove(-1)
	s.Remove(10)

	if s.Len() != 2 {
		t.Errorf("Len() = %d, want 2", s.Len())
	}
}

func TestSelector_Clear(t *testing.T) {
	s := New([]string{"a", "b", "c"})
	s.SetIndex(2)
	s.Clear()

	if s.Len() != 0 {
		t.Errorf("Len() = %d, want 0", s.Len())
	}
	if s.Index() != 0 {
		t.Errorf("Index() = %d, want 0", s.Index())
	}
}

func TestSelector_WithIntType(t *testing.T) {
	s := New([]int{100, 200, 300})

	if s.Selected() != 100 {
		t.Errorf("Selected() = %d, want 100", s.Selected())
	}

	s.Next()
	if s.Selected() != 200 {
		t.Errorf("Selected() = %d, want 200", s.Selected())
	}
}

func TestSelector_WithStructType(t *testing.T) {
	type Item struct {
		ID   int
		Name string
	}

	s := New([]Item{
		{ID: 1, Name: "one"},
		{ID: 2, Name: "two"},
	})

	if s.Selected().ID != 1 {
		t.Errorf("Selected().ID = %d, want 1", s.Selected().ID)
	}

	s.Next()
	if s.Selected().Name != "two" {
		t.Errorf("Selected().Name = %q, want %q", s.Selected().Name, "two")
	}
}
