package event_test

import (
	"testing"

	"github.com/gookit/event"
)

func BenchmarkManager_Fire_no_listener(b *testing.B) {
	em := event.NewManager("test")
	em.On("app.up", event.ListenerFunc(func(e event.Event) error {
		return nil
	}))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = em.Fire("app.up", nil)
	}
}

func BenchmarkManager_Fire_normal(b *testing.B) {
	em := event.NewManager("test")
	em.On("app.up", event.ListenerFunc(func(e event.Event) error {
		return nil
	}))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = em.Fire("app.up", nil)
	}
}

func BenchmarkManager_Fire_wildcard(b *testing.B) {
	em := event.NewManager("test")
	em.On("app.*", event.ListenerFunc(func(e event.Event) error {
		return nil
	}))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = em.Fire("app.up", nil)
	}
}
