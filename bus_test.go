package bus

import (
	"fmt"
	"testing"
	"time"
)

type demo struct{}

func (demo) EventName() string {
	return "demo"
}

func (demo) Data() interface{} {
	return "data"
}

func TestNewBus(t *testing.T) {
	bus := NewBus(0, 0, 0)
	c := make(chan interface{})
	bus.RegisterEvent("demo", c)

	go func() {
		for {
			d := <-c
			fmt.Println(d)
		}
	}()

	bus.GetInputChannel() <- demo{}

	time.Sleep(time.Second * 3)
}
