package gobus

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"
)

func test() {
	fmt.Println("global func")
}

type demo struct {
	a  int
	Fn func()
}

func (d demo) Test() {
	fmt.Println("a:", d.a)
}

func (d demo) Error() string {
	return "this is Error method"
}

func TestGoBus_ConvertType2(t *testing.T) {
	b := New(0, 0)
	_, _ = b.Bind("test", func(args ...interface{}) {
		fmt.Println(args)
	})
	_ = b.Trigger("test", "this is string", true, 2333, nil)
	time.Sleep(time.Second)
}

func TestGoBus_ConvertType(t *testing.T) {
	b := New(0, 0)
	_, _ = b.Bind("test", func(a uint64, b int8, c string) {
		fmt.Println(a, b, c)
	})
	_ = b.Trigger("test", int8(1), uint64(2), json.Number("json.Number"))
	time.Sleep(time.Second)
}

func TestGoBus_Bind2(t *testing.T) {
	b := New(0, 0)
	_, _ = b.Bind("test", func(err error) {
		fmt.Println("err:", err)
	})
	_ = b.Trigger("test", &demo{})
	time.Sleep(time.Second)
}

func TestGoBus_Trigger2(t *testing.T) {
	b := New(0, 0)
	_, _ = b.Bind("test", func() {
		panic("demo panic")
	})
	_ = b.Trigger("test")
	time.Sleep(time.Second)
}

func TestGoBus_Connect(t *testing.T) {
	b := New(0, 1)
	_, _ = b.Bind("test", func() {
		fmt.Println("test call")
	})
	_, _ = b.Bind("test1", func() {
		fmt.Println("test1 call")
	})
	d := &demo{}
	if err := b.Connect([]string{"test", "test1"}, &d.Fn); err != nil {
		fmt.Println(err)
	}
	d.Fn()
	time.Sleep(time.Second)
}

func TestGoBus_Triggers(t *testing.T) {
	bus := New(0, 1)
	_, _ = bus.Bind("test", func() {
		fmt.Println("test call")
	})
	_, _ = bus.Bind("test1", func() {
		fmt.Println("test1 call")
	})
	_ = bus.Triggers([]string{"test", "test1"}, "arg1")
	time.Sleep(time.Second)
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
