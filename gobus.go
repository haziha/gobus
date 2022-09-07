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

func (gb *GoBus) Cap() int {
	return gb.inChan.Cap()
}

func (gb *GoBus) Len() int {
	return gb.inChan.Len()
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
					log.Printf("gobus: call panic\n%v\n", err.Error())
					continue
				}
			}
		}

		gb.rwLock.RUnlock()
		inElementPool.Put(ie)
	}
}

func convArgs(in []reflect.Value, outType reflect.Type) (out []reflect.Value, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = fmt.Errorf("convArgs panic\n%v", err1)
		}
	}()
	if (outType.IsVariadic() && len(in) < outType.NumIn()-1) || (!outType.IsVariadic() && len(in) < outType.NumIn()) {
		err = fmt.Errorf("too few input arguments")
		return
	}

	out = make([]reflect.Value, 0, outType.NumIn())
	numIn := outType.NumIn()
	if outType.IsVariadic() {
		numIn--
	}

	for i := 0; i < numIn; i++ {
		if outType.In(i) == in[i].Type() { // 类型完全相同
			out = append(out, in[i])
		} else if in[i].CanConvert(outType.In(i)) { // 判断能否转换
			out = append(out, in[i].Convert(outType.In(i)))
		} else {
			err = fmt.Errorf("cannot convert %v to %v", in[i].Type(), outType.In(i))
			return
		}
	}

	if outType.IsVariadic() {
		elemType := outType.In(outType.NumIn() - 1).Elem()
		for i := numIn; i < len(in); i++ {
			if in[i].Kind() == reflect.Invalid {
				if elemType.Kind() != reflect.Interface {
					err = fmt.Errorf("cannot convert nil to %v", elemType)
					return
				} else {
					out = append(out, reflect.New(elemType).Elem())
				}
			} else if in[i].Type() == elemType {
				out = append(out, in[i])
			} else if in[i].CanConvert(elemType) {
				out = append(out, in[i].Convert(elemType))
			} else {
				err = fmt.Errorf("cannot convert element %v to %v", in[i].Type(), elemType)
				return
			}
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
