package freezable

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_String(t *testing.T) {
	var freezableObj String

	t.Log("Empty")
	{
		require.Equal(t, "", freezableObj.Get())
		require.Equal(t, false, freezableObj.IsFrozen())
	}

	t.Log("Set")
	{
		err := freezableObj.Set("initial value")
		require.NoError(t, err)
		require.Equal(t, "initial value", freezableObj.Get())
		require.Equal(t, "initial value", freezableObj.String())
		require.Equal(t, false, freezableObj.IsFrozen())
	}

	t.Log("fmt - Stringer")
	{
		require.Equal(t, "initial value", fmt.Sprintf("%s", freezableObj))
	}

	t.Log("Re-Set (not yet frozen)")
	{
		err := freezableObj.Set("frozen value")
		require.NoError(t, err)
		require.Equal(t, "frozen value", freezableObj.Get())
		require.Equal(t, false, freezableObj.IsFrozen())
	}

	t.Log("Freeze")
	{
		freezableObj.Freeze()
		require.Equal(t, "frozen value", freezableObj.Get())
		require.Equal(t, true, freezableObj.IsFrozen())
	}
	t.Log("Try to change - should error")
	{
		err := freezableObj.Set("something else")
		require.EqualError(t, err, "freezable.String: Object is already frozen. (Current value: frozen value) (New value was: something else)")
	}
}

func Test_StringSlice(t *testing.T) {
	var freezableObj StringSlice

	t.Log("Empty")
	{
		require.Equal(t, []string{}, freezableObj.Get())
		require.Equal(t, false, freezableObj.IsFrozen())
	}

	t.Log("Set")
	{
		err := freezableObj.Set([]string{"initial value"})
		require.NoError(t, err)
		require.Equal(t, []string{"initial value"}, freezableObj.Get())
		require.Equal(t, false, freezableObj.IsFrozen())
	}

	t.Log("fmt - Stringer")
	{
		require.Equal(t, "[initial value]", fmt.Sprintf("%s", freezableObj))
		require.Equal(t, "[initial value]", freezableObj.String())
	}

	t.Log("Re-Set (not yet frozen)")
	{
		err := freezableObj.Set([]string{"frozen value"})
		require.NoError(t, err)
		require.Equal(t, []string{"frozen value"}, freezableObj.Get())
		require.Equal(t, false, freezableObj.IsFrozen())
	}

	t.Log("Freeze")
	{
		freezableObj.Freeze()
		require.Equal(t, []string{"frozen value"}, freezableObj.Get())
		require.Equal(t, true, freezableObj.IsFrozen())
	}

	t.Log("Try to change - should error")
	{
		err := freezableObj.Set([]string{"something else"})
		require.EqualError(t, err, "freezable.StringSlice: Object is already frozen. (Current value: [frozen value]) (New value was: [something else])")
	}
}
