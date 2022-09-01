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

	select {
	case gb.inChan <- ie:
	case <-gb.ctx.Done():
		err = fmt.Errorf("gobus: closed")
	}

	return
}
