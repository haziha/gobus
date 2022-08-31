package bus

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type BusEvent interface {
	EventName() string
	Data() interface{}
}

type DefaultBusEvent struct {
	eventName string
	data      interface{}
}

func (_this *DefaultBusEvent) EventName() string {
	return _this.eventName
}

func (_this *DefaultBusEvent) Data() interface{} {
	return _this.data
}

func NewBus(inBufLen, timeout, routineCount int) (bus *Bus) {
	bus = new(Bus)
	bus.inCh = make(chan BusEvent, inBufLen)
	bus.outChMap = make(map[string]*list.List)
	bus.timeout = timeout
	bus.ctx, bus.cancelFunc = context.WithCancel(context.Background())

	if routineCount <= 0 {
		routineCount = 1
	}

	for i := 0; i < routineCount; i++ {
		go bus.goroutine()
	}

	return
}

type Bus struct {
	rwLock sync.RWMutex

	inCh     chan BusEvent
	outChMap map[string]*list.List
	timeout  int

	ctx        context.Context
	cancelFunc context.CancelFunc
}

func (bus *Bus) Close() {
	bus.cancelFunc()
}

func sendData(channel chan interface{}, data interface{}, timeout int) (err error) {
	defer func() {
		err1 := recover()
		if err1 != nil {
			err = fmt.Errorf("[%v] send data fail (%v)", channel, err1)
		}
	}()
	if timeout <= 0 {
		channel <- data
	} else {
		timer := time.NewTimer(time.Duration(timeout) * time.Second)
		select {
		case <-timer.C:
			err = fmt.Errorf("[%v] send data timeout", channel)
			break
		case channel <- data:
			break
		}
		timer.Stop()
	}
	return
}

func (bus *Bus) goroutine() {
	for {
		select {
		case <-bus.ctx.Done():
			return
		case data, ok := <-bus.inCh:
			if !ok {
				return
			}
			name, err := getEventName(data)
			if err != nil {
				log.Printf("Bus: get event name fail: %v\n", data)
				break
			}
			bus.rwLock.RLock()
			if l, ok := bus.outChMap[name]; ok {
				for element := l.Front(); element != nil; element = element.Next() {
					err = sendData(element.Value.(chan interface{}), data.Data(), bus.timeout)
					if err != nil {
						log.Println("Bus: " + err.Error())
					}
				}
			} else {
				log.Printf("Bus: not found event \"%s\"", name)
			}
			bus.rwLock.RUnlock()
		}
	}
}

func getEventName(event BusEvent) (name string, err error) {
	defer func() {
		err1 := recover()
		if err1 != nil {
			err = fmt.Errorf("get event name fail: %v", err1)
		}
	}()

	name = event.EventName()
	return
}

func (bus *Bus) Write(eventName string, data interface{}) (err error) {
	defer func() {
		err1 := recover()
		if err1 != nil {
			err = fmt.Errorf("bus: write fail (%v)", err1)
		}
	}()

	bus.GetInputChannel() <- &DefaultBusEvent{eventName: eventName, data: data}
	return
}

func (bus *Bus) GetInputChannel() (be chan<- BusEvent) {
	be = bus.inCh
	return
}

func (bus *Bus) CancelEvent(event string, channel chan interface{}) (err error) {
	bus.rwLock.Lock()
	defer bus.rwLock.Unlock()

	l, ok := bus.outChMap[event]
	if !ok {
		err = fmt.Errorf("not found event \"%s\"", event)
	}

	found := false
	for element := l.Front(); element != nil; element = element.Next() {
		if element.Value.(chan interface{}) == channel {
			l.Remove(element)
			found = true
			break
		}
	}

	if !found {
		err = fmt.Errorf("not found element \"%v\"", channel)
		return
	}

	if l.Len() == 0 {
		delete(bus.outChMap, event)
	}

	return
}

func (bus *Bus) RegisterEvent(event string, channel chan interface{}) {
	bus.rwLock.Lock()
	defer bus.rwLock.Unlock()

	_, ok := bus.outChMap[event]
	if !ok {
		bus.outChMap[event] = list.New()
	}

	bus.outChMap[event].PushBack(channel)

	return
}
