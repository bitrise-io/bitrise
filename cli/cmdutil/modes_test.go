package cmdutil

import (
	"os"
	"strconv"
	"testing"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setModeEnv(t *testing.T, key, value string, set bool) {
	t.Helper()
	// t.Setenv registers the original value for restoration; unset cases then
	// remove it so os.LookupEnv reports it as absent during the test.
	t.Setenv(key, value)
	if !set {
		require.NoError(t, os.Unsetenv(key))
	}
}

// boundBoolFlagSet returns a flag set with a bool flag bound to envKey (when set)
// via SetFlagEnvVar, with the flag optionally pre-set on the command line.
func boundBoolFlagSet(t *testing.T, name, envKey string, flag *bool) *pflag.FlagSet {
	t.Helper()
	fs := pflag.NewFlagSet("x", pflag.ContinueOnError)
	fs.Bool(name, false, "")
	if envKey != "" {
		SetFlagEnvVar(fs, name, envKey)
	}
	if flag != nil {
		require.NoError(t, fs.Set(name, strconv.FormatBool(*flag)))
	}
	return fs
}

func Test_envBool(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		envSet  bool
		want    *bool
		wantErr bool
	}{
		{name: "unset", want: nil},
		{name: "empty is treated as unset", env: "", envSet: true, want: nil},
		{name: "true", env: "true", envSet: true, want: pointers.NewBoolPtr(true)},
		{name: "false", env: "false", envSet: true, want: pointers.NewBoolPtr(false)},
		{name: "1 parses to true", env: "1", envSet: true, want: pointers.NewBoolPtr(true)},
		{name: "0 parses to false", env: "0", envSet: true, want: pointers.NewBoolPtr(false)},
		{name: "non-bool is an error", env: "notabool", envSet: true, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setModeEnv(t, configs.IsSecretEnvsFilteringKey, tt.env, tt.envSet)
			got, err := ResolveBoolEnv(configs.IsSecretEnvsFilteringKey)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// An env-bound flag (e.g. --ci, --secret-filtering) resolves flag-first, then the
// bound env parsed with ParseBool; a non-bool env value is an error.
func Test_resolveBoolFlagOrEnv_envBound(t *testing.T) {
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
		{name: "env 1 parses to true", env: "1", envSet: true, want: pointers.NewBoolPtr(true)},
		{name: "env empty falls through", env: "", envSet: true, want: nil},
		{name: "env non-bool is an error", env: "notabool", envSet: true, wantErr: true},
		{name: "flag false wins over env true", flag: pointers.NewBoolPtr(false), env: "true", envSet: true, want: pointers.NewBoolPtr(false)},
		{name: "flag true wins over env false", flag: pointers.NewBoolPtr(true), env: "false", envSet: true, want: pointers.NewBoolPtr(true)},
		{name: "flag true, no env", flag: pointers.NewBoolPtr(true), want: pointers.NewBoolPtr(true)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setModeEnv(t, configs.IsSecretFilteringKey, tt.env, tt.envSet)
			fs := boundBoolFlagSet(t, SecretFilteringKey, configs.IsSecretFilteringKey, tt.flag)
			got, err := ResolveBoolFlagOrEnv(fs, SecretFilteringKey)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// A flag with no env binding (e.g. --pr) resolves to the flag value or nil; its
// env var, if any, is never consulted.
func Test_resolveBoolFlagOrEnv_noEnvBinding(t *testing.T) {
	t.Run("no flag returns nil even when a same-named env is set", func(t *testing.T) {
		setModeEnv(t, configs.PRModeEnvKey, "true", true)
		fs := boundBoolFlagSet(t, PRKey, "", nil)
		got, err := ResolveBoolFlagOrEnv(fs, PRKey)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("flag value is returned", func(t *testing.T) {
		fs := boundBoolFlagSet(t, PRKey, "", pointers.NewBoolPtr(true))
		got, err := ResolveBoolFlagOrEnv(fs, PRKey)
		require.NoError(t, err)
		assert.Equal(t, pointers.NewBoolPtr(true), got)
	})
}
