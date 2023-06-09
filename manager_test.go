package event_test

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gookit/event"
	"github.com/stretchr/testify/assert"
)

func TestManager_FireEvent(t *testing.T) {
	em := event.NewManager("test")
	em.EnableLock = true

	e1 := event.NewBasic("e1", nil)
	em.AddEvent(e1)

	em.On("e1", &testListener{"HI"}, event.Min)
	em.On("e1", &testListener{"WEL"}, event.High)
	em.AddListener("e1", &testListener{"COM"}, event.BelowNormal)

	err := em.FireEvent(e1)
	assert.NoError(t, err)
	assert.Equal(t, "handled: e1(WEL) -> e1(COM) -> e1(HI)", e1.Get("result"))

	// not exist
	err = em.FireEvent(e1.SetName("e2"))
	assert.NoError(t, err)

	em.Clear()
}

func TestManager_FireEvent2(t *testing.T) {
	buf := new(bytes.Buffer)
	mgr := event.NewM("test")

	evt1 := event.New("evt1", nil).Fill(nil, event.M{"n": "inhere"})
	mgr.AddEvent(evt1)

	assert.True(t, mgr.HasEvent("evt1"))
	assert.False(t, mgr.HasEvent("not-exist"))

	mgr.On("evt1", event.ListenerFunc(func(e event.Event) error {
		_, _ = fmt.Fprintf(buf, "event: %s, params: n=%s", e.Name(), e.Get("n"))
		return nil
	}), event.Normal)

	assert.True(t, mgr.HasListeners("evt1"))
	assert.False(t, mgr.HasListeners("not-exist"))

	err := mgr.FireEvent(evt1)
	assert.NoError(t, err)
	assert.Equal(t, "event: evt1, params: n=inhere", buf.String())
	buf.Reset()

	mgr.On(event.Wildcard, event.ListenerFunc(func(e event.Event) error {
		buf.WriteString("|Wildcard handler")
		return nil
	}))

	err = mgr.FireEvent(evt1)
	assert.NoError(t, err)
	assert.Equal(t, "event: evt1, params: n=inhere|Wildcard handler", buf.String())
}

func TestManager_AsyncFire(t *testing.T) {
	em := event.NewManager("test")
	em.On("e1", event.ListenerFunc(func(e event.Event) error {
		assert.Equal(t, map[string]any{"k": "v"}, e.Data())
		e.Set("nk", "nv")
		return nil
	}))

	e1 := event.NewBasic("e1", event.M{"k": "v"})
	em.AsyncFire(e1)
	time.Sleep(time.Second / 10)
	assert.Equal(t, "nv", e1.Get("nk"))

	var wg sync.WaitGroup
	em.On("e2", event.ListenerFunc(func(e event.Event) error {
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
	em := event.NewManager("test")
	em.On("e1", event.ListenerFunc(func(e event.Event) error {
		assert.Equal(t, map[string]any{"k": "v"}, e.Data())
		e.Set("nk", "nv")
		return nil
	}))

	e1 := event.NewBasic("e1", event.M{"k": "v"})
	err := em.AwaitFire(e1)

	assert.NoError(t, err)
	assert.Contains(t, e1.Data(), "nk")
	assert.Equal(t, "nv", e1.Get("nk"))
}

func TestManager_AddSubscriber(t *testing.T) {
	em := event.NewManager("test")
	em.AddSubscriber(&testSubscriber{})

	assert.True(t, em.HasListeners("e1"))
	assert.True(t, em.HasListeners("e2"))
	assert.True(t, em.HasListeners("e3"))

	ers := em.FireBatch("e1", event.NewBasic("e2", nil))
	assert.Len(t, ers, 1)

	assert.Panics(t, func() {
		em.AddSubscriber(testSubscriber2{})
	})

	em.Clear()
}

func TestManager_ListenGroupEvent(t *testing.T) {
	em := event.NewManager("test")

	e1 := event.NewBasic("app.evt1", event.M{"buf": new(bytes.Buffer)})
	e1.AttachTo(em)

	l2 := event.ListenerFunc(func(e event.Event) error {
		e.Get("buf").(*bytes.Buffer).WriteString(" > 2 " + e.Name())
		return nil
	})
	l3 := event.ListenerFunc(func(e event.Event) error {
		e.Get("buf").(*bytes.Buffer).WriteString(" > 3 " + e.Name())
		return nil
	})

	em.On("app.evt1", event.ListenerFunc(func(e event.Event) error {
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
	em.On("app.*", event.ListenerFunc(func(e event.Event) error {
		return fmt.Errorf("an error")
	}))

	buf.Reset()
	err, e = em.Fire("app.evt1", nil)
	assert.Error(t, err)
	assert.Equal(t, "Hi > 1 app.evt1", buf.String())

	em.RemoveListeners("app.*")
	em.RemoveListener("", l3)
	em.On("app.*", l2) // re-add
	em.On("*", event.ListenerFunc(func(e event.Event) error {
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

func TestManager_Fire_WithWildcard(t *testing.T) {
	buf := new(bytes.Buffer)
	mgr := event.NewManager("test")

	const Event2FurcasTicketCreate = "kapal.furcas.ticket.create"

	handler := event.ListenerFunc(func(e event.Event) error {
		_, _ = fmt.Fprintf(buf, "%s-%s|", e.Name(), e.Get("user"))
		return nil
	})

	mgr.On("kapal.furcas.ticket.*", handler)
	mgr.On(Event2FurcasTicketCreate, handler)

	err, _ := mgr.Fire(Event2FurcasTicketCreate, event.M{"user": "inhere"})
	assert.NoError(t, err)
	assert.Equal(
		t,
		"kapal.furcas.ticket.create-inhere|kapal.furcas.ticket.create-inhere|",
		buf.String(),
	)
	buf.Reset()

	// add Wildcard listen
	mgr.On("*", handler)

	err, _ = mgr.Trigger(Event2FurcasTicketCreate, event.M{"user": "inhere"})
	assert.NoError(t, err)
	assert.Equal(
		t,
		"kapal.furcas.ticket.create-inhere|kapal.furcas.ticket.create-inhere|kapal.furcas.ticket.create-inhere|",
		buf.String(),
	)
}

func TestManager_Fire_usePathMode(t *testing.T) {
	buf := new(bytes.Buffer)
	em := event.NewManager("test", event.UsePathMode, event.EnableLock)

	em.Listen("db.user.*", event.ListenerFunc(func(e event.Event) error {
		_, _ = buf.WriteString("db.user.*|")
		return nil
	}))
	em.Listen("db.**", event.ListenerFunc(func(e event.Event) error {
		_, _ = buf.WriteString("db.**|")
		return nil
	}), 1)
	em.Listen("db.user.add", event.ListenerFunc(func(e event.Event) error {
		_, _ = buf.WriteString("db.user.add|")
		return nil
	}), 2)
	assert.True(t, em.HasListeners("db.user.*"))

	t.Run("fire case1", func(t *testing.T) {
		err, e := em.Fire("db.user.add", event.M{"user": "inhere"})
		assert.NoError(t, err)
		assert.Equal(t, "db.user.add", e.Name())
		assert.Equal(t, "inhere", e.Get("user"))
		str := buf.String()
		fmt.Println(str)
		assert.Contains(t, str, "db.**|")
		assert.Contains(t, str, "db.user.*|")
		assert.Contains(t, str, "db.user.add|")
		assert.True(t, strings.Count(str, "|") == 3)
	})
	buf.Reset()

	t.Run("fire case2", func(t *testing.T) {
		err, e := em.Fire("db.user.del", event.M{"user": "inhere"})
		assert.NoError(t, err)
		assert.Equal(t, "inhere", e.Get("user"))
		str := buf.String()
		fmt.Println(str)
		assert.Contains(t, str, "db.**|")
		assert.Contains(t, str, "db.user.*|")
		assert.True(t, strings.Count(str, "|") == 2)
	})
	buf.Reset()

	em.RemoveListeners("db.user.*")
	assert.False(t, em.HasListeners("db.user.*"))

	em.Listen("*", event.ListenerFunc(func(e event.Event) error {
		_, _ = buf.WriteString("*|")
		return nil
	}), 3)
	em.Listen("db.*.update", event.ListenerFunc(func(e event.Event) error {
		_, _ = buf.WriteString("db.*.update|")
		return nil
	}), 4)

	t.Run("fire case3", func(t *testing.T) {
		err, e := em.Fire("db.user.update", event.M{"user": "inhere"})
		assert.NoError(t, err)
		assert.Equal(t, "inhere", e.Get("user"))
		str := buf.String()
		fmt.Println(str)
		assert.Contains(t, str, "*|")
		assert.Contains(t, str, "db.**|")
		assert.Contains(t, str, "db.*.update|")
		assert.True(t, strings.Count(str, "|") == 3)
	})
	buf.Reset()

	t.Run("not-exist", func(t *testing.T) {
		err, e := em.Fire("not-exist", event.M{"user": "inhere"})
		assert.NoError(t, err)
		assert.Equal(t, "inhere", e.Get("user"))
		str := buf.String()
		fmt.Println(str)
		assert.Equal(t, "*|", str)
	})
}

func TestManager_Fire_AllNode(t *testing.T) {
	em := event.NewManager("test", event.UsePathMode, event.EnableLock)

	buf := new(bytes.Buffer)
	em.Listen("**.add", event.ListenerFunc(func(e event.Event) error {
		_, _ = buf.WriteString("**.add|")
		return nil
	}))

	err, e := em.Trigger("db.user.add", event.M{"user": "inhere"})
	assert.NoError(t, err)
	assert.Equal(t, "inhere", e.Get("user"))
	str := buf.String()
	assert.Equal(t, "**.add|", str)
}

func TestManager_FireC(t *testing.T) {
	em := event.NewManager("test", event.UsePathMode, event.EnableLock)
	defer func(em *event.Manager) {
		_ = em.Close()
	}(em)

	buf := new(bytes.Buffer)
	em.Listen("db.user.*", event.ListenerFunc(func(e event.Event) error {
		_, _ = buf.WriteString("db.user.*|")
		return nil
	}))
	em.Listen("db.**", event.ListenerFunc(func(e event.Event) error {
		_, _ = buf.WriteString("db.**|")
		return nil
	}), 1)

	em.Listen("db.user.add", event.ListenerFunc(func(e event.Event) error {
		_, _ = buf.WriteString("db.user.add|")
		return nil
	}), 2)

	assert.True(t, em.HasListeners("db.user.*"))

	em.FireC("db.user.add", event.M{"user": "inhere"})
	time.Sleep(time.Millisecond * 100) // wait for async

	str := buf.String()
	fmt.Println(str)
	assert.Contains(t, str, "db.**|")
	assert.Contains(t, str, "db.user.*|")
	assert.Contains(t, str, "db.user.add|")
	assert.True(t, strings.Count(str, "|") == 3)
	buf.Reset()

	em.WithOptions(func(o *event.Options) {
		o.ChannelSize = 0
		o.ConsumerNum = 0
	})
	em.Async("not-exist", event.M{"user": "inhere"})
	time.Sleep(time.Millisecond * 100) // wait for async
	assert.Empty(t, buf.String())
}

func TestManager_Wait(t *testing.T) {
	em := event.NewManager("test", event.UsePathMode, event.EnableLock)

	buf := new(bytes.Buffer)
	em.Listen("db.user.*", event.ListenerFunc(func(e event.Event) error {
		time.Sleep(time.Millisecond * 200)
		_, _ = buf.WriteString("db.user.*|")
		return nil
	}))
	assert.True(t, em.HasListeners("db.user.*"))

	em.Async("db.user.add", event.M{"user": "inhere"})
	assert.NoError(t, em.CloseWait())

	str := buf.String()
	fmt.Println(str)
	assert.Equal(t, "db.user.*|", str)
	buf.Reset()
}
