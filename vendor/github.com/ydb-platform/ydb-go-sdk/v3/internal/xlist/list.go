// Package xlist is a copy of standard container/list but uses generics
// for strict checks on compile time
package xlist

// Element is an element of a linked list.
type Element[T any] struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *Element[T]

	// The list to which this element belongs.
	list *list[T]

	// The value stored with this element.
	Value T
}

// Next returns the next list element or nil.
func (e *Element[T]) Next() *Element[T] {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}

	return nil
}

// Prev returns the previous list element or nil.
func (e *Element[T]) Prev() *Element[T] {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}

	return nil
}

// List represents a doubly linked list.
// The zero value for List is an empty list ready to use.
type (
	List[T any] interface {
		// PushBack inserts a new element e with value v at the back of list l and returns e.
		PushBack(v T) *Element[T]

		// PushBackList inserts a copy of another list at the back of list l.
		// The lists l and other may be the same. They must not be nil.
		PushBackList(other List[T])

		// PushFront inserts a new element e with value v at the front of list l and returns e.
		PushFront(v T) *Element[T]

		// PushFrontList inserts a copy of another list at the front of list l.
		// The lists l and other may be the same. They must not be nil.
		PushFrontList(other List[T])

		// MoveToFront moves element e to the front of list l.
		// If e is not an element of l, the list is not modified.
		// The element must not be nil.
		MoveToFront(e *Element[T])

		// MoveToBack moves element e to the back of list l.
		// If e is not an element of l, the list is not modified.
		// The element must not be nil.
		MoveToBack(e *Element[T])

		// MoveAfter moves element e to its new position after mark.
		// If e or mark is not an element of l, or e == mark, the list is not modified.
		// The element and mark must not be nil.
		MoveAfter(e, mark *Element[T])

		// MoveBefore moves element e to its new position before mark.
		// If e or mark is not an element of l, or e == mark, the list is not modified.
		// The element and mark must not be nil.
		MoveBefore(e, mark *Element[T])

		// Remove removes e from l if e is an element of list l.
		// It returns the element value e.Value.
		// The element must not be nil.
		Remove(e *Element[T]) T

		// InsertBefore inserts a new element e with value v immediately before mark and returns e.
		// If mark is not an element of l, the list is not modified.
		// The mark must not be nil.
		InsertBefore(v T, mark *Element[T]) *Element[T]

		// InsertAfter inserts a new element e with value v immediately after mark and returns e.
		// If mark is not an element of l, the list is not modified.
		// The mark must not be nil.
		InsertAfter(v T, mark *Element[T]) *Element[T]

		// Front returns the first element of list l or nil if the list is empty.
		Front() *Element[T]

		// Back returns the last element of list l or nil if the list is empty.
		Back() *Element[T]

		// Len returns the number of elements of list l.
		// The complexity is O(1).
		Len() int

		// Clear clears the list l.
		Clear()
	}
	list[T any] struct {
		root Element[T] // sentinel list element, only &root, root.prev, and root.next are used
		len  int        // current list length excluding (this) sentinel element
	}
)

// Clear clears the list l.
func (l *list[T]) Clear() {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
}

// New returns an initialized list.
func New[T any]() List[T] {
	l := &list[T]{}
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0

	return l
}

// Len returns the number of elements of list l.
// The complexity is O(1).
func (l *list[T]) Len() int { return l.len }

// Front returns the first element of list l or nil if the list is empty.
func (l *list[T]) Front() *Element[T] {
	if l.len == 0 {
		return nil
	}

	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *list[T]) Back() *Element[T] {
	if l.len == 0 {
		return nil
	}

	return l.root.prev
}

// insert inserts e after at, increments l.len, and returns e.
func (l *list[T]) insert(e, at *Element[T]) *Element[T] {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len++

	return e
}

// insertValue is a convenience wrapper for insert(&Element{Value: v}, at).
func (l *list[T]) insertValue(v T, at *Element[T]) *Element[T] {
	return l.insert(&Element[T]{Value: v}, at)
}

// remove removes e from its list, decrements l.len
func (l *list[T]) remove(e *Element[T]) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.len--
}

// move moves e to next to at.
func (l *list[T]) move(e, at *Element[T]) {
	if e == at {
		return
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.Value.
// The element must not be nil.
func (l *list[T]) Remove(e *Element[T]) T {
	if e.list == l {
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero Element) and l.remove will crash
		l.remove(e)
	}

	return e.Value
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *list[T]) PushFront(v T) *Element[T] {
	return l.insertValue(v, &l.root)
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *list[T]) PushBack(v T) *Element[T] {
	return l.insertValue(v, l.root.prev)
}

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *list[T]) InsertBefore(v T, mark *Element[T]) *Element[T] {
	if mark.list != l {
		return nil
	}

	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark.prev)
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *list[T]) InsertAfter(v T, mark *Element[T]) *Element[T] {
	if mark.list != l {
		return nil
	}

	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *list[T]) MoveToFront(e *Element[T]) {
	if e.list != l || l.root.next == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, &l.root)
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *list[T]) MoveToBack(e *Element[T]) {
	if e.list != l || l.root.prev == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, l.root.prev)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *list[T]) MoveBefore(e, mark *Element[T]) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark.prev)
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *list[T]) MoveAfter(e, mark *Element[T]) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark)
}

// PushBackList inserts a copy of another list at the back of list l.
// The lists l and other may be the same. They must not be nil.
func (l *list[T]) PushBackList(other List[T]) {
	for i, e := other.Len(), other.Front(); i > 0; i, e = i-1, e.Next() {
		l.insertValue(e.Value, l.root.prev)
	}
}

// PushFrontList inserts a copy of another list at the front of list l.
// The lists l and other may be the same. They must not be nil.
func (l *list[T]) PushFrontList(other List[T]) {
	for i, e := other.Len(), other.Back(); i > 0; i, e = i-1, e.Prev() {
		l.insertValue(e.Value, &l.root)
	}
}
