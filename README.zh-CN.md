# Event 

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/event?style=flat-square)
[![GoDoc](https://godoc.org/github.com/gookit/event?status.svg)](https://pkg.go.dev/github.com/gookit/event)
[![Actions Status](https://github.com/gookit/event/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/event/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/event/badge.svg?branch=master)](https://coveralls.io/github/gookit/event?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/event)](https://goreportcard.com/report/github.com/gookit/event)

Go 实现的轻量级的事件管理、调度工具库

- 支持自定义定义事件对象
- 支持对一个事件添加多个监听器
- 支持设置事件监听器的优先级，优先级越高越先触发
- 支持根据事件名称前缀 `PREFIX.*` 来进行一组事件监听.
  - 注册`app.*` 事件的监听，触发 `app.run` `app.end` 时，都将同时会触发 `app.*` 事件
- 支持使用通配符 `*` 来监听全部事件的触发
- 完善的单元测试，单元覆盖率 `> 95%`

## [English](README.md)

English introduction, please see **[EN README](README.md)**

## GoDoc

- [Godoc for github](https://pkg.go.dev/github.com/gookit/event)

## 主要方法

- `On/Listen(name string, listener Listener, priority ...int)` 注册事件监听
- `Subscribe/AddSubscriber(sbr Subscriber)`  订阅，支持注册多个事件监听
- `Trigger/Fire(name string, params M) (error, Event)` 触发事件
- `MustTrigger/MustFire(name string, params M) Event`   触发事件，有错误则会panic
- `FireEvent(e Event) (err error)`    根据给定的事件实例，触发事件
- `FireBatch(es ...interface{}) (ers []error)` 一次触发多个事件
- `AsyncFire(e Event)`   Async fire event by 'go' keywords

## 快速使用

```go
package main

import (
	"fmt"
	
	"github.com/gookit/event"
)

func main() {
	// 注册事件监听器
	event.On("evt1", event.ListenerFunc(func(e event.Event) error {
        fmt.Printf("handle event: %s\n", e.Name())
        return nil
    }), event.Normal)
	
	// 注册多个监听器
	event.On("evt1", event.ListenerFunc(func(e event.Event) error {
        fmt.Printf("handle event: %s\n", e.Name())
        return nil
    }), event.High)
	
	// ... ...
	
	// 触发事件
	// 注意：第二个监听器的优先级更高，所以它会先被执行
	event.MustFire("evt1", event.M{"arg0": "val0", "arg1": "val1"})
}
```

> Note: 注意：第二个监听器的优先级更高，所以它会先被执行

### 使用通配符

注册事件监听器和名称以通配符`*`结尾:

```go
func main() {
	dbListener1 := event.ListenerFunc(func(e event.Event) error {
		fmt.Printf("handle event: %s\n", e.Name())
		return nil
	})

	event.On("app.db.*", dbListener1, event.Normal)
}
```

在其他逻辑上触发事件:

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

像上面这样,触发 `app.db.create` `app.db.update` 事件,都会触发执行 `dbListener1` 监听器.

## 编写事件监听

### 使用匿名函数

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

### 使用结构体方法

**interface:**

```go
// Listener interface
type Listener interface {
	Handle(e Event) error
}
```

**示例:**

> 实现接口 `event.Listener`

```go
package mypgk

import (
	"fmt"
	
	"github.com/gookit/event"
)

type MyListener struct {
	// userData string
}

func (l *MyListener) Handle(e event.Event) error {
	e.Set("result", "OK")
	return nil
}
```

### 同时注册多个事件监听

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

**示例**

> 实现接口 `event.Subscriber`

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

## 编写自定义事件

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

**示例**

```go
package mypgk 

import (
	"fmt"
	
	"github.com/gookit/event"
)

type MyEvent struct{
	event.BasicEvent
	customData string
}

func (e *MyEvent) CustomData() string {
    return e.customData
}
```

**使用：**

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

## Gookit 工具包

- [gookit/ini](https://github.com/gookit/ini) INI配置读取管理，支持多文件加载，数据覆盖合并, 解析ENV变量, 解析变量引用
- [gookit/rux](https://github.com/gookit/rux) Simple and fast request router for golang HTTP 
- [gookit/gcli](https://github.com/gookit/gcli) Go的命令行应用，工具库，运行CLI命令，支持命令行色彩，用户交互，进度显示，数据格式化显示
- [gookit/slog](https://github.com/gookit/slog) 用Go编写的轻量级，可扩展，可配置的日志记录库
- [gookit/event](https://github.com/gookit/event) Go实现的轻量级的事件管理、调度程序库, 支持设置监听器的优先级, 支持对一组事件进行监听
- [gookit/cache](https://github.com/gookit/cache) 通用的缓存使用包装库，通过包装各种常用的驱动，来提供统一的使用API
- [gookit/config](https://github.com/gookit/config) Go应用配置管理，支持多种格式（JSON, YAML, TOML, INI, HCL, ENV, Flags），多文件加载，远程文件加载，数据合并
- [gookit/color](https://github.com/gookit/color) CLI 控制台颜色渲染工具库, 拥有简洁的使用API，支持16色，256色，RGB色彩渲染输出
- [gookit/filter](https://github.com/gookit/filter) 提供对Golang数据的过滤，净化，转换
- [gookit/validate](https://github.com/gookit/validate) Go通用的数据验证与过滤库，使用简单，内置大部分常用验证、过滤器
- [gookit/goutil](https://github.com/gookit/goutil) Go 的一些工具函数，格式化，特殊处理，常用信息获取等
- 更多请查看 https://github.com/gookit

## LICENSE

**[MIT](LICENSE)**
