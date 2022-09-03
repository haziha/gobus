package gobus

import (
	"fmt"
	"github.com/haziha/gochan"
	"github.com/haziha/golist"
	"log"
	"reflect"
	"sync"
)

func New(inBufLen, routineCount int) (gb *GoBus) {
	if inBufLen < 0 {
		inBufLen = -1
	}
	if routineCount <= 0 {
		routineCount = 1
	}

	gb = new(GoBus)
	gb.inChan = gochan.New[*inElement](inBufLen)
	gb.outMap = make(map[string]*golist.List[outElement])
	for i := 0; i < routineCount; i++ {
		go gb.goroutine()
	}

	return
}

type GoBus struct {
	rwLock sync.RWMutex

	inChan *gochan.GoChan[*inElement]
	outMap map[string]*golist.List[outElement]
}

func (gb *GoBus) goroutine() {
	for {
		ie, ok := gb.inChan.Pop()
		if !ok {
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
				func() {
					defer func() {
						if err1 := recover(); err1 != nil {
							err = fmt.Errorf("%v", err1)
						}
					}()

					e.Value.fnVal.Call(out)
				}()
				if err != nil {
					log.Printf("gobus: call panic %v\n", err.Error())
					continue
				}
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
		if outType.In(i) == in[i].Type() { // 类型完全相同
			out = append(out, in[i])
		} else if outType.In(i).Kind() == reflect.Interface { // 类型为 interface
			// 需要检查方法是否都有实现
			if outType.In(i).NumMethod() > in[i].NumMethod() { // 存在的方法比所需的方法少, 直接返回error
				err = fmt.Errorf("income num method > input num method")
				return
			}
			for j := 0; j < outType.In(i).NumMethod(); j++ {
				name := outType.In(i).Method(j).Name
				if !in[i].MethodByName(name).IsValid() {
					err = fmt.Errorf("not found method \"%s\" in %v", name, in[i].Type())
					return
				}
			}
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
	gb.inChan.Close()
	for k := range gb.outMap {
		for gb.outMap[k].Len() != 0 {
			gb.outMap[k].Remove(gb.outMap[k].Back())
		}
	}
}
