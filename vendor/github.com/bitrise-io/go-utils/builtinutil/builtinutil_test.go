package builtinutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCastInterfaceToInterfaceSlice(t *testing.T) {
	// string
	casted, err := CastInterfaceToInterfaceSlice([]string{"a", "b", "c"})
	require.NoError(t, err)
	require.Equal(t, "a", casted[0])
	require.Equal(t, "b", casted[1])
	require.Equal(t, "c", casted[2])

	// int
	casted, err = CastInterfaceToInterfaceSlice([]int{3, 2, 1})
	require.NoError(t, err)
	require.Equal(t, 3, casted[0])
	require.Equal(t, 2, casted[1])
	require.Equal(t, 1, casted[2])

	// empty
	casted, err = CastInterfaceToInterfaceSlice([]string{})
	require.NoError(t, err)
	require.Equal(t, []interface{}{}, casted)
}

func TestDeepEqualSlices(t *testing.T) {
	{
		//
		s1 := []string{"a", "b"}
		interfaceS1, err := CastInterfaceToInterfaceSlice(s1)
		require.NoError(t, err)
		//
		s2 := []string{"b", "a"}
		interfaceS2, err := CastInterfaceToInterfaceSlice(s2)
		require.NoError(t, err)
		// equal, order doesn't matter
		require.True(t, DeepEqualSlices(interfaceS1, interfaceS2))

		// NOT equal
		s3 := []string{"b", "a", "c"}
		interfaceS3, err := CastInterfaceToInterfaceSlice(s3)
		require.NoError(t, err)
		require.False(t, DeepEqualSlices(interfaceS1, interfaceS3))

		// NOT equal - same length but element differs
		s4 := []string{"b", "x"}
		interfaceS4, err := CastInterfaceToInterfaceSlice(s4)
		require.NoError(t, err)
		require.False(t, DeepEqualSlices(interfaceS1, interfaceS4))
	}

	// empty
	{
		require.True(t, DeepEqualSlices([]interface{}{}, []interface{}{}))
	}

	t.Log("Custom struct type")
	{
		type TestType struct {
			StrKey   string
			IntSlice []int
		}
		s1 := []TestType{
			{StrKey: "key1", IntSlice: []int{1, 2}},
			{StrKey: "key2", IntSlice: []int{3, 4, 5}},
		}
		s2 := []TestType{
			{StrKey: "key2", IntSlice: []int{3, 4, 5}},
			{StrKey: "key1", IntSlice: []int{1, 2}},
		}

		t.Log(" Should be equal")
		interfaceS1, err := CastInterfaceToInterfaceSlice(s1)
		require.NoError(t, err)
		interfaceS2, err := CastInterfaceToInterfaceSlice(s2)
		require.NoError(t, err)
		require.True(t, DeepEqualSlices(interfaceS1, interfaceS2))

		t.Log(" Should NOT be equal")
		s3 := []TestType{
			{StrKey: "key2", IntSlice: []int{3, 4, 5}},
			{StrKey: "key3", IntSlice: []int{6, 7}},
		}
		interfaceS3, err := CastInterfaceToInterfaceSlice(s3)
		require.NoError(t, err)
		require.False(t, DeepEqualSlices(interfaceS1, interfaceS3))
	}
}
