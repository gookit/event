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
	e.SetData(M{
		"arg0": "val0",
	})
	e.SetTarget("tgt")

	e.Add("arg1", "val1")

	assert.False(t, e.IsAborted())
	e.Abort(true)
	assert.True(t, e.IsAborted())

	assert.Equal(t, "n1", e.Name())
	assert.Equal(t, "tgt", e.Target())
	assert.Contains(t, e.Data(), "arg1")
	assert.Equal(t, "val0", e.Get("arg0"))
	assert.Equal(t, nil, e.Get("not-exist"))

	e.Set("arg1", "new val")
	assert.Equal(t, "new val", e.Get("arg1"))
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
	e := NewBasic("evt1", M{"k1": "inhere"})
	AddEvent(e)
	// add by AttachTo
	NewBasic("evt2", nil).AttachTo(DefaultEM)

	assert.False(t, e.IsAborted())
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
	assert.Panics(t, func() {
		On("++df", ListenerFunc(emptyListener), 0)
	})

	On("n1", ListenerFunc(emptyListener), Min)
	assert.Equal(t, 1, DefaultEM.ListenersCount("n1"))
	assert.Equal(t, 0, DefaultEM.ListenersCount("not-exist"))
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

	err := Fire("evt1", nil)
	assert.NoError(t, err)
	assert.Equal(t, "event: evt1", buf.String())

	NewBasic("evt2", nil).AttachTo(DefaultEM)
	On("evt2", ListenerFunc(func(e Event) error {
		assert.Equal(t, "evt2", e.Name())
		return nil
	}), AboveNormal)

	assert.True(t, HasListeners("evt2"))
	assert.NoError(t, Fire("evt2", nil))

	// clear all
	DefaultEM.Clear()
	assert.False(t, HasListeners("evt1"))
	assert.False(t, HasListeners("evt2"))
}

func TestFireEvent(t *testing.T) {
	buf := new(bytes.Buffer)

	evt1 := NewBasic("evt1", nil).Fill(nil, M{"n": "inhere"})
	AddEvent(evt1)

	assert.True(t, HasEvent("evt1"))
	assert.False(t, HasEvent("not-exist"))

	On("evt1", ListenerFunc(func(e Event) error {
		_, _ = fmt.Fprintf(buf, "event: %s, params: n=%s", e.Name(), e.Get("n"))
		return nil
	}), Normal)

	assert.True(t, HasListeners("evt1"))
	assert.False(t, HasListeners("not-exist"))

	err := FireEvent(evt1)
	assert.NoError(t, err)
	assert.Equal(t, "event: evt1, params: n=inhere", buf.String())
}

func TestMustFire(t *testing.T) {
	On("n1", ListenerFunc(func(e Event) error {
		return fmt.Errorf("an error")
	}), Max)
	On("n2", ListenerFunc(emptyListener), Min)

	assert.Panics(t, func() {
		MustFire("n1", nil)
	})

	assert.NotPanics(t, func() {
		MustFire("n2", nil)
	})
}

type testListener struct {
	userData string
}

func (l *testListener) Handle(e Event) error {
	e.Set("result", fmt.Sprintf("handled %s(%s)", e.Name(), l.userData))
	return nil
}

func TestManager_FireEvent(t *testing.T) {
	em := NewManager("test")

	e1 := NewBasic("e1", nil)
	em.AddEvent(e1)

	em.On("e1", &testListener{"HI"})

	err := em.FireEvent(e1)

	assert.NoError(t, err)
	assert.Equal(t, "handled e1(HI)", e1.Get("result"))

	em.Clear()
}
