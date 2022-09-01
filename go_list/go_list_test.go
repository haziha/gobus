package go_list

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	l := New[int]()
	for i := 0; i < 5; i++ {
		a := new(int)
		*a = i
		l.PushBack(a)

		a = new(int)
		*a = i * i
		l.PushFront(a)
	}

	fmt.Println(l.Len())
	for e := l.Front(); e != nil; e = e.Next() {
		fmt.Println(*e.Value)
	}

	l.Remove(l.Front())
	l.Remove(l.Back())
	for e := l.Front(); e != nil; e = e.Next() {
		fmt.Println(*e.Value)
	}
}
