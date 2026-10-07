package local

import (
	"os"
	"testing"

	"github.com/bitrise-io/bitrise/v3/internal/buildcache"
)

func TestMain(m *testing.M) {
	// A CI or VM export must not make a test run the real build cache CLI.
	for _, key := range []string{buildcache.EnvActivateAll, buildcache.EnvActivateGradleMirrors} {
		_ = os.Unsetenv(key)
	}
	os.Exit(m.Run())
}
