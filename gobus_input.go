package gobus

import (
	"fmt"
	"reflect"
	"sync"
)

type inElement struct {
	event string
	in    []reflect.Value
}

var inElementPool = sync.Pool{
	New: func() interface{} {
		return &inElement{}
	},
}

func (gb *GoBus) Connect(events []string, fn interface{}) (err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = fmt.Errorf("gobus: connect fail %v", err1)
		}
	}()
	fnVal := reflect.ValueOf(fn)
	if fnVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("fn must be ptr, but %v", fnVal.Kind())
		return
	} else if fnVal.IsZero() {
		err = fmt.Errorf("fn point to nil")
		return
	} else if fnVal.Elem().Kind() != reflect.Func {
		err = fmt.Errorf("fn must point to func, but %v", fnVal.Elem().Kind())
		return
	}
	fnVal = fnVal.Elem()
	if !fnVal.CanSet() {
		err = fmt.Errorf("fn must point to func and can set")
		return
	}
	var fnRtVal []reflect.Value
	if fnVal.Type().NumOut() > 0 {
		fnRtVal = make([]reflect.Value, 0, fnVal.Type().NumOut())
		for i := 0; i < fnVal.Type().NumOut(); i++ {
			fnRtVal = append(fnRtVal, reflect.New(fnVal.Type().Out(i)).Elem())
		}
	}

	newFnVal := reflect.MakeFunc(fnVal.Type(), func(in []reflect.Value) (out []reflect.Value) {
		out = fnRtVal
		for i := range events {
			ie := inElementPool.Get().(*inElement)
			ie.event = events[i]
			ie.in = in
			err := gb.inChan.Push(ie)
			if err != nil {
				return
			}
		}
		return
	})
	fnVal.Set(newFnVal)

	return
}

func (gb *GoBus) Triggers(events []string, args ...interface{}) (err error) {
	for i := range events {
		err1 := gb.Trigger(events[i], args...)
		if err1 != nil {
			err = err1
		}
	}
	return
}

func (gb *GoBus) Trigger(event string, args ...interface{}) (err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = fmt.Errorf("gobus: %v", err1)
		}
	}()
	ie := inElementPool.Get().(*inElement)

	ie.event = event
	ie.in = make([]reflect.Value, 0, len(args))

	for i := range args {
		ie.in = append(ie.in, reflect.ValueOf(args[i]))
	}

	err = gb.inChan.Push(ie)

	return
}
