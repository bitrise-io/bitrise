package integration

import (
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_SensitiveInput(t *testing.T) {
	configPth := "sensitive_input_test_bitrise.yml"
	secretsPth := "sensitive_input_test_secrets.yml"

	t.Log("env format tests")
	{
		cmd := command.New(binPath(), "run", "successful-formats", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		require.Equal(t, 2, strings.Count(out, "test content 1"), out)
		require.Equal(t, 2, strings.Count(out, "test content 2"), out)
	}
	t.Log("non sensitive")
	{
		cmd := command.New(binPath(), "run", "successful-non-sensitive", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		require.Equal(t, 1, strings.Count(out, "test env"), out)
	}
	configPth = "sensitive_input_nonsecret_test_bitrise.yml"
	t.Log("env is non secret")
	{
		cmd := command.New(binPath(), "run", "failure-non-secret", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)

		require.Equal(t, 1, strings.Count(out, "Security validation failed, error: security issue in script step: sensitive input (content) should be defined as a secret environment variable"), out)
	}
	configPth = "sensitive_input_nonisexpand_test_bitrise.yml"
	t.Log("disabled is_expand")
	{
		cmd := command.New(binPath(), "run", "failure-disabled-is-expand", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)

		require.Equal(t, 1, strings.Count(out, "security issue in script step: input (content) value should be defined as a secret environment variable, but is_expand set to: false"), out)
	}
	configPth = "sensitive_input_test_bitrise.yml"
	t.Log("no secrets file")
	{
		cmd := command.New(binPath(), "run", "successful-formats", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)

		require.Equal(t, 1, strings.Count(out, "security issue in script step: sensitive input (content) should be defined as a secret environment variable"), out)
	}
	configPth = "sensitive_input_test_bitrise.yml"
	secretsPth = "sensitive_input_test_empty_secrets.yml"
	t.Log("empty secrets file")
	{
		cmd := command.New(binPath(), "run", "successful-formats", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)

		require.Equal(t, 1, strings.Count(out, "empty config"), out)
	}
}
