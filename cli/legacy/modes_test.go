package legacy

import (
	"os"
	"strconv"
	"testing"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	secretFilteringFlag = "secret-filtering"
	debugModeKey        = "debug"
)

func secretFilteringTestCmd(t *testing.T, flag *bool) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "x"}
	cmd.Flags().Bool(secretFilteringFlag, false, "")
	if flag != nil {
		require.NoError(t, cmd.Flags().Set(secretFilteringFlag, strconv.FormatBool(*flag)))
	}
	return cmd
}

func setEnv(t *testing.T, key, value string, set bool) {
	t.Helper()
	// t.Setenv registers the original value for restoration; unset cases then
	// remove it so os.LookupEnv reports it as absent during the test.
	t.Setenv(key, value)
	if !set {
		require.NoError(t, os.Unsetenv(key))
	}
}

// run's --secret-filtering was never env-bound: the env is matched literally,
// non-bool values are ignored (no abort), and the flag wins when set.
func Test_RunSecretFilteringOverride(t *testing.T) {
	tests := []struct {
		name   string
		flag   *bool
		env    string
		envSet bool
		want   *bool
	}{
		{name: "no flag, no env", want: nil},
		{name: "env true", env: "true", envSet: true, want: pointers.NewBoolPtr(true)},
		{name: "env false", env: "false", envSet: true, want: pointers.NewBoolPtr(false)},
		{name: "env 0 is not literal false", env: "0", envSet: true, want: nil},
		{name: "env non-bool is ignored, not aborted", env: "yes", envSet: true, want: nil},
		{name: "env empty falls through", env: "", envSet: true, want: nil},
		{name: "flag false wins over env true", flag: pointers.NewBoolPtr(false), env: "true", envSet: true, want: pointers.NewBoolPtr(false)},
		{name: "flag true wins", flag: pointers.NewBoolPtr(true), want: pointers.NewBoolPtr(true)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, configs.IsSecretFilteringKey, tt.env, tt.envSet)
			assert.Equal(t, tt.want, RunSecretFilteringOverride(secretFilteringTestCmd(t, tt.flag), secretFilteringFlag))
		})
	}
}

// trigger's --secret-filtering was env-bound: the env is parsed with ParseBool,
// an empty value is false, and unset falls through. A non-bool value returns an
// error (the caller aborts).
func Test_SecretFilteringFlagOverride(t *testing.T) {
	tests := []struct {
		name    string
		flag    *bool
		env     string
		envSet  bool
		want    *bool
		wantErr bool
	}{
		{name: "no flag, no env", want: nil},
		{name: "env true", env: "true", envSet: true, want: pointers.NewBoolPtr(true)},
		{name: "env false", env: "false", envSet: true, want: pointers.NewBoolPtr(false)},
		{name: "env 0 parses to false", env: "0", envSet: true, want: pointers.NewBoolPtr(false)},
		{name: "env 1 parses to true", env: "1", envSet: true, want: pointers.NewBoolPtr(true)},
		{name: "env empty is false", env: "", envSet: true, want: pointers.NewBoolPtr(false)},
		{name: "env non-bool returns error", env: "notabool", envSet: true, wantErr: true},
		{name: "flag false wins over env true", flag: pointers.NewBoolPtr(false), env: "true", envSet: true, want: pointers.NewBoolPtr(false)},
		{name: "flag true wins", flag: pointers.NewBoolPtr(true), want: pointers.NewBoolPtr(true)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, configs.IsSecretFilteringKey, tt.env, tt.envSet)
			got, err := SecretFilteringFlagOverride(secretFilteringTestCmd(t, tt.flag), secretFilteringFlag)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// secret-envs-filtering has no flag and was never env-bound: literal match only.
func Test_SecretEnvsFilteringOverride(t *testing.T) {
	tests := []struct {
		name   string
		env    string
		envSet bool
		want   *bool
	}{
		{name: "no env", want: nil},
		{name: "env true", env: "true", envSet: true, want: pointers.NewBoolPtr(true)},
		{name: "env false", env: "false", envSet: true, want: pointers.NewBoolPtr(false)},
		{name: "env 0 is not literal false", env: "0", envSet: true, want: nil},
		{name: "env non-bool is ignored", env: "yes", envSet: true, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, configs.IsSecretEnvsFilteringKey, tt.env, tt.envSet)
			assert.Equal(t, tt.want, SecretEnvsFilteringOverride())
		})
	}
}

// debug mode is resolved before cobra parses: an explicit --debug flag wins,
// otherwise the DEBUG env (matched literally) decides.
func Test_IsDebugMode(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		env    string
		envSet bool
		want   bool
	}{
		{name: "No flag, no env", args: []string{}, want: false},
		{name: "Bare flag, one dash", args: []string{"-debug"}, want: true},
		{name: "Bare flag, two dashes", args: []string{"--debug"}, want: true},
		{name: "Value syntax true", args: []string{"-debug=true"}, want: true},
		{name: "Value syntax false", args: []string{"--debug=false"}, want: false},
		{name: "Invalid value falls back to env", args: []string{"--debug=notabool"}, want: false},
		{name: "Space syntax is not a bool flag", args: []string{"--debug true"}, want: false},
		{name: "Env enables debug", args: []string{}, env: "true", envSet: true, want: true},
		{name: "Explicit false wins over enabled env", args: []string{"--debug=false"}, env: "true", envSet: true, want: false},
		{name: "Explicit flag wins over disabled env", args: []string{"--debug"}, env: "false", envSet: true, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, configs.DebugModeEnvKey, tt.env, tt.envSet)
			assert.Equalf(t, tt.want, IsDebugMode(tt.args, debugModeKey), "IsDebugMode(%v) env=%q", tt.args, tt.env)
		})
	}
}
