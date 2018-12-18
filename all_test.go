package event

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var emptyListener = func(e Event) error {
	return nil
}

func TestEvent(t *testing.T) {
	e := &BasicEvent{}
	e.SetName("n1")
	e.SetData("arg0", "arg1")
	e.SetTarget("tgt")

	assert.False(t, e.Aborted())
	e.Abort()
	assert.True(t, e.Aborted())

	assert.Equal(t, "n1", e.Name())
	assert.Equal(t, "tgt", e.Target())
	assert.Equal(t, "arg0", e.Get(0))
	assert.Equal(t, nil, e.Get(10))
	assert.Equal(t, []interface{}{"arg0", "arg1"}, e.Data())
}

func TestAddEvent(t *testing.T) {
	DefaultEM.ClearEvents()

	// no name
	assert.Panics(t, func() {
		AddEvent(&BasicEvent{})
	})

	_, ok := GetEvent("evt1")
	assert.False(t, ok)

	// AddEvent
	e := NewBasic("evt1").Fill(nil, "inhere")
	AddEvent(e)
	e2 := (&BasicEvent{}).Init("evt2", nil)
	AddEvent(e2)

	assert.False(t, e.Aborted())
	assert.True(t, HasEvent("evt1"))
	assert.True(t, HasEvent("evt2"))
	assert.False(t, HasEvent("not-exist"))

	// GetEvent
	r1, ok := GetEvent("evt1")
	assert.True(t, ok)
	assert.Equal(t, e, r1)

	// DelEvent
	DefaultEM.DelEvent("evt2")
	assert.False(t, HasEvent("evt2"))

	// ClearEvents
	DefaultEM.ClearEvents()
	assert.False(t, HasEvent("evt1"))
}

func TestOn(t *testing.T) {
	assert.Panics(t, func() {
		On("", ListenerFunc(emptyListener), 0)
	})
	assert.Panics(t, func() {
		On("name", nil, 0)
	})

	On("n1", ListenerFunc(emptyListener), Min)
	assert.Equal(t, 1, DefaultEM.ListenersCount("n1"))
	assert.True(t, HasListeners("n1"))
	assert.False(t, HasListeners("name"))

	DefaultEM.ClearListeners("n1")
	assert.False(t, HasListeners("n1"))
}

func TestFire(t *testing.T) {
	buf := new(bytes.Buffer)
	fn := func(e Event) error {
		_, _ = fmt.Fprintf(buf, "event: %s", e.Name())
		return nil
	}

	On("evt1", ListenerFunc(fn), 0)
	On("evt1", ListenerFunc(emptyListener), High)
	assert.True(t, HasListeners("evt1"))

	err := Fire("evt1")
	assert.NoError(t, err)

	assert.Equal(t, "event: evt1", buf.String())
}

func TestFireEvent(t *testing.T) {
	buf := new(bytes.Buffer)

	evt1 := NewBasic("evt1").Fill(nil, "inhere")
	AddEvent(evt1)

	assert.True(t, HasEvent("evt1"))
	assert.False(t, HasEvent("not-exist"))

	On("evt1", ListenerFunc(func(e Event) error {
		_, _ = fmt.Fprintf(buf, "event: %s", e.Name())
		return nil
	}), Normal)

	assert.True(t, HasListeners("evt1"))
	assert.False(t, HasListeners("not-exist"))

	err := FireEvent(evt1)
	assert.NoError(t, err)

	assert.Equal(t, "event: evt1", buf.String())
}

func TestMustFire(t *testing.T) {
	On("n1", ListenerFunc(func(e Event) error {
		return fmt.Errorf("an error")
	}), Max)
	On("n2", ListenerFunc(emptyListener), Min)

	assert.Panics(t, func() {
		MustFire("n1")
	})

	assert.NotPanics(t, func() {
		MustFire("n2")
	})
}
