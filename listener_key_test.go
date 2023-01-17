package event

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getListenCompareKey(t *testing.T) {
	t.Run("ptr", func(t *testing.T) {
		makeFn := func(a int) ListenerFunc {
			return func(e Event) error {
				_ = a
				return nil
			}
		}

		f1 := makeFn(11)
		f2 := makeFn(22)
		f2same := makeFn(22)
		f2copy := f2
		require.NotEqual(t, getListenCompareKey(f1), getListenCompareKey(f2))
		require.NotEqual(t, getListenCompareKey(f2), getListenCompareKey(f2same))
		require.NotEqual(t, getListenCompareKey(f2same), getListenCompareKey(f2copy))
		require.Equal(t, getListenCompareKey(f2copy), getListenCompareKey(f2))

		f3ptr := &testListener{userData: "3"}
		ptr4 := &testListener{userData: "4"}
		ptr4Same := &testListener{userData: "4"}
		ptr4Copy := ptr4

		require.NotEqual(t, getListenCompareKey(f3ptr), getListenCompareKey(ptr4))
		require.NotEqual(t, getListenCompareKey(ptr4), getListenCompareKey(ptr4Same))
		require.NotEqual(t, getListenCompareKey(ptr4Same), getListenCompareKey(ptr4Copy))
		require.Equal(t, getListenCompareKey(ptr4Copy), getListenCompareKey(ptr4Copy))

	})

	t.Run("same value", func(t *testing.T) {

		f5struct := testListenerReadOnly{userData: "5"}
		struct6 := testListenerReadOnly{userData: "6"}
		struct6same := testListenerReadOnly{userData: "6"}
		struct6copy := struct6

		require.NotEqual(t, getListenCompareKey(f5struct), getListenCompareKey(struct6))
		require.Equal(t, getListenCompareKey(struct6), getListenCompareKey(struct6same))     // equal when struct data
		require.Equal(t, getListenCompareKey(struct6same), getListenCompareKey(struct6copy)) // equal when struct data
		require.Equal(t, getListenCompareKey(struct6copy), getListenCompareKey(struct6))

		f7func := ListenerFunc(testEmptyFunc1)
		func8 := ListenerFunc(testEmptyFunc2)
		func8same := ListenerFunc(testEmptyFunc2)
		func8copy := func8

		require.NotEqual(t, getListenCompareKey(f7func), getListenCompareKey(func8))
		require.Equal(t, getListenCompareKey(func8), getListenCompareKey(func8same))     // equal when struct data
		require.Equal(t, getListenCompareKey(func8same), getListenCompareKey(func8copy)) // equal when struct data
		require.Equal(t, getListenCompareKey(func8copy), getListenCompareKey(func8))

	})

}

func testEmptyFunc1(e Event) error {
	return nil
}
func testEmptyFunc2(e Event) error {
	return nil
}

type testListenerReadOnly struct {
	userData string
}

func (l testListenerReadOnly) Handle(e Event) error {
	return nil
}
