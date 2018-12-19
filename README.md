# event 

[![GoDoc](https://godoc.org/github.com/gookit/event?status.svg)](https://godoc.org/github.com/gookit/event)
[![Build Status](https://travis-ci.org/gookit/event.svg?branch=master)](https://travis-ci.org/gookit/event)
[![Coverage Status](https://coveralls.io/repos/github/gookit/event/badge.svg?branch=master)](https://coveralls.io/github/gookit/event?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/event)](https://goreportcard.com/report/github.com/gookit/event)

Go实现的轻量级的事件管理、调度工具库

- 支持定义事件对象
- 支持对一个事件添加多个监听器
- 支持设置监听器的优先级
- 支持根据事件名称来进行一组事件监听
- 支持使用通配符`*`来监听全部事件

## GoDoc

- [godoc for github](https://godoc.org/github.com/gookit/event)

## 主要方法

- `On(name string, listener Listener, priority ...int)`
- `Fire(name string, params M) error`
- `MustFire(name string, params M)`
- `FireEvent(e Event) (err error)`

## 快速使用

```go
package main

import (
	"fmt"
	"github.com/gookit/event"
)

func main() {
	// register event listener
	event.On("evt1", event.ListenerFunc(func(e event.Event) error {
        fmt.Printf("handle event: %s\n", e.Name())
        return nil
    }), event.Normal)
	
	// ... ...
	_ = event.Fire("evt1", "arg0", "arg1")
}
```

## LICENSE

**[MIT](LICENSE)**