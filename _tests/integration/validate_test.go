package integration

import (
	"fmt"
	"path/filepath"
	"testing"

	"os"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/bitrise/cli"
	"github.com/stretchr/testify/require"
)

func Test_ValidateTest(t *testing.T) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__validate_test__")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir))
	}()

	t.Log("valid bitrise.yml")
	{
		cfgPath := "trigger_params_test_bitrise.yml"
		cfgData := ""
		secretsPath := ""
		secretsData := ""
		cfgDeprecatePath := ""

		result, _, err := cli.Validate(cfgPath, cfgDeprecatePath, cfgData, secretsPath, secretsData)

		require.NoError(t, err)
		require.Equal(t, true, result.Config.IsValid)
	}

	t.Log("valid - warning test - `-p` flag is deprecated")
	{
		cfgPath := ""
		cfgDeprecatePath := "trigger_params_test_bitrise.yml"
		cfgData := ""
		secretsPath := ""
		secretsData := ""

		result, warns, err := cli.Validate(cfgPath, cfgDeprecatePath, cfgData, secretsPath, secretsData)

		require.NoError(t, err)
		require.Equal(t, true, result.Config.IsValid)
		require.Equal(t, "'path' key is deprecated, use 'config' instead!", warns[0])
	}

	t.Log("valid - invalid workflow id")
	{
		configPth := filepath.Join(tmpDir, "bitrise.yml")
		require.NoError(t, fileutil.WriteStringToFile(configPth, invalidWorkflowIDBitriseYML))

		cmd := command.New(binPath(), "validate", "-c", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err)
		expected := "Config is valid: \x1b[32;1mtrue\x1b[0m\nWarning(s):\n- invalid workflow ID (invalid:id): doesn't conform to: [A-Za-z0-9-_.]"
		require.Equal(t, expected, out)
	}

	t.Log("invalid - empty bitrise.yml")
	{
		configPth := filepath.Join(tmpDir, "bitrise.yml")
		require.NoError(t, fileutil.WriteStringToFile(configPth, emptyBitriseYML))

		cmd := command.New(binPath(), "validate", "-c", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
		expected := fmt.Sprintf("Config is valid: \x1b[31;1mfalse\x1b[0m\nError: \x1b[31;1mConfig (path:%s) is not valid: empty config\x1b[0m", configPth)
		require.Equal(t, expected, out)
	}
}

func Test_ValidateTestJSON(t *testing.T) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__validate_test__")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir))
	}()

	t.Log("valid bitrise.yml")
	{
		cmd := command.New(binPath(), "validate", "-c", "trigger_params_test_bitrise.yml", "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err)
		require.Equal(t, "{\"data\":{\"config\":{\"is_valid\":true}}}", out)
	}

	t.Log("valid - warning test - `-p` flag is deprecated")
	{
		cmd := command.New(binPath(), "validate", "-p", "trigger_params_test_bitrise.yml", "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err)
		require.Equal(t, "{\"data\":{\"config\":{\"is_valid\":true}},\"warnings\":[\"'path' key is deprecated, use 'config' instead!\"]}", out)
	}

	t.Log("valid - invalid workflow id")
	{
		configPth := filepath.Join(tmpDir, "bitrise.yml")
		require.NoError(t, fileutil.WriteStringToFile(configPth, invalidWorkflowIDBitriseYML))

		cmd := command.New(binPath(), "validate", "-c", configPth, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err)
		expected := "{\"data\":{\"config\":{\"is_valid\":true,\"warnings\":[\"invalid workflow ID (invalid:id): doesn't conform to: [A-Za-z0-9-_.]\"]}}}"
		require.Equal(t, expected, out)
	}

	t.Log("invalid - empty bitrise.yml")
	{
		configPth := filepath.Join(tmpDir, "bitrise.yml")
		require.NoError(t, fileutil.WriteStringToFile(configPth, emptyBitriseYML))

		cmd := command.New(binPath(), "validate", "-c", configPth, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
		expected := fmt.Sprintf("{\"data\":{\"config\":{\"is_valid\":false,\"error\":\"Config (path:%s) is not valid: empty config\"}}}", configPth)
		require.Equal(t, expected, out)
	}

	t.Log("invalid - only one space in bitrise.yml")
	{
		configPth := filepath.Join(tmpDir, "bitrise.yml")
		require.NoError(t, fileutil.WriteStringToFile(configPth, spaceBitriseYML))

		cmd := command.New(binPath(), "validate", "-c", configPth, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
		expected := fmt.Sprintf("{\"data\":{\"config\":{\"is_valid\":false,\"error\":\"Config (path:%s) is not valid: missing format_version\"}}}", configPth)
		require.Equal(t, expected, out)
	}
}

func Test_SecretValidateTest(t *testing.T) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__validate_test__")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir))
	}()

	t.Log("valid secret")
	{
		cmd := command.New(binPath(), "validate", "-i", "global_flag_test_secrets.yml")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err)
		require.Equal(t, "Secret is valid: \x1b[32;1mtrue\x1b[0m", out)
	}

	t.Log("invalid - empty secret")
	{
		secretsPth := filepath.Join(tmpDir, "secrets.yml")
		require.NoError(t, fileutil.WriteStringToFile(secretsPth, emptySecret))

		cmd := command.New(binPath(), "validate", "-i", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
		expected := "Secret is valid: \x1b[31;1mfalse\x1b[0m\nError: \x1b[31;1mempty config\x1b[0m"
		require.Equal(t, expected, out)
	}

	t.Log("invalid - invalid secret model")
	{
		secretsPth := filepath.Join(tmpDir, "secrets.yml")
		require.NoError(t, fileutil.WriteStringToFile(secretsPth, invalidSecret))

		cmd := command.New(binPath(), "validate", "-i", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
		expected := "Secret is valid: \x1b[31;1mfalse\x1b[0m\nError: \x1b[31;1mInvalid inventory format: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into models.EnvsSerializeModel\x1b[0m"
		require.Equal(t, expected, out)
	}
}

func Test_SecretValidateTestJSON(t *testing.T) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__validate_test__")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir))
	}()

	t.Log("valid secret")
	{
		cmd := command.New(binPath(), "validate", "-i", "global_flag_test_secrets.yml", "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err)
		require.Equal(t, "{\"data\":{\"secrets\":{\"is_valid\":true}}}", out)
	}

	t.Log("invalid - empty config")
	{
		secretsPth := filepath.Join(tmpDir, "secrets.yml")
		require.NoError(t, fileutil.WriteStringToFile(secretsPth, emptySecret))

		cmd := command.New(binPath(), "validate", "-i", secretsPth, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
		expected := "{\"data\":{\"secrets\":{\"is_valid\":false,\"error\":\"empty config\"}}}"
		require.Equal(t, expected, out)
	}

	t.Log("invalid - invalid secret model")
	{
		secretsPth := filepath.Join(tmpDir, "secrets.yml")
		require.NoError(t, fileutil.WriteStringToFile(secretsPth, invalidSecret))

		cmd := command.New(binPath(), "validate", "-i", secretsPth, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
		expected := "{\"data\":{\"secrets\":{\"is_valid\":false,\"error\":\"Invalid inventory format: yaml: unmarshal errors:\\n  line 1: cannot unmarshal !!seq into models.EnvsSerializeModel\"}}}"
		require.Equal(t, expected, out)
	}
}

const emptySecret = ""
const invalidSecret = `- TEST: test`

const emptyBitriseYML = ""
const spaceBitriseYML = ` `
const invalidWorkflowIDBitriseYML = `format_version: 1.3.0
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

workflows:
  invalid:id:
`
