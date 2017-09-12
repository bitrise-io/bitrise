package regexputil

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNamedFindStringSubmatch(t *testing.T) {
	t.Log("Both the name and age group are required")
	rexp := regexp.MustCompile(`(?P<name>[a-zA-Z]+) (?P<age>[0-9]+)`)

	t.Log("Simple name+age example")
	{
		results, isFound := NamedFindStringSubmatch(rexp, "MyName 42")
		require.Equal(t, true, isFound)
		require.Equal(t, map[string]string{
			"name": "MyName",
			"age":  "42",
		}, results)
	}

	t.Log("Includes an additional name at the end")
	{
		results, isFound := NamedFindStringSubmatch(rexp, "MyName 42 AnotherName")
		require.Equal(t, true, isFound)
		require.Equal(t, map[string]string{
			"name": "MyName",
			"age":  "42",
		}, results)
	}

	t.Log("Includes an additional name at the start")
	{
		results, isFound := NamedFindStringSubmatch(rexp, "AnotherName MyName 42")
		require.Equal(t, true, isFound)
		require.Equal(t, map[string]string{
			"name": "MyName",
			"age":  "42",
		}, results)
	}

	t.Log("Missing name group - should error")
	{
		results, isFound := NamedFindStringSubmatch(rexp, " 42")
		require.Equal(t, false, isFound)
		require.Equal(t, map[string]string(nil), results)
	}

	t.Log("Missing age group - should error")
	{
		results, isFound := NamedFindStringSubmatch(rexp, "MyName ")
		require.Equal(t, false, isFound)
		require.Equal(t, map[string]string(nil), results)
	}

	t.Log("Missing both groups - should error")
	{
		results, isFound := NamedFindStringSubmatch(rexp, "")
		require.Equal(t, false, isFound)
		require.Equal(t, map[string]string(nil), results)
	}

	t.Log("Optional name part")
	rexp = regexp.MustCompile(`(?P<name>[a-zA-Z]*) (?P<age>[0-9]+)`)

	t.Log("Name can now be empty - but should be included in the result!")
	{
		results, isFound := NamedFindStringSubmatch(rexp, " 42")
		require.Equal(t, true, isFound)
		require.Equal(t, map[string]string{
			"name": "",
			"age":  "42",
		}, results)
	}
}
