package gobus

import (
	"fmt"
	"gobus/go_list"
	"reflect"
)

type outElement struct {
	fnType reflect.Type
	fnVal  reflect.Value
}

func (gb *GoBus) Unbind(event string, fns ...interface{}) {
	gb.rwLock.Lock()
	defer gb.rwLock.Unlock()

	if _, ok := gb.outMap[event]; !ok {
		return
	}

	if len(fns) != 0 {
		for i := range fns {
			val := reflect.ValueOf(fns[i])
			ele := make([]*go_list.Element[outElement], 0)
			for element := gb.outMap[event].Front(); element != nil; element = element.Next() {
				if reflect.DeepEqual(val, element.Value.fnVal) {
					ele = append(ele, element)
				}
			}
			for j := range ele {
				gb.outMap[event].Remove(ele[j])
			}
		}
	}

	if len(fns) == 0 || gb.outMap[event].Len() == 0 {
		delete(gb.outMap, event)
	}
}

func (gb *GoBus) Bind(event string, fn interface{}) (unbindFunc func(), err error) {
	fnVal := reflect.ValueOf(fn)
	var oe *outElement = nil

	if fnVal.Kind() == reflect.Func {
		oe = new(outElement)
		oe.fnType = fnVal.Type()
		oe.fnVal = fnVal
	} else {
		err = fmt.Errorf("gobus: bind fail, unknown type")
	}

	if oe == nil {
		return
	}

	gb.rwLock.Lock()
	defer gb.rwLock.Unlock()

	if _, ok := gb.outMap[event]; !ok {
		gb.outMap[event] = go_list.New[outElement]()
	}
	element := gb.outMap[event].PushBack(oe)

	unbindFunc = func() {
		gb.rwLock.Lock()
		defer gb.rwLock.Unlock()

		if _, ok := gb.outMap[event]; !ok {
			return
		}

		gb.outMap[event].Remove(element)

		if gb.outMap[event].Len() == 0 {
			delete(gb.outMap, event)
		}
	}

	return
}