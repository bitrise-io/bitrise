package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvListToMap(t *testing.T) {
	m, err := envListToMap([]string{"TEST=test"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"TEST": "test"}, m)
}

func TestExpandEnv(t *testing.T) {
	t.Log("expands from empty externals")
	{
		external := []string{}
		m, err := ExpandEnv("KEY", external)
		require.NoError(t, err)
		require.Equal(t, "", m)
	}

	t.Log("expands something else")
	{
		external := []string{"EXTERNAL_KEY=some"}
		m, err := ExpandEnv("KEY", external)
		require.NoError(t, err)
		require.Equal(t, "", m)
	}

	t.Log("expands single")
	{
		external := []string{"EXTERNAL_KEY=some"}
		m, err := ExpandEnv("EXTERNAL_KEY", external)
		require.NoError(t, err)
		require.Equal(t, "some", m)
	}

	t.Log("expands same key multiple time - latest")
	{
		external := []string{"EXTERNAL_KEY=some", "EXTERNAL_KEY=other", "EXTERNAL_KEY=value"}
		m, err := ExpandEnv("EXTERNAL_KEY", external)
		require.NoError(t, err)
		require.Equal(t, "value", m)
	}

	t.Log("expands inherited")
	{
		external := []string{"EXTERNAL_KEY=some", "KEY=${EXTERNAL_KEY} value"}
		m, err := ExpandEnv("KEY", external)
		require.NoError(t, err)
		require.Equal(t, "some value", m)
	}

	t.Log("expands inherited and multiple times")
	{
		external := []string{"EXTERNAL_KEY=some", "KEY=${EXTERNAL_KEY} value", "KEY=${EXTERNAL_KEY} value 2"}
		m, err := ExpandEnv("KEY", external)
		require.NoError(t, err)
		require.Equal(t, "some value 2", m)
	}
}
