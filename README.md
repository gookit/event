# Event 

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/event?style=flat-square)
[![GoDoc](https://godoc.org/github.com/gookit/event?status.svg)](https://pkg.go.dev/github.com/gookit/event)
[![Actions Status](https://github.com/gookit/event/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/event/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/event/badge.svg?branch=master)](https://coveralls.io/github/gookit/event?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/event)](https://goreportcard.com/report/github.com/gookit/event)

Lightweight event management, dispatch tool library implemented by Go

- Support for custom definition event objects
- Support for adding multiple listeners to an event
- Support setting the priority of the event listener, the higher the priority, the first to trigger
- Support for a set of event listeners based on the event name prefix `PREFIX.*`.
  - `ModeSimple`(default) - `app.*` event listen, trigger `app.run` `app.end`, Both will fire the `app.*` listener
- New match mode: `ModePath`
  - `*` Only match a segment of characters that are not `.`, allowing for finer monitoring and matching
  - `**` matches any number of characters and can only be used at the beginning or end
- Support for using the wildcard `*` to listen for triggers for all events
- Support async trigger event by `go` channel consumers. use `Async(), FireAsync()`
- Complete unit testing, unit coverage `> 95%`

## [中文说明](README.zh-CN.md)

中文说明请看 **[README.zh-CN](README.zh-CN.md)**

## GoDoc

- [Godoc for GitHub](https://pkg.go.dev/github.com/gookit/event)

## Install

```shell
go get github.com/gookit/event
```

## Main method

- `On/Listen(name string, listener Listener, priority ...int)` Register event listener
- `Subscribe/AddSubscriber(sbr Subscriber)`  Subscribe to support registration of multiple event listeners
- `Trigger/Fire(name string, params M) (error, Event)` Trigger event by name and params
- `MustTrigger/MustFire(name string, params M) Event`   Trigger event, there will be panic if there is an error
- `FireEvent(e Event) (err error)`    Trigger an event based on a given event instance
- `FireBatch(es ...interface{}) (ers []error)` Trigger multiple events at once
- `Async/FireC(name string, params M)` Push event to `chan`, asynchronous consumption processing
- `FireAsync(e Event)`  Push event to `chan`, asynchronous consumption processing
- `AsyncFire(e Event)`  Async fire event by 'go' keywords

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

> Note: The second listener has a higher priority, so it will be executed first.

## Using the wildcard

### Match mode `ModePath`

Register event listener and name end with wildcard `*`:

```go
func main() {
	dbListener1 := event.ListenerFunc(func(e event.Event) error {
		fmt.Printf("handle event: %s\n", e.Name())
		return nil
	})

	event.On("app.db.*", dbListener1, event.Normal)
}
```

Trigger events on other logic:

```go
func doCreate() {
	// do something ...
	// Trigger event
	event.MustFire("app.db.create", event.M{"arg0": "val0", "arg1": "val1"})
}

func doUpdate() {
	// do something ...
	// Trigger event
	event.MustFire("app.db.update", event.M{"arg0": "val0"})
}
```

Like the above, triggering the `app.db.create` `app.db.update` event
will trigger the execution of the `dbListener1` listener.

### Match mode `ModePath`

`ModePath` It is a new pattern of `v1.1.0`, and the wildcard `*` matching logic has been adjusted:

- `*` Only match a segment of characters that are not `.`, allowing for finer monitoring and matching
- `**` matches any number of characters and can only be used at the beginning or end

```go
em := event.NewManager("test", event.UsePathMode)
```

## Async fire events

### Use `chan` fire events

You can use the `Async/FireC/FireAsync` method to trigger events, and the events will be written to chan for asynchronous consumption. 
You can use `CloseWait()` to close the chan and wait for all events to be consumed.

Added option configuration:

- `ChannelSize` Set buffer size for `chan`
- `ConsumerNum` Set how many coroutines to start to consume events

```go
func main() {
	// Note: close event chan on program exit
	defer event.CloseWait()
	// defer event.Close()
	
    // register event listener
    event.On("app.evt1", event.ListenerFunc(func(e event.Event) error {
        fmt.Printf("handle event: %s\n", e.Name())
        return nil
    }), event.Normal)
    
    event.On("app.evt1", event.ListenerFunc(func(e event.Event) error {
        fmt.Printf("handle event: %s\n", e.Name())
        return nil
    }), event.High)
    
    // ... ...
    
    // Asynchronous consumption of events
    event.FireC("app.evt1", event.M{"arg0": "val0", "arg1": "val1"})
}
```

> Note: The event chan should be closed when the program exits. 
> You can use the following method:

- `event.Close()` Close `chan` and no longer accept new events
- `event.CloseWait()` Close `chan` and wait for all event processing to complete

## Write event listeners

### Using anonymous functions

You can use anonymous function for quick write an event lister.

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

You can use struct write an event lister, and it should implementation interface `event.Listener`.

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

Can implementation interface `event.Subscriber` for register
multiple event listeners at once.

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

**Example**

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

If you want to customize the event object or define some fixed event information in advance,
you can implement the `event.Event` interface.

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
