// Package selector provides a generic selection manager for navigating lists.
// This follows the pattern from Huh for managing focus across items.
package selector

// Selector manages a list of items with a current selection index.
type Selector[T any] struct {
	items []T
	index int
}

// New creates a new selector with the given items.
func New[T any](items []T) *Selector[T] {
	return &Selector[T]{
		items: items,
		index: 0,
	}
}

// SetItems replaces the items and resets the index.
func (s *Selector[T]) SetItems(items []T) {
	s.items = items
	if s.index >= len(items) {
		s.index = max(0, len(items)-1)
	}
}

// Items returns all items.
func (s *Selector[T]) Items() []T {
	return s.items
}

// Len returns the number of items.
func (s *Selector[T]) Len() int {
	return len(s.items)
}

// Index returns the current selection index.
func (s *Selector[T]) Index() int {
	return s.index
}

// SetIndex sets the selection index.
func (s *Selector[T]) SetIndex(i int) {
	if i < 0 {
		i = 0
	}
	if i >= len(s.items) {
		i = len(s.items) - 1
	}
	s.index = i
}

// Selected returns the currently selected item.
// Returns the zero value if there are no items.
func (s *Selector[T]) Selected() T {
	if len(s.items) == 0 {
		var zero T
		return zero
	}
	return s.items[s.index]
}

// Next moves to the next item. Returns true if moved.
func (s *Selector[T]) Next() bool {
	if s.index < len(s.items)-1 {
		s.index++
		return true
	}
	return false
}

// Prev moves to the previous item. Returns true if moved.
func (s *Selector[T]) Prev() bool {
	if s.index > 0 {
		s.index--
		return true
	}
	return false
}

// OnFirst returns true if on the first item.
func (s *Selector[T]) OnFirst() bool {
	return s.index == 0
}

// OnLast returns true if on the last item.
func (s *Selector[T]) OnLast() bool {
	return s.index == len(s.items)-1 || len(s.items) == 0
}

// Range iterates over items with their index.
// Return false from f to stop iteration.
func (s *Selector[T]) Range(f func(i int, item T) bool) {
	for i, item := range s.items {
		if !f(i, item) {
			break
		}
	}
}

// Find returns the index of the first item matching the predicate.
// Returns -1 if not found.
func (s *Selector[T]) Find(predicate func(T) bool) int {
	for i, item := range s.items {
		if predicate(item) {
			return i
		}
	}
	return -1
}

// Get returns the item at the given index.
// Returns the zero value if index is out of bounds.
func (s *Selector[T]) Get(i int) T {
	if i < 0 || i >= len(s.items) {
		var zero T
		return zero
	}
	return s.items[i]
}

// Add appends an item to the list.
func (s *Selector[T]) Add(item T) {
	s.items = append(s.items, item)
}

// Remove removes the item at the given index.
func (s *Selector[T]) Remove(i int) {
	if i < 0 || i >= len(s.items) {
		return
	}
	s.items = append(s.items[:i], s.items[i+1:]...)
	if s.index >= len(s.items) {
		s.index = max(0, len(s.items)-1)
	}
}

// Clear removes all items.
func (s *Selector[T]) Clear() {
	s.items = nil
	s.index = 0
}
