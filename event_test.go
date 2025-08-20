package event_test

import (
	"fmt"
	"testing"

	"github.com/gookit/event"
	"github.com/gookit/goutil/testutil/assert"
)

type testListener struct {
	userData string
}

func (l *testListener) Handle(e event.Event) error {
	if ret := e.Get("result"); ret != nil {
		str := ret.(string) + fmt.Sprintf(" -> %s(%s)", e.Name(), l.userData)
		e.Set("result", str)
	} else {
		e.Set("result", fmt.Sprintf("handled: %s(%s)", e.Name(), l.userData))
	}
	return nil
}

type testSubscriber struct {
	// ooo
}

func (s *testSubscriber) SubscribedEvents() map[string]any {
	return map[string]any{
		"e1": event.ListenerFunc(s.e1Handler),
		"e2": event.ListenerItem{
			Priority: event.AboveNormal,
			Listener: event.ListenerFunc(func(e event.Event) error {
				return fmt.Errorf("an error")
			}),
		},
		"e3": &testListener{},
	}
}

func (s *testSubscriber) e1Handler(e event.Event) error {
	e.Set("e1-key", "val1")
	return nil
}

type testSubscriber2 struct{}

func (s testSubscriber2) SubscribedEvents() map[string]any {
	return map[string]any{
		"e1": "invalid",
	}
}

type testEvent struct {
	event.ContextTrait
	name  string
	data  map[string]any
	abort bool
}

func (t *testEvent) Name() string {
	return t.name
}

func (t *testEvent) Get(key string) any {
	return t.data[key]
}

func (t *testEvent) Set(key string, val any) {
	t.data[key] = val
}

func (t *testEvent) Add(key string, val any) {
	t.data[key] = val
}

func (t *testEvent) Data() map[string]any {
	return t.data
}

func (t *testEvent) SetData(m event.M) event.Event {
	t.data = m
	return t
}

func (t *testEvent) Abort(b bool) {
	t.abort = b
}

func (t *testEvent) IsAborted() bool {
	return t.abort
}

func TestEvent(t *testing.T) {
	e := &event.BasicEvent{}
	e.SetName("n1")
	e.SetData(event.M{
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

	e1 := &event.BasicEvent{}
	e1.Set("k", "v")
	assert.Equal(t, "v", e1.Get("k"))
	assert.NotEmpty(t, e1.Clone())
}
