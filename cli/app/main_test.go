package app

import (
	"os"
	"testing"

	"github.com/bitrise-io/bitrise/v2/analytics/analyticstest"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
)

// TestMain installs a no-op analytics tracker: RunE calls
// cmdutil.LogCommandParameters, which panics on the package-level tracker's
// zero value if nothing has called cmdutil.SetTracker (normally done once at
// real CLI startup, cli/cli.go).
func TestMain(m *testing.M) {
	cmdutil.SetTracker(analyticstest.NoOpTracker{})
	os.Exit(m.Run())
}
