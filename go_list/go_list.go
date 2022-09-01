package go_list

import "sync"

func New[T any]() (e *List[T]) {
	e = new(List[T])
	e.Init()
	return
}

type List[T any] struct {
	rwLock sync.RWMutex

	len  int
	head Element[T]
	tail Element[T]
}

func (l *List[T]) Remove(e *Element[T]) (v *T) {
	if e == nil || e.list != l {
		return
	}
	v = e.Value

	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil
	e.prev = nil
	e.list = nil

	l.len--

	return
}

func (l *List[T]) PushFront(v *T) (e *Element[T]) {
	e = new(Element[T])

	e.next = l.head.next
	e.prev = &l.head
	l.head.next.prev = e
	l.head.next = e
	e.list = l
	e.Value = v

	l.len++

	return
}

func (l *List[T]) PushBack(v *T) (e *Element[T]) {
	e = new(Element[T])

	e.next = &l.tail
	e.prev = l.tail.prev
	l.tail.prev.next = e
	l.tail.prev = e
	e.list = l
	e.Value = v

	l.len++

	return
}

func (l *List[T]) MoveToFront(e *Element[T]) {
	if e == nil || e.list != l {
		return
	}

	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = l.head.next
	e.prev = &l.head
	l.head.next.prev = e
	l.head.next = e
}

func (l *List[T]) MoveToBack(e *Element[T]) {
	if e == nil || e.list != l {
		return
	}

	e.prev.next = e.next
	e.next.prev = e.prev
	e.prev = l.tail.prev
	e.next = &l.tail
	l.tail.prev.next = e
	l.tail.prev = e
}

func (l *List[T]) MoveBefore(e, mark *Element[T]) {
	if e.list != l || mark.list != l || e == nil || mark == nil || e == mark {
		return
	}

	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = mark
	e.prev = mark.prev
	mark.prev.next = e
	mark.prev = e
}

func (l *List[T]) MoveAfter(e, mark *Element[T]) {
	if e.list != l || mark.list != l || e == nil || mark == nil || e == mark {
		return
	}

	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = mark.next
	e.prev = mark
	mark.next.prev = e
	mark.next = e
}

func (l *List[T]) Len() int {
	return l.len
}

func (l *List[T]) InsertBefore(v *T, mark *Element[T]) (e *Element[T]) {
	if mark.list != l || mark == nil {
		return nil
	}

	e = new(Element[T])

	e.next = mark
	e.prev = mark.prev
	mark.prev.next = e
	mark.prev = e
	e.list = l
	e.Value = v

	l.len++

	return
}

func (l *List[T]) InsertAfter(v *T, mark *Element[T]) (e *Element[T]) {
	if mark.list != l || mark == nil {
		return nil
	}

	e = new(Element[T])

	e.next = mark.next
	e.prev = mark
	mark.next.prev = e
	mark.next = e
	e.list = l
	e.Value = v

	l.len++

	return
}

func (l *List[T]) Init() *List[T] {
	l.len = 0
	l.head.next = &l.tail
	l.head.prev = &l.tail
	l.tail.next = &l.head
	l.tail.prev = &l.head
	return l
}

func (l *List[T]) Front() (e *Element[T]) {
	e = l.head.next
	if e == &l.tail {
		e = nil
	}
	return
}

func (l *List[T]) Back() (e *Element[T]) {
	e = l.tail.prev
	if e == &l.head {
		e = nil
	}
	return
}

type Element[T any] struct {
	next  *Element[T]
	prev  *Element[T]
	list  *List[T]
	Value *T
}

func (e *Element[T]) Prev() *Element[T] {
	if e.prev == &e.list.head {
		return nil
	}
	return e.prev
}

func (e *Element[T]) Next() *Element[T] {
	if e.next == &e.list.tail {
		return nil
	}
	return e.next
}
