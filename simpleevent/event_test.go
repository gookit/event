package simpleevent

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvent(t *testing.T) {
	buf := new(bytes.Buffer)

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
	assert.Equal(t, "event: n1, W-event: n1", buf.String())
	buf.Reset()

	assert.Panics(t, func() {
		MustFire("n2")
	})

	assert.True(t, Has("n2"))
	DefaultEM.ClearHandlers("n2")
	assert.False(t, Has("n2"))

	DefaultEM.Clear()
	assert.False(t, Has("n1"))
}
