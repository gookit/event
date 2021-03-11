package event_test

import (
	"testing"

	"github.com/gookit/event"
	"github.com/stretchr/testify/assert"
)

type testNotify struct {}

func (notify *testNotify) Handle(e event.Event) error {
	isRun = true
	return nil
}

var isRun = false

// https://github.com/gookit/event/issues/8
func TestIssue_8(t *testing.T) {
	notify := testNotify{}

	event.On("*", &notify)
	err, _ := event.Fire("test_notify", event.M{})
	assert.Nil(t, err)
	assert.True(t, isRun)

	event.On("test_notify", &notify)
	err, _ = event.Fire("test_notify", event.M{})
	assert.Nil(t, err)
	assert.True(t, isRun)
}
