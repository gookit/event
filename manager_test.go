package event

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddEvent(t *testing.T) {
	e := NewBasic("evt1").Fill(nil, "inhere")
	AddEvent(e)

	assert.True(t, HasEvent("evt1"))
	assert.False(t, HasEvent("not-exist"))
}

func TestManager_Fire(t *testing.T) {
	buf := new(bytes.Buffer)
	fn := func(e Event) error {
		_, _ = fmt.Fprintf(buf, "event: %s", e.Name())
		return nil
	}

	On("evt1", ListenerFunc(fn))
	assert.True(t, HasListeners("evt1"))

	err := Fire("evt1")
	assert.NoError(t, err)

	assert.Equal(t, "event: evt1", buf.String())
}

func TestManager_FireEvent(t *testing.T) {
	buf := new(bytes.Buffer)

	evt1 := NewBasic("evt1").Fill(nil, "inhere")
	AddEvent(evt1)

	assert.True(t, HasEvent("evt1"))
	assert.False(t, HasEvent("not-exist"))

	On("evt1", ListenerFunc(func(e Event) error {
		_, _ = fmt.Fprintf(buf, "event: %s", e.Name())
		return nil
	}))

	assert.True(t, HasListeners("evt1"))
	assert.False(t, HasListeners("not-exist"))

	err := FireEvent(evt1)
	assert.NoError(t, err)

	assert.Equal(t, "event: evt1", buf.String())
}
