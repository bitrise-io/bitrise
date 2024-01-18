//go:build linux_and_mac
// +build linux_and_mac

package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func TestSecretSharing(t *testing.T) {
	cmd := command.New(binPath(), "run", "secret-sharing", "--config", "secret_keys_sharing_test_bitrise.yml", "--inventory", "secret_keys_sharing_test_secrets.yml")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}
