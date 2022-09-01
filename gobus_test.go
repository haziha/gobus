package gobus

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func test() {
	fmt.Println("global func")
}

type demo struct {
	a int
}

func (d demo) Test() {
	fmt.Println("a:", d.a)
}

func TestGoBus_Trigger(t *testing.T) {
	d := &demo{a: 100}
	bus := New(0, 0)
	_, _ = bus.Bind("test", d.Test)
	_, _ = bus.Bind("test", test)
	_, _ = bus.Bind("test", func() {
		fmt.Println("trigger 0")
	})
	cf, _ := bus.Bind("test", func(str string) {
		fmt.Println("trigger 1", str)
	})
	_ = bus.Trigger("test 1")
	time.Sleep(time.Second)
	_ = bus.Trigger("test", "this is args")
	time.Sleep(time.Second)
	_ = bus.Trigger("test")
	time.Sleep(time.Second)

	cf()
	_ = bus.Trigger("test")
	time.Sleep(time.Second)

	bus.Unbind("test", test)
	_ = bus.Trigger("test")
	time.Sleep(time.Second)

	bus.Unbind("test")
	_ = bus.Trigger("test")
	time.Sleep(time.Second)
}

func TestGoBus_Bind(t *testing.T) {
	bus := New(0, 0)
	_, err := bus.Bind("test", func() {})
	if err != nil {
		log.Panicln(err)
	}
	_, err = bus.Bind("test", func(str string) {})
	if err != nil {
		log.Panicln(err)
	}
	bus.Close()
}

func TestNew(t *testing.T) {
	bus := New(0, 0)
	bus.Close()
}
