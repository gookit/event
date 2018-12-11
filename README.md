# event 

[![GoDoc](https://godoc.org/github.com/gookit/event?status.svg)](https://godoc.org/github.com/gookit/event)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/event)](https://goreportcard.com/report/github.com/gookit/event)

Lightweight event manager and dispatcher implements by Go

## GoDoc

- [godoc for github](https://godoc.org/github.com/gookit/event)

## Usage

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
    }))
	
	// ... ...
	_ = event.Fire("evt1", "arg0", "arg1")
}
```

## LICENSE

**[MIT](LICENSE)**