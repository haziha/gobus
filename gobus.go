package gobus

import (
	"context"
	"fmt"
	"github.com/haziha/golist"
	"log"
	"reflect"
	"sync"
)

func New(inBufLen, routineCount int) (gb *GoBus) {
	if inBufLen < 0 {
		inBufLen = 0
	}
	if routineCount <= 0 {
		routineCount = 1
	}

	gb = new(GoBus)
	gb.inChan = make(chan *inElement, inBufLen)
	gb.outMap = make(map[string]*golist.List[outElement])
	for i := 0; i < routineCount; i++ {
		go gb.goroutine()
	}

	gb.ctx, gb.cancelFunc = context.WithCancel(context.Background())

	return
}

type GoBus struct {
	rwLock sync.RWMutex

	inChan chan *inElement
	outMap map[string]*golist.List[outElement]

	ctx        context.Context
	cancelFunc context.CancelFunc
}

func (gb *GoBus) goroutine() {
	for {
		var ie *inElement
		select {
		case ie = <-gb.inChan:
			break
		case <-gb.ctx.Done():
			return
		}

		gb.rwLock.RLock()

		if _, ok := gb.outMap[ie.event]; !ok {
			log.Printf("gobus: not found event(%s) in map\n", ie.event)
		} else {
			for e := gb.outMap[ie.event].Front(); e != nil; e = e.Next() {
				out, err := convArgs(ie.in, e.Value.fnType)
				if err != nil {
					log.Printf("gobus: %v\n", err)
					continue
				}
				e.Value.fnVal.Call(out)
			}
		}

		gb.rwLock.RUnlock()
		inElementPool.Put(ie)
	}
}

func convArgs(in []reflect.Value, outType reflect.Type) (out []reflect.Value, err error) {
	if len(in) < outType.NumIn() {
		err = fmt.Errorf("income args (%d) less than input args (%d)", len(in), outType.NumIn())
		return
	}

	out = make([]reflect.Value, 0)
	for i := 0; i < outType.NumIn(); i++ {
		if outType.In(i) == in[i].Type() {
			out = append(out, in[i])
		} else {
			err = fmt.Errorf("income[%d] type(%v) != input[%d] type(%v)", i, in[i].Type(), i, outType.In(i))
			out = nil
			return
		}
	}

	return
}

func (gb *GoBus) Close() {
	gb.cancelFunc()
	close(gb.inChan)
	for k := range gb.outMap {
		for gb.outMap[k].Len() != 0 {
			gb.outMap[k].Remove(gb.outMap[k].Back())
		}
	}
}
