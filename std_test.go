package event_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/gookit/event"
	"github.com/stretchr/testify/assert"
)

var emptyListener = func(e event.Event) error {
	return nil
}

func TestAddEvent(t *testing.T) {
	defer event.Reset()
	event.Std().RemoveEvents()

	// no name
	assert.Panics(t, func() {
		event.AddEvent(&event.BasicEvent{})
	})

	_, ok := event.GetEvent("evt1")
	assert.False(t, ok)

	// event.AddEvent
	e := event.NewBasic("evt1", event.M{"k1": "inhere"})
	event.AddEvent(e)
	// add by AttachTo
	event.NewBasic("evt2", nil).AttachTo(event.Std())

	assert.False(t, e.IsAborted())
	assert.True(t, event.HasEvent("evt1"))
	assert.True(t, event.HasEvent("evt2"))
	assert.False(t, event.HasEvent("not-exist"))

	// event.GetEvent
	r1, ok := event.GetEvent("evt1")
	assert.True(t, ok)
	assert.Equal(t, e, r1)

	// RemoveEvent
	event.Std().RemoveEvent("evt2")
	assert.False(t, event.HasEvent("evt2"))

	// RemoveEvents
	event.Std().RemoveEvents()
	assert.False(t, event.HasEvent("evt1"))
}

func TestAddEventFc(t *testing.T) {
	// clear all
	event.Reset()
	event.AddEvent(&testEvent{
		name: "evt1",
	})
	assert.True(t, event.HasEvent("evt1"))

	event.AddEventFc("test", func() event.Event {
		return event.NewBasic("test", nil)
	})

	assert.True(t, event.HasEvent("test"))
}

func TestFire(t *testing.T) {
	buf := new(bytes.Buffer)
	fn := func(e event.Event) error {
		_, _ = fmt.Fprintf(buf, "event: %s", e.Name())
		return nil
	}

	event.On("evt1", event.ListenerFunc(fn), 0)
	event.On("evt1", event.ListenerFunc(emptyListener), event.High)
	assert.True(t, event.HasListeners("evt1"))

	err, e := event.Fire("evt1", nil)
	assert.NoError(t, err)
	assert.Equal(t, "evt1", e.Name())
	assert.Equal(t, "event: evt1", buf.String())

	event.NewBasic("evt2", nil).AttachTo(event.Std())
	event.On("evt2", event.ListenerFunc(func(e event.Event) error {
		assert.Equal(t, "evt2", e.Name())
		assert.Equal(t, "v", e.Get("k"))
		return nil
	}), event.AboveNormal)

	assert.True(t, event.HasListeners("evt2"))
	err, e = event.Trigger("evt2", event.M{"k": "v"})
	assert.NoError(t, err)
	assert.Equal(t, "evt2", e.Name())
	assert.Equal(t, map[string]any{"k": "v"}, e.Data())

	// clear all
	event.Reset()
	assert.False(t, event.HasListeners("evt1"))
	assert.False(t, event.HasListeners("evt2"))

	err, e = event.Fire("not-exist", nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, e)
}

func TestAddSubscriber(t *testing.T) {
	event.AddSubscriber(&testSubscriber{})

	assert.True(t, event.HasListeners("e1"))
	assert.True(t, event.HasListeners("e2"))
	assert.True(t, event.HasListeners("e3"))

	ers := event.FireBatch("e1", event.NewBasic("e2", nil))
	assert.Len(t, ers, 1)

	assert.Panics(t, func() {
		event.Subscribe(testSubscriber2{})
	})

	event.Reset()
}

func TestFireEvent(t *testing.T) {
	defer event.Reset()
	buf := new(bytes.Buffer)

	evt1 := event.NewBasic("evt1", nil).Fill(nil, event.M{"n": "inhere"})
	event.AddEvent(evt1)

	assert.True(t, event.HasEvent("evt1"))
	assert.False(t, event.HasEvent("not-exist"))

	event.Listen("evt1", event.ListenerFunc(func(e event.Event) error {
		_, _ = fmt.Fprintf(buf, "event: %s, params: n=%s", e.Name(), e.Get("n"))
		return nil
	}), event.Normal)

	assert.True(t, event.HasListeners("evt1"))
	assert.False(t, event.HasListeners("not-exist"))

	err := event.FireEvent(evt1)
	assert.NoError(t, err)
	assert.Equal(t, "event: evt1, params: n=inhere", buf.String())
	buf.Reset()

	err = event.TriggerEvent(evt1)
	assert.NoError(t, err)
	assert.Equal(t, "event: evt1, params: n=inhere", buf.String())
	buf.Reset()

	event.AsyncFire(evt1)
	time.Sleep(time.Second)
	assert.Equal(t, "event: evt1, params: n=inhere", buf.String())
}

func TestAsync(t *testing.T) {
	event.Reset()
	event.Config(event.UsePathMode)

	buf := new(bytes.Buffer)
	event.On("test", event.ListenerFunc(func(e event.Event) error {
		buf.WriteString("test:")
		buf.WriteString(e.Get("key").(string))
		buf.WriteString("|")
		return nil
	}))

	event.Async("test", event.M{"key": "val1"})
	te := &testEvent{name: "test", data: event.M{"key": "val2"}}
	event.FireAsync(te)

	assert.NoError(t, event.CloseWait())
	s := buf.String()
	assert.Contains(t, s, "test:val1|")
	assert.Contains(t, s, "test:val2|")
}

func TestFire_point_at_end(t *testing.T) {
	// clear all
	event.Reset()
	event.Listen("db.*", event.ListenerFunc(func(e event.Event) error {
		e.Set("key", "val")
		return nil
	}))

	err, e := event.Fire("db.add", nil)
	assert.NoError(t, err)
	assert.Equal(t, "val", e.Get("key"))

	err, e = event.Fire("db", nil)
	assert.NotEmpty(t, e)
	assert.NoError(t, err)

	err, e = event.Fire("db.", nil)
	assert.NotEmpty(t, e)
	assert.NoError(t, err)
}

func TestFire_notExist(t *testing.T) {
	// clear all
	event.Reset()

	err, e := event.Fire("not-exist", nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, e)
}

func TestMustFire(t *testing.T) {
	defer event.Reset()

	event.On("n1", event.ListenerFunc(func(e event.Event) error {
		return fmt.Errorf("an error")
	}), event.Max)
	event.On("n2", event.ListenerFunc(emptyListener), event.Min)

	assert.Panics(t, func() {
		_ = event.MustFire("n1", nil)
	})

	assert.NotPanics(t, func() {
		_ = event.MustTrigger("n2", nil)
	})
}

func TestOn(t *testing.T) {
	defer event.Reset()

	assert.Panics(t, func() {
		event.On("", event.ListenerFunc(emptyListener), 0)
	})
	assert.Panics(t, func() {
		event.On("name", nil, 0)
	})
	assert.Panics(t, func() {
		event.On("++df", event.ListenerFunc(emptyListener), 0)
	})

	std := event.Std()
	event.On("n1", event.ListenerFunc(emptyListener), event.Min)
	assert.Equal(t, 1, std.ListenersCount("n1"))
	assert.Equal(t, 0, std.ListenersCount("not-exist"))
	assert.True(t, event.HasListeners("n1"))
	assert.False(t, event.HasListeners("name"))

	assert.NotEmpty(t, std.Listeners())
	assert.NotEmpty(t, std.ListenersByName("n1"))

	std.RemoveListeners("n1")
	assert.False(t, event.HasListeners("n1"))
}
