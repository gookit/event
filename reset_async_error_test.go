package event_test

import (
	"strings"
	"testing"

	"github.com/gookit/event"
)

func TestResetClearsPreviousAsyncError(t *testing.T) {
	for _, reset := range []struct {
		name  string
		apply func(*event.Manager)
	}{
		{"Reset", func(m *event.Manager) { m.Reset() }},
		{"Clear", func(m *event.Manager) { m.Clear() }},
	} {
		t.Run(reset.name, func(t *testing.T) {
			manager := event.NewManager("reusable", event.WithConsumerNum(1))
			manager.On("old", event.ListenerFunc(func(event.Event) error { panic("old failure") }))
			manager.Async("old", nil)
			if err := manager.CloseWait(); err == nil || !strings.Contains(err.Error(), "old failure") {
				t.Fatalf("missing initial failure: %v", err)
			}
			reset.apply(manager)
			if err := manager.Wait(); err != nil {
				t.Errorf("reset retained old error: %v", err)
			}
			called := 0
			manager.On("new", event.ListenerFunc(func(event.Event) error { called++; return nil }))
			manager.Async("new", nil)
			if err := manager.CloseWait(); err != nil {
				t.Errorf("successful new lifecycle retained error: %v", err)
			}
			if called != 1 {
				t.Errorf("new listener calls=%d", called)
			}
		})
	}
	t.Run("clean reset", func(t *testing.T) {
		manager := event.NewManager("clean", event.WithConsumerNum(1))
		manager.Reset()
		manager.On("new", event.ListenerFunc(func(event.Event) error { return nil }))
		manager.Async("new", nil)
		if err := manager.CloseWait(); err != nil {
			t.Error(err)
		}
	})
	t.Run("new failure", func(t *testing.T) {
		manager := event.NewManager("new error", event.WithConsumerNum(1))
		manager.On("old", event.ListenerFunc(func(event.Event) error { panic("old failure") }))
		manager.Async("old", nil)
		_ = manager.CloseWait()
		manager.Reset()
		manager.On("new", event.ListenerFunc(func(event.Event) error { panic("new failure") }))
		manager.Async("new", nil)
		if err := manager.CloseWait(); err == nil || !strings.Contains(err.Error(), "new failure") {
			t.Errorf("new failure lost: %v", err)
		}
	})
}
