package event

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

var emptyListener = func(e Event) error {
	return nil
}

type testListener struct {
	userData string
}

func (l *testListener) Handle(e Event) error {
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

func (s *testSubscriber) SubscribeEvents() map[string]interface{} {
	return map[string]interface{}{
		"e1": ListenerFunc(s.e1Handler),
		"e2": ListenerItem{
			Priority: AboveNormal,
			Listener: ListenerFunc(func(e Event) error {
				return fmt.Errorf("an error")
			}),
		},
		"e3": &testListener{},
	}
}

func (s *testSubscriber) e1Handler(e Event) error {
	e.Set("e1-key", "val1")
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

	e1 := &BasicEvent{}
	e1.Set("k", "v")
	assert.Equal(t, "v", e1.Get("k"))
}

func TestAddEvent(t *testing.T) {
	DefaultEM.RemoveEvents()

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

	// RemoveEvent
	DefaultEM.RemoveEvent("evt2")
	assert.False(t, HasEvent("evt2"))

	// RemoveEvents
	DefaultEM.RemoveEvents()
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

	DefaultEM.RemoveListeners("n1")
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

	err, e := Fire("evt1", nil)
	assert.NoError(t, err)
	assert.Equal(t, "evt1", e.Name())
	assert.Equal(t, "event: evt1", buf.String())

	NewBasic("evt2", nil).AttachTo(DefaultEM)
	On("evt2", ListenerFunc(func(e Event) error {
		assert.Equal(t, "evt2", e.Name())
		assert.Equal(t, "v", e.Get("k"))
		return nil
	}), AboveNormal)

	assert.True(t, HasListeners("evt2"))
	err, e = Fire("evt2", M{"k": "v"})
	assert.NoError(t, err)
	assert.Equal(t, "evt2", e.Name())
	assert.Equal(t, map[string]interface{}{"k": "v"}, e.Data())

	// clear all
	DefaultEM.Clear()
	assert.False(t, HasListeners("evt1"))
	assert.False(t, HasListeners("evt2"))

	err, e = Fire("not-exist", nil)
	assert.NoError(t, err)
	assert.Nil(t, e)
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
		_, _ = MustFire("n1", nil)
	})

	assert.NotPanics(t, func() {
		_, _ = MustFire("n2", nil)
	})
}

func TestManager_FireEvent(t *testing.T) {
	em := NewManager("test")

	e1 := NewBasic("e1", nil)
	em.AddEvent(e1)

	em.On("e1", &testListener{"HI"}, Min)
	em.On("e1", &testListener{"WEL"}, High)
	em.AddListener("e1", &testListener{"COM"}, BelowNormal)

	err := em.FireEvent(e1)
	assert.NoError(t, err)
	assert.Equal(t, "handled: e1(WEL) -> e1(COM) -> e1(HI)", e1.Get("result"))

	// not exist
	err = em.FireEvent(e1.SetName("e2"))
	assert.NoError(t, err)

	em.Clear()
}

func TestListenGroupEvent(t *testing.T) {
	em := NewManager("test")

	e1 := NewBasic("app.evt1", M{"buf": new(bytes.Buffer)})
	e1.AttachTo(em)

	l2 := ListenerFunc(func(e Event) error {
		e.Get("buf").(*bytes.Buffer).WriteString(" > 2 " + e.Name())
		return nil
	})
	l3 := ListenerFunc(func(e Event) error {
		e.Get("buf").(*bytes.Buffer).WriteString(" > 3 " + e.Name())
		return nil
	})

	em.On("app.evt1", ListenerFunc(func(e Event) error {
		e.Get("buf").(*bytes.Buffer).WriteString("Hi > 1 " + e.Name())
		return nil
	}))
	em.On("app.*", l2)
	em.On("*", l3)

	buf := e1.Get("buf").(*bytes.Buffer)
	err, e := em.Fire("app.evt1", nil)
	assert.NoError(t, err)
	assert.Equal(t, e1, e)
	assert.Equal(t, "Hi > 1 app.evt1 > 2 app.evt1 > 3 app.evt1", buf.String())

	em.RemoveListener("app.*", l2)
	assert.Len(t, em.ListenedNames(), 2)
	em.On("app.*", ListenerFunc(func(e Event) error {
		return fmt.Errorf("an error")
	}))

	buf.Reset()
	err, e = em.Fire("app.evt1", nil)
	assert.Error(t, err)
	assert.Equal(t, "Hi > 1 app.evt1", buf.String())

	em.RemoveListeners("app.*")
	em.RemoveListener("", l3)
	em.On("app.*", l2) // re-add
	em.On("*", ListenerFunc(func(e Event) error {
		return fmt.Errorf("an error")
	}))
	assert.Len(t, em.ListenedNames(), 3)

	buf.Reset()
	err, e = em.Trigger("app.evt1", nil)
	assert.Error(t, err)
	assert.Equal(t, e1, e)
	assert.Equal(t, "Hi > 1 app.evt1 > 2 app.evt1", buf.String())

	em.RemoveListener("", nil)

	// clear
	em.Clear()
	buf.Reset()
}

func TestManager_AsyncFire(t *testing.T) {
	em := NewManager("test")
	em.On("e1", ListenerFunc(func(e Event) error {
		assert.Equal(t, map[string]interface{}{"k": "v"}, e.Data())
		e.Set("nk", "nv")
		return nil
	}))

	e1 := NewBasic("e1", M{"k": "v"})
	em.AsyncFire(e1)
	time.Sleep(time.Second / 10)
	assert.Equal(t, "nv", e1.Get("nk"))

	var wg sync.WaitGroup
	em.On("e2", ListenerFunc(func(e Event) error {
		defer wg.Done()
		assert.Equal(t, "v", e.Get("k"))
		return nil
	}))

	wg.Add(1)
	em.AsyncFire(e1.SetName("e2"))
	wg.Wait()

	em.Clear()
}

func TestManager_AwaitFire(t *testing.T) {
	em := NewManager("test")
	em.On("e1", ListenerFunc(func(e Event) error {
		assert.Equal(t, map[string]interface{}{"k": "v"}, e.Data())
		e.Set("nk", "nv")
		return nil
	}))

	e1 := NewBasic("e1", M{"k": "v"})
	err := em.AwaitFire(e1)

	assert.NoError(t, err)
	assert.Contains(t, e1.Data(), "nk")
	assert.Equal(t, "nv", e1.Get("nk"))
}

type testSubscriber2 struct{}

func (s testSubscriber2) SubscribeEvents() map[string]interface{} {
	return map[string]interface{}{
		"e1": "invalid",
	}
}

func TestManager_AddSubscriber(t *testing.T) {
	em := NewManager("test")
	em.AddSubscriber(&testSubscriber{})

	assert.True(t, em.HasListeners("e1"))
	assert.True(t, em.HasListeners("e2"))
	assert.True(t, em.HasListeners("e3"))

	ers := em.FireBatch("e1", NewBasic("e2", nil))
	assert.Len(t, ers, 1)

	assert.Panics(t, func() {
		em.AddSubscriber(testSubscriber2{})
	})

	em.Clear()
}
