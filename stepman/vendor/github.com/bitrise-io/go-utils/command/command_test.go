package command

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCommandSlice(t *testing.T) {
	t.Log("it fails if slice empty")
	{
		cmd, err := NewFromSlice([]string{})
		require.Error(t, err)
		require.Equal(t, (*Model)(nil), cmd)
	}

	t.Log("it creates cmd if cmdSlice has 1 element")
	{
		_, err := NewFromSlice([]string{"ls"})
		require.NoError(t, err)
	}

	t.Log("it creates cmd if cmdSlice has multiple elements")
	{
		_, err := NewFromSlice([]string{"ls", "-a", "-l", "-h"})
		require.NoError(t, err)
	}
}

func TestNewWithParams(t *testing.T) {
	t.Log("it fails if params empty")
	{
		cmd, err := NewWithParams()
		require.Error(t, err)
		require.Equal(t, (*Model)(nil), cmd)
	}

	t.Log("it creates cmd if params has 1 element")
	{
		_, err := NewWithParams("ls")
		require.NoError(t, err)
	}

	t.Log("it creates cmd if params has multiple elements")
	{
		_, err := NewWithParams("ls", "-a", "-l", "-h")
		require.NoError(t, err)
	}
}
