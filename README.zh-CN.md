# Event 

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/event?style=flat-square)
[![GoDoc](https://pkg.go.dev/badge/github.com/gookit/event.svg)](https://pkg.go.dev/github.com/gookit/event)
[![Actions Status](https://github.com/gookit/event/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/event/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/event/badge.svg?branch=master)](https://coveralls.io/github/gookit/event?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/event)](https://goreportcard.com/report/github.com/gookit/event)

`event` Go 实现的轻量级的事件管理、调度工具库

- 支持自定义创建预定义的事件对象
- 支持对一个事件添加多个监听器
- 支持设置事件监听器的优先级，优先级越高越先触发
- 支持通过通配符 `*` 来进行一组事件的匹配监听.
  - `ModeSimple` - 注册 `app.*` 事件的监听，触发 `app.run` `app.end` 时，都将同时会触发 `app.*` 监听器
  - `ModePath` - **NEW** `*` 只匹配一段非 `.` 的字符,可以进行更精细的监听; `**` 匹配任意多个字符,只能用于开头或结尾
- 支持直接使用通配符 `*` 来监听全部事件的触发
- 支持触发事件时投递到 `chan`, 异步进行消费处理. 触发: `Async(), FireAsync()`
- 完善的单元测试，单元覆盖率 `> 95%`

## [English](README.md)

English introduction, please see **[EN README](README.md)**

## GoDoc

- [Godoc for GitHub](https://pkg.go.dev/github.com/gookit/event)

## 安装

```shell
go get github.com/gookit/event
```

## 主要方法

- `On/Listen(name string, listener Listener, priority ...int)` 注册事件监听
- `Subscribe/AddSubscriber(sbr Subscriber)`  订阅，支持注册多个事件监听
- `Trigger/Fire(name string, params M) (error, Event)` 触发事件
- `MustTrigger/MustFire(name string, params M) Event` 触发事件，有错误则会panic
- `FireEvent(e Event) (err error)` 根据给定的事件实例，触发事件
- `FireBatch(es ...any) (ers []error)` 一次触发多个事件
- `Async/FireC(name string, params M)` 投递事件到 `chan`，异步消费处理
- `FireAsync(e Event)`  投递事件到 `chan`，异步消费处理
- `AsyncFire(e Event)`  简单的通过 `go` 异步触发事件

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

## 使用通配符

### 匹配模式 ModeSimple

`ModeSimple` 是默认模式, 注册事件监听器和名称以通配符 `*` 结尾:

```go
func main() {
	dbListener1 := event.ListenerFunc(func(e event.Event) error {
		fmt.Printf("handle event: %s\n", e.Name())
		return nil
	})

	event.On("app.db.*", dbListener1, event.Normal)
}
```

**在其他逻辑上触发事件**:

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

### 匹配模式 ModePath

`ModePath` 是 `v1.1.0` 新增的模式,通配符 `*` 匹配逻辑有调整:

- `*` 只匹配一段非 `.` 的字符,可以进行更精细的监听匹配
- `**` 则匹配任意多个字符,并且只能用于开头或结尾

```go
em := event.NewManager("test", event.UsePathMode)

// 注册事件监听器
em.On("app.**", appListener)
em.On("app.db.*", dbListener)
em.On("app.*.create", createListener)
em.On("app.*.update", updateListener)

// ... ...

// 触发事件
// TIP: 将会触发 appListener, dbListener, createListener
em.Fire("app.db.create", event.M{"arg0": "val0", "arg1": "val1"})
```

## 异步消费事件

### 使用 `chan` 消费事件

可以使用 `Async/FireC/FireAsync` 方法触发事件，事件将会写入 `chan` 异步消费。可以使用 `CloseWait()` 关闭chan并等待事件全部消费完成。

> **Note**: `event.NewBasic()/event.New()` 可以创建通用的Event实例； `Async/FireC` 无需构建 Event，内部根据参数构建的。

**新增配置选项**:

- `ChannelSize` 设置 `chan` 的缓冲大小
- `ConsumerNum` 设置启动多少个协程来消费事件

```go
func main() {
	// 注意：在程序退出时关闭事件chan
	// defer event.Close()
	defer event.CloseWait()

    // 注册事件监听器
    event.On("app.evt1", event.ListenerFunc(func(e event.Event) error {
        fmt.Printf("handle event: %s\n", e.Name())
        return nil
    }), event.Normal)
    
    // 注册多个监听器
    event.On("app.evt1", event.ListenerFunc(func(e event.Event) error {
        fmt.Printf("handle event: %s\n", e.Name())
        return nil
    }), event.High)
    
    // ... ...
    
    // 异步消费事件
    event.FireC("app.evt1", event.M{"arg0": "val0", "arg1": "val1"})
	event.FireAsync(event.New("app.evt1", event.M{"arg0": "val2"})
}
```

> Note: 应当在程序退出时关闭事件chan. 可以使用下面的方法:
 
- `event.Close()` 立即关闭 `chan` 不再接受新的事件
- `event.CloseWait()` 关闭 `chan` 并等待所有事件处理完成

## 编写事件监听器

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
	SubscribedEvents() map[string]any
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

func (s *MySubscriber) SubscribedEvents() map[string]any {
	return map[string]any{
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

如果你希望自定义事件对象或者提前定义好一些固定事件信息,可以实现 `event.Event` 接口.

**interface:**

```go
// Event interface
type Event interface {
	Name() string
	Get(key string) any
	Add(key string, val any)
	Set(key string, val any)
	Data() map[string]any
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

> **Note**: `AddEvent()` 是用于添加预先定义的公共事件信息，都是在初始化阶段添加，所以没加锁. 在业务中动态创建的Event可以直接使用 `FireEvent()` 触发

## Gookit 工具包

- [gookit/ini](https://github.com/gookit/ini) INI配置读取管理，支持多文件加载，数据覆盖合并, 解析ENV变量, 解析变量引用
- [gookit/rux](https://github.com/gookit/rux) 简单且快速的 Go web 框架，支持中间件，兼容 http.Handler 接口
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
