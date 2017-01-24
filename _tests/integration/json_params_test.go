package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_JsonParams(t *testing.T) {
	configPth := "json_params_test_bitrise.yml"

	t.Log("run test")
	{
		config := map[string]string{
			"config":   configPth,
			"workflow": "json_params_test_target",
		}

		cmd := command.New(binPath(), "run", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("run test - param override")
	{
		config := map[string]string{
			"config":   configPth,
			"workflow": "exit_code_test_fail",
		}

		cmd := command.New(binPath(), "run", "--json-params", toJSON(t, config), "--workflow", "json_params_test_target")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("trigger test")
	{
		config := map[string]string{
			"config":  configPth,
			"pattern": "json_params_test_target",
		}

		cmd := command.New(binPath(), "trigger", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("trigger test - param override")
	{
		config := map[string]string{
			"config":  configPth,
			"pattern": "exit_code_test_fail",
		}

		cmd := command.New(binPath(), "trigger", "--json-params", toJSON(t, config), "--pattern", "json_params_test_target")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("run test base64")
	{
		config := map[string]string{
			"config":   configPth,
			"workflow": "json_params_test_target",
		}

		cmd := command.New(binPath(), "run", "--json-params-base64", toBase64(t, toJSON(t, config)))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("trigger test base64")
	{
		config := map[string]string{
			"config":  configPth,
			"pattern": "json_params_test_target",
		}

		cmd := command.New(binPath(), "trigger", "--json-params-base64", toBase64(t, toJSON(t, config)))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("trigger check test")
	{
		config := map[string]string{
			"config":  configPth,
			"pattern": "json_params_test_target",
			"format":  "json",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"json_params_test_target","workflow":"json_params_test_target"}`, out)
	}

	t.Log("trigger check test - param override")
	{
		config := map[string]string{
			"config":  configPth,
			"pattern": "json_params_test_target",
			"format":  "raw",
		}

		cmd := command.New(binPath(), "trigger-check", "--json-params", toJSON(t, config), "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, `{"pattern":"json_params_test_target","workflow":"json_params_test_target"}`, out)
	}
}
