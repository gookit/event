package event

import (
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

type globalTestVal struct {
	n   int
	sum int
}
type testListenerCalc struct {
	bind  int
	owner *globalTestVal
}

func (l testListenerCalc) Handle(e Event) error {
	l.owner.n++
	l.owner.sum += l.bind
	return nil
}

func Test_RemoveListener(t *testing.T) {
	t.Run("make func", func(t *testing.T) {
		global := &globalTestVal{}
		makeFn := func(a int) ListenerFunc {
			return func(e Event) error {
				global.n++
				global.sum += a
				return nil
			}
		}

		evBus := NewManager("")
		const evName = "ev1"

		f1 := makeFn(11)
		f2 := makeFn(22)
		f3 := makeFn(33)
		p4 := &testListenerCalc{bind: 44, owner: global}
		p5 := &testListenerCalc{bind: 55, owner: global}
		p6 := &testListenerCalc{bind: 66, owner: global}

		evBus.On(evName, f1)
		evBus.On(evName, f2)
		evBus.On(evName, f3)
		evBus.On(evName, p4)
		evBus.On(evName, p5)
		evBus.On(evName, p6)

		evBus.MustTrigger(evName, nil)
		assert.Equal(t, global.n, 6)
		assert.Equal(t, global.sum, 231) // 11+22+33+44+55+66=231

		evBus.RemoveListener(evName, f2)
		evBus.RemoveListener(evName, p5)
		evBus.MustFire(evName, nil)
		assert.Equal(t, global.n, 6+4)
		assert.Equal(t, global.sum, 385) // 231+11+33+44+66=385

		evBus.RemoveListener(evName, f1)
		evBus.RemoveListener(evName, f1) // not exist function.
		evBus.MustFire(evName, nil)
		assert.Equal(t, global.n, 6+4+3)
		assert.Equal(t, global.sum, 528) // 385+33+44+66=528

		evBus.RemoveListener(evName, p6)
		evBus.RemoveListener(evName, p6) // not exist function.
		evBus.MustFire(evName, nil)
		assert.Equal(t, global.n, 6+4+3+2)
		assert.Equal(t, global.sum, 605) // 528+33+44=605
	})

	t.Run("same value struct", func(t *testing.T) {
		global := &globalTestVal{}
		f1 := testListenerCalc{bind: 11, owner: global}
		f2 := testListenerCalc{bind: 22, owner: global}
		f2same := testListenerCalc{bind: 22, owner: global}
		f2copy := f2 // testListenerCalc{bind: 22, owner: global}

		evBus := NewManager("")
		const evName = "ev1"
		assert.Panics(t, func() {
			evBus.On(evName, f1)
			evBus.On(evName, f2)
			evBus.On(evName, f2same)
			evBus.On(evName, f2copy)
		})

		evBus.MustFire(evName, nil)
		assert.Equal(t, global.n, 0)
		assert.Equal(t, global.sum, 0) // 11+22+22+22=77

		evBus.RemoveListener(evName, f1)
		evBus.MustFire(evName, nil)
		assert.Equal(t, global.n, 0)
		assert.Equal(t, global.sum, 0) // 77+22+22+22=143

		evBus.RemoveListener(evName, f2)
		evBus.MustFire(evName, nil)
		assert.Equal(t, global.n, 0)
		assert.Equal(t, global.sum, 0)
	})

	t.Run("same func", func(t *testing.T) {
		global := &globalStatic

		f1 := ListenerFunc(testFuncCalc1)
		f2 := ListenerFunc(testFuncCalc2)
		f2same := ListenerFunc(testFuncCalc2)
		f2copy := f2

		evBus := NewManager("")
		const evName = "ev1"
		evBus.On(evName, f1)
		evBus.On(evName, f2)
		evBus.On(evName, f2same)
		evBus.On(evName, f2copy)

		evBus.MustFire(evName, nil)
		assert.Equal(t, global.n, 4)
		assert.Equal(t, global.sum, 77) // 11+22+22+22=77

		evBus.RemoveListener(evName, f1)
		evBus.MustFire(evName, nil)
		assert.Equal(t, global.n, 7)
		assert.Equal(t, global.sum, 143) // 77+22+22+22=143

		evBus.RemoveListener(evName, f2)
		evBus.MustFire(evName, nil)
		assert.Equal(t, global.n, 7)
		assert.Equal(t, global.sum, 143) //
	})
}

var globalStatic = globalTestVal{}

func testFuncCalc1(e Event) error {
	globalStatic.n++
	globalStatic.sum += 11
	return nil
}
func testFuncCalc2(e Event) error {
	globalStatic.n++
	globalStatic.sum += 22
	return nil
}

// /
