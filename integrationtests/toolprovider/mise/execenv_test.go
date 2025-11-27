//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecEnv_RunMiseWithTimeout(t *testing.T) {
	testTimeout := 1 * time.Second

	miseInstallDir := t.TempDir()
	miseDataDir := t.TempDir()
	toolConfig := defaultTestToolConfig()
	miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir, toolConfig)
	require.NoError(t, err)

	err = miseProvider.Bootstrap()
	require.NoError(t, err)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "install command gets timeout",
			args: []string{"install", "ruby@3"},
		},
		{
			name: "help command",
			args: []string{"help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := miseProvider.ExecEnv.RunMiseWithTimeout(testTimeout, tt.args...)
			if err != nil {
				assert.Contains(t, err.Error(), "mise command timed out", "command output: %s", out)
			}
		})
	}
}
