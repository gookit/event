package event_test

import (
	"errors"
	"github.com/gookit/event"
	"reflect"
	"testing"
)

func TestFireBatchIgnoresInvalidValuesAfterError(t *testing.T) {
	for _, bad := range []struct {
		name  string
		value any
	}{
		{"nil", nil}, {"number", 42}, {"slice", []string{"ignored"}}, {"struct", struct{}{}},
	} {
		t.Run(bad.name, func(t *testing.T) {
			manager := event.NewManager("batch")
			first, second := errors.New("first"), errors.New("second")
			var calls []string
			for _, name := range []string{"first", "second", "success"} {
				name := name
				manager.On(name, event.ListenerFunc(func(e event.Event) error {
					calls = append(calls, e.Name())
					if name == "first" {
						return first
					}
					if name == "second" {
						return second
					}
					return nil
				}))
			}
			got := manager.FireBatch(bad.value, "first", bad.value, event.New("second", nil), bad.value, "success", bad.value)
			if len(got) != 2 || got[0] != first || got[1] != second {
				t.Fatalf("got errors %v, want [%v %v]", got, first, second)
			}
			if !reflect.DeepEqual(calls, []string{"first", "second", "success"}) {
				t.Fatalf("listener calls: %v", calls)
			}
		})
	}
}

func TestFireBatchValidValuesAndErrors(t *testing.T) {
	manager := event.NewManager("batch")
	fail := errors.New("failure")
	manager.On("fail", event.ListenerFunc(func(event.Event) error { return fail }))
	if got := manager.FireBatch(nil, 42, struct{}{}); len(got) != 0 {
		t.Fatalf("invalid-only batch: %v", got)
	}
	if got := manager.FireBatch("unhandled", nil); len(got) != 0 {
		t.Fatalf("successful batch: %v", got)
	}
	if got := manager.FireBatch("fail", event.New("fail", nil)); len(got) != 2 || got[0] != fail || got[1] != fail {
		t.Fatalf("actual failures: %v", got)
	}
	if got := manager.FireBatch("fail", ""); len(got) != 2 || got[0] != fail || got[1] == nil {
		t.Fatalf("invalid event name: %v", got)
	}
}
