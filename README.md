# Event 

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/event?style=flat-square)
[![GoDoc](https://godoc.org/github.com/gookit/event?status.svg)](https://pkg.go.dev/github.com/gookit/event)
[![Actions Status](https://github.com/gookit/event/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/event/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/event/badge.svg?branch=master)](https://coveralls.io/github/gookit/event?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/event)](https://goreportcard.com/report/github.com/gookit/event)

Lightweight event management, dispatch tool library implemented by Go

- Support for custom definition event objects
- Support for adding multiple listeners to an event
- Supports setting the priority of the event listener. The higher the priority, the higher the trigger.
- Support for a set of event listeners based on the event name prefix `PREFIX.*`.
  - add `app.*` event listen, trigger `app.run` `app.end`, Both will trigger the `app.*` event at the same time
- Support for using the wildcard `*` to listen for triggers for all events
- Complete unit testing, unit coverage `> 95%`

## [中文说明](README.zh-CN.md)

中文说明请看 **[README.zh-CN](README.zh-CN.md)**

## GoDoc

- [Godoc for github](https://pkg.go.dev/github.com/gookit/event)

## Main method

- `On/Listen(name string, listener Listener, priority ...int)` Register event listener
- `Subscribe/AddSubscriber(sbr Subscriber)`  Subscribe to support registration of multiple event listeners
- `Trigger/Fire(name string, params M) (error, Event)` Trigger event
- `MustTrigger/MustFire(name string, params M) Event`   Trigger event, there will be panic if there is an error
- `FireEvent(e Event) (err error)`    Trigger an event based on a given event instance
- `FireBatch(es ...interface{}) (ers []error)` Trigger multiple events at once
- `AsyncFire(e Event)`   Async fire event by 'go' keywords

## Quick start

```go
package main

import (
	"fmt"
	
	"github.com/gookit/event"
)

func main() {
	// Register event listener
	event.On("evt1", event.ListenerFunc(func(e event.Event) error {
		fmt.Printf("handle event: %s\n", e.Name())
		return nil
	}), event.Normal)

	// Register multiple listeners
	event.On("evt1", event.ListenerFunc(func(e event.Event) error {
		fmt.Printf("handle event: %s\n", e.Name())
		return nil
	}), event.High)

	// ... ...

	// Trigger event
	// Note: The second listener has a higher priority, so it will be executed first.
	event.MustFire("evt1", event.M{"arg0": "val0", "arg1": "val1"})
}
```

## Write event listeners

### Using anonymous functions

```go
package mypgk

import (
	"fmt"

	"github.com/gookit/event"
)

var fnHandler = func(e event.Event) error {
	fmt.Printf("handle event: %s\n", e.Name())
	return nil
}

func Run() {
	// register
	event.On("evt1", event.ListenerFunc(fnHandler), event.High)
}
```

### Using the structure method

**interface:**

```go
// Listener interface
type Listener interface {
	Handle(e Event) error
}
```

**example:**

> Implementation interface `event.Listener`

```go
package mypgk

import "github.com/gookit/event"

type MyListener struct {
	// userData string
}

func (l *MyListener) Handle(e event.Event) error {
	e.Set("result", "OK")
	return nil
}
```

## Register multiple event listeners

**interface:**

```go
// Subscriber event subscriber interface.
// you can register multi event listeners in a struct func.
type Subscriber interface {
	// SubscribedEvents register event listeners
	// key: is event name
	// value: can be Listener or ListenerItem interface
	SubscribedEvents() map[string]interface{}
}
```

**example**

> Implementation interface `event.Subscriber`

```go
package mypgk

import (
	"fmt"

	"github.com/gookit/event"
)

type MySubscriber struct {
	// ooo
}

func (s *MySubscriber) SubscribedEvents() map[string]interface{} {
	return map[string]interface{}{
		"e1": event.ListenerFunc(s.e1Handler),
		"e2": event.ListenerItem{
			Priority: event.AboveNormal,
			Listener: event.ListenerFunc(func(e Event) error {
				return fmt.Errorf("an error")
			}),
		},
		"e3": &MyListener{},
	}
}

func (s *MySubscriber) e1Handler(e event.Event) error {
	e.Set("e1-key", "val1")
	return nil
}
```

## Write custom events

**interface:**

```go
// Event interface
type Event interface {
	Name() string
	// Target() interface{}
	Get(key string) interface{}
	Add(key string, val interface{})
	Set(key string, val interface{})
	Data() map[string]interface{}
	SetData(M) Event
	Abort(bool)
	IsAborted() bool
}
```

**examples:**

```go
package mypgk

import "github.com/gookit/event"

type MyEvent struct {
	event.BasicEvent
	customData string
}

func (e *MyEvent) CustomData() string {
	return e.customData
}
```

Usage:

```go
e := &MyEvent{customData: "hello"}
e.SetName("e1")
event.AddEvent(e)

// add listener
event.On("e1", event.ListenerFunc(func(e event.Event) error {
	fmt.Printf("custom Data: %s\n", e.(*MyEvent).CustomData())
	return nil
}))

// trigger
event.Fire("e1", nil)
// OR
// event.FireEvent(e)
```

## Gookit packages

- [gookit/ini](https://github.com/gookit/ini) Go config management, use INI files
- [gookit/rux](https://github.com/gookit/rux) Simple and fast request router for golang HTTP 
- [gookit/gcli](https://github.com/gookit/gcli) build CLI application, tool library, running CLI commands
- [gookit/slog](https://github.com/gookit/slog) Lightweight, extensible, configurable logging library written in Go
- [gookit/event](https://github.com/gookit/event) Lightweight event manager and dispatcher implements by Go
- [gookit/cache](https://github.com/gookit/cache) Generic cache use and cache manager for golang. support File, Memory, Redis, Memcached.
- [gookit/config](https://github.com/gookit/config) Go config management. support JSON, YAML, TOML, INI, HCL, ENV and Flags
- [gookit/color](https://github.com/gookit/color) A command-line color library with true color support, universal API methods and Windows support
- [gookit/filter](https://github.com/gookit/filter) Provide filtering, sanitizing, and conversion of golang data
- [gookit/validate](https://github.com/gookit/validate) Use for data validation and filtering. support Map, Struct, Form data
- [gookit/goutil](https://github.com/gookit/goutil) Some utils for the Go: string, array/slice, map, format, cli, env, filesystem, test and more
- More, please see https://github.com/gookit

## LICENSE

**[MIT](LICENSE)**
