package env

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func restoreEnviron(environ []string) error {
	currEnviron := os.Environ()
	for _, currEnv := range currEnviron {
		currEnvKey, _ := SplitEnv(currEnv)
		if err := os.Unsetenv(currEnvKey); err != nil {
			return err
		}
	}

	for _, envVar := range environ {
		key, value := SplitEnv(envVar)
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set %s=%s: %s", key, value, err)
		}
	}

	return nil
}

func TestGetDeclarationsSideEffects(t *testing.T) {
	for _, test := range EnvmanSharedTestCases {
		t.Run(test.Name, func(t *testing.T) {
			// Arrange
			cleanEnvs := os.Environ()

			for _, envVar := range test.Envs {
				err := envVar.FillMissingDefaults()
				require.NoError(t, err, "FillMissingDefaults()")
			}
			// Act
			got, err := GetDeclarationsSideEffects(test.Envs, &DefaultEnvironmentSource{})
			require.NoError(t, err, "GetDeclarationsSideEffects()")

			err = restoreEnviron(cleanEnvs)
			require.NoError(t, err, "restoreEnviron()")

			// Assert
			require.NotNil(t, got)
			require.Equal(t, test.Want, got.CommandHistory)

			// Want envs
			envsWant := make(map[string]string)
			for _, envVar := range os.Environ() {
				key, value := SplitEnv(envVar)
				envsWant[key] = value
			}

			for _, envCommand := range got.CommandHistory {
				switch envCommand.Action {
				case SetAction:
					envsWant[envCommand.Variable.Key] = envCommand.Variable.Value
				case UnsetAction:
					delete(envsWant, envCommand.Variable.Key)
				case SkipAction:
				default:
					t.Fatalf("compare() failed, invalid action: %d", envCommand.Action)
				}
			}

			require.Equal(t, envsWant, got.ResultEnvironment)
		})
	}
}

func TestSplitEnv(t *testing.T) {
	tests := []struct {
		name      string
		env       string
		wantKey   string
		wantValue string
	}{
		{
			name:      "simple case",
			env:       "A=B",
			wantKey:   "A",
			wantValue: "B",
		},
		{
			name:      "equals sign",
			env:       "A==B",
			wantKey:   "A",
			wantValue: "=B",
		},
		{
			name:      "",
			env:       "A=B=C=D",
			wantKey:   "A",
			wantValue: "B=C=D",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotValue := SplitEnv(tt.env)
			require.Equal(t, tt.wantKey, gotKey, "parseOSEnvs() gotKey")
			require.Equal(t, tt.wantValue, gotValue, "parseOSEnvs() gotvalue")
		})
	}
}
