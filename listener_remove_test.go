package event

import (
	"github.com/stretchr/testify/require"
	"testing"
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
	t.Run("", func(t *testing.T) {
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
		f3 := &testListenerCalc{bind: 33, owner: global}

		evBus.On(evName, f1)
		evBus.On(evName, f2)
		evBus.On(evName, f3)

		evBus.MustFire(evName, nil)
		require.Equal(t, global.n, 3)
		require.Equal(t, global.sum, 66) //11+22+33=66

		evBus.RemoveListener(evName, f1)
		evBus.MustFire(evName, nil)
		require.Equal(t, global.n, 5)
		require.Equal(t, global.sum, 121) // 66+22+33=121

		evBus.RemoveListener(evName, f3)
		evBus.MustFire(evName, nil)
		require.Equal(t, global.n, 6)
		require.Equal(t, global.sum, 143) // 121+22=143
	})
	t.Run("", func(t *testing.T) {
		global := &globalTestVal{}
		f1 := testListenerCalc{bind: 11, owner: global}
		f2 := testListenerCalc{bind: 22, owner: global}
		f2same := testListenerCalc{bind: 22, owner: global}
		f2copy := f2 // testListenerCalc{bind: 22, owner: global}

		evBus := NewManager("")
		const evName = "ev1"
		evBus.On(evName, f1)
		evBus.On(evName, f2)
		evBus.On(evName, f2same)
		evBus.On(evName, f2copy)

		evBus.MustFire(evName, nil)
		require.Equal(t, global.n, 4)
		require.Equal(t, global.sum, 77) //11+22+22+22=77

		evBus.RemoveListener(evName, f1)
		evBus.MustFire(evName, nil)
		require.Equal(t, global.n, 7)
		require.Equal(t, global.sum, 143) // 77+22+22+22=143

		evBus.RemoveListener(evName, f2)
		evBus.MustFire(evName, nil)
		require.Equal(t, global.n, 7)
		require.Equal(t, global.sum, 143) //
	})
	t.Run("", func(t *testing.T) {
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
		require.Equal(t, global.n, 4)
		require.Equal(t, global.sum, 77) //11+22+22+22=77

		evBus.RemoveListener(evName, f1)
		evBus.MustFire(evName, nil)
		require.Equal(t, global.n, 7)
		require.Equal(t, global.sum, 143) // 77+22+22+22=143

		evBus.RemoveListener(evName, f2)
		evBus.MustFire(evName, nil)
		require.Equal(t, global.n, 7)
		require.Equal(t, global.sum, 143) //
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

///
