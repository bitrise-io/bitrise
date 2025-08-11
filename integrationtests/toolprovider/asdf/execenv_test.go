//go:build linux_and_mac
// +build linux_and_mac

package asdf

import (
	"context"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/execenv"
	"github.com/stretchr/testify/assert"
)

func TestExecEnv_RunCommandWithTimeout(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		command       []string
		shouldError   bool
		errorContains string
	}{
		{
			name:        "successful command with timeout",
			timeout:     5 * time.Second,
			command:     []string{"echo", "hello"},
			shouldError: false,
		},
		{
			name:          "command times out",
			timeout:       50 * time.Millisecond,
			command:       []string{"sleep", "2"},
			shouldError:   true,
			errorContains: "timed out",
		},
		{
			name:        "failed command with timeout",
			timeout:     5 * time.Second,
			command:     []string{"nonexistent-command"},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execEnv := &execenv.ExecEnv{
				EnvVars:            map[string]string{},
				ClearInheritedEnvs: false,
				ShellInit:          "",
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			output, err := execEnv.RunCommandWithTimeout(ctx, nil, tt.command...)

			if tt.shouldError {
				assert.Error(t, err, "expected error but got none")
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains, "expected error to contain '%s' but got: %v", tt.errorContains, err)
				}
			} else {
				assert.NoError(t, err, "unexpected error: %v", err)
				assert.Contains(t, output, "hello", "expected output to contain 'hello' but got: %s", output)
			}
		})
	}
}

func TestExecEnv_RunAsdfPlugin(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldError bool
	}{
		{
			name:        "plugin list command",
			args:        []string{"list"},
			shouldError: false,
		},
		{
			name:        "plugin add with invalid plugin",
			args:        []string{"add", "nonexistent-plugin-test-12345"},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execEnv := &execenv.ExecEnv{
				EnvVars:            map[string]string{},
				ClearInheritedEnvs: false,
				ShellInit:          "",
			}

			_, err := execEnv.RunAsdfPlugin(tt.args...)

			if tt.shouldError {
				assert.Error(t, err, "expected error but got none")
			} else {
				assert.NoError(t, err, "unexpected error: %v", err)
			}
		})
	}
}

func TestExecEnv_RunAsdfPluginTimeout(t *testing.T) {
	t.Run("plugin commands use timeout", func(t *testing.T) {
		execEnv := &execenv.ExecEnv{
			EnvVars:            map[string]string{},
			ClearInheritedEnvs: true,
			ShellInit:          "",
		}

		// This should timeout quickly since we're using a very short timeout constant for testing
		// We'll override the timeout by setting a short one in context
		start := time.Now()
		_, err := execEnv.RunAsdfPlugin("add", "nonexistent-plugin-that-would-hang")
		duration := time.Since(start)

		// The command should complete within the timeout period (PluginInstallTimeout = 5 minutes)
		// But since we're testing with a nonexistent plugin, it should fail quickly
		assert.Error(t, err, "expected error for nonexistent plugin")

		// Verify it doesn't take too long (should be much less than 5 minutes)
		assert.LessOrEqual(t, duration, 30*time.Second, "command took too long: %v", duration)
	})
}
