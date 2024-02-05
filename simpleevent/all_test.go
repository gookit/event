package simpleevent

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestEvent(t *testing.T) {
	buf := new(bytes.Buffer)

	assert.Panics(t, func() {
		On(" ", func(e *EventData) error {
			return nil
		})
	})

	On("n1", func(e *EventData) error {
		_, _ = fmt.Fprintf(buf, "event: %s", e.Name())
		return nil
	})

	On("n2", func(e *EventData) error {
		return fmt.Errorf("an error")
	})

	assert.True(t, Has("n1"))
	assert.False(t, Has("not-exist"))

	assert.NotEmpty(t, DefaultEM.GetEventHandlers("n1"))
	assert.Contains(t, DefaultEM.EventHandlers(), "n1")

	assert.NoError(t, Fire("n1"))
	assert.Equal(t, "event: n1", buf.String())
	buf.Reset()

	// add a wildcard handler
	On(Wildcard, func(e *EventData) error {
		_, _ = fmt.Fprintf(buf, ", W-event: %s", e.Name())
		return nil
	})
	assert.NoError(t, Fire("n1"))
	assert.NoError(t, Fire("not-exist"))
	assert.Equal(t, "event: n1, W-event: n1", buf.String())
	buf.Reset()

	assert.Panics(t, func() {
		MustFire("n2")
	})

	assert.Contains(t, DefaultEM.String(), "n1 handlers:")
	assert.Equal(t, 1, DefaultEM.EventNames()["n1"])
	assert.NotContains(t, "not-exist", DefaultEM.EventNames())

	assert.True(t, Has("n2"))
	DefaultEM.ClearHandlers("n2")
	assert.False(t, Has("n2"))

	DefaultEM.Clear()
	assert.False(t, Has("n1"))
}

func TestEventData_Abort(t *testing.T) {
	DefaultEM.Clear()
	buf := new(bytes.Buffer)

	// stop by Abort
	On("n1", func(e *EventData) error {
		buf.WriteString("handler0;")
		e.Abort() // abort
		return nil
	})
	On("n1", func(e *EventData) error {
		buf.WriteString("handler1;")
		return nil
	})

	assert.NoError(t, Fire("n1"))
	assert.Equal(t, "handler0;", buf.String())

	// clear buffer
	buf.Reset()

	// stop by return error
	On("n2", func(e *EventData) error {
		buf.WriteString("handler0;")
		return fmt.Errorf("an error")
	})
	On("n2", func(e *EventData) error {
		buf.WriteString("handler1;")
		return nil
	})

	assert.Error(t, Fire("n2"))
	assert.Equal(t, "handler0;", buf.String())
}
