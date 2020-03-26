package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

const emptySecret = ""
const invalidSecret = `- TEST: test`

const emptyBitriseYML = ""
const spaceBitriseYML = ` `
const invalidWorkflowIDBitriseYML = `format_version: 1.3.0
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

workflows:
  invalid:id:
`
const runtimeLimit = 1000 * time.Millisecond
const runningTimeMsg = "test case too slow: %s is %s above limit"

func Test_ValidateTest(t *testing.T) {
	start := time.Now().Unix()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().Unix()))
	}(start)

	tmpDir, err := pathutil.NormalizedOSTempDirPath("__validate_test__")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir))
	}()

	t.Log("valid bitrise.yml")
	{
		var out string
		var err error
		elapsed := withRunningTimeCheck(func() {
			cmd := command.New(binPath(), "validate", "-c", "trigger_params_test_bitrise.yml")
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.NoError(t, err)
		require.Equal(t, "Config is valid: \x1b[32;1mtrue\x1b[0m", out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}

	t.Log("valid - warning test - `-p` flag is deprecated")
	{
		var out string
		var err error
		elapsed := withRunningTimeCheck(func() {
			cmd := command.New(binPath(), "validate", "-p", "trigger_params_test_bitrise.yml")
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.NoError(t, err)
		require.Equal(t, "Config is valid: \x1b[32;1mtrue\x1b[0m\nWarning(s):\n- 'path' key is deprecated, use 'config' instead!", out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}

	t.Log("valid - invalid workflow id")
	{
		configPth := filepath.Join(tmpDir, "bitrise.yml")
		require.NoError(t, fileutil.WriteStringToFile(configPth, invalidWorkflowIDBitriseYML))

		var out string
		var err error
		elapsed := withRunningTimeCheck(func() {
			cmd := command.New(binPath(), "validate", "-c", configPth)
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.NoError(t, err)
		expected := "Config is valid: \x1b[32;1mtrue\x1b[0m\nWarning(s):\n- invalid workflow ID (invalid:id): doesn't conform to: [A-Za-z0-9-_.]"
		require.Equal(t, expected, out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}

	t.Log("invalid - empty bitrise.yml")
	{
		configPth := filepath.Join(tmpDir, "bitrise.yml")
		require.NoError(t, fileutil.WriteStringToFile(configPth, emptyBitriseYML))

		var out string
		var err error
		elapsed := withRunningTimeCheck(func() {
			cmd := command.New(binPath(), "validate", "-c", configPth)
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.Error(t, err, out)
		expected := fmt.Sprintf("Config is valid: \x1b[31;1mfalse\x1b[0m\nError: \x1b[31;1mConfig (path:%s) is not valid: empty config\x1b[0m", configPth)
		require.Equal(t, expected, out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}
}

func Test_ValidateTestJSON(t *testing.T) {
	start := time.Now().Unix()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().Unix()))
	}(start)

	tmpDir, err := pathutil.NormalizedOSTempDirPath("__validate_test__")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir))
	}()

	t.Log("valid bitrise.yml")
	{
		var out string
		var err error
		elapsed := withRunningTimeCheck(func() {
			cmd := command.New(binPath(), "validate", "-c", "trigger_params_test_bitrise.yml", "--format", "json")
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.NoError(t, err)
		require.Equal(t, "{\"data\":{\"config\":{\"is_valid\":true}}}", out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}

	t.Log("valid - warning test - `-p` flag is deprecated")
	{
		var out string
		var err error
		elapsed := withRunningTimeCheck(func() {
			cmd := command.New(binPath(), "validate", "-p", "trigger_params_test_bitrise.yml", "--format", "json")
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.NoError(t, err)
		require.Equal(t, "{\"data\":{\"config\":{\"is_valid\":true}},\"warnings\":[\"'path' key is deprecated, use 'config' instead!\"]}", out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}

	t.Log("valid - invalid workflow id")
	{
		configPth := filepath.Join(tmpDir, "bitrise.yml")
		require.NoError(t, fileutil.WriteStringToFile(configPth, invalidWorkflowIDBitriseYML))

		var out string
		var err error
		elapsed := withRunningTimeCheck(func() {
			cmd := command.New(binPath(), "validate", "-c", configPth, "--format", "json")
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.NoError(t, err)
		expected := "{\"data\":{\"config\":{\"is_valid\":true,\"warnings\":[\"invalid workflow ID (invalid:id): doesn't conform to: [A-Za-z0-9-_.]\"]}}}"
		require.Equal(t, expected, out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}

	t.Log("invalid - empty bitrise.yml")
	{
		configPth := filepath.Join(tmpDir, "bitrise.yml")
		require.NoError(t, fileutil.WriteStringToFile(configPth, emptyBitriseYML))

		var out string
		var err error
		elapsed := withRunningTimeCheck(func() {
			cmd := command.New(binPath(), "validate", "-c", configPth, "--format", "json")
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.Error(t, err, out)
		expected := fmt.Sprintf("{\"data\":{\"config\":{\"is_valid\":false,\"error\":\"Config (path:%s) is not valid: empty config\"}}}", configPth)
		require.Equal(t, expected, out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}

	t.Log("invalid - only one space in bitrise.yml")
	{
		configPth := filepath.Join(tmpDir, "bitrise.yml")
		require.NoError(t, fileutil.WriteStringToFile(configPth, spaceBitriseYML))

		var out string
		var err error
		elapsed := withRunningTimeCheck(func() {
			cmd := command.New(binPath(), "validate", "-c", configPth, "--format", "json")
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.Error(t, err, out)
		expected := fmt.Sprintf("{\"data\":{\"config\":{\"is_valid\":false,\"error\":\"Config (path:%s) is not valid: missing format_version\"}}}", configPth)
		require.Equal(t, expected, out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}
}

func Test_SecretValidateTest(t *testing.T) {
	start := time.Now().Unix()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().Unix()))
	}(start)

	tmpDir, err := pathutil.NormalizedOSTempDirPath("__validate_test__")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir))
	}()

	t.Log("valid secret")
	{
		var out string
		var err error
		elapsed := withRunningTimeCheck(func() {
			cmd := command.New(binPath(), "validate", "-i", "global_flag_test_secrets.yml")
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.NoError(t, err)
		require.Equal(t, "Secret is valid: \x1b[32;1mtrue\x1b[0m", out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}

	t.Log("invalid - empty secret")
	{
		secretsPth := filepath.Join(tmpDir, "secrets.yml")
		require.NoError(t, fileutil.WriteStringToFile(secretsPth, emptySecret))

		var out string
		var err error
		elapsed := withRunningTimeCheck(func() {
			cmd := command.New(binPath(), "validate", "-i", secretsPth)
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.Error(t, err, out)
		expected := "Secret is valid: \x1b[31;1mfalse\x1b[0m\nError: \x1b[31;1mempty config\x1b[0m"
		require.Equal(t, expected, out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}

	t.Log("invalid - invalid secret model")
	{
		secretsPth := filepath.Join(tmpDir, "secrets.yml")
		require.NoError(t, fileutil.WriteStringToFile(secretsPth, invalidSecret))

		var out string
		var err error
		elapsed := withRunningTimeCheck(func() {
			cmd := command.New(binPath(), "validate", "-i", secretsPth)
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.Error(t, err, out)
		expected := "Secret is valid: \x1b[31;1mfalse\x1b[0m\nError: \x1b[31;1mInvalid inventory format: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into models.EnvsSerializeModel\x1b[0m"
		require.Equal(t, expected, out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}
}

func Test_SecretValidateTestJSON(t *testing.T) {
	start := time.Now().Unix()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().Unix()))
	}(start)

	tmpDir, err := pathutil.NormalizedOSTempDirPath("__validate_test__")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir))
	}()

	t.Log("valid secret")
	{
		var out string
		var err error
		cmd := command.New(binPath(), "validate", "-i", "global_flag_test_secrets.yml", "--format", "json")

		elapsed := withRunningTimeCheck(func() {
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.NoError(t, err)
		require.Equal(t, "{\"data\":{\"secrets\":{\"is_valid\":true}}}", out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}

	t.Log("invalid - empty config")
	{
		secretsPth := filepath.Join(tmpDir, "secrets.yml")
		require.NoError(t, fileutil.WriteStringToFile(secretsPth, emptySecret))

		var out string
		var err error
		cmd := command.New(binPath(), "validate", "-i", secretsPth, "--format", "json")

		elapsed := withRunningTimeCheck(func() {
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.Error(t, err, out)
		expected := "{\"data\":{\"secrets\":{\"is_valid\":false,\"error\":\"empty config\"}}}"
		require.Equal(t, expected, out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}

	t.Log("invalid - invalid secret model")
	{
		secretsPth := filepath.Join(tmpDir, "secrets.yml")
		require.NoError(t, fileutil.WriteStringToFile(secretsPth, invalidSecret))

		var out string
		var err error
		cmd := command.New(binPath(), "validate", "-i", secretsPth, "--format", "json")
		elapsed := withRunningTimeCheck(func() {
			out, err = cmd.RunAndReturnTrimmedCombinedOutput()
		}, runtimeLimit)
		require.Error(t, err, out)
		expected := "{\"data\":{\"secrets\":{\"is_valid\":false,\"error\":\"Invalid inventory format: yaml: unmarshal errors:\\n  line 1: cannot unmarshal !!seq into models.EnvsSerializeModel\"}}}"
		require.Equal(t, expected, out)
		require.Equal(t, true, elapsed < runtimeLimit, runningTimeMsg, elapsed, elapsed-runtimeLimit)
	}
}
