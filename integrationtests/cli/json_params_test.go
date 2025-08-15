//go:build linux_and_mac
// +build linux_and_mac

package cli

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_JsonParams(t *testing.T) {
	configPth := "json_params_test_bitrise.yml"

	t.Log("run test")
	{
		config := map[string]interface{}{
			"config":   configPth,
			"workflow": "json_params_test_target",
		}

		cmd := command.New(testhelpers.BinPath(), "run", "--json-params", testhelpers.ToJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("run test - param override")
	{
		config := map[string]interface{}{
			"config":   configPth,
			"workflow": "exit_code_test_fail",
		}

		cmd := command.New(testhelpers.BinPath(), "run", "--json-params", testhelpers.ToJSON(t, config), "--workflow", "json_params_test_target")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("trigger test")
	{
		config := map[string]interface{}{
			"config":  configPth,
			"pattern": "json_params_test_target",
		}

		cmd := command.New(testhelpers.BinPath(), "trigger", "--json-params", testhelpers.ToJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("trigger test - param override")
	{
		config := map[string]interface{}{
			"config":  configPth,
			"pattern": "exit_code_test_fail",
		}

		cmd := command.New(testhelpers.BinPath(), "trigger", "--json-params", testhelpers.ToJSON(t, config), "--pattern", "json_params_test_target")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("run test base64")
	{
		config := map[string]interface{}{
			"config":   configPth,
			"workflow": "json_params_test_target",
		}

		cmd := command.New(testhelpers.BinPath(), "run", "--json-params-base64", testhelpers.ToBase64(testhelpers.ToJSON(t, config)))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("trigger test base64")
	{
		config := map[string]interface{}{
			"config":  configPth,
			"pattern": "json_params_test_target",
		}

		cmd := command.New(testhelpers.BinPath(), "trigger", "--json-params-base64", testhelpers.ToBase64(testhelpers.ToJSON(t, config)))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("trigger check test")
	{
		config := map[string]interface{}{
			"config":  configPth,
			"pattern": "json_params_test_target",
			"format":  "json",
		}

		cmd := command.New(testhelpers.BinPath(), "trigger-check", "--json-params", testhelpers.ToJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"json_params_test_target","workflow":"json_params_test_target"}`, out)
	}

	t.Log("trigger check test - param override")
	{
		config := map[string]interface{}{
			"config":  configPth,
			"pattern": "json_params_test_target",
			"format":  "raw",
		}

		cmd := command.New(testhelpers.BinPath(), "trigger-check", "--json-params", testhelpers.ToJSON(t, config), "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"json_params_test_target","workflow":"json_params_test_target"}`, out)
	}
}