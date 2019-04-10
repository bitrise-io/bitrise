package gows

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_filteredEnvsList(t *testing.T) {
	t.Log("Remove second item")
	{
		inputEnvs := []string{"env1=value1", "env2=value two"}
		filteredEnvs := filteredEnvsList(inputEnvs, "env2")
		require.Equal(t, []string{"env1=value1"}, filteredEnvs)
		// should not change the input env list
		require.Equal(t, []string{"env1=value1", "env2=value two"}, inputEnvs)
	}

	t.Log("Remove first item")
	{
		inputEnvs := []string{"env1=value1", "env2=value two"}
		filteredEnvs := filteredEnvsList(inputEnvs, "env1")
		require.Equal(t, []string{"env2=value two"}, filteredEnvs)
	}

	t.Log("Key is not in the list")
	{
		inputEnvs := []string{"env1=value1", "env2=value two"}
		filteredEnvs := filteredEnvsList(inputEnvs, "not-in-list")
		require.Equal(t, []string{"env1=value1", "env2=value two"}, filteredEnvs)
	}

	t.Log("Empty input list")
	{
		inputEnvs := []string{"env1=value1", "env2=value two"}
		filteredEnvs := filteredEnvsList(inputEnvs, "env1")
		require.Equal(t, []string{"env2=value two"}, filteredEnvs)
	}
}
