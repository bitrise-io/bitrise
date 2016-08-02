package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseJSONParams(t *testing.T) {
	t.Log("it parses json string-string map")
	{
		params, err := parseJSONParams(`{"config":"bitrise.yml","inventory":".secrets.bitrise.yml", "workflow":"primary"}`)
		require.NoError(t, err)
		require.Equal(t, 3, len(params))
		require.Equal(t, "bitrise.yml", params[ConfigKey])
		require.Equal(t, ".secrets.bitrise.yml", params[InventoryKey])
		require.Equal(t, "primary", params[WorkflowKey])
	}

	t.Log("null is a valid json string-string map")
	{
		params, err := parseJSONParams(`null`)
		require.NoError(t, err)
		require.Equal(t, 0, len(params))
	}

	t.Log("it returns error for not json string-string map")
	{
		params, err := parseJSONParams("primary")
		require.Error(t, err)
		require.Equal(t, 0, len(params))

		params, err = parseJSONParams(`{"number": 1}`)
		require.Error(t, err)
		require.Equal(t, 0, len(params))

		params, err = parseJSONParams(`{"array": [1,2,3]}`)
		require.Error(t, err)
		require.Equal(t, 0, len(params))

		params, err = parseJSONParams(`{"boolean":true}`)
		require.Error(t, err)
		require.Equal(t, 0, len(params))
	}
}

func TestParseJSONParamsBase64(t *testing.T) {
	t.Log("it parses base 64 encoded json string-string map")
	{
		params, err := parseJSONParamsBase64(`eyJjb25maWciOiJiaXRyaXNlLnltbCIsImludmVudG9yeSI6Ii5zZWNyZXRzLmJpdHJpc2UueW1sIiwgIndvcmtmbG93IjoicHJpbWFyeSJ9`)
		require.NoError(t, err)
		require.Equal(t, 3, len(params))
		require.Equal(t, "bitrise.yml", params[ConfigKey])
		require.Equal(t, ".secrets.bitrise.yml", params[InventoryKey])
		require.Equal(t, "primary", params[WorkflowKey])
	}

	t.Log("it returns error for not base 64 encoded json string-string map")
	{
		params, err := parseJSONParamsBase64(`{"config":"bitrise.yml","inventory":".secrets.bitrise.yml", "workflow":"primary"}`)
		require.Error(t, err)
		require.Equal(t, 0, len(params))
	}
}
