package cmdutil

import (
	"os"
	"testing"
)

// TestMain clears the ambient BITRISE_* inputs this package's resolvers read.
// Without it a developer who exports one of them gets different results than
// CI: an exported BITRISE_TOKEN, for instance, makes TestNewAPIClient_ErrNoToken
// and TestNewAPIClient_FallsBackToAuthFile stop testing what they name. Tests
// that want a value set use t.Setenv, which still wins.
func TestMain(m *testing.M) {
	for _, key := range []string{"BITRISE_TOKEN", EnvWorkspaceID, EnvAppID, EnvAppIDLegacy, EnvOutput, EnvTheme} {
		if err := os.Unsetenv(key); err != nil {
			panic(err)
		}
	}
	os.Exit(m.Run())
}
