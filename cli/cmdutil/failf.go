package cmdutil

import (
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/v2/internal/style"
)

// Failf prints a formatted error to stderr and exits the process with status
// 1. It must not go through the global logger (log.Errorf): that logger's
// writer is stdout, which would leak the error into piped machine-readable
// output (e.g. `stack list -o json | jq`), and re-pointing the global logger
// would also move `bitrise run`'s step output to stderr with it.
func Failf(format string, args ...interface{}) {
	s := style.New(os.Stderr)
	_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", s.Failure.Render("Error:"), fmt.Sprintf(format, args...))

	if globalTracker != nil {
		globalTracker.Wait()
	}
	os.Exit(1)
}
