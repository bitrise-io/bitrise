package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/envman/v2/env"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__envman__")
	require.NoError(t, err)

	envstore := filepath.Join(tmpDir, ".envstore")

	for _, tt := range env.EnvmanSharedTestCases {
		t.Run(tt.Name, func(t *testing.T) {
			// Clear and init
			err := EnvmanInitAtPath(envstore)
			require.NoError(t, err, "EnvmanInitAtPath()")

			for _, envVar := range tt.Envs {
				if err := envVar.FillMissingDefaults(); err != nil {
					require.NoError(t, err, "FillMissingDefaults()")
				}
			}

			err = ExportEnvironmentsList(envstore, tt.Envs)
			require.NoError(t, err, "ExportEnvironmentsList()")

			// -0 makes env NUL-terminate each entry instead of using newlines.
			// A value that itself ends in or contains a newline is then still
			// unambiguous, so a trailing newline survives the round-trip instead
			// of being swallowed as if it were the separator to the next entry.
			output, err := EnvmanRun(envstore, tmpDir, []string{"env", "-0"})
			require.NoError(t, err, "EnvmanRun()")

			gotOut, err := parseEnvRawOut(output)
			require.NoError(t, err, "parseEnvRawOut()")

			// Envman and this test process run in different folders,
			// so we can't compare the PWD env var.
			delete(gotOut, "PWD")

			// Want envs
			envsWant := make(map[string]string)
			for _, envVar := range os.Environ() {
				key, value := env.SplitEnv(envVar)
				if key == "PWD" {
					continue
				}
				envsWant[key] = value
			}

			for _, envCommand := range tt.Want {
				switch envCommand.Action {
				case env.SetAction:
					envsWant[envCommand.Variable.Key] = envCommand.Variable.Value
				case env.UnsetAction:
					delete(envsWant, envCommand.Variable.Key)
				case env.SkipAction:
				default:
					t.Fatalf("compare() failed, invalid action: %d", envCommand.Action)
				}
			}

			require.Equal(t, envsWant, gotOut)
		})
	}

}

// Used for tests only, to parse the NUL-separated output of `env -0`.
// Each record is a single KEY=VALUE pair whose value is verbatim, so a value
// containing newlines needs no reassembly.
func parseEnvRawOut(output string) (map[string]string, error) {
	envs := make(map[string]string)
	for _, record := range strings.Split(output, "\x00") {
		if record == "" {
			continue
		}

		key, value, found := strings.Cut(record, "=")
		if !found {
			return nil, fmt.Errorf("parseEnvRawOut() failed, no '=' in record (%q)", record)
		}
		envs[key] = value
	}

	return envs, nil
}

func Test_parseEnvRawOut(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   map[string]string
	}{
		{
			name:   "value with embedded newlines is preserved verbatim",
			output: "RBENV_SHELL=zsh\x00_=/usr/local/bin/go\n#!/bin/env bash\necho \"ff\"\x00A=\x00",
			want: map[string]string{
				"RBENV_SHELL": "zsh",
				"_": `/usr/local/bin/go
#!/bin/env bash
echo "ff"`,
				"A": "",
			},
		},
		{
			name:   "trailing newline in a value is kept",
			output: "KEY=-----END-----\n\x00",
			want: map[string]string{
				"KEY": "-----END-----\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEnvRawOut(tt.output)
			require.NoError(t, err, "parseEnvRawOut()")
			require.Equal(t, got, tt.want)
		})
	}
}
