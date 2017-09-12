package maputil

import (
	"testing"

	"github.com/bitrise-io/go-utils/testutil"
	"github.com/stretchr/testify/require"
)

func TestKeysOfStringInterfaceMap(t *testing.T) {
	t.Log("Empty map")
	{
		keys := KeysOfStringInterfaceMap(map[string]interface{}{})
		require.Equal(t, []string{}, keys)
		require.Equal(t, 0, len(keys))
	}

	t.Log("Nil map")
	{
		keys := KeysOfStringInterfaceMap(map[string]interface{}(nil))
		require.Equal(t, []string{}, keys)
		require.Equal(t, 0, len(keys))

		var nilMap map[string]interface{}
		keys = KeysOfStringInterfaceMap(nilMap)
		require.Equal(t, []string{}, keys)
		require.Equal(t, 0, len(keys))
	}

	t.Log("Single key")
	{
		keys := KeysOfStringInterfaceMap(map[string]interface{}{"a": "value"})
		require.Equal(t, 1, len(keys))
		require.Equal(t, []string{"a"}, keys)
	}

	t.Log("Multiple keys")
	{
		keys := KeysOfStringInterfaceMap(map[string]interface{}{"a": "value 1", "b": "value 2"})
		require.Equal(t, 2, len(keys))
		testutil.EqualSlicesWithoutOrder(t, []string{"a", "b"}, keys)
	}
}

func TestKeysOfStringStringMap(t *testing.T) {
	t.Log("Empty map")
	{
		keys := KeysOfStringStringMap(map[string]string{})
		require.Equal(t, []string{}, keys)
		require.Equal(t, 0, len(keys))
	}

	t.Log("Nil map")
	{
		keys := KeysOfStringStringMap(map[string]string(nil))
		require.Equal(t, []string{}, keys)
		require.Equal(t, 0, len(keys))

		var nilMap map[string]string
		keys = KeysOfStringStringMap(nilMap)
		require.Equal(t, []string{}, keys)
		require.Equal(t, 0, len(keys))
	}

	t.Log("Single key")
	{
		keys := KeysOfStringStringMap(map[string]string{"a": "value"})
		require.Equal(t, 1, len(keys))
		require.Equal(t, []string{"a"}, keys)
	}

	t.Log("Multiple keys")
	{
		keys := KeysOfStringStringMap(map[string]string{"a": "value 1", "b": "value 2"})
		require.Equal(t, 2, len(keys))
		testutil.EqualSlicesWithoutOrder(t, []string{"a", "b"}, keys)
	}
}

func TestCloneStringStringMap(t *testing.T) {
	t.Log("Should copy the map")
	{
		m1 := map[string]string{"key": "v1"}
		m1Clone := CloneStringStringMap(m1)
		require.Equal(t, map[string]string{"key": "v1"}, m1)
		require.Equal(t, map[string]string{"key": "v1"}, m1Clone)
		m1["key"] = "v2"
		require.Equal(t, map[string]string{"key": "v2"}, m1)
		require.Equal(t, map[string]string{"key": "v1"}, m1Clone)
	}

	t.Log("Should also work for empty map")
	{
		m1 := map[string]string{}
		m1Clone := CloneStringStringMap(m1)
		require.Equal(t, map[string]string{}, m1)
		require.Equal(t, map[string]string{}, m1Clone)
		m1["key"] = "v2"
		require.Equal(t, map[string]string{"key": "v2"}, m1)
		require.Equal(t, map[string]string{}, m1Clone)
	}
}

func TestMergeStringStringMap(t *testing.T) {
	t.Log("Merge maps - target should overwrite source's value, but should not modify either input map")
	{
		m1 := map[string]string{"key": "v1", "m1": "yes"}
		m2 := map[string]string{"key": "v2", "m2": "yes"}
		merged := MergeStringStringMap(m1, m2)
		require.Equal(t, map[string]string{"key": "v1", "m1": "yes"}, m1)
		require.Equal(t, map[string]string{"key": "v2", "m2": "yes"}, m2)
		require.Equal(t, map[string]string{"key": "v2", "m1": "yes", "m2": "yes"}, merged)
	}

	t.Log("Merge empty maps")
	{
		m1 := map[string]string{}
		m2 := map[string]string{}
		merged := MergeStringStringMap(m1, m2)
		require.Equal(t, map[string]string{}, merged)
	}

	t.Log("Merge maps where source is empty")
	{
		m1 := map[string]string{}
		m2 := map[string]string{"key": "v2", "m2": "yes"}
		merged := MergeStringStringMap(m1, m2)
		require.Equal(t, map[string]string{"key": "v2", "m2": "yes"}, merged)
	}

	t.Log("Merge maps where target is empty")
	{
		m1 := map[string]string{"key": "v1", "m1": "yes"}
		m2 := map[string]string{}
		merged := MergeStringStringMap(m1, m2)
		require.Equal(t, map[string]string{"key": "v1", "m1": "yes"}, merged)
	}
}
