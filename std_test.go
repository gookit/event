package event_test

import (
	"testing"

	"github.com/gookit/event"
	"github.com/stretchr/testify/assert"
)

var emptyListener = func(e event.Event) error {
	return nil
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
