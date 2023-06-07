package event_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

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
}
