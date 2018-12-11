# simple event

Very simple event manager implements.

## Usage

```go
package main

import (
	"fmt"
	"github.com/gookit/event/simpleevent"
)

func main() {
	// register event handler
	simpleevent.On("event1", func(e *simpleevent.EventData) error {
		fmt.Printf("handle the event: %s\n", e.Name())
	    return nil
	})
	
	// register more handler to the event.
	simpleevent.On("event1", func(e *simpleevent.EventData) error {
		fmt.Printf("oo, handle the event: %s\n", e.Name())
	    return nil
	})
	
	// ....
	
	// fire event
	_ = simpleevent.Fire("event1", "arg0", "arg1")
}
```

## LICENSE

**[MIT](../LICENSE)**