package integration

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func TestSensitiveRun(t *testing.T) {
	t.Log("prints a warning in case of security issue")
	{

		bitriseYML := `format_version: "5"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

workflows:
  primary:
    steps:
    - script:
        inputs:
        - content: echo "direct value"
          opts:
            is_sensitive: true
            is_expand: true`

		configBase64 := base64.StdEncoding.EncodeToString([]byte(bitriseYML))

		cmd := command.New(binPath(), "run", "primary", "--config-base64", configBase64)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err)
		require.True(t, strings.Contains(out, `Security validation failed: security issue in script step's content input: value should be defined as a secret environment variable, but does not starts with '$' mark`), out)
	}
}

func TestSensitiveTrigger(t *testing.T) {
	t.Log("prints a warning in case of security issue")
	{

		bitriseYML := `format_version: "5"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

trigger_map:
- push_branch: master
  workflow: primary

workflows:
  primary:
    steps:
    - script:
        inputs:
        - content: $SECRET_ENV
          opts:
            is_sensitive: true
            is_expand: false`

		configBase64 := base64.StdEncoding.EncodeToString([]byte(bitriseYML))

		cmd := command.New(binPath(), "trigger", "--push-branch", "master", "--config-base64", configBase64)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err)
		require.True(t, strings.Contains(out, `Security validation failed: security issue in script step's content input: value should be defined as a secret environment variable, but is_expand set to: false`), out)
	}
}

func TestSensitiveValidation(t *testing.T) {
	t.Log("affects sensitive inputs only")
	{
		bitriseYML := `format_version: "5"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

workflows:
  primary:
    steps:
    - script:
        inputs:
        - content: "direct value"
          opts:
            is_sensitive: false`

		configBase64 := base64.StdEncoding.EncodeToString([]byte(bitriseYML))

		cmd := command.New(binPath(), "validate", "--config-base64", configBase64, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("is expand is required")
	{

		bitriseYML := `format_version: "5"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

workflows:
  primary:
    steps:
    - script:
        inputs:
        - content: "direct value"
          opts:
            is_sensitive: true
            is_expand: false`

		configBase64 := base64.StdEncoding.EncodeToString([]byte(bitriseYML))

		cmd := command.New(binPath(), "validate", "--config-base64", configBase64, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err)
		require.Equal(t, `{"data":{"config":{"is_valid":false,"error":"security issue in script step's content input: value should be defined as a secret environment variable, but is_expand set to: false"}}}`, out, out)
	}

	t.Log("direct value is not allowed")
	{

		bitriseYML := `format_version: "5"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

workflows:
  primary:
    steps:
    - script:
        inputs:
        - content: "direct value"
          opts:
            is_sensitive: true
            is_expand: true`

		configBase64 := base64.StdEncoding.EncodeToString([]byte(bitriseYML))

		cmd := command.New(binPath(), "validate", "--config-base64", configBase64, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err)
		require.Equal(t, `{"data":{"config":{"is_valid":false,"error":"security issue in script step's content input: value should be defined as a secret environment variable, but does not starts with '$' mark"}}}`, out, out)
	}

	t.Log("value should start with '$' mark")
	{
		bitriseYML := `format_version: "5"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

workflows:
  primary:
    steps:
    - script:
        inputs:
        - content: "PREFIX_${SECRET_KEY}"
          opts:
            is_sensitive: true
            is_expand: true`

		configBase64 := base64.StdEncoding.EncodeToString([]byte(bitriseYML))

		cmd := command.New(binPath(), "validate", "--config-base64", configBase64, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err)
		require.Equal(t, `{"data":{"config":{"is_valid":false,"error":"security issue in script step's content input: value should be defined as a secret environment variable, but does not starts with '$' mark"}}}`, out, out)
	}

	t.Log("valid secrets")
	{
		bitriseYML := `format_version: "5"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

workflows:
  primary:
    steps:
    - script:
        inputs:
        - content: "$SECRET_KEY"
          opts:
            is_sensitive: true
            is_expand: true
    - script:
        inputs:
        - content: "${SECRET_KEY}"
          opts:
            is_sensitive: true
            is_expand: true
    - script:
        inputs:
        - content: "${SECRET_KEY}_WITH_SUFFIX"
          opts:
            is_sensitive: true
            is_expand: true`

		configBase64 := base64.StdEncoding.EncodeToString([]byte(bitriseYML))

		cmd := command.New(binPath(), "validate", "--config-base64", configBase64, "--format", "json")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}
