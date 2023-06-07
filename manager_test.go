package event_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gookit/event"
	"github.com/stretchr/testify/assert"
)

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
