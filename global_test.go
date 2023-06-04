package event_test

import (
	"testing"

	"github.com/gookit/event"
	"github.com/stretchr/testify/assert"
)

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
