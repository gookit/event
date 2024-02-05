package event_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/gookit/event"
	"github.com/gookit/goutil/testutil/assert"
)

type testNotify struct{}

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

// https://github.com/gookit/event/issues/9
func TestIssues_9(t *testing.T) {
	evBus := event.NewManager("")
	eName := "evt1"

	f1 := makeFn(11)
	evBus.On(eName, f1)

	f2 := makeFn(22)
	evBus.On(eName, f2)
	assert.Equal(t, 2, evBus.ListenersCount(eName))

	f3 := event.ListenerFunc(func(e event.Event) error {
		// dump.Println(e.Name())
		return nil
	})
	evBus.On(eName, f3)
	assert.Equal(t, 3, evBus.ListenersCount(eName))

	evBus.RemoveListener(eName, f1) // DON'T REMOVE ALL !!!
	assert.Equal(t, 2, evBus.ListenersCount(eName))

	evBus.MustFire(eName, event.M{"arg0": "val0", "arg1": "val1"})
}

func makeFn(a int) event.ListenerFunc {
	return func(e event.Event) error {
		e.Set("val", a)
		// dump.Println(a, e.Name())
		return nil
	}
}

// https://github.com/gookit/event/issues/20
func TestIssues_20(t *testing.T) {
	buf := new(bytes.Buffer)
	mgr := event.NewManager("test")

	handler := event.ListenerFunc(func(e event.Event) error {
		_, _ = fmt.Fprintf(buf, "%s-%s|", e.Name(), e.Get("user"))
		return nil
	})

	mgr.On("app.user.*", handler)
	// ERROR: if not register "app.user.add", will not trigger "app.user.*"
	// mgr.On("app.user.add", handler)

	err, _ := mgr.Fire("app.user.add", event.M{"user": "INHERE"})
	assert.NoError(t, err)
	assert.Equal(t, "app.user.add-INHERE|", buf.String())

	// dump.P(buf.String())
}

// https://github.com/gookit/event/issues/61
// prepare: 此时设置 ConsumerNum = 10, 每个任务耗时1s, 触发100个任务
// expected: 10s左右执行完所有任务
// actual: 执行了 100s左右
func TestIssues_61(t *testing.T) {
	var em = event.NewManager("default", func(o *event.Options) {
		o.ConsumerNum = 10
		o.EnableLock = false
	})
	defer em.CloseWait()

	var listener event.ListenerFunc = func(e event.Event) error {
		time.Sleep(1 * time.Second)
		fmt.Println("event received!", e.Name(), "index", e.Get("arg0"))
		return nil
	}

	em.On("app.evt1", listener, event.Normal)

	for i := 0; i < 20; i++ {
		em.FireAsync(event.New("app.evt1", event.M{"arg0": i}))
	}

	fmt.Println("publish event finished!")
}
