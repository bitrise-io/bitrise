package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

			output, err := EnvmanRun(envstore, tmpDir, []string{"env"})
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

// Used for tests only, to parse env command output
func parseEnvRawOut(output string) (map[string]string, error) {
	// matches a single line like MYENVKEY_1=myvalue
	// Shell uses upperscore letters (plus numbers and underscore); Step inputs are lowerscore.
	// https://pubs.opengroup.org/onlinepubs/9699919799/:
	// > Environment variable names used by the utilities in the Shell and Utilities volume of POSIX.1-2017
	// > consist solely of uppercase letters, digits, and the <underscore> ( '_' ) from the characters defined
	// > in Portable Character Set and do not begin with a digit.
	// > Other characters may be permitted by an implementation; applications shall tolerate the presence of such names.
	r := regexp.MustCompile("^([a-zA-Z_][a-zA-Z0-9_]*)=(.*)$")

	lines := strings.Split(output, "\n")

	envs := make(map[string]string)
	lastKey := ""
	for _, line := range lines {
		match := r.FindStringSubmatch(line)

		// If no env is mathced, treat the line as the continuation of the env in the previous line.
		// `env` command output does not distinguish between a new env in a new line and
		// and environment value containing newline character.
		// Newline can be added for example: **  myenv=A$'\n'B env  ** (bash/zsh only)
		// If called from a script step, the content of the script contains newlines:
		/*
			content=#!/usr/bin/env bash
			set -ex
			current_envman="..."
			# ...
			go test -v ./_tests/integration/..."
		*/
		if match == nil {
			if lastKey != "" {
				envs[lastKey] += "\n" + line
			}
			continue
		}

		// If match not nil, must have 3 mathces at this point (the matched string and its subexpressions)
		if len(match) != 3 {
			return nil, fmt.Errorf("parseEnvRawOut() failed, match (%s) length is not 3 for line (%s).", match, line)
		}

		lastKey = match[1]
		envs[match[1]] = match[2]
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
			output: `RBENV_SHELL=zsh
_=/usr/local/bin/go
#!/bin/env bash
echo "ff"
A=`,
			want: map[string]string{
				"RBENV_SHELL": "zsh",
				"_": `/usr/local/bin/go
#!/bin/env bash
echo "ff"`,
				"A": "",
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
